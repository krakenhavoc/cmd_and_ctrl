# ADR 0076 — Opponent boards are read-outs, not small cards: the summary panel and what expands it

**Status:** Proposed · 2026-09-19 · S32 — Playtest stabilisation, round 1
**Issue:** [#956](https://github.com/krakenhavoc/cmd_and_ctrl/issues/956) (in-app
report: "On a three player game, I cannot see my whole hand")
**Numbering:** on 2026-09-19 the merged set on `develop` was enumerated
(0001–0075, with 0005, 0024, 0029 and 0030 permanently unused per AGENTS.md §4,
and no other gaps), plus `docs/decisions/` on `feat/1032-engine-table-settings`,
`feat/1038-rotate-invite` and `docs/s33-persistence-plan`, plus every branch name
matching `docs/adr-NNNN-*` (highest 0075). **This is weaker than the AGENTS.md
recipe:** the repo has ~300 remote heads and they were not all tree-listed,
because this session reaches the repo only through the GitHub MCP and cannot run
`git log --all`. `TestADRNumbersAreUniqueAndMatchTheirHeading` is the backstop if
0076 turns out to be claimed on a branch not checked here.
**Builds on:** [ADR 0075](0075-table-settings-and-host-controls.md) (the display
settings this extends), PR #858 / `claude/board-ratios.md` (the `cqh` size ramp
whose `clamp()` floor is the thing that broke)
**Amends:** nothing. `display.tableLayout` survives, but see §7 — it now decides
less than it did.
**Designs, does not decide:** which of the two expansion mechanisms wins (§6);
whether `tableLayout` should be retired (§7).

---

## 1. The problem

The table's only answer to "less space" is "smaller cards", and that answer has
run out.

PR #858 made every card height a share of its panel's height: `.panel` is a size
container and sets `--card-h: clamp(168px, calc((42cqh - 30px) * var(--card-scale)), 240px)`,
with the older fixed sizes as the floors — 168px self, 123px upright opponent,
90px flipped opponent. A panel that loses height or width renders the same four
rows at a smaller scale.

The `clamp()` floor is where that stops working. Past the floor the cards no
longer shrink, so the panel **clips** instead, because `.panel` sets
`overflow: hidden`.

#956 is that failure reported from a real game: in a 3-player game the viewer's
panel was half-width and the reporter could not see their own hand. The layout
half is fixed separately — the viewer now gets the full bottom row — but that
only moves the pressure. At four players every panel is a quadrant and opponent
cards sit at 90px, already at the edge of legible. There is no headroom left.

There is a second observation behind this, independent of the bug. What a player
reads off an opponent's board mid-turn is not the art. It is *can they respond*,
*what can block*, *how dead am I*. Four rows of 90px card images answer those
slowly; a designed read-out answers them at a glance, in less space.

## 2. Decision — two representations, chosen by a setting

`display.opponentDetail`:

- **`"summary"` (the default)** — life and commander damage, untapped mana by
  colour, creature count and power split tapped/untapped, one pip per creature
  carrying P/T and its combat keywords, hand and pile counts, and a tile per
  *structural* permanent (§4).
- **`"full"`** — every opponent rendered as cards at all times, exactly as the
  table worked before.

The property the whole decision rests on: **the summary degrades by CHANGING
rather than SCALING, so it has no floor to hit.** It reads the same in a
4-player quadrant as in a third of a 3-player top row.

Every derivation feeding it is pure and lives in `client/src/lib/seatSummary.ts`,
unit-tested without rendering Svelte. A rule about what the player sees is
covered by a test, not buried in a template.

## 3. Decision — expansion is driven by interaction requirement, never relevance

The constraint that shapes this: **summary panels still have to be clickable at
card level.** You click an incoming attacker to block it; you click an opponent's
creature as a spell target. If the summary renders pips instead of cards, those
clicks have nowhere to land. Expansion is load-bearing, not a nicety.

| Trigger | Condition | Why it is safe |
|---|---|---|
| Targeting | a targeting prompt is live and that seat controls a legal target | the viewer opened the prompt |
| Blocking | `combatMode === "block"` and that seat has attackers | the viewer entered block mode |
| Attacking | `combatMode === "attack"` and that seat is a legal defender | the viewer entered attack mode |
| Active player | that seat is the active player, gated on `display.expandActivePlayer` (default on) | turn boundary; cannot land mid-click |
| Pinned | the viewer clicked that seat's avatar | the viewer asked |

Every one is either initiated by the viewer or changes on a turn boundary. A pin
outranks the rest and survives turn changes; only one seat can be pinned, so the
viewer never loses their own panel to a pile of expanded opponents.

**What must never expand a panel:** a trigger resolving, a spell being cast, a
permanent entering, life changing. Those go to the attention strip. A board that
moves under a click the player has already committed to is the worst class of bug
in a game with priority windows.

Each creature pip is a real hit-testable element carrying `data-card-id`, because
a pip is a legal target and a legal block. `CombatArrows` resolves an endpoint by
measuring the element that carries the card; if the arrow layer keys off a
different attribute, the pip matches it rather than introducing a second
contract.

## 4. Decision — the summary is deterministic, and never ranks

Nothing in the read-out scores or picks "the important permanents". A heuristic
that ranks threats will be wrong at the worst possible moment and will be blamed
for the loss — fairly, because it chose what to hide.

