// draftDeckStats.ts — what Main is made of, counted for the colours and
// curve strip and for the full stats overlay (#5848 §6). Pure functions
// over the viewer's own cards; the components only draw them.
//
// The overlay's "More detail" lists, for each row, the names of the cards
// it counts, so every count here is a list of cards, not a number: the
// number is the list's length. A row with no cards is still a row (the
// overlay fades it), because what a deck lacks is an answer too.
//
// Roles ("what they do") prefer what the server already sends: the card's
// printed `keywords` and its type line. Rules text is matched only where
// nothing else says it (removal, card draw, a counterspell…), with
// patterns that would sooner miss a card than claim one it isn't. A
// missed card is a count of one too few; a wrong one is a lie.

import { isLand, type DraftCard } from "./draft";
import { curve, type CurveColumn } from "./draftBuild";
import { cardTypes, STAT_COLOURS, type CardType, type StatColour } from "./draftStats";

/** A row of the overlay: a name, and the cards it counts. */
export interface StatRow {
  key: string;
  cards: DraftCard[];
}

/** The cards' names, a card that appears twice once, as "Name ×2". */
export function namesOf(cards: readonly DraftCard[]): string[] {
  const order: string[] = [];
  const n = new Map<string, number>();
  for (const c of cards) {
    if (!n.has(c.name)) order.push(c.name);
    n.set(c.name, (n.get(c.name) ?? 0) + 1);
  }
  return order.map((name) => (n.get(name)! > 1 ? `${name} ×${n.get(name)}` : name));
}

// ---- Colours ----

/**
 * The coloured mana symbols in a cost, per colour: "{1}{W}{W}" is two
 * white. A hybrid symbol ("{W/U}", "{2/G}", "{G/P}") counts once for each
 * colour in it.
 */
export function costPips(cost: string | undefined): Record<StatColour, number> {
  const pips = Object.fromEntries(STAT_COLOURS.map((c) => [c, 0])) as Record<StatColour, number>;
  for (const m of (cost ?? "").matchAll(/\{([^}]*)\}/g)) {
    for (const c of STAT_COLOURS) if (m[1].split("/").includes(c)) pips[c]++;
  }
  return pips;
}

export interface ColourRow {
  colour: StatColour;
  /** Coloured symbols in the costs of Main's cards. */
  pips: number;
  /** Cards of this colour in Main (a card of two colours is in both). */
  cards: number;
}

export interface ColourStats {
  /** The colours with at least one pip, most pips first (ties in WUBRG order). */
  rows: ColourRow[];
  totalPips: number;
  /** The basics to play for a 17-land deck by pips, as [colour, count], most first. */
  basics: [StatColour, number][];
}

/** The basic land of each colour, for "9 Swamp · 8 Plains". */
export const BASIC_OF: Record<StatColour, string> = {
  W: "Plains",
  U: "Island",
  B: "Swamp",
  R: "Mountain",
  G: "Forest",
};

/**
 * Splits `total` among the colours in proportion to their pips by the
 * largest-remainder method: each gets its whole share, and the lands
 * still unplaced go to the biggest fractions (ties to the colour with
 * more pips, then in WUBRG order). A colour whose share rounds to nothing
 * is left out.
 */
export function splitBasics(
  pips: Partial<Record<StatColour, number>>,
  total = 17,
): [StatColour, number][] {
  const have = STAT_COLOURS.filter((c) => (pips[c] ?? 0) > 0);
  const sum = have.reduce((n, c) => n + (pips[c] ?? 0), 0);
  if (sum === 0) return [];
  const shares = have.map((c) => {
    const exact = ((pips[c] ?? 0) / sum) * total;
    return { c, n: Math.floor(exact), frac: exact - Math.floor(exact), pips: pips[c] ?? 0 };
  });
  let left = total - shares.reduce((n, s) => n + s.n, 0);
  const byFraction = [...shares].sort((a, b) => b.frac - a.frac || b.pips - a.pips);
  for (const s of byFraction) {
    if (left <= 0) break;
    s.n++;
    left--;
  }
  return shares
    .filter((s) => s.n > 0)
    .sort((a, b) => b.pips - a.pips || STAT_COLOURS.indexOf(a.c) - STAT_COLOURS.indexOf(b.c))
    .map((s) => [s.c, s.n]);
}

/** The colours of Main: pips, cards, and the basics split. */
export function colourStats(main: readonly DraftCard[]): ColourStats {
  const pips = Object.fromEntries(STAT_COLOURS.map((c) => [c, 0])) as Record<StatColour, number>;
  const cards = Object.fromEntries(STAT_COLOURS.map((c) => [c, 0])) as Record<StatColour, number>;
  for (const card of main) {
    const own = costPips(card.mana_cost);
    for (const c of STAT_COLOURS) {
      pips[c] += own[c];
      if ((card.colors ?? []).includes(c)) cards[c]++;
    }
  }
  const rows = STAT_COLOURS.filter((c) => pips[c] > 0)
    .map((colour) => ({ colour, pips: pips[colour], cards: cards[colour] }))
    .sort((a, b) => b.pips - a.pips);
  return {
    rows,
    totalPips: rows.reduce((n, r) => n + r.pips, 0),
    basics: splitBasics(pips),
  };
}

