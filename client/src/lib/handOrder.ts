// handOrder.ts — #1524. Rearrange the cards in your own hand.
//
// Hand order has no rules meaning and opponents only ever see a count,
// so the order is the viewer's alone: it lives in this browser, in
// localStorage, one entry per game and viewer (owner decision 1). There
// is no server, protocol or snapshot change. With nothing saved — a new
// device, a private window, storage that throws — the hand shows in the
// server's order, exactly as before.
//
// The saved value is a list of instance IDs. The server hand is the
// truth about WHICH cards are in hand; the saved list only says in what
// order to show them. applyHandOrder reconciles the two quietly:
//
//   - a card in both keeps its place relative to the other known cards;
//   - a card the list does not know (a draw, or a card an undo or a
//     restore put back) goes on the right end, in server order;
//   - an ID the hand no longer holds is dropped.
//
// The three sorts (owner decision 3) are one-time: the result becomes
// the saved order and the player may adjust it by hand afterwards.
// Nothing keeps the hand sorted.

import type { CardView } from "./protocol";
import {
  isArtifact,
  isBattle,
  isCreature,
  isEnchantment,
  isLand,
  isPlaneswalker,
} from "./cardTypes";
import { manaValueOf } from "./targetPrices";

// applyHandOrder returns the server hand in display order. `saved` may
// be null (nothing saved), stale, or name cards from a different hand;
// see the file comment for the reconcile rule. Never mutates its input.
export function applyHandOrder(
  hand: readonly CardView[],
  saved: readonly string[] | null,
): CardView[] {
  if (!saved || saved.length === 0) return [...hand];
  const byID = new Map<string, CardView>();
  for (const c of hand) byID.set(c.instance_id, c);
  const out: CardView[] = [];
  const placed = new Set<string>();
  for (const id of saved) {
    const c = byID.get(id);
    if (!c || placed.has(id)) continue;
    out.push(c);
    placed.add(id);
  }
  for (const c of hand) {
    if (placed.has(c.instance_id)) continue;
    out.push(c);
    placed.add(c.instance_id);
  }
  return out;
}

// idsOf is the instance IDs of a hand, in its order.
export function idsOf(cards: readonly CardView[]): string[] {
  return cards.map((c) => c.instance_id);
}

