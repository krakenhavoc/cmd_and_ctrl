// permissionTypes.ts — #2167: the card type a cast spends under a
// permission that opens one of each type (Muldrotha, the Gravetide;
// Aminatou's Augury).
//
// The server publishes the types a face may spend as `permission_types`
// (protocol.CastSurfaceView), ranked so the first is the one the
// permission's other cards need least. The cast chain asks only when
// there are two or more; with one, the server settles it and nothing is
// sent.

import type { CardView } from "./protocol";

// permissionTypesOf returns the types this face of the card may spend,
// in the server's ranked order. Empty for every cast that spends no
// per-type budget.
export function permissionTypesOf(card: CardView): string[] {
  return card.permission_types ?? [];
}

// needsPermissionTypePicker reports whether casting this face asks which
// card type it uses — "If a card has multiple permanent types, choose one
// as you play it."
export function needsPermissionTypePicker(card: CardView): boolean {
  return permissionTypesOf(card).length > 1;
}

// permissionTypeLabel capitalises a type for a row: "artifact" →
// "Artifact".
export function permissionTypeLabel(t: string): string {
  return t.length === 0 ? t : t[0].toUpperCase() + t.slice(1);
}

// permissionTypeVerb is the confirm button's verb phrase: "Cast as an
// artifact", "Play as a land".
export function permissionTypeVerb(card: CardView, t: string): string {
  const verb = t === "land" ? "Play" : "Cast";
  const article = /^[aeiou]/.test(t) ? "an" : "a";
  return `${verb} ${card.name} as ${article} ${t}`;
}
