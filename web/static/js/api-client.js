// API Client for Tournament Service
// Provides functions to interact with the REST API

const API_BASE = '/v1/public/namespaces';

/**
 * Fetch a single tournament
 * @param {string} namespace - Namespace ID
 * @param {string} tournamentId - Tournament ID
 * @returns {Promise<Object>} Tournament object
 */
async function fetchTournament(namespace, tournamentId) {
    const url = `${API_BASE}/${namespace}/tournaments/${tournamentId}`;
    const response = await fetch(url);
    
    if (!response.ok) {
        throw new Error(`Failed to fetch tournament: ${response.statusText}`);
    }
    
    return await response.json();
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
