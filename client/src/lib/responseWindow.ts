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
//     instants, targeted abilities, untargeted abilities and special
//     actions, each one a category the player can switch off. Mana and
//     land never count, and untargeted abilities are off by default
//     (#2853).
//   hasPlay — is there anything to do on a step the player ticked?
//     A response, or a land, a sorcery-speed cast, a declaration.
//
// Both read the combat window (#2871): crewing a Vehicle, animating a
// manland or granting flying is a response there and nowhere else.
//
// and one about the moment: keyWindow says whether the cursor is in a
// window where a response is worth stopping for even without a tick.
// ADR 0009, amendment #1307.
//
// Pure functions over the frame. The decision lives in
// autopassDecision.ts; this module only answers questions.

import type { GameView, LegalMoveView } from "./protocol";
import { attackersDefendedBy } from "./attackTargets";
import { hasDeclaredAttackers, owesBlockDecision } from "./priority";
import { ownsEveryStackItem } from "./holdPriority";
import { hasPriority, isActivePlayer, isMainPhase, stackEmpty } from "./timing";

// MoveClass is what a move means to autopass.
//   none        — pass, mana (but see ability), and an activation of a
//                 permanent another player controls (ADR 0106 §1):
//                 never a reason to hold.
//   play        — a land, or a cast / activation in the viewer's own
//                 sorcery window. Only ever counts on a ticked step.
//   counter     — a cast or activation that targets the stack.
//   instant     — any other cast (instant, flash, split second's
//                 exceptions — whatever the enumerator offered).
//   ability     — any other activated ability that targets or protects
//                 (#2853: `has_targets`, or `interacts`: a sacrifice
//                 outlet, regeneration, a pump, a blink, a shield; a
//                 mana ability only when it is a sacrifice outlet), or
//                 one that changes a fight, in a combat window only
//                 (#2871: `combat_interacts`: crew, a manland, a
//                 granted keyword, an extra block; a creature token
//                 only while the viewer defends, `combat_defender_only`).
//   untargeted  — any other non-mana activated ability: pure value
//                 like Mind Stone, a fetch land, a Clue, cycling. Its
//                 own class so it does not count as a response by
//                 default.
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
  | "untargeted"
  | "special"
  | "declaration"
  | "other";

// ResponseCategories are the five gameplay.respond* toggles.
export interface ResponseCategories {
  counter: boolean;
  instant: boolean;
  ability: boolean;
  untargeted: boolean;
  special: boolean;
}

export const ALL_RESPONSES: ResponseCategories = {
  counter: true,
  instant: true,
  ability: true,
  untargeted: true,
  special: true,
};

// DEFAULT_RESPONSES is what a player who never touched "Stop for"
// gets (#2853, owner decision 1): an opponent's stack item stops you
// for an instant, a counter, or an ability that targets or protects
// (owner answer 2). A pure value ability is not interaction.
export const DEFAULT_RESPONSES: ResponseCategories = {
  ...ALL_RESPONSES,
  untargeted: false,
};

