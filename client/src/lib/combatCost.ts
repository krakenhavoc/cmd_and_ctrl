// combatCost.ts — #6190: paying a NON-MANA cost to attack or block
// (CR 508.1h, 509.1d).
//
// Exalted Dragon "can't attack unless you sacrifice a land", Hollow
// Warrior "can't attack or block unless you tap an untapped creature you
// control", Floodtide Serpent "can't attack unless you return an
// enchantment you control to its owner's hand".
//
// The price is the server's: each creature's `attack_cost` / `block_cost`
// states the clause, how many permanents it asks for and which ones
// could pay (`CombatCostDemandView`). Nothing here prices a declaration
// or decides what may pay — it collects the player's picks and turns
// them into the three id lists the declaration carries, the same
// `sacrifice_ids` / `tap_ids` / `return_ids` a cast and an activation
// already send.
//
// The cost is NOT paid for the player. CR 508.1d and CR 509.1c: a player
// is never required to pay a cost to attack or block, so the dock asks,
// and cancelling sends nothing.

import type { CardView, CombatCostDemandView, GameView } from "./protocol";

/** One demand the dock asks about, with its candidate cards resolved. */
export interface CombatCostAsk {
  // The creature whose clause this is, for the heading.
  creature: string;
  creatureName: string;
  // The demand as the server stated it.
  demand: CombatCostDemandView;
  // The candidates, in the server's order, as cards the sheet can show.
  options: CardView[];
}

/** The whole question: every demand one declaration owes. */
export interface CombatCostPrompt {
  // The verb the confirm button carries ("Attack with 2").
  verb: string;
  asks: CombatCostAsk[];
}

/** The payload fields a paid declaration adds. */
export interface CombatCostParams {
  sacrifice_ids?: string[];
  tap_ids?: string[];
  return_ids?: string[];
}

// cardsByID indexes the battlefield once per question.
function cardsByID(view: GameView | null | undefined): Map<string, CardView> {
  const out = new Map<string, CardView>();
  for (const c of view?.battlefield?.cards ?? []) out.set(c.instance_id, c);
  return out;
}

// asksFor collects the demands of one creature, resolving candidates.
function asksFor(
  card: CardView | undefined,
  demands: CombatCostDemandView[] | undefined,
  index: Map<string, CardView>,
): CombatCostAsk[] {
  if (!card || !demands?.length) return [];
  return demands.map((demand) => ({
    creature: card.instance_id,
    creatureName: card.name,
    demand,
    options: (demand.candidates ?? []).map((id) => index.get(id)).filter((c): c is CardView => !!c),
  }));
}

/**
 * attackCostAsks is every non-mana demand a declaration of `attackers`
 * owes, one entry per (creature, clause). Empty when attacking costs no
 * permanents, which is every declaration at nearly every table.
 */
export function attackCostAsks(
  view: GameView | null | undefined,
  attackers: readonly string[],
): CombatCostAsk[] {
  const index = cardsByID(view);
  const out: CombatCostAsk[] = [];
  for (const id of attackers) {
    const card = index.get(id);
    out.push(...asksFor(card, card?.attack_cost, index));
  }
  return out;
}

/**
 * blockCostAsks is every non-mana demand the seat's STAGED blocks owe:
 * read off each blocking creature the seat controls, because that is the
 * set `finish_blocks` will be priced against (CR 509.1d).
 */
export function blockCostAsks(
  view: GameView | null | undefined,
  seatID: string | null,
): CombatCostAsk[] {
  if (!seatID) return [];
  const index = cardsByID(view);
  const out: CombatCostAsk[] = [];
  for (const card of view?.battlefield?.cards ?? []) {
    if (card.controller !== seatID || !card.blocking_target) continue;
    out.push(...asksFor(card, card.block_cost, index));
  }
  return out;
}

/**
 * combatCostPrompt is the question the sheet shows, or null when the
 * declaration owes nothing.
 */
export function combatCostPrompt(asks: CombatCostAsk[], verb: string): CombatCostPrompt | null {
  if (asks.length === 0) return null;
  return { verb, asks };
}

/**
 * combatCostUnpayable names the first demand the board cannot fill —
 * fewer candidates than the clause asks for, so the declaration cannot be
 * made at all and the dock says so instead of opening a picker that can
 * never be confirmed.
 */
export function combatCostUnpayable(asks: CombatCostAsk[]): CombatCostAsk | null {
  return asks.find((a) => a.options.length < a.demand.count) ?? null;
}

