// autoAnswerPref.ts — standing answers to repeated prompts (ADR 0127).
//
// A player may answer a repeated yes/no prompt once ("never pay for
// Rhystic Study", "always draw two off Consecrated Sphinx"). The rules
// live in the synced setting gameplay.autoAnswers; the SERVER answers
// with them (so it works with the browser closed), from its own copy of
// the seat's rules. This module is the client half, all pure:
//
//   - the reconcile (autoAnswersToSend): the triggerOrderPref.ts shape —
//     when the seat's view disagrees with the setting, send the list;
//   - the rule edits (withRule, withoutRule) and the "Remember this
//     answer" mapping from a pressed button to a rule;
//   - the notice (autoAnswerNotice): what the dock says after the server
//     answered for the viewer, and whether its Undo still works.

import type { DockRequest } from "./dock";
import { L } from "./labels";
import type { GameView, LogEvent, PendingChoiceView, PlayerView } from "./protocol";
import {
  MAX_AUTO_ANSWERS,
  canonicalJSON,
  syncedSubset,
  trimAutoAnswerText,
  utf8Length,
  type AutoAnswerRule,
  type Settings,
} from "./settings";

// ---- the reconcile ----------------------------------------------------------

export interface AutoAnswersPrefState {
  // The list this client last sent and has not yet seen reflected, as
  // canonical JSON. Stops a refused or slow action being re-sent every
  // frame.
  lastSent: string | null;
}

export function newAutoAnswersPrefState(): AutoAnswersPrefState {
  return { lastSent: null };
}

/** The wire form of a rule list: `{key, answer}`, sorted by key. */
export function wireRules(rules: readonly AutoAnswerRule[]): { key: string; answer: string }[] {
  return rules
    .map((r) => ({ key: r.key, answer: r.answer }))
    .sort((a, b) => (a.key < b.key ? -1 : a.key > b.key ? 1 : 0));
}

function seatRules(me: PlayerView): { key: string; answer: string }[] {
  return (me.auto_answers ?? [])
    .map((r) => ({ key: r.key, answer: r.answer }))
    .sort((a, b) => (a.key < b.key ? -1 : a.key > b.key ? 1 : 0));
}

/**
 * autoAnswersToSend returns the `rules` to send in set_auto_answers, or
 * null for nothing. Updates state. A viewer who is not seated
 * (spectator, unseated admin), an eliminated one, or a bot seat never
 * sends.
 */
export function autoAnswersToSend(
  state: AutoAnswersPrefState,
  view: GameView | null | undefined,
  viewerID: string | null,
  desired: readonly AutoAnswerRule[],
): { key: string; answer: string }[] | null {
  if (!view || !viewerID) return null;
  const me = view.seats.find((s) => s.id === viewerID);
  if (!me || me.eliminated || me.is_bot) return null;
  const want = wireRules(desired);
  const wantJSON = canonicalJSON(want);
  if (canonicalJSON(seatRules(me)) === wantJSON) {
    state.lastSent = null;
    return null;
  }
  if (state.lastSent === wantJSON) return null;
  state.lastSent = wantJSON;
  return want;
}

// ---- the rule edits ---------------------------------------------------------

// The account keeps at most 32 KiB of settings (ADR 0110 §4); a rule is
// refused if the synced settings would pass this, leaving room for the
// rest of the object.
export const AUTO_ANSWER_SETTINGS_BUDGET = 30 * 1024;

export type RuleRefusal = "too_many" | "too_large";

/** ruleFor is the rule for `key`, or undefined (Ask). */
export function ruleFor(
  rules: readonly AutoAnswerRule[],
  key: string | undefined,
): AutoAnswerRule | undefined {
  if (!key) return undefined;
  return rules.find((r) => r.key === key);
}

/**
 * withRule returns the list with `rule` set (replacing any rule for the
 * same key), or a refusal: a 101st rule (ADR 0127 §3), or a list that
 * would push the synced settings past AUTO_ANSWER_SETTINGS_BUDGET.
 * `settings` is the current settings, for the size check; omit it to
 * skip that check.
 */
export function withRule(
  rules: readonly AutoAnswerRule[],
  rule: AutoAnswerRule,
  settings?: Settings,
): AutoAnswerRule[] | RuleRefusal {
  const others = rules.filter((r) => r.key !== rule.key);
  if (others.length >= MAX_AUTO_ANSWERS) return "too_many";
  const next = [...others, rule];
  if (settings) {
    const synced = syncedSubset({
      ...settings,
      gameplay: { ...settings.gameplay, autoAnswers: next },
    });
    if (utf8Length(JSON.stringify(synced)) > AUTO_ANSWER_SETTINGS_BUDGET) return "too_large";
  }
  return next;
}

/** withoutRule returns the list with no rule for `key` (Ask). */
export function withoutRule(rules: readonly AutoAnswerRule[], key: string): AutoAnswerRule[] {
  return rules.filter((r) => r.key !== key);
}

