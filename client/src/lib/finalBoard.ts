// finalBoard.ts — #4912: an eliminated player's final board, what they
// had when they left the game.
//
// CR 800.4a takes every object a player owns out of the game as they
// leave, so their panel is empty under the tombstone from then on. The
// server records their zones first and sends the record on the seat
// (`PlayerView.final_board`), the same for every viewer and on every frame,
// so a reload or a late spectator sees it too. The tombstone's Final board
// button opens a read-only viewer of it (FinalBoardModal.svelte), mounted
// once by Board.svelte from the store below, like the zone browser.
//
// The hand is shown to everyone: the record exists only once the player
// has left, and the owner decided (2026-11-04) the table may see it then.
// The library is never in it.

import type { Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import type { CardView, FinalBoardView, FinalCardView, PlayerView } from "./protocol";

export type FinalZone = "battlefield" | "hand" | "graveyard" | "exile" | "command";

/** The viewer's tabs, in the order a seat's zones read. */
export const FINAL_ZONES: readonly { kind: FinalZone; label: string }[] = [
  { kind: "battlefield", label: "Battlefield" },
  { kind: "hand", label: "Hand" },
  { kind: "graveyard", label: "Graveyard" },
  { kind: "exile", label: "Exile" },
  { kind: "command", label: "Command zone" },
];

/** The open viewer: whose final board, by seat ID. */
export interface FinalBoardOpen {
  seatID: string;
}

export const finalBoardViewer: Writable<FinalBoardOpen | null> = guardedWritable(
  null,
  "finalBoardViewer",
);

export function openFinalBoard(seatID: string): void {
  finalBoardViewer.set({ seatID });
}

export function closeFinalBoard(): void {
  finalBoardViewer.set(null);
}

/** Whether a seat has a final board to show: it has left, and the record came with it. */
export function hasFinalBoard(seat: PlayerView | null | undefined): boolean {
  return seat?.eliminated === true && seat.final_board != null;
}

/** The cards of one zone of a final board. */
export function finalZoneCards(
  fb: FinalBoardView | null | undefined,
  zone: FinalZone,
): FinalCardView[] {
  return fb?.[zone] ?? [];
}

/** The total number of cards on a final board. */
export function finalBoardCount(fb: FinalBoardView | null | undefined): number {
  return FINAL_ZONES.reduce((n, z) => n + finalZoneCards(fb, z.kind).length, 0);
}

/** The tab the viewer opens on: the battlefield, or the first zone with a card in it. */
export function firstFinalZone(fb: FinalBoardView | null | undefined): FinalZone {
  return FINAL_ZONES.find((z) => finalZoneCards(fb, z.kind).length > 0)?.kind ?? "battlefield";
}

/**
 * The battlefield in reading order: each permanent, with whatever was
 * attached to it straight after it. An attachment whose host is not on
 * this board (an Aura on another player's creature, a Curse on a player)
 * keeps its place.
 */
export function finalBattlefieldOrder(cards: readonly FinalCardView[]): FinalCardView[] {
  const here = new Set(cards.map((c) => c.instance_id));
  const onHost = new Map<string, FinalCardView[]>();
  for (const c of cards) {
    if (c.attached_to && here.has(c.attached_to) && c.attached_to !== c.instance_id) {
      const list = onHost.get(c.attached_to) ?? [];
      list.push(c);
      onHost.set(c.attached_to, list);
    }
  }
  const out: FinalCardView[] = [];
  const placed = new Set<string>();
  const place = (c: FinalCardView) => {
    if (placed.has(c.instance_id)) return;
    placed.add(c.instance_id);
    out.push(c);
    for (const a of onHost.get(c.instance_id) ?? []) place(a);
  };
  for (const c of cards) {
    const hosted = c.attached_to && here.has(c.attached_to) && c.attached_to !== c.instance_id;
    if (!hosted) place(c);
  }
  // A loop of attachments (never expected) still shows every card.
  for (const c of cards) place(c);
  return out;
}

/** The line under a recorded card: what it was attached to, and whether it was tapped. */
export function finalCardNotes(card: FinalCardView): string[] {
  const notes: string[] = [];
  if (card.attached_to_name) notes.push(`on ${card.attached_to_name}`);
  if (card.tapped) notes.push("tapped");
  return notes;
}

/**
 * A recorded card as the CardView Card.svelte draws. It is drawn upright
 * (a tapped card would turn into its neighbour in the grid, so the note
 * says it was tapped instead) and carries nothing that acts.
 */
export function finalCardAsView(card: FinalCardView, ownerID: string): CardView {
  return {
    instance_id: card.instance_id,
    name: card.name ?? "",
    owner: ownerID,
    controller: ownerID,
    scryfall_id: card.scryfall_id,
    layout: card.layout,
    active_face: card.active_face,
    type_line: card.type_line,
    power: card.power,
    toughness: card.toughness,
    counters: card.counters,
    is_token: card.is_token,
    is_commander: card.is_commander,
    face_down: card.face_down,
  } as CardView;
}
