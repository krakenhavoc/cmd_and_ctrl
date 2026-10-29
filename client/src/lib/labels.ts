// labels.ts — the contract labels (ADR 0125 §2, ADR 0076 §2.4, ADR 0111 §10).
//
// The names the tutorial, the hints and the e2e suite rely on, one entry
// each, with the files that render it. A component renders a registered
// name as `L.<key>` (or `L.<key>(…)` for one with a variable part), never
// as a literal, so a rename is an edit to one value here and every
// anchor follows it. labels.test.ts holds the line:
//
//   1. each entry's owner files reference `L.<key>`;
//   2. no other file under client/src carries a literal copy of a
//      registered `aria` name;
//   3. every tutorial anchor and every first-use hint's anchor is
//      registered here;
//   4. docs/labels.md, which is generated from this file, is current
//      (`UPDATE_LABELS_DOC=1 npm test -- labels` rewrites it).
//
// The registry is the contract, not an inventory: other aria-labels stay
// literals. What goes in is every label an anchor points at, every label
// the e2e walks select on, and every name AGENTS.md §5 used to list.
//
// This file imports nothing, so the e2e package can import it too.

declare const brand: unique symbol;

/**
 * `aria`: an aria-label, or a dialog's or group's name given through a
 * prop that becomes one. `text`: a name a button, link or menu item gets
 * from its visible text. Only `aria` names can be anchors, because the
 * anchor resolver matches `[aria-label]`.
 */
export type LabelKind = "aria" | "text";

/**
 * A registered name, exactly as rendered. At run time it is the string
 * itself; the brand only exists for the type checker, so that an anchor
 * cannot be written as a raw string.
 */
export type Name<K extends LabelKind = LabelKind> = string & { readonly [brand]: K };

/** Where a dynamic label's fixed part sits in the rendered name. */
export type StemMatch = "prefix" | "suffix" | "contains";

/**
 * A dynamic label used as an anchor without its variable part: "any
 * element whose aria-label has this stem". The resolver turns `match`
 * into the CSS operators `^=`, `$=` or `*=`.
 */
export interface Stem {
  readonly [brand]: "stem";
  readonly key: string;
  readonly stem: string;
  readonly match: StemMatch;
}

/**
 * What an anchor's `label` and `within` take: a registered `aria` name
 * (exact match) or a dynamic `aria` entry's stem. Only `L` produces
 * either, so `npm run check` refuses an anchor spelled as a raw string.
 */
export type LabelRef = Name<"aria"> | Stem;

interface Common<K extends LabelKind> {
  kind: K;
  /** The element's role, for docs/labels.md: region, group, dialog, button, … */
  role: string;
  /** The files, relative to client/src, that render it. */
  owners: readonly string[];
  /** One line for docs/labels.md: what it is and who relies on it. */
  doc: string;
}

export interface StaticSpec<K extends LabelKind = LabelKind> extends Common<K> {
  readonly type: "static";
  name: string;
}

export interface DynamicSpec<K extends LabelKind, A extends unknown[]> extends Common<K> {
  readonly type: "dynamic";
  /** The fixed part every rendered name contains. */
  stem: string;
  match: StemMatch;
  /** The name's shape for docs/labels.md, with <placeholders>. */
  shape: string;
  make: (...args: A) => string;
  /** Arguments the guard calls `make` with, to check it keeps the stem. */
  example: A;
}

/** Any dynamic entry, whatever its arguments. */
export interface AnyDynamicSpec extends Common<LabelKind> {
  readonly type: "dynamic";
  stem: string;
  match: StemMatch;
  shape: string;
  make: (...args: never) => string;
  example: readonly unknown[];
}

export type LabelSpec = StaticSpec | AnyDynamicSpec;

/** A dynamic entry: call it for the full name; `.any` is its stem, for an anchor. */
export type DynamicLabel<K extends LabelKind, A extends unknown[]> = ((...args: A) => Name<K>) &
  (K extends "aria" ? { readonly any: Stem } : unknown);

function label<K extends LabelKind>(spec: Omit<StaticSpec<K>, "type">): StaticSpec<K> {
  return { type: "static", ...spec };
}

function dynamicLabel<K extends LabelKind, A extends unknown[]>(
  spec: Omit<DynamicSpec<K, A>, "type">,
): DynamicSpec<K, A> {
  return { type: "dynamic", ...spec };
}

// "A", "A and B", "A, B and C".
const andList = (names: readonly string[]): string =>
  names.length <= 1
    ? (names[0] ?? "")
    : `${names.slice(0, -1).join(", ")} and ${names[names.length - 1]}`;

const plural = (n: number, word: string): string => `${n} ${word}${n === 1 ? "" : "s"}`;

// ---- The registry ----

const BOARD = "lib/components/board/";

/**
 * Every contract label. Keys are camelCase and stable; a rename changes
 * the value, not the key. docs/labels.md lists them in this order.
 */