// ---- Curve ----

export interface CurveStats {
  columns: CurveColumn[];
  /** Cards that are not lands. */
  spells: number;
  /** Their average mana value; 0 with none. */
  average: number;
}

/** The curve of Main (lands left out) and its average mana value. */
export function curveStats(main: readonly DraftCard[]): CurveStats {
  const spells = main.filter((c) => !isLand(c));
  const total = spells.reduce((n, c) => n + (c.mana_value ?? 0), 0);
  return {
    columns: curve(main),
    spells: spells.length,
    average: spells.length === 0 ? 0 : total / spells.length,
  };
}

// ---- Card types ----

export interface TypeStats {
  /** One row per type, in CARD_TYPES order; a card with two types is in two rows. */
  rows: { type: CardType; cards: DraftCard[] }[];
  /** Cards that are neither creatures nor lands. */
  otherSpells: DraftCard[];
  creatures: number;
}

export function typeStats(main: readonly DraftCard[]): TypeStats {
  const order: CardType[] = [
    "creature",
    "instant",
    "sorcery",
    "enchantment",
    "artifact",
    "planeswalker",
    "land",
  ];
  const rows = order.map((type) => ({ type, cards: [] as DraftCard[] }));
  const otherSpells: DraftCard[] = [];
  for (const card of main) {
    const own = cardTypes(card);
    for (const r of rows) if (own.includes(r.type)) r.cards.push(card);
    if (!own.includes("creature") && !own.includes("land")) otherSpells.push(card);
  }
  return { rows, otherSpells, creatures: rows[0].cards.length };
}

/** Below this many creatures the overlay says limited decks want more. */
export const CREATURES_WANTED = 13;

// ---- Type lines and creature types ----

/** The front face's type line: "Artifact Creature — Golem" → ["Artifact Creature", "Golem"]. */
function splitTypeLine(card: Pick<DraftCard, "type_line">): [string, string] {
  const front = (card.type_line ?? "").split("//")[0].trim();
  const at = front.search(/—|\s-\s/);
  if (at < 0) return [front, ""];
  return [
    front.slice(0, at).trim(),
    front
      .slice(at)
      .replace(/^(—|\s-\s)/, "")
      .trim(),
  ];
}

/** "Artifact Creature", "Enchantment (Aura)", "Artifact (Equipment)", "Land". */
export function typeLineOf(card: Pick<DraftCard, "type_line">): string {
  const [front, sub] = splitTypeLine(card);
  if (!front) return "";
  // A creature's subtypes are its race and class, which have their own rows.
  return /\bCreature\b/.test(front) || !sub ? front : `${front} (${sub})`;
}

function byCountThenName(a: StatRow, b: StatRow): number {
  return b.cards.length - a.cards.length || a.key.localeCompare(b.key);
}

function groupBy(cards: readonly DraftCard[], keys: (c: DraftCard) => string[]): StatRow[] {
  const map = new Map<string, DraftCard[]>();
  for (const card of cards) {
    for (const key of new Set(keys(card))) {
      if (!key) continue;
      map.set(key, [...(map.get(key) ?? []), card]);
    }
  }
  return [...map].map(([key, list]) => ({ key, cards: list })).sort(byCountThenName);
}

/** Every type line in Main, most common first. */
export function typeLines(main: readonly DraftCard[]): StatRow[] {
  return groupBy(main, (c) => [typeLineOf(c)]);
}

/** The most common creature types of Main's creatures (a Changeling's "Changeling" included, as printed). */
export function creatureTypes(main: readonly DraftCard[], top = 8): StatRow[] {
  return groupBy(main, (c) => {
    const [front, sub] = splitTypeLine(c);
    return /\bCreature\b/.test(front) ? sub.split(/\s+/).filter(Boolean) : [];
  }).slice(0, top);
}

// ---- Creatures ----

/** A printed power or toughness that is a plain number; "*" and "1+*" are not. */
function plain(v: string | undefined): number | null {
  return v !== undefined && /^-?\d+$/.test(v.trim()) ? Number(v) : null;
}

export interface CreatureStats {
  creatures: DraftCard[];
  /** Average power and toughness over the creatures that print both as numbers; null with none. */
  avgPower: number | null;
  avgToughness: number | null;
  small: DraftCard[];
  medium: DraftCard[];
  big: DraftCard[];
  evasive: DraftCard[];
}

/** The keywords that make a creature hard to block. */
const EVASION = ["flying", "trample", "menace"];

const isCreature = (c: DraftCard): boolean => cardTypes(c).includes("creature");

const hasKeyword = (c: DraftCard, word: string): boolean =>
  (c.keywords ?? []).some((k) => k.toLowerCase() === word);

