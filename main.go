// Copyright (c) 2023-2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"extend-tournament-service/pkg/common"
	"extend-tournament-service/pkg/server"
	"extend-tournament-service/pkg/service"
	"extend-tournament-service/pkg/storage"

	serviceextension "extend-tournament-service/pkg/pb"

	sdkAuth "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth"
	prometheusGrpc "github.com/grpc-ecosystem/go-grpc-prometheus"
	prometheusCollectors "github.com/prometheus/client_golang/prometheus/collectors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const (
	metricsEndpoint     = "/metrics"
	metricsPort         = 8080
	grpcServerPort      = 6565
	grpcGatewayHTTPPort = 8000
)

var (
	serviceName = "extend-app-service-extension"
	logLevelStr = common.GetEnv("LOG_LEVEL", "info")
	basePath    = common.GetBasePath()
)

func parseSlogLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "fatal", "panic":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func main() {
	slogLevel := parseSlogLevel(logLevelStr)
	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	loggingOptions := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall, logging.PayloadReceived, logging.PayloadSent),
		logging.WithFieldsFromContext(func(ctx context.Context) logging.Fields {
			if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
				return logging.Fields{"traceID", span.TraceID().String()}
			}

			return nil
		}),
		logging.WithLevels(logging.DefaultClientCodeToLevel),
		logging.WithDurationField(logging.DurationToDurationField),
	}

	unaryServerInterceptors := []grpc.UnaryServerInterceptor{
		prometheusGrpc.UnaryServerInterceptor,
		logging.UnaryServerInterceptor(common.InterceptorLogger(logger), loggingOptions...),
	}
	streamServerInterceptors := []grpc.StreamServerInterceptor{
		prometheusGrpc.StreamServerInterceptor,
		logging.StreamServerInterceptor(common.InterceptorLogger(logger), loggingOptions...),
	}

	// Preparing the IAM authorization
	var tokenRepo repository.TokenRepository = sdkAuth.DefaultTokenRepositoryImpl()
	var configRepo repository.ConfigRepository = sdkAuth.DefaultConfigRepositoryImpl()
	var refreshRepo repository.RefreshTokenRepository = &sdkAuth.RefreshTokenImpl{RefreshRate: 0.8, AutoRefresh: true}

	oauthService := iam.OAuth20Service{
		Client:                 factory.NewIamClient(configRepo),
		TokenRepository:        tokenRepo,
		RefreshTokenRepository: refreshRepo,
		ConfigRepository:       configRepo,
	}

	if strings.ToLower(common.GetEnv("PLUGIN_GRPC_SERVER_AUTH_ENABLED", "true")) == "true" {
		refreshInterval := common.GetEnvInt("REFRESH_INTERVAL", 600)
		common.Validator = common.NewTokenValidator(oauthService, time.Duration(refreshInterval)*time.Second, true)
		err := common.Validator.Initialize(ctx)
		if err != nil {
			logger.Info(err.Error())
		}

		unaryServerInterceptor := common.NewUnaryAuthServerIntercept()
		serverServerInterceptor := common.NewStreamAuthServerIntercept()

		unaryServerInterceptors = append(unaryServerInterceptors, unaryServerInterceptor)
		streamServerInterceptors = append(streamServerInterceptors, serverServerInterceptor)
		logger.Info("added auth interceptors")
	}

	// Create gRPC Server
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(unaryServerInterceptors...),
		grpc.ChainStreamInterceptor(streamServerInterceptors...),
	)

	// Configure IAM authorization
	clientId := configRepo.GetClientId()
	clientSecret := configRepo.GetClientSecret()
	err := oauthService.LoginClient(&clientId, &clientSecret)
	if err != nil {
		logger.Error("error unable to login using clientId and clientSecret", "error", err)
		os.Exit(1)
	}

	// Initialize MongoDB storage
	docdbHost := common.GetEnv("DOCDB_HOST", "mongodb:27017")
	docdbUsername := common.GetEnv("DOCDB_USERNAME", "admin")
	docdbPassword := common.GetEnv("DOCDB_PASSWORD", "password")
	docdbCaCertFilePath := common.GetEnv("DOCDB_CA_CERT_FILE_PATH", "")
	mongoConnectionString := fmt.Sprintf("mongodb://%s:%s@%s/?tls=true&tlsCAFile=%s", docdbUsername, docdbPassword, docdbHost, docdbCaCertFilePath)
	minPoolSize := uint64(common.GetEnvInt("DOCDB_MIN_POOL_SIZE", 5))
	maxPoolSize := uint64(common.GetEnvInt("DOCDB_MAX_POOL_SIZE", 30))
	mongoDatabase := common.GetEnv("DOCDB_DATABASE_NAME", "tournament_service")

	mongoClient, err := mongo.Connect(ctx, options.Client().
		ApplyURI(mongoConnectionString).
		SetRetryWrites(false).
		SetMinPoolSize(minPoolSize).
		SetMaxPoolSize(maxPoolSize))
	if err != nil {
		logger.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			logger.Error("failed to disconnect from MongoDB", "error", err)
		}
	}()

	// Ping MongoDB to verify connection
	if err := mongoClient.Ping(ctx, nil); err != nil {
		logger.Error("failed to ping MongoDB", "error", err)
		os.Exit(1)
	}
	logger.Info("connected to MongoDB", "uri", mongoConnectionString, "database", mongoDatabase)

	// Initialize the AccelByte CloudSave service
	adminGameRecordService := cloudsave.AdminGameRecordService{
		Client:          factory.NewCloudsaveClient(configRepo),
		TokenRepository: tokenRepo,
	}

	cloudSaveStorage := storage.NewCloudSaveStorage(&adminGameRecordService)

	// Initialize storage registry with MongoDB
	storageRegistry := storage.NewStorageRegistry(mongoClient, mongoDatabase, logger)

	// Create all storage instances using registry
	tournamentStorage := storageRegistry.NewTournamentStorage()
	participantStorage := storageRegistry.NewParticipantStorage()
	matchStorage := storageRegistry.NewMatchStorage()

	// Ensure all database indexes are created
	if err := storageRegistry.EnsureAllIndexes(ctx); err != nil {
		logger.Error("failed to create storage indexes", "error", err)
		// Continue execution but log the error
	}

	// Initialize Participant service
	participantService := service.NewParticipantService(
		participantStorage,
		tournamentStorage,
		logger,
	)

	// Initialize Tournament authentication interceptor
	tournamentAuthInterceptor := common.NewTournamentAuthInterceptor(oauthService, common.Validator, logger)

	// Initialize Match service
	matchService := service.NewMatchService(
		matchStorage,
		tournamentStorage,
		tournamentAuthInterceptor,
		logger,
	)

	// Initialize Tournament service
	tournamentService := service.NewTournamentServiceServer(tokenRepo, configRepo, refreshRepo, tournamentStorage, participantStorage, tournamentAuthInterceptor, logger)

	// Add tournament auth interceptors to chain
	unaryServerInterceptors = append(unaryServerInterceptors, tournamentAuthInterceptor.TournamentUnaryInterceptor())
	streamServerInterceptors = append(streamServerInterceptors, tournamentAuthInterceptor.TournamentStreamInterceptor())

	// Register Tournament Service with participant and match integration
	tournamentServer := server.NewTournamentServer(
		tournamentService,
		participantService,
		matchService,
		logger,
	)
	serviceextension.RegisterTournamentServiceServer(s, tournamentServer)

	// Enable gRPC Reflection
	reflection.Register(s)

	// Enable gRPC Health Check
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())

	// Create a new HTTP server for the gRPC-Gateway
	grpcGateway, err := common.NewGateway(ctx, fmt.Sprintf("localhost:%d", grpcServerPort), basePath)
	if err != nil {
		logger.Error("failed to create gRPC-Gateway", "error", err)
		os.Exit(1)
	}

	// Start the gRPC-Gateway HTTP server
	go func() {
		swaggerDir := "gateway/apidocs" // Path to swagger directory
		grpcGatewayHTTPServer := newGRPCGatewayHTTPServer(fmt.Sprintf(":%d", grpcGatewayHTTPPort), grpcGateway, logger, swaggerDir)
		logger.Info("starting gRPC-Gateway HTTP server", "port", grpcGatewayHTTPPort)
		if err := grpcGatewayHTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to run gRPC-Gateway HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	prometheusGrpc.Register(s)

	// Register Prometheus Metrics
	prometheusRegistry := prometheus.NewRegistry()
	prometheusRegistry.MustRegister(
		prometheusCollectors.NewGoCollector(),
		prometheusCollectors.NewProcessCollector(prometheusCollectors.ProcessCollectorOpts{}),
		prometheusGrpc.DefaultServerMetrics,
	)

	go func() {
		http.Handle(metricsEndpoint, promhttp.HandlerFor(prometheusRegistry, promhttp.HandlerOpts{}))
		if err := http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil); err != nil {
			logger.Error("failed to start metrics server", "error", err)
			os.Exit(1)
		}
	}()
	logger.Info("serving prometheus metrics", "port", metricsPort, "endpoint", metricsEndpoint)

	// Set Tracer Provider
	if val := common.GetEnv("OTEL_SERVICE_NAME", ""); val != "" {
		serviceName = "extend-app-se-" + strings.ToLower(val)
	}
	tracerProvider, err := common.NewTracerProvider(serviceName)
	if err != nil {
		logger.Error("failed to create tracer provider", "error", err)
		os.Exit(1)
	}
	otel.SetTracerProvider(tracerProvider)
	defer func(ctx context.Context) {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			logger.Error("failed to shutdown tracer provider", "error", err)
			os.Exit(1)
		}
	}(ctx)

	// Set Text Map Propagator
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			b3.New(),
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// Start gRPC Server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcServerPort))
	if err != nil {
		logger.Error("failed to listen to tcp", "port", grpcServerPort, "error", err)
		os.Exit(1)
	}
	go func() {
		if err = s.Serve(lis); err != nil {
			logger.Error("failed to run gRPC server", "error", err)
			os.Exit(1)
		}
	}()

	logger.Info("app server started", "service", serviceName)

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	logger.Info("signal received")
}

