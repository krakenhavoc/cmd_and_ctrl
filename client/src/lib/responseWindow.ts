// #1307 — what smart autopass counts as "something to do", and which
// windows it asks in.
//
// Before this module the whole question was `legal_moves.some(m =>
// m.kind !== "pass")`. That counted a mana ability as a response, so
// any untapped land held the window and smart autopass almost never
// skipped anything. It was also only asked on ticked steps: an
// opponent's spell always stopped you, whether or not you could
// answer it, and an unticked step always passed, even with a
// counterspell in hand.
//
// Two questions now, over the same server-enumerated move list:
//
//   hasResponse — could the viewer answer what is happening? Counters,
//     instants, non-mana abilities and special actions, each one a
//     category the player can switch off. Mana and land never count.
//   hasPlay — is there anything to do on a step the player ticked?
//     A response, or a land, a sorcery-speed cast, a declaration.
//
// and one about the moment: keyWindow says whether the cursor is in a
// window where a response is worth stopping for even without a tick.
// ADR 0009, amendment #1307.
//
// Pure functions over the frame. The decision lives in
// autopassDecision.ts; this module only answers questions.

import type { GameView, LegalMoveView } from "./protocol";
import { hasDeclaredAttackers, owesBlockDecision } from "./priority";
import { ownsEveryStackItem } from "./holdPriority";
import { hasPriority, isActivePlayer, isMainPhase, stackEmpty } from "./timing";

// MoveClass is what a move means to autopass.
//   none        — pass, mana: never a reason to hold.
//   play        — a land, or a cast / activation in the viewer's own
//                 sorcery window. Only ever counts on a ticked step.
//   counter     — a cast or activation that targets the stack.
//   instant     — any other cast (instant, flash, split second's
//                 exceptions — whatever the enumerator offered).
//   ability     — any other non-mana activated ability.
//   special     — a CR 116.2 special action (foretell, suspend, …).
//   declaration — an attack or a block.
//   other       — a choice answer, a mulligan, a kind this client
//                 does not know. Counts as a play, never a response.
export type MoveClass =
  | "none"
  | "play"
  | "counter"
  | "instant"
  | "ability"
  | "special"
  | "declaration"
  | "other";

// ResponseCategories are the four gameplay.respond* toggles.
export interface ResponseCategories {
  counter: boolean;
  instant: boolean;
  ability: boolean;
  special: boolean;
}

export const ALL_RESPONSES: ResponseCategories = {
  counter: true,
  instant: true,
  ability: true,
  special: true,
};

// classifyMove sorts one enumerated move. `sorceryWindow` is "the
// viewer's own main phase with an empty stack" — the one window where
// a cast or activation is a play the viewer came to make rather than
// an answer to something.
//
// An older server sends no `targets_stack`, so a counterspell there
// reads as `instant`. With both categories on (the default) that
// changes nothing; it only matters to a player who turned instants
// off and kept counters on.
export function classifyMove(m: LegalMoveView, sorceryWindow: boolean): MoveClass {
  switch (m.kind) {
    case "pass":
    case "mana":
      return "none";
    case "land":
      return "play";
    case "cast":
    case "activate":
      if (sorceryWindow) return "play";
      if (m.targets_stack) return "counter";
      return m.kind === "cast" ? "instant" : "ability";
    case "special_action":
      return "special";
    case "attack":
    case "block":
      return "declaration";
    default:
      return "other";
  }
}

// inSorceryWindow: the viewer's own main phase, stack empty.
export function inSorceryWindow(view: GameView | null | undefined, me: string | null): boolean {
  return isActivePlayer(view, me) && isMainPhase(view) && stackEmpty(view);
}

function isEnabledResponse(c: MoveClass, cats: ResponseCategories): boolean {
  switch (c) {
    case "counter":
      return cats.counter;
    case "instant":
      return cats.instant;
    case "ability":
      return cats.ability;
    case "special":
      return cats.special;
    default:
      return false;
  }
}

// hasResponse reports whether the viewer holds priority and has a move
// in one of the enabled response categories.
//
// No move list while holding priority → true. That is a pre-S31 server
// or a dropped field, and ADR 0009 §3 says to stop rather than guess.
export function hasResponse(
  view: GameView | null | undefined,
  me: string | null,
  cats: ResponseCategories,
): boolean {
  if (!view || !me) return false;
  if (!hasPriority(view, me)) return false;
  const moves = view.legal_moves;
  if (!moves) return true;
  const sw = inSorceryWindow(view, me);
  return moves.some((m) => isEnabledResponse(classifyMove(m, sw), cats));
}

// hasPlay reports whether a ticked step has anything in it for the
// viewer. Everything hasResponse counts, plus lands, sorcery-speed
// casts, declarations and choice answers, plus the two windows the
// engine has no move for: an owed block (#328) and the attack the
// viewer just declared (#599).
//
// Lands count here and only here, so a hand holding nothing but a land
// still stops on the viewer's own main phase — and never holds an
// opponent's spell.
export function hasPlay(
  view: GameView | null | undefined,
  me: string | null,
  cats: ResponseCategories,
): boolean {
  if (!view || !me) return false;
  if (owesBlockDecision(view, me)) return true;
  if (!hasPriority(view, me)) return false;
  if (hasDeclaredAttackers(view, me)) return true;
  const moves = view.legal_moves;
  if (!moves) return true;
  const sw = inSorceryWindow(view, me);
  return moves.some((m) => {
    const c = classifyMove(m, sw);
    if (c === "play" || c === "declaration" || c === "other") return true;
    return isEnabledResponse(c, cats);
  });
}

// KeyWindow names the windows where a response is worth a stop even
// on a step the player did not tick.
//   stackOpp — something on the stack the viewer did not put there
//              (or a trigger still queuing, which could be anyone's).
//   combat   — declare attackers / blockers with an attack declared.
//   oppEnd   — an opponent's end step, the classic flash window.
export interface KeyWindow {
  stackOpp: boolean;
  combat: boolean;
  oppEnd: boolean;
}

export function keyWindow(view: GameView | null | undefined, me: string | null): KeyWindow {
  const none: KeyWindow = { stackOpp: false, combat: false, oppEnd: false };
  if (!view || !me) return none;
  const step = view.turn?.step;
  return {
    stackOpp: !stackEmpty(view) && !ownsEveryStackItem(view, me),
    combat:
      (step === "declare_attackers" || step === "declare_blockers") &&
      (view.battlefield?.cards ?? []).some((c) => !!c.attacking_target),
    oppEnd: step === "end" && !isActivePlayer(view, me),
  };
}
