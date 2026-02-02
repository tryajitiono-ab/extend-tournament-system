// API Client for Tournament Service REST API
const API_BASE = '';  // Same origin, no base path needed

/**
 * Fetch all tournaments from REST API
 * @returns {Promise<Array>} Array of tournament objects
 */
async function fetchTournaments() {
    const response = await fetch(`${API_BASE}/v1/public/tournaments`, {
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
    const url = `${API_BASE}/${namespace}/tournaments/${tournamentId}/participants`;
    const response = await fetch(url);
    
    if (!response.ok) {
        throw new Error(`Failed to fetch participants: ${response.statusText}`);
    }
    
    const data = await response.json();
    return data.participants || [];
}
