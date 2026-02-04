# Design Review Skill

Review the tournament bracket UI for visual design quality, consistency, and distinctive aesthetics (avoiding generic "AI-generated" look).

## Instructions

Perform a comprehensive design review of the web UI, focusing on making the bracket visualization distinctive and polished.

### 1. Visual Hierarchy Assessment

Read `web/templates/*.html` and `web/static/styles.css`:

- **Tournament list page:** Are tournaments easy to scan? Clear visual priority?
- **Bracket view:** Does the bracket draw attention? Are matches clearly grouped by round?
- **Empty states:** How does it look with 0 tournaments? 0 participants? BYE matches?
- **Information density:** Too cluttered or too sparse?
- **Sizing relationships:** Proper visual rhythm between elements?

### 2. Typography Review

Check CSS for font choices and usage:

- **Font selection:** Are we using distinctive, appropriate fonts or defaulting to system fonts?
  - Avoid: Generic sans (Arial, Helvetica) unless intentional
  - Consider: Sport/esports-appropriate typefaces for tournament context
- **Hierarchy:** Clear size/weight distinction between headings, body, labels?
- **Readability:** Line height, letter spacing appropriate for sport/competition context?
- **Responsive scaling:** Font sizes scale smoothly across viewports?
- **Special cases:** How does BYE text stand out from participant names?

### 3. Color & Theme Consistency

Analyze color usage:

- **CSS variables:** Are colors defined as variables for consistency?
- **Color roles:** Clear semantic meaning (primary, danger, success, neutral)?
- **Contrast:** Sufficient contrast ratios (WCAG AA minimum)?
- **Bracket-specific:** Match status colors (scheduled, in-progress, completed) clearly differentiated?
- **BYE styling:** Does BYE participant have distinct, appropriate styling?
- **Atmosphere:** Is the palette distinctive or generic? Does it fit a tournament/competition theme?

### 4. Spacing & Layout

Check layout consistency:

- **Spacing system:** Consistent spacing units (8px/16px grid or similar)?
- **Bracket spacing:** Progressive spacing for mobile mentioned in commits - is it implemented well?
- **Alignment:** Elements properly aligned within bracket structure?
- **White space:** Adequate breathing room or cramped?
- **Responsive behavior:** Graceful degradation on mobile vs desktop?

### 5. Interactive Elements

Review buttons, links, and interactive components:

- **Touch targets:** Minimum 44px for mobile tap targets?
- **Hover states:** Clear affordance for interactive elements?
- **Focus states:** Keyboard navigation visible and styled?
- **Loading states:** How do matches/brackets appear while loading?
- **Empty matches:** Clear visual treatment for TBD participants?

### 6. Motion & Animation (if present)

Check for animations in CSS or JavaScript:

- **Purpose:** Do animations serve function or just decoration?
- **Performance:** CSS animations preferred over JavaScript?
- **Subtlety:** Animations enhance without annoying?
- **Reduce motion:** Respect `prefers-reduced-motion`?

### 7. Bracket Visualization Quality

Specific to tournament brackets:

- **Lines/connectors:** Clean, crisp rendering of match connections?
- **Round labels:** Clear indication of bracket rounds (Round 1, Semifinals, Finals)?
- **Match numbering:** Logical and visible?
- **Participant alignment:** Names aligned consistently, no overlap?
- **Scalability:** Works for 2, 8, 16, 32+ participants?
- **BYE display:** Visually distinct from regular matches, clear meaning?

### 8. Distinctiveness Assessment

Evaluate against "distributional convergence" (generic AI look):

- **Personality:** Does the UI have character appropriate for tournament management?
- **Generic indicators:**
  - Flat, stark white backgrounds?
  - Overly rounded corners everywhere?
  - Generic blue primary color (#007bff)?
  - Lack of visual interest or depth?
- **Improvement opportunities:**
  - Add subtle backgrounds, gradients, or textures?
  - Use competitive/sport-appropriate visual language?
  - Distinctive iconography for match status?
  - Custom bracket rendering (not just boxes and lines)?

### 9. Report Format

Provide feedback as:

```
## Visual Hierarchy
- Observations
- Suggestions for improvement

## Typography
- Current state
- Recommendations (specific font suggestions if needed)

## Color & Theme
- Assessment
- Specific color palette suggestions

## Spacing & Layout
- What's working
- What needs adjustment

## Interactive Elements
- Usability findings
- Enhancement suggestions

## Bracket Visualization
- Quality assessment
- Specific bracket rendering improvements

## Distinctiveness Score: [1-10]
- Score explanation
- Top 3 actions to improve personality and avoid generic look

## Priority Fixes
1. Most impactful design improvement
2. Second priority
3. Third priority
```

## Available Tools
Read, Glob, Grep

## Output
A comprehensive design assessment with actionable recommendations prioritized by impact.