export const LABEL_SPECS = {
  // -- The table: the viewer's own board (the tutorial's anchors) --
  yourBoard: label({
    name: "your board",
    kind: "aria",
    role: "region",
    owners: [`${BOARD}PlayerPanel.svelte`],
    doc: "the viewer's own board panel; scopes the tutorial's lands and creatures anchors",
  }),
  seatBoard: dynamicLabel({
    stem: " board",
    match: "suffix",
    shape: "<name> board",
    make: (name: string) => `${name} board`,
    example: ["Practice Bot"],
    kind: "aria",
    role: "region",
    owners: [`${BOARD}PlayerPanel.svelte`, `${BOARD}SeatSummary.svelte`],
    doc: "another seat's board panel; the tutorial's attack step points at the bot's while it plays (the stem also ends other names, so anchor by the full name)",
  }),
  reflowNextUp: label({
    name: "Next up",
    kind: "text",
    role: "generic",
    owners: ["lib/expansion.ts"],
    doc: "ADR 0077 (amended 2026-10-17, #4102): the chip on a reflow avatar card whose seat plays straight after the active player (drawn upper case)",
  }),
  reflowAfterYou: label({
    name: "After you",
    kind: "text",
    role: "generic",
    owners: ["lib/expansion.ts"],
    doc: "ADR 0077 (amended 2026-10-17, #4102): the chip on a reflow avatar card whose seat plays straight after the viewer (drawn upper case)",
  }),
  reflowAfter: dynamicLabel({
    stem: "After ",
    match: "prefix",
    shape: "After <name>",
    make: (name: string) => `After ${name}`,
    example: ["Practice Bot"],
    kind: "text",
    role: "generic",
    owners: ["lib/expansion.ts"],
    doc: "ADR 0077 (amended 2026-10-17, #4102): the chip on a reflow avatar card naming the opponent who plays right before it (drawn upper case)",
  }),
  yourHand: label({
    name: "your hand",
    kind: "aria",
    role: "generic",
    owners: [`${BOARD}Hand.svelte`],
    doc: "the viewer's hand; the tutorial's read-hand, play-land and cast-creature steps anchor here",
  }),
  lands: label({
    name: "lands",
    kind: "aria",
    role: "list",
    owners: [`${BOARD}PlayerPanel.svelte`],
    doc: "a board's land row; every board has one, so anchor it within your board",
  }),
  creatures: label({
    name: "creatures",
    kind: "aria",
    role: "list",
    owners: [`${BOARD}PlayerPanel.svelte`],
    doc: "a board's creature row; every board has one, so anchor it within your board",
  }),
  attractionDeck: label({
    name: "attractions",
    kind: "text",
    role: "button",
    owners: [`${BOARD}PileBar.svelte`],
    doc: "a seat's face-down Attraction deck in the rail (ADR 0207, CR 717.2), drawn only while it holds a card; a count, not openable",
  }),
  junkyard: label({
    name: "junkyard",
    kind: "text",
    role: "button",
    owners: [`${BOARD}PileBar.svelte`],
    doc: "a seat's face-up junkyard in the rail (ADR 0207, CR 717.6a), drawn only while it holds a card; opens the zone browser's junkyard tab",
  }),
  tickets: label({
    name: "tickets",
    kind: "aria",
    role: "status",
    owners: [`${BOARD}PlayerIdentity.svelte`],
    doc: "a seat's ticket counters ({TK}, ADR 0208, CR 107.17) in its marker column, drawn only while it has one; a count, no stepper",
  }),
  enduringStory: label({
    name: "an enduring story",
    kind: "aria",
    role: "status",
    owners: [`${BOARD}PlayerIdentity.svelte`],
    doc: "a seat's enduring story (CR 702.195b, #4475) in its marker column, drawn only once storied has given it; read-only, there is no toggle",
  }),
  stickerSheets: label({
    name: "sticker sheets",
    kind: "aria",
    role: "status",
    owners: [`${BOARD}PlayerIdentity.svelte`],
    doc: "a seat's three sticker sheets (ADR 0208, CR 123.2c): how many stickers are still free, the sheets and their stickers on hover",
  }),
  stickersOnCard: label({
    name: "stickers",
    kind: "aria",
    role: "status",
    owners: [`${BOARD}StickerBadge.svelte`],
    doc: "the stickers on a card (ADR 0208, CR 123): a count on the card's left edge, the stickers on hover",
  }),
  commandZone: dynamicLabel({
    stem: " command zone, ",
    match: "contains",
    shape: "<name> command zone, <N> card(s)",
    make: (seat: string, n: number) => `${seat} command zone, ${plural(n, "card")}`,
    example: ["Player", 1],
    kind: "aria",
    role: "group",
    owners: [`${BOARD}ExileStrip.svelte`, `${BOARD}CommandStrip.svelte`],
    doc: "a seat's command zone beside its hand (#2349): yours in the castable strip, others' in their command strip; the tutorial's commander step and the table.commander hint anchor to yours, within your board",
  }),
  companion: dynamicLabel({
    stem: " companion, ",
    match: "contains",
    shape: "<seat> companion, <card name>",
    make: (seat: string, card: string) => `${seat} companion, ${card}`,
    example: ["Player", "Jegantha, the Wellspring"],
    kind: "aria",
    role: "group",
    owners: [`${BOARD}CompanionSlot.svelte`],
    doc: "a seat's revealed companion while it waits outside the game (ADR 0172), beside its commanders",
  }),
  takeCompanion: dynamicLabel({
    stem: " into your hand (companion)",
    match: "suffix",
    shape: "Pay <cost>: put <card name> into your hand (companion)",
    make: (cost: string, card: string) => `Pay ${cost}: put ${card} into your hand (companion)`,
    example: ["{3}", "Jegantha, the Wellspring"],
    kind: "aria",
    role: "button",
    owners: [`${BOARD}CompanionSlot.svelte`],
    doc: "the companion special action (CR 116.2g, ADR 0172): pay {3} to put your companion into your hand; greyed outside your main phase",
  }),
  ignoreEffect: dynamicLabel({
    stem: " (sacrifice a permanent)",
    match: "suffix",
    shape: "Ignore <card name> (sacrifice a permanent)",
    make: (card: string) => `Ignore ${card} (sacrifice a permanent)`,
    example: ["Damping Engine"],
    kind: "aria",
    role: "button",
    owners: [`${BOARD}IgnoreEffects.svelte`],
    doc: "the CR 116.2d special action (#1986): sacrifice a permanent of your choice to ignore a static ability's effect until end of turn (Damping Engine); on your own panel while the effect bears on you, and it opens the sacrifice picker",
  }),
  expandBoard: dynamicLabel({
    stem: "Expand ",
    match: "prefix",
    shape: "Expand <name>'s board, or Expand your own board",
    make: (name: string | null) =>
      name === null ? "Expand your own board" : `Expand ${name}'s board`,
    example: ["Practice Bot"],
    kind: "aria",
    role: "button",
    owners: [`${BOARD}Board.svelte`],
    doc: "a board's expand button (ADR 0120); null names the viewer's own",
  }),
  stackPile: dynamicLabel({
    stem: "stack: ",
    match: "prefix",
    shape: "stack: <N> on the stack",
    make: (n: number) => `stack: ${n} on the stack`,
    example: [2],
    kind: "aria",
    role: "region",
    owners: [`${BOARD}StackLaneHost.svelte`, `${BOARD}StackOverlay.svelte`],
    doc: "the stack pile (ADR 0119), and the floating stack overlay; the tutorial's stack step anchors here",
  }),
  stackAbility: dynamicLabel({
    stem: " on the stack: ",
    match: "contains",
    shape: "Activate <card name> on the stack: <cost>",
    make: (card: string, cost: string) => `Activate ${card} on the stack: ${cost}`,
    example: ["Lightning Storm", "Discard a land card"],
    kind: "aria",
    role: "button",
    owners: [`${BOARD}StackAbilityButtons.svelte`],
    doc: "#1865: a spell's ability that works while it is on the stack (Lightning Storm's discard a land card), on every stack surface beside Counter; offered to each seat that may activate it",
  }),
  stackAbilityGone: dynamicLabel({
    stem: " left the stack",
    match: "suffix",
    shape: "<card name> left the stack",
    make: (card: string) => `${card} left the stack`,
    example: ["Lightning Storm"],
    kind: "text",
    role: "status",
    owners: [`${BOARD}Board.svelte`],
    doc: "#4043: the notice that closes the discard picker when the spell whose stack ability was being announced resolves or is countered first",
  }),
  attention: label({
    name: "attention",
    kind: "aria",
    role: "region",
    owners: [`${BOARD}Board.svelte`],
    doc: "the attention strip, top left: bot chatter, reveals, the roll call",
  }),
  turnPhase: label({
    name: "turn and phase indicator",
    kind: "aria",
    role: "generic",
    owners: [`${BOARD}PhaseDisplay.svelte`],
    doc: "the turn number, the active player and the step",
  }),
  abilityChip: dynamicLabel({
    stem: " abilit",
    match: "contains",
    shape: "<N> <kind> ability, or <N> <kind> abilities (kind: triggered, static or activated)",
    make: (n: number, kind: "triggered" | "static" | "activated") =>
      `${n} ${kind} ${n === 1 ? "ability" : "abilities"}`,
    example: [2, "triggered"],
    kind: "aria",
    role: "button",
    owners: [`${BOARD}AbilityKindChips.svelte`],
    doc: "an art tile's ability chip (#2219): one per kind the card has, its hover or focus lists the server's labels; the e2e suite reads Mulldrifter's",
  }),

  crewedBadge: dynamicLabel({
    stem: "crewed ",
    match: "prefix",
    shape: "crewed by <names>, or crewed this turn",
    make: (names: readonly string[]) =>
      names.length > 0 ? `crewed by ${andList(names)}` : "crewed this turn",
    example: [["Llanowar Elves", "Birds of Paradise"]],
    kind: "aria",
    role: "generic",
    owners: [`${BOARD}Card.svelte`],
    doc: "a Vehicle's crewed badge (ADR 0186 §6): the creatures tapped this turn to crew it; the names come from the board, and an ID with no card there is left out",
  }),
  saddledBadge: dynamicLabel({
    stem: "saddled ",
    match: "prefix",
    shape: "saddled by <names> until end of turn, or saddled until end of turn",
    make: (names: readonly string[]) =>
      names.length > 0
        ? `saddled by ${andList(names)} until end of turn`
        : "saddled until end of turn",
    example: [["Llanowar Elves", "Birds of Paradise"]],
    kind: "aria",
    role: "generic",
    owners: [`${BOARD}Card.svelte`],
    doc: "a Mount's saddled badge (ADR 0186 §6): the creatures tapped this turn to saddle it; the names come from the board, and an ID with no card there is left out",
  }),

  // -- The action dock (ADR 0111) --
  actions: label({
    name: "actions",
    kind: "aria",
    role: "region",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the action dock; the tutorial's move-along step and its detours anchor here",
  }),
  waitToRespond: label({
    name: "wait, let me respond",
    kind: "aria",
    role: "button",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the stack hold countdown's one-click hold (#2853): keeps priority on the stack auto-pass was about to pass",
  }),
  jumpChip: dynamicLabel({
    stem: "Your decision on ",
    match: "prefix",
    shape: "Your decision on <Bob's | your own> board",
    make: (whose: string) => `Your decision on ${whose} board`,
    example: ["Bob's"],
    kind: "aria",
    role: "status",
    owners: [`${BOARD}HandoffChip.svelte`],
    doc: "the small chip over the dock while the viewer plays another seat's turn and the next decision is owed on the board not shown (#4434); the board never switches by itself",
  }),
  jumpToBoard: label({
    name: "Jump",
    kind: "text",
    role: "button",
    owners: [`${BOARD}HandoffChip.svelte`],
    doc: "the jump chip's button: switches to the board that owes the decision (#4434)",
  }),
  priorityControls: label({
    name: "priority controls",
    kind: "aria",
    role: "group",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's toggles: Skip to my turn, undo, the ⋯ menu",
  }),
  skipToMyTurn: label({
    name: "Skip to my turn",
    kind: "aria",
    role: "button",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's Skip to my turn toggle (ADR 0143 §4.2, formerly autopass; state in aria-pressed): passes for you until your next main phase; the tutorial's let-the-bot-play step anchors here, within actions",
  }),
  recordLoop: label({
    name: "Record loop",
    kind: "text",
    role: "menuitem",
    owners: [`${BOARD}GameMenu.svelte`],
    doc: "the dock's ⋯ menu entry while you hold priority (ADR 0193 first amendment, CR 732.2a): every action from now on, any seat's, is one iteration of a loop; once recording, Stop and Cancel are the dock's status line",
  }),
  stopLoop: label({
    name: "Stop recording the loop",
    kind: "aria",
    role: "button",
    owners: [`${BOARD}LoopControl.svelte`],
    doc: "ends the demonstration (ADR 0193): the server checks the board came back, keeps the loop to propose, or discards it and says why in the log",
  }),
  cancelLoop: label({
    name: "Cancel the loop",
    kind: "aria",
    role: "button",
    owners: [`${BOARD}LoopControl.svelte`, `${BOARD}LoopProposeSheet.svelte`],
    doc: "discards the recording and any proposal still being answered (ADR 0193)",
  }),
  proposeLoop: label({
    name: "Propose the loop",
    kind: "aria",
    role: "dialog",
    owners: [`${BOARD}LoopProposeSheet.svelte`],
    doc: "the recorder's sheet once the loop closed (ADR 0193 §9): the iteration, its net change and how many more times to repeat it",
  }),
  loopProposal: label({
    name: "Loop shortcut proposal",
    kind: "aria",
    role: "dialog",
    owners: [`${BOARD}LoopProposalSheet.svelte`],
    doc: "another player's proposal to repeat a loop (CR 732.2b): accept it, or shorten it to one of your own steps",
  }),
  pass: label({
    name: "Pass",
    kind: "aria",
    role: "button",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's primary while the stack holds something (#3267): pass priority so the top item can resolve (its visible text and its aria-label); the tutorial's stack step anchors here, within actions",
  }),
  next: label({
    name: "Next",
    kind: "aria",
    role: "button",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's primary on an empty stack (#3267): pass priority so the game moves to the next step (its visible text and its aria-label); the tutorial's stack step anchors here too, within actions",
  }),
  nextEndsTurn: label({
    name: "Pass · turn ends",
    kind: "aria",
    role: "button",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's primary at your last priority of your own turn (#3826, turn.pass_hint ends_turn): a pass hands the turn over; ringed, and named so End turn stays the one button of that name (its visible text and its aria-label)",
  }),
  nextYourTurn: label({
    name: "Pass · your turn next",
    kind: "aria",
    role: "button",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's primary at your last priority before your turn begins (#3826, turn.pass_hint your_turn_next): your next decision is in your own turn; ringed (its visible text and its aria-label)",
  }),
  attackWith: dynamicLabel({
    stem: "Attack with ",
    match: "prefix",
    shape: "Attack with <N>",
    make: (n: number) => `Attack with ${n}`,
    example: [2],
    kind: "aria",
    role: "button",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's primary in declare attackers while creatures are staged (#3266): sends them as one declaration, then passes (its visible text and its aria-label)",
  }),
  noAttacks: label({
    name: "No attacks",
    kind: "aria",
    role: "button",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's primary in declare attackers while creatures could attack and none has (#3267): pass priority with no attackers, which skips the rest of combat (CR 508.8)",
  }),
  attackAll: dynamicLabel({
    stem: " with all ",
    match: "contains",
    shape: "Attack <name> with all <N> creature(s)",
    make: (name: string, n: number) => `Attack ${name} with all ${plural(n, "creature")}`,
    example: ["Practice Bot", 2],
    kind: "aria",
    role: "button",
    owners: ["lib/attackAll.ts"],
    doc: "the dock's attack-with-all button in a duel (its aria-label equals its visible text); the tutorial's attack step anchors here, within actions",
  }),
  endTurn: label({
    name: "End turn",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's secondary, shown to the active player only (ADR 0143 §4.2, formerly Pass turn): passes for you through every step of the rest of your turn",
  }),
  sandboxPassTurn: label({
    name: "Skip to next turn (sandbox)",
    kind: "text",
    role: "menuitem",
    owners: [`${BOARD}GameMenu.svelte`],
    doc: "the ⋯ menu's sandbox jump to the next turn (pass_turn): skips the end step, its triggers and the cleanup discard, and the log says so",
  }),
  declareAttackers: label({
    name: "declare attackers",
    kind: "aria",
    role: "dialog, group",
    owners: ["lib/combatDock.ts"],
    doc: "the attack request's dialog and its row group",
  }),
  attackPlain: label({
    name: "Attack",
    kind: "text",
    role: "button",
    owners: ["lib/combatDock.ts"],
    doc: "ADR 0130 §7: the dock's choice for a creature that may be exerted, attacking without exerting it",
  }),
  attackAndExert: label({
    name: "Attack and exert",
    kind: "text",
    role: "button",
    owners: ["lib/combatDock.ts"],
    doc: "ADR 0130 §7: the dock's choice for a creature that may be exerted, attacking and exerting it",
  }),
  declareAttackerAndExert: label({
    name: "Declare attacker and exert",
    kind: "text",
    role: "menuitem",
    owners: ["lib/contextMenu.logic.ts"],
    doc: "ADR 0130 §7: the card menu's row beside Declare attacker, for a creature that may be exerted",
  }),
  exertAttacker: dynamicLabel({
    stem: "Exert ",
    match: "prefix",
    shape: "Exert <name>",
    make: (name: string) => `Exert ${name}`,
    example: ["Oketra's Avenger"],
    kind: "aria",
    role: "switch",
    owners: [`${BOARD}AttackDeclarationModal.svelte`],
    doc: "ADR 0130 §7: the Choose attackers picker's per-row Exert toggle",
  }),
  attackPayWithLife: label({
    name: "pay the attack tax with life",
    kind: "aria",
    role: "group",
    owners: [`${BOARD}AttackLifeStepper.svelte`],
    doc: "#4198: the stepper that pays a Phyrexian attack tax (Norn's Annex's {W/P}) with 2 life per symbol; drawn in the Choose attackers picker and in the Pay the attack tax sheet",
  }),
  declareBlockers: label({
    name: "declare blockers",
    kind: "aria",
    role: "dialog, group",
    owners: ["lib/combatDock.ts"],
    doc: "the block request's dialog and its row group",
  }),
  removeBlock: label({
    name: "Remove block",
    kind: "text",
    role: "button",
    owners: ["lib/combatDock.ts"],
    doc: "#3898: the dock's button that takes the selected blocker's block back while blockers are declared on the board",
  }),
  chooseBlockers: label({
    name: "choose blockers",
    kind: "aria",
    role: "dialog",
    owners: [`${BOARD}BlockListSheet.svelte`],
    doc: "#3898: the block declaration as a list, opened from the block request's Show as list; the keyboard path",
  }),

  // -- Dock requests --
  mulligan: label({
    name: "keep or mulligan your hand",
    kind: "aria",
    role: "dialog",
    owners: ["routes/Game.svelte"],
    doc: "the opening hand sheet; the tutorial's play-land and cast-creature steps detour here while it waits",
  }),
  keepHand: label({
    name: "Keep hand",
    kind: "text",
    role: "button",
    owners: ["routes/Game.svelte"],
    doc: "the opening hand sheet's primary",
  }),
  openingHandDecisions: label({
    name: "opening hand decisions",
    kind: "aria",
    role: "generic",
    owners: ["routes/Game.svelte"],
    doc: "the strip's roll call while hands are kept or mulliganed",
  }),
  discard: dynamicLabel({
    stem: "Discard ",
    match: "prefix",
    shape: "Discard <N> card(s)",
    make: (n: number) => `Discard ${plural(n, "card")}`,
    example: [2],
    kind: "aria",
    role: "dialog",
    owners: [`${BOARD}DiscardPromptModal.svelte`],
    doc: "the discard request (cleanup or an effect)",
  }),
  crewCountsAs: dynamicLabel({
    stem: "counts as ",
    match: "prefix",
    shape: "counts as <N>",
    make: (n: number) => `counts as ${n}`,
    example: [3],
    kind: "text",
    role: "generic",
    owners: [`${BOARD}CrewCostModal.svelte`],
    doc: "ADR 0186 Decision 6 (#3550): the crew or saddle picker's note on a creature that counts for something other than its power on this cost",
  }),
  selectTarget: dynamicLabel({
    stem: "Select target for ",
    match: "prefix",
    shape: "Select target for <card>",
    make: (card: string) => `Select target for ${card}`,
    example: ["Lightning Bolt"],
    kind: "aria",
    role: "dialog",
    owners: ["lib/targetingDock.ts"],
    doc: "the targeting request",
  }),
  targetPlayersLead: label({
    name: "Players \u2192",
    kind: "text",
    role: "generic",
    owners: ["lib/targetingDock.ts"],
    doc: "#4435: the lead of the targeting request's row of legal players, so a player a dialog or a hand covers on the board is still one press away",
  }),
  targetPlayer: dynamicLabel({
    stem: "Target ",
    match: "prefix",
    shape: "Target <player>",
    make: (name: string) => `Target ${name}`,
    example: ["Practice Bot"],
    kind: "aria",
    role: "button",
    owners: ["lib/targetingDock.ts"],
    doc: "#4435: the targeting request's button for one legal player (a seat-colour dot and the name); it picks that player exactly as a click on their avatar does",
  }),
  chooseOnBoard: dynamicLabel({
    stem: "Choose on the board for ",
    match: "prefix",
    shape: "Choose on the board for <card>",
    make: (card: string) => `Choose on the board for ${card}`,
    example: ["Finale of Revelation"],
    kind: "aria",
    role: "dialog",
    owners: ["lib/targetingDock.ts"],
    doc: "a card-set pick answered by clicking permanents on the board (#2394)",
  }),
  showAsList: label({
    name: "Show as list",
    kind: "text",
    role: "button",
    owners: ["lib/targetingDock.ts", "lib/combatDock.ts"],
    doc: "a board-answered card-set pick's fallback: open the same choice as the list; #3898: the block request's way to the blockers list",
  }),
  castPayingOnPaper: dynamicLabel({
    stem: " and pay the rest on paper?",
    match: "suffix",
    shape: "Cast <card> and pay the rest on paper?",
    make: (card: string) => `Cast ${card} and pay the rest on paper?`,
    example: ["Craw Wurm"],
    kind: "aria",
    role: "dialog",
    owners: ["lib/targetingDock.ts"],
    doc: "Cast anyway's confirm (ADR 0118 §2, renamed by #3855), with its Cast and Cancel",
  }),
  paperPayment: dynamicLabel({
    stem: "What did you pay for ",
    match: "prefix",
    shape: "What did you pay for <card>?",
    make: (card: string) => `What did you pay for ${card}?`,
    example: ["Painful Truths"],
    kind: "aria",
    role: "dialog",
    owners: [`${BOARD}PaperPaymentModal.svelte`],
    doc: "the paper payment picker (#3855): what a waived payment something reads paid outside the pool",
  }),
  paperSlotOnX: label({
    name: "on X",
    kind: "text",
    role: "generic",
    owners: [`${BOARD}PaperPaymentModal.svelte`],
    doc: "the paper payment picker's tag on a slot of the cost's {X} part, the mana a spend-only-on-X clause restricts (#3941)",
  }),
  castFree: label({
    name: "Cast without paying its mana cost",
    kind: "text",
    role: "button",
    owners: [`${BOARD}PaperPaymentModal.svelte`],
    doc: "the paper payment picker's other answer: an effect on paper let you cast it free (#3855)",
  }),
  cast: label({
    name: "Cast",
    kind: "text",
    role: "button",
    owners: ["lib/targetingDock.ts"],
    doc: "Cast anyway's confirm: cast it, the pool paying what it can and the rest paid on paper",
  }),
  cancel: label({
    name: "Cancel",
    kind: "text",
    role: "button",
    owners: ["lib/targetingDock.ts", "lib/dock.ts"],
    doc: "Cast anyway's confirm, and every flow's Cancel (the dock's cancelAction), the first-turn confirm's included",
  }),
  castAnyway: label({
    name: "Cast anyway (pay on paper)",
    kind: "text",
    role: "menuitem",
    owners: ["lib/castAnyway.ts"],
    doc: "the card menu's row that casts what the pool can't pay, the rest paid on paper (ADR 0118 §2, renamed by #3855)",
  }),
  energyToPay: label({
    name: "energy to pay",
    kind: "aria",
    role: "group",
    owners: [`${BOARD}ChoicePromptModal.svelte`],
    doc: "the pay_amount prompt's stepper (ADR 0129 §3): how much energy to pay, from its floor to the seat's energy",
  }),
  payEnergy: dynamicLabel({
    stem: " {E}",
    match: "suffix",
    shape: "Pay <N> {E}",
    make: (n: number) => `Pay ${n} {E}`,
    example: [3],
    kind: "text",
    role: "button",
    owners: ["lib/choiceDock.ts"],
    doc: "the pay_amount prompt's primary: pays the stepper's amount",
  }),
  payNothing: label({
    name: "Pay nothing",
    kind: "text",
    role: "button",
    owners: ["lib/choiceDock.ts"],
    doc: "the pay_amount prompt's decline when the card says any amount; it reads Don't pay when the card says one or more",
  }),
  lifeToPay: label({
    name: "life to pay",
    kind: "aria",
    role: "group",
    owners: [`${BOARD}ChoicePromptModal.svelte`],
    doc: "the pay_amount prompt's stepper when it asks for life (#1941): how much life to pay, from nothing to the seat's life total",
  }),
  payLife: dynamicLabel({
    stem: " life",
    match: "suffix",
    shape: "Pay <N> life",
    make: (n: number) => `Pay ${n} life`,
    example: [4],
    kind: "text",
    role: "button",
    owners: ["lib/choiceDock.ts"],
    doc: "the pay_amount prompt's primary when it asks for life (#1941): pays the stepper's amount",
  }),
  sacrifice: label({
    name: "Sacrifice",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ChoicePromptModal.svelte`],
    doc: "a sacrifice choice's confirm in the dock (sacrifice_choice, entry_sacrifice, an edict's choose_cards): enabled once the permanents are picked, in the sheet or on the board (#2880)",
  }),
  numberToChoose: label({
    name: "number to choose",
    kind: "aria",
    role: "group",
    owners: [`${BOARD}ChoicePromptModal.svelte`],
    doc: "the pay_amount prompt's field when it asks for a number that is not paid (#1941): Volcano Hellion's amount of damage",
  }),
  chooseNumber: dynamicLabel({
    stem: "Choose ",
    match: "prefix",
    shape: "Choose <N>",
    make: (n: number) => `Choose ${n}`,
    example: [3],
    kind: "text",
    role: "button",
    owners: ["lib/choiceDock.ts"],
    doc: "the pay_amount prompt's primary when it asks for a number that is not paid (#1941): answers the field's number",
  }),
  payLifeForMana: label({
    name: "Pay life for {B}…",
    kind: "text",
    role: "menuitem",
    owners: ["lib/payLifeForMana.ts"],
    doc: "the card menu's row that opens the life stepper for the {B} in a cost, under K'rrik (ADR 0131 §4)",
  }),

  // -- The opening roll (ADR 0121) --
  rollForFirstTurn: label({
    name: "roll for the first turn",
    kind: "aria",
    role: "dialog",
    owners: [`${BOARD}OpeningRollDock.svelte`],
    doc: "the opening roll's request; the tutorial's roll step anchors here while the viewer owes a die",
  }),
  roll: label({
    name: "Roll",
    kind: "text",
    role: "button",
    owners: [`${BOARD}OpeningRollDock.svelte`],
    doc: "the opening roll's primary",
  }),
  rollForEveryone: label({
    name: "Roll for everyone left",
    kind: "text",
    role: "button",
    owners: [`${BOARD}OpeningRollDock.svelte`],
    doc: "the host's secondary on the opening roll",
  }),
  chooseFirstTurn: label({
    name: "choose who takes the first turn",
    kind: "aria",
    role: "dialog",
    owners: [`${BOARD}OpeningRollDock.svelte`],
    doc: "the roll winner's sheet; the tutorial's roll step detours here",
  }),
  goesFirst: dynamicLabel({
    stem: " goes first",
    match: "suffix",
    shape: "<name> goes first",
    make: (name: string) => `${name} goes first`,
    example: ["Practice Bot"],
    kind: "text",
    role: "button",
    owners: ["lib/openingRoll.ts"],
    doc: "the chooser's button for another seat",
  }),
  iGoFirst: label({
    name: "I go first",
    kind: "text",
    role: "button",
    owners: ["lib/openingRoll.ts"],
    doc: "the chooser's button for their own seat",
  }),
  giveFirstTurn: dynamicLabel({
    stem: " take the first turn?",
    match: "suffix",
    shape: "Let <name> take the first turn?",
    make: (name: string) => `Let ${name} take the first turn?`,
    example: ["Practice Bot"],
    kind: "aria",
    role: "dialog",
    owners: ["lib/openingRoll.ts"],
    doc: "the confirm when the chooser gives the first turn away, with Confirm and Cancel; in the tutorial's roll detour",
  }),
  confirm: label({
    name: "Confirm",
    kind: "text",
    role: "button",
    owners: [`${BOARD}OpeningRollDock.svelte`],
    doc: "the first-turn confirm's primary",
  }),
  openingRoll: label({
    name: "opening roll",
    kind: "aria",
    role: "group",
    owners: [`${BOARD}OpeningRollBanner.svelte`],
    doc: "the strip's banner during the roll; stands in for opening hand decisions until the winner chooses; the tutorial's roll step anchors here while the table waits on the bot",
  }),

  // -- The ⋯ menu --
  moreActions: label({
    name: "more actions",
    kind: "aria",
    role: "button",
    owners: [`${BOARD}GameMenu.svelte`],
    doc: "the ⋯ button that opens the game menu",
  }),
  gameActions: label({
    name: "game actions",
    kind: "aria",
    role: "menu",
    owners: [`${BOARD}GameMenu.svelte`],
    doc: "the ⋯ menu itself",
  }),
  dice: label({
    name: "Dice",
    kind: "text",
    role: "none",
    owners: [`${BOARD}GameMenu.svelte`],
    doc: "the ⋯ menu's heading over the table rolls; plain text, not a group, with no name of its own",
  }),
  sortHand: label({
    name: "Sort hand",
    kind: "text",
    role: "none",
    owners: [`${BOARD}GameMenu.svelte`],
    doc: "the ⋯ menu's heading over the one-time hand sorts (#3028), and the sortHand shortcut's row in the keymap",
  }),
  rollD6: label({
    name: "Roll a d6",
    kind: "text",
    role: "menuitem",
    owners: ["lib/gameMenu.ts"],
    doc: "a table roll (ADR 0121 §5); disabled for 2 s after the viewer's own roll",
  }),
  rollD20: label({
    name: "Roll a d20",
    kind: "text",
    role: "menuitem",
    owners: ["lib/gameMenu.ts"],
    doc: "a table roll (ADR 0121 §5)",
  }),
  flipCoin: label({
    name: "Flip a coin",
    kind: "text",
    role: "menuitem",
    owners: ["lib/gameMenu.ts"],
    doc: "a table roll (ADR 0121 §5)",
  }),

  rulings: label({
    name: "Rulings",
    kind: "text",
    role: "menuitem",
    owners: [`${BOARD}CardContextMenu.svelte`, `${BOARD}ManaAbilityMenu.svelte`],
    doc: "a card menu's link to the card's official rulings in a new tab (#3831); not offered for a hidden or face-down card or a token",
  }),

  // -- The admin views (ADR 0124 §7, §8) --
  adminViews: label({
    name: "admin views",
    kind: "aria",
    role: "navigation",
    owners: ["routes/Admin.svelte"],
    doc: "the admin views' tabs: Live now, Games, Accounts",
  }),
  adminLiveNowTab: label({
    name: "Live now",
    kind: "text",
    role: "link",
    owners: ["routes/Admin.svelte"],
    doc: "the admin views' first tab",
  }),
  adminGamesTab: label({
    name: "Games",
    kind: "text",
    role: "link",
    owners: ["routes/Admin.svelte"],
    doc: "the admin views' tab for every table",
  }),
  adminAccountsTab: label({
    name: "Accounts",
    kind: "text",
    role: "link",
    owners: ["routes/Admin.svelte"],
    doc: "the admin views' tab for every account",
  }),
  adminReportsTab: label({
    name: "Reports",
    kind: "text",
    role: "link",
    owners: ["routes/Admin.svelte"],
    doc: "the admin views' tab for what players have reported at their tables (#2984)",
  }),
  adminStatsTab: label({
    name: "Stats",
    kind: "text",
    role: "link",
    owners: ["routes/Admin.svelte"],
    doc: "the admin views' tab for game statistics (ADR 0165 §8, #3425)",
  }),
  liveNow: label({
    name: "live now",
    kind: "aria",
    role: "region",
    owners: ["lib/components/admin/LiveNow.svelte"],
    doc: "who is on now and the tables in play",
  }),
  gamesTable: label({
    name: "games",
    kind: "aria",
    role: "table",
    owners: ["lib/components/admin/GamesList.svelte"],
    doc: "the admin view's list of tables",
  }),
  accountsTable: label({
    name: "accounts",
    kind: "aria",
    role: "table",
    owners: ["lib/components/admin/AccountsList.svelte"],
    doc: "the admin view's list of accounts",
  }),

  // -- The site pages (ADR 0125 §3.7): the first-use hints' anchors --
  createGame: label({
    name: "create game",
    kind: "aria",
    role: "form",
    owners: ["routes/Lobby.svelte"],
    doc: "the Tables page's new-table form, opened by + New table (ADR 0158 §1 item 1)",
  }),
  newTable: label({
    name: "new table",
    kind: "aria",
    role: "button",
    owners: ["routes/Lobby.svelte"],
    doc: "the Tables page head's + New table, which opens the new-table form; the lobby.create hint points at it",
  }),
  suggestTableName: label({
    name: "suggest a name",
    kind: "aria",
    role: "button",
    owners: ["lib/components/NewTableRow.svelte"],
    doc: "the dice next to the Lobby's game-name box: fills it with a generated table name (#2630)",
  }),
  deckLink: label({
    name: "deck link",
    kind: "aria",
    role: "textbox",
    owners: ["routes/Decks.svelte"],
    doc: "the Decks page's deck-link field; the decks.check hint points at it",
  }),
  choosePrintings: label({
    name: "Choose printings",
    kind: "text",
    role: "button",
    owners: ["lib/components/decks/DeckCard.svelte"],
    doc: "saved-deck printing picker",
  }),
  printingPicker: label({
    name: "card printings",
    kind: "aria",
    role: "region",
    owners: ["lib/components/decks/PrintingPicker.svelte"],
    doc: "saved-deck printing picker",
  }),
  printingCard: label({
    name: "Card to customize",
    kind: "aria",
    role: "combobox",
    owners: ["lib/components/decks/PrintingPicker.svelte"],
    doc: "saved-deck printing picker",
  }),
  filterPrintings: label({
    name: "Filter printings",
    kind: "aria",
    role: "searchbox",
    owners: ["lib/components/decks/PrintingPicker.svelte"],
    doc: "saved-deck printing picker",
  }),
  savePrintings: label({
    name: "Save printings",
    kind: "text",
    role: "button",
    owners: ["lib/components/decks/PrintingPicker.svelte"],
    doc: "saved-deck printing picker",
  }),
  cancelPrintings: label({
    name: "Cancel printing changes",
    kind: "text",
    role: "button",
    owners: ["lib/components/decks/PrintingPicker.svelte"],
    doc: "saved-deck printing picker",
  }),
  updateDeck: label({
    name: "Update deck",
    kind: "text",
    role: "button",
    owners: ["lib/components/decks/DeckCard.svelte"],
    doc: "a saved deck's action that opens the update sheet (#2036)",
  }),
  deckUpdateSheet: label({
    name: "update deck",
    kind: "aria",
    role: "region",
    owners: ["lib/components/decks/UpdateDeckSheet.svelte"],
    doc: "the sheet that updates a saved deck from its link, a new link or a pasted list (#2036)",
  }),
  deckUpdateChoices: label({
    name: "how to update",
    kind: "aria",
    role: "group",
    owners: ["lib/components/decks/UpdateDeckSheet.svelte"],
    doc: "the update sheet's three ways in: re-pull, new link, paste list (#2036)",
  }),
  previewUpdate: label({
    name: "Preview changes",
    kind: "text",
    role: "button",
    owners: ["lib/components/decks/UpdateDeckSheet.svelte"],
    doc: "the update sheet's first step: read the new list and show the diff, writing nothing (#2036)",
  }),
  saveUpdate: label({
    name: "Save update",
    kind: "text",
    role: "button",
    owners: ["lib/components/decks/UpdateDeckSheet.svelte"],
    doc: "the update sheet's confirm: commits the previewed list (#2036)",
  }),
  cancelUpdate: label({
    name: "Cancel update",
    kind: "text",
    role: "button",
    owners: ["lib/components/decks/UpdateDeckSheet.svelte"],
    doc: "closes the update sheet without saving (#2036)",
  }),
  yourDecks: label({
    name: "your decks",
    kind: "aria",
    role: "list",
    owners: ["routes/Decks.svelte"],
    doc: "a signed-in person's saved decks; the decks.library hint points at it",
  }),
  settingsSections: label({
    name: "settings sections",
    kind: "aria",
    role: "navigation",
    owners: ["lib/components/Settings.svelte"],
    doc: "the Settings dialog's section tabs; the settings.display hint points at it",
  }),
  searchCatalogue: label({
    name: "search the catalogue",
    kind: "aria",
    role: "textbox",
    owners: ["routes/Catalog.svelte"],
    doc: "the Catalogue's search field; the catalog.search hint points at it",
  }),
  searchRoadmap: label({
    name: "search the roadmap",
    kind: "aria",
    role: "textbox",
    owners: ["routes/Roadmap.svelte"],
    doc: "the Roadmap's search field; the roadmap.search hint points at it",
  }),

  lobbyTitle: label({
    name: "cmd_and_ctrl · lobby",
    kind: "aria",
    role: "heading",
    owners: ["routes/Lobby.svelte"],
    doc: "the Lobby's title (its visible text is Tables); the lobby.practice hint anchors here",
  }),

  // -- A table card's toolbar and menu (#3499) --
  copyInvite: label({
    name: "Copy invite",
    kind: "text",
    role: "button",
    owners: ["routes/Lobby.svelte"],
    doc: "the wide part of a lobby table's invite split button: copies the player invite link; the lobby.create hint names it",
  }),
  newInviteLink: label({
    name: "new invite link",
    kind: "aria",
    role: "button",
    owners: ["routes/Lobby.svelte"],
    doc: "the ↻ part of the invite split button: mints a new player invite after a confirm, then copies it",
  }),
  copySpectatorLink: label({
    name: "Copy spectator link",
    kind: "text",
    role: "button",
    owners: ["routes/Lobby.svelte"],
    doc: "the wide part of a table's spectator split button: copies the read-only link",
  }),
  newSpectatorLink: label({
    name: "new spectator link",
    kind: "aria",
    role: "button",
    owners: ["routes/Lobby.svelte"],
    doc: "the ↻ part of the spectator split button: mints a new spectator link after a confirm, then copies it",
  }),
  tableSettingsLine: dynamicLabel({
    stem: "table settings: ",
    match: "prefix",
    shape: "table settings: <summary>",
    make: (summary: string) => `table settings: ${summary}`,
    example: ["Standard rules"],
    kind: "aria",
    role: "button",
    owners: ["routes/Lobby.svelte"],
    doc: "a table card's rules line, e.g. 30 life · no undos: opens the table settings panel (read-only for a non-host)",
  }),
  tableMenu: dynamicLabel({
    stem: "actions for ",
    match: "prefix",
    shape: "actions for <table name>",
    make: (table: string) => `actions for ${table}`,
    example: ["Friday Pod"],
    kind: "aria",
    role: "menu",
    owners: ["routes/Lobby.svelte"],
    doc: "a table card's right-click menu (also the Menu key, Shift+F10 or a long press)",
  }),

  reportABug: label({
    name: "Report a bug or idea",
    kind: "aria",
    role: "button",
    owners: ["routes/Game.svelte", "lib/components/SiteHeader.svelte"],
    doc: "the bug button: the table's command bar, and the site header on every other page for someone signed in with Discord (#3492); the Guide and About name it",
  }),

  // -- The site header (ADR 0156 §3) --
  navTables: label({
    name: "Tables",
    kind: "text",
    role: "link",
    owners: ["lib/components/SiteHeader.svelte"],
    doc: "the site header's link to #/lobby, first in the signed-in nav (it read Lobby before ADR 0156)",
  }),

  siteMenu: label({
    name: "Site menu",
    kind: "aria",
    role: "button",
    owners: ["lib/components/SiteHeader.svelte"],
    doc: "the site header's menu button below 600px, which opens the nav links (a disclosure: aria-expanded, Escape closes it); hidden at tablet width and up (#3877)",
  }),

  // -- The front door (ADR 0156 §4–6) --
  signInWithDiscord: label({
    name: "Sign in with Discord",
    kind: "text",
    role: "link",
    owners: ["routes/FrontPage.svelte"],
    doc: "the front page's Discord sign-in, shown when the server has Discord sign-in configured",
  }),
  joinDiscordServer: label({
    name: "Join our Discord server",
    kind: "text",
    role: "link",
    owners: ["routes/FrontPage.svelte"],
    doc: "the front page's link to the C2 Discord invite; hidden when GET /community names none",
  }),
  tryPracticeGame: label({
    name: "Try a practice game",
    kind: "text",
    role: "link",
    owners: ["routes/FrontPage.svelte", "routes/RequestAccess.svelte"],
    doc: "the front page's and the gate page's way in without an account: #/practice, a guest practice game when signed out",
  }),
  findAPod: label({
    name: "Find a pod",
    kind: "text",
    role: "link",
    owners: ["routes/Lobby.svelte"],
    doc: "the Lobby's empty state: the C2 Discord invite, in the No pod yet? card; hidden when GET /community names none",
  }),

  // -- Help (ADR 0125 §6) --
  help: label({
    name: "help",
    kind: "aria",
    role: "button",
    owners: ["lib/components/HelpMenu.svelte"],
    doc: "the site header's Help button; the menu it opens is named by it (aria-labelledby), and the site.help hint anchors here",
  }),
  tipsForThisPage: label({
    name: "Tips for this page",
    kind: "text",
    role: "menuitem",
    owners: ["lib/components/HelpMenu.svelte"],
    doc: "Help: shows this page's tips again, one after another, even with tips off",
  }),
  tipsForTheTable: label({
    name: "Tips for the table",
    kind: "text",
    role: "menuitem",
    owners: [`${BOARD}GameMenu.svelte`],
    doc: "the ⋯ menu's Help group: shows the table's tips again at the next quiet moments",
  }),
  showAllTipsAgain: label({
    name: "Show all tips again",
    kind: "text",
    role: "menuitem",
    owners: ["lib/components/HelpMenu.svelte", "lib/components/Settings.svelte"],
    doc: "Help, and Settings → Advanced: forgets every dismissed tip and turns tips back on",
  }),
  practiceGame: label({
    name: "Practice game",
    kind: "text",
    role: "menuitem",
    owners: ["lib/components/HelpMenu.svelte", "routes/Lobby.svelte"],
    doc: "Help, and the Lobby's empty-state card: opens a practice table, a guest one when signed out (ADR 0156 §2)",
  }),
  gettingStarted: label({
    name: "Getting started",
    kind: "text",
    role: "menuitem",
    owners: ["lib/components/HelpMenu.svelte"],
    doc: "Help: opens the getting-started guide (#/guide)",
  }),
  keyboardShortcutsItem: label({
    name: "Keyboard shortcuts",
    kind: "text",
    role: "menuitem",
    owners: ["lib/components/HelpMenu.svelte", `${BOARD}GameMenu.svelte`],
    doc: "Help, and the ⋯ menu's Help group: opens the keymap overlay",
  }),
  replayTutorial: label({
    name: "Replay the tutorial",
    kind: "text",
    role: "menuitem",
    owners: [`${BOARD}GameMenu.svelte`, "lib/components/Settings.svelte"],
    doc: "Settings → Advanced, and the ⋯ menu on the practice table only: opens a fresh practice table",
  }),
  legalNotice: label({
    name: "Legal notice",
    kind: "text",
    role: "menuitem",
    owners: [`${BOARD}GameMenu.svelte`],
    doc: "the ⋯ menu's Help group: opens the fan-content notice, which the site footer carries everywhere else (#2982)",
  }),
  showTips: label({
    name: "Show tips",
    kind: "text",
    role: "checkbox",
    owners: ["lib/components/Settings.svelte"],
    doc: "Settings → Advanced: on unless Hide tips was pressed on a tip",
  }),
  stackArrivalToggle: label({
    name: "Stack: trace each spell or ability in from its source",
    kind: "text",
    role: "checkbox",
    owners: ["lib/components/Settings.svelte"],
    doc: "Settings → Animations: the arrival trace, animations.stackArrival (ADR 0119 amendment 2026-10-16 §D)",
  }),
  effectsSetting: label({
    name: "Table effects",
    kind: "text",
    role: "combobox",
    owners: ["lib/components/Settings.svelte"],
    doc: "Settings → Animations: Full, Short or Off for the generic table effects, animations.effects (ADR 0209 §1, #4396)",
  }),
  effectBadgeDestroyed: label({
    name: "Destroyed",
    kind: "text",
    role: "status",
    owners: [`${BOARD}TableEffects.svelte`],
    doc: "the still badge a destroyed permanent leaves where it was, when effect motion is off or reduced (ADR 0209 §1)",
  }),
  effectBadgeDied: label({
    name: "Died",
    kind: "text",
    role: "status",
    owners: [`${BOARD}TableEffects.svelte`],
    doc: "the still badge for a permanent that died to a state-based action (lethal damage, zero toughness)",
  }),
  effectBadgeExiled: label({
    name: "Exiled",
    kind: "text",
    role: "status",
    owners: [`${BOARD}TableEffects.svelte`],
    doc: "the still badge for a permanent that was exiled, so an exile is never read as a death",
  }),
  effectBadgeSacrificed: label({
    name: "Sacrificed",
    kind: "text",
    role: "status",
    owners: [`${BOARD}TableEffects.svelte`],
    doc: "the still badge for a sacrificed permanent",
  }),
  effectBadgeBounced: label({
    name: "Returned to hand",
    kind: "text",
    role: "status",
    owners: [`${BOARD}TableEffects.svelte`],
    doc: "the still badge for a permanent that went back to a hand",
  }),
  effectBadgePwEnter: label({
    name: "Planeswalker arrived",
    kind: "text",
    role: "status",
    owners: [`${BOARD}TableEffects.svelte`],
    doc: "the still badge for a planeswalker entering the battlefield",
  }),
  effectBadgePwLeave: label({
    name: "Planeswalker gone",
    kind: "text",
    role: "status",
    owners: [`${BOARD}TableEffects.svelte`],
    doc: "the still badge for a planeswalker that died",
  }),
  effectBadgeWipe: dynamicLabel({
    stem: "Board wipe: ",
    match: "prefix",
    shape: "Board wipe: <N> permanent(s)",
    make: (n: number) => `Board wipe: ${plural(n, "permanent")}`,
    example: [5],
    kind: "text",
    role: "status",
    owners: [`${BOARD}TableEffects.svelte`],
    doc: "the one still badge a board wipe leaves over the table (three or more permanents leaving in one batch from one source)",
  }),
  startPractice: label({
    name: "Start practice",
    kind: "text",
    role: "link",
    owners: ["routes/LobbyPractice.hint.ts", "routes/Lobby.svelte"],
    doc: "the lobby.practice tip's action, to #/practice, and the Lobby's Practice game card; the tip's Got it reads Not now beside it",
  }),

  // -- Everywhere --
  tip: label({
    name: "tip",
    kind: "aria",
    role: "complementary",
    owners: ["lib/components/hints/HintCard.svelte"],
    doc: "a first-use hint's card (ADR 0125 §3.6); one at a time, and Go to the tip (i) moves focus to it",
  }),
  keyboardShortcuts: label({
    name: "keyboard shortcuts",
    kind: "text",
    role: "dialog",
    owners: ["lib/components/ShortcutsOverlay.svelte"],
    doc: "the keymap overlay (?), named by its heading",
  }),

  // -- Automatic answers (ADR 0127) --
  rememberThisAnswer: label({
    name: "Remember this answer",
    kind: "text",
    role: "button",
    owners: ["lib/choiceDock.ts"],
    doc: "the toggle under a prompt that can take a standing answer; on, the next button pressed also sets the rule (Yes or Pay: Always; No or Don't pay: Never)",
  }),
  recommendedAnswer: label({
    name: "Recommended",
    kind: "text",
    role: "generic",
    owners: ["lib/choiceDock.ts"],
    doc: 'the tag after the recommended answer of the which-skip question, "Use my skip" (ADR 0188 Decision 12); Enter takes it',
  }),
  choosingNextTurn: label({
    name: "Choosing the next turn",
    kind: "text",
    role: "generic",
    owners: ["lib/components/board/PhaseDisplay.svelte"],
    doc: "the step label while the turn change waits on the which-skip question (turn.between_turns, ADR 0188 Decision 12): the old turn is over and the next has not begun",
  }),
  automaticAnswer: label({
    name: "automatic answer",
    kind: "aria",
    role: "dialog",
    owners: ["lib/autoAnswerPref.ts"],
    doc: 'the dock\'s notice after the server answered a prompt for you ("Rhystic Study: paid {1} for you"), for about six seconds',
  }),
  undoAutomaticAnswer: label({
    name: "Undo",
    kind: "text",
    role: "button",
    owners: ["lib/autoAnswerPref.ts"],
    doc: "the automatic-answer notice's undo; greyed once anything else is the top undo entry",
  }),
  autoPassNotice: label({
    name: "auto-pass notice",
    kind: "aria",
    role: "dialog",
    owners: ["lib/passReason.ts"],
    doc: 'the dock\'s chip after the client passed priority for you ("Passed for you: nothing to respond with", #3343), for about six seconds',
  }),
  undoAutoPass: label({
    name: "Take back the pass",
    kind: "aria",
    role: "button",
    owners: ["lib/passReason.ts"],
    doc: 'the auto-pass chip\'s Undo (visible text "Undo"); sends the ordinary undo and is shown only while that undo would take the pass back. Named apart from the dock\'s "Undo (N left)" so a /^Undo/ locator still finds one button',
  }),
  askMeNextTime: label({
    name: "Ask me next time",
    kind: "text",
    role: "button",
    owners: ["lib/autoAnswerPref.ts"],
    doc: "the automatic-answer notice's button that removes the rule (sets Ask); the answer just given stands",
  }),
  automaticAnswers: label({
    name: "automatic answers",
    kind: "aria",
    role: "region",
    owners: ["lib/components/AutoAnswersSettings.svelte"],
    doc: "Settings → Gameplay's list of standing answers; the settings.auto-answers hint points at it",
  }),
  forgetAutoAnswer: label({
    name: "Forget",
    kind: "text",
    role: "button",
    owners: ["lib/components/AutoAnswersSettings.svelte"],
    doc: "removes one standing answer in Settings → Gameplay, so its prompt is asked again",
  }),
  useThisOrderAgain: label({
    name: "Use this order again",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ChoicePromptModal.svelte`],
    doc: "the toggle under the trigger-order list (#3016, ADR 0018's eighth amendment); on, the next batch of the same abilities is ordered the same way for you. Disabled on an order that splits one card's triggers apart",
  }),
  searchLibraryAndGraveyard: label({
    name: "Library and graveyard",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ChoicePromptModal.svelte`],
    doc: 'the primary of the question a "search your library and/or graveyard" prompt asks before it shows anything (#4519, CR 701.23): search both zones. The library is shuffled afterwards',
  }),
  searchLibraryAndHand: label({
    name: "Library and hand",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ChoicePromptModal.svelte`],
    doc: 'the primary of the question a "search your hand and/or library" prompt asks before it shows anything (#4519, CR 701.23): search both zones. The library is shuffled afterwards',
  }),
  searchLibraryOnly: label({
    name: "Library only",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ChoicePromptModal.svelte`],
    doc: "an and/or search's zone question (#4519): search the library alone; it is shuffled afterwards",
  }),
  searchGraveyardOnly: label({
    name: "Graveyard only",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ChoicePromptModal.svelte`],
    doc: 'a "search your library and/or graveyard" prompt zone question (#4519): search the graveyard alone; the library is never shown or shuffled',
  }),
  searchHandOnly: label({
    name: "Hand only",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ChoicePromptModal.svelte`],
    doc: 'a "search your hand and/or library" prompt zone question (#4519): search the hand alone; the library is never shown or shuffled',
  }),
  forgetTriggerOrders: dynamicLabel({
    stem: "Forget remembered trigger orders",
    match: "prefix",
    shape: "Forget remembered trigger orders (<N>)",
    make: (n: number) => `Forget remembered trigger orders (${n})`,
    example: [2],
    kind: "text",
    role: "menuitem",
    owners: [`${BOARD}GameMenu.svelte`],
    doc: "the ⋯ menu's entry that forgets every trigger order you asked to reuse this game (#3016); shown while you have one",
  }),

  // -- Settings → Audio: a volume for each sound category (ADR 0209 §2, #4394) --
  volumeMaster: label({
    name: "Master volume",
    kind: "text",
    role: "generic",
    owners: ["lib/components/Settings.svelte"],
    doc: "the slider in Settings → Audio that sits over every category and the music",
  }),
  volumeInterface: label({
    name: "Interface volume",
    kind: "text",
    role: "generic",
    owners: ["lib/components/Settings.svelte"],
    doc: "the slider for your-turn, priority, prompt and timer cues (ADR 0209 §2); an interface cue ducks the other buses while it plays",
  }),
  volumeCards: label({
    name: "Cards volume",
    kind: "text",
    role: "generic",
    owners: ["lib/components/Settings.svelte"],
    doc: "the slider for draw, play, tap, untap and shuffle sounds (what Effects volume was before settings v29)",
  }),
  volumeCombat: label({
    name: "Combat volume",
    kind: "text",
    role: "generic",
    owners: ["lib/components/Settings.svelte"],
    doc: "the slider for attack, block and combat damage sounds",
  }),
  volumeEffects: label({
    name: "Effects volume",
    kind: "text",
    role: "generic",
    owners: ["lib/components/Settings.svelte"],
    doc: "the slider for the generic effect cues and signature card effects (ADR 0209 §3, §4)",
  }),
  volumeVoice: label({
    name: "Voice volume",
    kind: "text",
    role: "generic",
    owners: ["lib/components/Settings.svelte"],
    doc: "the slider for planeswalker voice lines (ADR 0209 §5); none ship yet",
  }),
  volumeMusic: label({
    name: "Music volume",
    kind: "text",
    role: "generic",
    owners: ["lib/components/Settings.svelte"],
    doc: "the slider for the ambient music bed",
  }),

  // -- Combat damage assignment (#2956, ADR 0147) --
  damageAutoAssigned: label({
    name: "combat damage assigned",
    kind: "aria",
    role: "dialog",
    owners: ["lib/damageAssignment.ts"],
    doc: "the dock's notice after Auto-assign combat damage assigned an attacker's damage for you (lethal to each blocker, the rest over), for about six seconds",
  }),
  alwaysAskCombatDamage: label({
    name: "Always ask",
    kind: "text",
    role: "button",
    owners: ["lib/damageAssignment.ts"],
    doc: "the combat-damage notice's button that turns Auto-assign combat damage off; the assignment just made stands",
  }),
  addDamage: dynamicLabel({
    stem: "Add 1 damage to ",
    match: "prefix",
    shape: "Add 1 damage to <name>",
    make: (name: string) => `Add 1 damage to ${name}`,
    example: ["Corsair Captain"],
    kind: "aria",
    role: "button",
    owners: [`${BOARD}DamageStepper.svelte`, `${BOARD}ChoicePromptModal.svelte`],
    doc: "the + on a blocker (on the board and in the damage sheet) while you assign combat damage; the defending player's trample line uses it too",
  }),
  removeDamage: dynamicLabel({
    stem: "Remove 1 damage from ",
    match: "prefix",
    shape: "Remove 1 damage from <name>",
    make: (name: string) => `Remove 1 damage from ${name}`,
    example: ["Corsair Captain"],
    kind: "aria",
    role: "button",
    owners: [`${BOARD}DamageStepper.svelte`, `${BOARD}ChoicePromptModal.svelte`],
    doc: "the − beside addDamage",
  }),
  awayOffer: label({
    name: "move on without an away player",
    kind: "aria",
    role: "dialog",
    owners: ["lib/away.ts"],
    doc: "the dock sheet a decider sees when a seat the table waits on has been offline for the away time (ADR 0177 §2)",
  }),
  awayStandIn: dynamicLabel({
    stem: "'s seat to a bot",
    match: "suffix",
    shape: "Hand <name>'s seat to a bot",
    make: (name: string) => `Hand ${name}'s seat to a bot`,
    example: ["Bo"],
    kind: "text",
    role: "button",
    owners: ["lib/away.ts"],
    doc: "the away offer's hand-over: a heuristic bot plays the seat until its player takes it back",
  }),
  awayConcede: dynamicLabel({
    stem: "Concede for ",
    match: "prefix",
    shape: "Concede for <name>",
    make: (name: string) => `Concede for ${name}`,
    example: ["Bo"],
    kind: "text",
    role: "button",
    owners: ["lib/away.ts"],
    doc: "the away offer's concession, made for the away player",
  }),
  awayWait: label({
    name: "Keep waiting",
    kind: "text",
    role: "button",
    owners: ["lib/away.ts"],
    doc: "the away offer's third answer: hide the offer for another away time",
  }),
  takeBackSeat: label({
    name: "Take it back",
    kind: "text",
    role: "button",
    owners: [`${BOARD}StandInBanner.svelte`],
    doc: "the banner a player sees on their own seat while a bot plays it: they play it again (ADR 0177 §6)",
  }),
  holdShiftToFlip: label({
    name: "Hold Shift to flip",
    kind: "text",
    role: "note",
    owners: [`${BOARD}HoverZoomOverlay.svelte`],
    doc: "the hover zoom's hint under a two-faced card's inset, on pointer devices only (#4106): holding Shift shows the other face big",
  }),
} satisfies Record<string, LabelSpec>;

