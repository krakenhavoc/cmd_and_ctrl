// draftBuild.ts — the rules of the deckbuilding screen (ADR 0222 §5,
// #5440): what is in the deck and what is left in the sideboard, the
// count, the curve and the land-count suggestion.
//
// Pure functions over the viewer's own pool. The server's check is the
// one that counts (deck.ValidateLimited: at least 40 cards, every card
// that is not a basic land from the pool, any number of basic lands,
// CR 100.2b); this file says the same thing before the player submits.

import { isLand, type DraftBuild, type DraftCard, type DraftDeck } from "./draft";

/** The basic lands a deck may add, in colour order. */
export const BASIC_LANDS = ["Plains", "Island", "Swamp", "Mountain", "Forest"] as const;
export type BasicLand = (typeof BASIC_LANDS)[number];

/** The colour of mana each basic land makes. */
export const BASIC_COLOR: Record<BasicLand, string> = {
  Plains: "W",
  Island: "U",
  Swamp: "B",
  Mountain: "R",
  Forest: "G",
};

/** The most of one basic land the stepper offers; the server allows 99. */
export const MAX_BASICS = 40;

/** A deck being built: the pool cards in it, and the basics added. */
export interface BuildState {
  main: string[];
  basics: Record<string, number>;
}

/**
 * The deck the screen opens with: the one the server holds for this
 * seat (submitted earlier, or on a reload), less any id the pool does
 * not have; else an empty deck, so the player chooses every card.
 */
export function initialBuild(
  pool: readonly DraftCard[],
  deck: DraftDeck | undefined | null,
): BuildState {
  const ids = new Set(pool.map((c) => c.id));
  const main = (deck?.main ?? []).filter((id, i, all) => ids.has(id) && all.indexOf(id) === i);
  const basics: Record<string, number> = {};
  for (const name of BASIC_LANDS) {
    const n = deck?.basics?.[name] ?? 0;
    if (n > 0) basics[name] = Math.min(Math.floor(n), MAX_BASICS);
  }
  return { main, basics };
}

/** Moves a pool card into the deck, or out to the sideboard. */
export function moveCard(state: BuildState, id: string, to: "deck" | "side"): BuildState {
  const has = state.main.includes(id);
  if (to === "deck") return has ? state : { ...state, main: [...state.main, id] };
  return has ? { ...state, main: state.main.filter((x) => x !== id) } : state;
}

/** Moves a pool card to the other side: a click on it. */
export function toggleCard(state: BuildState, id: string): BuildState {
  return moveCard(state, id, state.main.includes(id) ? "side" : "deck");
}

/** Sets how many of one basic land the deck adds, held to 0..MAX_BASICS. */
export function setBasic(state: BuildState, name: BasicLand, n: number): BuildState {
  const count = Math.max(0, Math.min(MAX_BASICS, Math.floor(Number.isFinite(n) ? n : 0)));
  const basics = { ...state.basics };
  if (count === 0) delete basics[name];
  else basics[name] = count;
  return { ...state, basics };
}

/** How many basic lands the deck adds. */
export function basicsCount(state: BuildState): number {
  return BASIC_LANDS.reduce((n, name) => n + (state.basics[name] ?? 0), 0);
}

/** How many cards the deck holds. */
export function deckSize(state: BuildState): number {
  return state.main.length + basicsCount(state);
}

/** The pool split by the deck: what it plays, and the sideboard. */
export function splitPool(
  pool: readonly DraftCard[],
  state: BuildState,
): { deck: DraftCard[]; side: DraftCard[] } {
  const inDeck = new Set(state.main);
  const deck: DraftCard[] = [];
  const side: DraftCard[] = [];
  for (const c of pool) (inDeck.has(c.id) ? deck : side).push(c);
  return { deck, side };
}

/** How many lands the deck holds: the basics added, and the pool's own. */
export function landCount(deck: readonly DraftCard[], state: BuildState): number {
  return basicsCount(state) + deck.filter((c) => isLand(c)).length;
}

/** The curve's last column gathers everything this dear or dearer. */
export const CURVE_TOP = 6;

export interface CurveColumn {
  /** The mana value; CURVE_TOP stands for that and more. */
  value: number;
  label: string;
  count: number;
}

/** The deck's cards that are not lands, counted by mana value, 0 to 6+. */
export function curve(deck: readonly DraftCard[]): CurveColumn[] {
  const cols: CurveColumn[] = [];
  for (let v = 0; v <= CURVE_TOP; v++) {
    cols.push({ value: v, label: v === CURVE_TOP ? `${v}+` : String(v), count: 0 });
  }
  for (const c of deck) {
    if (isLand(c)) continue;
    cols[Math.min(Math.max(c.mana_value ?? 0, 0), CURVE_TOP)].count++;
  }
  return cols;
}

