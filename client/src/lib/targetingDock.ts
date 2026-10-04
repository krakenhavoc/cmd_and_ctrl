// targetingDock.ts — targeting and payment in the action dock (ADR 0111
// Delivery PR 4). Pure builders, like combatDock.ts: the targeting
// store and Game.svelte own the state and the handlers, and these turn
// them into the DockRequest the dock draws.
//
//   - Targeting (ADR 0111 §6). Board targeting stays on the board: the
//     highlights, the rings and the click that picks. What was the
//     attention strip's TargetingBanner is now a request: its sentence
//     is the question line, Done is the primary (only where a pick list
//     needs one, disabled until enough are picked) and Cancel the
//     secondary. A cast's walk is a `flow` (the player started it); a
//     trigger's pick_target or a retarget is a `choice` (the game waits
//     on it), and has no Cancel, as before: the trigger needs a target.
//   - Insufficient mana (S15). A strict cast the server refused for
//     mana. Auto-tap & cast is the primary (it opens the auto-tap
//     preview, which still confirms before anything is tapped), Cast
//     anyway and Cancel the secondaries. It was a strip toast.
//
// The names are the old ones (ADR 0111 §10): the request's dialog is
// "Select target for <card>", which s19-triggers.spec.ts waits on, and
// the buttons keep "Done", "Cancel", "Auto-tap & cast", "Cast anyway".

import type { DockAction, DockRequest } from "./dock";
import type { GameView } from "./protocol";
import {
  canConfirm,
  isMultiPick,
  legalTargetCount,
  type TargetingMode,
  type TargetingState,
} from "./targeting";
import { doubledTriggerLabel } from "./triggerDoubling";
import { targetPriceSummary } from "./targetPrices";

function modeHint(mode: TargetingMode | string | undefined): string {
  switch (mode) {
    case "any":
      return "a player or creature";
    case "player":
      return "a player";
    case "creature":
      return "a creature";
    case "permanent":
      return "a permanent";
    case "stack_spell":
      return "a spell on the stack";
    // #1211, CR 115.4. An ability on the stack is drawn in the same
    // overlay row a spell is, so the sentence is the only thing that
    // tells a player which half of the stack they may click.
    case "stack_ability":
      return "an ability on the stack";
    case "stack_item":
      return "a spell or ability on the stack";
    case "card_in_graveyard":
      return "a card in a graveyard";
    default:
      return "a target";
  }
}

// targetingQuestion is the question line's sentence.
function targetingQuestion(state: TargetingState): string {
  const name = state.card.name;
  if (state.choiceID && state.choiceKind === "retarget") {
    // #1196, CR 115.7: the source is a spell that is RESOLVING, not a
    // trigger, and it is pointing something else somewhere else.
    // "Deflecting Swat triggered" would be the wrong sentence.
    return `${name} — click ${state.label || "a new target"}`;
  }
  if (state.choiceID) return `${name} triggered — click ${state.label || "a target"}`;
  // #764: a multi-clause or per-mode announcement is a WALK, so the
  // line names the clause it is asking about right now (and the detail
  // says where in the walk the player is). Without this a two-slot card
  // asks the same question twice with no way to tell which slot is open.
  if (state.steps.length > 1) return `${name} — click ${state.label || modeHint(state.mode)}`;
  return `Click ${modeHint(state.mode)} to target ${name}`;
}

// #1659: the CURRENT step's divided amount, so a player sees "divide 4
// damage among up to 3 targets" while still picking. X-based amounts
// show the announced X when the walk resolved it; divideFromXUnresolved
// is the defensive fallback, so this never prints the 0 an unresolved X
// would otherwise resolve to.
function divideText(state: TargetingState): string {
  if (state.divide === undefined) return "";
  const amount =
    (state.divideUpTo ? "up to " : "") + (state.divideFromXUnresolved ? "X" : `${state.divide}`);
  const targets =
    state.max > 0
      ? `up to ${state.max} target${state.max === 1 ? "" : "s"}`
      : "any number of targets";
  return `divide ${amount} damage among ${targets}`;
}

// #1296: an ability whose price reads its target (Dragonfire Blade's
// "{1} less for each color of the creature it targets") shows each
// price with the targets that pay it. An equip's one click is also its
// confirm, so this is the last place the price is visible before it is
// paid.
function priceLine(state: TargetingState, view: GameView | null | undefined): string {
  const prices = state.ability?.prices;
  if (!prices || !view) return "";
  const names = new Map<string, string>();
  for (const c of view.battlefield.cards) names.set(c.instance_id, c.name);
  for (const p of view.seats) names.set(p.id, p.name);
  return targetPriceSummary(prices, (id) => names.get(id));
}

// targetingDetail is everything after the sentence, " · "-separated:
// where the walk is, how many targets are legal, the set rules, the
// doubler, the division, the pick tally and the price.
export function targetingDetail(state: TargetingState, view?: GameView | null): string {
  const parts: string[] = [];
  if (!state.choiceID && state.steps.length > 1) {
    parts.push(`step ${state.step + 1} of ${state.steps.length}`);
  }
  const count = legalTargetCount(state);
  if (count >= 0) parts.push(`${count} legal`);
  // #1559: the clause's rule over the chosen set. Candidates that would
  // break it are greyed; this says why.
  if (state.different) parts.push(`targets must ${state.different.label}`);
  // #1807: "from a single graveyard".
  if (state.same) parts.push(`targets must ${state.same.label}`);
  const doubled = doubledTriggerLabel(state.doubledBy, state.doubledByName);
  if (doubled) parts.push(doubled);
  const divide = divideText(state);
  if (divide) parts.push(divide);
  // S20 sub-PR 5: a pick list shows its tally (and has a Done).
  if (isMultiPick(state)) {
    const n = state.picked.length;
    parts.push(state.max > 0 ? `${n}/${state.max} picked` : `${n} picked`);
  }
  const price = priceLine(state, view);
  if (price) parts.push(`costs ${price}`);
  return parts.join(" · ");
}

