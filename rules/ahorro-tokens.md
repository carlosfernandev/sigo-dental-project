# Token discipline

Reduce context cost without weakening verification.

## Browser and MCP tools

- Never take screenshots or dump full page HTML/DOM.
- Extract only the visible text or the specific CSS selector under test.
- To verify an API or JSON response, issue a direct HTTP request instead of opening a browser.

## Memory

Save to persistent memory only:

1. Root cause and final fix of recurring environment or configuration failures.
2. Local ports and key project variables.
3. Decisions, conventions, and non-obvious discoveries.

Never store full source files, stack traces, or chat transcripts as memory content.

## Failures and tests

- While iterating on a failure, run only the failing test.
- Before declaring a work unit done, run the full suite once. Skipping it is how a regression
  reaches the branch.
- If the same error survives two fix attempts, stop editing and state the hypothesis. More edits
  without a hypothesis convert a bug into a refactor.
