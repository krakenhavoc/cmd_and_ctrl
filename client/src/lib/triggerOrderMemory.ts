// triggerOrderMemory.ts — "Use this order again" (#3016, ADR 0018's
// eighth amendment). All pure.
//
// A seat that orders the same triggers again and again may keep the
// order: the trigger-order prompt's toggle sends `remember` with the
// answer, and the server answers the next prompt for the same set of
// abilities itself, as an ADR 0127 automatic answer (the dock's notice
// with Undo and Ask me next time). Only an order that keeps each
// ability's copies together is a ranking the server can reuse.

import type { LogEvent, PendingChoiceView, ReplacementOptionView } from "./protocol";

/** Every trigger-order prompt's auto_answer_key starts with this. */
export const TRIGGER_ORDER_KEY_PREFIX = "trigger_order|";

/** isTriggerOrderKey reports whether an auto_answer_key is a remembered trigger order's. */
export function isTriggerOrderKey(key: string | undefined | null): boolean {
  return !!key && key.startsWith(TRIGGER_ORDER_KEY_PREFIX);
}

/** The toggle's hint while the order splits one card's triggers apart. */
export const SPLIT_ORDER_HINT =
  "Only an order that keeps each card's triggers together can be reused.";

/** The toggle's hint otherwise. */
export const REUSE_ORDER_HINT =
  "Order these abilities the same way for you the next time they trigger together this game. You can undo it, or turn it off.";

/** canRememberTriggerOrder reports whether the prompt shows the toggle: a keyed trigger_order. */
export function canRememberTriggerOrder(c: PendingChoiceView | null | undefined): boolean {
  return c?.kind === "trigger_order" && !!c.auto_answer_key;
}

/**
 * keepsRowsTogether reports whether `order` (option IDs, in resolution
 * order, possibly not yet complete) keeps each row's copies together: no
 * row comes back after another row. Options with no row never split.
 */
export function keepsRowsTogether(
  order: readonly string[],
  options: readonly ReplacementOptionView[],
): boolean {
  const rowOf = new Map(options.map((o) => [o.id, o.row ?? 0]));
  const done = new Set<number>();
  let last = 0;
  for (const id of order) {
    const row = rowOf.get(id) ?? 0;
    if (row === 0) continue;
    if (row === last) continue;
    if (done.has(row)) return false;
    done.add(row);
    last = row;
  }
  return true;
}

/**
 * triggerOrderAnswer is the resolve_choice params for a trigger order:
 * `{order}`, plus `remember` on a keyed prompt. The toggle counts only
 * on an order it can take, so a split order sends `remember: false`
 * (which also forgets an order remembered before).
 */
export function triggerOrderAnswer(
  c: PendingChoiceView,
  order: readonly string[],
  remember: boolean,
): { order: string[]; remember?: boolean } {
  if (!canRememberTriggerOrder(c)) return { order: [...order] };
  return {
    order: [...order],
    remember: remember && keepsRowsTogether(order, c.trigger_options ?? []),
  };
}

/**
 * triggerOrderNoticeText is the automatic-answer notice's line for a
 * remembered order: "Ordered for you: A, then B".
 */
export function triggerOrderNoticeText(e: LogEvent): string {
  return e.label ? `Ordered for you: ${e.label}` : "Your triggers were ordered for you";
}