export interface TargetingHandlers {
  onDone: () => void;
  onCancel: () => void;
}

export function targetingRequest(
  state: TargetingState,
  view: GameView | null | undefined,
  h: TargetingHandlers,
): DockRequest {
  const multi = isMultiPick(state);
  const confirmable = canConfirm(state);
  // A single pick completes on the click, so it has no Done (as the
  // banner had none): the corner stays empty, like a combat selection.
  const primary: DockAction | null = multi
    ? {
        id: "done",
        label: "Done",
        disabled: !confirmable,
        title:
          state.min > 0 && state.picked.length < state.min
            ? `pick at least ${state.min}`
            : "confirm targets",
        keyShortcuts: "Enter",
        cap: "⏎",
        onPress: h.onDone,
      }
    : null;
  // A pending choice's target cannot be refused: the trigger needs one.
  const secondary: DockAction[] = state.choiceID
    ? []
    : [
        {
          id: "cancel",
          label: "Cancel",
          title: "cancel the cast",
          keyShortcuts: "Escape",
          cap: "Esc",
          onPress: h.onCancel,
        },
      ];
  return {
    rank: state.choiceID ? "choice" : "flow",
    label: `Select target for ${state.card.name}`,
    tag: "target",
    tone: "gold",
    live: true,
    question: targetingQuestion(state),
    detail: targetingDetail(state, view) || undefined,
    primary,
    secondary,
  };
}

export interface InsufficientManaHandlers {
  onAutoTap: () => void;
  onCastAnyway: () => void;
  onCancel: () => void;
}

// insufficientManaRequest: a strict cast was refused for mana (the
// server's insufficient_mana, with what it is missing).
export function insufficientManaRequest(
  missing: readonly string[],
  cardName: string | undefined,
  h: InsufficientManaHandlers,
): DockRequest {
  return {
    rank: "flow",
    label: "insufficient mana",
    tag: "mana",
    tone: "gold",
    question: cardName ? `Insufficient mana for ${cardName}` : "Insufficient mana",
    detail: missing.length > 0 ? `missing ${missing.join(" ")}` : undefined,
    primary: {
      id: "auto-tap",
      label: "Auto-tap & cast",
      title: "preview which of your sources would pay, then cast",
      keyShortcuts: "Enter",
      cap: "⏎",
      onPress: h.onAutoTap,
    },
    secondary: [
      {
        id: "cancel",
        label: "Cancel",
        title: "don't cast it",
        keyShortcuts: "Escape",
        cap: "Esc",
        onPress: h.onCancel,
      },
      {
        id: "cast-anyway",
        label: "Cast anyway",
        title: "cast it without paying the missing mana (the sandbox override)",
        onPress: h.onCastAnyway,
      },
    ],
  };
}

export interface CastAnywayConfirmHandlers {
  onCast: () => void;
  onCancel: () => void;
}

// castAnywayConfirmLabel is the confirmation's dialog name (ADR 0118 §2,
// owner decision 6), e.g. "Cast Craw Wurm without paying its mana cost?".
// A label contract (AGENTS.md §5, ADR 0111 §10): cast-anyway-2188.spec.ts
// selects the dialog by it.
export function castAnywayConfirmLabel(cardName: string): string {
  return `Cast ${cardName} without paying its mana cost?`;
}

// castAnywayConfirmRequest: the "Cast anyway (don't pay)" row was chosen
// from a card's menu, and nothing has been sent yet (ADR 0118 §2, owner
// decision 6). The question is the dialog's name; Cast and Cancel are
// its two buttons, and both names are a label contract. Cancel answers
// Escape. Cast has no key, for the reason the insufficient-mana
// request's Cast anyway has none: no keystroke should reach an unpaid
// cast. A flow does not move focus today; should the dock ever move it,
// it goes to the dialog, never to Cast (`focus: "dialog"`).
//
// The dock's own Cast anyway, on a refused cast (insufficientManaRequest
// above), does not open this: it is already the second step of a choice
// the player made.
export function castAnywayConfirmRequest(
  cardName: string,
  h: CastAnywayConfirmHandlers,
): DockRequest {
  const question = castAnywayConfirmLabel(cardName);
  return {
    rank: "flow",
    label: question,
    tag: "unpaid",
    tone: "gold",
    question,
    hint: "No mana is spent. Life and any other costs are still paid, and the game log shows the table.",
    focus: "dialog",
    primary: {
      id: "cast",
      label: "Cast",
      title: "cast it without paying its mana cost; the game log says so",
      onPress: h.onCast,
    },
    secondary: [
      {
        id: "cancel",
        label: "Cancel",
        title: "don't cast it",
        keyShortcuts: "Escape",
        cap: "Esc",
        onPress: h.onCancel,
      },
    ],
  };
}
