# ADR 0178 — Life instructions after a replaced move

Date: 2026-10-14
Status: Accepted
Sprint: S58 — Deck requests, October batch follow-up
Issues: #2950, #2910

## Context

The Darkness Crystal replaces a death with exile and a gain of 2 life. Its
callback currently starts the life event before the exile physically occurs.
When life replacements ask for CR 616 ordering, the parent move continues and
runs the original effect's continuation while that life question remains open.
A real destroy-then-draw regression reproduces the early continuation.

The repository-pinned 2026-09-25 rules require printed instruction order
(608.2c), allow a modified event to contain several instructions (614.6), and
consider replacements of contained events after the first replacement is
chosen (616.1g). An impossible move instruction does not erase a separate
life instruction. A stale original event that never occurs is different
(614.7). Existing move and life replacement pipelines already handle all
individual operations; their structured continuations need one link.

## Decision

1. Add a narrow value-only life instruction list to the transient
   ReplacementEvent for a replaced move. Each entry stores source ID, affected
   player ID and signed life delta. A locked-context helper validates the
   parent kind and appends it. The Crystal redirects the move and stamps its
   existing exile link, then declares its gain instead of executing it there.
2. The shared battlefield-leave finisher lands the modified move first. It
   then runs the declared life instructions through the ordinary life
   replacement pipeline, holding the outer route continuation until they all
   settle. This covers both inline moves and resumed move ordering.
3. A typed private continuation on lifeTail stores the remaining instruction
   values, the outer zoneRoute and its simultaneous-exit context. It captures
   no mutable parent event or route in a Go closure. The ordinary life tail
   clears itself before resuming this sequence; the sequence consumes each
   instruction once, then runs the outer route tail once. Publish the carried
   simultaneous-exit context while resuming that tail, as ordinary exits do.
4. Declared life instructions survive a later destination redirect or
   cancellation of the move subevent, per 614.6. A stale parent move abandoned
   before it occurs discards its instructions but still settles the existing
   route continuation. A prohibited or replaced-away gain settles its own
   life tail with zero and advances normally. Player/source identities are
   snapshots of the replacement's application, so later source departure
   does not rewrite who was to gain life.
5. Cost moves pass their existing mustSettleNow posture into the contained
   life events. An indivisible cost never pauses on an inner ordering prompt.
   Effect-time moves retain their ordinary choice window.
6. Clone replacement resumes with fresh instruction slices and a fresh typed
   continuation/route, including simultaneous context, so answering a life
   prompt cannot consume an undo snapshot's outer continuation. Additive
   private frame data adds no persisted card, event or snapshot field and
   does not change schema v8 or frozen fixtures. Existing continuation census
   and closure inventory must continue to count the entire paused resolution.

## Validation

Pin exile-before-life-before-original-continuation for inline and nested
ordering paths. Test a parent move ordering prompt followed by a life prompt,
multiple sequential life instructions, prohibited/cancelled life, redirected
and cancelled moves, stale parent withdrawal, affected-player departure,
source departure, must-settle-now costs, simultaneous-exit context and
independent clone/undo replay. Keep the Crystal's discount, nontoken/opponent
filter, linked targeted reanimation, tapped entry and additional counters.
Run affected game/effects/legal/protocol/coverage/roadmap suites and the
snapshot/census/source guards. Do not close the issue on its existing Full
catalog declaration; close after verified PR merge.

## Number reservation

Fetched all remote refs and checked 13 branches on 2026-10-14: origin,
origin/cost-ledger, origin/develop, origin/docs/3337-away-seats,
origin/docs/issue-audit, origin/feat/750-conditional-block-restrictions,
origin/feat/playmats, origin/fix/2948-liliana-simultaneous-sacrifice,
origin/fix/3159-3160-attachments, origin/fix/3338-tablet-card-reading,
origin/fix/caddy-reload-admin-off, origin/main, origin/wip/836-one-click-default.
0175–0177 were occupied; 0178 was the first free number.
