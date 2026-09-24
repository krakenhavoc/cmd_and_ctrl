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