// sameOrder: do two ID lists name the same cards in the same order?
export function sameOrder(a: readonly string[] | null, b: readonly string[] | null): boolean {
  if (a === b) return true;
  if (!a || !b || a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
  return true;
}

// moveCard takes the entry at `from` out and puts it back so it ends at
// index `to` of the result. An out-of-range `from` returns a copy
// unchanged; `to` is clamped to the list.
export function moveCard<T>(order: readonly T[], from: number, to: number): T[] {
  const out = [...order];
  if (from < 0 || from >= out.length) return out;
  const [item] = out.splice(from, 1);
  const at = Math.max(0, Math.min(out.length, to));
  out.splice(at, 0, item);
  return out;
}

// ---- sorts ---------------------------------------------------------

export type HandSort = "manaValue" | "type" | "color";

export const HAND_SORTS: readonly { key: HandSort; label: string }[] = [
  { key: "manaValue", label: "By mana value" },
  { key: "type", label: "By type (lands first)" },
  { key: "color", label: "By colour" },
];

// manaValue is the card's mana value, read off the printed mana cost the
// wire already carries ({2}{U} is 3, a land with no cost is 0, X is 0).
export function manaValue(c: CardView): number {
  return c.mana_cost ? manaValueOf(c.mana_cost) : 0;
}

// typeRank: lands, then creatures, then other permanents, then instants
// and sorceries, then anything the type line does not place. A land
// creature (Dryad Arbor) is a land here — "lands first" — and an
// artifact creature is a creature.
export function typeRank(c: CardView): number {
  if (isLand(c)) return 0;
  if (isCreature(c)) return 1;
  if (isArtifact(c) || isEnchantment(c) || isPlaneswalker(c) || isBattle(c)) return 2;
  if (c.type_line && /\b(instant|sorcery)\b/i.test(c.type_line)) return 3;
  return 4;
}

const WUBRG = ["W", "U", "B", "R", "G"];

// colorRank: mono-coloured cards in WUBRG order (0-4), then multicolour
// (5), then colourless (6). Reads the card's effective `colors`, never
// its mana cost (protocol.ts: clients must not infer colours from it).
export function colorRank(c: CardView): number {
  const colors = new Set((c.colors ?? []).filter((x) => WUBRG.includes(x)));
  if (colors.size === 0) return 6;
  if (colors.size > 1) return 5;
  return WUBRG.indexOf([...colors][0]);
}

function byName(a: CardView, b: CardView): number {
  return (a.name ?? "").localeCompare(b.name ?? "");
}

const COMPARATORS: Record<HandSort, (a: CardView, b: CardView) => number> = {
  manaValue: (a, b) => manaValue(a) - manaValue(b) || byName(a, b),
  type: (a, b) => typeRank(a) - typeRank(b) || manaValue(a) - manaValue(b) || byName(a, b),
  color: (a, b) => colorRank(a) - colorRank(b) || manaValue(a) - manaValue(b) || byName(a, b),
};

// sortHand returns the hand's IDs in the chosen order. The sort is
// stable, so full ties (two copies of one card) keep the order they had.
export function sortHand(cards: readonly CardView[], by: HandSort): string[] {
  const cmp = COMPARATORS[by];
  return cards
    .map((c, i) => ({ c, i }))
    .sort((a, b) => cmp(a.c, b.c) || a.i - b.i)
    .map((x) => x.c.instance_id);
}

// ---- persistence ---------------------------------------------------

// The key is namespaced like the client's other keys ("cmdctrl.…") and
// carries the game and the viewer, so a saved order never leaks into
// another game or another seat on a shared browser.
export const HAND_ORDER_KEY_PREFIX = "cmdctrl.handOrder.v1";

export function handOrderKey(gameID: string, viewerID: string): string {
  return `${HAND_ORDER_KEY_PREFIX}:${gameID}:${viewerID}`;
}

// Storage-like, so tests can pass a fake or one that throws. The default
// is globalThis.localStorage, read inside the try: merely touching it
// throws in some sandboxed iframes.
export interface OrderStorage {
  getItem(key: string): string | null;
  setItem(key: string, value: string): void;
}

function defaultStorage(): OrderStorage | null {
  const ls = (globalThis as { localStorage?: OrderStorage }).localStorage;
  return ls ?? null;
}

// loadHandOrder returns the saved order, or null for nothing saved,
// something unreadable, or storage that is missing or throws. Null
// means "server order".
export function loadHandOrder(
  gameID: string | null | undefined,
  viewerID: string | null | undefined,
  storage?: OrderStorage | null,
): string[] | null {
  if (!gameID || !viewerID) return null;
  try {
    const s = storage === undefined ? defaultStorage() : storage;
    if (!s) return null;
    const raw = s.getItem(handOrderKey(gameID, viewerID));
    if (!raw) return null;
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed) || !parsed.every((x) => typeof x === "string")) return null;
    return parsed as string[];
  } catch {
    return null;
  }
}

// saveHandOrder writes the order and reports whether it stuck. A quota
// error or a storage that throws is swallowed: the order still applies
// for this page, it just does not survive a reload.
export function saveHandOrder(
  gameID: string | null | undefined,
  viewerID: string | null | undefined,
  order: readonly string[],
  storage?: OrderStorage | null,
): boolean {
  if (!gameID || !viewerID) return false;
  try {
    const s = storage === undefined ? defaultStorage() : storage;
    if (!s) return false;
    s.setItem(handOrderKey(gameID, viewerID), JSON.stringify(order));
    return true;
  } catch {
    return false;
  }
}
