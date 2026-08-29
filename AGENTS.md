# Workflow preferences

- **Language**: GitHub issues, commits, PRs, and code comments must always be in
  **English** (titles and bodies), regardless of the language used in chat.
- Always propose the solution and discuss it with the user BEFORE writing any
  code. Do not implement a chosen technical direction on your own initiative.
- After changes, restart the stack with `make down` and `make up` to test
  (force-recreating containers is not always sufficient).
- Delegate implementation to the dedicated subagents whenever the work fits
  their scope: `backend` for Go/Postgres/Redis/API, `frontend` for SvelteKit/
  TypeScript/Tailwind, `finanza` for financial/statistical analysis. Use
  `explore`/`general` for research or cross-cutting tasks. Do not implement
  domain work yourself when a dedicated subagent exists.
