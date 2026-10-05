# Contract labels

<!-- Generated from client/src/lib/labels.ts by client/src/lib/labels.test.ts. Do not edit by hand:
     change the registry, then run `UPDATE_LABELS_DOC=1 npm test -- labels` in client/. -->

The accessible names the tutorial, the first-use hints and the e2e suite rely on
([ADR 0125](decisions/0125-a-walkthrough-that-keeps-up.md) §2,
[ADR 0076](decisions/0076-tutorial.md) §2.4, [ADR 0111](decisions/0111-action-dock.md) §10).
Each lives once, in `client/src/lib/labels.ts`, and its owner files render it as `L.<key>`
(or `L.<key>(…)`), never as a literal. `labels.test.ts` fails when an owner stops rendering
its entry, when a registered `aria` name is copied as a literal anywhere else under
`client/src`, when a tutorial step or a first-use hint anchors to a name that is not
registered, or when this file is stale.

- **Kind** `aria` is an `aria-label`, or a dialog's or group's name passed through a prop
  that becomes one; only these can be anchors. `text` is a name a button, link or menu item
  gets from its visible text.
- A name with a variable part is matched by its fixed stem when used as an anchor
  (`L.<key>.any`), with the CSS operators `^=`, `$=` or `*=`.
- A trigger's dialog is named by the trigger's reason, which comes from the game, so it has
  no entry here.
- To rename a label, change its value in the registry: tutorial steps and hints follow on
  their own. The older e2e specs select on literals on purpose, so the nightly E2E goes red
  until they are updated; run it on the branch (`gh workflow run "cmd_and_ctrl E2E" --ref <branch>`).