The structural rule instead: a permanent earns a tile if it is a **planeswalker**
(loyalty is a second life total), the **commander** (the game is named after it),
or **wearing an attachment** (an Aura or Equipment drawn nowhere reads as a
permanent that isn't there). That is explainable and wrong only in ways a player
can predict.

The host set comes off the WHOLE battlefield, not one seat's slice: CR 303.4f
lets an Aura you control enchant a creature an opponent controls.

## 5. Decision — the mana read under-reports, on purpose

At ADR level because it is a contract, not an implementation detail.

`manaAvailable()` is a **floor, not a total**. A dual land counts once rather than
once per printed ability, because it can only be tapped once — summing would
report a Triome as three mana. Anything whose colour cannot be pinned down (an
any-colour source, a hybrid symbol, a generic amount, an absent `produced`
string) is counted in `flexible` and never folded into a colour. Floating mana,
rituals in hand, and costs the client does not model are excluded entirely.

A confidently wrong mana read is worse than no mana read, because the player acts
on it. Read as "they have at least this much" it is right; read as "they have
exactly this much" it is wrong, and the dashed `?` pip plus the spoken source
count exist to say so.

Cost checks reuse `abilityBlocked` from `contextMenu.logic`, so the summary and
the card menu cannot disagree about whether a source is live.

## 6. Undecided — which expansion mechanism wins

Two are implemented, selected by `display.expandStyle`:

- **`"reflow"` (default)** — the expanded seat takes a larger grid share and the
  summaries shrink. Everything stays in one plane, so `CombatArrows` keeps
  measuring both endpoints against `boardEl`. Harder to animate smoothly.
- **`"overlay"`** — the expanded panel floats over the table, anchored to its
  seat. Much easier to animate, but it covers other boards and puts one arrow
  endpoint across a z-index boundary.

Shipping both is deliberate: the difference is a feel judgement a prototype
cannot settle, and switching mid-game is the only way to compare them on the same
board. **One of the two, and the setting, will be deleted** once a few games have
answered it. Nothing else should be built on `expandStyle`.

## 7. Consequences

**Existing players' tables change appearance on upgrade.** The v10 → v11 settings
migration fills `opponentDetail` from the default, which is `"summary"`. Every
earlier migration in that chain changed how the client *behaved*; this is the
first that changes what the table *looks like* without the player asking. `"full"`
restores the old rendering exactly, and the row sits at the top of the Display tab
because that is what an upgrading player is looking for.

**The aria contract is preserved.** A summary is still that seat's board:
`role="region"` and `aria-label="<name> board"` are unchanged, so
`board-layout.spec.ts` and `game.spec.ts` keep finding seats. The expand control
is labelled separately so the full board reads as reachable.

**`display.tableLayout` decides less than it did.** Once opponents are read-outs,
the arrangement carries much less weight — a summary is legible at any of these
sizes. Combined with #956 making quadrant and row identical at three players,
that setting is a candidate for retirement. Deliberately not decided here.

**`cardSize`'s opponent ramp applies only to expanded panels** once summaries are
the default. `--card-scale-opponent` may stop earning its complexity; same note,
same "revisit later".

**e2e breaks when the default flips.** `game.spec.ts` and `players.ts` assert
opponent board selectors that a summary does not have. They need either updating
or forcing to `opponentDetail: "full"` — probably both, so the summary path gets
its own spec.

## 8. Alternatives considered

**Full-screen focus with wheel navigation between boards.** One board fills the
screen, the wheel scrolls between seats, avatars jump. Rejected: Commander asks a
player to track opponents *continuously*, and a design whose default state shows
one board hides exactly what the format runs on. Three concrete breakages beyond
that — the wheel is already taken (`.rail` sets `overflow-y: auto`, and
battlefield rows scroll at `large` card size); targeting and combat would need
off-screen seats scrolled to mid-prompt; and `CombatArrows` measures both
endpoints against `boardEl`, so an attacker and a defender never on screen
together have no arrow to draw. The avatar-as-jump-target idea survives, as
pinning.

**Hover to magnify.** Rejected as a cursor-crossing trap: the pointer traverses
opponent panels constantly on the way to the stack, the phase widget, a target.
`hoverDelayMs` already exists, which suggests hover timing has bitten this
codebase before.

**Auto-magnify when something relevant happens.** Rejected because it steals focus
mid-interaction — a trigger goes on the stack while the player is clicking a land
and the board moves under them. The attention strip is the right surface for
"something happened over there". Note this is *not* the same as §3's triggers,
which fire on actions the viewer initiated or on turn boundaries.

**Threat ranking in the summary.** See §4.

**Raising the hand-peek fraction.** Considered as a fix for #956 and rejected. The
`0.62` self peek is a documented vertical budget (`cre ≤ 0.42·H − 30`, #858) and
raising it shrinks the creature row; the hover lift already reveals full cards.
The missing fix was width, not height.

## 9. Status note

Proposed rather than Accepted: the derivations, the component and the settings
have landed, but nothing is wired into `Board.svelte` yet (that waits for the
#956 layout fix to merge, since both touch the same file), so no game has been
played against this. It becomes Accepted when the wiring lands and §6 has an
answer.
