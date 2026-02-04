# UI Lint Skill

Check web code quality and catch common HTML/CSS/JavaScript issues before they cause bugs.

## Instructions

When invoked, systematically analyze the web UI files and report issues:

### 1. JavaScript Analysis (`web/static/*.js`)
Read all JavaScript files and check for:
- **Console statements:** `console.log`, `console.error` left in code (should be removed or gated)
- **Null safety:** Missing null/undefined checks, especially for API response data
- **Participant data:** Check bracket.js for proper handling of BYE participants and empty slots
- **Magic numbers:** Hardcoded values that should be constants (spacing, dimensions)
- **Event listeners:** Check if listeners are properly removed to prevent memory leaks
- **Error handling:** API calls should have proper try/catch or error callbacks
- **Type coercion:** Loose equality (`==`) instead of strict (`===`)

### 2. CSS Analysis (`web/static/*.css`)
Read all CSS files and check for:
- **Unused selectors:** CSS rules that don't match any elements in templates
- **!important overuse:** More than 2-3 uses suggests specificity problems
- **Color consistency:** Hardcoded colors should use CSS variables (e.g., `--color-primary`)
- **Magic numbers:** Hardcoded spacing/sizes (use rem/em units or CSS variables)
- **Responsive issues:** Missing mobile breakpoints or inconsistent breakpoint usage
- **Vendor prefixes:** Check if modern flexbox/grid/transforms need prefixes
- **Z-index chaos:** More than 5 different z-index values suggests layering issues
- **Accessibility:** Insufficient color contrast, missing focus states

### 3. HTML Template Analysis (`web/templates/*.html`)
Read all template files and check for:
- **Semantic HTML:** Divs where semantic elements (article, section, nav) should be used
- **Accessibility:** Missing alt attributes, ARIA labels, keyboard navigation support
- **Script placement:** Scripts should be at end of body or async/defer
- **Inline styles:** Should use CSS classes instead
- **ID vs Class:** IDs used for styling instead of classes

### 4. Bracket-Specific Checks
Focus on common issues from recent commits:
- **BYE handling:** Check if bracket.js properly styles and displays BYE participants
- **Participant display:** Verify display_name vs username vs user_id extraction logic
- **Match numbering:** Check calculation logic for edge cases (odd participants)
- **Responsive spacing:** Verify progressive spacing adapts to viewport width
- **Touch targets:** Mobile elements should be min 44px for touch

### 5. Report Format

Present findings as:

```
## Critical Issues (breaks functionality)
- [file:line] description and suggested fix

## Warnings (potential bugs)
- [file:line] description and suggested fix

## Suggestions (code quality)
- [file:line] description and suggested fix

## Bracket-Specific Findings
- description of bracket visualization issues
```

## Available Tools
Read, Grep, Glob

## Output
A categorized report of code quality issues with file locations and actionable suggestions.