func newGRPCGatewayHTTPServer(
	addr string, handler http.Handler, logger *slog.Logger, swaggerDir string,
) *http.Server {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Add the gRPC-Gateway handler
	mux.Handle("/", handler)

	// Serve Swagger UI and JSON
	serveSwaggerUI(mux)
	serveSwaggerJSON(mux, swaggerDir)

	// Add logging middleware
	loggedMux := loggingMiddleware(logger, mux)

	return &http.Server{
		Addr:     addr,
		Handler:  loggedMux,
		ErrorLog: log.New(os.Stderr, "httpSrv: ", log.LstdFlags), // Configure the logger for the HTTP server
	}
}

// loggingMiddleware is a middleware that logs HTTP requests
func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", duration,
		)
	})
}

func serveSwaggerUI(mux *http.ServeMux) {
	swaggerUIDir := "third_party/swagger-ui"
	fileServer := http.FileServer(http.Dir(swaggerUIDir))
	swaggerUiPath := fmt.Sprintf("%s/apidocs/", basePath)
	mux.Handle(swaggerUiPath, http.StripPrefix(swaggerUiPath, fileServer))
}

func serveSwaggerJSON(mux *http.ServeMux, swaggerDir string) {
	fileHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		matchingFiles, err := filepath.Glob(filepath.Join(swaggerDir, "*.swagger.json"))
		if err != nil || len(matchingFiles) == 0 {
			http.Error(w, "Error finding Swagger JSON file", http.StatusInternalServerError)

			return
		}

		firstMatchingFile := matchingFiles[0]
		swagger, err := loads.Spec(firstMatchingFile)
		if err != nil {
			http.Error(w, "Error parsing Swagger JSON file", http.StatusInternalServerError)

			return
		}

		// Update the base path
		swagger.Spec().BasePath = basePath

		updatedSwagger, err := swagger.Spec().MarshalJSON()
		if err != nil {
			http.Error(w, "Error serializing updated Swagger JSON", http.StatusInternalServerError)

			return
		}
		var prettySwagger bytes.Buffer
		err = json.Indent(&prettySwagger, updatedSwagger, "", "  ")
		if err != nil {
			http.Error(w, "Error formatting updated Swagger JSON", http.StatusInternalServerError)

			return
		}

		_, err = w.Write(prettySwagger.Bytes())
		if err != nil {
			http.Error(w, "Error writing Swagger JSON response", http.StatusInternalServerError)

			return
		}
	})
	apidocsPath := fmt.Sprintf("%s/apidocs/api.json", basePath)
	mux.Handle(apidocsPath, fileHandler)
}
