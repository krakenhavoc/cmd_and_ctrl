# ADR 0182 — Your own model key for AI seats

**Status:** Accepted · 2026-10-14 · Backlog
**Issue:** [#3460](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3460)
**Numbering:** every ADR and migration file added on any remote head was listed (`git log --remotes --name-only --format= --diff-filter=A -- server/internal/db/migrations docs/decisions`); the highest migration was **0024**, so this is migration **0025**. It was drafted as ADR 0180; before pushing, 0180 (`0180-secret-council.md`) and 0181 (`0181-deploy-credentials-fail-closed.md`) had appeared on other heads, so it is **0182**.
**Builds on:** [ADR 0033](0033-ai-bot-seat.md) (the AI seat: the model selects from a closed list, the table never waits on a bot), [ADR 0051](0051-user-database.md) decision 5 (sealing: AES-256-GCM, a version byte, the key from `CMDCTRL_IDENTITY_KEY`), [ADR 0128](0128-playmats.md) §3 (the SSRF guard on a link a person pastes), [ADR 0149](0149-privacy-notice-and-account-deletion.md) (the privacy notice and the deletion inventory that cannot drift), #2980 (the server's own spend caps).

---

## Context

The `assisted` and `strong` bots need a model. Until now the only model was the server's: an operator sets `CMDCTRL_ANTHROPIC_API_KEY` or points `CMDCTRL_OPENAI_ENDPOINT` at a local server, and every model seat on the deployment shares it under the #2980 caps. A server without one offers `random` and `heuristic` only.

A player who has a model API key of their own can now bring it: the AI seats **they** add run on **their** key and their account with the provider. That makes the stronger bots available on any server, and puts the cost and the choice of model with the person who wants them.

A key is a credential, and a configurable endpoint is a way for a stranger to make the server send requests somewhere. Both are the substance of this ADR.

## Decision

### 1. Settings › AI opponents: add, test, replace, remove

A signed-in person gets an **AI opponents** tab in Settings (`ModelKeySettings.svelte`), shown only when the server says the feature is on. It holds one key per person:

- **Provider:** Anthropic, OpenAI, OpenRouter, Groq, Together AI, Mistral, DeepSeek, or "Other (OpenAI-compatible)" with an https address (§3).
- **Model:** the model id the seats ask for by default; each seat may name another when it is added.
- **API key:** write-only. The field starts empty, is cleared after every save, and the key never comes back from the server.
- **Limits** the owner sets (§5): games a day (1–100, default 10) and calls per game (1–5000, default 500).
- **Test key** makes one small call (`"Reply with the single word: ok"`, 16 tokens) and shows a fixed sentence: it works, or why not.
- **Remove key** deletes the row.

The tab shows the provider, the model, the **last four characters**, when the key was added and last used, how many games used it today, and the **last error**, as one of the server's own sentences.

Routes (`lobby/modelkeys_http.go`): `GET`, `PUT`, `DELETE /me/model-key` and `POST /me/model-key/test`, signed-in only (403 otherwise), 503 with `"available": false` while the feature is off. Writes share a 1/s bucket per person; the test, a real request, has a slower one (one per ten seconds, burst 3).

### 2. The key: sealed, never returned, never logged

- **Sealed at rest** with the ADR 0051 `users.Sealer` (AES-256-GCM, version byte 1, fresh nonce), under `CMDCTRL_IDENTITY_KEY`. The associated data is `"cmd_and_ctrl model key v1" ‖ user id ‖ provider`, distinct from the refresh tokens', so a sealed key copied onto another person's row, or kept across a change of provider, does not open. A key saved without a key typed in is only accepted when the provider and address are unchanged.
- **No sealing key → the feature is off.** `modelkeys.New` returns nil without a database or a sealer, every route answers 503, and the Settings tab and the picker's option do not appear. There is no plaintext fallback.
- **Never returned.** The HTTP layer only ever gets `modelkeys.View`, which has no field that could carry the key. A body that fails to decode is answered "invalid body", never the decoder's error, which can quote the body.
- **Never in a log, an error, a snapshot, a replay or a metric.** The key goes from the sealed row into the provider client and nowhere else. A provider's own error text can echo part of a key (`Incorrect API key provided: sk-…abcd`), so nothing a provider says is passed on: every failure is classified (`modelkeys/failures.go`) and becomes one of a fixed set of sentences, and the error the bot policy logs is rebuilt from the status code and that sentence. The seat client's `Name()` is `own-key`. The bot metrics are labelled by tier only. Tests feed a fake key through a provider that quotes it back and assert it appears in no error, log line, notice, stored row or HTTP answer (`TestARefusedKeyStopsTheSeatTellsTheOwnerAndLeaksNothing`, `TestModelKeyRoundTripNeverReturnsTheKey`).

### 3. Endpoint safety: the check is the dial

An OpenAI-compatible base URL is an SSRF vector into the host's network. The requirement is met by a new package, `internal/netguard`, which every own-key request goes through — the listed providers as well as "custom":

- **https only.** The transport refuses any other scheme before a request is sent (`capTransport.RoundTrip`).
- **Public unicast addresses only, checked at connect time.** `netguard.AddrAllowed` refuses loopback, private (RFC 1918 and IPv6 unique-local), link-local (the cloud metadata address among it), carrier-grade NAT (100.64/10), unspecified, multicast and the reserved and documentation ranges, in any IPv6 spelling (v4-mapped, NAT64, 6to4). The transport's `DialContext` resolves the name **once**, refuses the name if **any** answer is refused, and dials the checked address rather than the name; `net.Dialer.Control` checks the address the socket is actually connecting to as the last word. A name that is public when saved and private when dialled (DNS rebinding) is refused at the dial.
- **No redirects.** `CheckRedirect` returns the 3xx to the caller; nothing is followed. A redirect answer ends the seat's model use like any other failure that will not pass.
- **Short timeout and a response cap.** 15 seconds end to end (a seat's own deadline, the tier's think time, is shorter and wins); a 256 KiB body cap that fails the read rather than truncating.
- **No proxy from the environment, no cookies.**
- **Save-time checks are for the message, not the guard.** `ValidateCustomURL` refuses non-https, a user name or password, a query or fragment, a port other than 443, an IP-literal host, and names that are plainly local (`localhost`, `.local`, `.internal`, …), so the person is told at once. The dial check runs on every request regardless.

The listed providers are an allowlist in code (`modelkeys/providers.go`), each https on 443; "custom" is the escape hatch behind the same checks. Tests cover each refused class, a host name resolving to 127.0.0.1 and to 192.168.x, a mixed answer, an address literal, the Control hook on its own, plain http, a redirect, the body cap and a slow host (`netguard_test.go`), and the same refusals end to end through a seat and the Test button (`TestACustomAddressThatResolvesPrivatelyIsRefusedAtDial`). `playmat.AddrAllowed` now delegates to `netguard.AddrAllowed`, so the two features share one address policy.

### 4. Using the key at a table

When a signed-in person adds a bot, the picker offers **Run on my model key (provider)** if they have a key (`GET /bot/options` → `own_key`). Ticking it narrows the tier list to `assisted` and `strong` (available whether or not the server has a model of its own) and shows a model field, prefilled with the key's model. `POST /games/{id}/seats/bot` takes `own_key: true` and an optional `model`.

The seat records its key owner, **always the caller's own user id from their session** — the request has no field that could name anyone else's key — and the model. They ride `SeatInfo.BotKeyOwner` / `BotKeyModel` (never sent to clients; the public `bot_own_key` flag says only that the bot is on a player's own key), the seats row (`seats.bot_key_user_id`, `bot_key_model`, so a restart brings the seat back on the same key), and `aiseat.SeatSpec.KeyOwner` / `KeyModel`. A rematch keeps an own-key bot only when its owner is the one asking for the rematch. A saved table setup does not carry the key: it seats the bot only if the server offers the tier itself.

`aiseat/tiers.Factory` builds an own-key seat through `OwnKeyClients` (`*modelkeys.Service`), not through the server's client, and not behind the server's spend gate: these calls are on the owner's account. The seat client (`modelkeys/seat.go`) reads and opens the key **on the seat's first call**, on the seat's own goroutine — the Manager builds policies under its lock — and re-reads it when the owner saves or removes the key, so a replaced key is used from the next call and a removed one not at all.

### 5. Limits, failures and the table

- **Games a day** (UTC): a game is counted once, on the seat's first model call, by one SQL statement that cannot overrun the limit; a seat restarted by a deploy does not count its game again. The picker refuses a new own-key bot once the day is used up, with a sentence saying so.
- **Calls per game**: per seat, counted in the seat client.
- **A failure that will not pass** — 401 or 403 (key refused), 402 or a quota/credit answer (out of quota), 404 or `model_not_found` (no such model), an address the guard refuses, a redirect — ends that seat's model use for the game. Every later call answers `model.ErrBudgetExceeded` at once, which the model policy already treats as "play the heuristic's move, quietly" (#2980): Layer B's answer was computed first, so the table loses nothing but the call. The owner is told twice: the seat posts **one** chat line (a new optional policy extension, `aiseat.Noticer`, asked after every decision) — "I can't use my owner's model key any more, so I'm playing on CMD & CTRL's built-in heuristic for the rest of this game." plus the reason — and the reason is stored as the key's last error, shown on Settings. A limit reached posts the same line with its reason and is not recorded as an error.
- **A failure that will pass** (a rate limit, a 5xx, a timeout) costs that window and the next call tries again, exactly as for the server's own model.
- **Never stall the table.** An own-key seat is always buildable: with the feature off, the key removed, the account deleted or the sealing key changed, it is a complete policy that plays the heuristic under its tier's name. A seat whose `NewPolicy` failed would sit idle and stall the table (ADR 0033 §10).

### 6. Data, deletion and the privacy notice

Migration **0025** adds `model_keys` (one row per person: provider, base URL for "custom", model, sealed key, last four, the two limits, the day's count, created, updated, last used, last error and when) and the two nullable seat columns.

Account deletion (ADR 0149) **deletes** the person's `model_keys` row in its transaction (`accountdelete.Inventory`: Deleted), and tells the service to forget the person so a running seat stops using a key it had already opened. `seats.bot_key_user_id` is **Anonymised** like `seats.user_id`: it points at the tombstone, which carries no identity and has no key, so a restored seat plays the heuristic rather than sitting idle. `seats.bot_key_model` is Kept: a model id names no one.

The privacy notice gains a row, "Your own model key for AI opponents": the key is stored encrypted and used only for your seats, and each request goes to the provider you chose, under its terms, carrying the game the bot is looking at. `PRIVACY_SCHEMA_VERSION` is 25.

## Consequences

- Any server with a database and a sealing key can offer the model bots to people who bring a key, whatever the operator configured.
- The provider sees what the server's own model would see for that seat: the seat's filtered view of the game (ADR 0033 §3) and its move list. Nothing about the other players beyond what that seat can see leaves the server.
- OpenAI-compatible providers differ. The client omits the local-server-only `think` and `reasoning_effort` fields for an own key, and falls back from a `json_schema` response format a provider refuses (#2196). A model that wants `max_completion_tokens` instead of `max_tokens` will fail the Test button with a status sentence; that is the owner's to fix by choosing another model.
- A deployment with no sealing key shows nothing new.

## Not done

- Sharing a key with other players, and any charging for model seats, are out of scope (#3460).
- One key per person. A second provider means replacing the key.

## Back-out

Revert the change and roll the schema back by hand (docs/environments.md, migration 0025): drop `model_keys` and the two seat columns. A seat already on a key then comes back as an ordinary model-tier seat: on the server's own model where it has one. Where it has none, the older binary cannot build that seat, so end the active tables with an own-key bot before rolling back.
