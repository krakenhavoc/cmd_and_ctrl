// zoneBrowser.logic — pure helpers behind ZoneBrowserModal. Lives
// outside the component file so vitest can exercise the filtering
// without pulling in a Svelte/jsdom renderer (the client is
// node-only at test time). The component delegates derivations
// here so there's no behavioural drift.

import type { CardView, GameView } from "./protocol";
import type { BrowsableZone } from "./zoneBrowser";

// cardsForZone returns the viewable card slice for a given zone +
// owner. Mirrors the derivations in ZoneBrowserModal.svelte. Stack
// is a shared zone so ownerID is ignored for it; graveyard / exile /
// command are all owner-scoped.
export function cardsForZone(view: GameView, zoneKind: BrowsableZone, ownerID: string): CardView[] {
  if (zoneKind === "exile") {
    return view.exile.cards.filter((c) => c.owner === ownerID);
  }
  if (zoneKind === "stack") {
    return view.stack.cards;
  }
  const seat = view.seats.find((s) => s.id === ownerID);
  if (!seat) return [];
  if (zoneKind === "graveyard") return seat.graveyard.cards;
  if (zoneKind === "command") return seat.command.cards;
  return [];
}