/** The usual shape of a 40-card deck: 17 lands beside 23 other cards. */
const USUAL_LANDS = 17;
const USUAL_SPELLS = 23;

export interface LandSuggestion {
  /** The lands the deck wants in all. */
  lands: number;
  /** The basics to add to get there, by land; empty with no colour to go by. */
  basics: Record<string, number>;
  /** The basics to add in all (lands, less the lands the pool gave). */
  add: number;
}

/** The coloured mana symbols of a cost, a hybrid counting for each colour. */
function colourSymbols(cost: string | undefined): string[] {
  const out: string[] = [];
  for (const m of (cost ?? "").matchAll(/\{([^}]+)\}/g)) {
    for (const part of m[1].split("/")) {
      if ("WUBRG".includes(part) && part.length === 1) out.push(part);
    }
  }
  return out;
}

/**
 * The land-count suggestion: 17 lands for 23 other cards, in proportion
 * for more or fewer, and never fewer than it takes to reach the minimum
 * deck. The basics are split by the coloured mana symbols in the deck's
 * costs (a card with no cost counts by its colours), every colour the
 * deck plays getting at least one.
 */
export function landSuggestion(deck: readonly DraftCard[], deckMin: number): LandSuggestion {
  const spells = deck.filter((c) => !isLand(c));
  const ownLands = deck.length - spells.length;
  if (spells.length === 0) return { lands: 0, basics: {}, add: 0 };
  const lands = Math.max(
    Math.round((spells.length * USUAL_LANDS) / USUAL_SPELLS),
    deckMin - spells.length,
  );
  const add = Math.max(lands - ownLands, 0);

  const weight: Record<string, number> = {};
  for (const c of spells) {
    const symbols = colourSymbols(c.mana_cost);
    for (const s of symbols.length > 0 ? symbols : (c.colors ?? [])) {
      weight[s] = (weight[s] ?? 0) + 1;
    }
  }
  const played = BASIC_LANDS.filter((name) => (weight[BASIC_COLOR[name]] ?? 0) > 0);
  const basics: Record<string, number> = {};
  if (played.length === 0 || add === 0) return { lands, basics, add };

  // Largest remainder, after one land for each colour the deck plays.
  const total = played.reduce((n, name) => n + weight[BASIC_COLOR[name]], 0);
  const floor = add >= played.length ? 1 : 0;
  const spare = add - floor * played.length;
  const shares = played.map((name) => {
    const exact = (spare * weight[BASIC_COLOR[name]]) / total;
    return { name, n: floor + Math.floor(exact), rest: exact - Math.floor(exact) };
  });
  let left = add - shares.reduce((n, s) => n + s.n, 0);
  for (const s of [...shares].sort((a, b) => b.rest - a.rest)) {
    if (left <= 0) break;
    s.n++;
    left--;
  }
  for (const s of shares) if (s.n > 0) basics[s.name] = s.n;
  return { lands, basics, add };
}

/** Whether the deck's basics are already what the suggestion says. */
export function matchesSuggestion(state: BuildState, s: LandSuggestion): boolean {
  return BASIC_LANDS.every((name) => (state.basics[name] ?? 0) === (s.basics[name] ?? 0));
}

/** Sets the deck's basics to the suggestion's. */
export function applySuggestion(state: BuildState, s: LandSuggestion): BuildState {
  return { ...state, basics: { ...s.basics } };
}

/** What stops the deck being submitted, in words; "" when nothing does. */
export function buildProblem(size: number, deckMin: number): string {
  if (size >= deckMin) return "";
  const short = deckMin - size;
  return `Your deck has ${size} ${size === 1 ? "card" : "cards"}. Add ${short} more to reach ${deckMin}.`;
}

/** Whether two decks hold the same cards and basics. */
export function sameDeck(a: BuildState, deck: DraftDeck | undefined | null): boolean {
  if (!deck) return false;
  const main = deck.main ?? [];
  if (main.length !== a.main.length) return false;
  const ids = new Set(main);
  if (!a.main.every((id) => ids.has(id))) return false;
  return BASIC_LANDS.every((name) => (a.basics[name] ?? 0) === (deck.basics?.[name] ?? 0));
}

/**
 * Whole seconds left on the build timer: the server's `deadline - now`
 * when the frame was built, less the time passed here since it arrived,
 * so a clock that is wrong on this device does not move the deadline.
 */
export function buildSecondsLeft(build: DraftBuild | undefined, elapsedMs: number): number | null {
  if (!build) return null;
  const deadline = Date.parse(build.deadline);
  const now = Date.parse(build.now);
  if (Number.isNaN(deadline) || Number.isNaN(now)) return null;
  return Math.max(0, Math.ceil((deadline - now - elapsedMs) / 1000));
}
