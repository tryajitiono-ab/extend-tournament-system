// Bracket Adapter - Data transformation layer for brackets-viewer.js
// Transforms REST API match data to brackets-model format
// 
// Dependencies: 
// - Match[] from REST API (/v1/public/namespace/{namespace}/tournaments/{tournament_id}/matches)
// - Participant[] from REST API (/v1/public/namespace/{namespace}/tournaments/{tournament_id}/participants)
// - Tournament from REST API (/v1/public/namespace/{namespace}/tournaments/{tournament_id})

/**
 * Transform REST API match data to brackets-viewer.js format
 * 
 * Converts tournament match data from our protobuf-based REST API format
 * into the brackets-model data structure required by brackets-viewer.js.
 * 
 * Key transformations:
 * - Status enum: MATCH_STATUS_* strings → numeric codes (2=Pending, 3=Running, 4=Completed, 5=Archived)
 * - Round indexing: 1-based API rounds → 0-based brackets-model rounds
 * - Participant mapping: user_id as ID, username with user_id fallback for name
 * - Match structure: participant1/participant2 → opponent1/opponent2 with position calculation
 * 
 * @param {Array} matches - Array of Match objects from REST API
 * @param {Array} participants - Array of Participant objects from REST API
 * @param {Object} tournament - Tournament object from REST API
 * @returns {Object} Data in brackets-model format with stages, matches, participants, matchGames
 */
function transformToBracketsModel(matches, participants, tournament) {
    // Handle empty matches array gracefully (tournament not started)
    if (!matches || matches.length === 0) {
        return {
            stages: [],
            matches: [],
            participants: [],
            matchGames: [],
        };
    }

    // Validate required parameters
    if (!participants) {
        console.warn('transformToBracketsModel: participants array is null/undefined, using empty array');
        participants = [];
    }
    if (!tournament) {
        console.warn('transformToBracketsModel: tournament object is null/undefined');
        tournament = { tournament_id: 'unknown', name: 'Unknown Tournament' };
    }

    // Create participant lookup map for efficient access
    const participantMap = new Map();
    participants.forEach(p => {
        const userId = p.userId || p.user_id || p.userid;
        if (userId) {
            participantMap.set(userId, p);
        }
    });

    // Transform matches to brackets-model format
    const transformedMatches = matches.map((match, index) => {
        // Convert opponent data with null handling for BYE matches and unknown participants
        // Handle both object format {userid: "player1"} and string format "player1"
        const getParticipantId = (participant) => {
            if (!participant) return null;
            if (typeof participant === 'string') return participant;
            return participant.userid || participant.userId || participant.user_id || null;
        };
        
        const participant1Id = getParticipantId(match.participant1);
        const participant2Id = getParticipantId(match.participant2);
        
        const opponent1 = participant1Id ? {
            id: participant1Id,
        } : null;

        const opponent2 = participant2Id ? {
            id: participant2Id,
        } : null;

        // Determine winner result for completed matches
        // The brackets-viewer library uses 'win' string on the winning opponent
        let result = null;
        if (match.status === 'MATCH_STATUS_COMPLETED' && match.winner) {
            if (participant1Id === match.winner) {
                result = 'win'; // opponent1 won
            } else if (participant2Id === match.winner) {
                result = 'win'; // opponent2 won, but we need to swap the result format
            }
        }

        // Brackets-viewer expects result field with 'win' for opponent1 or 'loss' for opponent1 (opponent2 won)
        // Simpler approach: set score/result per opponent
        if (opponent1 && match.winner === participant1Id) {
            opponent1.result = 'win';
        }
        if (opponent2 && match.winner === participant2Id) {
            opponent2.result = 'win';
        }

        return {
            id: index, // Use array index as numeric ID since match_id is a string
            stage_id: 0,
            group_id: 0,
            round_id: match.round - 1, // Convert 1-based API rounds to 0-based brackets-model rounds
            number: match.position + 1, // Convert 0-based position to 1-based match number
            opponent1: opponent1,
            opponent2: opponent2,
            status: mapMatchStatus(match.status),
        };
    });

    // Transform participants to brackets-model format
    const transformedParticipants = participants.map(p => {
        const userId = p.userId || p.user_id || p.userid;
        const userName = p.username || p.user_name || p.userName || userId;
        return {
            id: userId,
            tournament_id: tournament.tournamentId || tournament.tournament_id,
            name: userName
        };
    });
    
    // Add a special TBD participant for empty slots (instead of showing "BYE")
    transformedParticipants.push({
        id: null,
        tournament_id: tournament.tournamentId || tournament.tournament_id,
        name: 'TBD'
    });

    // Create single stage for single-elimination tournament
    const stage = {
        id: 0,
        tournament_id: tournament.tournamentId || tournament.tournament_id,
        name: tournament.name || 'Tournament',
        type: 'single_elimination',
        number: 1,
        settings: {
            seedOrdering: ['natural'],
            grandFinal: 'simple',
            skipFirstRound: false,
            consolationFinal: false,
            matchesChildCount: 0,
        }
    };

    return {
        stages: [stage],
        matches: transformedMatches,
        participants: transformedParticipants,
        matchGames: [], // Not used for basic bracket display
    };
}

/**
 * Map REST API match status to brackets-viewer numeric status code
 * 
 * Status mapping:
 * - MATCH_STATUS_SCHEDULED → 2 (Pending)
 * - MATCH_STATUS_IN_PROGRESS → 3 (Running)
 * - MATCH_STATUS_COMPLETED → 4 (Completed)
 * - MATCH_STATUS_CANCELLED → 5 (Archived)
 * 
 * @param {string} apiStatus - Match status from REST API (protobuf enum string)
 * @returns {number} Numeric status code for brackets-viewer.js
 */
function mapMatchStatus(apiStatus) {
    // Map protobuf MatchStatus enum strings to brackets-viewer numeric status codes
    const statusMap = {
        'MATCH_STATUS_SCHEDULED': 2,    // Pending
        'MATCH_STATUS_IN_PROGRESS': 3,  // Running
        'MATCH_STATUS_COMPLETED': 4,    // Completed
        'MATCH_STATUS_CANCELLED': 5,    // Archived
    };

    const mappedStatus = statusMap[apiStatus];
    
    if (mappedStatus === undefined) {
        console.warn(`Unknown match status: ${apiStatus}, defaulting to SCHEDULED (2)`);
        return 2; // Default to Pending for unknown statuses
    }

    return mappedStatus;
}

/**
 * Validate brackets-model data structure before rendering
 * 
 * Checks for common issues that can cause rendering failures:
 * - Missing required fields
 * - Invalid data types
 * - Inconsistent participant references
 * 
 * @param {Object} data - Transformed brackets-model data
 * @returns {boolean} True if data is valid, false otherwise
 */
function validateBracketsData(data) {
    if (!data) {
        console.error('validateBracketsData: data is null/undefined');
        return false;
    }

    if (!Array.isArray(data.stages) || data.stages.length === 0) {
        console.error('validateBracketsData: stages array is empty or invalid');
        return false;
    }

    if (!Array.isArray(data.matches) || data.matches.length === 0) {
        console.error('validateBracketsData: matches array is empty or invalid');
        return false;
    }

    if (!Array.isArray(data.participants)) {
        console.error('validateBracketsData: participants array is invalid');
        return false;
    }

    // Verify all matches have valid round_id (0-indexed)
    const invalidMatches = data.matches.filter(m => m.round_id < 0);
    if (invalidMatches.length > 0) {
        console.error('validateBracketsData: found matches with negative round_id', invalidMatches);
        return false;
    }

    return true;
}
