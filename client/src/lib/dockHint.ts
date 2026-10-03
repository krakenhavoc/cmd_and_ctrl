// dockHint.ts — ADR 0111 §1, the action dock's status line: what
// pressing `next` will do. It reads public state only (the top of the
// stack and the step), so it says nothing a seat across the table
// could not work out from the same board.
//
// It names what happens once EVERY player passes, which is what a pass
// is for: with a spell on the stack the top item resolves, and with an
// empty stack the turn moves on. Your own pass alone only hands
// priority to the next seat, and the line does not pretend otherwise.
//
// Every step-to-step line here is one the engine always takes:
//   - first strike is folded into "combat damage" (CR 510.4), because
//     whether the turn has a first-strike step depends on the board;
//   - declare attackers goes to end of combat when nothing attacked
//     (CR 508.8), and to declare blockers otherwise.

import type { GameView } from "./protocol";
import { buildStackLane } from "./stackLane";
import { STEP_LABELS, type StepID } from "./turn";

// The step an all-pass on an empty stack moves to, for every step that
// grants priority. Untap and cleanup grant none (CR 502.4, 514.3), so a
// pass there is not a thing the dock can offer.
const NEXT_STEP: Partial<Record<StepID, StepID | "next_turn">> = {
  upkeep: "draw",
  draw: "precombat_main",
  precombat_main: "begin_combat",
  begin_combat: "declare_attackers",
  // declare_attackers is decided by the board, below.
  declare_blockers: "combat_damage",
  first_strike_damage: "combat_damage",
  combat_damage: "end_combat",
  end_combat: "postcombat_main",
  postcombat_main: "end",
  end: "next_turn",
};

/**
 * passHint is the status line's "what does `next` do" sentence, or ""
 * when there is nothing to say: the viewer does not hold priority, the
 * opening hands are still being decided, or the step is one the client
 * does not know.
 *
 * #1501: one window is worth a line without priority. While defenders
 * declare blockers nobody holds priority (CR 509.1 takes the
 * declaration first), `next` is greyed for everyone, and the line says
 * who the table is waiting on. A defender still declaring has the
 * block request in the action bar instead, which hides this line.
 */
export function passHint(view: GameView | null | undefined, viewerHasPriority: boolean): string {
  if (!view || view.mulligans_open === true) return "";
  if (!viewerHasPriority) return blockWaitHint(view);

  const top = topOfStackName(view);
  if (top) return `passing lets ${top} resolve`;

  const step = view.turn?.step as StepID | undefined;
  if (!step) return "";
  const next = step === "declare_attackers" ? afterDeclareAttackers(view) : NEXT_STEP[step];
  if (!next) return "";
  if (next === "next_turn") return "passing ends the turn";
  return `passing moves to ${STEP_LABELS[next]}`;
}

// blockWaitHint names the defenders still declaring blockers while
// priority is parked for them (#1501), or "" in every other window.
function blockWaitHint(view: GameView): string {
  const turn = view.turn;
  if (turn?.step !== "declare_blockers" || (turn.priority_holder ?? -1) >= 0) return "";
  const names = (turn.block_pending_seats ?? [])
    .map((i) => view.seats?.[i]?.name)
    .filter((n): n is string => !!n);
  if (names.length === 0) return "";
  const who =
    names.length === 1
      ? names[0]
      : `${names.slice(0, -1).join(", ")} and ${names[names.length - 1]}`;
  return `waiting for ${who} to declare blockers`;
}

function topOfStackName(view: GameView): string {
  if ((view.stack_items?.length ?? 0) === 0 && (view.stack?.count ?? 0) === 0) return "";
  const model = buildStackLane({
    stack: view.stack,
    stackItems: view.stack_items,
    pendingTriggers: view.pending_triggers,
    seats: view.seats,
    battlefield: view.battlefield,
    exile: view.exile,
    viewerID: null,
    priorityHolder: null,
    splitSecondActive: view.split_second_active === true,
  });
  return model.stackItems[0]?.name ?? "";
}

// CR 508.8: with no attackers declared, the declare blockers and
// combat damage steps are skipped.
function afterDeclareAttackers(view: GameView): StepID {
  const attacking = (view.battlefield?.cards ?? []).some((c) => !!c.attacking_target);
  return attacking ? "declare_blockers" : "end_combat";
}
