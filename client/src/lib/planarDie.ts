// planarDie.ts — the action dock's planar die button at a Planechase
// table (ADR 0216 §4, §10).
//
// CR 901.9 makes rolling the planar die a special action of the active
// player, in their main phase, with priority and an empty stack; the
// first roll each turn is free and each one after costs {1} more. The
// server decides all of it: the button is live exactly when the
// viewer's own legal_moves carry the `special_action` move whose kind is
// `planar_die`, and pressing it sends that move's params (strict and
// auto_tap, as the enumerator wrote them). The cost shown is the
// server's `die_cost`, priced by the CR 601.2f pass.
//
// NOTHING HERE IS A RULE. The reason a disabled button gives is read
// off the frame for the title only; the server refuses whatever the
// title says.

import type { GameView, LegalMoveView } from "./protocol";

export const PLANAR_DIE_KIND = "planar_die";

/** The planar die move in the viewer's legal_moves, or null. */
export function planarDieMove(
  view: GameView | null | undefined,
  viewerID: string | null | undefined,
): LegalMoveView | null {
  if (!view?.planechase || !viewerID) return null;
  return (
    (view.legal_moves ?? []).find(
      (m) =>
        m.type === "special_action" &&
        m.params?.kind === PLANAR_DIE_KIND &&
        (m.player === undefined || m.player === "" || m.player === viewerID),
    ) ?? null
  );
}

const MAIN_STEPS = new Set(["precombat_main", "postcombat_main"]);

/** The dock's planar die button: what it shows and what pressing it sends. */
export interface PlanarDieButton {
  /** The cost in brace notation, "{0}" for the first roll of a turn. */
  cost: string;
  enabled: boolean;
  /** The button's title: what it does, or why it can't be pressed now. */
  title: string;
  /** The `special_action` params to send, when enabled. */
  params: Record<string, unknown> | null;
}

/**
 * planarDieButton is the button for a seated viewer at a Planechase
 * table, or null at any other table (and for a spectator). Disabled, not
 * hidden (ADR 0111 §1), outside the window the server offers it in.
 */
export function planarDieButton(
  view: GameView | null | undefined,
  viewerID: string | null | undefined,
): PlanarDieButton | null {
  if (!view?.planechase || !viewerID) return null;
  const me = view.seats.find((s) => s.id === viewerID);
  if (!me || me.eliminated) return null;
  const cost = view.planechase.die_cost || "{0}";
  const free = cost === "{0}";
  const move = planarDieMove(view, viewerID);
  if (move) {
    return {
      cost,
      enabled: true,
      title: free
        ? "Roll the planar die: chaos, planeswalk or blank. Your first roll this turn is free"
        : `Roll the planar die again for ${cost}: each roll this turn costs 1 more`,
      params: move.params ?? { kind: PLANAR_DIE_KIND },
    };
  }
  return { cost, enabled: false, title: planarDieBlocked(view, me.seat, cost), params: null };
}

/** Why the planar die can't be rolled now, for the disabled button's title. */
export function planarDieBlocked(view: GameView, seat: number, cost: string): string {
  const t = view.turn;
  if (view.state !== "active" || view.mulligans_open)
    return "The planar die is rolled once the game is under way";
  if (t.active_seat !== seat) return "Only the player whose turn it is can roll the planar die";
  if (!MAIN_STEPS.has(t.step)) return "Roll the planar die in your main phase";
  if ((view.stack_items ?? []).length > 0 || (view.stack?.count ?? 0) > 0) {
    return "Roll the planar die when the stack is empty";
  }
  if (t.priority_holder !== seat) return "Roll the planar die when you have priority";
  return `You can't pay ${cost} for another roll this turn`;
}
