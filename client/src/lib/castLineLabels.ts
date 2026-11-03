// castLineLabels.ts — telling apart cast lines that read the same (#4709).
//
// A resolution_cast sheet (ResolutionCastSheet) lists one line per legal
// way to cast the card on offer, and the server names each line's targets
// by card name only ("…targeting Grizzly Bears"). Two permanents with one
// name, a creature and its token copy say, then give two identical lines.
// The server's label stays the source of truth; this only adds, to lines
// that would otherwise read the same, what tells their target cards apart:
// that a card is a token, and whose it is when it is not the viewer's.
// Lines still alike after that (two of your own nontoken Bears) are
// numbered, so no two buttons ever share a name.

import type { CardView, GameView, LegalMoveView } from "./protocol";
import { findCard } from "./contextMenu.logic";

interface WireTarget {
  kind?: string;
  id?: string;
}

// targetCardIDs is the card targets a move names, in its order. A player
// target is already told apart by its name.
export function targetCardIDs(m: LegalMoveView): string[] {
  const raw = (m.params as { targets?: unknown } | undefined)?.targets;
  if (!Array.isArray(raw)) return [];
  return (raw as WireTarget[])
    .filter((t) => t && t.kind !== "player" && typeof t.id === "string" && t.id !== "")
    .map((t) => t.id as string);
}

// cardQualifier is what tells one card apart from another of its name:
// "token" for a token, and the controller's name when it is not the
// viewer. "" when there is nothing to say.
export function cardQualifier(card: CardView, view: GameView, viewerID: string | null): string {
  const parts: string[] = [];
  if (card.is_token) parts.push("token");
  const who = card.controller;
  if (who && who !== viewerID) {
    const seat = view.seats.find((s) => s.id === who);
    if (seat?.name) parts.push(`${seat.name}'s`);
  }
  return parts.join(", ");
}

// distinctCastLabels returns one label per move, in order: the move's own
// label where no other line reads the same, and otherwise that label with
// its targets' qualifiers, then a number if lines are still alike.
export function distinctCastLabels(
  moves: LegalMoveView[],
  labelOf: (m: LegalMoveView) => string,
  view: GameView,
  viewerID: string | null,
): string[] {
  const base = moves.map(labelOf);
  const count = new Map<string, number>();
  for (const l of base) count.set(l, (count.get(l) ?? 0) + 1);

  const qualified = moves.map((m, i) => {
    if ((count.get(base[i]) ?? 0) < 2) return base[i];
    const notes = targetCardIDs(m)
      .map((id) => findCard(view, id))
      .map((c) => {
        if (!c) return "";
        // Beside another target's note, your own plain card says "yours".
        return cardQualifier(c, view, viewerID) || (c.controller === viewerID ? "yours" : "");
      });
    if (notes.every((n) => n === "" || n === "yours")) return base[i];
    return `${base[i]} (${notes.map((n) => n || "—").join("; ")})`;
  });

  const seen = new Map<string, number>();
  const again = new Map<string, number>();
  for (const l of qualified) again.set(l, (again.get(l) ?? 0) + 1);
  return qualified.map((l) => {
    if ((again.get(l) ?? 0) < 2) return l;
    const n = (seen.get(l) ?? 0) + 1;
    seen.set(l, n);
    return `${l} #${n}`;
  });
}
