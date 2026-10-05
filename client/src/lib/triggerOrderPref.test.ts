import { describe, it, expect } from "vitest";
import { newTriggerOrderPrefState, triggerOrderPrefToSend } from "./triggerOrderPref";
import type { GameView } from "./protocol";

function view(flag: boolean | undefined, eliminated = false): GameView {
  return {
    seats: [{ id: "me", trigger_order_always_ask: flag, eliminated }, { id: "opp" }],
  } as unknown as GameView;
}

describe("triggerOrderPrefToSend", () => {
  it("sends nothing when the server already agrees", () => {
    const st = newTriggerOrderPrefState();
    expect(triggerOrderPrefToSend(st, view(undefined), "me", false)).toBeNull();
    expect(triggerOrderPrefToSend(st, view(true), "me", true)).toBeNull();
  });

  it("sends the setting when the server disagrees, once", () => {
    const st = newTriggerOrderPrefState();
    expect(triggerOrderPrefToSend(st, view(undefined), "me", true)).toBe(true);
    // Same stale frame again: not re-sent.
    expect(triggerOrderPrefToSend(st, view(undefined), "me", true)).toBeNull();
    // The server caught up.
    expect(triggerOrderPrefToSend(st, view(true), "me", true)).toBeNull();
  });

  it("re-sends after the server and the setting agreed and then split", () => {
    const st = newTriggerOrderPrefState();
    triggerOrderPrefToSend(st, view(undefined), "me", true);
    triggerOrderPrefToSend(st, view(true), "me", true);
    // Server lost it (restart from an older snapshot): reconcile again.
    expect(triggerOrderPrefToSend(st, view(undefined), "me", true)).toBe(true);
    // Turning it off.
    triggerOrderPrefToSend(st, view(true), "me", true);
    expect(triggerOrderPrefToSend(st, view(true), "me", false)).toBe(false);
  });

  it("never sends for a spectator, a missing view or an eliminated seat", () => {
    const st = newTriggerOrderPrefState();
    expect(triggerOrderPrefToSend(st, view(undefined), null, true)).toBeNull();
    expect(triggerOrderPrefToSend(st, null, "me", true)).toBeNull();
    expect(triggerOrderPrefToSend(st, view(undefined, true), "me", true)).toBeNull();
    expect(triggerOrderPrefToSend(st, view(undefined), "stranger", true)).toBeNull();
  });
});
