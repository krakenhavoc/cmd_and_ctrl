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
| `tombstoneRip` | `R. I. P.` | generic | text | `lib/components/board/Tombstone.svelte` | #4449: the epitaph on the tombstone that stays on an eliminated player's panel |
| `tombstoneOf` | `<name> is out of the game` (ends with ` is out of the game`) | group | aria | `lib/components/board/Tombstone.svelte` | #4449: the tombstone on an eliminated player's panel; its text is the player's name, their commander and the epitaph |
| `reflowNextUp` | `Next up` | generic | text | `lib/expansion.ts` | ADR 0077 (amended 2026-10-17, #4102): the chip on a reflow avatar card whose seat plays straight after the active player (drawn upper case) |
| `reflowAfterYou` | `After you` | generic | text | `lib/expansion.ts` | ADR 0077 (amended 2026-10-17, #4102): the chip on a reflow avatar card whose seat plays straight after the viewer (drawn upper case) |
| `reflowAfter` | `After <name>` (starts with `After `) | generic | text | `lib/expansion.ts` | ADR 0077 (amended 2026-10-17, #4102): the chip on a reflow avatar card naming the opponent who plays right before it (drawn upper case) |
| `yourHand` | `your hand` | generic | aria | `lib/components/board/Hand.svelte` | the viewer's hand; the tutorial's read-hand, play-land and cast-creature steps anchor here |
| `lands` | `lands` | list | aria | `lib/components/board/PlayerPanel.svelte` | a board's land row; every board has one, so anchor it within your board |
| `creatures` | `creatures` | list | aria | `lib/components/board/PlayerPanel.svelte` | a board's creature row; every board has one, so anchor it within your board |
| `attractionDeck` | `attractions` | button | text | `lib/components/board/PileBar.svelte` | a seat's face-down Attraction deck in the rail (ADR 0207, CR 717.2), drawn only while it holds a card; a count, not openable |
| `junkyard` | `junkyard` | button | text | `lib/components/board/PileBar.svelte` | a seat's face-up junkyard in the rail (ADR 0207, CR 717.6a), drawn only while it holds a card; opens the zone browser's junkyard tab |
| `tickets` | `tickets` | status | aria | `lib/components/board/PlayerIdentity.svelte` | a seat's ticket counters ({TK}, ADR 0208, CR 107.17) in its marker column, drawn only while it has one; a count, no stepper |
| `enduringStory` | `an enduring story` | status | aria | `lib/components/board/PlayerIdentity.svelte` | a seat's enduring story (CR 702.195b, #4475) in its marker column, drawn only once storied has given it; read-only, there is no toggle |
| `stickerSheets` | `sticker sheets` | status | aria | `lib/components/board/PlayerIdentity.svelte` | a seat's three sticker sheets (ADR 0208, CR 123.2c): how many stickers are still free, the sheets and their stickers on hover |
| `stickersOnCard` | `stickers` | status | aria | `lib/components/board/StickerBadge.svelte` | the stickers on a card (ADR 0208, CR 123): a count on the card's left edge, the stickers on hover |
| `commandZone` | `<name> command zone, <N> card(s)` (contains ` command zone, `) | group | aria | `lib/components/board/ExileStrip.svelte`<br>`lib/components/board/CommandStrip.svelte` | a seat's command zone beside its hand (#2349): yours in the castable strip, others' in their command strip; the tutorial's commander step and the table.commander hint anchor to yours, within your board |
| `companion` | `<seat> companion, <card name>` (contains ` companion, `) | group | aria | `lib/components/board/CompanionSlot.svelte` | a seat's revealed companion while it waits outside the game (ADR 0172), beside its commanders |
| `takeCompanion` | `Pay <cost>: put <card name> into your hand (companion)` (ends with ` into your hand (companion)`) | button | aria | `lib/components/board/CompanionSlot.svelte` | the companion special action (CR 116.2g, ADR 0172): pay {3} to put your companion into your hand; greyed outside your main phase |
| `ignoreEffect` | `Ignore <card name> (sacrifice a permanent)` (ends with ` (sacrifice a permanent)`) | button | aria | `lib/components/board/IgnoreEffects.svelte` | the CR 116.2d special action (#1986): sacrifice a permanent of your choice to ignore a static ability's effect until end of turn (Damping Engine); on your own panel while the effect bears on you, and it opens the sacrifice picker |
| `expandBoard` | `Expand <name>'s board, or Expand your own board` (starts with `Expand `) | button | aria | `lib/components/board/Board.svelte` | a board's expand button (ADR 0120); null names the viewer's own |
| `stackPile` | `stack: <N> on the stack` (starts with `stack: `) | region | aria | `lib/components/board/StackLaneHost.svelte`<br>`lib/components/board/StackOverlay.svelte` | the stack pile (ADR 0119), and the floating stack overlay; the tutorial's stack step anchors here |
| `stackAbility` | `Activate <card name> on the stack: <cost>` (contains ` on the stack: `) | button | aria | `lib/components/board/StackAbilityButtons.svelte` | #1865: a spell's ability that works while it is on the stack (Lightning Storm's discard a land card), on every stack surface beside Counter; offered to each seat that may activate it |
| `stackAbilityGone` | `<card name> left the stack` (ends with ` left the stack`) | status | text | `lib/components/board/Board.svelte` | #4043: the notice that closes the discard picker when the spell whose stack ability was being announced resolves or is countered first |
| `attention` | `attention` | region | aria | `lib/components/board/Board.svelte` | the attention strip, top left: bot chatter, reveals, the roll call |
| `turnPhase` | `turn and phase indicator` | generic | aria | `lib/components/board/PhaseDisplay.svelte` | the turn number, the active player and the step |
| `abilityChip` | `<N> <kind> ability, or <N> <kind> abilities (kind: triggered, static or activated)` (contains ` abilit`) | button | aria | `lib/components/board/AbilityKindChips.svelte` | an art tile's ability chip (#2219): one per kind the card has, its hover or focus lists the server's labels; the e2e suite reads Mulldrifter's |
| `crewedBadge` | `crewed by <names>, or crewed this turn` (starts with `crewed `) | generic | aria | `lib/components/board/Card.svelte` | a Vehicle's crewed badge (ADR 0186 §6): the creatures tapped this turn to crew it; the names come from the board, and an ID with no card there is left out |
| `saddledBadge` | `saddled by <names> until end of turn, or saddled until end of turn` (starts with `saddled `) | generic | aria | `lib/components/board/Card.svelte` | a Mount's saddled badge (ADR 0186 §6): the creatures tapped this turn to saddle it; the names come from the board, and an ID with no card there is left out |
| `actions` | `actions` | region | aria | `lib/components/board/ActionDock.svelte` | the action dock; the tutorial's move-along step and its detours anchor here |
| `waitToRespond` | `wait, let me respond` | button | aria | `lib/components/board/ActionDock.svelte` | the stack hold countdown's one-click hold (#2853): keeps priority on the stack auto-pass was about to pass |
| `jumpChip` | `Your decision on <Bob's | your own> board` (starts with `Your decision on `) | status | aria | `lib/components/board/HandoffChip.svelte` | the small chip over the dock while the viewer plays another seat's turn and the next decision is owed on the board not shown (#4434); the board never switches by itself |
| `jumpToBoard` | `Jump` | button | text | `lib/components/board/HandoffChip.svelte` | the jump chip's button: switches to the board that owes the decision (#4434) |
| `priorityControls` | `priority controls` | group | aria | `lib/components/board/ActionDock.svelte` | the dock's toggles: Skip to my turn, undo, the ⋯ menu |
| `skipToMyTurn` | `Skip to my turn` | button | aria | `lib/components/board/ActionDock.svelte` | the dock's Skip to my turn toggle (ADR 0143 §4.2, formerly autopass; state in aria-pressed): passes for you until your next main phase; the tutorial's let-the-bot-play step anchors here, within actions |
| `recordLoop` | `Record loop` | menuitem | text | `lib/components/board/GameMenu.svelte` | the dock's ⋯ menu entry while you hold priority (ADR 0193 first amendment, CR 732.2a): every action from now on, any seat's, is one iteration of a loop; once recording, Stop and Cancel are the dock's status line |
| `stopLoop` | `Stop recording the loop` | button | aria | `lib/components/board/LoopControl.svelte` | ends the demonstration (ADR 0193): the server checks the board came back, keeps the loop to propose, or discards it and says why in the log |
| `cancelLoop` | `Cancel the loop` | button | aria | `lib/components/board/LoopControl.svelte`<br>`lib/components/board/LoopProposeSheet.svelte` | discards the recording and any proposal still being answered (ADR 0193) |
| `proposeLoop` | `Propose the loop` | dialog | aria | `lib/components/board/LoopProposeSheet.svelte` | the recorder's sheet once the loop closed (ADR 0193 §9): the iteration, its net change and how many more times to repeat it |
| `loopProposal` | `Loop shortcut proposal` | dialog | aria | `lib/components/board/LoopProposalSheet.svelte` | another player's proposal to repeat a loop (CR 732.2b): accept it, or shorten it to one of your own steps |
| `pass` | `Pass` | button | aria | `lib/components/board/ActionDock.svelte` | the dock's primary while the stack holds something (#3267): pass priority so the top item can resolve (its visible text and its aria-label); the tutorial's stack step anchors here, within actions |
| `next` | `Next` | button | aria | `lib/components/board/ActionDock.svelte` | the dock's primary on an empty stack (#3267): pass priority so the game moves to the next step (its visible text and its aria-label); the tutorial's stack step anchors here too, within actions |
| `nextEndsTurn` | `Pass · turn ends` | button | aria | `lib/components/board/ActionDock.svelte` | the dock's primary at your last priority of your own turn (#3826, turn.pass_hint ends_turn): a pass hands the turn over; ringed, and named so End turn stays the one button of that name (its visible text and its aria-label) |
| `nextYourTurn` | `Pass · your turn next` | button | aria | `lib/components/board/ActionDock.svelte` | the dock's primary at your last priority before your turn begins (#3826, turn.pass_hint your_turn_next): your next decision is in your own turn; ringed (its visible text and its aria-label) |
| `attackWith` | `Attack with <N>` (starts with `Attack with `) | button | aria | `lib/components/board/ActionDock.svelte` | the dock's primary in declare attackers while creatures are staged (#3266): sends them as one declaration, then passes (its visible text and its aria-label) |
| `noAttacks` | `No attacks` | button | aria | `lib/components/board/ActionDock.svelte` | the dock's primary in declare attackers while creatures could attack and none has (#3267): pass priority with no attackers, which skips the rest of combat (CR 508.8) |
| `attackAll` | `Attack <name> with all <N> creature(s)` (contains ` with all `) | button | aria | `lib/attackAll.ts` | the dock's attack-with-all button in a duel (its aria-label equals its visible text); the tutorial's attack step anchors here, within actions |
| `endTurn` | `End turn` | button | text | `lib/components/board/ActionDock.svelte` | the dock's secondary, shown to the active player only (ADR 0143 §4.2, formerly Pass turn): passes for you through every step of the rest of your turn |
| `sandboxPassTurn` | `Skip to next turn (sandbox)` | menuitem | text | `lib/components/board/GameMenu.svelte` | the ⋯ menu's sandbox jump to the next turn (pass_turn): skips the end step, its triggers and the cleanup discard, and the log says so |
| `declareAttackers` | `declare attackers` | dialog, group | aria | `lib/combatDock.ts` | the attack request's dialog and its row group |
| `attackPlain` | `Attack` | button | text | `lib/combatDock.ts` | ADR 0130 §7: the dock's choice for a creature that may be exerted, attacking without exerting it |
| `attackAndExert` | `Attack and exert` | button | text | `lib/combatDock.ts` | ADR 0130 §7: the dock's choice for a creature that may be exerted, attacking and exerting it |
| `declareAttackerAndExert` | `Declare attacker and exert` | menuitem | text | `lib/contextMenu.logic.ts` | ADR 0130 §7: the card menu's row beside Declare attacker, for a creature that may be exerted |
| `exertAttacker` | `Exert <name>` (starts with `Exert `) | switch | aria | `lib/components/board/AttackDeclarationModal.svelte` | ADR 0130 §7: the Choose attackers picker's per-row Exert toggle |
| `attackPayWithLife` | `pay the attack tax with life` | group | aria | `lib/components/board/AttackLifeStepper.svelte` | #4198: the stepper that pays a Phyrexian attack tax (Norn's Annex's {W/P}) with 2 life per symbol; drawn in the Choose attackers picker and in the Pay the attack tax sheet |
| `declareBlockers` | `declare blockers` | dialog, group | aria | `lib/combatDock.ts` | the block request's dialog and its row group |
| `removeBlock` | `Remove block` | button | text | `lib/combatDock.ts` | #3898: the dock's button that takes the selected blocker's block back while blockers are declared on the board |
| `chooseBlockers` | `choose blockers` | dialog | aria | `lib/components/board/BlockListSheet.svelte` | #3898: the block declaration as a list, opened from the block request's Show as list; the keyboard path |
| `mulligan` | `keep or mulligan your hand` | dialog | aria | `routes/Game.svelte` | the opening hand sheet; the tutorial's play-land and cast-creature steps detour here while it waits |
| `keepHand` | `Keep hand` | button | text | `routes/Game.svelte` | the opening hand sheet's primary |
| `openingHandDecisions` | `opening hand decisions` | generic | aria | `routes/Game.svelte` | the strip's roll call while hands are kept or mulliganed |
| `discard` | `Discard <N> card(s)` (starts with `Discard `) | dialog | aria | `lib/components/board/DiscardPromptModal.svelte` | the discard request (cleanup or an effect) |
| `crewCountsAs` | `counts as <N>` (starts with `counts as `) | generic | text | `lib/components/board/CrewCostModal.svelte` | ADR 0186 Decision 6 (#3550): the crew or saddle picker's note on a creature that counts for something other than its power on this cost |
| `selectTarget` | `Select target for <card>` (starts with `Select target for `) | dialog | aria | `lib/targetingDock.ts` | the targeting request |
| `targetPlayersLead` | `Players →` | generic | text | `lib/targetingDock.ts` | #4435: the lead of the targeting request's row of legal players, so a player a dialog or a hand covers on the board is still one press away |
| `targetPlayer` | `Target <player>` (starts with `Target `) | button | aria | `lib/targetingDock.ts` | #4435: the targeting request's button for one legal player (a seat-colour dot and the name); it picks that player exactly as a click on their avatar does |
| `chooseOnBoard` | `Choose on the board for <card>` (starts with `Choose on the board for `) | dialog | aria | `lib/targetingDock.ts` | a card-set pick answered by clicking permanents on the board (#2394) |
| `showAsList` | `Show as list` | button | text | `lib/targetingDock.ts`<br>`lib/combatDock.ts` | a board-answered card-set pick's fallback: open the same choice as the list; #3898: the block request's way to the blockers list |
| `castPayingOnPaper` | `Cast <card> and pay the rest on paper?` (ends with ` and pay the rest on paper?`) | dialog | aria | `lib/targetingDock.ts` | Cast anyway's confirm (ADR 0118 §2, renamed by #3855), with its Cast and Cancel |
| `paperPayment` | `What did you pay for <card>?` (starts with `What did you pay for `) | dialog | aria | `lib/components/board/PaperPaymentModal.svelte` | the paper payment picker (#3855): what a waived payment something reads paid outside the pool |
| `paperSlotOnX` | `on X` | generic | text | `lib/components/board/PaperPaymentModal.svelte` | the paper payment picker's tag on a slot of the cost's {X} part, the mana a spend-only-on-X clause restricts (#3941) |
| `castFree` | `Cast without paying its mana cost` | button | text | `lib/components/board/PaperPaymentModal.svelte` | the paper payment picker's other answer: an effect on paper let you cast it free (#3855) |
| `cast` | `Cast` | button | text | `lib/targetingDock.ts` | Cast anyway's confirm: cast it, the pool paying what it can and the rest paid on paper |
| `cancel` | `Cancel` | button | text | `lib/targetingDock.ts`<br>`lib/dock.ts` | Cast anyway's confirm, and every flow's Cancel (the dock's cancelAction), the first-turn confirm's included |
| `castAnyway` | `Cast anyway (pay on paper)` | menuitem | text | `lib/castAnyway.ts` | the card menu's row that casts what the pool can't pay, the rest paid on paper (ADR 0118 §2, renamed by #3855) |
| `energyToPay` | `energy to pay` | group | aria | `lib/components/board/ChoicePromptModal.svelte` | the pay_amount prompt's stepper (ADR 0129 §3): how much energy to pay, from its floor to the seat's energy |
| `payEnergy` | `Pay <N> {E}` (ends with ` {E}`) | button | text | `lib/choiceDock.ts` | the pay_amount prompt's primary: pays the stepper's amount |
| `payNothing` | `Pay nothing` | button | text | `lib/choiceDock.ts` | the pay_amount prompt's decline when the card says any amount; it reads Don't pay when the card says one or more |
| `lifeToPay` | `life to pay` | group | aria | `lib/components/board/ChoicePromptModal.svelte` | the pay_amount prompt's stepper when it asks for life (#1941): how much life to pay, from nothing to the seat's life total |
| `payLife` | `Pay <N> life` (ends with ` life`) | button | text | `lib/choiceDock.ts` | the pay_amount prompt's primary when it asks for life (#1941): pays the stepper's amount |
| `sacrifice` | `Sacrifice` | button | text | `lib/components/board/ChoicePromptModal.svelte` | a sacrifice choice's confirm in the dock (sacrifice_choice, entry_sacrifice, an edict's choose_cards): enabled once the permanents are picked, in the sheet or on the board (#2880) |
| `numberToChoose` | `number to choose` | group | aria | `lib/components/board/ChoicePromptModal.svelte` | the pay_amount prompt's field when it asks for a number that is not paid (#1941): Volcano Hellion's amount of damage |
| `chooseNumber` | `Choose <N>` (starts with `Choose `) | button | text | `lib/choiceDock.ts` | the pay_amount prompt's primary when it asks for a number that is not paid (#1941): answers the field's number |
| `payLifeForMana` | `Pay life for {B}…` | menuitem | text | `lib/payLifeForMana.ts` | the card menu's row that opens the life stepper for the {B} in a cost, under K'rrik (ADR 0131 §4) |
| `rollForFirstTurn` | `roll for the first turn` | dialog | aria | `lib/components/board/OpeningRollDock.svelte` | the opening roll's request; the tutorial's roll step anchors here while the viewer owes a die |
| `roll` | `Roll` | button | text | `lib/components/board/OpeningRollDock.svelte` | the opening roll's primary |
| `rollForEveryone` | `Roll for everyone left` | button | text | `lib/components/board/OpeningRollDock.svelte` | the host's secondary on the opening roll |
| `chooseFirstTurn` | `choose who takes the first turn` | dialog | aria | `lib/components/board/OpeningRollDock.svelte` | the roll winner's sheet; the tutorial's roll step detours here |
| `goesFirst` | `<name> goes first` (ends with ` goes first`) | button | text | `lib/openingRoll.ts` | the chooser's button for another seat |
| `iGoFirst` | `I go first` | button | text | `lib/openingRoll.ts` | the chooser's button for their own seat |
| `giveFirstTurn` | `Let <name> take the first turn?` (ends with ` take the first turn?`) | dialog | aria | `lib/openingRoll.ts` | the confirm when the chooser gives the first turn away, with Confirm and Cancel; in the tutorial's roll detour |
| `confirm` | `Confirm` | button | text | `lib/components/board/OpeningRollDock.svelte` | the first-turn confirm's primary |
| `openingRoll` | `opening roll` | group | aria | `lib/components/board/OpeningRollBanner.svelte` | the strip's banner during the roll; stands in for opening hand decisions until the winner chooses; the tutorial's roll step anchors here while the table waits on the bot |
| `moreActions` | `more actions` | button | aria | `lib/components/board/GameMenu.svelte` | the ⋯ button that opens the game menu |
| `gameActions` | `game actions` | menu | aria | `lib/components/board/GameMenu.svelte` | the ⋯ menu itself |
| `dice` | `Dice` | none | text | `lib/components/board/GameMenu.svelte` | the ⋯ menu's heading over the table rolls; plain text, not a group, with no name of its own |
| `sortHand` | `Sort hand` | none | text | `lib/components/board/GameMenu.svelte` | the ⋯ menu's heading over the one-time hand sorts (#3028), and the sortHand shortcut's row in the keymap |
| `rollD6` | `Roll a d6` | menuitem | text | `lib/gameMenu.ts` | a table roll (ADR 0121 §5); disabled for 2 s after the viewer's own roll |
| `rollD20` | `Roll a d20` | menuitem | text | `lib/gameMenu.ts` | a table roll (ADR 0121 §5) |
| `flipCoin` | `Flip a coin` | menuitem | text | `lib/gameMenu.ts` | a table roll (ADR 0121 §5) |
| `powerUpUsed` | `already activated (power-up)` | generic | text | `lib/contextMenu.logic.ts` | the reason on a power-up ability row the server marked power_up_used (#4527, CR 702.193a: activate each power-up ability only once), in the card menu and the mana-ability popover |
| `rulings` | `Rulings` | menuitem | text | `lib/components/board/CardContextMenu.svelte`<br>`lib/components/board/ManaAbilityMenu.svelte` | a card menu's link to the card's official rulings in a new tab (#3831); not offered for a hidden or face-down card or a token |
| `adminViews` | `admin views` | navigation | aria | `routes/Admin.svelte` | the admin views' tabs: Live now, Games, Accounts |
| `adminLiveNowTab` | `Live now` | link | text | `routes/Admin.svelte` | the admin views' first tab |
| `adminGamesTab` | `Games` | link | text | `routes/Admin.svelte` | the admin views' tab for every table |
| `adminAccountsTab` | `Accounts` | link | text | `routes/Admin.svelte` | the admin views' tab for every account |
| `adminReportsTab` | `Reports` | link | text | `routes/Admin.svelte` | the admin views' tab for what players have reported at their tables (#2984) |
| `adminStatsTab` | `Stats` | link | text | `routes/Admin.svelte` | the admin views' tab for game statistics (ADR 0165 §8, #3425) |
| `liveNow` | `live now` | region | aria | `lib/components/admin/LiveNow.svelte` | who is on now and the tables in play |
| `gamesTable` | `games` | table | aria | `lib/components/admin/GamesList.svelte` | the admin view's list of tables |
| `accountsTable` | `accounts` | table | aria | `lib/components/admin/AccountsList.svelte` | the admin view's list of accounts |
| `createGame` | `create game` | form | aria | `routes/Lobby.svelte` | the Tables page's new-table form, opened by + New table (ADR 0158 §1 item 1) |
| `newTable` | `new table` | button | aria | `routes/Lobby.svelte` | the Tables page head's + New table, which opens the new-table form; the lobby.create hint points at it |
| `suggestTableName` | `suggest a name` | button | aria | `lib/components/NewTableRow.svelte` | the dice next to the Lobby's game-name box: fills it with a generated table name (#2630) |
| `deckLink` | `deck link` | textbox | aria | `routes/Decks.svelte` | the Decks page's deck-link field; the decks.check hint points at it |
| `choosePrintings` | `Choose printings` | button | text | `lib/components/decks/DeckCard.svelte` | saved-deck printing picker |
| `printingPicker` | `card printings` | region | aria | `lib/components/decks/PrintingPicker.svelte` | saved-deck printing picker |
| `printingCard` | `Card to customize` | combobox | aria | `lib/components/decks/PrintingPicker.svelte` | saved-deck printing picker |
| `filterPrintings` | `Filter printings` | searchbox | aria | `lib/components/decks/PrintingPicker.svelte` | saved-deck printing picker |
| `savePrintings` | `Save printings` | button | text | `lib/components/decks/PrintingPicker.svelte` | saved-deck printing picker |
| `cancelPrintings` | `Cancel printing changes` | button | text | `lib/components/decks/PrintingPicker.svelte` | saved-deck printing picker |
| `updateDeck` | `Update deck` | button | text | `lib/components/decks/DeckCard.svelte` | a saved deck's action that opens the update sheet (#2036) |
| `deckUpdateSheet` | `update deck` | region | aria | `lib/components/decks/UpdateDeckSheet.svelte` | the sheet that updates a saved deck from its link, a new link or a pasted list (#2036) |
| `deckUpdateChoices` | `how to update` | group | aria | `lib/components/decks/UpdateDeckSheet.svelte` | the update sheet's three ways in: re-pull, new link, paste list (#2036) |
| `previewUpdate` | `Preview changes` | button | text | `lib/components/decks/UpdateDeckSheet.svelte` | the update sheet's first step: read the new list and show the diff, writing nothing (#2036) |
| `saveUpdate` | `Save update` | button | text | `lib/components/decks/UpdateDeckSheet.svelte` | the update sheet's confirm: commits the previewed list (#2036) |
| `cancelUpdate` | `Cancel update` | button | text | `lib/components/decks/UpdateDeckSheet.svelte` | closes the update sheet without saving (#2036) |
| `yourDecks` | `your decks` | list | aria | `routes/Decks.svelte` | a signed-in person's saved decks; the decks.library hint points at it |
| `settingsSections` | `settings sections` | navigation | aria | `lib/components/Settings.svelte` | the Settings dialog's section tabs; the settings.display hint points at it |
| `searchCatalogue` | `search the catalogue` | textbox | aria | `routes/Catalog.svelte` | the Catalogue's search field; the catalog.search hint points at it |
| `searchRoadmap` | `search the roadmap` | textbox | aria | `routes/Roadmap.svelte` | the Roadmap's search field; the roadmap.search hint points at it |
| `lobbyTitle` | `cmd_and_ctrl · lobby` | heading | aria | `routes/Lobby.svelte` | the Lobby's title (its visible text is Tables); the lobby.practice hint anchors here |
| `copyInvite` | `Copy invite` | button | text | `routes/Lobby.svelte` | the wide part of a lobby table's invite split button: copies the player invite link; the lobby.create hint names it |
| `newInviteLink` | `new invite link` | button | aria | `routes/Lobby.svelte` | the ↻ part of the invite split button: mints a new player invite after a confirm, then copies it |
| `copySpectatorLink` | `Copy spectator link` | button | text | `routes/Lobby.svelte` | the wide part of a table's spectator split button: copies the read-only link |
| `newSpectatorLink` | `new spectator link` | button | aria | `routes/Lobby.svelte` | the ↻ part of the spectator split button: mints a new spectator link after a confirm, then copies it |
| `tableSettingsLine` | `table settings: <summary>` (starts with `table settings: `) | button | aria | `routes/Lobby.svelte` | a table card's rules line, e.g. 30 life · no undos: opens the table settings panel (read-only for a non-host) |
| `tableMenu` | `actions for <table name>` (starts with `actions for `) | menu | aria | `routes/Lobby.svelte` | a table card's right-click menu (also the Menu key, Shift+F10 or a long press) |
| `reportABug` | `Report a bug or idea` | button | aria | `routes/Game.svelte`<br>`lib/components/SiteHeader.svelte` | the bug button: the table's command bar, and the site header on every other page for someone signed in with Discord (#3492); the Guide and About name it |
| `navTables` | `Tables` | link | text | `lib/components/SiteHeader.svelte` | the site header's link to #/lobby, first in the signed-in nav (it read Lobby before ADR 0156) |
| `siteMenu` | `Site menu` | button | aria | `lib/components/SiteHeader.svelte` | the site header's menu button below 600px, which opens the nav links (a disclosure: aria-expanded, Escape closes it); hidden at tablet width and up (#3877) |
| `signInWithDiscord` | `Sign in with Discord` | link | text | `routes/FrontPage.svelte` | the front page's Discord sign-in, shown when the server has Discord sign-in configured |
| `joinDiscordServer` | `Join our Discord server` | link | text | `routes/FrontPage.svelte` | the front page's link to the C2 Discord invite; hidden when GET /community names none |
| `tryPracticeGame` | `Try a practice game` | link | text | `routes/FrontPage.svelte`<br>`routes/RequestAccess.svelte` | the front page's and the gate page's way in without an account: #/practice, a guest practice game when signed out |
| `findAPod` | `Find a pod` | link | text | `routes/Lobby.svelte` | the Lobby's empty state: the C2 Discord invite, in the No pod yet? card; hidden when GET /community names none |
| `help` | `help` | button | aria | `lib/components/HelpMenu.svelte` | the site header's Help button; the menu it opens is named by it (aria-labelledby), and the site.help hint anchors here |
| `tipsForThisPage` | `Tips for this page` | menuitem | text | `lib/components/HelpMenu.svelte` | Help: shows this page's tips again, one after another, even with tips off |
| `tipsForTheTable` | `Tips for the table` | menuitem | text | `lib/components/board/GameMenu.svelte` | the ⋯ menu's Help group: shows the table's tips again at the next quiet moments |
| `showAllTipsAgain` | `Show all tips again` | menuitem | text | `lib/components/HelpMenu.svelte`<br>`lib/components/Settings.svelte` | Help, and Settings → Advanced: forgets every dismissed tip and turns tips back on |
| `practiceGame` | `Practice game` | menuitem | text | `lib/components/HelpMenu.svelte`<br>`routes/Lobby.svelte` | Help, and the Lobby's empty-state card: opens a practice table, a guest one when signed out (ADR 0156 §2) |
| `gettingStarted` | `Getting started` | menuitem | text | `lib/components/HelpMenu.svelte` | Help: opens the getting-started guide (#/guide) |
| `keyboardShortcutsItem` | `Keyboard shortcuts` | menuitem | text | `lib/components/HelpMenu.svelte`<br>`lib/components/board/GameMenu.svelte` | Help, and the ⋯ menu's Help group: opens the keymap overlay |
| `replayTutorial` | `Replay the tutorial` | menuitem | text | `lib/components/board/GameMenu.svelte`<br>`lib/components/Settings.svelte` | Settings → Advanced, and the ⋯ menu on the practice table only: opens a fresh practice table |
| `legalNotice` | `Legal notice` | menuitem | text | `lib/components/board/GameMenu.svelte` | the ⋯ menu's Help group: opens the fan-content notice, which the site footer carries everywhere else (#2982) |
| `showTips` | `Show tips` | checkbox | text | `lib/components/Settings.svelte` | Settings → Advanced: on unless Hide tips was pressed on a tip |
| `stackArrivalToggle` | `Stack: trace each spell or ability in from its source` | checkbox | text | `lib/components/Settings.svelte` | Settings → Animations: the arrival trace, animations.stackArrival (ADR 0119 amendment 2026-10-16 §D) |
| `opponentCardFacing` | `Opponents' cards` | combobox | text | `lib/components/Settings.svelte` | Settings → Display: whether a top-row opponent's cards face them (turned 180°, the default) or the viewer, display.opponentCardFacing (#4555) |
| `effectsSetting` | `Table effects` | combobox | text | `lib/components/Settings.svelte` | Settings → Animations: Full, Short or Off for the generic table effects, animations.effects (ADR 0209 §1, #4396) |
| `effectBadgeDestroyed` | `Destroyed` | status | text | `lib/components/board/TableEffects.svelte` | the still badge a destroyed permanent leaves where it was, when effect motion is off or reduced (ADR 0209 §1) |
| `effectBadgeDied` | `Died` | status | text | `lib/components/board/TableEffects.svelte` | the still badge for a permanent that died to a state-based action (lethal damage, zero toughness) |
| `effectBadgeExiled` | `Exiled` | status | text | `lib/components/board/TableEffects.svelte` | the still badge for a permanent that was exiled, so an exile is never read as a death |
| `effectBadgeSacrificed` | `Sacrificed` | status | text | `lib/components/board/TableEffects.svelte` | the still badge for a sacrificed permanent |
| `effectBadgeBounced` | `Returned to hand` | status | text | `lib/components/board/TableEffects.svelte` | the still badge for a permanent that went back to a hand |
| `effectBadgePwEnter` | `Planeswalker arrived` | status | text | `lib/components/board/TableEffects.svelte` | the still badge for a planeswalker entering the battlefield |
| `effectBadgePwLeave` | `Planeswalker gone` | status | text | `lib/components/board/TableEffects.svelte` | the still badge for a planeswalker that died |
| `effectBadgeWipe` | `Board wipe: <N> permanent(s)` (starts with `Board wipe: `) | status | text | `lib/components/board/TableEffects.svelte` | the one still badge a board wipe leaves over the table (three or more permanents leaving in one batch from one source) |
| `effectExtort` | `Extort` | status | text | `lib/components/board/ExtortScales.svelte` | the caption under the Scales of Souls, the table effect an extort drain plays (#4544); the whole scene is aria-hidden, the log and the life popups carry the news |
| `effectExtortMany` | `Extort ×<N>` (starts with `Extort ×`) | status | text | `lib/components/board/ExtortScales.svelte` | the same caption when several extort triggers resolved in one scale's wait and were merged into it |
| `startPractice` | `Start practice` | link | text | `routes/LobbyPractice.hint.ts`<br>`routes/Lobby.svelte` | the lobby.practice tip's action, to #/practice, and the Lobby's Practice game card; the tip's Got it reads Not now beside it |
| `tip` | `tip` | complementary | aria | `lib/components/hints/HintCard.svelte` | a first-use hint's card (ADR 0125 §3.6); one at a time, and Go to the tip (i) moves focus to it |
| `keyboardShortcuts` | `keyboard shortcuts` | dialog | text | `lib/components/ShortcutsOverlay.svelte` | the keymap overlay (?), named by its heading |
| `rememberThisAnswer` | `Remember this answer` | button | text | `lib/choiceDock.ts` | the toggle under a prompt that can take a standing answer; on, the next button pressed also sets the rule (Yes or Pay: Always; No or Don't pay: Never) |
| `recommendedAnswer` | `Recommended` | generic | text | `lib/choiceDock.ts` | the tag after the recommended answer of the which-skip question, "Use my skip" (ADR 0188 Decision 12); Enter takes it |
| `choosingNextTurn` | `Choosing the next turn` | generic | text | `lib/components/board/PhaseDisplay.svelte` | the step label while the turn change waits on the which-skip question (turn.between_turns, ADR 0188 Decision 12): the old turn is over and the next has not begun |
| `automaticAnswer` | `automatic answer` | dialog | aria | `lib/autoAnswerPref.ts` | the dock's notice after the server answered a prompt for you ("Rhystic Study: paid {1} for you"), for about six seconds |
| `undoAutomaticAnswer` | `Undo` | button | text | `lib/autoAnswerPref.ts` | the automatic-answer notice's undo; greyed once anything else is the top undo entry |
| `autoPassNotice` | `auto-pass notice` | dialog | aria | `lib/passReason.ts` | the dock's chip after the client passed priority for you ("Passed for you: nothing to respond with", #3343), for about six seconds |
| `undoAutoPass` | `Take back the pass` | button | aria | `lib/passReason.ts` | the auto-pass chip's Undo (visible text "Undo"); sends the ordinary undo and is shown only while that undo would take the pass back. Named apart from the dock's "Undo (N left)" so a /^Undo/ locator still finds one button |
| `askMeNextTime` | `Ask me next time` | button | text | `lib/autoAnswerPref.ts` | the automatic-answer notice's button that removes the rule (sets Ask); the answer just given stands |
| `automaticAnswers` | `automatic answers` | region | aria | `lib/components/AutoAnswersSettings.svelte` | Settings → Gameplay's list of standing answers; the settings.auto-answers hint points at it |
| `forgetAutoAnswer` | `Forget` | button | text | `lib/components/AutoAnswersSettings.svelte` | removes one standing answer in Settings → Gameplay, so its prompt is asked again |
| `useThisOrderAgain` | `Use this order again` | button | text | `lib/components/board/ChoicePromptModal.svelte` | the toggle under the trigger-order list (#3016, ADR 0018's eighth amendment); on, the next batch of the same abilities is ordered the same way for you. Disabled on an order that splits one card's triggers apart |
| `searchLibraryAndGraveyard` | `Library and graveyard` | button | text | `lib/components/board/ChoicePromptModal.svelte` | the primary of the question a "search your library and/or graveyard" prompt asks before it shows anything (#4519, CR 701.23): search both zones. The library is shuffled afterwards |
| `searchLibraryAndHand` | `Library and hand` | button | text | `lib/components/board/ChoicePromptModal.svelte` | the primary of the question a "search your hand and/or library" prompt asks before it shows anything (#4519, CR 701.23): search both zones. The library is shuffled afterwards |
| `searchLibraryOnly` | `Library only` | button | text | `lib/components/board/ChoicePromptModal.svelte` | an and/or search's zone question (#4519): search the library alone; it is shuffled afterwards |
| `searchGraveyardOnly` | `Graveyard only` | button | text | `lib/components/board/ChoicePromptModal.svelte` | a "search your library and/or graveyard" prompt zone question (#4519): search the graveyard alone; the library is never shown or shuffled |
| `searchHandOnly` | `Hand only` | button | text | `lib/components/board/ChoicePromptModal.svelte` | a "search your hand and/or library" prompt zone question (#4519): search the hand alone; the library is never shown or shuffled |
| `forgetTriggerOrders` | `Forget remembered trigger orders (<N>)` (starts with `Forget remembered trigger orders`) | menuitem | text | `lib/components/board/GameMenu.svelte` | the ⋯ menu's entry that forgets every trigger order you asked to reuse this game (#3016); shown while you have one |
| `volumeMaster` | `Master volume` | generic | text | `lib/components/Settings.svelte` | the slider in Settings → Audio that sits over every category and the music |
| `volumeInterface` | `Interface volume` | generic | text | `lib/components/Settings.svelte` | the slider for your-turn, priority, prompt and timer cues (ADR 0209 §2); an interface cue ducks the other buses while it plays |
| `volumeCards` | `Cards volume` | generic | text | `lib/components/Settings.svelte` | the slider for draw, play, tap, untap and shuffle sounds (what Effects volume was before settings v29) |
| `volumeCombat` | `Combat volume` | generic | text | `lib/components/Settings.svelte` | the slider for attack, block and combat damage sounds |
| `volumeEffects` | `Effects volume` | generic | text | `lib/components/Settings.svelte` | the slider for the generic effect cues and signature card effects (ADR 0209 §3, §4) |
| `volumeVoice` | `Voice volume` | generic | text | `lib/components/Settings.svelte` | the slider for planeswalker voice lines (ADR 0209 §5); none ship yet |
| `volumeMusic` | `Music volume` | generic | text | `lib/components/Settings.svelte` | the slider for the ambient music bed |
| `damageAutoAssigned` | `combat damage assigned` | dialog | aria | `lib/damageAssignment.ts` | the dock's notice after Auto-assign combat damage assigned an attacker's damage for you (lethal to each blocker, the rest over), for about six seconds |
| `alwaysAskCombatDamage` | `Always ask` | button | text | `lib/damageAssignment.ts` | the combat-damage notice's button that turns Auto-assign combat damage off; the assignment just made stands |
| `addDamage` | `Add 1 damage to <name>` (starts with `Add 1 damage to `) | button | aria | `lib/components/board/DamageStepper.svelte`<br>`lib/components/board/ChoicePromptModal.svelte` | the + on a blocker (on the board and in the damage sheet) while you assign combat damage; the defending player's trample line uses it too |
| `removeDamage` | `Remove 1 damage from <name>` (starts with `Remove 1 damage from `) | button | aria | `lib/components/board/DamageStepper.svelte`<br>`lib/components/board/ChoicePromptModal.svelte` | the − beside addDamage |
| `awayOffer` | `move on without an away player` | dialog | aria | `lib/away.ts` | the dock sheet a decider sees when a seat the table waits on has been offline for the away time (ADR 0177 §2) |
| `awayStandIn` | `Hand <name>'s seat to a bot` (ends with `'s seat to a bot`) | button | text | `lib/away.ts` | the away offer's hand-over: a heuristic bot plays the seat until its player takes it back |
| `awayConcede` | `Concede for <name>` (starts with `Concede for `) | button | text | `lib/away.ts` | the away offer's concession, made for the away player |
| `awayWait` | `Keep waiting` | button | text | `lib/away.ts` | the away offer's third answer: hide the offer for another away time |
| `takeBackSeat` | `Take it back` | button | text | `lib/components/board/StandInBanner.svelte` | the banner a player sees on their own seat while a bot plays it: they play it again (ADR 0177 §6) |
| `holdShiftToFlip` | `Hold Shift to flip` | note | text | `lib/components/board/HoverZoomOverlay.svelte` | the hover zoom's hint under a two-faced card's inset, on pointer devices only (#4106): holding Shift shows the other face big |
