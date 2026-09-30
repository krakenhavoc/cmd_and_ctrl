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

*Implemented in #517:* `auth.HMACAuthenticator`, keyed by
`CMDCTRL_SESSION_KEY` (at least 32 bytes, not equal to the admin
token, generated on each host by CD). `Revoke` is advisory. The
credential also carries `UserID` and a millisecond `IssuedAt`, for
[ADR 0051](0051-user-database.md) decisions 3 and 6.

*Bounded revocation, S34 sub-PR 7 ([ADR 0051](0051-user-database.md)
decision 6):* a session with a `UserID` can now be withdrawn.
`auth.WithRevocation` wraps this authenticator and refuses a token
issued at or before `users.sessions_invalid_before`, which
`POST /logout/everywhere` and `POST /admin/users/{id}/revoke-sessions`
set. The user's open sockets are closed with `1000 "session revoked"`,
a second sender of the terminal code in the decision 2 table.
Sessions with no user (admin, guest, spectator) keep the advisory
`Revoke` described here.

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

## Amendment 2026-09-24: owner decisions for the rest of S33

Recorded on [#515](https://github.com/krakenhavoc/cmd_and_ctrl/issues/515)
to settle the design questions left open in S33's sub-issues, and
implemented starting with [#524](https://github.com/krakenhavoc/cmd_and_ctrl/issues/524)
(this PR). Decisions 1-7 are the owner's, verbatim in substance;
"#524's own design" at the end is this PR's.

1. **#522 ability check** — when a card is restored with FEWER
   abilities than were captured, restore the table anyway and report
   it loudly: an ERROR naming the game and the card, counted in the
   census, and the card flagged as not automated for the game (the
   `manual` chip). A card restored with MORE abilities than captured
   is not a mismatch — a new build adding abilities to a card is the
   ordinary case, not a schema problem.
2. **#522 fixtures** — both kinds belong in the corpus decision 7
   (above) calls for:
   - generated fixtures, written by a tool from scripted boards,
     append-only, never regenerated, one frozen set per schema
     version;
   - a few real restore files pulled from cmd-dev, scrubbed of player
     data.
3. **#523 restore generation** — a counter saved in the restore file,
   bumped on every restore. The client drops its cached view when the
   value changes. Players see a toast only when the table actually
   rewound ("The table was restored to an earlier point after a
   server restart"); a restore with no rewind shows nothing.
4. **Exit criterion 3 is amended to match #520's design.** A player
   who cleared their storage and opens the invite link on a started
   table is told how to get back: signed in, "My Games" rejoins in
   one click; as a guest, ask the host for a seat-reclaim link.
   `Join.svelte` stops telling them to "ask your host for a spectator
   link".
5. **Dead session on reconnect** — after a few failed dials, the
   client checks whether the session is still valid. On a 401 it
   stops retrying and shows "Your session ended — sign in again",
   linking to Login, or to My Games for a signed-in identity. A server
   that is only restarting keeps the normal retry ladder.
6. **Order.**
   1. a live restart check on cmd-dev, then close #518;
   2. #524 (this PR);
   3. #522 (the ability check, then the fixtures);
   4. #523, together with the reconnect and rewind sections of
      `protocol.md`;
   5. the return-path fixes (decisions 4 and 5), tracked as their own
      issue.
7. **Next sprint: ADR 0041 phase 3, "effects as data."** It starts
   with the effects that freeze a restore point for longest (earthbend,
   amass, control exchange, Sower-style control) and reuses
   [ADR 0093](0093-abilities-granted-to-other-permanents.md)'s
   `ScopedGrant` record. #524's census numbers set the priorities.

**#524's own design.** `ws.Room` gains `lastRestorePoint`
(`{Seq uint64; At time.Time}`), a small bookkeeping struct guarded by
the room's own mutex and set only inside `writeRestorePointLocked` on
a successful write — and seeded at boot, in `restoreOne`, from the
restore file's own `Seq` and `Snapshot.TakenAt`, since at that instant
the file IS the room's last-written point. `Room.ShutdownReport()`
pairs that bookkeeping with a fresh `CaptureSnapshot()` — a read, no
disk write — to answer, per room: whether the CURRENT state is itself
a clean boundary (`Clean`), the `ContinuationCensus` when it is not,
and (when a restore point has ever been written) that point's `seq`,
its age, and how many actions have happened since (`SeqBehind`,
`LiveSeq - RestoreSeq`).

`ws.LogShutdownCensus(log, rooms, archived)` calls that once per live
room and logs one line each — `Info` for a clean table, `Warn` for one
that would rewind, naming its census labels — plus one `Info` summary
line with totals (`games`, `clean`, `would_rewind`,
`actions_rewound`). `archived` is a predicate the caller supplies
(main.go looks it up via `lobby.Lobby.Get`), because `ws` does not
otherwise know about `lobby.GameMeta.ArchivedAt` and importing it
would invert the dependency; this is also what keeps the function
testable with a fake. `main.go` calls it once, at SIGTERM, after the
HTTP listener closes and before `hub.Shutdown` — a read of state the
process already holds, so it adds no measurable time to the 5s
shutdown grace period and cannot block it (decision 1's "log line, not
a write").

On the boot side, `RestoreFromDisk`'s own call is now timed in
`main.go` and logged (`"restore from disk complete"`,
`duration`, `resumed`) — the number decision 1's "measurement task"
asked for, to confirm or refute the "startup dominates, not this
pass" hypothesis. Each per-game `"game restored"` line grows a
`restore_point_age` field (`time.Since(Snapshot.TakenAt)`), and
`RestoreOutcome` grows a `Reason` field so `LogRestoreSummary` can
tally *why* a table was abandoned (`abandoned_reasons`) rather than
only how many were.

**Deliberately not computed here: how far a restored table rewound,
at boot.** That number is the restored seq against the seq the
*previous* process was showing to clients right before it died — and
the only place that fact is recorded is the previous process's own
shutdown census line, in its own log stream, which this process has no
reliable way to read back (log rotation, a crashed process that never
reached the shutdown handler, a multi-host log aggregator that is not
this codebase's concern). Inventing a number here would be worse than
not having one. The shutdown line already carries the true rewind
delta — `seq_behind` — for the process that is about to stop; the
boot line carries only what boot can honestly know, the restore
point's own age.

## Amendment 2026-09-28 (#1698): one sanctioned regeneration of the v7 corpus

Decision 7's rule is "appended to and never rewritten," in that order,
and this is the one time the second half is set aside on purpose — not
a reversal of the rule, a recorded exception to it.

**What happened.** Five PRs in one day — #1661, and the four ADR 0027
amendments that followed it, #1675, #1679, #1682 and #1689 — each added
one `omitempty` field to `Event`, all in the same place: right after
`CauseItem`, all stamped only on `EventLTB`, all carrying a leaving
permanent's last-known battlefield state (CR 603.10a) for a dies /
leaves-the-battlefield trigger to read. Nine fields in total:
`attacking_target`, `blocking_target`, `blocked`, `last_known_types`,
`last_known_subtypes`, `last_known_all_creature_types`,
`last_known_supertypes`, `last_known_controller`, `last_known_colors`.
None of the five PRs touched the snapshot corpus, because none of them
had reason to think they were a schema decision — each looked, in
isolation, like adding a field that zero-values correctly, the case
`SnapshotSchemaVersion`'s own comment says needs no bump.

Diffing the regenerated output against the committed fixtures (below)
turned up a second, older, unrelated source of the same kind of drift:
`Characteristic.BlockRequirements` (`internal/game/characteristic.go`),
added by #1597 (PR #1683) and widened in scope, not in shape, by #1684
(PR #1693). It carries no `omitempty` — every `Characteristic` on the
wire has always printed `"AttackRequirements": null` for a permanent
with none, and after #1683 it started needing to print
`"BlockRequirements": null` right beside it — and #1683 never
regenerated the corpus either. So the writer has in fact refused since
#1683 landed, on any board a `Characteristic` reaches (which is every
board), independent of and earlier than the LKI cluster; #1696 is only
the first PR that happened to need a NEW board while the writer was
in that state, which is what surfaced the refusal as a filed issue.

Collectively these ten fields (nine on `Event`, one on `Characteristic`)
left every fixture in `internal/game/testdata/snapshots/v7/` rendering
one byte different from what is committed, which is exactly what
`TestWriteSnapshotCorpus` (this decision's guard) exists to catch — it
refused to write anything, including the two new boards #1696 had added
to `corpusBoards()` (`detection_tower`, `arcane_lighthouse`) with
nowhere to land.

**Why this is the additive case, not a silent change.** Decision 7
exists to catch a rename, a repurpose or a unit change wearing an
additive disguise — the failure mode is a fixture that still decodes
but means something different. This is not that. Applying ADR 0041
Decision P10's test for "additive" (`docs/decisions/0041-game-persistence.md`,
"Decision P10"): today's binary is the first to write any of these ten
fields, so an older binary was never asked to restore a file holding
them, and dropping them from its own files makes nothing worse for it.
Three things confirm it rather than assume it:

- `Event` and `Characteristic` have no custom `(Un)MarshalJSON`, and
  `internal/game` has no use of `json.DisallowUnknownFields` anywhere.
  Decoding is the standard library's ordinary permissive behaviour both
  directions: a pre-#1661/#1597 file missing these fields decodes into
  today's binary with the documented empty/zero fallback each field's
  own comment describes (`WasType`, `WasColor`, `LeftUnderControlOf`,
  a nil `BlockRequirements` slice reading as "none," exactly what an
  older file's absence of the field means), and a file carrying them
  would decode into an older binary with the extra keys silently
  ignored. This is unlike the `ErrUnknownEffectKey` vocabulary this
  same ADR's decision 7 and ADR 0041's P4/P10 built for the "effects
  as data" surface (stack-item bodies, `params`, scoped effect kinds)
  — that machinery is a deliberate, manually-checked refusal gate, and
  it does not run over plain `Event` or `Characteristic` fields, so
  nothing here is being routed around it.
- Every fixture that existed before this regeneration — all 39 of
  them, written across v7's whole life, from before any of tiers 1
  through 4 — still passes `TestSnapshotCorpusRestores` unchanged
  under the binary that added all ten fields. That is the property
  decision 7 actually cares about (a fixture this binary did not write
  still restores), and it was never broken. Only the WRITER's
  overly-blunt byte-for-byte check — which cannot distinguish "the
  shape changed" from "a new field was appended" — was refusing to
  run.
- Diffing every regenerated file against its committed original,
  hunk by hunk, confirms nothing else moved: every line that
  disappears from a fixture reappears byte-for-byte (its trailing
  comma aside) among the lines that replace it, in the same hunk. No
  fixture lost a key, changed a value, or reordered anything. The only
  wholly new content, across all 39 files, is the ten fields above and
  — on the handful of fixtures whose scripted board already put a
  permanent through a stamped `EventLTB` — the real (nonzero)
  `last_known_*` values for that permanent, which is new information a
  pre-#1661 binary could not have captured, not a change to anything
  it did capture.

**What was done.** The 39 existing files under
`internal/game/testdata/snapshots/v7/` were deleted and rewritten by
`TestWriteSnapshotCorpus -write-corpus` against the current binary, in
the same commit as this amendment, alongside the two new files the
regeneration unblocked (`detection_tower.json`, `arcane_lighthouse.json`).
A sample diff is in the PR body for #1698.

**What this does not authorise.** The append-only rule stands for
everything else. A future PR that adds a field is not free to
regenerate on the strength of this precedent — it has to make the same
case this amendment makes (no decode-time refusal gate, the field's own
comment describes the safe empty fallback, and old fixtures still
restore unchanged before touching anything), and it has to write that
case down, here or in a new dated amendment, before running
`-write-corpus` over a file that already exists. A field that removes,
renames or repurposes anything, or that needs a refusal channel because
an older binary would otherwise misread it, is Q1/P10's other branch —
a new schema version, not a regeneration.

## Amendment 2026-09-28 (#1712): a second sanctioned regeneration — a test-helper fix, not a wire-format addition

The #1698 amendment above covers a schema addition (ten new fields, one
new value). This one is a different shape of "additive" — no field was
added or changed meaning; a test double that had been silently
skipping a step started performing it — and it is recorded separately
because the case for it is not identical.

**What happened.** #1707's second commit fixed `pushCatalogPermanent`
(`internal/cards/effects/activated_test.go`), the helper two scripted
corpus boards (`gingerbrute`, `whip_redirect`) use to seed a catalog
permanent straight onto the battlefield. Before the fix it only
appended the card to `g.Battlefield.Cards` — no `EmitEvent`, so no
`EventZoneMove`, so `stampBattlefieldEntryLocked` never ran and the
card's `EnteredBattlefieldAt` stayed at its zero value. That is not a
state any REAL battlefield entry can produce: production always routes
through `EmitEvent` (`Game.MoveCard` and every mutation that pushes a
card onto the battlefield emits `EventZoneMove`), so a zero
`EnteredBattlefieldAt` on a live permanent never happens. The two
frozen fixtures were capturing an artifact of the test double, not a
value a real snapshot has ever held. #1707 fixed the helper to emit the
event the same way `pushBattlefieldCardWithTimestamp` does, but did not
re-run `-write-corpus` over the two files this changed the output of —
that is what #1712 tracks.

**Why this is not a compatibility break.** Nothing about the
`enteredBattlefieldAt` field itself changed — same key, same type
(int64 nanoseconds), same meaning, no rename or repurpose. Applying
Decision P10's test again: the two existing fixtures, unmodified, with
the old `0`, still pass `TestSnapshotCorpusRestores` today, unchanged,
under the binary that carries the `pushCatalogPermanent` fix (verified
before touching either file). Regenerating them is not fixing a
restore failure; it is only unblocking `TestWriteSnapshotCorpus`'s
self-consistency check, which cannot tell "the wire format changed"
apart from "the scripted board now performs one more real action than
it used to."

**Why the diff is bigger than one field, and why that's still the same
change.** Emitting the extra `EventZoneMove` is not a value patched in
after the fact — it is one more real event appended to the game's
event log, mid-script, before the actions the rest of each board's
`build` function already performs. Once the deterministic corpus clock
(`corpusEpoch`, ticking `+1000` per call) is asked for one more
timestamp before the point it used to be asked for the first one, three
things shift by construction, not by coincidence:

- the new `zone_move` event itself appears in `events`, at the seq slot
  it now occupies;
- every event logged afterward in the same board carries a `seq` one
  higher than before, because the counter is monotonic and one more
  event landed ahead of it;
- every `enteredAt` / `PinnedEnteredAt` / scoped-effect `timestamp` /
  `layerVersion` that reads off the corpus clock or off this
  permanent's own `EnteredBattlefieldAt` advances by the same one tick,
  because they are all derived from the same clock and the same stamp,
  not independent literals.

Diffed hunk by hunk against the committed originals (sample in the PR
body for #1712), that is the whole of it for both files: one new event,
and every downstream value that was always going to be "whatever the
clock said last" shifted by the one call the fix inserted. No key was
renamed, no value took on a new meaning, and nothing before the fixed
helper's call moved at all.

**What was done.** `gingerbrute.json` and `whip_redirect.json` under
`internal/game/testdata/snapshots/v7/` were deleted and rewritten by
`TestWriteSnapshotCorpus -write-corpus`, in the same PR as this
amendment, alongside the new `duration_copy.json` (#1593) that file's
absence was blocking. `detection_tower.json` and `arcane_lighthouse.json`
(#1696) needed no action here — #1707 already wrote them, new, in its
own regeneration pass.

**What this does not authorise.** Same boundary as the #1698 amendment:
this is not a general license to regenerate a fixture because a test
helper changed. It applies specifically to a scripted board where the
only thing that changed is that a **test double started doing what
production already always did** — production's own `EnteredBattlefieldAt`
semantics did not move. A fix to a helper that changes what a board
actually represents (a different card, a different zone, a different
player) is a new board under a new name, not a regeneration of the old
one.
