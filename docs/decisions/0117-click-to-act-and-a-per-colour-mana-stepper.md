# ADR 0117 — Click to act, and a per-colour mana stepper

**Status:** Proposed · 2026-10-04 · S59 — Automated table: clicks that act, payment that counts (tracker [#2189](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2189))
**Issues:** [#2187](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2187) (this change). Strict payment by default is [#2188](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2188), which gets its own ADR (0118).
**Owner decisions:** the direction of 2026-10-04 and the four answers given the same day, quoted under [Owner direction and answers](#owner-direction-and-answers-2026-10-04). They are binding. This ADR also makes six smaller calls the answers did not cover. They are listed under [Calls made here](#calls-made-here) so the owner can overturn any of them in review. The ADR is accepted when its PR merges.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-04. I ran `git fetch --all --prune` and listed `docs/decisions/` on all 36 remote heads: `origin/develop`, `origin/main`, the `feat/1107-*`, `feat/1112-*`, `feat/1117-*` and `feat/batch*` card branches, and the other chore, docs, feat, fix, repro and wip branches. The highest number on any of them is 0116, on `origin/develop` and `origin/feat/1548-heuristic-gang-blocks`. No branch has 0117, so this ADR takes **0117**. 0118 is reserved for #2188 by the tracker; it does not exist on any branch yet.
**Amends:** [ADR 0020](0020-activated-abilities.md) Consequences (the activation popover also opens on a left-click). [ADR 0093](0093-abilities-granted-to-other-permanents.md) Decision 8, "Click behaviour" (the sandbox tap moves from the context menu to the popover's Sandbox section; a granted ability is clicked through the same rule as any other). [ADR 0040](0040-mana-pipeline.md), the 2026-09-24 amendment (#1443), its Client paragraph (two or more colour picks are a stepper, not a list of combinations). [ADR 0076](0076-tutorial.md) §2.1, step 7's hint.
**Builds on:** [ADR 0011](0011-mana-pool-and-auto-tapper.md) (the pool and the auto-tapper, unchanged here), [ADR 0028](0028-admin-context-menu.md) §5 (right-click with admin overrides opens the override menu, unchanged), [ADR 0105](0105-legal-action-highlights.md) §7 (pips open the popover), [ADR 0106](0106-five-small-seams-from-the-s50-rechecks.md) §1 decision 6 (any-player rows), [ADR 0111](0111-action-dock.md) owner decision 5 (card-local menus and the mana picker stay at the card).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

The client was built for a sandbox. A permanent's left-click turned it sideways, because tapping was the one thing every permanent could do. Over time the click grew exceptions: a planeswalker opens its menu (#329), a utility land opens its menu (#368), an untapped mana source taps for mana (#1438), and a permanent with a granted ability opens its menu (ADR 0093). The table is automated now, with almost 4,000 cards in the catalog. A click should do what the card does.

Every claim below was checked on `origin/develop` at `6f1dd6f3`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`).

### Owner direction and answers (2026-10-04)

The direction, verbatim:

> previously we were building the game around it being a sandbox and now we are building it around being automated with almost 4k cards catalogged. That means the game should be smarter and more intuitive. left clicking prompts to activate the abilities unless it is a mana ability in which case it automatically taps for mana (unless a color is needed to be selected). In the case of Vivi they may generate a lot of mana which could have like a blue and red ticker to resolve how much blue and red is being tapped at once

The answers:

1. **A plain click on a permanent with no usable ability does nothing.** That covers a vanilla creature, and a tapped creature whose only abilities cost {T}. Tap and Untap move into the card's ability popover as a Sandbox row on every permanent. Alt-click keeps its raw tap.
2. **A card with a usable mana ability and another usable activated ability opens the popover,** with the mana rows first. It does not tap for mana at once.
3. **The stepper** is for any activation that adds two or more mana with a colour choice: one −/+ row per colour, with a running "N of N" count, at the card. One mana, or one colour, keeps the plain buttons. This removes the 12-combination cap.
4. **Strict mana payment by default** is wanted, but in a separate ADR, 0118 ([#2188](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2188)). It is out of scope here.

### The rules

- **CR 602.2:** "Only an object's controller … can activate its activated ability unless the object specifically says otherwise." The click rule acts only on the viewer's own permanents, except for any-player rows (ADR 0106).
- **CR 605.1a:** an activated ability is a mana ability if, among other things, "it could add mana to a player's mana pool when it resolves." A power-0 Vivi Ornitier still has a mana ability. Activating it is legal. It adds nothing.
- **CR 605.3a:** a player may activate a mana ability "whenever they have priority, whenever they are casting a spell or activating an ability that requires a mana payment, or whenever a rule or effect asks for a mana payment". The engine's `activate_mana_ability` has no priority gate, and this ADR adds none.
- **CR 605.3b:** a mana ability "doesn't go on the stack … it resolves immediately after it is activated." Nothing can be undone after the click, so the colour split has to be decided before the activation is sent.
- **CR 106.12:** to "tap [a permanent] for mana" is to activate a mana ability "that includes the {T} symbol in its activation cost." Vivi's "{0}" ability does not tap Vivi. Whether a row can be used is a question about its cost, not about the card's tapped state.
- **CR 302.6:** a creature's ability with {T} in its cost "can't be activated unless the creature has been under its controller's control continuously since their most recent turn began."
- **CR 602.1b:** "Activate only during your turn and only once each turn" is an activation instruction. Vivi carries both as its `Condition`.
- **CR 106.1b:** there are six types of mana: white, blue, black, red, green and colorless. A stepper row is one of these.
- **CR 106.4:** added mana goes into the pool and "can stay in the player's mana pool as unspent mana" until the step ends. What the stepper adds floats like any other mana.
- **CR 106.5 and 106.7:** an ability that would produce mana of an undefined type produces none, and a permanent that "wouldn't produce any mana under these conditions" has no type of mana it could produce. Exotic Orchard with no opposing lands is that case.
- **CR 701.26a and 701.26b:** "Only untapped permanents can be tapped." "Only tapped permanents can be untapped." The Sandbox row shows whichever one applies.
- **CR 116.2b:** turning a face-down creature face up is a special action. The popover lists it (ADR 0105 §7), so the click counts it.

### What exists

**The click router.** `PlayerPanel.svelte` `handleCardClick` (`:490`) tries, in order, the targeting intercept, a combat select on the viewer's own creature, an attack on a listed planeswalker or battle, and a block on an incoming attacker. Then it asks `battlefieldClickIntent` (`contextMenu.logic.ts:357-404`) and switches on the answer (`PlayerPanel.svelte:550-566`):

| Case, in order | Intent | What happens |
|---|---|---|
| `canOverride` fails (not yours, not admin) | `abilities` if an any-player row, else `none` | the heavy override menu, or nothing |
| Alt-click | `tap` | raw `tap` / `untap` |
| planeswalker | `abilities` | `openCardMenu`: the heavy `CardContextMenu` |
| a granted activated ability (ADR 0093 Decision 8) | `abilities` | the heavy override menu |
| `manaClick && !card.tapped && mana_abilities` (#1438) | `mana` | `clickForMana` → `manaClickPlan` |
| non-creature land with `activated_abilities` (#368) | `abilities` | the heavy override menu |
| anything else | `tap` | raw `tap` / `untap` |

`canOverride` (`:258`) returns true for any admin, so an admin's left-click on another seat's permanent raw-taps it. Only the viewer's own panel wires `activateManaAbility`, so an admin never activates mana on another seat (comment at `PlayerPanel.svelte:547-549`). Keyboard Enter and Space reach the same router (`Card.svelte` `handleKeydown`, `onClick`). Token groups' "Use this one" also calls `handleCardClick` (`PlayerPanel.svelte:783`, `TokenGroupModal.svelte:194-199`).

**The two menus.** `"abilities"` opens the heavy override menu (`CardContextMenu`, ADR 0028), whatever the `adminOverrides` setting says. Right-click opens it only with that setting on (`Card.svelte:510-523`). Otherwise right-click, and a tap on a pip (ADR 0105 §7, `Card.svelte` `openFromPip`), open the light popover `ManaAbilityMenu`: special actions, then mana rows, then activated rows. Those two paths, and no left-click path, emit the tutorial event `ability-menu-opened` (`Card.svelte:522`, `:549`). The popover has a "Tap (no mana)" row only for an untapped permanent with a mana ability, on the viewer's own panel (`ManaAbilityMenu.svelte:313-330`, wired at `Card.svelte:1055`). It has no plain Tap or Untap otherwise. `hasMenu` (`Card.svelte:303-307`) is false for a vanilla creature, so its right-click does nothing.

**Three judgements of "can this row be used".**

1. The popover's `abilityBlocked` (`ManaAbilityMenu.svelte:161-171`): a {T} cost on a tapped card, summoning sickness, the row's `timing_closed`, then the shared `abilityBlocked` (`contextMenu.logic.ts:709`) with no loyalty context, then the digest for sorcery-speed and any-player rows. With no `legal_actions` digest, which is every frame in which the seat owes no decision (`protocol/view.go:211-217`), only the row fields judge (`ManaAbilityMenu.svelte:159-160`).
2. The override menu's rows (`abilityItems`, `contextMenu.logic.ts:858`): the shared `abilityBlocked` with a full `LoyaltyContext`, plus the card-level restrictions `cant_activate` (Arrest, `:936`) and `cant_activate_mana` (`:885`). For a planeswalker with no catalog abilities the same menu adds manual loyalty rows (`loyaltyAbilityItems`, `:1034`): "+1: activate a loyalty ability" and so on, each an `activate_loyalty` with a delta, greyed by `canActivateLoyalty`. Only the override menu has them.
3. The mana picker's `blockedReason` (`manaSource.ts:85-92`): `cant_activate_mana`, `exhausted`, `condition_unmet`, `adds_no_mana`. It deliberately skips a tapped source and summoning sickness, and sends the activation for the server to refuse (#1438).

So the popover misses Arrest's restriction and a planeswalker's "already activated this turn", and the picker misses summoning sickness. The client also never reads the row-level `cant_activate` string the server stamps on both ability views (#1210, `protocol/view.go:3081`, `:3636`). It is not in `protocol.ts`.

**Mana with a colour choice.** The view stamps `ManaAbilityView.color_options`: one list per picking slot of the output, narrowed (CR 903.4f) and ordered identity-first (#843) by the same function the `mana_pick` prompt uses (`game/mana_color_upfront.go` `ManaAbilityColorOptions`, stamped in `protocol/view.go` `stampManaIdentity`). The activation takes `color` or `colors`, one per picking slot, and refuses a wrong count or a colour a slot does not offer before anything is paid (`actions.go:2046-2057`, `game/mana_color_upfront.go` `validateUpfrontManaColors`). `docs/protocol.md` says clients render the options in the order sent. With no colours, each picking slot queues its own `mana_pick`, answered one at a time in the dock.

The left-click picker (`ManaSourcePicker.svelte`) expands each ability into one button per distinct answer (`manaAbilityOptionsFor`, `colorCombos`, `manaSource.ts:186-257`). Past `MAX_COLOR_OPTIONS = 12` (`:178`) it gives up and shows one plain option, which sends no colours, so the server asks N times. Vivi's two colours stay under the cap until power 12 (power N is N+1 answers), but at power 6 that is already seven buttons to read. A five-colour source passes the cap at two picks (15 answers), so Gwenna, Eyes of Gaea, Cascading Cataracts and a Selvala with power 2 or more all fall back to N prompts today. The right-click popover's mana rows never send colours (`Card.svelte:1045`), so they always take the N-prompt path.

**The "any combination" cards.** All are `CompletenessFull`:

| Card | Output | Picks |
|---|---|---|
| Vivi Ornitier | `{U\|R}` × power, `{0}`, your turn, once per turn | power |
| Orcish Lumberjack | `{R\|G}{R\|G}{R\|G}` | 3 |
| Relic of Sauron | `{U\|B\|R}{U\|B\|R}`, plus "{3}, {T}: Draw two cards, then discard a card." | 2 |
| Gwenna, Eyes of Gaea | `AnyCombinationOfColors(2)` | 2 |
| Cascading Cataracts | `{C}`, or `AnyCombinationOfColors(5)` for {5}, {T} | 5 |
| Selvala, Heart of the Wilds | `AnyCombinationOfColors(greatest power)` | X |
| the filter lands (Mystic Gate) | `{W\|U}{W\|U}` | 2 |
| Burnt Offering | a spell: `{B\|R}` × X at resolution | X (not a mana ability) |

`AnyCombinationOfColors` is `effects/color_choice.go:171-202`. Its counterpart `OneColorOfAmount` ("three mana of any one color", #742) is one pick that adds three, and is not this shape.

**`adds_no_mana`.** `ManaAbilityView.adds_no_mana` is `game.ManaAbilityAddsNoMana` (`game/mutations.go:7226`), which today answers only CR 903.4f: Command Tower and its three siblings under no commander or a colourless one. It returns false for every ability without `NarrowToCommanderIdentity` (`:7227`). The legal-move enumerator also reads it, and skips the move (`legal/abilities.go:1696`). Both client judgements grey such a row with `NO_COMMANDER_IDENTITY` (`contextMenu.logic.ts:597`, `:800`; `manaSource.ts:90`). The popover's button for it is **disabled** (`ManaAbilityMenu.svelte:244`).

**Two bugs this ADR fixes.**

- **A tapped Vivi untaps.** The `mana` intent needs `!card.tapped`. Vivi attacked, so she is tapped, so the click falls to `tap` and untaps her. Her ability has no {T} and was usable the whole time.
- **A power-0 Vivi wastes her turn's activation.** Her `ProducedFunc` returns `""`, so she has no `color_options` and no `adds_no_mana`. `manaClickPlan` sees one live option and activates at once. Nothing is added, and the once-per-turn record is spent.

**Manual tapping elsewhere.** Alt-click; the override menu; the dock's ⋯ Sandbox "Untap all" (`GameMenu.svelte:169`); token groups' "Tap N" / "Untap N". The server's only gate on `tap` / `untap` is `requireCardController` (`actions.go:758`).

**The tutorial** (`client/src/lib/tutorialSteps.ts`). Step 5, `tap-land`, asks for a left-click on a Forest and completes when the pool is non-empty. Step 7, `right-click` (`:275-295`), anchors to `abilityCardID` (`:106-115`), which prefers the newest creature with a menu: in practice the mana creature just cast in step 6, which is summoning sick. Its hint says "Right-click, not left: a left click taps it for mana." It completes on `ability-menu-opened`, and its recover check fires when the count of tapped permanents rises.

**Tests that pin today's routing.** About thirty `battlefieldClickIntent` cases, in `manaSource.test.ts` (11), `loyalty.test.ts` (10), `anyPlayerActivation.test.ts` (6) and `grantedAbilities.test.ts` (3); the click render tests in `manaClick.render.test.ts`; "Use this one" in `tokenGroups.render.test.ts`. The Playwright suite clicks battlefield cards only as targets, which the intercept takes first. `s19-triggers.spec.ts:50-56` has a comment describing the old fall-through to tap.

---

## Decision

### 1. The click rule

The intercepts stay first and unchanged: targeting, the combat selects, attack on a listed target, block. Enter and Space on a focused card follow the same rule. Alt-click keeps its raw `tap` / `untap` wherever `canOverride` holds today, an admin's Alt-click on any permanent included.

After the intercepts:

1. **Usable rows.** Work out the card's usable rows with the one shared predicate of §2. These are its special-action rows, its activated rows (own, granted, loyalty, and the manual loyalty rows of §3) and its mana rows. For a permanent the viewer does not control, only the any-player rows count (ADR 0106).
2. **Any usable non-mana row** (an activated row, a loyalty row or a special action): open the light ability popover at the card, mana rows first, and emit `ability-menu-opened`. This is owner answer 2. A left-click never opens the override menu.
3. **Otherwise, usable mana rows only:**
   - one ability with no colour choice: activate it at once (#1438 as today);
   - one ability with one picking slot: today's colour buttons;
   - one ability with two or more picking slots: the stepper (§4);
   - two or more mana abilities, or any granted one: the picker, which never chooses silently between abilities (ADR 0093). A row in the picker with two or more picking slots opens the stepper in place.
4. **Otherwise: nothing.** No action is sent.

Who the card belongs to is the viewer's seat, not `canOverride`. An admin's plain left-click on another seat's permanent therefore no longer raw-taps it. It opens the popover for an any-player row, or does nothing. Admin tools stay where ADR 0028 put them: right-click with admin overrides on, and Alt-click.

The not-yours case keeps its condition (an any-player row opens, nothing else does), but it opens the light popover, the same one a right-click opens for that viewer (`across`), instead of the override menu.

**The cursor.** When no intercept applies and the rule's answer is "nothing", the card drops its `clickable` class, so it shows no pointer cursor and no hover lift. It keeps `role="button"` and its tab stop. A keyboard player still needs to reach it to open the popover, and the e2e suite selects cards by role. There is no tooltip. Hovering already opens the card zoom, and the ADR 0105 ready ring already marks the cards that can act.

What a click does, before and after:

| Permanent you control | Today | After |
|---|---|---|
| Untapped Forest | adds {G} | adds {G} |
| Tapped Forest | untaps | nothing (Untap is in the popover) |
| Vanilla creature | taps | nothing |
| Summoning-sick Llanowar Elves | sends the activation; the server refuses it | nothing (the row is greyed: summoning sickness) |
| Birds of Paradise | five colour buttons | five colour buttons |
| Mystic Gate, `{W\|U}{W\|U}` row | WW / WU / UU buttons (with `{C}`) | picker: `{C}`, or the stepper for the filter row |
| Vivi, untapped or tapped, power ≥ 2 | a button per split, or untaps if tapped | the stepper |
| Vivi, power 1 | U / R buttons | U / R buttons |
| Vivi, power 0 | activates, adds nothing, spends the turn's use | nothing |
| Relic of Sauron, {3} draw row usable | six combination buttons | the popover, mana row first |
| Rogue's Passage | adds {C} | the popover, mana row first |
| Prodigal Sorcerer | taps | the popover |
| An Equipment in your main phase | taps | the popover (Equip) |
| An Equipment off your turn | taps | nothing (Equip's timing is closed) |
| A planeswalker, catalogued or not | the override menu | the popover, or nothing when no loyalty ability can be activated (one was this turn, or the window is shut) |
| A face-down creature | taps | the popover (turn face up) |
| A land under Squirrel Nest | the override menu | the popover |

### 2. One predicate for "usable"

A single exported function in `contextMenu.logic.ts` decides whether a row can be used. The click rule, the popover's greying, the picker's greying and the override menu's rows all call it, so a click never disagrees with the menu. It takes the row, its kind (mana, activated, special) and a context: the card, `tapped`, `summoning_sick`, the payer's life, the timing words, the two digests (`legal`, `legalGate`), `across`, and a `LoyaltyContext` (the card, the view, the viewer).

A row is not usable when, in this order:

1. the card's restrictions stop it: `cant_activate` for an activated row, `cant_activate_mana` for a mana row;
2. the row's own `cant_activate` string is set (#1210; `protocol.ts` gains the field);
3. it has a {T} cost and the card is tapped (CR 106.12 makes this a cost question: the card's tapped flag matters only when the cost has {T});
4. it has a {T} cost and the card is summoning sick (CR 302.6);
5. anything the shared `abilityBlocked` already says, with the loyalty context, so a planeswalker's rows grey once one was activated this turn and a −N it cannot pay greys;
6. `timing_closed`;
7. `exhausted`, `condition_unmet` (Vivi's turn and once-per-turn gates);
8. `adds_no_mana` (§3);
9. the digest refuses it, for a sorcery-speed or any-player row, exactly as the popover does today.

It does not judge whether the viewer can afford a mana cost. The auto-tapper pays, and ADR 0118 decides what happens when it cannot.

A special-action row is usable when `specialActionItems` leaves it enabled.

The picker's old posture, sending a summoning-sick activation for the server to refuse, ends. The popover has greyed that row since ADR 0020, which put `summoning_sick` on the wire "so the menu can grey the entry instead of failing the click."

### 3. The popover's Sandbox row, and manual loyalty

Every permanent the viewer controls gets a Sandbox section at the bottom of the popover, under a divider. It holds one row: **Tap** or **Untap**, whichever applies (CR 701.26). The row sends the raw `tap` / `untap`, and its title says it is a manual change that adds no mana and activates nothing. On a permanent with a mana ability the row keeps today's label, "Tap (no mana)", and its `data-raw-tap` attribute.

The Sandbox section replaces "Tap (no mana)". `hasMenu` becomes true for every permanent the viewer controls, so a right-click on a vanilla creature now opens the popover with its Tap row. The Sandbox section draws no pip. Token groups' "Tap N" / "Untap N" and "Untap all" are unchanged.

**Manual loyalty rows.** A planeswalker with no catalog abilities also gets the override menu's manual loyalty rows (`loyaltyAbilityItems`), in the popover's activated section, not in Sandbox. They are real activations (`activate_loyalty` with a delta, whose text the players resolve), so they count as usable non-mana rows when `canActivateLoyalty` allows. Today a left-click on such a planeswalker opens the override menu for exactly these rows (#329). Without this, the new rule would leave them reachable only with admin overrides on.

### 4. The per-colour stepper

It lives at the card, inside the anchored picker (ADR 0111 owner decision 5).

**When.** An activation whose `color_options` has two or more lists. A list with one option is a fixed slot (Command Tower narrowed to one colour, CR 903.4f). It is shown as a fixed count and cannot be stepped. If every list has one option, there is no choice, and the ability activates at once like a Forest. An ability with one picking slot keeps its buttons, including a `OneColorOfAmount` slot ("three mana of any one color" is one choice). If a picking slot's option in `produced` carries a count (#742) and the ability also has a second picking slot, the stepper is not used and the picker keeps one button per answer. No catalog card has that shape today.

**Rows.** One row per colour any list offers, in the order the server sent them. The order is the union of the lists, by first appearance. That is identity-first (#843), and `docs/protocol.md` asks clients to keep it. The brief for this ADR suggested WUBRG order; the server's order wins for that reason. Each row has −, the count, and +. The colour name and its mana symbol label the row, never the colour alone.

**The total.** It is fixed: the number of lists. A status line reads "N of N" (for example "3 of 5"). **Add mana** is enabled only when the counts add up to N and are feasible.

**Feasibility.** Lists can differ, so a count vector is valid only if each slot can be given a colour it offers. That is a bipartite matching between slots and colour units. Hall's condition decides it: for every non-empty set T of colours, the count summed over T must not exceed the number of lists that offer any colour in T. There are at most six colours (CR 106.1b), so at most 63 sets. The same test on a partial vector says whether it can still be completed, so **+** is disabled when it would break the condition or the total is reached. **−** is disabled at 0 and on a fixed slot's share. Every card in the table above offers the same list in every slot, so the condition is trivially met for all of them. It matters for a fixed slot beside a wider one, and for any future card with different lists.

**Sending.** Add mana builds `colors[]` in slot order from a matching (augmenting paths over at most a few dozen slots) and sends `activate_mana_ability` with it. The server already takes one colour per slot and checks each, so there is **no server change** for the stepper.

**Start state.** The split last confirmed for this card in this game, held in client memory only, if it is still feasible for the current lists. Otherwise all N on the first colour every list offers. Otherwise each slot on its own first option. The start state always fills the total, so a player who wants the same split again confirms in one press.

**Keyboard.** Up and Down move between rows. Left and Right, or − and +, change the focused row. Enter confirms when Add mana is enabled. Escape cancels and sends nothing.

**A frame that changes the lists.** The stepper re-reads `color_options` on every frame. If the number of lists changes (Vivi's power changed), it resets to the start state and shows the new total. If the server refuses a stale answer, nothing was paid (ADR 0040, #1443 decision 2).

**The cap goes.** `MAX_COLOR_OPTIONS` and the fallback to a plain option are deleted. `colorCombos` stays for one picking slot only, where it is just the slot's list.

**Right-click.** A popover mana row with `color_options` opens the same picker, buttons or stepper, for that ability, instead of sending no colours. The override menu's mana rows are an admin tool and are unchanged. The per-slot `mana_pick` prompt in the dock stays as the server's path for anything that sends no colours: the auto-tapper, the bots, Burnt Offering's resolution and an older client.

**Accessible names.** New names, nothing renamed (AGENTS.md §5, "labels are a contract"):

| Element | Name |
|---|---|
| the stepper's dialog | `Split N mana from <card name>` |
| one row | `group` named by the colour: `blue` |
| its buttons | `less blue`, `more blue` |
| its count | `blue count` |
| the status line | `N of N` (a polite live region) |
| the confirm button | `Add mana` |

The picker's existing dialog name, `Tap <card name> for mana`, is unchanged where buttons are shown.

### 5. `adds_no_mana` also means "adds nothing right now"

`game.ManaAbilityAddsNoMana` also returns true when the ability's output is computed (`ProducedFunc`, `ProducedForPaid` or `DerivedMatch`) and computes to no mana right now, read the way every "what would this make" reader reads it: `manaAbilityProducedLocked` with the largest counter payment. Examples are a power-0 Vivi, a Selvala whose greatest power is 0, and an Exotic Orchard with nothing to copy (CR 106.5, 106.7). A static `Produced: ""` stays out of it, as now ("a broken or empty declaration is not this rule's business").

What follows from that:

- **The view.** The row ships `adds_no_mana: true` and no `color_options`. The field already exists, so this is an additive meaning. No protocol or snapshot version changes, and nothing is persisted.
- **The client.** The predicate (§2) makes the row unusable, so a click never activates it. In the popover it stays shown and greyed, as Command Tower's is today. Activating it is legal (CR 605.1a), but the popover has never offered a disabled row, and offering this one would only spend Vivi's once-per-turn use. The reason becomes generic, "adds no mana right now", because the row's own label already says why ("where X is Vivi Ornitier's power", "any color in your commander's color identity"). That is the reasoning `condition_unmet` already uses.
- **The bots.** The legal enumerator stops offering the move (`legal/abilities.go:1696`). A bot no longer wastes Vivi's activation at power 0, for the reason #844 gave for Command Tower.
- **The auto-tapper** is unchanged. It already does not plan a source with no output (the Vivi file's comment, `effects/vivi_ornitier.go`, and ADR 0011's 2026-09-30 amendment).
- **The engine** still accepts the activation. `docs/protocol.md`'s `adds_no_mana` sentence is widened to say so.

### 6. The tutorial

Step 5 is unchanged: an untapped Forest still adds {G} on a left-click.

Step 7 still anchors to the newest creature with a menu, and its title and completion stay. That creature is usually the mana creature cast in step 6, so it is summoning sick. Under the new rule a left-click on it does nothing, so the hint's "a left click taps it for mana" is no longer true. The hint becomes: "Right-click it, or tap its pip. A left click only acts when an ability is ready to use." The recover check stays as it is: a left-click on an untapped land still taps it for mana, which is still the wrong action for this step, and its body ("A left click taps a permanent for mana") stays true for that case.

A left-click that opens the popover also emits `ability-menu-opened`, so a player who opens the menu that way completes the step too. That is the lesson, reached by another route.

This amends ADR 0076 §2.1's step 7 (the hint only). The anchor and the event are unchanged, so §2.4's anchors contract holds.

### 7. Tests

**Vitest, changed.** The thirty `battlefieldClickIntent` cases move to the new intents. A tapped source, a permanent with no ability and an admin's click on another seat's land become "nothing". Rogue's Passage, a planeswalker, a Squirrel Nest land and an any-player row become the popover. A planeswalker with no catalog abilities becomes "nothing". `manaClick.render.test.ts`'s "keeps click-to-tap" and "untaps a tapped mana source with one click" become "does nothing", and its "Tap (no mana)" case reads the Sandbox section. `tokenGroups.render.test.ts`'s "Use this one" is re-pinned on a token with a usable ability.

**Vitest, new.**

- The predicate: one case per arm in §2, and one test that runs the click rule and the popover's greying over the same fixtures and checks they agree.
- The two bugs: a tapped Vivi at power 3 opens the stepper and sends no `untap`; a power-0 Vivi sends nothing.
- The mixed card: Relic of Sauron opens the popover with the mana row first. With its draw row's own `cant_activate` set and the mana row live, the click goes to the stepper. Tapped, it does nothing.
- The stepper's logic, as pure functions: the Hall test (equal lists; one fixed slot beside a wide one; two different lists where a vector is infeasible); slot assignment that respects every list; the start state (remembered, first common colour, per-slot fallback); + and − enablement; `colors[]` in slot order.
- The stepper rendered: the names in the §4 table, "N of N", Add mana disabled until the total is met, Enter and Escape.
- Right-click on Vivi opens the stepper, not an activation with no colours.
- The Sandbox section: Tap on an untapped vanilla creature, Untap on a tapped one. An uncatalogued planeswalker's manual loyalty rows in the popover, greyed once one was activated this turn.
- `ability-menu-opened` is emitted by the left-click popover path.

**Go (PR 2).** `ManaAbilityAddsNoMana` for a power-0 and a power-2 Vivi, a Selvala at 0, an Exotic Orchard with and without a match, and Command Tower's existing cases. The view ships `adds_no_mana` and no `color_options` for the zero case. `legal.EnumerateFor` offers no activation of a power-0 Vivi and does offer it at power 2. The activation itself is still accepted at power 0.

**Playwright.** No spec drives a card's own click today, but the click's meaning changes, so the nightly E2E runs on PR 3's and PR 4's branches before they merge (`gh workflow run "cmd_and_ctrl E2E" --ref <branch>`), and on `develop` after PR 4. PR 3 also rewrites the stale comment in `s19-triggers.spec.ts`.

---

## Delivery

Each PR goes into `develop`, Sprint S59, Issue #2187.

| PR | What | Needs | Parallel with |
|---|---|---|---|
| 1 | **This ADR,** the S59 section and index row in `docs/sprints.md`, and the AGENTS.md §3 ADR range line. Docs only. | — | — |
| 2 | **Server: `adds_no_mana` covers "adds nothing right now"** (§5). `ManaAbilityAddsNoMana`, its comment and the `ManaAbilityView` field comment; the Go tests in §7; `docs/protocol.md`'s `adds_no_mana` sentence. | 1 | 3 |
| 3 | **Client: the click rule** (§1, §2, §3, §6). The shared predicate and `cant_activate` in `protocol.ts`; `battlefieldClickIntent` and `handleCardClick`; opening the popover from a left-click (a small store keyed by instance ID, in the shape of `manaSourcePicker.ts`, so PlayerPanel and Card open the same popover; a grouped token's "Use this one" opens it on the group's drawn card); the cursor; the Sandbox section; the tapped-Vivi fix; the generic "adds no mana right now" reason; the tutorial hint and event; the vitest in §7; the nightly E2E on the branch. A power-0 Vivi does nothing on click once PR 2 is deployed; before that, the predicate cannot see it. | 1 | 2 |
| 4 | **Client: the stepper** (§4). The pure stepper module and its component inside `ManaSourcePicker`; `MAX_COLOR_OPTIONS` removed; the popover's colour-choice mana rows opening the picker; the vitest in §7; the nightly E2E on the branch. | 3 | — |

After PR 4: run the nightly E2E on `develop`, then close #2187 with evidence (the run, and a cmd-dev check of a tapped Vivi at power 3 splitting U and R).

## Consequences

- A left-click on a permanent does what the card does. A card with nothing to do does nothing, instead of turning sideways.
- Mana with several colour picks is one dialog with a running count, not a list of combinations or N prompts in a row. The 12-answer cap is gone, so five-colour sources (Gwenna, Cascading Cataracts, Selvala) get the dialog too.
- A tapped Vivi no longer untaps on a click, and a power-0 Vivi no longer spends her activation.
- The click, the popover, the picker and the override menu judge a row the same way. The popover gains the Arrest restriction and the planeswalker's once-per-turn check it was missing. The picker greys summoning sickness instead of sending a refused activation.
- Utility lands that also make mana (Rogue's Passage, a manland) and cards like Relic of Sauron open the popover rather than adding mana. That is owner answer 2: one more click on those cards, and no silent choice between their abilities.
- Turning a card sideways by hand takes the popover's Sandbox row, or Alt-click. An admin's plain left-click no longer taps other seats' permanents.
- A player without admin overrides can no longer reach the override menu with a left-click. An uncatalogued planeswalker's manual loyalty rows move to the popover so it keeps them. Everything else in that menu (moving zones, arbitrary counters) was always meant for the admin-overrides setting (ADR 0028 §4).

## Out of scope

- **Strict mana payment and auto-tap by default.** ADR 0118, [#2188](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2188).
- **The auto-tapper's colour choices.** ADR 0011, unchanged.
- **The bots.** They send moves, not clicks. Only §5 reaches them, through the enumerator.
- **Burnt Offering**, and any spell that adds split mana on resolution. Its X picks are `mana_pick` prompts the resolving spell queues, one per slot, and the stepper answers an activation, not a prompt. Answering several prompts in one dialog needs either a multi-slot prompt on the server or the client answering a run of prompts at once. It stays N prompts in the dock. A follow-up issue can take it if it is wanted.
- **Hand and command-zone cards.** Their clicks cast or open the zone browser through `Hand.svelte` and `CommandZone.svelte`. Neither calls `battlefieldClickIntent`, so neither changes. A Spirit Guide's hand mana ability stays on right-click.
- **The override menu's mana rows** still send no colours. They are an admin tool.

## Calls made here

The owner's answers did not settle these. Each is decided above, and each can be overturned in review:

1. **A row that adds nothing stays disabled in the popover** (§5). The brief suggested the popover still allow it, labelled, because activating it is legal. The popover has never offered a disabled row, and the only effect would be to spend Vivi's once-per-turn activation.
2. **Stepper rows follow the server's order** (identity first), not WUBRG (§4), because `docs/protocol.md` and #843 say so.
3. **An uncatalogued planeswalker's manual loyalty rows move into the popover** (§3), so it does not lose them when the left-click stops opening the override menu.
4. **Special actions count as usable rows** (§1), so a face-down creature's click opens the popover to turn it face up (CR 116.2b).
5. **The not-yours click opens the light popover,** not the override menu (§1). Its condition is unchanged.
6. **No tooltip** on a card whose click does nothing (§1): the hover zoom and the ready ring already say enough.
