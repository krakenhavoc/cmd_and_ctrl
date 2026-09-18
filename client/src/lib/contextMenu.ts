// contextMenu — module-scope store behind the admin card context
// menu (#170). Card.svelte's oncontextmenu handler calls
// openCardMenu() with the clicked card and the cursor position;
// Board.svelte mounts one CardContextMenu bound to this store.
//
// Same shape (and same reasoning) as zoneBrowser.ts: lifting the
// open/closed state to module scope keeps every Card render site —
// BattlefieldRow, Hand, CommandZone, ZoneBrowserModal — free of new
// props, which matters because the menu has to be reachable from
// every zone a card can sit in. The menu itself resolves the card's
// zone from the snapshot (see locateCard in contextMenu.logic.ts),
// so the opener doesn't have to know where the card lives either.
//
// Only one menu is open at a time. Opening a second one replaces the
// first; closing writes null.

import { type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import type { CardView } from "./protocol";

export interface CardMenuOpen {
  card: CardView;
  // Viewport coordinates of the cursor at right-click time. The menu
  // is position: fixed, so these are used verbatim (clamped to the
  // viewport by the component once it knows its own size).
  x: number;
  y: number;
}

export const cardMenu: Writable<CardMenuOpen | null> = guardedWritable(null, "cardMenu");

export function openCardMenu(open: CardMenuOpen): void {
  cardMenu.set(open);
}

export function closeCardMenu(): void {
  cardMenu.set(null);
}