// classifyMove sorts one enumerated move. `sorceryWindow` is "the
// viewer's own main phase with an empty stack" — the one window where
// a cast or activation is a play the viewer came to make rather than
// an answer to something.
//
// An older server sends no `targets_stack`, so a counterspell there
// reads as `instant`. With both categories on (the default) that
// changes nothing; it only matters to a player who turned instants
// off and kept counters on. Nor does it send `has_targets`, so every
// non-counter activation there reads as `untargeted`.
//
// `combat` is inCombatWindow: there, an activation the server marks
// `combat_interacts` is an `ability`; elsewhere it is `untargeted`.
// `defending` is isDefending: an activation also marked
// `combat_defender_only` (a creature-token maker) is an `ability` only
// when the viewer is being attacked as well (owner answer, #2871).
//
// ADR 0106 §1 decision 7 (owner decision 1, #1793): `controllers` maps
// a battlefield permanent's instance ID to its controller, and `me` is
// the viewer. An activate move whose source is a permanent somebody
// else controls is an "Any player may activate this ability" row
// (CR 602.2), and that is a legal move for every seat at every
// priority window while the permanent is on the table. Counting it
// would stop smart autopass on every window for the rest of the game,
// so it is `none`: neither a response nor a play. A source the map
// does not know (a card in hand, a caller with no map) keeps the
// ordinary classification.
export function classifyMove(
  m: LegalMoveView,
  sorceryWindow: boolean,
  controllers?: ReadonlyMap<string, string>,
  me?: string | null,
  combat = false,
  defending = false,
): MoveClass {
  if (m.kind === "activate" && controllers && me && activatesAcross(m, controllers, me)) {
    return "none";
  }
  switch (m.kind) {
    case "pass":
      return "none";
    case "mana":
      // #2853: a sacrifice outlet that makes mana (Ashnod's Altar)
      // still answers removal. Never a play: it is not why the viewer
      // stopped on their own main phase.
      return m.interacts && !sorceryWindow ? "ability" : "none";
    case "land":
      return "play";
    case "cast":
    case "activate":
      if (sorceryWindow) return "play";
      if (m.targets_stack) return "counter";
      if (m.kind === "cast") return "instant";
      if (m.has_targets || m.interacts) return "ability";
      if (combat && m.combat_interacts && (!m.combat_defender_only || defending)) {
        return "ability";
      }
      return "untargeted";
    case "special_action":
      return "special";
    // #1501: finishing a block declaration is part of the declaration.
    case "attack":
    case "block":
    case "finish_blocks":
      return "declaration";
    default:
      return "other";
  }
}

// activatesAcross: the move's source is on the battlefield under a
// controller other than `me`.
function activatesAcross(
  m: LegalMoveView,
  controllers: ReadonlyMap<string, string>,
  me: string,
): boolean {
  if (!m.source) return false;
  const controller = controllers.get(m.source);
  return !!controller && controller !== me;
}

// battlefieldControllers is classifyMove's `controllers` for a frame:
// the controller of every permanent on the battlefield, by instance ID.
export function battlefieldControllers(
  view: GameView | null | undefined,
): ReadonlyMap<string, string> {
  const out = new Map<string, string>();
  for (const c of view?.battlefield?.cards ?? []) {
    if (c.controller) out.set(c.instance_id, c.controller);
  }
  return out;
}

// COMBAT_STEPS are the steps where a combat ability is a response
// (#2871): the last moment to make a blocker or an attacker, and the
// two declarations. Combat damage and end of combat are too late.
const COMBAT_STEPS: ReadonlySet<string> = new Set([
  "begin_combat",
  "declare_attackers",
  "declare_blockers",
]);

// ATTACK_OR_BLOCK names a stack item about an attack or a block: an
// attack trigger, a "whenever this blocks" trigger.
const ATTACK_OR_BLOCK = /\b(attack|attacks|attacking|attacked|block|blocks|blocking|blocked)\b/i;

// inCombatWindow reports whether a combat ability counts as a response
// now (#2871): beginning of combat, declare attackers or declare
// blockers, or an attack or block trigger on the stack. Elsewhere a
// board of them would stop the viewer on every spell.
export function inCombatWindow(view: GameView | null | undefined): boolean {
  if (!view) return false;
  if (COMBAT_STEPS.has(view.turn?.step ?? "")) return true;
  return (view.stack_items ?? []).some(
    (it) => it.kind !== "spell" && ATTACK_OR_BLOCK.test(it.label ?? ""),
  );
}

// isDefending reports whether the viewer is a defending player in this
// combat (#2871): a creature is attacking them, or a planeswalker or
// battle they defend. A creature token is a blocker, so a token maker
// counts as a response only then.
export function isDefending(view: GameView | null | undefined, me: string | null): boolean {
  if (!view || !me) return false;
  return attackersDefendedBy(view, me).length > 0;
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
    case "untargeted":
      return cats.untargeted;
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
  const combat = inCombatWindow(view);
  const defending = isDefending(view, me);
  const controllers = battlefieldControllers(view);
  return moves.some((m) =>
    isEnabledResponse(classifyMove(m, sw, controllers, me, combat, defending), cats),
  );
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
  const combat = inCombatWindow(view);
  const defending = isDefending(view, me);
  const controllers = battlefieldControllers(view);
  return moves.some((m) => {
    const c = classifyMove(m, sw, controllers, me, combat, defending);
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