| Key | Name | Role | Kind | Owners | What it is |
|---|---|---|---|---|---|
| `yourBoard` | `your board` | region | aria | `lib/components/board/PlayerPanel.svelte` | the viewer's own board panel; scopes the tutorial's lands and creatures anchors |
| `seatBoard` | `<name> board` (ends with ` board`) | region | aria | `lib/components/board/PlayerPanel.svelte`<br>`lib/components/board/SeatSummary.svelte` | another seat's board panel; the tutorial's attack step points at the bot's while it plays (the stem also ends other names, so anchor by the full name) |
| `yourHand` | `your hand` | generic | aria | `lib/components/board/Hand.svelte` | the viewer's hand; the tutorial's read-hand, play-land and cast-creature steps anchor here |
| `lands` | `lands` | list | aria | `lib/components/board/PlayerPanel.svelte` | a board's land row; every board has one, so anchor it within your board |
| `creatures` | `creatures` | list | aria | `lib/components/board/PlayerPanel.svelte` | a board's creature row; every board has one, so anchor it within your board |
| `commandZone` | `<name> command zone, <N> card(s)` (contains ` command zone, `) | group | aria | `lib/components/board/ExileStrip.svelte`<br>`lib/components/board/CommandStrip.svelte` | a seat's command zone beside its hand (#2349): yours in the castable strip, others' in their command strip |
| `expandBoard` | `Expand <name>'s board, or Expand your own board` (starts with `Expand `) | button | aria | `lib/components/board/Board.svelte` | a board's expand button (ADR 0120); null names the viewer's own |
| `stackPile` | `stack: <N> on the stack` (starts with `stack: `) | region | aria | `lib/components/board/StackLaneHost.svelte`<br>`lib/components/board/StackOverlay.svelte` | the stack pile (ADR 0119), and the floating stack overlay |
| `attention` | `attention` | region | aria | `lib/components/board/Board.svelte` | the attention strip, top left: bot chatter, reveals, the roll call |
| `turnPhase` | `turn and phase indicator` | generic | aria | `lib/components/board/PhaseDisplay.svelte` | the turn number, the active player and the step |
| `actions` | `actions` | region | aria | `lib/components/board/ActionDock.svelte` | the action dock; the tutorial's move-along step and its detours anchor here |
| `priorityControls` | `priority controls` | group | aria | `lib/components/board/ActionDock.svelte` | the dock's toggles: autopass, undo, the ⋯ menu |
| `autopass` | `autopass` | button | aria | `lib/components/board/ActionDock.svelte` | the dock's autopass toggle (state in aria-pressed); the tutorial's let-the-bot-play step anchors here, within actions |
| `next` | `next` | button | text | `lib/components/board/ActionDock.svelte` | the dock's primary: pass priority |
| `passTurn` | `Pass turn` | button | text | `lib/components/board/ActionDock.svelte` | the dock's secondary: skip the rest of the turn |
| `declareAttackers` | `declare attackers` | dialog, group | aria | `lib/combatDock.ts` | the attack request's dialog and its row group |
| `declareBlockers` | `declare blockers` | dialog, group | aria | `lib/combatDock.ts` | the block request's dialog and its row group |
| `mulligan` | `keep or mulligan your hand` | dialog | aria | `routes/Game.svelte` | the opening hand sheet |
| `keepHand` | `Keep hand` | button | text | `routes/Game.svelte` | the opening hand sheet's primary |
| `openingHandDecisions` | `opening hand decisions` | generic | aria | `routes/Game.svelte` | the strip's roll call while hands are kept or mulliganed |
| `discard` | `Discard <N> card(s)` (starts with `Discard `) | dialog | aria | `lib/components/board/DiscardPromptModal.svelte` | the discard request (cleanup or an effect) |
| `selectTarget` | `Select target for <card>` (starts with `Select target for `) | dialog | aria | `lib/targetingDock.ts` | the targeting request |
| `castWithoutPaying` | `Cast <card> without paying its mana cost?` (ends with ` without paying its mana cost?`) | dialog | aria | `lib/targetingDock.ts` | Cast anyway's confirm (ADR 0118 §2), with its Cast and Cancel |
| `cast` | `Cast` | button | text | `lib/targetingDock.ts` | Cast anyway's confirm: cast it unpaid |
| `cancel` | `Cancel` | button | text | `lib/targetingDock.ts`<br>`lib/dock.ts` | Cast anyway's confirm, and every flow's Cancel (the dock's cancelAction), the first-turn confirm's included |
| `castAnyway` | `Cast anyway (don't pay)` | menuitem | text | `lib/castAnyway.ts` | the card menu's row that casts without paying (ADR 0118 §2) |
| `rollForFirstTurn` | `roll for the first turn` | dialog | aria | `lib/components/board/OpeningRollDock.svelte` | the opening roll's request |
| `roll` | `Roll` | button | text | `lib/components/board/OpeningRollDock.svelte` | the opening roll's primary |
| `rollForEveryone` | `Roll for everyone left` | button | text | `lib/components/board/OpeningRollDock.svelte` | the host's secondary on the opening roll |
| `chooseFirstTurn` | `choose who takes the first turn` | dialog | aria | `lib/components/board/OpeningRollDock.svelte` | the roll winner's sheet |
| `goesFirst` | `<name> goes first` (ends with ` goes first`) | button | text | `lib/openingRoll.ts` | the chooser's button for another seat |
| `iGoFirst` | `I go first` | button | text | `lib/openingRoll.ts` | the chooser's button for their own seat |
| `giveFirstTurn` | `Let <name> take the first turn?` (ends with ` take the first turn?`) | dialog | aria | `lib/openingRoll.ts` | the confirm when the chooser gives the first turn away, with Confirm and Cancel |
| `confirm` | `Confirm` | button | text | `lib/components/board/OpeningRollDock.svelte` | the first-turn confirm's primary |
| `openingRoll` | `opening roll` | group | aria | `lib/components/board/OpeningRollBanner.svelte` | the strip's banner during the roll; stands in for opening hand decisions until the winner chooses |
| `moreActions` | `more actions` | button | aria | `lib/components/board/GameMenu.svelte` | the ⋯ button that opens the game menu |
| `gameActions` | `game actions` | menu | aria | `lib/components/board/GameMenu.svelte` | the ⋯ menu itself |
| `dice` | `Dice` | none | text | `lib/components/board/GameMenu.svelte` | the ⋯ menu's heading over the table rolls; plain text, not a group, with no name of its own |
| `rollD6` | `Roll a d6` | menuitem | text | `lib/gameMenu.ts` | a table roll (ADR 0121 §5); disabled for 2 s after the viewer's own roll |
| `rollD20` | `Roll a d20` | menuitem | text | `lib/gameMenu.ts` | a table roll (ADR 0121 §5) |
| `flipCoin` | `Flip a coin` | menuitem | text | `lib/gameMenu.ts` | a table roll (ADR 0121 §5) |
| `adminViews` | `admin views` | navigation | aria | `routes/Admin.svelte` | the admin views' tabs: Live now, Games, Accounts |
| `adminLiveNowTab` | `Live now` | link | text | `routes/Admin.svelte` | the admin views' first tab |
| `adminGamesTab` | `Games` | link | text | `routes/Admin.svelte` | the admin views' tab for every table |
| `adminAccountsTab` | `Accounts` | link | text | `routes/Admin.svelte` | the admin views' tab for every account |
| `liveNow` | `live now` | region | aria | `lib/components/admin/LiveNow.svelte` | who is on now and the tables in play |
| `gamesTable` | `games` | table | aria | `lib/components/admin/GamesList.svelte` | the admin view's list of tables |
| `accountsTable` | `accounts` | table | aria | `lib/components/admin/AccountsList.svelte` | the admin view's list of accounts |
| `createGame` | `create game` | form | aria | `routes/Lobby.svelte` | the Lobby's create-a-table form; the lobby.create hint points at it |
| `deckLink` | `deck link` | textbox | aria | `routes/Decks.svelte` | the Decks page's deck-link field; the decks.check hint points at it |
| `yourDecks` | `your decks` | list | aria | `routes/Decks.svelte` | a signed-in person's saved decks; the decks.library hint points at it |
| `settingsSections` | `settings sections` | navigation | aria | `lib/components/Settings.svelte` | the Settings dialog's section tabs; the settings.display hint points at it |
| `searchCatalogue` | `search the catalogue` | textbox | aria | `routes/Catalog.svelte` | the Catalogue's search field; the catalog.search hint points at it |
| `searchRoadmap` | `search the roadmap` | textbox | aria | `routes/Roadmap.svelte` | the Roadmap's search field; the roadmap.search hint points at it |
| `tip` | `tip` | complementary | aria | `lib/components/hints/HintCard.svelte` | a first-use hint's card (ADR 0125 §3.6); one at a time, and Go to the tip (i) moves focus to it |
| `keyboardShortcuts` | `keyboard shortcuts` | dialog | text | `lib/components/ShortcutsOverlay.svelte` | the keymap overlay (?), named by its heading |
