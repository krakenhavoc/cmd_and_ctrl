# ADR 0184 — Weighted model budget and a share per table

Date: 2026-10-14
Sprint: S70 — Public beta: a table strangers can find
Status: Accepted. Amends [ADR 0033](0033-ai-bot-seat.md) (the spend guard of #2980). Issue
[#3476](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3476), from #3341.

## Context

The hosted-model guard summed input, output and cache tokens at equal weight, and kept one server-wide
pool. Two things were wrong with that. A million output tokens of the strongest model and a million
cache reads of the cheapest cost a hundred times apart, so the cap did not track spend. And in a public
beta one table could use the whole pool and take the model tiers away from everyone else.

## Decision

1. **Count weighted tokens.** Each response's usage is priced by the model that answered
   (`Response.Model`, else `Request.Model`): input, output, cache read and cache write tokens are each
   multiplied by that model's list price per million tokens. The unit is therefore the micro-dollar at
   list price. Defaults (US dollars per million, input / output / cache read / cache write), matched
   by case-insensitive substring of the model id: Haiku `1 / 5 / 0.10 / 1.25`, Sonnet
   `3 / 15 / 0.30 / 3.75`, Opus `5 / 25 / 0.50 / 6.25`. An unrecognised model is priced as Opus: a cap
   that under-counts fails in the worse direction. `CMDCTRL_BOT_MODEL_PRICES` overrides or extends the
   table. It is a weight table, not an invoice; list prices move and the table is the thing to refresh.
2. **Keep the names.** The metrics `cmdctrl_bot_budget_tokens` / `_limit_tokens` / `_blocked`, the
   `daily_tokens` / `monthly_tokens` fields and the `CMDCTRL_BOT_MODEL_*_TOKENS` variables keep their
   names. Both sides of the alert ratio are in the same unit, so `BotModelSpendHigh` and
   `BotModelBlocked` need no new expression, only wording and docs. The default caps are restated in
   the new unit: 5,000,000 a day (about $5) and 50,000,000 a month (about $50), the same order as the
   1M / 10M plain tokens they replace. A counter persisted before the upgrade is read in the new unit
   for the rest of that day, which is harmless.
3. **A share per table.** A table (the game id) may use `CMDCTRL_BOT_MODEL_TABLE_SHARE_PERCENT` of the
   daily cap in a UTC day; the default is 25, and 0 or 100 turns it off. Past it that table's model
   seats receive `model.ErrBudgetExceeded` and play the heuristic, the same fallback as a cap. Other
   tables are untouched and the picker stays open: a table at its share is not a server state, so
   `cmdctrl_bot_budget_blocked` does not move. The readout gains `table_limit` and `tables_at_share`.
   The share is on the daily window only; the monthly pool bounds the rest.
4. **Per-table counts are in memory.** They reset at the UTC day and on a restart. A restart gives a
   table a fresh share, but the persisted pool counters still bound the sum, which is what protects the
   bill. Persisting them would add a table to the schema for little.
5. **Own-key seats stay out.** A seat that runs on its owner's key (ADR 0182) never gets the guarded
   client: `ForTable` binds only a client that `Wrap` produced, and the factory's own-key path does not
   call it. Their calls count in neither the pool nor a share.

## Consequences

- A cheap model at a busy table costs a fifth of what the old count said; a strong model costs more
  than it did. The default caps, read as dollars, are easier to reason about.
- The table's share is enforced per call, as the pool is, so calls in flight can overrun it by a few
  replies.
- Self-hosted models still get no caps (`budget.Limits{}`), which also means no share.

Branches checked for the number: every `origin/*` head and `pr/2326`; the highest used was 0183.
