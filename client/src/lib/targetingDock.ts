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
//     anyway and Cancel the secondaries. It was a strip toast. Since
//     ADR 0118 §1 a clicked cast is already auto-tapped, and its
//     refusal drops Auto-tap & cast: Cancel is the primary.
//
// The names are the old ones (ADR 0111 §10): the request's dialog is
// "Select target for <card>", which s19-triggers.spec.ts waits on, and
// the buttons keep "Done", "Cancel", "Auto-tap & cast", "Cast anyway".

import type { DockAction, DockRequest } from "./dock";
import { L } from "./labels";
import type { GameView } from "./protocol";
import {
  canConfirm,
  effectiveShareScope,
  isCardSetPick,
  isMultiPick,
  legalTargetCount,
  type ShareScope,
  type TargetingMode,
  type TargetingState,
} from "./targeting";
import { doubledTriggerLabel } from "./triggerDoubling";
import { targetPriceSummary } from "./targetPrices";
import { batchPosition, type BatchPosition } from "./triggerBatch";

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
    // #3259: the cards live in a player's exile pile, which the viewer
    // opens like a graveyard; the sentence says where to look.
    case "card_in_exile":
      return "a card in exile (open a player's exile pile)";
    default:
      return "a target";
  }
}

// cardSetCount is how many a card-set pick asks for, in words.
function cardSetCount(state: TargetingState): string {
  if (state.min === state.max) return `${state.max}`;
  if (state.min === 0) return `up to ${state.max}`;
  return `${state.min} to ${state.max}`;
}

// targetingQuestion is the question line's sentence.
function targetingQuestion(state: TargetingState, batch: BatchPosition | null = null): string {
  const name = state.card.name;
  if (isCardSetPick(state)) {
    // #2394: a choice made as the card resolves, not a target. The
    // label is the card's own question ("Finale of Revelation — untap
    // up to five lands"); the sentence says where to answer it.
    return `${state.label || name} — click ${cardSetCount(state)} of the highlighted permanents`;
  }
  if (state.choiceID && state.choiceKind === "retarget") {
    // #1196, CR 115.7: the source is a spell that is RESOLVING, not a
    // trigger, and it is pointing something else somewhere else.
    // "Deflecting Swat triggered" would be the wrong sentence.
    return `${name} — click ${state.label || "a new target"}`;
  }
  if (state.choiceID) {
    // #3817: one of many — twenty Blood Artists and two deaths are forty
    // prompts, each looking like the last. Say how many.
    const times = batch && batch.total > 1 ? ` ${batch.total} times` : "";
    return `${name} triggered${times} — click ${state.label || "a target"}`;
  }
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
  // #3817: which of a batch of the same trigger this prompt answers.
  const batch = view ? batchPosition(view, state.choiceID) : null;
  if (batch) parts.push(`trigger ${batch.index} of ${batch.total}`);
  if (!state.choiceID && state.steps.length > 1) {
    parts.push(`step ${state.step + 1} of ${state.steps.length}`);
  }
  const count = legalTargetCount(state);
  if (count >= 0) parts.push(isCardSetPick(state) ? `${count} to choose from` : `${count} legal`);
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
  // #2394: a card-set pick answered on the board offers the same
  // choice as the list. Called with the choice's ID.
  onShowList?: (choiceID: string) => void;
  // #3265: the same ability on this source has `count` more instances
  // waiting for a target, and the player may answer them all with this
  // pick. #3817: `card` is the reach over every copy of the card (the
  // other instances, and the sources in all), offered as a second
  // choice when more than one source has them waiting.
  share?: {
    scope: ShareScope;
    count: number;
    card?: { count: number; sources: number };
    onScope: (scope: ShareScope) => void;
  };
}

// sameTargetSourceLabel is the source scope's button: this source's
// instances, this prompt included.
export function sameTargetSourceLabel(name: string, n: number): string {
  return `Same target for this ${name}'s ${n} triggers`;
}

// sameTargetCardLabel is the card scope's button (#3817): every copy's
// instances, this prompt included.
export function sameTargetCardLabel(name: string, n: number): string {
  return `Same target for all ${n} ${name} triggers`;
}

// shareRow is the dock's row of share choices: none for a lone
// instance, the source scope when this source repeats, and the card
// scope when more than one source of the card has instances waiting.
// Each button turns its scope on, or off again when it is the one on.
function shareRow(name: string, share: NonNullable<TargetingHandlers["share"]>): DockAction[] {
  const cardOK = (share.card?.sources ?? 0) > 1 && (share.card?.count ?? 0) > 0;
  const scope = effectiveShareScope(share.scope, cardOK);
  const press = (s: "source" | "card") => () => share.onScope(scope === s ? "off" : s);
  const row: DockAction[] = [];
  if (share.count > 0) {
    row.push({
      id: "same-target",
      label: sameTargetSourceLabel(name, share.count + 1),
      title:
        scope === "source"
          ? `your pick also answers this ${name}'s other triggers, where it is legal`
          : `use one target for this ${name}'s triggers`,
      pressed: scope === "source",
      alignEnd: true,
      onPress: press("source"),
    });
  }
  if (cardOK) {
    row.push({
      id: "same-target-card",
      label: sameTargetCardLabel(name, share.card!.count + 1),
      title:
        scope === "card"
          ? `your pick also answers every ${name}'s triggers, where it is legal`
          : `use one target for the triggers of every ${name} you control`,
      pressed: scope === "card",
      alignEnd: true,
      onPress: press("card"),
    });
  }
  return row;
}