/** Flying, trample, menace, or "can't be blocked" (not "can't block"). */
export function isEvasive(c: DraftCard): boolean {
  return (
    EVASION.some((w) => hasKeyword(c, w)) || /\bcan['’]t be blocked\b/i.test(c.oracle_text ?? "")
  );
}

export function creatureStats(main: readonly DraftCard[]): CreatureStats {
  const creatures = main.filter(isCreature);
  const sized = creatures.filter((c) => plain(c.power) !== null && plain(c.toughness) !== null);
  const mean = (f: (c: DraftCard) => number | null): number | null =>
    sized.length === 0 ? null : sized.reduce((n, c) => n + (f(c) ?? 0), 0) / sized.length;
  const power = (c: DraftCard): number => plain(c.power) ?? 0;
  return {
    creatures,
    avgPower: mean((c) => plain(c.power)),
    avgToughness: mean((c) => plain(c.toughness)),
    small: sized.filter((c) => power(c) <= 2),
    medium: sized.filter((c) => power(c) >= 3 && power(c) <= 4),
    big: sized.filter((c) => power(c) >= 5),
    evasive: creatures.filter(isEvasive),
  };
}

// ---- What the cards do ----

export const ROLES = [
  "removal",
  "fliers",
  "draw",
  "tricks",
  "counters",
  "bounce",
  "tokens",
  "ramp",
  "lifegain",
  "discard",
  "answers",
] as const;
export type Role = (typeof ROLES)[number];

const text = (c: DraftCard): string => c.oracle_text ?? "";
const isInstantSpeed = (c: DraftCard): boolean =>
  cardTypes(c).includes("instant") || hasKeyword(c, "flash");

/**
 * Each role's test. Keywords and type lines first; rules text only where
 * the printed keywords say nothing, and then in the phrasing Oracle uses.
 */
const ROLE_TEST: Record<Role, (c: DraftCard) => boolean> = {
  removal: (c) =>
    /\b(destroy|exile) target [^.\n]*\bcreature\b/i.test(text(c)) ||
    /\bdeals? \d+ damage to (any target|target creature)\b/i.test(text(c)) ||
    /\bdeals? \d+ damage divided as you choose\b/i.test(text(c)) ||
    /\btarget creature gets -\d+\/-\d+\b/i.test(text(c)) ||
    /\benchanted creature can['’]t attack or block\b/i.test(text(c)),
  fliers: (c) => isCreature(c) && hasKeyword(c, "flying"),
  draw: (c) => /\bdraw (a|an|two|three|four|\d+) cards?\b/i.test(text(c)),
  tricks: (c) =>
    isInstantSpeed(c) &&
    !cardTypes(c).includes("land") &&
    /\btarget creature (you control )?gets \+\d+\/\+\d+ until end of turn\b/i.test(text(c)),
  counters: (c) => /\bcounter target [^.\n]*\bspell\b/i.test(text(c)),
  bounce: (c) =>
    /\breturn target [^.\n]*\b(creature|permanent)\b[^.\n]* to its owner['’]s hand\b/i.test(
      text(c),
    ),
  tokens: (c) => /\bcreates? [^.\n]*\btokens?\b/i.test(text(c)),
  ramp: (c) =>
    !isLand(c) &&
    (/\badd (\{[^}]+\})+/i.test(text(c)) ||
      /\bsearch your library for [^.\n]*\bbasic land card\b/i.test(text(c))),
  lifegain: (c) => hasKeyword(c, "lifelink") || /\bgains? (\d+|x|that much) life\b/i.test(text(c)),
  discard: (c) =>
    /\b(target (player|opponent)|each (opponent|player)|defending player) discards?\b/i.test(
      text(c),
    ),
  answers: (c) => /\b(destroy|exile) target [^.\n]*\b(artifact|enchantment)\b/i.test(text(c)),
};

/** The roles a card plays, in ROLES order; lands have none. */
export function cardRoles(card: DraftCard): Role[] {
  if (isLand(card)) return [];
  return ROLES.filter((r) => ROLE_TEST[r](card));
}

// ---- Everything the overlay shows ----

export interface DeckStats {
  /** Cards in Main. */
  total: number;
  colours: ColourStats;
  types: TypeStats;
  curve: CurveStats;
  typeLines: StatRow[];
  creatureTypes: StatRow[];
  creatures: CreatureStats;
  roles: { role: Role; cards: DraftCard[] }[];
  /** Instants and cards with flash. */
  instant: DraftCard[];
  /** Every other spell. */
  sorcery: DraftCard[];
  multicolour: DraftCard[];
  colourless: DraftCard[];
}

export function deckStats(main: readonly DraftCard[]): DeckStats {
  const spells = main.filter((c) => !isLand(c));
  const roles = main.map((c) => cardRoles(c));
  const instant = spells.filter(isInstantSpeed);
  return {
    total: main.length,
    colours: colourStats(main),
    types: typeStats(main),
    curve: curveStats(main),
    typeLines: typeLines(main),
    creatureTypes: creatureTypes(main),
    creatures: creatureStats(main),
    roles: ROLES.map((role) => ({
      role,
      cards: main.filter((_, i) => roles[i].includes(role)),
    })),
    instant,
    sorcery: spells.filter((c) => !instant.includes(c)),
    multicolour: spells.filter((c) => (c.colors ?? []).length > 1),
    colourless: spells.filter((c) => (c.colors ?? []).length === 0),
  };
}
