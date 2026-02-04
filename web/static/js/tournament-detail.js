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
let bracketSectionEl;
let bracketLoadingEl;
let bracketErrorEl;
let bracketContainerEl;
let bracketMobileWarningEl;

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
    bracketSectionEl = document.getElementById('bracket-section');
    bracketLoadingEl = document.getElementById('bracket-loading');
    bracketErrorEl = document.getElementById('bracket-error');
    bracketContainerEl = document.getElementById('bracket-container');
    bracketMobileWarningEl = document.getElementById('bracket-mobile-warning');

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
        
        // Load bracket after participants
        await loadBracket();
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
        showParticipantEmpty();
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

    // Tournament header
    tournamentNameEl.textContent = tournament.name || 'Untitled Tournament';
    tournamentDescriptionEl.textContent = tournament.description || 'No description provided';

    // Status badge
    const statusEl = document.getElementById('tournament-status');
    statusEl.innerHTML = formatStatusBadge(tournament.status);

    // Participant count
    const participantCount = tournament.currentParticipants || 0;
    const maxParticipants = tournament.maxParticipants || 0;
    const participantsEl = document.getElementById('tournament-participants');
    participantsEl.textContent = `${participantCount} / ${maxParticipants}`;

    // Created date
    const createdEl = document.getElementById('tournament-created');
    const createdDate = new Date(tournament.createdAt);
    createdEl.textContent = createdDate.toLocaleDateString();

    // Store tournament for later use
    window.currentTournament = tournament;
}

/**
 * Show tournament winner banner
 * @param {string} winnerUserId - Winner user ID
 * @param {Array} participants - Array of participant objects
 */
function showWinnerBanner(winnerUserId, participants) {
    // Check if banner already exists
    let banner = document.getElementById('tournament-winner-banner');
    if (banner) {
        banner.remove();
    }

    // Find winner's username from participants
    let winnerName = winnerUserId;
    if (participants && participants.length > 0) {
        const winner = participants.find(p => p.userId === winnerUserId || p.user_id === winnerUserId);
        if (winner) {
            winnerName = winner.username || winner.userName || winner.userId || winner.user_id || winnerUserId;
        }
    }

    // Create winner banner
    banner = document.createElement('div');
    banner.id = 'tournament-winner-banner';
    banner.className = 'tournament-winner-banner';
    banner.innerHTML = `
        <span class="trophy">🏆</span>
        <h2>Tournament Winner</h2>
        <div class="winner-name">${escapeHtml(winnerName)}</div>
    `;

    // Insert after tournament header
    const headerEl = document.querySelector('.tournament-header');
    headerEl.insertAdjacentElement('afterend', banner);
}

/**
 * Format status as colored badge
 * @param {string} status - Tournament status enum
 * @returns {string} HTML for status badge
 */
