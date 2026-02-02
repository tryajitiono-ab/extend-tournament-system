// API Client for Tournament Service REST API
// BASE_PATH environment variable prepends a path prefix to all API routes
// For dev container setup, this is typically /tournament
const API_BASE = '/tournament';  // Must match BASE_PATH environment variable

/**
 * Fetch all tournaments from REST API
 * @param {string} namespace - Namespace to fetch tournaments from (defaults to 'test-ns')
 * @returns {Promise<Array>} Array of tournament objects
 */
async function fetchTournaments(namespace = 'test-ns') {
    const response = await fetch(`${API_BASE}/v1/public/namespace/${namespace}/tournaments`, {
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
    const response = await fetch(
        `${API_BASE}/v1/public/namespaces/${namespace}/tournaments/${tournamentId}`,
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
    
    return response.json();
}

/**
 * Fetch participants for a tournament
 * @param {string} namespace - Namespace ID
 * @param {string} tournamentId - Tournament ID
 * @returns {Promise<Array>} Array of participant objects
 */
async function fetchParticipants(namespace, tournamentId) {
    const url = `${API_BASE}/v1/public/namespace/${namespace}/tournaments/${tournamentId}/participants`;
    const response = await fetch(url);
    
    if (!response.ok) {
        throw new Error(`Failed to fetch participants: ${response.statusText}`);
    }
    
    const data = await response.json();
    return data.participants || [];
}
