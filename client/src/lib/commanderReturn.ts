import { nameList } from "./openingRoll";
import type { CardView, GameView } from "./protocol";

// commanderReturn.ts — ADR 0115 PR 5: the two small facts the
// commander_return prompt needs from a view. The card the question is
// about, wherever it sits (a graveyard or exile for CR 903.9a, the
// battlefield or the stack for a CR 903.9b replacement), and the line
// the seats that are NOT deciding read while it is open.

type Zones = Pick<GameView, "battlefield" | "exile" | "seats" | "stack_items">;

/**
 * findCardView looks a card up across the public zones a commander can
 * be asked about. A card the viewer may not see arrives with an empty
 * name; the caller decides what to do with that.
 */
export function findCardView(view: Partial<Zones>, id: string | undefined): CardView | null {
  if (!id) return null;
  const seek = (cards: CardView[] | undefined): CardView | null =>
    cards?.find((c) => c.instance_id === id) ?? null;
  const found = seek(view.battlefield?.cards) ?? seek(view.exile?.cards);
  if (found) return found;
  for (const s of view.seats ?? []) {
    const hit = seek(s.graveyard?.cards) ?? seek(s.command?.cards) ?? seek(s.hand?.cards);
    if (hit) return hit;
  }
  return null;
}

/**
 * commanderReturnWaiting is "Alice is deciding about their commander",
 * or "" when no commander_return prompt is open. Public state only: the
 * chooser of each open prompt is on every seat's view. The deciding
 * seat reads the prompt itself in the dock, which hides this line.
 */
export function commanderReturnWaiting(
  view: Pick<GameView, "pending_choices" | "seats"> | null | undefined,
): string {
  const names: string[] = [];
  for (const c of view?.pending_choices ?? []) {
    if (c.kind !== "commander_return") continue;
    const name = view?.seats?.find((s) => s.id === c.chooser)?.name;
    if (name && !names.includes(name)) names.push(name);
  }
  if (names.length === 0) return "";
  return names.length === 1
    ? `${names[0]} is deciding about their commander`
    : `${nameList(names)} are deciding about their commanders`;
}
