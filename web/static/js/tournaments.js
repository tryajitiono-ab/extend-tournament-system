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
        
        return `
            <article>
                <header>
                    <h3><a href="${detailUrl}">${escapeHtml(tournament.name)}</a></h3>
                </header>
                <p>${escapeHtml(tournament.description || '')}</p>
                <footer>
                    <small>${tournament.status} · ${tournament.currentParticipants || 0}/${tournament.maxParticipants} participants</small>
                </footer>
            </article>
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
