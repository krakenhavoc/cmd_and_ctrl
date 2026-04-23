// zoneBrowser.logic — pure helpers behind ZoneBrowserModal. Lives
// outside the component file so vitest can exercise the filtering
// and action-payload shapes without pulling in a Svelte/jsdom
// renderer (the client is node-only at test time). The component
// delegates all derivations here so there's no behavioural drift.

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

// canManageZone tells the UI whether the viewer may invoke owner
// actions on cards in this zone. Stack is view-only (cards on the
// stack have their own counter/resolve flow via StackOverlay).
// Graveyard / exile / command are manage-able only by their owner,
// who is also the controller of any card in those zones server-side.
export function canManageZone(
  zoneKind: BrowsableZone,
  viewerID: string | null,
  ownerID: string,
): boolean {
  if (zoneKind === "stack") return false;
  if (!viewerID) return false;
  return viewerID === ownerID;
}

// buildMovePayload constructs the params object for a move_card
// action dispatched from the zone browser. Separated so tests can
// pin the wire shape without rendering a component.
//
// Destination semantics:
// - hand / library: dst.owner = viewer (the card moves to the
//   viewer's own private zone; they're also the owner of the
//   source zone so no cross-player transfer happens).
// - battlefield: shared zone, no owner stamp.
export interface ZoneRef {
  kind: string;
  owner?: string;
}
export interface MovePayload {
  src: ZoneRef;
  dst: ZoneRef;
  instance_id: string;
}
export function buildMovePayload(
  zoneKind: BrowsableZone,
  ownerID: string,
  viewerID: string,
  card: Pick<CardView, "instance_id" | "controller">,
  dest: "hand" | "battlefield" | "library",
): MovePayload | null {
  if (!canManageZone(zoneKind, viewerID, ownerID)) return null;
  // Defence in depth against a stale controller: the server gates
  // on the card's current Controller, and the browser mustn't
  // speculatively fire a move the server will reject.
  if (card.controller && card.controller !== viewerID) return null;
  const src: ZoneRef = { kind: zoneKind, owner: ownerID };
  const dst: ZoneRef = { kind: dest };
  if (dest === "hand" || dest === "library") dst.owner = viewerID;
  return { src, dst, instance_id: card.instance_id };
}
