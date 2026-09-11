// zoneBrowser.logic — pure helpers behind ZoneBrowserModal. Lives
// outside the component file so vitest can exercise the filtering
// and action-payload shapes without pulling in a Svelte/jsdom
// renderer (the client is node-only at test time). The component
// delegates all derivations here so there's no behavioural drift.

import type { CardView, ExilePlayView, GameView } from "./protocol";
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

// --- S21 sub-PR 6: impulse exile ---------------------------------
//
// A card in exile can carry a grant naming a player who may play it
// this turn — Ragavan exiles off the top of the player he hit and
// lets YOU cast it. So the affordance is keyed on the grant, not on
// zone ownership like the move buttons: the card is in the victim's
// exile slice, and the thief is the one who gets a button.

// impulseGrantFor returns the grant on `card` if the viewer is the
// one it names AND its window is open, or null.
//
// `turn` is the current turn number, for S29 warp's "you may cast it
// from exile on a LATER turn": the grant is stamped on the permanent
// the moment it is exiled, at the end step of the turn it was warped
// in, and stays dark until the next turn begins. Callers that have
// no turn number to hand pass undefined and get the pre-S29
// behaviour, which is correct for every grant that carries no floor.
export function impulseGrantFor(
  card: CardView,
  zoneKind: BrowsableZone,
  viewerID: string | null,
  turn?: number,
): ExilePlayView | null {
  if (zoneKind !== "exile" || !viewerID) return null;
  const grant = card.exile_play;
  if (!grant || grant.player !== viewerID) return null;
  if (grant.not_before_turn !== undefined && turn !== undefined && turn < grant.not_before_turn) {
    return null;
  }
  return grant;
}

// impulseActionLabel is the verb for the button, or null when there
// is no action to offer. Ragavan says "you may CAST that card", so
// a land under a cast-only grant is stranded and gets no button;
// Breeches says "you may PLAY those cards", so its land is playable.
export function impulseActionLabel(
  card: CardView,
  zoneKind: BrowsableZone,
  viewerID: string | null,
  turn?: number,
): "cast" | "play" | null {
  const grant = impulseGrantFor(card, zoneKind, viewerID, turn);
  if (!grant) return null;
  if ((card.type_line ?? "").toLowerCase().includes("land")) {
    return grant.cast_only ? null : "play";
  }
  return "cast";
}

// --- S29: alternative cast paths from non-hand zones -------------
//
// The impulse button above is keyed on a grant stamped on one exiled
// INSTANCE. Flashback and escape are the other shape: the permission
// is printed on the CARD, so the server answers it per card per zone
// and sends the answer down as `castable_here`.
//
// The client deliberately does not know what "flashback" means. It
// asks whether the card is castable from the zone it is looking at,
// and hands the cast to the Board's ordinary prompt chain, which
// reads the cost out of `alternative_costs` — already filtered
// server-side to the offers claimable from this zone.

// castableFromZone reports whether the viewer may cast `card` out of
// the zone the browser is showing.
//
// Two gates, and the ownership one is not redundant with the
// server's. `castable_here` is PUBLIC — the graveyard is a public
// zone and a flashback cost is printed on the card, so an opponent's
// snapshot carries the bit too. Without the ownership check the
// browser would offer a button on someone else's graveyard card that
// the server then refuses with ErrCardNotFound, which reads to the
// player as a bug rather than as a rule.
export function castableFromZone(
  card: CardView,
  zoneKind: BrowsableZone,
  viewerID: string | null,
  ownerID: string,
): boolean {
  if (zoneKind !== "graveyard") return false;
  if (!viewerID || viewerID !== ownerID) return false;
  return card.castable_here === true;
}
