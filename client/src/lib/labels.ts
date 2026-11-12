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
const DRAFT = "lib/components/draft/";

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
  tombstoneRip: label({
    name: "R. I. P.",
    kind: "text",
    role: "generic",
    owners: [`${BOARD}Tombstone.svelte`],
    doc: "#4449: the epitaph on the tombstone that stays on an eliminated player's panel",
  }),
  tombstoneDecked: label({
    name: "Decked",
    kind: "text",
    role: "generic",
    owners: [`${BOARD}Tombstone.svelte`],
    doc: "#4879: the epitaph on the stone of a player who lost by drawing from an empty library (CR 704.5b), in place of the R. I. P.",
  }),
  tombstoneOf: dynamicLabel({
    stem: " is out of the game",
    match: "suffix",
    shape: "<name> is out of the game",
    make: (name: string) => `${name} is out of the game`,
    example: ["Practice Bot"],
    kind: "aria",
    role: "group",
    owners: [`${BOARD}Tombstone.svelte`],
    doc: "#4449: the tombstone on an eliminated player's panel; its text is the player's name, their commander and the epitaph",
  }),
  finalBoard: label({
    name: "Final board",
    kind: "text",
    role: "button",
    owners: [`${BOARD}Tombstone.svelte`, `${BOARD}FinalBoardModal.svelte`],
    doc: "#4912: the button on an eliminated player's tombstone that opens their final board, what they had when they left the game; also the final board dialog's title",
  }),
  finalBoardOpen: dynamicLabel({
    stem: "'s final board",
    match: "suffix",
    shape: "show <name>'s final board",
    make: (name: string) => `show ${name}'s final board`,
    example: ["Practice Bot"],
    kind: "aria",
    role: "button",
    owners: [`${BOARD}Tombstone.svelte`],
    doc: "#4912: the tombstone's Final board button, named for the player whose board it opens",
  }),
  finalBoardOf: dynamicLabel({
    stem: "final board of ",
    match: "prefix",
    shape: "final board of <name>",
    make: (name: string) => `final board of ${name}`,
    example: ["Practice Bot"],
    kind: "aria",
    role: "dialog",
    owners: [`${BOARD}FinalBoardModal.svelte`],
    doc: "#4912: the read-only dialog of an eliminated player's final board: battlefield, hand, graveyard, exile and command zone as they left the game, never the library",
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
  rollPlanarDie: label({
    name: "Roll planar die",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's planar die button at a Planechase table (ADR 0216 §10): the CR 901.9 special action, its cost beside it; disabled, not hidden, outside the active player's main phase",
  }),
  planarZone: label({
    name: "planes",
    kind: "aria",
    role: "group",
    owners: [`${BOARD}PlanarZone.svelte`],
    doc: "the face-up plane at a Planechase table (ADR 0216 §10), drawn landscape at the table's left edge; the planar-zone hint points at it",
  }),
  planarDeck: dynamicLabel({
    stem: "planar deck",
    match: "prefix",
    shape: "planar deck, <n> cards",
    make: (n: number) => `planar deck, ${n} ${n === 1 ? "card" : "cards"}`,
    example: [74],
    kind: "aria",
    role: "status",
    owners: [`${BOARD}PlanarZone.svelte`],
    doc: "the communal planar deck's count beside the face-up plane (ADR 0216 §10)",
  }),
  planechaseSetting: label({
    name: "Planechase",
    kind: "text",
    role: "checkbox",
    owners: ["lib/components/TableSettingsPanel.svelte", "routes/Lobby.svelte"],
    doc: "the Planechase table option (ADR 0216 §1): the new-table form's box and the table settings switch, locked once the game starts; offered where the server offers it",
  }),
  tableFormat: label({
    name: "Format",
    kind: "aria",
    role: "radiogroup",
    owners: ["routes/Lobby.svelte"],
    doc: "the new-table form's format picker (ADR 0222 §1): one radio per format, Commander first and chosen by default; the format is fixed once a seat has a deck",
  }),
  formatCommander: label({
    name: "Commander",
    kind: "text",
    role: "radio",
    owners: ["lib/formats.ts"],
    doc: "the Commander format's name (ADR 0222 §1): the format picker's first radio, and a deck's or a table's format wherever it is named",
  }),
  formatStandard: label({
    name: "Standard",
    kind: "text",
    role: "radio",
    owners: ["lib/formats.ts"],
    doc: "the Standard format's name (ADR 0222 §1): the format picker's radio, a Standard table's chip and rules line, and a Standard deck's tag in the pickers",
  }),
  formatLimited: label({
    name: "Limited",
    kind: "text",
    role: "text",
    owners: ["lib/formats.ts"],
    doc: "the limited format's name (ADR 0222 §6, #5440): the format of a table a draft opens for one of its matches; never in the format picker",
  }),
  matchBestOfOne: label({
    name: "Best of one",
    kind: "text",
    role: "radio",
    owners: ["lib/matchLength.ts"],
    doc: "the match length picker's first radio (ADR 0222 §3): one game, the default at a table whose format plays matches",
  }),
  matchBestOfThree: label({
    name: "Best of three",
    kind: "text",
    role: "radio",
    owners: ["lib/matchLength.ts"],
    doc: "the match length picker's second radio (ADR 0222 §3), and a best-of-three table's chip in the table list",
  }),
  matchScore: label({
    name: "match score",
    kind: "aria",
    role: "group",
    owners: [`${BOARD}MatchScore.svelte`],
    doc: "the command bar's score of a best-of-three (ADR 0222 §3): the game number and each seat's games won; absent at a Commander table and a best-of-one",
  }),
  sideboarding: label({
    name: "sideboarding between games",
    kind: "aria",
    role: "dialog",
    owners: [`${BOARD}SideboardingSheet.svelte`],
    doc: "the screen between two games of a match (ADR 0222 §3): your deck and sideboard, the play-or-draw choice, the timer and Ready",
  }),
  sideboardingBanner: label({
    name: "between games of the match",
    kind: "aria",
    role: "group",
    owners: ["routes/Game.svelte"],
    doc: "the strip's banner between two games of a match (ADR 0222 §3): the last result, the score, the timer and who is ready; a spectator sees this and no sideboarding screen",
  }),
  sideboardReady: label({
    name: "Ready for the next game",
    kind: "text",
    role: "button",
    owners: [`${BOARD}SideboardingSheet.svelte`],
    doc: "the sideboarding screen's primary: readies you up; the next game starts when both players are ready or the timer runs out",
  }),
  concedeMatch: label({
    name: "Concede match",
    kind: "text",
    role: "button",
    owners: [`${BOARD}SideboardingSheet.svelte`, `${BOARD}GameMenu.svelte`],
    doc: "concedes the whole match (ADR 0222 §3), from the sideboarding screen and from the game menu's concede confirmation; conceding the game alone is the menu's other button",
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
    doc: "the ⋯ menu entry (ADR 0193 first amendment, CR 732.2a), never hidden (2026-10-30 amendment, #4736): any seat in an active game may start one, partway through a resolution too, and it begins at their next priority with nothing open; every action from then on, any seat's, is one iteration of a loop. Disabled with a one-line reason for a spectator, a game not active, or a loop already standing; once recording, Stop and Cancel are the dock's status line",
  }),
  stopLoop: label({
    name: "Stop recording the loop",
    kind: "aria",
    role: "button",
    owners: [`${BOARD}LoopControl.svelte`],
    doc: "ends the demonstration (ADR 0193): at your priority, the next one if you do not hold it now (#4736), the server checks the board came back, keeps the loop to propose, or discards it and says why in the log",
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

  opensFlashNeededNow: label({
    name: "needed to cast it now",
    kind: "text",
    role: "generic",
    owners: [`${BOARD}AlternativeCostModal.svelte`],
    doc: "#5363: under an optional cost in the cost picker that the chosen cost can be cast with right now only if it is paid (CR 601.3c); the picker ticks it",
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
  facePickerOmen: label({
    name: "omen · CR 720.3",
    kind: "text",
    role: "generic",
    owners: [`${BOARD}FacePickerModal.svelte`],
    doc: "#5191: the face picker's source tag on an omen card (CR 720.3: cast it normally or as an Omen), which Scryfall lays out as an adventure; known by the Omen spell type on its second half's type line",
  }),
  facePickerPrototype: label({
    name: "prototype · CR 718.3",
    kind: "text",
    role: "generic",
    owners: [`${BOARD}FacePickerModal.svelte`],
    doc: "#5656: the face picker's source tag on a prototype card (CR 718.3: cast it normally or prototyped)",
  }),
  facePrototyped: label({
    name: "Prototyped",
    kind: "text",
    role: "generic",
    owners: [`${BOARD}FacePickerModal.svelte`],
    doc: "#5656: the face picker's caption on a prototype card's inset, the choice that casts it with the smaller cost and size (CR 718.3b)",
  }),
  faceNotPrototyped: label({
    name: "Not prototyped",
    kind: "text",
    role: "generic",
    owners: [`${BOARD}FacePickerModal.svelte`],
    doc: "#5656: the face picker's caption on a prototype card cast normally",
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
  xPickerTurnFaceUp: label({
    name: "Turn face up",
    kind: "text",
    role: "button",
    owners: [`${BOARD}Board.svelte`],
    doc: "the X picker's confirm verb for a turn-face-up cost with an {X} (Aurelia's Vindicator's disguise {X}{3}{W}, #5249); the button reads Turn face up with X = <n>",
  }),
  xPickerManaCeiling: dynamicLabel({
    stem: "You can pay for X up to ",
    match: "prefix",
    shape: "You can pay for X up to <n> right now.",
    make: (n: number) => `You can pay for X up to ${n} right now.`,
    example: [2],
    kind: "text",
    role: "paragraph",
    owners: [`${BOARD}XCostModal.svelte`],
    doc: "the X picker's ceiling for a special action's {X} (#5249, CR 107.3d): the largest X the server says the viewer can pay from the pool and auto-tap",
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
  playerManaAbility: dynamicLabel({
    stem: " life, add ",
    match: "contains",
    shape: "<source>: pay <N> life, add <mana>",
    make: (source: string, life: number, produced: string) =>
      `${source}: pay ${life} life, add ${produced}`,
    example: ["Channel", 1, "{C}"],
    kind: "aria",
    role: "button",
    owners: [`${BOARD}PlayerManaAbilities.svelte`],
    doc: "a mana ability the viewer has as a player, bound to no card (#4885, Channel): beside the mana pool on your own panel until end of turn; each press pays the life and adds the mana, greyed when the life can't be paid (CR 119.4)",
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

  powerUpUsed: label({
    name: "already activated (power-up)",
    kind: "text",
    role: "generic",
    owners: ["lib/contextMenu.logic.ts"],
    doc: "the reason on a power-up ability row the server marked power_up_used (#4527, CR 702.193a: activate each power-up ability only once), in the card menu and the mana-ability popover",
  }),

  onceUsed: label({
    name: "already activated",
    kind: "text",
    role: "generic",
    owners: ["lib/contextMenu.logic.ts"],
    doc: "the reason on an ability row the server marked once_used (#5221, CR 602.5b: activate only once, without the exhaust keyword), in the card menu",
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
  // -- Advanced stats: the setting (ADR 0217 §2, owner answer 1, #4880) --
  advancedStatsSwitch: label({
    name: "Advanced stats",
    kind: "text",
    role: "checkbox",
    owners: ["lib/components/StatsSettings.svelte", "routes/Stats.svelte"],
    doc: "Settings → Stats and the head of #/stats: the person's own advanced-stats switch (PUT /me/stats/prefs); off unless they, an admin or the server default turned it on",
  }),
  advancedStatsOptIn: label({
    name: "Turn on advanced stats",
    kind: "text",
    role: "button",
    owners: ["lib/components/AdvancedStatsOptIn.svelte"],
    doc: "the opt-in card on #/stats while advanced stats are off for the person",
  }),
  advancedStatsDefault: label({
    name: "Advanced stats for everyone who hasn't chosen",
    kind: "text",
    role: "checkbox",
    owners: ["lib/components/admin/AdvancedStatsDefault.svelte"],
    doc: "admin Accounts: the server's advanced-stats default (PUT /admin/stats/advanced-default); anyone who chose for themselves keeps their answer",
  }),
  accountAdvancedStats: dynamicLabel({
    stem: "advanced stats for ",
    match: "prefix",
    shape: "advanced stats for <name>",
    make: (name: string) => `advanced stats for ${name}`,
    example: ["Ian"],
    kind: "aria",
    role: "checkbox",
    owners: ["lib/components/admin/AccountAdvancedStats.svelte"],
    doc: "admin Accounts and one account: an admin's advanced-stats switch for that person (PUT /admin/users/{id}/advanced-stats)",
  }),
  // -- Advanced stats: the pages and the share link (ADR 0217 §6 and §8, slice 6, #4880) --
  statsMakeupFilter: label({
    name: "Table makeup",
    kind: "text",
    role: "group",
    owners: ["lib/components/stats/StatsFilters.svelte"],
    doc: "#/stats and a shared view: the radio group that filters by who else sat at the table",
  }),
  statsRangeFilter: label({
    name: "Date range",
    kind: "text",
    role: "group",
    owners: ["lib/components/stats/StatsFilters.svelte"],
    doc: "#/stats and a shared view: the radio group that filters by when the game ended",
  }),
  statsMakeupAll: label({
    name: "All tables",
    kind: "text",
    role: "radio",
    owners: ["lib/advancedStats.ts"],
    doc: "the table makeup filter: every counted game (the default)",
  }),
  statsMakeupPeople: label({
    name: "People only",
    kind: "text",
    role: "radio",
    owners: ["lib/advancedStats.ts"],
    doc: "the table makeup filter: games where every other seat was a person",
  }),
  statsMakeupWithBots: label({
    name: "With bots",
    kind: "text",
    role: "radio",
    owners: ["lib/advancedStats.ts"],
    doc: "the table makeup filter: games with at least one bot and one person among the other seats",
  }),
  statsMakeupBots: label({
    name: "Bots only",
    kind: "text",
    role: "radio",
    owners: ["lib/advancedStats.ts"],
    doc: "the table makeup filter: games where every other seat was a bot",
  }),
  statsRangeAll: label({
    name: "All time",
    kind: "text",
    role: "radio",
    owners: ["lib/advancedStats.ts"],
    doc: "the date range filter: every game (the default)",
  }),
  statsRange90: label({
    name: "90 days",
    kind: "text",
    role: "radio",
    owners: ["lib/advancedStats.ts"],
    doc: "the date range filter: games that ended in the last 90 days",
  }),
  statsRange30: label({
    name: "30 days",
    kind: "text",
    role: "radio",
    owners: ["lib/advancedStats.ts"],
    doc: "the date range filter: games that ended in the last 30 days",
  }),
  showAsTable: label({
    name: "Show as table",
    kind: "text",
    role: "disclosure",
    owners: ["lib/components/stats/ChartFigure.svelte"],
    doc: "under every stats chart: the same numbers as a table with headers (ADR 0217 §8)",
  }),
  shareStats: label({
    name: "Share my stats",
    kind: "text",
    role: "button",
    owners: ["routes/Stats.svelte"],
    doc: "#/stats: opens the share dialog (ADR 0217 §6)",
  }),
  shareStatsDialog: label({
    name: "share your stats",
    kind: "aria",
    role: "dialog",
    owners: ["lib/components/stats/ShareStatsDialog.svelte"],
    doc: "the share dialog: make, copy and revoke the link to a read-only copy of your stats",
  }),
  makeShareLink: label({
    name: "Make a link",
    kind: "text",
    role: "button",
    owners: ["lib/components/stats/ShareStatsDialog.svelte"],
    doc: "the share dialog, with no link yet: mints one (POST /me/stats/share-link)",
  }),
  makeNewShareLink: label({
    name: "Make a new link",
    kind: "text",
    role: "button",
    owners: ["lib/components/stats/ShareStatsDialog.svelte"],
    doc: "the share dialog, with a link: mints a new one, and the old one stops working",
  }),
  copyShareLink: label({
    name: "Copy link",
    kind: "text",
    role: "button",
    owners: ["lib/components/stats/ShareStatsDialog.svelte"],
    doc: "the share dialog: copies the link just made; it is shown only once",
  }),
  revokeShareLink: label({
    name: "Revoke link",
    kind: "text",
    role: "button",
    owners: ["lib/components/stats/ShareStatsDialog.svelte"],
    doc: "the share dialog: the link answers not found from now on (DELETE /me/stats/share-link)",
  }),
  deckStats: label({
    name: "Deck stats",
    kind: "text",
    role: "link",
    owners: ["lib/components/decks/DeckCard.svelte"],
    doc: "a deck's card on #/decks: opens that deck's advanced stats page, #/stats/deck/<deck_id> (ADR 0217 §8)",
  }),
  statsVersionPicker: label({
    name: "Deck version",
    kind: "text",
    role: "combobox",
    owners: ["routes/StatsDeck.svelte"],
    doc: "a deck's stats page: picks which list's games to show, since the last edit, all time or one earlier version (ADR 0217 §5)",
  }),
  statsAllDeckGames: label({
    name: "Show all time",
    kind: "text",
    role: "button",
    owners: ["routes/StatsDeck.svelte"],
    doc: "a deck's stats page after an edit with no games yet: switches to every game with the deck",
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
    owners: ["routes/Decks.svelte", "lib/components/DeckPicker.svelte"],
    doc: "a signed-in person's saved decks: the Decks page's grid (the decks.library hint points at it) and the deck picker's My decks grid",
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
  filterCatalogueBySet: label({
    name: "filter the catalogue by set",
    kind: "aria",
    role: "combobox",
    owners: ["lib/components/SetPicker.svelte"],
    doc: "the Catalogue's set picker (#4790): a combobox that filters the sets by name or set code as you type, narrows the grid to the picked set's cards and shows how many of its cards the game plays",
  }),
  clearCatalogueSet: label({
    name: "clear the set filter",
    kind: "aria",
    role: "button",
    owners: ["lib/components/SetPicker.svelte"],
    doc: "the set picker's clear button (#4790): empties the box and goes back to All sets",
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

  // -- A table card: its seats, its footer, the Invite menu and the
  //    right-click menu (#3499, ADR 0212) --
  seatTile: dynamicLabel({
    stem: "seat ",
    match: "prefix",
    shape: "seat <n>: <name>",
    make: (n: number, name: string) => `seat ${n}: ${name}`,
    example: [1, "Luke"],
    kind: "aria",
    role: "listitem",
    owners: ["lib/components/SeatTile.svelte"],
    doc: "a lobby table's seat tile, numbered from 1 as people count; an open one is `seat <n>: open` (ADR 0212 §1)",
  }),
  readinessLine: dynamicLabel({
    stem: "ready ",
    match: "prefix",
    shape: "ready <n> of <seated>",
    make: (ready: number, seated: number) => `ready ${ready} of ${seated}`,
    example: [2, 3],
    kind: "aria",
    role: "group",
    owners: ["routes/Lobby.svelte"],
    doc: "a lobby table's readiness line, in its footer: how many of the seated players have a deck in (ADR 0212 §2)",
  }),
  everyDeckIn: dynamicLabel({
    stem: "Every deck is in",
    match: "prefix",
    shape:
      "Every deck is in[ at <table>]. You can start the game. (or: Waiting for <host> to start.)",
    make: (startable: boolean, host: string, table?: string) =>
      `Every deck is in${table ? ` at ${table}` : ""}. ${
        startable ? "You can start the game." : `Waiting for ${host} to start.`
      }`,
    example: [true, "Luke"],
    kind: "text",
    role: "status",
    owners: ["lib/readyMoment.ts"],
    doc: "the lobby page's polite live region, said once when this client sees a lobby table go from not-all-ready to two or more seated with every deck in (ADR 0212 §2, lib/readyMoment.ts); the table's name follows only when you take part in more than one lobby table",
  }),
  chooseYourDeck: label({
    name: "Choose your deck",
    kind: "text",
    role: "button",
    owners: ["lib/tablesPage.ts", "routes/Lobby.svelte"],
    doc: "a lobby table's primary while your seat has no deck: opens the deck picker under the seats (ADR 0212 §2)",
  }),
  startGame: dynamicLabel({
    stem: "Start game",
    match: "prefix",
    shape: "Start game · <n> players",
    make: (players: number) => (players < 4 ? `Start game · ${players} players` : "Start game"),
    example: [3],
    kind: "text",
    role: "button",
    owners: ["lib/tablesPage.ts"],
    doc: "a lobby table's primary once every seated player has a deck, for whoever may start it; plain Start game at four (ADR 0212 §2)",
  }),
  waitingFor: dynamicLabel({
    stem: "Waiting for ",
    match: "prefix",
    shape: "Waiting for <who>",
    make: (who: string) => `Waiting for ${who}`,
    example: ["2 players"],
    kind: "text",
    role: "button",
    owners: ["lib/tablesPage.ts"],
    doc: "a lobby table's disabled primary: Waiting for 2 players, Waiting for players, Waiting for <host> to start (ADR 0212 §2)",
  }),
  openTable: label({
    name: "Open table",
    kind: "text",
    role: "button",
    owners: ["lib/tablesPage.ts", "routes/Lobby.svelte"],
    doc: "opens a lobby table on the board: the primary for a seat held elsewhere and for the admin, otherwise a ghost button in the footer and a right-click item",
  }),
  inviteButton: label({
    name: "Invite",
    kind: "text",
    role: "button",
    owners: ["routes/Lobby.svelte", "lib/components/SeatTile.svelte"],
    doc: "opens a table's Invite menu: the footer's Invite ▾, and the same on each open seat (ADR 0212 §4)",
  }),
  inviteMenu: dynamicLabel({
    stem: "invite to ",
    match: "prefix",
    shape: "invite to <table name>",
    make: (table: string) => `invite to ${table}`,
    example: ["Friday Pod"],
    kind: "aria",
    role: "menu",
    owners: ["routes/Lobby.svelte"],
    doc: "a table's Invite menu: the two links, Invite by DM, #looking-for-game and Make new links (ADR 0212 §4)",
  }),
  copyInvite: label({
    name: "Copy invite",
    kind: "text",
    role: "menuitem",
    owners: ["routes/Lobby.svelte", "lib/components/PlayedWithDrawer.svelte"],
    doc: "copies a lobby table's player invite link, from the Invite menu and the right-click menu; a Played with row's button where the server has no bot to DM it (ADR 0213 §1)",
  }),
  newInviteLink: label({
    name: "New invite link…",
    kind: "text",
    role: "menuitem",
    owners: ["routes/Lobby.svelte"],
    doc: "mints a new player invite after a confirm, then copies it; the current link stops working",
  }),
  copySpectatorLink: label({
    name: "Copy spectator link",
    kind: "text",
    role: "menuitem",
    owners: ["routes/Lobby.svelte"],
    doc: "copies a table's read-only link, from the Invite menu and the right-click menu",
  }),
  newSpectatorLink: label({
    name: "New spectator link…",
    kind: "text",
    role: "menuitem",
    owners: ["routes/Lobby.svelte"],
    doc: "mints a new spectator link after a confirm, then copies it; the current link stops working",
  }),
  tableSettingsLine: dynamicLabel({
    stem: "table settings: ",
    match: "prefix",
    shape: "table settings: <summary>",
    make: (summary: string) => `table settings: ${summary}`,
    example: ["40 life · 1 undo · standard rules"],
    kind: "aria",
    role: "button",
    owners: ["routes/Lobby.svelte"],
    doc: "a table card's settings summary, life and undos and then what else differs: opens the table settings panel (read-only for a non-host)",
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

  // -- The deck picker and the bot picker under a lobby table's seats,
  //    and the table route's deck dialog (ADR 0212 §3, owner decision 8) --
  deckPicker: label({
    name: "choose your deck",
    kind: "aria",
    role: "region",
    owners: ["routes/Lobby.svelte"],
    doc: "the deck picker under a lobby table's seats, its notch at your seat: the tabs, the grid and the detail line (ADR 0212 §3)",
  }),
  myDecksTab: label({
    name: "My decks",
    kind: "text",
    role: "tab",
    owners: ["lib/components/DeckPicker.svelte"],
    doc: "the deck picker's tab of saved decks, with their count; absent for a guest or an empty library (owner decision 7)",
  }),
  prebuiltTab: label({
    name: "Pre-built",
    kind: "text",
    role: "tab",
    owners: ["lib/components/DeckPicker.svelte"],
    doc: "the deck picker's tab of pre-built decks, with their count; where a guest opens",
  }),
  pasteListTab: label({
    name: "Paste a list",
    kind: "text",
    role: "tab",
    owners: ["lib/components/DeckPicker.svelte"],
    doc: "the deck picker's tab with the decklist box: an Archidekt or Moxfield link, or a plain-text list",
  }),
  prebuiltDecks: label({
    name: "pre-built decks",
    kind: "aria",
    role: "radiogroup",
    owners: [
      "lib/components/DeckPicker.svelte",
      "lib/components/BotPicker.svelte",
      "routes/Decks.svelte",
    ],
    doc: "the Pre-built grid, a radio group of deck tiles, in the deck picker and the Add bot picker; the Decks page's read-only list of them",
  }),
  decklistSource: label({
    name: "decklist source",
    kind: "aria",
    role: "textbox",
    owners: ["lib/components/DeckPicker.svelte"],
    doc: "Paste a list's box: a deck link, or a plain-text decklist",
  }),
  sitDownWithThisDeck: label({
    name: "Sit down with this deck",
    kind: "text",
    role: "button",
    owners: ["lib/components/DeckPicker.svelte"],
    doc: "seats the deck selected in My decks or Pre-built (#4655): the lobby card's footer primary while the picker is open, or the deck dialog's button (ADR 0212 §2)",
  }),
  uploadDeck: label({
    name: "Upload deck",
    kind: "text",
    role: "button",
    owners: ["lib/components/DeckPicker.svelte"],
    doc: "seats the decklist pasted in Paste a list: the footer primary, or the deck dialog's button",
  }),
  importDeck: label({
    name: "Import deck",
    kind: "text",
    role: "button",
    owners: ["lib/components/DeckPicker.svelte"],
    doc: "what Upload deck says when Paste a list holds a deck link",
  }),
  addBot: label({
    name: "Add bot",
    kind: "text",
    role: "button",
    owners: ["lib/components/SeatTile.svelte", "lib/components/BotPicker.svelte"],
    doc: "an open seat's Add bot, which opens the bot picker; then the footer primary that seats the bot (owner decision 8)",
  }),
  addBotTo: dynamicLabel({
    stem: "Add a bot to seat ",
    match: "prefix",
    shape: "Add a bot to seat <n>",
    make: (seat: number) => `Add a bot to seat ${seat}`,
    example: [3],
    kind: "text",
    role: "region",
    owners: ["lib/components/BotPicker.svelte"],
    doc: "the bot picker under the seats, named by its title: the tier, your own model key, and the Pre-built grid (owner decision 8)",
  }),
  botTier: label({
    name: "bot tier",
    kind: "aria",
    role: "radiogroup",
    owners: ["lib/components/BotPicker.svelte"],
    doc: "the bot picker's tier control; a tier the server cannot run is disabled, and the line under it says why",
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

  // -- Pods (ADR 0220 §11, slice 3) --
  navPods: label({
    name: "Pods",
    kind: "text",
    role: "link",
    owners: ["lib/components/SiteHeader.svelte", "routes/Stats.svelte"],
    doc: "the signed-in nav's link to #/pods, after Stats, and the Pods link on #/stats (ADR 0220 §11)",
  }),
  newPod: label({
    name: "New pod",
    kind: "text",
    role: "button",
    owners: ["routes/Pods.svelte"],
    doc: "#/pods' one primary: asks for a name and makes a pod with you as its owner (ADR 0220 §5)",
  }),
  acceptPodInvite: label({
    name: "Join pod",
    kind: "text",
    role: "button",
    owners: ["routes/Pods.svelte"],
    doc: "accepts a friend's invite to a pod, on #/pods; the invite's own yes (ADR 0220 §3a)",
  }),
  declinePodInvite: label({
    name: "Decline",
    kind: "text",
    role: "button",
    owners: ["routes/Pods.svelte"],
    doc: "takes a pod invite off your list, silently: the pod still sees it pending (ADR 0220 §3a)",
  }),
  inviteToPod: label({
    name: "Invite a friend",
    kind: "text",
    role: "button",
    owners: ["routes/Pod.svelte"],
    doc: "opens the friend picker on a pod's page; any member can invite one of their own friends (ADR 0220 §3a)",
  }),
  leavePod: label({
    name: "Leave pod",
    kind: "text",
    role: "button",
    owners: ["routes/Pod.svelte"],
    doc: "leaves a pod after a one-line confirm; silent, nobody is told (ADR 0220 §3)",
  }),
  removePodMember: label({
    name: "Remove",
    kind: "text",
    role: "button",
    owners: ["routes/Pod.svelte"],
    doc: "the owner's removal of a member, with its Also make a new invite link checkbox (ADR 0220 §3)",
  }),
  makePodOwner: label({
    name: "Make owner",
    kind: "text",
    role: "button",
    owners: ["routes/Pod.svelte"],
    doc: "the owner hands the pod to a member and stays a member (ADR 0220 §2)",
  }),
  makePodLink: label({
    name: "New link",
    kind: "text",
    role: "button",
    owners: ["routes/Pod.svelte"],
    doc: "mints or rotates the pod's invite link; the old one stops at once and the new one is shown once (ADR 0220 §3b)",
  }),
  copyPodLink: label({
    name: "Copy invite link",
    kind: "text",
    role: "button",
    owners: ["routes/Pod.svelte"],
    doc: "copies the just-minted invite link; the token is never shown again (ADR 0220 §3b)",
  }),
  makeSuggestedPod: label({
    name: "Make this pod",
    kind: "text",
    role: "button",
    owners: ["routes/Pods.svelte"],
    doc: "a suggested pod's yes, on #/pods: asks for a name, makes the pod with you as its owner and invites the others in the suggestion (ADR 0220 §4, slice 4)",
  }),
  dismissPodSuggestion: label({
    name: "Not now",
    kind: "text",
    role: "button",
    owners: ["routes/Pods.svelte"],
    doc: "hides a suggested pod from you for 180 days and from nobody else (ADR 0220 §4, slice 4)",
  }),
  sharePodsSwitch: label({
    name: "Share my stats with my pods",
    kind: "text",
    role: "checkbox",
    owners: ["lib/components/PodShareSwitch.svelte"],
    doc: "Settings → Stats and a pod's page: the person's one pod-sharing switch (PUT /me/stats/prefs share_pods); it covers every pod they are in, and joining a pod never turns it on (ADR 0220 §7)",
  }),
  podStatsLink: label({
    name: "Pod stats",
    kind: "text",
    role: "link",
    owners: ["routes/Pod.svelte"],
    doc: "a pod's page: the link to #/pods/<id>/stats, the pod's leaderboard and the sharing members' records (ADR 0220 §11)",
  }),
  joinPod: label({
    name: "Join",
    kind: "text",
    role: "button",
    owners: ["routes/PodJoin.svelte"],
    doc: "the invite link page's yes: joins the pod (ADR 0220 §3b)",
  }),

  // -- Friends (ADR 0213 §2, §3, §9, slice 3) --
  friendsButton: label({
    name: "Friends",
    kind: "aria",
    role: "button",
    owners: ["lib/components/SiteHeader.svelte"],
    doc: "the site header's people button, for someone signed in with Discord (ADR 0213 §9; it read People you've played with in slice 2); opens the friends drawer, which fetches then and never polls",
  }),
  friendsDrawer: label({
    name: "Friends list",
    kind: "aria",
    role: "dialog",
    owners: ["lib/components/PlayedWithDrawer.svelte"],
    doc: "the header's drawer (ADR 0213 §9): Requests, Friends and Played with, each Played with row with Add friend and Invite; covers the page at tablet width (it was the Played with dialog in slice 2)",
  }),
  playedWith: label({
    name: "Played with",
    kind: "text",
    role: "heading",
    owners: ["lib/components/PlayedWithDrawer.svelte"],
    doc: "the friends drawer's section of the people you have shared a table with and are not friends with yet, newest first (ADR 0213 §1)",
  }),
  addFriend: label({
    name: "Add friend",
    kind: "text",
    role: "button",
    owners: ["lib/friends.ts", "lib/components/SeatTile.svelte"],
    doc: "sends a friend request (POST /me/friends/requests): a Played with row by account, another signed-in person's seat at your own table by seat (ADR 0213 §2)",
  }),
  acceptFriend: label({
    name: "Accept",
    kind: "text",
    role: "button",
    owners: ["lib/friends.ts", "lib/components/PlayedWithDrawer.svelte"],
    doc: "answers yes to a friend request: a Requests row in the drawer, or a Played with row of someone who asked you (ADR 0213 §2)",
  }),
  ignoreFriend: label({
    name: "Ignore",
    kind: "text",
    role: "button",
    owners: ["lib/components/PlayedWithDrawer.svelte"],
    doc: "takes a friend request off your list, silently: the sender still sees it pending until it expires (ADR 0213 §2)",
  }),
  manageFriends: label({
    name: "Requests sent, blocks and settings",
    kind: "text",
    role: "link",
    owners: ["lib/components/PlayedWithDrawer.svelte"],
    doc: "the friends drawer's link to #/friends (ADR 0213 §9)",
  }),
  cancelFriendRequest: label({
    name: "Cancel request",
    kind: "text",
    role: "button",
    owners: ["routes/Friends.svelte"],
    doc: "withdraws a friend request you sent, on #/friends",
  }),
  removeFriend: label({
    name: "Remove friend",
    kind: "text",
    role: "button",
    owners: ["routes/Friends.svelte"],
    doc: "ends a friendship from your side, silently (ADR 0213 §2), on #/friends after a confirm",
  }),
  blockPerson: label({
    name: "Block",
    kind: "text",
    role: "button",
    owners: ["routes/Friends.svelte"],
    doc: "blocks a friend, a requester or someone you played with, after a confirm that says what it ends (ADR 0213 §3); on #/friends",
  }),
  unblockPerson: label({
    name: "Unblock",
    kind: "text",
    role: "button",
    owners: ["routes/Friends.svelte"],
    doc: "lifts a block on #/friends; it brings back nothing the block ended",
  }),
  whoCanRequest: label({
    name: "Who can send me friend requests",
    kind: "aria",
    role: "radiogroup",
    owners: ["routes/Friends.svelte"],
    doc: "the #/friends setting: people I've played with (default) or nobody (ADR 0213 §2)",
  }),
  whoCanSee: label({
    name: "Who can see what I'm doing on CMD & CTRL",
    kind: "aria",
    role: "radiogroup",
    owners: ["routes/Friends.svelte"],
    doc: "the #/friends setting: my friends (default), my friends and people I've played with, or nobody (ADR 0213 §5); stored now, shown to others from slice 4",
  }),
  historyWithPeople: label({
    name: "With people",
    kind: "text",
    role: "button",
    owners: ["lib/tablesPage.ts"],
    doc: "Tables › History's filter for a table where another person (not a bot) sat; ADR 0158 called it With friends until ADR 0213 §9",
  }),
  historyWithFriends: label({
    name: "With friends",
    kind: "text",
    role: "button",
    owners: ["lib/tablesPage.ts"],
    doc: "Tables › History's filter for a table where one of your current friends sat (ADR 0213 §9)",
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
  opponentCardFacing: label({
    name: "Opponents' cards",
    kind: "text",
    role: "combobox",
    owners: ["lib/components/Settings.svelte"],
    doc: "Settings → Display: whether a top-row opponent's cards face them (turned 180°, the default) or the viewer, display.opponentCardFacing (#4555)",
  }),
  keywordFlourishesToggle: label({
    name: "Show keyword flourishes on battlefield cards",
    kind: "text",
    role: "checkbox",
    owners: ["lib/components/Settings.svelte"],
    doc: "Settings → Display: the quiet per-keyword cues on battlefield cards (flying hover, reach lean, deathtouch edge, strike blades, lifelink glow), display.keywordFlourishes (#4556)",
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
  effectExtort: label({
    name: "Extort",
    kind: "text",
    role: "status",
    owners: [`${BOARD}ExtortScales.svelte`],
    doc: "the caption under the Scales of Souls, the table effect an extort drain plays (#4544); the whole scene is aria-hidden, the log and the life popups carry the news",
  }),
  effectExtortMany: dynamicLabel({
    stem: "Extort ×",
    match: "prefix",
    shape: "Extort ×<N>",
    make: (n: number) => `Extort ×${n}`,
    example: [3],
    kind: "text",
    role: "status",
    owners: [`${BOARD}ExtortScales.svelte`],
    doc: "the same caption when several extort triggers resolved in one scale's wait and were merged into it",
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
  gameClosed: label({
    name: "This table was closed.",
    kind: "text",
    role: "status",
    owners: ["lib/gameClosed.ts", "lib/connectionBanner.ts"],
    doc: "ADR 0211 §4, #5082: what a refused join, spectate, start, deck, bot or setup says for the server's `game_closed` code, and the connection banner's headline once a failed redial finds the table archived",
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
  respondValueInstantsToggle: label({
    name: "Value instants (Opt, Brainstorm) on an opponent's spell",
    kind: "text",
    role: "checkbox",
    owners: ["lib/components/Settings.svelte"],
    doc: "Settings → Gameplay → Advanced, What counts as a response: off by default, so a value instant does not stop you on an opponent's spell; it still does in combat and at an opponent's end step, gameplay.respondValueInstants (#2877)",
  }),
  respondTurnFaceUpToggle: label({
    name: "Turning a face-down creature face up",
    kind: "text",
    role: "checkbox",
    owners: ["lib/components/Settings.svelte"],
    doc: "Settings → Gameplay → Advanced, What counts as a response: on by default, so a morph, disguise or manifested creature you can turn face up stops you on an opponent's spell, gameplay.respondTurnFaceUp (#2861)",
  }),
  respondOtherSpecialActionsToggle: label({
    name: "Other special actions (foretell, suspend, plot, unlock)",
    kind: "text",
    role: "checkbox",
    owners: ["lib/components/Settings.svelte"],
    doc: "Settings → Gameplay → Advanced, What counts as a response: off by default, so a special action that answers nothing does not stop you; on your own main phase it still counts as something to do, gameplay.respondOtherSpecialActions (#2861)",
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
  searchEveryZone: label({
    name: "Graveyard, hand and library",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ChoicePromptModal.svelte`],
    doc: 'the primary of the question a "search your graveyard, hand, and/or library" prompt asks before it shows anything (#5408, CR 701.23): search all three zones. The library is shuffled afterwards',
  }),
  searchGraveyardAndHand: label({
    name: "Graveyard and hand",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ChoicePromptModal.svelte`],
    doc: 'a "search your graveyard, hand, and/or library" prompt zone question (#5408): search the graveyard and the hand, not the library; it is never shown or shuffled',
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
    doc: 'a "search your library and/or graveyard" (or "graveyard, hand, and/or library", #5408) prompt zone question (#4519): search the graveyard alone; the library is never shown or shuffled',
  }),
  searchHandOnly: label({
    name: "Hand only",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ChoicePromptModal.svelte`],
    doc: 'a "search your hand and/or library" (or "graveyard, hand, and/or library", #5408) prompt zone question (#4519): search the hand alone; the library is never shown or shuffled',
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
  // -- The draftable sets page (ADR 0222 §4, #5437) --
  draftPlayBooster: label({
    name: "Play booster",
    kind: "text",
    role: "generic",
    owners: ["lib/draftSets.ts"],
    doc: "the draftable sets page: a set whose packs are modelled on a play booster (sets released from February 2024 on)",
  }),
  draftDraftBooster: label({
    name: "Draft booster",
    kind: "text",
    role: "generic",
    owners: ["lib/draftSets.ts"],
    doc: "the draftable sets page: a set whose packs are modelled on a draft booster (sets before play boosters)",
  }),
  // -- Card ratings and colour pairs (ADR 0222 §7, #5703) --
  draftRatingsCredit: label({
    name: "Card ratings from 17Lands",
    kind: "text",
    role: "link",
    owners: [`${DRAFT}DraftRatingsBar.svelte`],
    doc: "the credit 17Lands' data is used on: shown wherever a 17Lands rating shows beside a draft card; links to 17lands.com (the owner's condition, 2026-11-11)",
  }),
  draftRatingsShow: label({
    name: "Show card ratings",
    kind: "text",
    role: "checkbox",
    owners: [`${DRAFT}DraftRatingsBar.svelte`],
    doc: "the viewer's switch for the rating beside each draft card, kept in this browser only; on by default",
  }),
  draftArchetypesTitle: label({
    name: "Colour pairs",
    kind: "text",
    role: "heading",
    owners: ["routes/DraftSets.svelte", `${DRAFT}DraftArchetypesPanel.svelte`],
    doc: "a set's ten colour pairs with a short label and the gold uncommons that name each: the disclosure on the draftable sets page and the panel beside a draft",
  }),
  // -- The booster draft (ADR 0222 §5, #5438) --
  draftThisSet: label({
    name: "Draft this set",
    kind: "text",
    role: "link",
    owners: ["routes/DraftSets.svelte"],
    doc: "the draftable sets page: a ready set's link to the create form with that set chosen (#/draft?set=<code>)",
  }),
  draftAgainstBots: label({
    name: "Draft against bots",
    kind: "text",
    role: "link",
    owners: ["routes/FrontPage.svelte"],
    doc: "the front page's way into a draft without an account: #/draft, a solo draft against bots when signed out",
  }),
  draftNewTitle: label({
    name: "Start a draft",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's page title, at #/draft",
  }),
  draftSetField: label({
    name: "Set to draft",
    kind: "text",
    role: "combobox",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's set picker: the draftable sets only, newest first",
  }),
  draftPickMode: label({
    name: "Pick mode",
    kind: "text",
    role: "radiogroup",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's pick mode group: pick one or pick two",
  }),
  draftPickOne: label({
    name: "Pick one",
    kind: "text",
    role: "radio",
    owners: ["lib/draft.ts"],
    doc: "a draft's pick mode: one card from each pack (the create form's option, the join page, the draft's head)",
  }),
  draftPickTwo: label({
    name: "Pick two",
    kind: "text",
    role: "radio",
    owners: ["lib/draft.ts"],
    doc: "a draft's pick mode: two cards from each pack, offered first at four seats",
  }),
  draftSeatsField: label({
    name: "Seats",
    kind: "text",
    role: "combobox",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's number of seats, 2 to 8; a seat no person takes is a bot",
  }),
  draftNameField: label({
    name: "Draft name",
    kind: "text",
    role: "textbox",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's optional name for the draft",
  }),
  draftCreate: label({
    name: "Create draft",
    kind: "text",
    role: "button",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's primary: POST /draft, then the draft's own page",
  }),
  draftReturn: label({
    name: "Return to your draft",
    kind: "text",
    role: "link",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's link back to a draft this browser still holds a seat in",
  }),
  draftSeeSets: label({
    name: "See which sets can be drafted",
    kind: "text",
    role: "link",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's link to #/draft/sets",
  }),
  draftGuestName: label({
    name: "Name at the draft",
    kind: "text",
    role: "textbox",
    owners: [`${DRAFT}DraftJoin.svelte`],
    doc: "the join page's name field, shown to a visitor who is not signed in",
  }),
  draftJoin: label({
    name: "Join draft",
    kind: "text",
    role: "button",
    owners: [`${DRAFT}DraftJoin.svelte`],
    doc: "the join page's primary: POST /draft/{id}/join with the link's invite",
  }),
  draftStart: label({
    name: "Start draft",
    kind: "text",
    role: "button",
    owners: [`${DRAFT}DraftRoom.svelte`],
    doc: "the draft lobby's primary, for the host: every open seat becomes a bot and the first pack is dealt",
  }),
  draftLeave: label({
    name: "Leave draft",
    kind: "text",
    role: "button",
    owners: [`${DRAFT}DraftRoom.svelte`],
    doc: "leave a draft: frees the seat in the lobby; once picking has begun it asks first, because a bot takes the seat",
  }),
  draftLeaveConfirm: label({
    name: "Leave and let a bot draft",
    kind: "text",
    role: "button",
    owners: [`${DRAFT}DraftRoom.svelte`],
    doc: "the second step of leaving a draft that is picking: the seat becomes a bot for the rest of the draft",
  }),
  draftStay: label({
    name: "Keep drafting",
    kind: "text",
    role: "button",
    owners: [`${DRAFT}DraftRoom.svelte`],
    doc: "the way back from the leave question in a draft that is picking",
  }),
  draftCopyInvite: label({
    name: "Copy draft invite",
    kind: "text",
    role: "button",
    owners: [`${DRAFT}DraftRoom.svelte`],
    doc: "the draft lobby's copy button beside the host's invite link; never shown in a solo draft",
  }),
  draftInviteLink: label({
    name: "draft invite link",
    kind: "aria",
    role: "textbox",
    owners: [`${DRAFT}DraftRoom.svelte`],
    doc: "the read-only field that holds the host's invite link in the draft lobby",
  }),
  draftSeats: label({
    name: "seats at this draft",
    kind: "aria",
    role: "list",
    owners: [`${DRAFT}DraftSeats.svelte`],
    doc: "the draft's seat list: each seat's name and status (host, bot, picking, picked, away) and how many cards it has drafted, never the cards",
  }),
  draftPack: label({
    name: "the pack you hold",
    kind: "aria",
    role: "group",
    owners: [`${DRAFT}DraftPack.svelte`],
    doc: "the pick screen's pack: one toggle button per card, named for the card",
  }),
  draftConfirmPick: dynamicLabel({
    stem: "Pick ",
    match: "prefix",
    shape: "Pick <n> card(s)",
    make: (take: number) => `Pick ${take} ${take === 1 ? "card" : "cards"}`,
    example: [2],
    kind: "text",
    role: "button",
    owners: ["lib/draft.ts"],
    doc: "the pick screen's primary: sends draft_pick with the chosen cards, enabled only when exactly that many are chosen",
  }),
  draftTimer: label({
    name: "pick timer",
    kind: "aria",
    role: "timer",
    owners: [`${DRAFT}DraftRoom.svelte`],
    doc: "the pick screen's countdown, run from the server's clock; urgent in the last 10 seconds",
  }),
  draftWaiting: label({
    name: "Waiting for the others",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftRoom.svelte`],
    doc: "the pick screen after the viewer has picked, above the names of the seats still picking",
  }),
  draftYourPicks: label({
    name: "Your picks",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftRoom.svelte`],
    doc: "the pick screen's list of the viewer's own picks so far, grouped by colour and mana value",
  }),
  draftYourPool: label({
    name: "Your pool",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftRoom.svelte`],
    doc: "a finished draft's card pool, grouped by colour and mana value",
  }),
  draftFinished: label({
    name: "The draft is finished",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftRounds.svelte`],
    doc: "a finished draft's headline, once its rounds are over or every person has left",
  }),
  // -- After the picks: the build and the rounds (ADR 0222 §5, §6, #5440) --
  draftBuildDeck: label({
    name: "Build your deck",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftBuild.svelte`],
    doc: "the deckbuilding screen's headline, once a draft's last pick is made",
  }),
  draftBuildTimer: label({
    name: "build timer",
    kind: "aria",
    role: "timer",
    owners: [`${DRAFT}DraftRoom.svelte`],
    doc: "the deckbuilding screen's countdown, run from the server's clock; when it ends a deck is built for anyone who has not submitted one",
  }),
  draftDeckZone: label({
    name: "Main",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftSplit.svelte`],
    doc: "#5699: the picks' main area, beside the pack and on the deckbuilding screen: the cards to play, each a button that moves it to the sideboard, and a drop target; also the stats' heading for the main cards",
  }),
  draftSideboardZone: label({
    name: "Sideboard",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftSplit.svelte`],
    doc: "#5699: the picks' sideboard, beside the pack and on the deckbuilding screen: every pick not in main, each a button that moves it back, and a drop target",
  }),
  draftBasicLands: label({
    name: "Basic lands",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftBuild.svelte`],
    doc: "the deckbuilding screen's steppers, one per basic land type; any number may be added",
  }),
  draftBasicStepper: dynamicLabel({
    stem: "how many ",
    match: "prefix",
    shape: "how many <basic land>",
    make: (land: string) => `how many ${land}`,
    example: ["Forest"],
    kind: "aria",
    role: "group",
    owners: [`${DRAFT}DraftBuild.svelte`],
    doc: "one basic land's stepper on the deckbuilding screen: a fewer button, the count and a more button",
  }),
  draftCurve: label({
    name: "mana curve of your deck",
    kind: "aria",
    role: "img",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "the stats' curve of the main cards (#5702), on the pick and deckbuilding screens: the cards that are not lands, counted by mana value from 0 to 6 and more",
  }),
  draftCurvePool: label({
    name: "mana curve of your pool",
    kind: "aria",
    role: "img",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "the stats' curve of the whole pool (#5702): main and sideboard together, lands left out, counted by mana value from 0 to 6 and more",
  }),
  draftStats: label({
    name: "Draft stats",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "#5702: the area on the pick and deckbuilding screens that tallies the picks: creatures against the rest, card types, the mana curve and the colours, for main and for the whole pool",
  }),
  draftStatsPool: label({
    name: "Whole pool",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "#5702: the stats' heading for every card picked, main and sideboard together",
  }),
  draftStatNonCreatures: label({
    name: "Non-creatures",
    kind: "text",
    role: "term",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "#5702: the stats' count of cards that are not creatures, beside the creature count",
  }),
  draftStatTypes: label({
    name: "Card types",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "#5702: the stats' count for each card type; a card with two types is in both counts",
  }),
  draftStatCurve: label({
    name: "Mana curve",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "#5702: the heading above a stats curve",
  }),
  draftStatColours: label({
    name: "Colours",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "#5702: the stats' count for each colour, then several colours and none; a card of two colours is in both counts",
  }),
  draftTypeCreature: label({
    name: "Creatures",
    kind: "text",
    role: "term",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "#5702: the stats' creature count",
  }),
  draftTypeInstant: label({
    name: "Instants",
    kind: "text",
    role: "term",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "#5702: the stats' instant count",
  }),
  draftTypeSorcery: label({
    name: "Sorceries",
    kind: "text",
    role: "term",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "#5702: the stats' sorcery count",
  }),
  draftTypeArtifact: label({
    name: "Artifacts",
    kind: "text",
    role: "term",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "#5702: the stats' artifact count",
  }),
  draftTypeEnchantment: label({
    name: "Enchantments",
    kind: "text",
    role: "term",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "#5702: the stats' enchantment count",
  }),
  draftTypePlaneswalker: label({
    name: "Planeswalkers",
    kind: "text",
    role: "term",
    owners: [`${DRAFT}DraftStats.svelte`],
    doc: "#5702: the stats' planeswalker count",
  }),
  draftUseSuggestion: label({
    name: "Use the suggested lands",
    kind: "text",
    role: "button",
    owners: [`${DRAFT}DraftBuild.svelte`],
    doc: "the deckbuilding screen's button that sets the basic land steppers to the land-count suggestion",
  }),
  draftSubmitDeck: label({
    name: "Submit deck",
    kind: "text",
    role: "button",
    owners: [`${DRAFT}DraftBuild.svelte`],
    doc: "the deckbuilding screen's primary: sends draft_deck; disabled below the 40-card minimum",
  }),
  draftMatchLength: label({
    name: "Matches",
    kind: "text",
    role: "radiogroup",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's choice of the draft's match length: best of one (the default) or best of three",
  }),
  draftBuildTime: label({
    name: "Deck building",
    kind: "text",
    role: "radiogroup",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's choice of how deck building after the last pick runs (#5700): on a timer (the default) or untimed",
  }),
  draftBuildTimed: label({
    name: "Timed",
    kind: "text",
    role: "radio",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's deck building radio for a timer: when it ends, a deck is built for anyone who has not submitted one",
  }),
  draftBuildUntimed: label({
    name: "Untimed",
    kind: "text",
    role: "radio",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's deck building radio for no timer: the rounds begin when every person has submitted a deck",
  }),
  draftBuildMinutes: label({
    name: "Minutes to build",
    kind: "text",
    role: "combobox",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's length of the deck building timer, 2 to 30 minutes, shown while deck building is timed",
  }),
  draftRound: dynamicLabel({
    stem: "Round ",
    match: "prefix",
    shape: "Round <n> of <m>",
    make: (round: number, rounds: number) => `Round ${round} of ${rounds}`,
    example: [2, 3],
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftRounds.svelte`],
    doc: "the rounds screen's headline while a round is in play",
  }),
  draftPlayMatch: label({
    name: "Play your match",
    kind: "text",
    role: "button",
    owners: [`${DRAFT}DraftRounds.svelte`],
    doc: "the rounds screen's primary: POST /draft/{id}/match, then the limited table the match is played at",
  }),
  draftRecords: label({
    name: "Win-loss table",
    kind: "text",
    role: "heading",
    owners: [`${DRAFT}DraftRounds.svelte`],
    doc: "the rounds screen's table of every seat's wins and losses; the draft's only standings",
  }),
  draftBackToDraft: label({
    name: "Back to the draft",
    kind: "text",
    role: "button",
    owners: ["routes/Game.svelte"],
    doc: "the way out of a draft match's table: the game-over corner button and the leave control go to the draft, not the lobby",
  }),
  myDrafts: label({
    name: "My drafts",
    kind: "text",
    role: "link",
    owners: ["routes/MyDrafts.svelte", `${DRAFT}DraftCreate.svelte`, `${DRAFT}DraftRounds.svelte`],
    doc: "a signed-in person's finished drafts at #/drafts: the page's title, and the link to it from the draft pages",
  }),
  // -- Cubes (ADR 0222 §8, #5441) --
  myCubes: label({
    name: "My cubes",
    kind: "text",
    role: "link",
    owners: ["routes/MyCubes.svelte", `${DRAFT}DraftCreate.svelte`],
    doc: "a signed-in person's saved cubes at #/cubes: the page's title, and the link to it from the create-draft form",
  }),
  cubeNew: label({
    name: "Add a cube",
    kind: "text",
    role: "heading",
    owners: ["routes/MyCubes.svelte"],
    doc: "the My cubes page's form for a new cube: a name, a pasted list or a deck link, and the singleton box",
  }),
  cubeNameField: label({
    name: "Cube name",
    kind: "text",
    role: "textbox",
    owners: ["routes/MyCubes.svelte"],
    doc: "the new-cube form's name; with a deck link it may be left empty and the deck's own name is used",
  }),
  cubeListField: label({
    name: "Card list",
    kind: "text",
    role: "textbox",
    owners: ["routes/MyCubes.svelte"],
    doc: "the new-cube form's pasted list, one card a line, and the same field when a saved cube's list is edited",
  }),
  cubeLinkField: label({
    name: "Deck link",
    kind: "text",
    role: "textbox",
    owners: ["routes/MyCubes.svelte"],
    doc: "the new-cube form's link to a list on a supported deck site, in place of a pasted list",
  }),
  cubeSingleton: label({
    name: "Singleton: one copy of each card",
    kind: "text",
    role: "checkbox",
    owners: ["routes/MyCubes.svelte"],
    doc: "a cube's singleton setting, on by default: off, the list may repeat a card; basic lands are never limited",
  }),
  cubeSave: label({
    name: "Save cube",
    kind: "text",
    role: "button",
    owners: ["routes/MyCubes.svelte"],
    doc: "the new-cube form's primary: POST /me/cubes, then the cube appears in the list with its check",
  }),
  cubeCheck: label({
    name: "Check cube",
    kind: "text",
    role: "button",
    owners: ["routes/MyCubes.svelte"],
    doc: "a saved cube's check: GET /me/cubes/{id}, then the cards that block it are named under the cube",
  }),
  cubeDropBlockers: label({
    name: "Drop the cards that are not ready",
    kind: "text",
    role: "button",
    owners: ["routes/MyCubes.svelte"],
    doc: "under a checked cube with cards that are unknown or not fully built: saves the list without them (PUT /me/cubes/{id})",
  }),
  cubeEditList: label({
    name: "Edit list",
    kind: "text",
    role: "button",
    owners: ["routes/MyCubes.svelte"],
    doc: "under a checked cube: opens its list in a text field to change and save",
  }),
  cubeSaveList: label({
    name: "Save list",
    kind: "text",
    role: "button",
    owners: ["routes/MyCubes.svelte"],
    doc: "saves an edited cube list (PUT /me/cubes/{id}) and shows the new check",
  }),
  cubeDelete: label({
    name: "Delete cube",
    kind: "text",
    role: "button",
    owners: ["routes/MyCubes.svelte"],
    doc: "a saved cube's delete; it asks once before DELETE /me/cubes/{id}",
  }),
  cubeDeleteConfirm: label({
    name: "Delete it for good",
    kind: "text",
    role: "button",
    owners: ["routes/MyCubes.svelte"],
    doc: "the second step of deleting a cube",
  }),
  cubeDraft: label({
    name: "Draft this cube",
    kind: "text",
    role: "link",
    owners: ["routes/MyCubes.svelte"],
    doc: "a ready cube's link to the create-draft form with that cube chosen (#/draft?cube=<id>)",
  }),
  draftSource: label({
    name: "What to draft",
    kind: "text",
    role: "radiogroup",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's choice between a set's packs and one of the caller's cubes; signed-in people only",
  }),
  draftSourceSet: label({
    name: "A set",
    kind: "text",
    role: "radio",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's source: packs of one draftable set",
  }),
  draftSourceCube: label({
    name: "One of my cubes",
    kind: "text",
    role: "radio",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's source: packs cut from a saved cube",
  }),
  draftCubeField: label({
    name: "Cube to draft",
    kind: "text",
    role: "combobox",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's cube picker: the caller's cubes, the ones that cannot be drafted disabled with the reason",
  }),
  draftPackSize: label({
    name: "Cards in a pack",
    kind: "text",
    role: "combobox",
    owners: [`${DRAFT}DraftCreate.svelte`],
    doc: "the create form's pack size for a cube draft, 9 to 20, 15 by default; a set's packs have their own size",
  }),
  homeDraftCard: label({
    name: "Draft a set",
    kind: "text",
    role: "heading",
    owners: ["routes/Lobby.svelte"],
    doc: "the signed-in Home's start card for a booster draft, beside Practice game and Bring a deck; opens #/draft",
  }),
  draftGone: label({
    name: "This draft is gone",
    kind: "text",
    role: "heading",
    owners: ["routes/Draft.svelte", `${DRAFT}DraftJoin.svelte`, `${DRAFT}DraftRoom.svelte`],
    doc: "a draft that closed, or a seat that is no longer the viewer's: a plain message and the way to a new draft",
  }),
  draftNewDraft: label({
    name: "Start another draft",
    kind: "text",
    role: "link",
    owners: ["routes/Draft.svelte", `${DRAFT}DraftJoin.svelte`, `${DRAFT}DraftRoom.svelte`],
    doc: "the link to the create form from a finished or closed draft",
  }),
  draftGroupWhite: label({
    name: "White",
    kind: "text",
    role: "heading",
    owners: ["lib/draft.ts"],
    doc: "a drafted pool's group of white cards",
  }),
  draftGroupBlue: label({
    name: "Blue",
    kind: "text",
    role: "heading",
    owners: ["lib/draft.ts"],
    doc: "a drafted pool's group of blue cards",
  }),
  draftGroupBlack: label({
    name: "Black",
    kind: "text",
    role: "heading",
    owners: ["lib/draft.ts"],
    doc: "a drafted pool's group of black cards",
  }),
  draftGroupRed: label({
    name: "Red",
    kind: "text",
    role: "heading",
    owners: ["lib/draft.ts"],
    doc: "a drafted pool's group of red cards",
  }),
  draftGroupGreen: label({
    name: "Green",
    kind: "text",
    role: "heading",
    owners: ["lib/draft.ts"],
    doc: "a drafted pool's group of green cards",
  }),
  draftGroupMulticolour: label({
    name: "Multicolour",
    kind: "text",
    role: "heading",
    owners: ["lib/draft.ts"],
    doc: "a drafted pool's group of cards of two or more colours",
  }),
  draftGroupColourless: label({
    name: "Colourless",
    kind: "text",
    role: "heading",
    owners: ["lib/draft.ts"],
    doc: "a drafted pool's group of colourless cards that are not lands",
  }),
  draftGroupLands: label({
    name: "Lands",
    kind: "text",
    role: "heading",
    owners: ["lib/draft.ts"],
    doc: "a drafted pool's lands, kept apart from the colours",
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
