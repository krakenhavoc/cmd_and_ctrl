// @vitest-environment jsdom
//
// specialActionX.render.test.ts — #5249. A turn-face-up row whose cost has
// an {X} (Aurelia's Vindicator's disguise {X}{3}{W}, CR 107.3d) is not
// sent straight from the menu: the Board finds the row the payload names,
// asks for X with the X picker, held under the server's `x_max`, and sends
// the payload again with `x_value`.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";

const previewed = vi.hoisted(() => ({ calls: 0 }));
vi.mock("./api", async (orig) => ({
  ...(await orig<typeof import("./api")>()),
  fetchAutoTapPreview: vi.fn(async () => {
    previewed.calls++;
    return { ok: true, cost: "", plan: [] };
  }),
}));

import XCostModal from "./components/board/XCostModal.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { defaultSettings, settings } from "./settings";
import { L } from "./labels";
import type { CardView, GameView } from "./protocol";
import { specialActionItems } from "./contextMenu.logic";
import { specialActionNeedsX, withX } from "./specialActionX";
import { render, cleanup, flushSync } from "./test/render.svelte";
import { barPrimary, nameOf, sheetPanel } from "./test/dockView";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  settings.set(defaultSettings());
  resetDock();
  resetModals();
  previewed.calls = 0;
});
afterEach(cleanup);

const vindicator = {
  instance_id: "av-1",
  name: "Aurelia's Vindicator",
  owner: "me",
  controller: "me",
  face_down: true,
  special_actions: [
    {
      kind: "turn_face_up",
      label: "Turn face up {X}{3}{W}",
      cost: "{X}{3}{W}",
      available: true,
      route: "disguise",
      demands_x: true,
      x_max: 2,
    },
  ],
} as unknown as CardView;

const morph = {
  instance_id: "mo-1",
  name: "Willbender",
  owner: "me",
  controller: "me",
  special_actions: [
    {
      kind: "turn_face_up",
      label: "Turn face up {1}{U}",
      cost: "{1}{U}",
      available: true,
      route: "morph",
    },
  ],
} as unknown as CardView;

const view = { battlefield: { cards: [vindicator, morph] } } as unknown as GameView;

// The payload the card's menu row sends (the popover and the admin menu
// build it with the same function).
function rowParams(card: CardView): Record<string, unknown> {
  const item = specialActionItems(card, "me")[0];
  return item.action!.params as Record<string, unknown>;
}

describe("a special action's {X}", () => {
  it("finds the row that demands X, and only that one", () => {
    const found = specialActionNeedsX(view, rowParams(vindicator));
    expect(found?.card.instance_id).toBe("av-1");
    expect(found?.row.x_max).toBe(2);
    expect(specialActionNeedsX(view, rowParams(morph))).toBeNull();
    // A payload that already carries its X is sent as it is.
    expect(specialActionNeedsX(view, withX(rowParams(vindicator), 1))).toBeNull();
  });

  it("sends the chosen X with the rest of the payload", () => {
    expect(withX(rowParams(vindicator), 2)).toEqual({
      card_id: "av-1",
      kind: "turn_face_up",
      route: "disguise",
      strict: true,
      auto_tap: true,
      x_value: 2,
    });
  });

  it("opens the picker at the server's ceiling, holds X under it, and confirms the chosen X", async () => {
    const row = specialActionNeedsX(view, rowParams(vindicator))!.row;
    const confirmed: number[] = [];
    render(
      DockHarness as never,
      {
        component: XCostModal,
        props: {
          gameID: "g1",
          card: vindicator,
          suggestedMax: row.x_max,
          costLabel: row.cost,
          maxX: row.x_max,
          maxXUnit: "mana",
          confirmVerb: L.xPickerTurnFaceUp,
          onConfirm: (x: number) => confirmed.push(x),
          onCancel: () => {},
        },
      } as never,
    );
    flushSync();
    expect(nameOf(barPrimary()!)).toBe("Turn face up with X = 2");
    expect(sheetPanel()!.textContent).toContain(L.xPickerManaCeiling(2));
    const input = sheetPanel()!.querySelector('input[aria-label="X value"]') as HTMLInputElement;
    expect(input.getAttribute("max")).toBe("2");
    input.value = "9";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    expect(nameOf(barPrimary()!)).toBe("Turn face up with X = 2");
    input.value = "1";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    barPrimary()!.click();
    expect(confirmed).toEqual([1]);
    await Promise.resolve();
    // The auto-tap preview prices casts and abilities, never this.
    expect(previewed.calls).toBe(0);
  });
});