export function targetingRequest(
  state: TargetingState,
  view: GameView | null | undefined,
  h: TargetingHandlers,
): DockRequest {
  const multi = isMultiPick(state);
  const confirmable = canConfirm(state);
  const cardSet = isCardSetPick(state);
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
            : cardSet
              ? "confirm your choice"
              : "confirm targets",
        keyShortcuts: "Enter",
        cap: "⏎",
        onPress: h.onDone,
      }
    : null;
  // A pending choice's target cannot be refused: the trigger needs one.
  // A card-set pick offers its list instead (#2394): the same question,
  // in the modal's grid, for a player who would rather read names.
  const showList = h.onShowList;
  const share = multi ? undefined : h.share;
  const rowOf = share ? shareRow(state.card.name, share) : [];
  const choiceID = state.choiceID;
  const secondary: DockAction[] = choiceID
    ? cardSet && showList
      ? [
          {
            id: "show-list",
            label: L.showAsList,
            title: "choose from a list instead of on the board",
            onPress: () => showList(choiceID),
          },
        ]
      : []
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
    label: cardSet ? L.chooseOnBoard(state.card.name) : L.selectTarget(state.card.name),
    tag: cardSet ? "choose" : "target",
    tone: "gold",
    live: true,
    question: targetingQuestion(state, view ? batchPosition(view, state.choiceID) : null),
    detail: targetingDetail(state, view) || undefined,
    primary,
    secondary,
    ...(rowOf.length > 0 ? { row: rowOf } : {}),
  };
}

export interface InsufficientManaHandlers {
  onAutoTap: () => void;
  onCastAnyway: () => void;
  onCancel: () => void;
}

export interface InsufficientManaOptions {
  // The refused cast's payload already had `auto_tap` (every clicked
  // cast under strict since ADR 0118 §1): the planner found no plan, so
  // the preview would find none either.
  autoTapped?: boolean;
}

// insufficientManaRequest: a strict cast was refused for mana (the
// server's insufficient_mana, with what it is missing).
//
// ADR 0118 §1 (amends ADR 0111 PR 4): a refusal of a cast that was
// already auto-tapped drops "Auto-tap & cast"; Cancel becomes the
// primary and Cast anyway stays a secondary with no key. A refusal
// without `auto_tap` (no client surface sends one now, but the server
// accepts it) keeps all three buttons. The labels are unchanged (ADR
// 0111 §10); only Cast anyway's title is.
export function insufficientManaRequest(
  missing: readonly string[],
  cardName: string | undefined,
  h: InsufficientManaHandlers,
  opts: InsufficientManaOptions = {},
): DockRequest {
  const cancel: DockAction = {
    id: "cancel",
    label: "Cancel",
    title: "don't cast it",
    keyShortcuts: "Escape",
    cap: "Esc",
    onPress: h.onCancel,
  };
  const castAnyway: DockAction = {
    id: "cast-anyway",
    label: "Cast anyway",
    title: "cast it without paying its mana cost; the game log says so",
    onPress: h.onCastAnyway,
  };
  const base = {
    rank: "flow" as const,
    label: "insufficient mana",
    tag: "mana",
    tone: "gold" as const,
    question: cardName ? `Insufficient mana for ${cardName}` : "Insufficient mana",
    detail: missing.length > 0 ? `missing ${missing.join(" ")}` : undefined,
  };
  if (opts.autoTapped) {
    return { ...base, primary: cancel, secondary: [castAnyway] };
  }
  return {
    ...base,
    primary: {
      id: "auto-tap",
      label: "Auto-tap & cast",
      title: "preview which of your sources would pay, then cast",
      keyShortcuts: "Enter",
      cap: "⏎",
      onPress: h.onAutoTap,
    },
    secondary: [cancel, castAnyway],
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
  return L.castWithoutPaying(cardName);
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
      label: L.cast,
      title: "cast it without paying its mana cost; the game log says so",
      onPress: h.onCast,
    },
    secondary: [
      {
        id: "cancel",
        label: L.cancel,
        title: "don't cast it",
        keyShortcuts: "Escape",
        cap: "Esc",
        onPress: h.onCancel,
      },
    ],
  };
}
