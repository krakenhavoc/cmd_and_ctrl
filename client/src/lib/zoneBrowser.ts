// zoneBrowser — tiny module-scope store for the S18.5 zone-browser
// modal. Pile chips in PileBar / CommandStrip call openZoneBrowser()
// with the zone kind + owner; Board.svelte renders ZoneBrowserModal
// bound to the store. Lifting the open/close state to module scope
// (rather than prop-drilling a handler through every PlayerPanel)
// keeps PileButton's surface unchanged — it still speaks only a bare
// onClick — and lets any future consumer (e.g. the stack overlay's
// "show all on stack" affordance) open the same modal with one call.
//
// Only one modal is open at a time. Opening a second one replaces
// the first; closing writes null.

import { type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";

// "sideboard" is the viewer's OWN cards outside the game (ADR 0191, amended
// 2026-10-19): read only, a tab of the zone browser and no pile on the table,
// never offered for another seat. A card in it moves only through a live
// permission's cast button.
export type BrowsableZone = "graveyard" | "exile" | "command" | "stack" | "sideboard";

export interface ZoneBrowserOpen {
  zoneKind: BrowsableZone;
  ownerID: string;
  // ownerName is cached at open time so we don't need to re-derive it
  // from the game snapshot inside the modal component (the modal
  // already has the GameView to filter cards against, but duplicating
  // the seat lookup there would be redundant).
  ownerName: string;
}

export const zoneBrowser: Writable<ZoneBrowserOpen | null> = guardedWritable(null, "zoneBrowser");

export function openZoneBrowser(open: ZoneBrowserOpen): void {
  zoneBrowser.set(open);
}

export function closeZoneBrowser(): void {
  zoneBrowser.set(null);
}