/** The message for a refused rule. */
export function ruleRefusalText(why: RuleRefusal): string {
  return why === "too_many"
    ? `You have ${MAX_AUTO_ANSWERS} automatic answers already. Remove one in Settings → Gameplay to add another.`
    : "Your automatic answers are full. Remove one in Settings → Gameplay to add another.";
}

/**
 * rememberedRule is the rule "Remember this answer" sets when the player
 * presses Yes / Pay (Always) or No / Don't pay (Never) on `c`, or null
 * when the prompt can take none.
 */
export function rememberedRule(c: PendingChoiceView, apply: boolean): AutoAnswerRule | null {
  if (!c.auto_answer_key) return null;
  return {
    key: c.auto_answer_key,
    card: trimAutoAnswerText(c.auto_answer_card),
    prompt: trimAutoAnswerText(c.auto_answer_prompt || c.reason),
    answer: apply ? "always" : "never",
  };
}

/** canRemember reports whether a prompt shows "Remember this answer". */
export function canRemember(c: PendingChoiceView | null | undefined): boolean {
  if (!c?.auto_answer_key) return false;
  return c.kind === "pay_unless" || c.kind === "trigger_prompt" || c.kind === "confirm";
}

/**
 * askedByHandText is the line a prompt that has a rule but is asked
 * anyway shows, saying why (ADR 0127 §4), or null.
 */
export function askedByHandText(c: PendingChoiceView | null | undefined): string | null {
  switch (c?.asked_by_hand) {
    case "no_mana":
      return "Always pay: not enough mana, so you're asked.";
    case "empty_library":
      return "Always: your library is empty, so you're asked.";
    case "loop":
      return "A loop was spotted, so automatic answers are paused. You're asked.";
    case "undone":
      return "You undid the automatic answer, so you're asked this time.";
    default:
      return null;
  }
}

// ---- the notice -------------------------------------------------------------

/** What the dock shows after the server answered for the viewer. */
export interface AutoAnswerNotice {
  // The auto_answer log entry's seq: the notice's identity.
  seq: number;
  // "Rhystic Study: paid {1} for you".
  text: string;
  // The rule's key, for "Ask me next time".
  key: string;
  // True while the answer is still the viewer's top undo entry.
  undoable: boolean;
}

/** What the answer was, in the notice's words. */
function answerWords(e: LogEvent): string {
  switch (e.call) {
    case "pay":
      return e.label ? `paid ${e.label}` : "paid";
    case "dont_pay":
      return "didn't pay";
    case "yes":
      return "answered Yes";
    default:
      return "answered No";
  }
}

/**
 * newAutoAnswerEntries is the viewer's own auto_answer log entries with
 * a seq above `since`, oldest first.
 */
export function newAutoAnswerEntries(
  view: GameView | null | undefined,
  viewerID: string | null,
  since: number,
): LogEvent[] {
  if (!view || !viewerID) return [];
  const me = view.seats.find((s) => s.id === viewerID);
  if (!me) return [];
  return (view.log ?? []).filter(
    (e) => e.kind === "auto_answer" && e.seat === me.seat && e.seq > since && !!e.auto_answer_key,
  );
}

/**
 * autoAnswerNotice builds the notice for `entry`, an auto_answer line of
 * the viewer's. `cardName` is the asking card's name as the viewer sees
 * it ("" when hidden). The Undo works exactly while the room stamps this
 * entry's seq as the viewer's top undo entry (owner amendment
 * 2026-10-07).
 */
export function autoAnswerNotice(
  entry: LogEvent,
  cardName: string,
  view: GameView | null | undefined,
  viewerID: string | null,
): AutoAnswerNotice {
  const me = view?.seats.find((s) => s.id === viewerID);
  return {
    seq: entry.seq,
    text: `${cardName || "A card"}: ${answerWords(entry)} for you`,
    key: entry.auto_answer_key ?? "",
    undoable: (me?.undo_auto_answer ?? 0) === entry.seq,
  };
}

/** How long the notice stays, unless another request takes the dock. */
export const AUTO_ANSWER_NOTICE_MS = 6000;

/**
 * autoAnswerNoticeRequest is the notice as a dock request (ADR 0127 §6):
 * a status, not a question. It is a `step` request, so it does not take
 * the action bar, binds no key and blocks nothing; any other request
 * outranks it. Undo is greyed with "Someone has acted since" once the
 * answer is no longer the viewer's top undo entry.
 */
export function autoAnswerNoticeRequest(
  notice: AutoAnswerNotice,
  onUndo: () => void,
  onAskNextTime: () => void,
): DockRequest {
  return {
    rank: "step",
    label: L.automaticAnswer,
    tag: "automatic",
    tone: "plain",
    question: notice.text,
    live: true,
    row: [
      {
        id: "auto-answer-undo",
        label: L.undoAutomaticAnswer,
        disabled: !notice.undoable,
        title: notice.undoable
          ? "Take this answer back and answer by hand"
          : "Someone has acted since",
        onPress: onUndo,
      },
      {
        id: "auto-answer-ask",
        label: L.askMeNextTime,
        title: "Remove this rule; the answer just given stands",
        onPress: onAskNextTime,
      },
    ],
  };
}