/** A selection: the ids picked for each ask, by its index. */
export type CombatCostPicks = readonly (readonly string[])[];

/** emptyPicks is one empty list per ask. */
export function emptyPicks(asks: CombatCostAsk[]): string[][] {
  return asks.map(() => []);
}

/**
 * togglePick adds or removes `id` from the ask at `index`, holding the
 * selection inside the clause's count: picking past it drops the oldest
 * pick, the way a multi-select cost picker does, so the player is never
 * stuck having to deselect first.
 */
export function togglePick(
  picks: CombatCostPicks,
  asks: CombatCostAsk[],
  index: number,
  id: string,
): string[][] {
  const out = picks.map((p) => [...p]);
  const row = out[index];
  if (!row) return out;
  const at = row.indexOf(id);
  if (at >= 0) {
    row.splice(at, 1);
    return out;
  }
  // A permanent pays ONE demand: picking it here takes it off any other.
  for (const other of out) {
    const i = other.indexOf(id);
    if (i >= 0) other.splice(i, 1);
  }
  row.push(id);
  const count = asks[index]?.demand.count ?? 1;
  while (row.length > count) row.shift();
  return out;
}

/**
 * pickedElsewhere reports whether `id` is already paying another demand,
 * so the sheet can show it as spoken for rather than letting it pay
 * twice (the server refuses a permanent named twice).
 */
export function pickedElsewhere(picks: CombatCostPicks, index: number, id: string): boolean {
  return picks.some((row, i) => i !== index && row.includes(id));
}

/**
 * canConfirmCombatCost is true once every demand has exactly as many
 * permanents as its clause asks for. A cost is not partially payable
 * (CR 118.3), so the confirm stays disabled until then.
 */
export function canConfirmCombatCost(picks: CombatCostPicks, asks: CombatCostAsk[]): boolean {
  if (asks.length === 0) return false;
  return asks.every((a, i) => (picks[i]?.length ?? 0) === a.demand.count);
}

/**
 * chooseCombatCostForMe fills every demand with the first candidates the
 * server offered, in its order, skipping permanents already spoken for.
 * It never confirms: the player can change the picks.
 */
export function chooseCombatCostForMe(asks: CombatCostAsk[]): string[][] {
  const taken = new Set<string>();
  return asks.map((a) => {
    const row: string[] = [];
    for (const c of a.options) {
      if (row.length >= a.demand.count) break;
      if (taken.has(c.instance_id)) continue;
      taken.add(c.instance_id);
      row.push(c.instance_id);
    }
    return row;
  });
}

/**
 * combatCostParams is the picks as the declaration's payload fields: one
 * id list per cost component, each omitted when empty. The ORDER within a
 * list does not matter — the server matches the permanents to the
 * demands.
 */
export function combatCostParams(picks: CombatCostPicks, asks: CombatCostAsk[]): CombatCostParams {
  const out: CombatCostParams = {};
  const push = (key: keyof CombatCostParams, ids: readonly string[]): void => {
    if (ids.length === 0) return;
    out[key] = [...(out[key] ?? []), ...ids];
  };
  asks.forEach((a, i) => {
    const ids = picks[i] ?? [];
    switch (a.demand.kind) {
      case "tap":
        push("tap_ids", ids);
        break;
      case "return":
        push("return_ids", ids);
        break;
      default:
        push("sacrifice_ids", ids);
    }
  });
  return out;
}

/** combatCostVerb is the sheet's sentence for one demand's clause. */
export function combatCostVerb(demand: CombatCostDemandView): string {
  const clause = demand.clause || "a permanent you control";
  switch (demand.kind) {
    case "tap":
      return `Tap ${clause}`;
    case "return":
      return `Return ${clause} to its owner's hand`;
    default:
      return `Sacrifice ${clause}`;
  }
}

/**
 * combatCostNote is the dock's one-line summary of what a declaration
 * will cost in permanents: "sacrificing a land", "tapping an untapped
 * creature you control and sacrificing a land". "" when it costs none.
 */
export function combatCostNote(asks: CombatCostAsk[]): string {
  const parts: string[] = [];
  for (const a of asks) {
    const clause = a.demand.clause || "a permanent you control";
    const s =
      a.demand.kind === "tap"
        ? `tapping ${clause}`
        : a.demand.kind === "return"
          ? `returning ${clause} to its owner's hand`
          : `sacrificing ${clause}`;
    if (!parts.includes(s)) parts.push(s);
  }
  return parts.join(" and ");
}
