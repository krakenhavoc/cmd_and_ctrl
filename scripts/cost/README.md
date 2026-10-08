# Project cost tracker

Three answers to "what did this cost to build?", refreshed weekly in the
**Project cost tracker** GitHub Discussion:

1. **Human developers.** What a team would have charged to write this codebase.
2. **Claude API list price.** The tokens we actually used, priced per token.
3. **Subscriptions.** What we actually paid.

## Adding your usage (each of us, every few weeks)

```bash
make cost-push
```

This reads your local Claude Code transcripts (`~/.claude/projects/**/*.jsonl`,
subagents included) and sums every response whose working directory was
this repo or one of its worktrees. It writes the per-day, per-model token
totals to `data/cost/claude-usage-<github-login>.json`, and commits that
file to the `cost-ledger` branch without touching your checkout.

**Run it at least once every 30 days.** Claude Code deletes transcripts
after `cleanupPeriodDays` (default 30), and a day that is gone from your
transcripts can only be kept if it is already in your ledger. You can also
raise `cleanupPeriodDays` in `~/.claude/settings.json`.

The ledger holds token counts only: no prompts, code or file names. The
repo is public, so the ledgers and the discussion are public too.

`make cost-usage` does the same without publishing. `make cost-report`
prints the full report locally.

## How each number is made

**Human developers** (`config.json` → `human`). Two independent estimates
that should land close together:

- *By component*: production lines of engine/server Go and client code
  divided by a per-developer-day rate; hours per catalog card; doc lines
  per day; plus a project overhead share (planning, review, QA, CI). Tests
  are not counted separately: the per-day rates cover writing the tests
  that go with the code.
- *By merged PR*: developer-days per merged PR.

Each comes in low, mid and high scenarios. The headline is mid. The report
also gives the textbook COCOMO figure, for reference only: it counts lines
and assumes a traditional team, so it overstates AI-written code.

**Claude API list price** (`config.json` → `pricing_per_mtok`). Every
response's input, output, cache-write (5-minute and 1-hour) and cache-read
tokens are priced at Anthropic's first-party list price for its model. Fast
mode is priced at 2x. A person with no ledger, and any PR merged before a
person's ledger starts, is estimated at the measured cost per merged PR.
It is an estimate, and it is replaced by measured data as ledgers fill in.

**Subscriptions** (`config.json` → `people[].plans`). Months on each plan
multiplied by its monthly price, through the current month. Edit your own
plan history when it changes.

## Files

| File | Does |
|---|---|
| `claude_usage.py` | Transcripts → your ledger; `--push` publishes it |
| `cost_report.py` | Repo + GitHub + ledgers + config → the report |
| `post_discussion.py` | Creates or updates the discussion; `--comment` adds the weekly change |
| `config.json` | Prices, plans and every human-estimate assumption |
| `../../.github/workflows/cost-report.yml` | Mondays 12:00 UTC, and on demand |
