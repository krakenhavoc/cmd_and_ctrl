# ADR 0227 — Conspiracies in the command zone, and hidden agenda

**Status:** Accepted · 2026-11-24 · S82 — Card coverage: sets outside Standard, newest first. Everything here follows the Comprehensive Rules (CR 315, 702.106, 905.4) on a C2 table; nothing in it needs an owner decision. Draft-time abilities do need one, and it is asked on [#6102](https://github.com/krakenhavoc/cmd_and_ctrl/issues/6102), not answered here ([What this does not decide](#what-this-does-not-decide)).
**Issue:** [#6102](https://github.com/krakenhavoc/cmd_and_ctrl/issues/6102) (Conspiracy Draft cards). This ADR builds the half of it that needs no decision.
**Numbering:** checked on 2026-11-24 after `git fetch --all --prune`, with the loop in [README.md](README.md#picking-a-number) over `docs/decisions/` on all 538 remote branches. `develop` ends at 0225 and `feat/6103-subgames` holds 0226 (subgames and dexterity cards), so this is **0227**.
**Builds on:** [ADR 0064](0064-emblems.md) (the second command-zone container), [ADR 0069](0069-face-down-objects.md) (face-down kinds), [ADR 0172](0172-companion-a-card-outside-the-game.md) (a card that starts the game outside the library, and a special action derived from the card), [ADR 0191](0191-a-sideboard-is-the-outside-the-game-pool.md) (a deck's sideboard rides into the game), ADR 0041 P3 (the closure ratchet) and #522 (snapshots additive within a version).

---

## Context

The engine had no conspiracy card type, nothing that starts the game with a card in the command zone other than a commander, and no hidden agenda. The rules, from the pinned text:

- **CR 315.2 / 905.4.** At the start of a game, before decks are shuffled, each player may put any number of conspiracy cards from their sideboard into the command zone. Those with hidden agenda go in face down.
- **CR 315.3.** They stay in the command zone. They are not permanents, can't be cast and can't be in a deck. If one would leave the command zone, it stays.
- **CR 315.5, 315.5b.** While face up, a conspiracy's statics apply and its triggers trigger. Face down it has no characteristics.
- **CR 315.6, 315.7.** Its owner is the player who put it in and its controller is its owner. You may look at your own face-down conspiracy, and nobody else may.
- **CR 702.106a, 702.106f.** Hidden agenda: as it goes into the command zone, turn it face down and secretly choose a card name. Double agenda chooses two different names.
- **CR 702.106c, 116.2j.** Any time you have priority you may turn your face-down conspiracy face up, revealing the name. This is a special action.
- **CR 702.106d.** "The chosen name" means the name its own hidden agenda chose (CR 607.2d).
- **CR 702.106e.** A player who leaves the game reveals their face-down conspiracies, and every face-down conspiracy is revealed at the end of the game.

None of the conspiracy cards in CNS and CN2 has a draft-time ability: those are on the creatures, spells and land of the sets. So the conspiracy cards can be built on a Commander table exactly as printed.

## Decision

### 1. Where a conspiracy is: the emblem container

A conspiracy goes into `Player.Emblems`, the container ADR 0064 made "the other half of this player's command zone". It is the command zone, every reader of `Command.Cards` means "commander card", no `ZoneRef` resolves to it and the card index does not cover it. So nothing can move, cast, target or sacrifice a conspiracy, which is CR 315.3 enforced by the container rather than by a check. Its statics already join the layer pass (`emblemContinuousEffectsLocked`), its triggers the harvester (`harvestFromEmblemsLocked`) and its cost modifiers the CR 601.2f pass (`appendEmblemCostModifiersLocked`), so a face-up conspiracy works from there with no new walk. CR 800.4a's sweep already takes the container with a leaving player.

`game.IsConspiracy(c)` (the printed type line's first word is `Conspiracy`) tells the two kinds apart. `EmblemsForPlayer` skips conspiracies and `ConspiraciesOf` reads only them.

**Why not a new `Player` container.** It would be a new route from `Game` to cards, a new `census:IntrinsicAbilityCards` line, and that ceiling may only fall without the owner's approval (ADR 0216 needed it for the planar deck). The emblem container already has every property CR 315 asks for.

### 2. Which conspiracies go in

Every conspiracy card in a player's sideboard goes in, as the deal begins and before the libraries are shuffled (`placeConspiraciesLocked`, called at the top of `dealAndBeginLocked`). Listing a conspiracy in the deck's sideboard is the choice to put it in (CR 315.2's "may"), as listing a companion is the choice to reveal it (ADR 0172). Its `IsSideboard` mark is cleared, its owner and controller are the player (CR 315.6), and it leaves the sideboard, so a wish can never fetch it. A conspiracy listed in the main deck is not moved: the deck checker's business, as for any card a format does not allow.

### 3. Face down: a new face-down kind

A hidden agenda conspiracy is face down with `FaceDownConspiracy` (`"conspiracy"`). It is not a CR 708.2 permanent state, so it has no 2/2 body. Its owner is its one viewer (CR 315.7), which `faceDownViewersLocked`'s default answer already gives, since its controller is its owner. `printedTextKeyOf` answers `""` for it, so it has no catalog entry, and with that no static, trigger or cost modifier, until it is turned face up (CR 315.5b). A face-up conspiracy is public: every seat is a knower.

A card declares hidden agenda in its catalog entry, `Spec.HiddenAgenda` (`effects.HiddenAgendaOneName`, or `HiddenAgendaTwoNames` for double agenda), filed as `CardDef.HiddenAgenda`. A conspiracy with no catalog entry goes in face up and does nothing on its own, like any uncatalogued card.

### 4. The name is chosen before any hand is dealt

The name prompt is the ordinary `choose_card_name` prompt, in its keyed form, with the tail `conspiracy/agenda-named` (an on-disk key), which stores the answer on the conspiracy's `ChosenName` (and `ChosenSecondName` for a double agenda; a second answer equal to the first is asked again). The prompt is **private**: every other viewer sees "Secretly choosing a hidden agenda name" with no source and no name suggestions. The answer is not an event, because the event log is public.

CR 905.4 puts conspiracies in before the libraries are shuffled, so the name is chosen without a hand in view. **The deal waits.** When an agenda prompt is open, `dealAndBeginLocked` records the starting seat in `Game.PendingDealSeat` and returns; the last answer deals (`dealIfAgendasChosenLocked`). Keeping or mulliganing is refused meanwhile (`ErrDealPending`). A player who leaves meanwhile has their prompts dropped, and the deal goes on without them. `PendingDealSeat` is plain data, carried by `Clone`, `RestoreFrom` and the snapshot (additive within v9, `pendingDealSeat`, `omitempty`), so a table waiting on a name is a restore point.

### 5. The reveal: a special action

`reveal_agenda` (`SpecialActionRevealAgenda`) is CR 116.2j, one more kind of the one special-action verb, in its tables and nowhere else: its zone is the command zone, its card is looked up among the actor's own conspiracies, its window is "any time you have priority" (the dispatcher's priority check; legal under split second, as turn_face_up is), it is free, and its offer is derived from the card being a face-down conspiracy (`RevealAgendaOffer`), never declared. The performer turns the card face up, makes every seat a knower, gives it a new timestamp and recomputes the layers. The log line is the special action's own label, "Reveal hidden agenda: <name>". The legal-move enumerator offers it like the companion's row.

### 6. Revealing at the end and on leaving (CR 702.106e)

`revealHiddenAgendasLocked` makes every seat a knower of each face-down conspiracy, and logs a reveal naming the chosen name: for every seat as the game ends (beside the foretold-card reveal), and for the leaving player's before CR 800.4a's sweep. The card stays face down: revealing is showing.

### 7. On the wire and in the client

`PlayerView.conspiracies` is the seat's conspiracy cards as `CardView`s, in the order they went in. A face-up one is public. A face-down one is redacted for every viewer but its owner by the ordinary non-knower redaction, so another seat sees a face-down back with no name and no `chosen_name`; once revealed, or revealed at the end, everyone sees the card and its name. The owner's face-down conspiracy carries the reveal row in `special_actions`, stripped for every other viewer.

The client's `ConspiracySlot` shows them beside the commander and companion, as `CompanionSlot` does, with the chosen name as a caption and, on your own face-down one, a Reveal button built from the server's row (`L.revealAgenda`). The group is `L.conspiracies`.

### 8. Cards

A conspiracy card is one file and one `init()` like any other, keyed by oracle ID. Its "the chosen name" abilities read `effects.HasTheChosenName`, `CreaturesYouControlWithTheChosenName`, `SpellsWithTheChosenNameYouCast` and `YouCastASpellWithTheChosenName` (`effects/hidden_agenda.go`), which compare with `game.CardNameMatches` / `game.SpellOrLandHasName` as every chosen name is compared. The first four, all `Full`: Weight Advantage (face up), Immediate Action, Brago's Favor and Iterative Analysis (hidden agenda).

## What this does not decide

**Draft-time abilities.** Many CNS and CN2 creatures, spells and one land have abilities that work during the draft, or read values noted while drafting (CR 905.2): "reveal this card as you draft it and note …", "draft this card face up", "before you shuffle your deck to start the game, … exile a card you drafted that isn't in your deck". C2's tables are Commander and constructed tables with no draft, and ADR 0222's booster draft is a 1v1 limited format that does not draft these sets. Whether that text is inert outside a Conspiracy Draft (as ante text is, #5724), is built into a draft, or something else, is the owner's call, and the options are on #6102. None of those cards is built here.

## Consequences

- Conspiracy cards with hidden agenda or plain face-up text can now ship `Full` with no caveat, one file each, as long as their own abilities are expressible.
- The emblem container now holds two kinds of object. Any new reader of `Player.Emblems` that means "emblem" asks `IsConspiracy` (or `IsEmblem`) first.
- One more face-down kind, one more special action, one more on-disk tail key and one more snapshot field. No closure-census line, no snapshot version bump.
