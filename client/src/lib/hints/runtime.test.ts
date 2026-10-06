// @vitest-environment jsdom
//
// runtime.test.ts — closing the tip on screen (#2422, #2372). A card
// someone closed is marked seen even if the engine let go of it after
// it was drawn, and a press on the card is told apart from a press on
// the table.

import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { get } from "svelte/store";

import { defaultSettings, settings } from "../settings";
import { L } from "../labels";
import type { Hint } from "./hint";
import { _resetHintsForTests, activeTip, dismissTip, isInsideTip, setActiveTip } from "./runtime";

const DOCK: Hint = {
  id: "table.dock",
  version: 1,
  place: "table",
  order: 10,
  anchor: { label: L.actions },
  title: "Your controls",
  body: "A body.",
};

function showOnly(h: Hint): void {
  setActiveTip({
    hint: h,
    title: "Your controls",
    body: "A body.",
    ring: { left: 0, top: 0, width: 10, height: 10 },
    placement: { kind: "strip" },
    stripBottom: 6,
    reduceMotion: true,
  });
}

beforeEach(() => {
  settings.set(defaultSettings());
  _resetHintsForTests({ log: () => {} });
});

afterEach(() => {
  _resetHintsForTests();
  settings.set(defaultSettings());
});

describe("dismissTip", () => {
  it("marks the card on screen seen even when the engine no longer holds it", () => {
    // The engine never ticked (or let the hint go): only the card is up.
    showOnly(DOCK);
    dismissTip("got-it");
    expect(get(activeTip)).toBeNull();
    expect(get(settings).help.seen[DOCK.id]).toBe(DOCK.version);
  });

  it("Hide tips does the same and turns tips off", () => {
    showOnly(DOCK);
    dismissTip("hide");
    expect(get(activeTip)).toBeNull();
    expect(get(settings).help.seen[DOCK.id]).toBe(DOCK.version);
    expect(get(settings).help.tipsOff).toBe(true);
  });
});

describe("isInsideTip", () => {
  it("is true for the card and what is in it, and false elsewhere", () => {
    const card = document.createElement("aside");
    card.setAttribute("aria-label", L.tip);
    const button = document.createElement("button");
    card.append(button);
    const other = document.createElement("div");
    document.body.append(card, other);
    try {
      expect(isInsideTip(card)).toBe(true);
      expect(isInsideTip(button)).toBe(true);
      expect(isInsideTip(other)).toBe(false);
      expect(isInsideTip(null)).toBe(false);
      expect(isInsideTip(window)).toBe(false);
    } finally {
      card.remove();
      other.remove();
    }
  });
});
