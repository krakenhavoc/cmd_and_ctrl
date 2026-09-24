// zoneBrowser.logic — pure helpers behind ZoneBrowserModal. Lives
// outside the component file so vitest can exercise the filtering
// and action-payload shapes without pulling in a Svelte/jsdom
// renderer (the client is node-only at test time). The component
// delegates all derivations here so there's no behavioural drift.

import { cardAsFace, castableFaces } from "./faces";
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
// `turn` is the current turn sequence, for S29 warp's "you may cast it
// from exile on a LATER turn": the grant is stamped on the permanent
// the moment it is exiled, at the end step of the turn it was warped
// in, and stays dark until the next turn begins. Callers that have
// no turn sequence to hand pass undefined and get the pre-S29
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
  if (grant.not_before_seq !== undefined && turn !== undefined && turn < grant.not_before_seq) {
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
  // The land test reads the GRANTED face, not the face the card is
  // sitting in exile wearing (S32). A defeated Siege is exiled
  // battle-side-up under a grant for its back face, so asking the
  // front face whether it is a land answers a question about the
  // wrong card. Every grant that names no face resolves to the card
  // itself, which is the pre-S32 behaviour to the byte.
  const played = grantedFace(card, grant);
  if ((played.type_line ?? "").toLowerCase().includes("land")) {
    return grant.cast_only ? null : "play";
  }
  return "cast";
}

// grantedFaceIndex is the ONE face a grant opens, or undefined when
// it speaks about no faces (every impulse, airbend, warp and cascade
// grant) or about several (a choice the caster has not made, which
// nothing declares today).
//
// A list on the wire rather than a number, because zero had to mean
// two things and could not: a defeated Siege's grant names the BACK
// face and CR 715.4's Adventure grant names the CREATURE face, which
// is face 0. `if (!grant.face)` read the second as "no opinion" and
// re-opened the face picker on a cast with exactly one legal face.
export function grantedFaceIndex(grant: ExilePlayView | null): number | undefined {
  if (!grant || grant.faces?.length !== 1) return undefined;
  return grant.faces[0];
}

// grantedFace returns the view of `card` that the grant actually
// plays: `faces[i]` materialised over the card when the grant names
// one face, and the card unchanged when it does not.
//
// Exported because the button label and its aria-label both have to
// name the half being cast — "cast Refraction Elemental from exile",
// not "cast Invasion of Karsus from exile", which would name a card
// the click cannot produce.
export function grantedFace(card: CardView, grant: ExilePlayView | null): CardView {
  const i = grantedFaceIndex(grant);
  // A grant naming the face that is ALREADY up returns the card
  // untouched, which since #992 is an economy rather than a
  // correctness rule: cardAsFace swaps a face's announce block in
  // instead of clearing the card's, and for the face that is up the
  // two blocks are the same answer. CR 715.4's Adventure grant is
  // this case — face 0, on a card exile is already showing front-up
  // (CR 712.8).
  if (i === undefined || i === (card.active_face ?? 0)) return card;
  return cardAsFace(card, i);
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
// ONE gate, because `castable_here` is the viewer's own answer since
// #1055. It used to be the PILE OWNER's and it used to be public, so
// this function needed a second gate — "you own the pile, or
// `exile_play` names you" — to stop the browser offering a button on
// someone else's graveyard card that the server then refused with
// ErrCardNotFound, which reads to the player as a bug rather than as a
// rule. That pair-read was the bug #1055 fixed at the source: the
// field's name is a statement about the viewer, and now so is its
// value.
//
// A permission is still a statement about an OBJECT, not about a pile
// — Wrexial's "cast target instant or sorcery card from that player's
// graveyard" — so a card in an opponent's pile IS castable by its
// holder. The server computes that holder's own offers, targets and
// gate and stamps them, the bit included, on their frame and nobody
// else's. Reading the bit alone is reading exactly that.
//
// `exile_play` keeps its own job: it is public, it names the seat that
// granted the permission, and the impulse button above reads it for
// the grant's face and label. It is no longer part of ANSWERING
// whether this viewer may cast.
//
// #1173: the answer is the UNION over the card's block and its
// castable `faces[i]` blocks, not the card's alone. #1171 made the
// server ask "does this card's own text open this zone" of every
// castable face rather than of the card's bare (face 0) oracle ID, so
// a card whose BACK face prints flashback, escape or a Gravecrawler-
// shaped permission is stamped with `castable_here` on `faces[1]` and
// `false` on the card — casting the FRONT half out of the graveyard is
// genuinely not legal. Reading the card alone answered face 0's
// question for every face; `castableFaces` (faces.ts) is the shared
// walk that also backs canCastFromHand's face-aware gates (#1168), so
// the two readers can't drift into different opinions about which
// faces a cast may choose.
export function castableFromZone(card: CardView, zoneKind: BrowsableZone): boolean {
  if (zoneKind !== "graveyard") return false;
  return castableFaces(card).some((f) => f.castable_here === true);
}