export type LabelKey = keyof typeof LABEL_SPECS;

type LabelOf<S> =
  S extends StaticSpec<infer K>
    ? Name<K>
    : S extends DynamicSpec<infer K, infer A>
      ? DynamicLabel<K, A>
      : never;

function build(): { readonly [P in LabelKey]: LabelOf<(typeof LABEL_SPECS)[P]> } {
  const out: Record<string, unknown> = {};
  for (const [key, spec] of Object.entries(LABEL_SPECS) as [string, LabelSpec][]) {
    if (spec.type === "static") {
      out[key] = spec.name;
      continue;
    }
    const make = spec.make as (...a: unknown[]) => string;
    const fn = (...a: unknown[]): string => make(...a);
    if (spec.kind === "aria") {
      const any = { key, stem: spec.stem, match: spec.match } as unknown as Stem;
      Object.defineProperty(fn, "any", { value: Object.freeze(any), enumerable: true });
    }
    out[key] = Object.freeze(fn);
  }
  return Object.freeze(out) as { readonly [P in LabelKey]: LabelOf<(typeof LABEL_SPECS)[P]> };
}

/** The contract labels. Render `L.<key>`, never the literal. */
export const L = build();

// ---- Checks, for the anchor resolver and the guard ----

/** isStem tells a stem anchor from an exact name. */
export function isStem(ref: LabelRef): ref is Stem {
  return typeof ref !== "string";
}

