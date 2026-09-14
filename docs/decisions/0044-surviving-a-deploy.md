# ADR 0044 — Surviving a deploy: the return path

**Status:** Accepted · 2026-09-14 · Sprint S33 · Issue #515
**Builds on:** [ADR 0041](0041-game-persistence.md), whose decisions all
stand. **Supersedes in part:** ADR 0041's claim that nothing on the
client needs to change.

## Context

ADR 0041 set out to make a deploy cost a live table a page refresh
instead of the game, and it did the hard half. The engine snapshot is
schema-versioned and round-trip-exact, `snapshot_drift_test.go` fails
CI on an unclassified field, the RNG's stream position rides the file,
the invite tokens are persisted so a restored table is openable, and
`RestoreFromDisk` runs before the first route is registered. The data
directory is a host path on a dedicated disk; the deploy rsyncs a
binary and restarts a unit, and never touches it. **The board really
does survive `systemctl restart`.**

The players do not get back to it.

ADR 0041 rested its client story on one sentence:

> A refresh is acceptable. Connection continuity is not the
> requirement — surviving a process restart is. That matters, because
> the client already auto-reconnects and resyncs by `seq`, so nothing
> on the client needs to change.

Both halves are false, and the S33 audit found a third failure behind
them.

- **The client does not auto-reconnect on a deploy.** `ws.ts`'s close
  handler treats close **1000 and 1001 alike as terminal**, because
  when that branch was written `hub.EvictGame`'s 1000 ("game deleted")
  was the only deliberate server close and the reasoning — *don't
  redial a game that is gone* — was correct. `hub.Shutdown` now sends
  **1001 on every restart**, and the game is emphatically not gone.
  The exponential-backoff ladder immediately below that branch is dead
  code on the exact path it was written for. This is the stale-comment
  shape: true when written, false once ADR 0041 landed.
- **There is no resync by `seq`.** The protocol has no resume frame and
  no `since` field; the client never tells the server what it has seen.
  The server unconditionally stages a full snapshot at upgrade time.
  `seq` is an ordering value, not a resume cursor.
- **And the session is gone anyway.** `auth.MemoryAuthenticator` is an
  in-process map. Every token minted by the previous process fails
  `Validate`, which the WS authorizer turns into a 401 on the upgrade —
  which a browser reports as close 1006, indistinguishable from a
  network blip.

ADR 0041 named the recovery: *"after a deploy every player
re-authenticates through their invite link."* That path does not
exist. `Lobby.JoinWithIdentity` refuses any game that has left the
lobby, and there is no seat-reclaim route. A player whose credentials
died with the process can spectate their own game. They cannot play
it.

So: state persistence works, and the return path was never built. This
ADR builds it, and makes four smaller calls the audit forced.

## Decision 1 — Cold restore, not drain. The topology forbids one.

A drain is the reflex answer and it is wrong here, so it is recorded
as a decision rather than left as an omission for someone to
"discover".

Production is one systemd unit on one VM, binding `127.0.0.1:8080`
behind Caddy. `systemctl restart` is stop-then-start of the same
binary on the same port: the old process **must** be fully dead before
the new one can bind. There is no second instance, no rolling
replacement, and nowhere to drain *to*. Draining would mean refusing
work while still holding the port, which is downtime wearing a
different hat.

Nor would a longer grace period buy anything. Durability is already
per-`Apply` — `captureLocked` writes the restore point on every
action — and the shutdown path deliberately writes nothing. A
shutdown-time capture could not improve matters either, because
`writeRestorePointLocked` skips a game holding continuations, which is
precisely the game whose restore point is stale. The fix for staleness
is decision 5, not a drain.

**So the deploy stays a cold restart, and the work goes into making
the restart cheap and the return automatic.** Revisit only if the
topology gains a second node.

What the shutdown path *should* gain is a **log line**, not a write:
the instant before exit is the only moment where "which tables are at
a clean boundary, and how far back is each one's restore point" is
knowable, and today it is thrown away.

## Decision 2 — Close 1001 is retryable; 1000 stays terminal

The two codes mean different things and must stop sharing a branch.

| Code | Sender | Meaning | Client |
|---|---|---|---|
| 1000 | `hub.EvictGame` | the game is deleted | terminal — render an ended state |
| 1001 | `hub.Shutdown` | the process is restarting | **reconnect** |
| 1006 | anything abnormal | blip, crash, or a rejected upgrade | reconnect |

There is a second, nastier problem in the same place: `hub.Shutdown`
does `WriteControl` immediately followed by `conn.Close()`, so whether
the browser observes 1001 or 1006 depends on whether the close frame
beats the FIN. **The same binary behaves differently on different
deploys**, which is worse than either outcome consistently, and it is
why this looked like a flaky network problem rather than a code
problem. The close frame gets flushed deterministically.

