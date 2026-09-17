import type { CardView, PlayerView } from "./protocol";

/** Whether a tapped permanent will miss its controller's next untap step. */
export function noUntapAppliesToController(
  card: Pick<CardView, "controller" | "tapped" | "no_untap">,
): boolean {
  if (!card.tapped || !card.no_untap) return false;
  return !!card.no_untap.static || card.no_untap.next?.includes(card.controller) === true;
}

/** Footer lines for the permanent's current untap-step restrictions. */
export function noUntapFooterLines(
  card: Pick<CardView, "controller" | "no_untap">,
  seats: Pick<PlayerView, "id" | "name">[],
): string[] {
  const state = card.no_untap;
  if (!state) return [];
  const lines: string[] = [];
  if (state.static) lines.push("doesn't untap during its controller's untap step");
  for (const id of state.next ?? []) {
    const player = seats.find((seat) => seat.id === id);
    const name = player?.name || "that player";
    lines.push(`doesn't untap during ${name}'s next untap step`);
  }
  return lines;
}
