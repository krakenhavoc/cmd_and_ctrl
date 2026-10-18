// settleGate holds a layout change until the viewer's pointer is at rest
// (ADR 0077 as amended 2026-10-18, #4115).
//
// expansion.ts's invariant is that the table never moves under a click
// the viewer has started. A stack item that targets an avatar-card seat
// is a table event, yet it should open that seat's board. The gate makes
// it safe: a requested value is applied only once no pointer button is
// down and the pointer has not moved for `quietMs`. A newer request
// replaces a pending one, so only the latest value ever lands.

export interface SettleGate<T> {
  /** Ask for `value` to be applied once the pointer is at rest. */
  request(value: T): void;
  /** Drop any pending value and stop listening. */
  destroy(): void;
}

export interface SettleGateOptions {
  /** How long the pointer must be still. Default 150ms. */
  quietMs?: number;
  /** Where pointer events are heard. Default `window`. */
  target?: EventTarget;
}

export function createSettleGate<T>(
  apply: (value: T) => void,
  { quietMs = 150, target = window }: SettleGateOptions = {},
): SettleGate<T> {
  let pending: { value: T } | null = null;
  let pressed = false;
  let timer: ReturnType<typeof setTimeout> | null = null;

  function schedule(): void {
    if (timer !== null) clearTimeout(timer);
    timer = null;
    if (!pending || pressed) return;
    timer = setTimeout(() => {
      timer = null;
      if (!pending || pressed) return;
      const { value } = pending;
      pending = null;
      apply(value);
    }, quietMs);
  }

  const onDown = (): void => {
    pressed = true;
    schedule();
  };
  const onUp = (): void => {
    pressed = false;
    schedule();
  };
  const onMove = (): void => {
    if (pending) schedule();
  };
  const opts = { capture: true, passive: true } as const;
  target.addEventListener("pointerdown", onDown, opts);
  target.addEventListener("pointerup", onUp, opts);
  target.addEventListener("pointercancel", onUp, opts);
  target.addEventListener("pointermove", onMove, opts);

  return {
    request(value: T): void {
      pending = { value };
      schedule();
    },
    destroy(): void {
      if (timer !== null) clearTimeout(timer);
      timer = null;
      pending = null;
      target.removeEventListener("pointerdown", onDown, opts);
      target.removeEventListener("pointerup", onUp, opts);
      target.removeEventListener("pointercancel", onUp, opts);
      target.removeEventListener("pointermove", onMove, opts);
    },
  };
}
