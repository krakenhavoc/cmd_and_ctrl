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
  commandZone: dynamicLabel({
    stem: " command zone, ",
    match: "contains",
    shape: "<name> command zone, <N> card(s)",
    make: (seat: string, n: number) => `${seat} command zone, ${plural(n, "card")}`,
    example: ["Player", 1],
    kind: "aria",
    role: "group",
    owners: [`${BOARD}ExileStrip.svelte`, `${BOARD}CommandStrip.svelte`],
    doc: "a seat's command zone beside its hand (#2349): yours in the castable strip, others' in their command strip",
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
    doc: "the stack pile (ADR 0119), and the floating stack overlay",
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

  // -- The action dock (ADR 0111) --
  actions: label({
    name: "actions",
    kind: "aria",
    role: "region",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the action dock; the tutorial's move-along step and its detours anchor here",
  }),
  priorityControls: label({
    name: "priority controls",
    kind: "aria",
    role: "group",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's toggles: autopass, undo, the ⋯ menu",
  }),
  autopass: label({
    name: "autopass",
    kind: "aria",
    role: "button",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's autopass toggle (state in aria-pressed); the tutorial's let-the-bot-play step anchors here, within actions",
  }),
  next: label({
    name: "next",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's primary: pass priority",
  }),
  passTurn: label({
    name: "Pass turn",
    kind: "text",
    role: "button",
    owners: [`${BOARD}ActionDock.svelte`],
    doc: "the dock's secondary: skip the rest of the turn",
  }),
  declareAttackers: label({
    name: "declare attackers",
    kind: "aria",
    role: "dialog, group",
    owners: ["lib/combatDock.ts"],
    doc: "the attack request's dialog and its row group",
  }),
  declareBlockers: label({
    name: "declare blockers",
    kind: "aria",
    role: "dialog, group",
    owners: ["lib/combatDock.ts"],
    doc: "the block request's dialog and its row group",
  }),

  // -- Dock requests --
  mulligan: label({
    name: "keep or mulligan your hand",
    kind: "aria",
    role: "dialog",
    owners: ["routes/Game.svelte"],
    doc: "the opening hand sheet",
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
  castWithoutPaying: dynamicLabel({
    stem: " without paying its mana cost?",
    match: "suffix",
    shape: "Cast <card> without paying its mana cost?",
    make: (card: string) => `Cast ${card} without paying its mana cost?`,
    example: ["Craw Wurm"],
    kind: "aria",
    role: "dialog",
    owners: ["lib/targetingDock.ts"],
    doc: "Cast anyway's confirm (ADR 0118 §2), with its Cast and Cancel",
  }),
  cast: label({
    name: "Cast",
    kind: "text",
    role: "button",
    owners: ["lib/targetingDock.ts"],
    doc: "Cast anyway's confirm: cast it unpaid",
  }),
  cancel: label({
    name: "Cancel",
    kind: "text",
    role: "button",
    owners: ["lib/targetingDock.ts", "lib/dock.ts"],
    doc: "Cast anyway's confirm, and every flow's Cancel (the dock's cancelAction), the first-turn confirm's included",
  }),
  castAnyway: label({
    name: "Cast anyway (don't pay)",
    kind: "text",
    role: "menuitem",
    owners: ["lib/castAnyway.ts"],
    doc: "the card menu's row that casts without paying (ADR 0118 §2)",
  }),

  // -- The opening roll (ADR 0121) --
  rollForFirstTurn: label({
    name: "roll for the first turn",
    kind: "aria",
    role: "dialog",
    owners: [`${BOARD}OpeningRollDock.svelte`],
    doc: "the opening roll's request",
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
    doc: "the roll winner's sheet",
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
    doc: "the confirm when the chooser gives the first turn away, with Confirm and Cancel",
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
    doc: "the strip's banner during the roll; stands in for opening hand decisions until the winner chooses",
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
    doc: "the Lobby's create-a-table form; the lobby.create hint points at it",
  }),
  deckLink: label({
    name: "deck link",
    kind: "aria",
    role: "textbox",
    owners: ["routes/Decks.svelte"],
    doc: "the Decks page's deck-link field; the decks.check hint points at it",
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
    owners: ["lib/components/HelpMenu.svelte", "routes/Home.svelte"],
    doc: "Help, and Home's Help tile: opens a practice table; signed out it reads Practice game (sign in first) and goes to sign in",
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
  showTips: label({
    name: "Show tips",
    kind: "text",
    role: "checkbox",
    owners: ["lib/components/Settings.svelte"],
    doc: "Settings → Advanced: on unless Hide tips was pressed on a tip",
  }),
  startPractice: label({
    name: "Start practice",
    kind: "text",
    role: "link",
    owners: ["routes/LobbyPractice.hint.ts"],
    doc: "the lobby.practice tip's action, to #/practice; the tip's Got it reads Not now beside it",
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
