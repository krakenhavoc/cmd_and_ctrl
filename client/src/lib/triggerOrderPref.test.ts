import { describe, it, expect } from "vitest";
import {
  newTriggerOrderPrefState,
  serverTriggerOrder,
  triggerOrderPrefRefused,
  triggerOrderPrefSent,
  triggerOrderPrefToSend,
} from "./triggerOrderPref";
import type { GameView } from "./protocol";

function view(mode: "always" | "never" | undefined, eliminated = false): GameView {
  return {
    seats: [
      {
        id: "me",
        trigger_order: mode,
        trigger_order_always_ask: mode === "always" ? true : undefined,
        eliminated,
      },
      { id: "opp" },
    ],
  } as unknown as GameView;
}

describe("serverTriggerOrder", () => {
  it("reads the mode, and the #1530 boolean from an older server", () => {
    expect(serverTriggerOrder({})).toBe("when_it_matters");
    expect(serverTriggerOrder({ trigger_order: "never" })).toBe("never");
    expect(serverTriggerOrder({ trigger_order: "always", trigger_order_always_ask: true })).toBe(
      "always",
    );
    expect(serverTriggerOrder({ trigger_order_always_ask: true })).toBe("always");
    expect(serverTriggerOrder({ trigger_order: "sometimes" })).toBe("when_it_matters");
  });
});

describe("triggerOrderPrefToSend", () => {
  it("sends nothing when the server already agrees", () => {
    const st = newTriggerOrderPrefState();
    expect(triggerOrderPrefToSend(st, view(undefined), "me", "when_it_matters")).toBeNull();
    expect(triggerOrderPrefToSend(st, view("always"), "me", "always")).toBeNull();
    expect(triggerOrderPrefToSend(st, view("never"), "me", "never")).toBeNull();
  });

  it("sends the mode when the server disagrees, once", () => {
    const st = newTriggerOrderPrefState();
    expect(triggerOrderPrefToSend(st, view(undefined), "me", "never")).toBe("never");
    triggerOrderPrefSent(st, "frame-1");
    // Same stale frame again: not re-sent.
    expect(triggerOrderPrefToSend(st, view(undefined), "me", "never")).toBeNull();
    // The server caught up.
    expect(triggerOrderPrefToSend(st, view("never"), "me", "never")).toBeNull();
  });

  it("re-sends after the server and the setting agreed and then split", () => {
    const st = newTriggerOrderPrefState();
    triggerOrderPrefToSend(st, view(undefined), "me", "always");
    triggerOrderPrefToSend(st, view("always"), "me", "always");
    // Server lost it (restart from an older snapshot): reconcile again.
    expect(triggerOrderPrefToSend(st, view(undefined), "me", "always")).toBe("always");
    // Back to the default.
    triggerOrderPrefToSend(st, view("always"), "me", "always");
    expect(triggerOrderPrefToSend(st, view("always"), "me", "when_it_matters")).toBe(
      "when_it_matters",
    );
  });

  it("retries a send the server refused", () => {
    const st = newTriggerOrderPrefState();
    expect(triggerOrderPrefToSend(st, view(undefined), "me", "never")).toBe("never");
    triggerOrderPrefSent(st, "frame-1");
    // Somebody else's refusal is not ours.
    expect(triggerOrderPrefRefused(st, { replyTo: "frame-2" })).toBe(false);
    expect(triggerOrderPrefRefused(st, null)).toBe(false);
    expect(triggerOrderPrefToSend(st, view(undefined), "me", "never")).toBeNull();
    // Ours: the next frame sends again.
    expect(triggerOrderPrefRefused(st, { replyTo: "frame-1" })).toBe(true);
    expect(triggerOrderPrefToSend(st, view(undefined), "me", "never")).toBe("never");
  });

  it("retries a send that never left the client", () => {
    const st = newTriggerOrderPrefState();
    expect(triggerOrderPrefToSend(st, view(undefined), "me", "always")).toBe("always");
    triggerOrderPrefSent(st, null);
    expect(triggerOrderPrefToSend(st, view(undefined), "me", "always")).toBe("always");
  });

  it("never sends for a spectator, a missing view or an eliminated seat", () => {
    const st = newTriggerOrderPrefState();
    expect(triggerOrderPrefToSend(st, view(undefined), null, "always")).toBeNull();
    expect(triggerOrderPrefToSend(st, null, "me", "always")).toBeNull();
    expect(triggerOrderPrefToSend(st, view(undefined, true), "me", "always")).toBeNull();
    expect(triggerOrderPrefToSend(st, view(undefined), "stranger", "always")).toBeNull();
  });
});
