// API Client for Tournament Service REST API
// BASE_PATH environment variable prepends a path prefix to all API routes
// For dev container setup, this is typically /tournament
const API_BASE = '/tournament';  // Must match BASE_PATH environment variable
const FETCH_TIMEOUT = 10000;  // 10 second timeout for API calls

/**
 * Fetch with timeout wrapper
 * @param {string} url - URL to fetch
 * @param {Object} options - Fetch options
 * @param {number} timeout - Timeout in milliseconds
 * @returns {Promise<Response>} Fetch response
 */
async function fetchWithTimeout(url, options = {}, timeout = FETCH_TIMEOUT) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);
    
    try {
        const response = await fetch(url, {
            ...options,
            signal: controller.signal
        });
        clearTimeout(timeoutId);
        return response;
    } catch (error) {
        clearTimeout(timeoutId);
        if (error.name === 'AbortError') {
            throw new Error('Request timeout');
        }
        throw error;
    }
}

/**
 * Fetch all tournaments from REST API
 * @param {string} namespace - Namespace to fetch tournaments from (defaults to 'test-ns')
 * @returns {Promise<Array>} Array of tournament objects
 */
async function fetchTournaments(namespace = 'test-ns') {
    const response = await fetchWithTimeout(`${API_BASE}/v1/public/namespace/${namespace}/tournaments`, {
        method: 'GET',
        headers: {
            'Accept': 'application/json',
        },
    });
    
    if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    const data = await response.json();
    return data.tournaments || [];
}

/**
 * Fetch single tournament by ID
 * @param {string} namespace - Tournament namespace
 * @param {string} tournamentId - Tournament ID
 * @returns {Promise<Object>} Tournament object
 */
async function fetchTournament(namespace, tournamentId) {
    const response = await fetchWithTimeout(
        `${API_BASE}/v1/public/namespace/${namespace}/tournaments/${tournamentId}`,
        {
            method: 'GET',
            headers: {
                'Accept': 'application/json',
            },
        }
    );
    
    if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    const data = await response.json();
    return data.tournament;
}

/**
 * Fetch participants for a tournament
 * @param {string} namespace - Namespace ID
 * @param {string} tournamentId - Tournament ID
 * @returns {Promise<Array>} Array of participant objects
 */
async function fetchParticipants(namespace, tournamentId) {
    const url = `${API_BASE}/v1/public/namespace/${namespace}/tournaments/${tournamentId}/participants`;
    const response = await fetchWithTimeout(url);
    
    if (!response.ok) {
        throw new Error(`Failed to fetch participants: ${response.statusText}`);
    }
    
    const data = await response.json();
    return data.participants || [];
}

/**
 * Fetch all matches for a tournament
 * @param {string} namespace - Namespace ID
 * @param {string} tournamentId - Tournament ID
 * @returns {Promise<Object>} Object with matches, totalRounds, currentRound
 */
async function fetchMatches(namespace, tournamentId) {
    const url = `${API_BASE}/v1/public/namespace/${namespace}/tournaments/${tournamentId}/matches`;
    const response = await fetchWithTimeout(url);
    
    if (!response.ok) {
        throw new Error(`Failed to fetch matches: ${response.statusText}`);
    }
    
    const data = await response.json();
    return {
        matches: data.matches || [],
        totalRounds: data.total_rounds || 0,
        currentRound: data.current_round || 0,
    };
}
