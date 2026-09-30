# ADR 0105 — Highlighting every legal action

**Status:** Proposed · 2026-09-30 · S48 — Client robustness and the surfaces that lie
**Issue:** [#1789](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1789). Relates to [#1621](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1621) (Vivi as a mana source: this ADR is its "make it obvious" half; the auto-tap half is being built separately as a last-resort tier) and [#1622](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1622) (casting from exile).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-09-30. I ran `git fetch --all --prune` and listed every `docs/decisions/` file name ever committed on any branch, local or remote (`git log --all --name-only -- docs/decisions/`). The highest number in use is 0102; 0103 and 0104 are reserved for ADRs being written in parallel. No branch has 0105, so this one takes **0105**.
**Builds on:** [ADR 0033](0033-ai-bot-seat.md) §1 (the legal-move enumerator, "shared with the client"), [ADR 0009](0009-smart-priority-autopass.md) and its #1307 amendment (autopass reads the same move list), [ADR 0047](0047-keyboard-shortcuts.md) §4 ("a shortcut is enabled exactly when the move is legal, and the server says so"), [ADR 0066](0066-granted-cast-and-play-permissions.md) (per-holder cast permissions), [ADR 0062](0062-abilities-and-special-actions-from-the-hand.md) (special actions), [ADR 0093](0093-abilities-granted-to-other-permanents.md) (ability `ref`s), [ADR 0037](0037-unimplemented-card-signal.md) (the unimplemented flag) and [ADR 0077](0077-opponent-board-summary.md) (opponent summary panels).

---

## Context

On 2026-09-30 the owner asked for this:

> in general I want client highlights on playable cards, activated abilities, etc so a player has an easy view of all their legal actions.

The trigger was #1621. A player had Vivi Ornitier on the battlefield with a free, once-per-turn mana ability. Nothing on the board said it was available, so they cast a spell without it and were told `insufficient_mana`. The rules were fine. What failed was the board: it did not tell the player what they could do.

### The server already answers the question

`internal/legal` enumerates every move a seat may make right now (ADR 0033 §1). The protocol projection runs it for every seat on every frame (`enumerateLegalMoves`, `protocol/view.go`) and ships each seat **its own** list as `GameView.legal_moves`. The list is projected out of an unexported per-seat map by `FilterViewFor`, so it never reaches an opponent, a spectator, the crash dump or the replay log (`TestLegalMovesOwnSeatOnly`, `…NeverLeakThroughTheRawView`, `…SpectatorGetsNone`).

The list has these properties:

- **It is exact about what it lists.** Every move would be accepted by `actions.Dispatch` if sent right now. The tests hold this by dispatching every enumerated move against a clone.
- **It pays for what it offers.** A cast move exists only if the auto-tapper can fund it (`canPayExcluding` runs a real auto-tap solve). So the list answers "can I afford it", which `castable_here` deliberately does not (CR 601.2g lets the caster tap afterwards, #695).
- **It is empty unless the seat owes a decision.** That means holding priority, owing a pending choice, a combat declaration, the mulligan or a cleanup discard. On a quiet frame the field is omitted and costs 0 bytes (`TestLegalMovesQuietFrameIsFree`). The client already reads an absent list as "no information", never as "nothing is legal" (`timing.ts` `movesFor`).
- **It is capped.** There are two caps. The enumerator expands at most 12 moves per source (`MaxExpansionPerSource`). The wire keeps at most 48 moves per seat (`legalMovesWireCap`). Past 48, `capLegalMoves` keeps the **first** move of each `(source, kind, targets_stack)` tuple, so every card that had a move still has one. What is lost is the alternatives within a tuple. That includes a card's second activated ability, a second castable face, and a second zone's price.
- **It does not list sandbox verbs.** It never lists `activate_loyalty`, `move_card`, `change_life`, `advance_step` or `finish_blocks`. Catalog loyalty abilities go through `activate_ability` and *are* listed.

### What the client does with it today

The client reads the list for gates and never for highlights:

- `canCastFromHand` takes its verdict from the list and builds only the tooltip from other fields.
- The pass key is live when `hasPassMove` is true (ADR 0047 §4).
- Smart autopass classifies the moves (`responseWindow.ts`).

Nothing uses the list to say "this one is live". Here is the inventory of what the board marks today:

| Surface | Signal read | What it looks like | Positive or negative |
|---|---|---|---|
| Hand card (`Hand.svelte`) | `legal_moves` cast/land, via `canCastFromHand` | `.timing-disabled`: 50% opacity, greyscale, no click, tooltip reason | Negative only. A playable card looks like any other card. |
| Drag-to-cast ghost (#1508) | `dragVerdict` plus auto-tap preview | Gold/red ghost, board highlight of planned sources | Both, but only while dragging |
| Exile strip (#1389) | `castable_here`, `cast_prices`, `exile_play`, then the move list | `.strip-slot.ready` / `.blocked` / `.later`, `.toggle-count.ready` | **The only positive "ready" mark on the board** |
| Zone browser, exile | `exileEntryLegality` | Button `disabled` | Negative |
| Zone browser, graveyard (flashback, escape) | `castable_here` on any face | Button shown or hidden | Hidden, never greyed. Not affordability-aware. |
| Library top (`PileBar`, `libraryTop.ts`) | `castable_here`; special actions filtered to `available` | "play" pill shown or hidden | Hidden |
| Command zone cast hint (`CommandZone.svelte`) | none | Always live | **Never greyed.** A commander you cannot cast looks castable. |
| Mana/ability popover (right-click, `ManaAbilityMenu.svelte`) | Per-row server fields through `abilityBlocked`, plus a client `sorcerySpeedBlocked` | Row `disabled` with a reason | Negative, and only once you have opened it. `sorcerySpeedBlocked` keys on the printed `sorcery_speed`, not the server's `timing_closed`, so it can disagree with the admin context menu. |
| Card context menu (admin overrides only) | Same row fields; `special_actions[].available`; `canActivateLoyalty` | Row `disabled`, "not right now" | Negative. **Special actions are reachable only here**, behind `gameplay.adminOverrides`. |
| Permanent with a usable ability | — | Nothing | **Gap.** Vivi (#1621). |
| Attack / block candidates | Client-derived `combatMode`; `attackAll.ts` `attackBlocker` re-derives eligibility from `restrictions`, `defender`, `summoning_sick` and `tapped` | Nothing before declaring. `.attacking` (red), `.blocking` (blue), `.selected` after. `MUST ATTACK` / `MUST BLOCK` badges. | **Gap**, plus a second copy of the rules |
| Targets (`targeting.ts`) | Server `legal_targets` | `.targetable` green ring, `.picked` yellow, "· N legal" in `TargetingBanner` | Positive, and already right |
| Pending choices (`ChoicePromptModal`) | `pending_choices` | Modal with per-control `disabled` | Already right |
| Priority | `turn.priority_holder` | Gold avatar ring, `next` button enabled | Already right |

Three facts from the survey shape the design:

1. **The positive half is missing, not the data.** Every gap in the table is a card whose move is already in `legal_moves`.
2. **Two readers re-derive rules the list already answers:** `attackBlocker` for combat, and `sorcerySpeedBlocked` in the popover. They are the drift class ADR 0033 §1 was written to remove.
3. **Some legal actions have no ordinary route to perform them.** Special actions (foretell, suspend, plot, turn face up) sit behind the admin context menu. Every ability menu is right-click only; there is no long-press anywhere, so a touch player cannot open one. A highlight that points at something the player cannot then do is worse than none.

### Measurements

These were taken on 2026-09-30 in `golang:1.22` on the workstation (16 threads) with `go test ./internal/protocol -run TestLegalMoves -bench BenchmarkViewOfGame -benchmem`. The board is `busyTable`: four seats and six creatures each.

- `ViewOfGame` takes **1.75 ms/op** (1.79 MB, 8,262 allocs). Of that, the four-seat enumeration is **0.60 ms/op**. It is already paid on every `Room.Apply`, highlights or not.
- On the frame-budget boards, `legal_moves` is **14.8 KB for 40 moves**, about 370 B a move, against a 24 KiB budget. The frame grows from about 19–25 KB to 34–39 KB.
- On an idle seat's frame, `legal_moves` costs **0 B**.

---

## Options

### A. Client only: read `legal_moves` as it is

Build a per-frame map `source → moves` from the list and draw from it. No server change.

- **For:** zero wire or server cost; ships in one client PR; one answer shared with the bot, autopass and the shortcuts.
- **Against:** the wire cap. Past 48 moves the list keeps one move per `(source, kind)`, so it can say "this permanent has an ability", but not **which** row in its menu. It also cannot say "this commander is castable for its printed cost *and* its alternative". A count badge ("3 abilities") would be wrong exactly on the busy boards where it matters most. The card-level highlight survives the cap; the row-level highlight does not.

### B. New per-card stamps: `can_activate_now`, `can_cast_now`, `can_attack`

Add booleans to `CardView` and the ability rows, computed by new predicates in the projection.

- **Rejected.** This is a second answer to "is this legal now", next to the enumerator. It is the exact shape of #1012 (the view and the enumerator answering "which prices may this cast claim" with two functions) and of #544 (a bot offered what the engine refuses). Most of these stamps would also be private, and then need #1037's per-holder machinery on every surface. The enumerator already runs; stamping from its output is option C.

### C. A digest of the same enumeration, next to the list — recommended

In `enumerateLegalMoves`, **before** `capLegalMoves`, fold each seat's full move list into a compact per-source digest. Ship it beside `legal_moves` through the same unexported-map projection, so it is own-seat only by construction.

- **For:** one enumeration and one answer. Exact past the cap, down to the ability row. It costs a few hundred bytes more on priority frames and nothing on quiet frames. The client joins it to the rows by the `ref` both already carry (ADR 0093).
- **Against:** a new wire field and its tests. That is small.

### D. Client re-derivation from the row fields

Extend `abilityBlocked`: `tap_cost` plus `tapped` plus `summoning_sick` plus `timing_closed` plus priority, and so on.

- **Rejected.** It is the S13.3 timing layer that S31 sub-PR 2 deleted. It cannot answer affordability. It would also be wrong silently on every new cost component.

**Recommendation: C.** Highlight from the enumerator's own output, and add the digest so the output survives the cap. If the owner would rather not touch the server at all, A is an acceptable first step: the card-level highlights ship identically, and only the row-level precision and the counts wait.

---

## Decision

### 1. Source of truth: `GameView.legal_actions`, the enumerator's digest

`legal_moves` stays as it is: the bot, autopass and the shortcuts read it, and ADR 0033's contract is about it. Beside it goes a digest of the **uncapped** list:

```jsonc
"legal_actions": {
  "pass": true,                      // a pass move exists (what hasPassMove answers today)
  "sources": {                       // keyed by card instance ID
    "<id>": {
      "kinds": ["cast", "activate", "mana"],   // distinct legal.Kind values, in enumeration order
      "moves": 7,                              // uncapped move count for this source
      "abilities": ["own:0", "grant:x:0:1"],   // activate moves: the row refs (ADR 0093)
      "mana_abilities": ["own:1"],             // mana moves: the row refs
      "special_actions": ["foretell"],         // special_action moves: the kinds
      "zones": ["hand", "graveyard"],          // cast/land moves: from_zone values
      "faces": [0, 1],                         // cast moves: castable face indices
      "attack_targets": ["<player or permanent id>"],  // attack moves: target
      "blocks": ["<attacker id>"]              // block moves: the attackers this blocker may block
    }
  }
}
```

The rules for the digest:

- **Built from the list, never beside it.** One function (`digestLegalMoves`) reads `[]legal.Move` and nothing else. It reads the fields `Params` already names: `ability_index`/`ref`, `from_zone`, `face`, `kind`, `attacker`, `target` and `blocker`. No new game read, and no new lock.
- **Own seat only, by construction.** It lives in the same unexported per-seat map as `legal_moves` and is promoted by `FilterViewFor` for the viewer's key only. The unfiltered view carries none. Spectators and admins get none. It reveals nothing new, because the viewer already receives the list it summarises.
- **Absent means "no information".** When there is no list, there is no digest, and the client highlights nothing. It does not dim anything new either (§3).
- **Omitempty throughout.** A quiet frame costs 0 bytes, as now.
- **Named `legal_actions`,** to sit beside `legal_moves` and say what it is for.

A card joins the digest by instance ID. A menu row joins it by `ref`. `ActivatedAbilityView` and `ManaAbilityView` both carry `ref`, and it is the same string the move's `Params.ref` names. `choice` and `mulligan` moves are not digested: their surfaces (§2) already read `pending_choices`.

The client gets one pure module, `lib/legalActions.ts`. It takes a frame (plus `legal_moves` as a fallback, see sub-PR 2) and returns lookups: `isReady(cardID)`, `readyAbilityRefs(cardID)`, `readyManaRefs(cardID)`, `readySpecialActions(cardID)`, `canAttack(cardID)`, `blockableAttackers(cardID)` and `readyCount(zone, seat)`. It is computed once per snapshot in a `$derived`, so every card reads it in O(1). Nothing else in the client decides whether an action is legal. `attackBlocker`'s eligibility half and the popover's `sorcerySpeedBlocked` gate move onto it (§2, sub-PRs 3 and 5), which deletes the two re-derivations the survey found.

### 2. What gets highlighted, and how

There is **one "ready" treatment** in one colour token, `--ready`. It is a cool cyan, chosen to be distinct from the green `.targetable`, the yellow `.picked`, the red `.attacking`, the blue `.blocking` and the gold priority ring. Small **pips** carry the kind by shape, not by colour. A player learns one colour ("I can do something with this") and reads the pip only when they need the detail.

| Surface | Digest read | Treatment |
|---|---|---|
| Hand card castable or land playable | `kinds` ∋ `cast` or `land`, `zones` ∋ `hand` | Ready ring (2px outline plus a soft outer glow). The existing `.timing-disabled` dim stays for cards with no move. A land gets the same ring. |
| Hand card with only a special action (foretell, plot, suspend) or a hand ability (cycling, a Spirit Guide's mana) | `special_actions`, `abilities` or `mana_abilities` on a hand card | Ready ring plus the pip for the kind. The card is **not** dimmed. It is not castable, but it is not dead either. |
| Permanent with an activatable ability | `abilities` non-empty | Corner **bolt pip**, with a count when there are two or more, and the ready ring. |
| Permanent with a usable mana ability | `mana_abilities` non-empty | Corner **drop pip**, with no ring and subject to §4's noise rule. |
| Face-down permanent that can be turned face up | `special_actions` ∋ `turn_face_up` | Special-action pip plus ring |
| Ability and mana menu rows (popover and context menu) | row `ref` ∈ `abilities` / `mana_abilities` | Ready rows get the ready accent and sort first. Blocked rows keep today's `disabled` plus reason, which still come from the row fields because those supply the *sentence*, as `canCastFromHand` does. |
| Exile strip, zone browser (exile, graveyard), library-top pill, command-zone cast hint | `kinds` ∋ `cast`/`land` with the matching `zones` | Ready ring on the card, or ready accent on the button. The **command-zone hint is greyed when there is no move**, which closes the "never greyed" gap. The graveyard button is shown-but-disabled rather than hidden when `castable_here` says yes but the list says no, so "you cannot afford it" is visible. |
| Pile counts (graveyard, exile toggle, library top) | `readyCount` | A small "N ready" count on the pile. The exile strip's `.toggle-count.ready` becomes this. |
| Attack candidates (declare attackers, your turn) | `attack_targets` non-empty | Ready ring on each creature that may attack. While a creature is selected, the defenders in its `attack_targets` light (today's `.identity.targetable` already does the player half). |
| Block candidates (declare blockers, defending) | `blocks` non-empty | Ready ring on each creature that may block. While a blocker is selected, the attackers in its `blocks` get the ready ring. |
| Targets, pending choices, pass | — | **Unchanged.** Targeting's green ring and the choice modal are already the right answer. The digest's `pass` only replaces `hasPassMove`'s scan. |

What highlights do **not** say:

- **They do not mean "the engine automates this card".** A card with ADR 0037's `unimplemented` flag is still castable and still gets the ring; its `manual` chip stays separate text.
- **They do not mark "worth doing".** A seat with an untapped Sol Ring and nothing to spend on it has a legal mana move. §4 decides what is noise; this ADR never ranks moves.
- **They never mark opponents' cards,** and never spectators' views. There is no digest for them.

**Room doors (unlock, CR 709.5)** have no engine support yet: `game.SpecialActionKind` is foretell, suspend, turn face up and plot. When unlock lands it is one more `special_action` kind. The pip falls back to a generic special-action glyph for a kind the client does not know, so it lights with no client change.

### 3. Timing: only while you owe a decision

Highlights render **only on a frame that carries `legal_actions`**, which is exactly when the seat owes a decision. On an opponent's turn with nothing on the stack and no priority for you, nothing glows. When the priority cursor reaches you, what you can do lights. This is recommended because:

- It is the only honest answer the server has. A "you could activate this once you get priority" tier would need a hypothetical enumeration ("as if this seat held priority now"). That is a second, speculative answer. It would also be wrong often, because priority arrives after something has changed.
- A board that glows all the time stops meaning anything.

Two details:

- **Autopass flicker.** Smart autopass decides on the frame it receives. If its verdict is "pass", the highlights for that frame are suppressed, so a window the client is about to skip does not flash. `autopassDecision.ts` already computes the verdict. `Game.svelte` passes a `highlightsLive` boolean down with the frame.
- **The negative half stays as it is.** Today's dimming (hand `.timing-disabled`, disabled menu rows) keeps its current rules, including its permissive reading of an absent list. The toggle in §6 turns off only the positive treatment.

### 4. Mana abilities and noise

If every mana move lit up, every untapped land would glow on every priority window, and the one signal #1621 needs (Vivi's free ability) would drown. The recommended rule, which is open question 2:

- **Lands' mana abilities get no pip.** Whether a land is untapped is already visible, and every player knows a land makes mana.
- **A mana ability gets the drop pip when it is not the obvious kind:** a nonland source, a source in hand (a Spirit Guide), or an ability without a `{T}` cost (Vivi, a Treasure's sacrifice, a Lotus Petal). All of these are facts on the `ManaAbilityView` row (`tap_cost`, `sacrifice_cost`) and on the card (`type_line`). This is presentation, not legality: legality still comes only from the digest.

### 5. Performance

- **Server.** The digest is one pass over a list the projection has already built, so it is O(moves), well under the 0.60 ms enumeration. Sub-PR 1 extends `BenchmarkViewOfGame` with the digest and records the delta in the PR.
- **Wire.** An entry is about 60–150 B, depending on refs and IDs. A busy priority frame with 15–25 sources is about 1.5–3 KB. A budget test pins it at **4 KiB** on the same worst boards `TestLegalMovesFrameBudget` uses. Quiet frames stay 0 B.
- **Client.**
  - One map build per snapshot; each card and row does an O(1) lookup.
  - The ring is a static `outline` plus a pseudo-element glow, with no `box-shadow` animation across forty cards.
  - The optional single pulse when priority *arrives* is one opacity transition on the pseudo-element. It is off under `reduceMotion` and `prefers-reduced-motion`.
- **The bot's enumeration is unchanged.** The runner enumerates in-process as it does today (`runner.go`).

### 6. Settings

Add one toggle: **`gameplay.highlightLegalActions: boolean`, default on.** This bumps the settings schema from v14 to v15, with a migration that adds the key as `true`. It goes on the Gameplay tab next to the autopass settings, with the usual "✓ saved" pattern. It is a per-player, browser-local preference in `localStorage` like every other client setting, not a table setting (ADR 0075 is house rules).

**Off** removes the ready rings, pips, counts and menu-row accents. It does not remove today's dimming or disabled rows, because those are gates (the click is withheld), not decoration. Whether mana pips get their own sub-toggle is open question 2.

### 7. Accessibility and phone width

- **Not colour alone.** Readiness is carried by the outline, and the kind is carried by pip **shape** (bolt, drop, special-action star, sword for attack, shield for block).
  - The `high-contrast` theme draws the ring as a solid 3px outline with no glow.
  - `colorblindPalette` swaps `--ready` for the alternate set.
  - The ring is outset so it never overlaps the keyboard focus ring, which is inset under `alwaysShowFocus`.
- **Screen readers.** A ready card's accessible name gains one phrase: "castable", "playable land", "has an ability you can activate", "can attack" or "can block". A menu row's name gains "available". The phase display's live region says "N actions available" once, when priority arrives, not on every frame.
- **Phone width.** The board has no width media queries today; card size is a CSS variable. The ring and pips scale with `--card-size`, and pips have a 16px minimum hit and draw size. At the `small` card size, counts are hidden and the pip shows alone.
- **A touch route to the highlighted action.** Ability and special-action menus open only on right-click today, and there is no long-press in the client. Tapping a pip therefore **opens the same popover** a right-click opens. This is the smallest change that makes every highlight actionable on a phone. It is in sub-PR 4, and it is open question 6 whether it belongs here or in its own issue.
- **Special actions reach the ordinary menu.** Foretell, suspend, plot and turn face up move from the admin-only context menu into the default popover, as rows next to the abilities. Without that, the pip in §2 would point at nothing for anyone not running admin overrides.

### 8. The bot does not change

- No policy reads `legal_actions`. Policies pick from `legal.Move` lists they enumerate in-process at full fidelity.
- The digest will appear in a bot's `Input.View`, because that is `ViewOfGameFor`, byte-identical to a human's frame (ADR 0033). So it also appears in decision logs and harvested positions. That is additive and `omitempty`. Old positions parse unchanged.
- The model prompt renderer picks fields; it does not marshal the view, and it does not add this one. A test pins that a rendered prompt is byte-identical with and without the digest.
- `heuristic/imports_test.go` is untouched: nothing new is imported.

### 9. Test strategy

**Server (`internal/protocol`)**

- `TestLegalActionsOwnSeatOnly`, `…NeverInTheRawView` and `…SpectatorGetsNone`, copying the three `legal_moves` privacy tests.
- `TestLegalActionsAgreeWithTheList`. On `busyTable` and on the pathological board that trips `legalMovesWireCap`:
  - every source with a move in the **uncapped** enumeration has an entry;
  - every entry's refs, zones, faces, targets and blockers name at least one enumerated move;
  - `moves` equals the uncapped count.

  This is the test that proves the digest survives the cap and option A does not.
- `TestLegalActionRefsMatchTheRows`: every `abilities` / `mana_abilities` ref is the `ref` of a row on that card's `activated_abilities`, `zone_abilities`, `mana_abilities` or `zone_mana_abilities`. That is the client's join key.
- `TestLegalActionsFrameBudget` (4 KiB) and the quiet-frame zero-byte test, with measurements logged like the existing budget test.
- A Vivi case: on the controller's own turn with priority, Vivi's `ref` is in `mana_abilities`; on an opponent's turn, and after it has been used this turn, it is not.

**Contract fixture.** Generate `server/internal/legal/testdata/legal_actions_agreement.json` from a few boards. `client/src/lib/legalActions.test.ts` reads it, the way `timingAgreement.test.ts` reads `timing_agreement.json`, so the client's lookups and the server's digest are checked against the same bytes.

**Client, pure (`vitest`, node)**

`legalActions.ts` lookups; the absent-list rule; the mana-noise rule; the toggle; and the autopass suppression.

**Client, render (`vitest`, jsdom via `lib/test/render.svelte.ts`)**

- A hand card with and without a move: ready class, accessible name, and `.timing-disabled` unchanged.
- A permanent with two abilities where one is ready: pip count and row states.
- The command-zone hint greyed when there is no move.
- Pips tap-open the popover.

**E2E (`tests-e2e/`, one spec)**

On your own main phase, the land and a castable spell in hand have the ready class and an uncastable one does not. After passing to the opponent's turn, nothing has it. It runs in the E2E workflow, since Playwright does not run on the workstation.

**Not automated:** the look. Each client sub-PR attaches screenshots at the three card sizes and at 390px width.

### 10. Delivery

Each sub-PR is small, lands on `develop` on its own, and leaves the board coherent.

1. **Server: the digest.** `legal_actions` on `GameView`, `digestLegalMoves`, the privacy, agreement, ref-join and budget tests, the agreement fixture, and `docs/protocol.md`. No client change.
2. **Client: the module, the setting and the cast surfaces.**
   - `legalActions.ts` reads `legal_actions`, and falls back to scanning `legal_moves` when the field is absent, so it also works against an older server.
   - `gameplay.highlightLegalActions` (settings v15).
   - The ready ring on hand cards, the exile strip, the zone browser, the library-top pill and the command-zone hint (which is greyed at last); pile counts; autopass suppression.
   - `hasPassMove` reads `legal_actions.pass`.
3. **Client: permanents and menus.** Bolt and drop pips with §4's rule. Ready accents on popover and context-menu rows. `ManaAbilityMenu`'s `sorcerySpeedBlocked` gate is replaced by the digest. This is the PR that closes #1621's client half.
4. **Client: special actions and the touch route.** Special-action rows in the default popover. Special-action pips on hand cards and face-down permanents. Tap-a-pip opens the popover.
5. **Client: combat.** Attack and block candidate rings, and selection-driven defender and attacker rings. `attackAll.ts` reads eligibility from the digest instead of `attackBlocker`'s re-derivation.
6. **Accessibility and phone pass, and the e2e spec.** Accessible names, the live-region line, high-contrast and colour-blind variants, the phone-width screenshots, and `tests-e2e/legal-highlights-1789.spec.ts`.

Sub-PRs 2–6 depend on 1 only for row precision. Sub-PR 2's fallback means the client work can start in parallel.

---

## Consequences

- **Good.** The owner's view is met: on your priority, everything you can do is marked, and each mark comes from the same function that decides whether the server accepts the click. #1621's Vivi case shows as a drop pip. The command-zone hint stops lying. Two client re-derivations (`attackBlocker`'s eligibility, `sorcerySpeedBlocked`) are deleted. Special actions stop being an admin-only feature. Phone players get a route into ability menus.
- **Cost.** One wire field of roughly 1.5–3 KB on priority frames and 0 B otherwise. One server function and its tests. A settings migration.
- **Risk.** A highlight that is wrong is worse than none. The mitigation is structural: the digest is the enumerator's own output, and the enumerator is held to dispatch-soundness by its tests. A card the enumerator misses (a sandbox verb) gets no ring but is not dimmed either, so the failure mode is "unmarked", never "marked but refused".
- **Not solved here.** Whether "castable from exile" should be **public** for other players (#1622's second half) is a question about `exile_play`'s visibility, not about highlights. The digest stays own-seat only.

## Out of scope

- Ranking or suggesting moves ("you should cast this"). That would be a coach, like ADR 0076's tutorial, not a rules readout.
- Highlights for sandbox verbs the enumerator does not list: `activate_loyalty`, moving cards by hand, `change_life`.
- Room doors: they have no engine support yet. §2 covers them when they land.

---

## Open questions for the owner

1. **Timing.** Should highlights show only while you owe a decision (priority, a combat declaration, a choice), or also a dimmer "you could, once you get priority" tier at other times? *Recommended: only while you owe a decision (§3). The server has no honest answer for the other tier.*
2. **Mana sources.** Should ordinary lands' mana abilities be marked? *Recommended: no. Mark only nonland, from-hand or no-`{T}` mana abilities (Vivi, Spirit Guides, Treasures) (§4).* And should mana pips get their own on/off sub-toggle?
3. **One colour or several.** Should there be one "ready" colour with pip shapes for the kind, or a colour per kind (cast, ability, mana, combat)? *Recommended: one colour plus pips (§2).*
4. **Default.** Should the toggle default to on for everyone, existing players included? *Recommended: on.*
5. **Autopass.** Should highlights be hidden on a frame smart autopass is about to pass, or always drawn for the instant the frame is on screen? *Recommended: hidden (§3).*
6. **Touch route.** Should tapping a pip opening the ability popover be in scope here (sub-PR 4), or split into its own issue about phone access to ability menus? *Recommended: in scope. Otherwise the highlight points at an action a phone player cannot take.*
7. **Server digest or client only.** Should sub-PR 1 add `legal_actions` (option C), or should the client read `legal_moves` as it is and accept that row-level highlights and counts go vague on boards past 48 moves (option A)? *Recommended: C.*
