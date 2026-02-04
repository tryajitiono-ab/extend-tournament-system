// Tournament List Page Logic

// DOM elements
let tournamentListEl;
let loadingStateEl;
let emptyStateEl;
let errorBannerEl;
let refreshBtn;
let retryBtn;

// Initialize page
document.addEventListener('DOMContentLoaded', () => {
    tournamentListEl = document.getElementById('tournament-list');
    loadingStateEl = document.getElementById('loading-state');
    emptyStateEl = document.getElementById('empty-state');
    errorBannerEl = document.getElementById('error-banner');
    refreshBtn = document.getElementById('refresh-btn');
    retryBtn = document.getElementById('retry-btn');

    // Attach event listeners
    refreshBtn.addEventListener('click', loadTournaments);
    retryBtn.addEventListener('click', loadTournaments);

    // Initial load
    loadTournaments();
});

/**
 * Load and render tournaments
 */
async function loadTournaments() {
    showLoading();
    hideError();

    try {
        const tournaments = await fetchTournaments();
        renderTournaments(tournaments);
    } catch (error) {
        showError();
        console.error('Failed to load tournaments:', error);
    }
}

/**
 * Render tournament cards
 * @param {Array} tournaments - Array of tournament objects
 */
function renderTournaments(tournaments) {
    hideLoading();

    if (!tournaments || tournaments.length === 0) {
        showEmpty();
        return;
    }

    hideEmpty();

    tournamentListEl.innerHTML = tournaments.map(tournament => {
        const detailUrl = `/tournament?namespace=test-ns&id=${encodeURIComponent(tournament.tournamentId)}`;
        const statusBadge = formatStatusBadge(tournament.status);
        const participantInfo = formatParticipantCount(tournament.currentParticipants || 0, tournament.maxParticipants);

        return `
            <article onclick="window.location.href='${detailUrl}'">
                <header>
                    <h3>${escapeHtml(tournament.name)}</h3>
                    ${statusBadge}
                </header>
                <p>${escapeHtml(tournament.description || 'No description provided')}</p>
                <footer>
                    ${participantInfo}
                </footer>
            </article>
        `;
    }).join('');
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
 * Format participant count with icon
 * @param {number} current - Current participants
 * @param {number} max - Maximum participants
 * @returns {string} HTML for participant count
 */
function formatParticipantCount(current, max) {
    const percentage = max > 0 ? Math.round((current / max) * 100) : 0;
    const isFull = current >= max;

    return `
        <div class="participant-count">
            <span>👥</span>
            <span><strong>${current}</strong> / ${max} participants ${isFull ? '(Full)' : `(${percentage}%)`}</span>
        </div>
    `;
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
    tournamentListEl.innerHTML = '';
    emptyStateEl.style.display = 'none';
}

function hideLoading() {
    loadingStateEl.style.display = 'none';
}

function showError() {
    errorBannerEl.style.display = 'block';
    hideLoading();
}

function hideError() {
    errorBannerEl.style.display = 'none';
}

function showEmpty() {
    emptyStateEl.style.display = 'block';
}

function hideEmpty() {
    emptyStateEl.style.display = 'none';
}
