# Workflow preferences

- **Language**: GitHub issues, commits, PRs, and code comments must always be in
  **English** (titles and bodies), regardless of the language used in chat.
- Always propose the solution and discuss it with the user BEFORE writing any
  code. Do not implement a chosen technical direction on your own initiative.
- After changes, restart the stack with `make down` and `make up` to test
  (force-recreating containers is not always sufficient).
- Run e2e/smoke tests ONLY against the isolated test stack: `make test-e2e`
  (`docker-compose.test.yml`, project `vaultlab-test`, separate DB `vaultlab_test`,
  ports 8081/5433/6380, Yahoo finance disabled). NEVER test against the dev/prod
  stack (`docker-compose.yml` / `make up`, DB `vaultlab`, port 8080): it holds real
  data and must stay clean.
- Delegate implementation to the dedicated subagents whenever the work fits
  their scope: `backend` for Go/Postgres/Redis/API, `frontend` for SvelteKit/
  TypeScript/Tailwind, `finanza` for financial/statistical analysis. Use
  `explore`/`general` for research or cross-cutting tasks. Do not implement
  domain work yourself when a dedicated subagent exists.
