// settleGate.test.ts — #4115: a layout change waits for the pointer to
// be at rest, so a board never moves under a click already started.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createSettleGate } from "./settleGate";

beforeEach(() => vi.useFakeTimers());
afterEach(() => vi.useRealTimers());

function setup() {
  const target = new EventTarget();
  const applied: string[] = [];
  const gate = createSettleGate<string>((v) => applied.push(v), { quietMs: 150, target });
  const fire = (type: string) => target.dispatchEvent(new Event(type));
  return { gate, applied, fire };
}

describe("createSettleGate", () => {
  it("applies a request once the pointer has been still for quietMs", () => {
    const { gate, applied } = setup();
    gate.request("a");
    vi.advanceTimersByTime(149);
    expect(applied).toEqual([]);
    vi.advanceTimersByTime(1);
    expect(applied).toEqual(["a"]);
  });

  it("restarts the wait while the pointer moves", () => {
    const { gate, applied, fire } = setup();
    gate.request("a");
    vi.advanceTimersByTime(100);
    fire("pointermove");
    vi.advanceTimersByTime(100);
    expect(applied).toEqual([]);
    vi.advanceTimersByTime(50);
    expect(applied).toEqual(["a"]);
  });

  it("holds the change while a button is down, then waits after release", () => {
    const { gate, applied, fire } = setup();
    fire("pointerdown");
    gate.request("a");
    vi.advanceTimersByTime(5000);
    expect(applied).toEqual([]);
    fire("pointerup");
    vi.advanceTimersByTime(149);
    expect(applied).toEqual([]);
    vi.advanceTimersByTime(1);
    expect(applied).toEqual(["a"]);
  });

  it("treats a cancelled pointer as released", () => {
    const { gate, applied, fire } = setup();
    fire("pointerdown");
    gate.request("a");
    fire("pointercancel");
    vi.advanceTimersByTime(150);
    expect(applied).toEqual(["a"]);
  });

  it("applies only the latest of several requests", () => {
    const { gate, applied } = setup();
    gate.request("a");
    vi.advanceTimersByTime(100);
    gate.request("b");
    vi.advanceTimersByTime(150);
    expect(applied).toEqual(["b"]);
  });

  it("drops a pending request when destroyed", () => {
    const { gate, applied, fire } = setup();
    gate.request("a");
    gate.destroy();
    fire("pointermove");
    vi.advanceTimersByTime(1000);
    expect(applied).toEqual([]);
  });
});
