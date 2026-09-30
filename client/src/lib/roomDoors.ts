/**
 * roomDoors.ts — the client half of a Room's doors (CR 709.5, ADR 0103).
 *
 * A Room on the battlefield carries `doors` (which of its two halves
 * are unlocked) and, for each locked door, an `unlock` row in
 * `special_actions` naming the door, its price and the server's own
 * timing answer. This module turns the two into the door strip's rows,
 * and carries an unlock request from a door button to the Board — the
 * one place that sends actions — through a module-scope store, the
 * same way the card context menu does (contextMenu.ts), so Card.svelte
 * needs no new callback threaded through every row.
 */

import { type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import type { CardView, SpecialActionView } from "./protocol";

export type DoorSide = "left" | "right";

/** One half of a Room as the door strip draws it. */
export interface DoorRow {
  door: DoorSide;
  /** The half's printed name, from `faces`. */
  name: string;
  /** The half's printed mana cost — its unlock cost. */
  cost: string;
  unlocked: boolean;
  /**
   * The unlock row the server published for this door, when it is
   * locked. Absent for an unlocked door.
   */
  unlock?: SpecialActionView;
}

/**
 * doorRows is the strip for `card`, or an empty list for anything that
 * is not a Room on the battlefield (no `doors` on the wire).
 */
export function doorRows(card: CardView): DoorRow[] {
  const doors = card.doors;
  const faces = card.faces ?? [];
  if (!doors || faces.length < 2) return [];
  return (["left", "right"] as const).map((door, i) => {
    const unlocked = doors[door];
    return {
      door,
      name: faces[i]?.name ?? "",
      cost: faces[i]?.mana_cost ?? "",
      unlocked,
      unlock: unlocked
        ? undefined
        : (card.special_actions ?? []).find((sa) => sa.kind === "unlock" && sa.door === door),
    };
  });
}

/**
 * unlockPrice is the price a door button shows: the charged cost when
 * the server priced one (a discount is real), the printed one
 * otherwise, and "{0}" for a door that costs nothing.
 */
export function unlockPrice(row: DoorRow): string {
  const charged = row.unlock?.charged_cost;
  const price = charged !== undefined ? charged : (row.unlock?.cost ?? row.cost);
  return price === "" ? "{0}" : price;
}

/** unlockParams is the special_action payload that unlocks `door`. */
export function unlockParams(cardID: string, door: DoorSide): Record<string, unknown> {
  return { card_id: cardID, kind: "unlock", door, strict: true, auto_tap: true };
}

/** A door button's request, read and cleared by the Board. */
export interface UnlockRequest {
  cardID: string;
  door: DoorSide;
}

export const unlockRequest: Writable<UnlockRequest | null> = guardedWritable(null, "unlockRequest");

export function requestUnlock(cardID: string, door: DoorSide): void {
  unlockRequest.set({ cardID, door });
}
