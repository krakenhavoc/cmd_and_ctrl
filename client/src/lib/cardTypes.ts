// Card type predicates and battlefield bucketing for the per-type
// player panel layout. The server sends one Battlefield zone per
// game with every permanent in a single Cards[] slice; the client
// splits that visually into one group per permanent type (#4087):
// CREATURES, LANDS, PLANESWALKERS, BATTLES, ARTIFACTS and ENCHANTMENTS,
// plus an OTHER group for a card whose type line names none of them.
//
// Predicates match Scryfall's printed type-line text. They're case-
// insensitive and use word boundaries so "Saga — Enchantment" doesn't
// false-match on "Sage" and "Forestland" doesn't false-match the bare
// /\bland\b/. Cards without a type_line (placeholder / unresolved
// demo cards) fall through every predicate as false.

import { type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import type { CardView, PlayerView } from "./protocol";

export function isCreature(c: CardView): boolean {
  return !!c.type_line && /\bcreature\b/i.test(c.type_line);
}

export function isLand(c: CardView): boolean {
  return !!c.type_line && /\bland\b/i.test(c.type_line);
}

export function isArtifact(c: CardView): boolean {
  return !!c.type_line && /\bartifact\b/i.test(c.type_line);
}

export function isEnchantment(c: CardView): boolean {
  return !!c.type_line && /\benchantment\b/i.test(c.type_line);
}

export function isPlaneswalker(c: CardView): boolean {
  return !!c.type_line && /\bplaneswalker\b/i.test(c.type_line);
}

export function isBattle(c: CardView): boolean {
  return !!c.type_line && /\bbattle\b/i.test(c.type_line);
}

// BattlefieldBucket names the per-player visual group a card lands in.
// "other" holds a permanent whose type line names none of the six (a
// card with no type line at all, a placeholder), so it is still drawn
// somewhere rather than dropped.
export type BattlefieldBucket =
  | "creature"
  | "land"
  | "planeswalker"
  | "battle"
  | "artifact"
  | "enchantment"
  | "other";

// bucketForBattlefield maps a card to its visual group. A permanent
// lands in EXACTLY ONE group, the first that applies, because real
// Magic cards mix types and drawing one in two groups would double it:
//
//   creature > planeswalker > battle > land > artifact > enchantment
//
//   - Creature first: an Artifact Creature (Solemn Simulacrum), an
//     Enchantment Creature (Eidolon of Blood) and a Land Creature (Dryad
//     Arbor, an animated Mutavault) are tracked in combat, which is
//     where players look for creatures. A planeswalker that is also a
//     creature (a Gideon until end of turn) goes there too, and still
//     shows its loyalty pip.
//   - Planeswalker and battle before land / artifact / enchantment: they
//     are the permanents an attacker picks as a target and that carry a
//     counter to track, so they must never be lost among the rest.
//   - Land before artifact and enchantment: an Artifact Land (Seat of the
//     Synod) taps for mana beside the basics and piles with them.
//   - Artifact before enchantment: an Enchantment Artifact (Bident of
//     Thassa) sits with the artifacts.
export function bucketForBattlefield(c: CardView): BattlefieldBucket {
  if (isCreature(c)) return "creature";
  if (isPlaneswalker(c)) return "planeswalker";
  if (isBattle(c)) return "battle";
  if (isLand(c)) return "land";
  if (isArtifact(c)) return "artifact";
  if (isEnchantment(c)) return "enchantment";
  return "other";
}

// SeatPosition is one of four quadrants on the table:
//   - self          : bottom-right, upright (the viewer)
//   - next          : bottom-left, upright (sits next to self)
//   - across        : top-left, rotated 180° (two seats on, diagonal from self)
//   - across_next   : top-right, rotated 180° (three seats on, directly across)
//
// Quadrants — not rotated 90° edges — because a 3- or 4-player game
// reads better when every panel is either upright (the bottom row) or
// flipped (the top row, "across the table"). Card art on a side-
// rotated panel was unreadable. Across-the-table flipping is the only
// rotation we apply.
export type SeatPosition = "self" | "next" | "across" | "across_next";

// SeatPlacements tells Board.svelte which seat goes where. Empty
// quadrants stay null so the layout can collapse gracefully in 2- /
// 3-player games rather than reserving empty space.
export type SeatPlacements = Record<SeatPosition, PlayerView | null>;

// seatPlacements assigns every seat to a quadrant. The viewer always
// lands at "self"; remaining seats fill clockwise from the viewer:
//
//   1 opponent  → across                            (2-player)
//   2 opponents → next + across                     (3-player)
//   3 opponents → next + across + across_next       (4-player)
//
// "Clockwise from self" matches the prior table-renderer ordering and
// the way real Commander tables seat people: the next player to your
// left takes the next turn. With three opponents, that's
// bottom-left → top-left → top-right, which maps to next → across →
// across_next in the quadrant scheme (Board.svelte places them so the
// order really does run clockwise on screen).
//
// Spectators (no viewerID) fall back to original seat order across
// self → next → across → across_next so the board still renders
// something coherent.
export function seatPlacements(seats: PlayerView[], viewerID: string | null): SeatPlacements {
  const out: SeatPlacements = { self: null, next: null, across: null, across_next: null };
  const selfIdx = viewerID ? seats.findIndex((s) => s.id === viewerID) : -1;
  if (selfIdx >= 0) {
    out.self = seats[selfIdx];
    const others = [...seats.slice(selfIdx + 1), ...seats.slice(0, selfIdx)];
    let spots: SeatPosition[];
    switch (others.length) {
      case 1:
        spots = ["across"];
        break;
      case 2:
        spots = ["next", "across"];
        break;
      default:
        spots = ["next", "across", "across_next"];
    }
    others.slice(0, 3).forEach((s, i) => {
      out[spots[i]] = s;
    });
    return out;
  }
  const fallback: SeatPosition[] = ["self", "next", "across", "across_next"];
  seats.slice(0, 4).forEach((s, i) => {
    out[fallback[i]] = s;
  });
  return out;
}

// hoveredCard is the global "what is the cursor over right now"
// store. Card.svelte writes to it on pointerenter / clears it on
// pointerleave; HoverZoomOverlay subscribes to render the big
// preview. Lifted to module scope (rather than threaded through
// props) because every Card writes to it and only one overlay
// reads, so prop-drilling would be noisy and the global state is
// genuinely scene-wide.
export const hoveredCard: Writable<CardView | null> = guardedWritable(null, "hoveredCard");