/** stemMatches reports whether a rendered name has a dynamic entry's stem where it says. */
export function stemMatches(name: string, stem: string, match: StemMatch): boolean {
  switch (match) {
    case "prefix":
      return name.startsWith(stem);
    case "suffix":
      return name.endsWith(stem);
    case "contains":
      return name.includes(stem);
  }
}

/**
 * registeredAria names the entry an anchor's label belongs to, or null
 * when it belongs to none: an exact static `aria` name, a name a dynamic
 * `aria` entry renders, or one of their stems. A static name is checked
 * first, so "your board" is yourBoard, not a seat board.
 */
export function registeredAria(ref: LabelRef | string): LabelKey | null {
  const entries = Object.entries(LABEL_SPECS) as [LabelKey, LabelSpec][];
  if (typeof ref !== "string") {
    const spec = LABEL_SPECS[ref.key as LabelKey] as LabelSpec | undefined;
    return spec && spec.type === "dynamic" && spec.kind === "aria" && spec.stem === ref.stem
      ? (ref.key as LabelKey)
      : null;
  }
  for (const [key, spec] of entries) {
    if (spec.type === "static" && spec.kind === "aria" && spec.name === ref) return key;
  }
  for (const [key, spec] of entries) {
    if (spec.type === "dynamic" && spec.kind === "aria" && stemMatches(ref, spec.stem, spec.match))
      return key;
  }
  return null;
}
