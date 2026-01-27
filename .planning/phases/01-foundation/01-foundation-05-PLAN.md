---
phase: 01-foundation
plan: 05
type: execute
wave: 1
depends_on: []
files_modified: ["pkg/proto/tournament.proto"]
autonomous: true
gap_closure: true

must_haves:
  truths:
    - "Service token authentication enables game server access to tournament operations"
  artifacts:
    - path: "pkg/proto/tournament.proto"
      provides: "Service token security definitions"
      contains: "securityDefinitions"
    - path: "pkg/proto/tournament.proto"
      provides: "Security requirement annotations"
      contains: "security"
  key_links:
    - from: "pkg/proto/tournament.proto"
      to: "Game server auth"
      via: "securityDefinitions"
      pattern: "securityDefinitions.*ServiceToken"
    - from: "pkg/proto/tournament.proto"
      to: "validateServiceToken"
      via: "security requirement annotations"
      pattern: "security:.*ServiceToken"
---

<objective>
Add missing service token authentication security definitions to tournament.proto to complete AUTH-03 requirement.

Purpose: Enable game servers to authenticate using service tokens for API access as specified in AUTH-03.
Output: Complete service token authentication infrastructure from proto definition to validation.
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md

# Gap context from verification
@.planning/phases/01-foundation/01-foundation-VERIFICATION.md

# Existing service token implementation
@pkg/common/auth_interceptors.go
@pkg/proto/tournament.proto
</context>

<tasks>

<task type="auto">
  <name>Add service token security definitions to tournament.proto</name>
  <files>pkg/proto/tournament.proto</files>
  <action>
    1. Read current tournament.proto to understand existing structure
    2. Add securityDefinitions section after import statements:
       ```
       // Security definitions for authentication methods
       security_definitions: {
         security_definition: {
           key: "BearerToken"
           value: {
             type: "apiKey"
             name: "Authorization"
             in: "header"
             description: "OAuth2 Bearer token for user authentication"
           }
         }
         security_definition: {
           key: "ServiceToken"
           value: {
             type: "apiKey"
             name: "X-Service-Token"
             in: "header"
             description: "Service token for game server authentication"
           }
         }
       }
       ```
    3. Add security requirements to methods that game servers need:
       - StartTournament (game servers need to report results)
       - GetTournament (game servers need tournament info)
       - ListTournaments (game servers need to discover tournaments)
       Add to each rpc definition:
       ```
       option (google.api.api_auth) = {
         security: {
           security_requirement: {
             key: "BearerToken"
           }
           security_requirement: {
             key: "ServiceToken"
           }
         }
       }
       ```
  </action>
  <verify>grep -q "securityDefinitions" pkg/proto/tournament.proto && grep -q "ServiceToken" pkg/proto/tournament.proto</verify>
  <done>Service token security definitions added to proto with proper method annotations</done>
</task>

<task type="auto">
  <name>Regenerate protobuf files to include security definitions</name>
  <files>pkg/pb/tournament.pb.go, pkg/pb/tournament.pb.gw.go</files>
  <action>
    1. Run protobuf generation command to update generated files:
       ```
       protoc --go_out=. --go-grpc_out=. --grpc-gateway_out=. \
         --go_opt=paths=source_relative \
         --go-grpc_opt=paths=source_relative \
         --grpc-gateway_opt=paths=source_relative \
         pkg/proto/tournament.proto
       ```
    2. Check that generated files include the new security definitions
    3. The REST gateway generation should now include security requirements in OpenAPI
  </action>
  <verify>ls -la pkg/pb/tournament.*.go && grep -q "ServiceToken" pkg/pb/tournament.pb.gw.go</verify>
  <done>Protobuf files regenerated with service token security definitions</done>
</task>

</tasks>

<verification>
- Security definitions are present in tournament.proto
- Service token authentication is defined alongside Bearer token
- Appropriate methods have security requirements that accept either token
- Generated files include the new security definitions
- No syntax errors in protobuf (generation succeeds)
</verification>

<success_criteria>
Game servers can authenticate using service tokens for tournament operations, completing AUTH-03 requirement.

Service token authentication flow:
1. Game server sends X-Service-Token header
2. validateServiceToken method in auth_interceptors.go validates the token
3. Tournament operations accept service token authentication for designated methods
4. REST API properly documents service token authentication in OpenAPI
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation/01-foundation-05-SUMMARY.md`
</output>