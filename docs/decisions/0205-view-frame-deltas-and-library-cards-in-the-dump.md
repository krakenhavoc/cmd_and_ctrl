# ADR 0205 — View frames on the wire: compress before we diff, and library cards in the dump

**Status:** Proposed · 2026-10-20 · S73 — Big boards: layout and per-action cost
**Issue:** [#3922](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3922) (split out of [#3823](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3823))
**Numbering:** `git fetch --all --prune`, then every `docs/decisions/` file name in the history of all remote heads (`git log --remotes --name-only -- docs/decisions`) and the open PRs. The highest in use is **0204**. Reserved on #3922 before this file was written.
**Builds on:** [ADR 0044](0044-surviving-a-deploy.md) decision 5 (seq and generation), [ADR 0148](0148-capacity-limits-for-a-public-launch.md) (capacity), [ADR 0194](0194-snapshot-only-what-someone-can-undo.md) and [ADR 0041](0041-game-persistence.md)'s 2026-10-15 amendments (the per-commit cost work of #3823), [ADR 0122](0122-an-agent-at-the-table-a-local-mcp-seat.md) (the MCP seat reads the same wire), [ADR 0149](0149-privacy-notice-and-account-deletion.md) (what the replay holds).

This ADR was written plan-first. No code changed with it.

---

## Context

#3823 cut the per-commit cost of a table: #3880 the view's allocations and the per-frame log, #3884 one broadcast frame per viewer ID, #3887 rate-limited restore points that encode each event once. On the 5,000-event midgame board a whole commit went from 16.5 ms / 35 MB to 2.4 ms / 3.3 MB. Two optional items were left, and #3922 holds them:

1. **Wire deltas.** Send a seat what changed since its last frame instead of the whole view on every commit.
2. **Library card views in the crash dump.** Decide whether they are worth their bytes.

The issue says to pick these up "when a soak or prod profile shows view frames or dumps costing again". No such profile exists. This ADR records what the code does today, measures what it can, and recommends what to do first.

### What is on the wire today

- **One frame kind carries state: `snapshot`, a whole filtered view.** `SnapshotPayload` is `{seq, generation, game}` (`server/internal/protocol/protocol.go:322`). Its doc comment says S03 chose full frames on purpose and "diff-based delta frames can be added later without a protocol version bump" (`protocol.go:319-321`); `docs/protocol.md` §`snapshot` says the same: "new frame kinds can be added inside v0 additively".
- **A frame goes out on every commit, to every connection on the table.** `Hub.broadcastToRoom` (`server/internal/ws/hub.go:856`) filters the view per viewer ID, marshals it once per viewer (#3884) and queues the same bytes on each of that viewer's connections. It is reached from every action path (`hub.go:1264`, `1312`, `1421`), from HTTP-side mutations through `BroadcastState` (`hub.go:789`) and from presence changes (`presence.go:274`), so a seat going away and coming back is a whole frame to everyone too.
- **No compression.** `baseUpgrader` sets only buffer sizes (`hub.go:49-52`); `EnableCompression` is unset, so no permessage-deflate is negotiated. The write pump sends each frame as one `TextMessage` (`hub.go:1828`). gorilla/websocket is v1.5.3 (`server/go.mod`).
- **The send buffer is 16 frames, and a full one disconnects.** `sendRaw` (`hub.go:1780`) drops the connection rather than a frame, and the buffer is `make(chan []byte, 16)` (`hub.go:570`). A slow link on a bot table that commits in bursts reconnects rather than falling behind.
- **Readers of the wire.** The browser client (`client/src/lib/ws.ts`, the `snapshot` case at line 1027: drop an older `seq` within a generation, take any frame of a new generation), the MCP seat (`server/internal/mcpseat/seat.go:420`, the same rule) and `gamecli` (`server/cmd/gamecli/main.go:167`). The in-process bot seat does **not** read the wire: `aiseat` takes the view through `room.Subscribe()` (`server/internal/aiseat/runner.go:307`).
- **Reconnect is a whole frame by construction.** `ServeWS` stages a fresh filtered snapshot on the new connection before admitting it (`hub.go:584-616`), and ADR 0044 decision 5 makes a generation change mean "discard and re-render".
- **What is measured in production.** `cmdctrl_ws_frames_total` counts frames by direction and kind and `cmdctrl_ws_broadcast_seconds` times the broadcast (`server/internal/metrics/ws.go:119-124`). **No metric counts bytes.** The host dashboard has node-level network receive and transmit rates (`deploy/monitoring/dashboards/hosts.json`, `node_network_transmit_bytes_total`), which cannot separate WebSocket frames from images, the card cache and backups. The dev frame inspector records each frame's serialised size in the browser (`ws.ts:113`, ADR 0023), which is a manual tool, not a series.

### What a frame weighs

Measured for this ADR with a throwaway test (not committed) on `midgameTable` (`server/internal/protocol/view_midgame_bench_test.go`: four seats, 100-card decks, about 25 permanents a seat, 5,000 events), marshalling `SnapshotPayload` exactly as `marshalSnapshotFrame` does. Deflate is Go's `compress/flate` at the default level (6); gorilla's own default level is lower, so its numbers would be somewhat larger and are not measured here.

| frame | bytes | deflated | of which: `log` (200 entries) | `battlefield` (100 cards) | libraries | `legal_moves` |
|---|---|---|---|---|---|---|
| seat 1 | 70,672 | 9,026 (−87 %) | 22,330 | 19,044 | 9,123 | 2,820 (11 moves) |
| spectator | 56,650 | 7,110 (−87 %) | 22,330 | 19,044 | 348 | 4 |
| admin (`""`) | 73,463 | 8,024 (−89 %) | 22,330 | 34,004 | 348 | 4 |
| unfiltered (dump, replay line) | 111,419 | 14,274 (−87 %) | 22,330 | — | 36,725 (208 cards) | — |

Then one commit: seat 1 draws a card (the top library card moves to the hand, one `draw_card` event).

| for seat 1 | bytes |
|---|---|
| the next whole frame | 70,804 |
| an RFC 7386 merge patch between the two frames (arrays replaced whole) | 45,826 (65 % of the frame), deflated 5,816 |
| of that patch: `seats` | 23,474 (every seat's zones are arrays, so a draw resends the hand and the library) |
| of that patch: `log` | 22,326 (the 200-entry window slides by one, so the whole array differs) |
| the next whole frame, deflated with the previous frame's last 32 KB as a dictionary | 8,979 (vs 9,061 without: −1 %) |

Three things follow:

- **Compression alone takes a frame from ~70 KB to ~9 KB.** The view is repetitive JSON: field names, UUIDs, the same card shape a hundred times.
- **A naive diff saves little.** A generic JSON patch saves a third, because the view is made of arrays (zones, the log window) that a merge patch can only replace whole. A diff that pays has to know the view's shape: zones keyed by card instance, the log as entries after a sequence number. That is a protocol of its own, not a library call.
- **Context takeover would not help.** permessage-deflate's window is at most 32 KB and a frame is more than twice that, so carrying the compressor's state from frame to frame finds almost none of the previous frame. gorilla/websocket supports only "no context takeover" anyway.

The filler cards of the benchmark board are small (177 B per library card in the unfiltered view). Real catalog cards carry more text and fields; #3823's estimate was about 1.5 KB per library `CardView`. These numbers are a floor for a real table and should be re-taken on a real game (open question 1).

What is **not** measured: the commit rate of a live table, the bytes per second a table puts on the VM's uplink, and whether the uplink or the browser's parse is ever the bottleneck. ADR 0148's load test found CPU first, then memory; bandwidth was not mentioned.

### Library card views in the dump

The issue text says the dump *omits* library card views. It does not: `ViewOfGame` builds a full `CardView` for every library card, and `captureLocked` (`server/internal/ws/room.go:969`) marshals that unfiltered view into both the crash dump (`games/<id>.json`, the last view, overwritten) and a line of the replay (`replays/<id>.jsonl`, one unfiltered view per commit, ADR 0149). On the benchmark board the libraries are 36.7 KB of the 111 KB unfiltered view (33 %), with filler cards.

Who reads them:

- **`FilterViewFor`**, which sends the owner their own library as one redacted placeholder per card (`redactZone`, `view.go:10638`; 9.1 KB of seat 1's frame, mostly its own ~52 placeholders), keeps only a known top card for an opponent (`keepKnownTopInLibraryZone`, `view.go:11735`) and hides the zone from spectators and the admin view (`view.go:10706`).
- **The log's name resolver**, which indexes card pointers in the view's zones, libraries included (`newLogNameResolver`, `server/internal/protocol/log.go:2347`), to name a card a log line mentions.
- **The client**, which looks cards up by instance ID in the viewer's own library (`client/src/lib/contextMenu.logic.ts:272`, `Board.svelte:1637`, `ChoicePromptModal.svelte:1088`). It gets placeholders, but needs the IDs.
- **The replay and the dump**, through `streamFilteredReplay` (`server/internal/lobby/http.go:3020`) for a player's download, the operator's `GET /games/{id}/replay` and the bug-report bundle (ADR 0017). The replay is the only history of what a library held at each commit: the restore point (`restore/<id>.json`) holds the library too, but only the newest restorable state.

The cost on the commit path is the build in `ViewOfGame` and the marshal into the dump and the replay. Since #3887 the whole commit is 2.4 ms on the midgame board, and #3823's breakdown put the dump and the replay line at ~0.57 ms. The library share of that was not separated and is not measured here.

---

## Options for the wire

### W1. Keep full frames, and compress them (permessage-deflate)

Turn on `EnableCompression` in the upgrader and send each viewer's frame as a `websocket.PreparedMessage`, which gorilla compresses once per negotiated setting and shares across connections, so #3884's "one frame per viewer" holds for the compressed bytes too.

- **Wire:** ~87 % fewer bytes on the benchmark board (table above), for every frame, on every path: commits, presence, reconnect.
- **Protocol:** none. Compression is negotiated per connection in the handshake. Every browser offers it, so `ws.ts` changes nothing. The MCP seat's and `gamecli`'s `websocket.Dialer` do not ask for it (`mcpseat/conn.go:88`), so they keep receiving uncompressed frames, byte for byte what they get today.
- **Cost:** CPU to deflate one frame per viewer per commit, at a level we choose (`SetCompressionLevel`), plus a compressor per write. Not measured; it is the thing to measure before turning it on (owner question 3). ADR 0148 says CPU is what runs out first, so this trade is not free.
- **Proxy path:** Caddy passes WebSocket extensions through. Whether the Cloudflare edge in front of `labxp.io` keeps permessage-deflate end to end has to be checked on cmd-dev (open question 2).
- **Reconnect, snapshots, replay:** untouched.

### W2. Section deltas: a `snapshot_patch` frame against a named base

A new frame kind, added inside v0: `{base: {generation, seq}, seq, generation, set: {<top-level key>: <whole value>}, log_append: [...]}`. The server keeps, per viewer ID in a room, the last frame it sent; on a commit it compares the new filtered view section by section (each top-level `GameView` key, each `PlayerView`), sends only the sections that changed, and sends the log as the entries after the last one sent. A client applies a patch only when `base` equals what it rendered; otherwise it asks for, or is sent, a whole `snapshot`.

- **Wire:** the draw above would carry the drawing seat's `PlayerView`, the turn fields that moved and one log entry instead of 70 KB. Not measured as built; it depends on the section granularity. Board-wide changes (a wrath, a new turn) still resend most of the view.
- **Protocol and client:** a new frame kind, a base check, a resync path, and a merge in `ws.ts` that rebuilds a whole `GameView` before the stores see it, so components keep reading whole views. `seq` stays the ADR 0044 watermark; `generation` still means "discard". Every wire reader has to learn it or opt out: the MCP seat (`mcpseat/seat.go`) and `gamecli` would need a "full frames only" switch in the handshake (for example `?frames=full`) so a third-party agent client written against `snapshot` keeps working.
- **Server:** a last-sent view per viewer per room (one filtered view each; ~70 KB as JSON on this board, more as Go structs), the comparison on every commit (more CPU, not less: it is added to the build and the filter, which stay), and the bookkeeping on undo, rewind, admin seat switches and presence.
- **Ordering:** a delta depends on every frame before it, so the 16-frame send buffer can never drop or merge one. Today a full frame supersedes every earlier one, which is what would make "keep only the newest pending frame" (W4) possible. Deltas close that door.
- **Snapshots and replay (#522):** the restore point and the replay are not wire frames and stay full; no `SnapshotSchemaVersion` change. The replay must never be written as deltas: `streamFilteredReplay`, `statsbackfill` and the ADR 0149 scrub read one whole view per line.

### W3. Per-zone dirty flags from the engine

The engine marks what a commit touched (a zone, a seat, the turn, the log) and `ViewOfGame` rebuilds and sends only those parts. This is W2 with the comparison replaced by bookkeeping in `game`: cheaper per commit, but every mutation path in the engine would have to set the right flag, and a missed flag is a seat that silently shows a stale board. Effects that change a derived value on cards they never touch (a lord's anthem, a layer-7 static) make "what changed" a question about the projection, not the mutation. Not recommended at this project's size.

### W4. Keep full frames, and conflate a slow connection

Not a byte saving on a healthy link, but the other half of the cost a slow link pays today: when a connection's buffer is full, replace the queued snapshot with the newest one instead of disconnecting (`sendRaw`, `hub.go:1780`). It is safe only because every `snapshot` is whole. Listed because it is the option W2 rules out. No measured need: the disconnect has not been reported as a problem.

### W5. Trim the log window out of the frame

The log is 22 KB of every frame (32 % of a seat's), and the client already keys life cues and the log on sequence numbers (`docs/protocol.md`, `life_history.seq`). Sending only new log entries, with the window on the initial snapshot, is the cheapest slice of W2. It still needs the base check and resync of W2, so it is W2's machinery for one section.

## Options for the dump

- **D1. Keep library card views in the dump and the replay (today).** The replay stays the only per-commit history of library order, which a bug about a tutor, a shuffle, a "put on the bottom" or a reveal needs.
- **D2. Build only known and top library cards in `ViewOfGame`.** Saves the build and the bytes in the dump and the replay. It changes what a forensic replay shows, the owner's placeholders (`redactZone` needs the IDs and the count) would have to come from somewhere else, and the log resolver would lose names for library cards no one knows (which it does not print anyway, but the change has to prove it).
- **D3. Keep them in `games/<id>.json` and drop them from replay lines.** Halves the problem: the dump is one overwritten file and costs nothing that grows, and the replay is the one that grows by a full view per commit. But it is the replay that bug reports read.

---

## Decision (proposed)

1. **Wire: W1 first, and only behind a measurement.**
   - Add a bytes-out counter beside `cmdctrl_ws_frames_total` (by kind), so frames' share of the uplink is a series and not a guess.
   - Benchmark deflate on `BenchmarkBroadcastWithSpectators` and the midgame board at levels 1 and 6, and report CPU per commit against today's 2.4 ms.
   - If the CPU is acceptable, turn on permessage-deflate with prepared messages behind an env switch that defaults on, check it end to end through Cloudflare on cmd-dev, and confirm the MCP seat still receives uncompressed frames.
   No protocol change, no client change, no snapshot change.
2. **Wire: do not build W2, W3 or W5 now.** W1 removes most of the bytes for none of the protocol, resync, client-merge and per-viewer memory W2 costs, and W2 adds CPU on the path #3823 just cut. Reopen W2 only if the bytes metric shows frames costing after W1, and then start with W5 (the log), since it is the largest single section and already sequenced.
3. **Dump: D1, keep library card views.** The replay is the only record of library order over a game, the cost is a fraction of a commit that is now 2.4 ms, and nothing has shown it costing. Revisit if a real-game profile shows the library build as a large share of `ViewOfGame`.
4. **Correct the issue's premise in the record:** the dump does include library card views today.

## Consequences

- If W1 lands: a browser's frames shrink by roughly an order of magnitude; the server pays a deflate per viewer per commit; the MCP seat, `gamecli` and any agent client keep the plain wire; reconnect, rewind, the replay and the restore point are unchanged.
- If W1's CPU cost is too high for ADR 0148's budget, nothing is lost: the switch stays off, and the bytes metric still answers whether bandwidth was ever the problem.
- W2 stays possible. `protocol.go` and `docs/protocol.md` already promise that a new frame kind can be added inside v0, and nothing here spends that.
- Choosing W2 later rules out W4: a delta stream cannot be conflated.

## Owner questions

1. **Is there a real need?** The issue was parked until a profile shows frames or dumps costing. Do you want the bytes-out metric shipped first and this ADR left Proposed until it shows a number, or W1 built now on the benchmark numbers?
2. **Cloudflare:** is `cmd.labxp.io` (and cmd-dev) proxied through Cloudflare for WebSocket traffic, and is it acceptable to verify permessage-deflate end to end on cmd-dev before turning it on in prod?
3. **CPU budget:** ADR 0148 says CPU runs out first. What deflate cost per commit is acceptable, for example "no more than +10 % of a commit on the midgame board", before W1 ships default-on?
4. **Wire deltas (W2) at all?** Do you agree to defer them indefinitely behind a measured need, or do you want W5 (the log as new entries since a sequence number, 32 % of a frame) designed now as the first delta?
5. **Library cards (D1/D2/D3):** keep them in the dump and the replay (recommended), or trade replay forensics of library order for the bytes?
6. **Third-party wire clients:** if W2 ever ships, is a "full frames only" handshake switch for the MCP seat and other agent clients the right compatibility line, or should agent seats move to deltas too?