function formatStatusBadge(status) {
    const statusMap = {
        'TOURNAMENT_STATUS_DRAFT': { text: 'Draft', class: 'draft' },
        'TOURNAMENT_STATUS_ACTIVE': { text: 'Active', class: 'active' },
        'TOURNAMENT_STATUS_STARTED': { text: 'Started', class: 'started' },
        'TOURNAMENT_STATUS_COMPLETED': { text: 'Completed', class: 'completed' },
        'TOURNAMENT_STATUS_CANCELLED': { text: 'Cancelled', class: 'cancelled' }
    };

    const statusInfo = statusMap[status] || { text: 'Unknown', class: 'draft' };
    return `<span class="status-badge ${statusInfo.class}">${statusInfo.text}</span>`;
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

    participantListEl.innerHTML = participants.map((participant, index) => {
        const username = participant.username || participant.userName || participant.userId || participant.user_id || 'Unknown';
        const userId = participant.userId || participant.user_id || '';
        const initial = username.charAt(0).toUpperCase();

        return `
            <div class="participant-card">
                <div class="participant-avatar">${initial}</div>
                <div class="participant-info">
                    <strong>${escapeHtml(username)}</strong>
                    <small>${escapeHtml(userId)}</small>
                </div>
            </div>
        `;
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

/**
 * Load and render bracket
 */
async function loadBracket() {
    showBracketSection();
    showBracketLoading();
    
    try {
        // Fetch all required data
        const matchData = await fetchMatches(currentNamespace, currentTournamentId);
        const participantData = await fetchParticipants(currentNamespace, currentTournamentId);
        
        // Re-fetch tournament for latest data (may have changed)
        const tournament = await fetchTournament(currentNamespace, currentTournamentId);
        
        // Check if tournament has started
        if (!matchData.matches || matchData.matches.length === 0) {
            showBracketError('Bracket not yet generated. Tournament must be started to view bracket.');
            return;
        }
        
        // Transform to brackets-model format
        console.log('Match data:', matchData);
        console.log('Participant data:', participantData);
        console.log('Tournament:', tournament);
        
        const bracketData = transformToBracketsModel(
            matchData.matches,
            participantData,
            tournament
        );
        
        console.log('Transformed bracket data:', bracketData);
        
        // Show mobile warning for large tournaments
        const participantCount = participantData.length;
        if (participantCount >= 32 && window.innerWidth < 768) {
            showMobileWarning();
        }
        
        // Render bracket
        renderBracket(bracketData);

        // Show winner banner if tournament is completed
        if (tournament.status === 'TOURNAMENT_STATUS_COMPLETED') {
            const winner = findTournamentWinner(matchData.matches);
            if (winner) {
                showWinnerBanner(winner, participantData);
            }
        }

    } catch (error) {
        // Non-critical error - don't show error banner at top
        showBracketError('Failed to load bracket');
        console.error('Bracket loading error:', error);
    }
}

/**
 * Find tournament winner from matches (highest round winner)
 * @param {Array} matches - Array of match objects
 * @returns {string|null} Winner user ID
 */
function findTournamentWinner(matches) {
    if (!matches || matches.length === 0) return null;

    // Find the highest round number
    const maxRound = Math.max(...matches.map(m => m.round));

    // Find the completed match in the highest round
    const finalMatch = matches.find(m => m.round === maxRound && m.status === 'MATCH_STATUS_COMPLETED');

    return finalMatch ? finalMatch.winner : null;
}

/**
 * Render bracket using brackets-viewer.js
 * @param {Object} data - Bracket data in brackets-model format
 */
function renderBracket(data) {
    console.log('=== RENDER BRACKET DEBUG ===');
    console.log('Bracket data:', JSON.stringify(data, null, 2));
    console.log('Stages count:', data.stages ? data.stages.length : 0);
    console.log('Matches count:', data.matches ? data.matches.length : 0);
    console.log('Participants count:', data.participants ? data.participants.length : 0);
    
    // Check if brackets-viewer library is loaded
    if (typeof window.bracketsViewer === 'undefined') {
        const errorMsg = 'brackets-viewer.js library failed to load. Please refresh the page.';
        console.error(errorMsg);
        showBracketError(errorMsg);
        return;
    }
    
    console.log('bracketsViewer is loaded:', typeof window.bracketsViewer);
    console.log('bracketsViewer.render:', typeof window.bracketsViewer.render);
    
    // Check if container exists
    const container = document.querySelector('#bracket-container');
    console.log('Container found:', !!container);
    console.log('Container classes:', container ? container.className : 'N/A');
    console.log('Container display:', container ? container.style.display : 'N/A');
    
    showBracket();
    
    // Render using brackets-viewer.js
    try {
        const renderData = {
            stages: data.stages,
            matches: data.matches,
            matchGames: data.matchGames || [],
            participants: data.participants
        };
        
        console.log('Calling bracketsViewer.render with:', JSON.stringify(renderData, null, 2));
        
        window.bracketsViewer.render(
            renderData,
            {
                selector: '#bracket-container',
                clear: true,  // Always clear previous render
            }
        );
        
        console.log('Bracket rendered successfully');
        console.log('Container innerHTML length after render:', container ? container.innerHTML.length : 0);
        console.log('Container children after render:', container ? container.children.length : 0);
    } catch (error) {
        console.error('Bracket rendering error:', error);
        console.error('Error stack:', error.stack);
        showBracketError('Bracket rendering failed: ' + error.message);
    }
}

// Bracket UI state management functions
function showBracketSection() {
    bracketSectionEl.style.display = 'block';
}

function hideBracketSection() {
    bracketSectionEl.style.display = 'none';
}

function showBracketLoading() {
    bracketLoadingEl.style.display = 'block';
    bracketErrorEl.style.display = 'none';
    bracketContainerEl.style.display = 'none';
}

function showBracketError(message) {
    bracketLoadingEl.style.display = 'none';
    bracketErrorEl.style.display = 'block';
    bracketErrorEl.textContent = message;
    bracketContainerEl.style.display = 'none';
}

function showBracket() {
    bracketLoadingEl.style.display = 'none';
    bracketErrorEl.style.display = 'none';
    bracketContainerEl.style.display = 'block';
}

function showMobileWarning() {
    bracketMobileWarningEl.style.display = 'block';
}
