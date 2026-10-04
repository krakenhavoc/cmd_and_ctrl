// manaSourcePicker — module-scope store behind the anchored mana
// picker a multi-ability source opens on left-click (#1438).
//
// Same shape as contextMenu.ts, for the same reason: the click lands
// in PlayerPanel, several components below Board, and the popover has
// to be position: fixed OUTSIDE every card — a Card carries a CSS
// transform (the tap rotation, the hover lift), and a transformed
// ancestor turns position: fixed into position: absolute. So the
// click writes here and Board mounts the one picker.

import { get, type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import type { AnchorRect } from "./manaSource";

export interface ManaSourcePickerOpen {
  /** The source permanent; the picker re-reads it from each snapshot. */
  cardID: string;
  /** The card's on-screen box when it was clicked. */
  anchor: AnchorRect;
  /**
   * ADR 0117 §4: open on this one mana ability only. A lone ability
   * with two or more picking slots opens straight on its stepper; the
   * popover's mana row with a colour choice opens its buttons or its
   * stepper. Absent lists every mana ability the card has.
   */
  abilityIndex?: number;
}

export const manaSourcePicker: Writable<ManaSourcePickerOpen | null> = guardedWritable(
  null,
  "manaSourcePicker",
);

export function openManaSourcePicker(open: ManaSourcePickerOpen): void {
  manaSourcePicker.set(open);
}

export function closeManaSourcePicker(): void {
  manaSourcePicker.set(null);
}

/** Whether the picker is open on this card right now. */
export function manaSourcePickerOpenFor(cardID: string): boolean {
  return get(manaSourcePicker)?.cardID === cardID;
}