Reconnecting is necessary and not sufficient: while the socket is
down, the board stays fully rendered and fully clickable, and every
click lands in a log store that only the bug-report modal reads. The
player sees a frozen table and no explanation — which is
indistinguishable from the state-freeze bug (#266) and gets reported
as one. Connection state becomes visible, and input is gated on it.

## Decision 3 — Sessions become stateless (HMAC), not persisted

ADR 0041 already named this and already named the seam:

> Swapping `MemoryAuthenticator` for an HMAC implementation is a
> one-line change in `main.go` — the `Authenticator` interface is the
> only seam the rest of the server sees.

That is still true, and it is the right shape. The alternative —
persisting the session map to disk beside the other artifacts — buys
the same durability and costs a new file to write, sweep, secure and
version. Signing the `Principal` into the credential needs no store at
all, and a store that does not exist cannot go stale, leak, or
disagree with the snapshot beside it.

**The honest cost is revocation.** A stateless credential cannot be
withdrawn by forgetting it; `Revoke` becomes advisory unless a
bounded denylist is added. That is recorded on the type rather than
papered over, because `Authenticator.Revoke`'s own doc comment already
tells callers that the answer depends on the implementation, and
leaving a silent `return nil` would make that warning useless.

Key handling: sourced from the environment alongside the admin token.
A **missing key falls back to `MemoryAuthenticator` with a loud
warning naming the variable** — a dev box keeps working, and
production cannot quietly boot with a fixed or absent key. Failing
open on a signing key is how you get an authenticator that authorises
everyone.

## Decision 4 — Seat reclaim is the backstop, not the mechanism

With decision 3 the common case is that the stored token still
validates and the reconnect simply succeeds. Reclaim covers what is
left: cleared storage, a new device, a token past its TTL, a closed
tab.

Making reclaim the *primary* path was considered and rejected. It
would mean every player re-walking an invite link after every deploy —
which is what ADR 0041 assumed and what its readers reasonably
believed already worked — and it makes a routine deploy a manual
ceremony for four people rather than an event they can miss entirely.
It also puts the heaviest possible load on the newest authentication
surface.

Because it **is** an authentication surface, it is scoped tightly. The
invite token alone must not be sufficient to take a seat in a game in
progress: today that token admits you to a lobby, and reclaim makes it
strictly more valuable. Reclaim requires proof of *which seat is
yours* — the engine snapshot already carries `Player.DiscordID` and
the persisted `GameMeta` already carries `SeatInfo`, and `GameMeta` is
already written `0600` precisely because it holds secrets, so it is
the right home for a per-seat reclaim secret if identity alone is
insufficient. `JoinWithIdentity` itself is **not** widened: a brand-new
seat in a started game remains wrong.

## Decision 5 — Rewind is a protocol event, and it is explicit

ADR 0041's restore-point rule is that a snapshot holding
un-rebuildable continuations is not written, so a restart rewinds to
the last state that can be rebuilt exactly. Sound. Its consequence for
the wire was not followed through.

`docs/protocol.md` promises `seq` is "monotonically **non-decreasing**"
and that clients use it to detect dropped or out-of-order frames. A
restored room sets `room.seq` to the restore point's `seq` — which,
when the game ran on through states that were never restorable, is
**lower than the value connected clients already rendered**. ADR 0041
addressed only the weaker case ("a restored room that restarted its
counter"); restarting at zero is prevented, rewinding to an earlier
real value is not.

It works today by accident. The client zeroes its `highestSeq`
watermark on every socket open, so the "older than highest; ignored"
guard never fires after a reconnect and the rewound snapshot lands.
The comment there argues, correctly, against persisting the watermark
across reconnects — but it means correctness rests on a side effect,
and the obvious "fix" of making the watermark durable would wedge
every rewound game permanently, showing a stale board against a server
that has moved on.

**So the rewind gets a name.** A restore generation is bumped when a
room is rebuilt from disk, travels in the restore-point file beside
`seq`, and is exposed on the snapshot frame. Within a generation `seq`
is non-decreasing and the existing guard is safe to trust; a
generation change means *discard and re-render*. The protocol doc is
corrected to say that, because a contract the code does not honour is
worse than no contract.

## Decision 6 — The undo stack is deliberately not persisted

A restored room starts with an empty undo stack and returns
`ErrNothingToUndo`. This is a choice, recorded so the next reader does
not have to re-derive it or file it as an oversight.

`Room.undoStack` is up to 32 pre-mutation `*game.Game` clones. Carrying
it would multiply a room's on-disk footprint by up to 32× and add 32
more surfaces for the schema-compatibility problem in decision 7 —
all to preserve the one piece of state whose loss has a trivial
player-level remedy: take the action again. `freeUndo` and the
per-turn `UndosRemaining` budget both ride the engine snapshot
already, so the *policy* survives; only the history does not.

## Decision 7 — Compatibility is enforced by fixtures, not by discipline

This is the one that bites hardest at several deploys a day, and it is
the easiest to believe is already handled.

`snapshot_drift_test.go` is a strong guard and this ADR does not touch
it. It forces every field on the domain types to be classified
**carried / rebuilt / dropped**, and it fails CI naming the field. But
it enforces *classification*, never *compatibility*, and the two are
not the same thing. Its round-trip property —

```
capture(restore(decode(encode(capture(g))))) == capture(g)
```

— has the **same binary on both sides**. Rename a snapshot mirror
field, repurpose one, or change a unit, and the domain field is still
classified, the round trip still passes, and yesterday's file decodes
into today's binary with a zero value or the wrong meaning. Silently.

`SnapshotSchemaVersion` and `checkSchema` handle skew correctly once
someone bumps the constant. Nothing makes anyone bump it, and the
constant's own comment concedes the judgement call ("Adding a field
that zero-values correctly does NOT need a bump"). Meanwhile **no test
in the tree has ever decoded a snapshot this binary did not write** —
there is no fixture corpus at all.

So: a committed corpus of real restore-point files, restored on every
CI run, **appended to and never rewritten**. Rewriting a fixture to
make a test pass converts the guard back into the comment it replaced.

The same section fixes a narrower instance of the same trust. Capture
records `ManaAbilityCount` / `ActivatedAbilityCount`; restore looks the
abilities up in the catalog by oracle ID and **never compares the
result to the recorded count**. A card whose catalog entry moved,
changed name, or vanished between binaries comes back on the
battlefield silently stripped of its abilities. The numbers needed to
catch it are already on disk and nothing reads them. A `rebuilt`
classification has to prove the rebuild happened, or it is a `dropped`
in disguise.

## Decision 8 — Tokens get catalog keys, so a Treasure stops blocking every restore point

ADR 0041 flagged this as phase 3's top priority and it is smaller than
that framing suggests.

The census check is:

```go
if (len(c.ManaAbilities) > 0 || len(c.ActivatedAbilities) > 0) && c.OracleID == "" {
    cen.IntrinsicAbilityCards++
```

It tests **whether there is an oracle ID to look up**, not whether
anything unserialisable is present. A Treasure token sets `TapCost`,
`SacrificeCost`, `Produced` and `Label` — every field pure data, every
closure field on `ManaAbilityShape` nil — and blocks every subsequent
restore point in the game for as long as it sits on the battlefield.
Food, Clue, Blood, Gold and Eldrazi Spawn do the same. These are
ordinary output of the current catalog, so in practice many real games
have **no usable restore point at all** and a deploy rewinds them
much further than the ADR's "last clean boundary" implies.

Carrying the data-only shapes verbatim would fix Treasure and not
Clue, whose `ActivatedAbilityShape.Effect` is always a closure. The
general answer is the one the codebase already uses for printed cards
and for token *copies*: **look the abilities up in the catalog by a
stable key**. Token templates get a synthetic key namespace
(`token:treasure` and friends) and are registered, so restore
re-derives their abilities exactly as it already does for a card with
a printing.

The namespace must not collide with real oracle IDs — a token copy
carries the copied card's real oracle ID under CR 707.2 and must keep
resolving to the printed card, not to a template. And the census
counter is narrowed rather than deleted: it must still fire for an
ability the registry genuinely cannot return, or the guard becomes
decorative.

The rest of ADR 0041's phase 3 — `StackItem.Effect`,
`DelayedTrigger.Effect`, the resume frames, the turn-scoped statics
and replacements — stays deferred. Those block only the window they
are open in; a token blocks everything after it. Decision 1's shutdown
logging exists partly to put real census numbers behind the order that
work is eventually done in.

## Consequences

- A deploy becomes what ADR 0041 promised: clients reconnect on their
  own, re-authenticate with the credential they already hold, and
  re-render the table. No refresh, no invite link, no lost seat.
- Revocation weakens (decision 3) and the restore point gets
  dramatically fresher (decision 8).
- The protocol gains a generation field and loses a promise it was not
  keeping (decision 5).
- Snapshot changes get slower and safer: adding a field now means
  thinking about a fixture corpus rather than only about a
  classification (decision 7).
- The shutdown path gains observability and, deliberately, no new
  writes (decision 1).
