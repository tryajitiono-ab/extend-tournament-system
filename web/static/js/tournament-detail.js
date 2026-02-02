// Tournament Detail Page Logic

// DOM elements
let tournamentInfoEl;
let tournamentNameEl;
let tournamentMetaEl;
let tournamentDescriptionEl;
let participantListEl;
let participantLoadingEl;
let participantEmptyEl;
let loadingStateEl;
let errorBannerEl;
let retryBtn;

// Tournament data
let currentNamespace;
let currentTournamentId;

// Initialize page
document.addEventListener('DOMContentLoaded', () => {
    tournamentInfoEl = document.getElementById('tournament-info');
    tournamentNameEl = document.getElementById('tournament-name');
    tournamentMetaEl = document.getElementById('tournament-meta');
    tournamentDescriptionEl = document.getElementById('tournament-description');
    participantListEl = document.getElementById('participant-list');
    participantLoadingEl = document.getElementById('participant-loading');
    participantEmptyEl = document.getElementById('participant-empty');
    loadingStateEl = document.getElementById('loading-state');
    errorBannerEl = document.getElementById('error-banner');
    retryBtn = document.getElementById('retry-btn');

    // Parse URL parameters
    const urlParams = new URLSearchParams(window.location.search);
    currentNamespace = urlParams.get('namespace');
    currentTournamentId = urlParams.get('id');

    if (!currentNamespace || !currentTournamentId) {
        showError('Missing namespace or tournament ID');
        return;
    }

    // Attach event listeners
    retryBtn.addEventListener('click', loadTournamentData);

    // Initial load
    loadTournamentData();
});

/**
 * Load tournament data and participants
 */
async function loadTournamentData() {
    showLoading();
    hideError();

    try {
        // Fetch tournament details
        const tournament = await fetchTournament(currentNamespace, currentTournamentId);
        renderTournament(tournament);

        // Fetch participants
        await loadParticipants();
    } catch (error) {
        showError('Failed to load tournament data');
        console.error('Failed to load tournament:', error);
    }
}

/**
 * Load and render participants
 */
async function loadParticipants() {
    showParticipantLoading();

    try {
        const participants = await fetchParticipants(currentNamespace, currentTournamentId);
        renderParticipants(participants);
    } catch (error) {
        hideParticipantLoading();
        console.error('Failed to load participants:', error);
        // Don't show error banner - just hide loading state
        // Participant error is less critical than tournament error
    }
}

/**
 * Render tournament information
 * @param {Object} tournament - Tournament object
 */
function renderTournament(tournament) {
    hideLoading();
    tournamentInfoEl.style.display = 'block';

    tournamentNameEl.textContent = tournament.name || 'Untitled Tournament';
    
    const participantCount = tournament.current_participants || 0;
    const maxParticipants = tournament.max_participants || 0;
    tournamentMetaEl.textContent = `${tournament.status} · ${participantCount}/${maxParticipants} participants`;
    
    tournamentDescriptionEl.textContent = tournament.description || '';
}

/**
 * Render participant list
 * @param {Array} participants - Array of participant objects
 */
function renderParticipants(participants) {
    hideParticipantLoading();

    if (!participants || participants.length === 0) {
        showParticipantEmpty();
        return;
    }

    hideParticipantEmpty();

    participantListEl.innerHTML = participants.map(participant => {
        const username = participant.username || participant.user_id || 'Unknown';
        return `<li>${escapeHtml(username)}</li>`;
    }).join('');
}

/**
 * Escape HTML to prevent XSS
 * @param {string} text - Text to escape
 * @returns {string} Escaped text
 */
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// UI state management functions
function showLoading() {
    loadingStateEl.style.display = 'block';
    tournamentInfoEl.style.display = 'none';
}

function hideLoading() {
    loadingStateEl.style.display = 'none';
}

function showError(message) {
    errorBannerEl.style.display = 'block';
    hideLoading();
}

function hideError() {
    errorBannerEl.style.display = 'none';
}

function showParticipantLoading() {
    participantLoadingEl.style.display = 'block';
    participantListEl.innerHTML = '';
    participantEmptyEl.style.display = 'none';
}

function hideParticipantLoading() {
    participantLoadingEl.style.display = 'none';
}

function showParticipantEmpty() {
    participantEmptyEl.style.display = 'block';
}

function hideParticipantEmpty() {
    participantEmptyEl.style.display = 'none';
}
