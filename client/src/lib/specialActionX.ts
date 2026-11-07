/**
 * specialActionX.ts — a special action whose cost has an {X} (#5249).
 *
 * CR 107.3d: the player taking a special action chooses the X in its
 * cost immediately before paying it. Aurelia's Vindicator's "Turn face up
 * {X}{3}{W}" is the row: the server marks it `demands_x` and stamps
 * `x_max`, the largest X the viewer can pay right now. A row like that is
 * not sent straight from the menu. The Board asks for X first, with the X
 * picker every {X} spell and ability uses, and then sends the payload
 * with `x_value`.
 */

import type { CardView, GameView, SpecialActionView } from "./protocol";

/** The card and the row a `special_action` payload is about, when it needs an X. */
export interface SpecialActionXRow {
  card: CardView;
  row: SpecialActionView;
}

/**
 * specialActionNeedsX finds the row a `special_action` payload names when
 * that row demands an X and the payload has none yet, or null. The row is
 * matched on the payload's kind, route and door, the same fields the menu
 * row put there (contextMenu.logic.ts specialActionItems).
 */
export function specialActionNeedsX(view: GameView, params: unknown): SpecialActionXRow | null {
  if (!params || typeof params !== "object") return null;
  const p = params as Record<string, unknown>;
  if (p.x_value !== undefined || typeof p.card_id !== "string") return null;
  const card = (view.battlefield?.cards ?? []).find((c) => c.instance_id === p.card_id);
  const row = card?.special_actions?.find(
    (sa) =>
      sa.kind === p.kind &&
      (sa.route ?? undefined) === (p.route ?? undefined) &&
      (sa.door ?? undefined) === (p.door ?? undefined),
  );
  return card && row?.demands_x ? { card, row } : null;
}

/** withX is the payload with the chosen X. */
export function withX(params: unknown, x: number): Record<string, unknown> {
  return { ...(params as Record<string, unknown>), x_value: x };
}
