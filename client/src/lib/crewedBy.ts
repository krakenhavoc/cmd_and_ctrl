import type { GameView } from "./protocol";

// crewedBy.ts — ADR 0186 §6: the names behind CardView.crewed_by /
// saddled_by. The wire carries card IDs; the names come from the board's
// own card views. An ID with no card on the battlefield (it left after
// tapping, or the caller has no frame) is left out, never shown as an ID.

type Board = Pick<GameView, "battlefield">;

export function crewNames(
  view: Board | null | undefined,
  ids: readonly string[] | undefined,
): string[] {
  if (!ids || ids.length === 0) return [];
  const byID = new Map<string, string>();
  for (const c of view?.battlefield?.cards ?? []) byID.set(c.instance_id, c.name);
  const out: string[] = [];
  for (const id of ids) {
    const name = byID.get(id);
    if (name) out.push(name);
  }
  return out;
}
