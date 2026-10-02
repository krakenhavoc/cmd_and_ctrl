import { describe, expect, it, vi } from "vitest";

import { emit, subscribe, type TutorialEvent } from "./tutorialBus";

describe("tutorialBus (ADR 0076 §2.5)", () => {
  it("an emit with no subscriber is a harmless no-op", () => {
    expect(() => emit("hand-hovered")).not.toThrow();
  });

  it("delivers each of the three events to a subscriber, in order", () => {
    const seen: TutorialEvent[] = [];
    const off = subscribe((n) => seen.push(n));
    emit("ability-menu-opened");
    emit("hand-hovered");
    emit("pile-hovered");
    off();
    expect(seen).toEqual(["ability-menu-opened", "hand-hovered", "pile-hovered"]);
  });

  it("retains nothing: a late subscriber does not hear an earlier emit", () => {
    emit("ability-menu-opened");
    const h = vi.fn();
    const off = subscribe(h);
    expect(h).not.toHaveBeenCalled();
    off();
  });

  it("stops delivering after unsubscribe", () => {
    const h = vi.fn();
    const off = subscribe(h);
    off();
    emit("pile-hovered");
    expect(h).not.toHaveBeenCalled();
  });

  it("serves several subscribers, and one may unsubscribe mid-emit", () => {
    const a = vi.fn();
    const b = vi.fn();
    const offA = subscribe(() => {
      a();
      offA();
    });
    const offB = subscribe(b);
    emit("hand-hovered");
    emit("hand-hovered");
    expect(a).toHaveBeenCalledTimes(1);
    expect(b).toHaveBeenCalledTimes(2);
    offB();
  });
});
