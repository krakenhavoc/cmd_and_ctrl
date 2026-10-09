// @vitest-environment jsdom
//
// xCeiling.render.test.ts — #2581. A spell's printed "X can't be
// greater than <count>" reaches the X picker as the card's `x_max`
// (and the live preview's): the server refuses a larger X, so the
// picker opens at no more than it, holds the field under it, and says
// why.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";

const preview = vi.hoisted(() => ({ xMax: undefined as number | undefined }));
vi.mock("./api", async (orig) => ({
  ...(await orig<typeof import("./api")>()),
  fetchAutoTapPreview: vi.fn(async () => ({
    ok: true,
    cost: "{X}{G}{G}",
    plan: [],
    x_max: preview.xMax,
  })),
}));

import XCostModal from "./components/board/XCostModal.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { defaultSettings, settings } from "./settings";
import type { CardView } from "./protocol";
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
  preview.xMax = undefined;
});
afterEach(cleanup);

const openTheWay = {
  instance_id: "way-1",
  name: "Open the Way",
  owner: "me",
  controller: "me",
  mana_cost: "{X}{G}{G}",
  x_max: 4,
} as CardView;

function mount(xCeiling: number | undefined, suggestedMax = 7) {
  return render(
    DockHarness as never,
    {
      component: XCostModal,
      props: {
        gameID: "g1",
        card: openTheWay,
        suggestedMax,
        xCeiling,
        onConfirm: () => {},
        onCancel: () => {},
      },
    } as never,
  );
}

function field(): HTMLInputElement {
  return sheetPanel()!.querySelector('input[aria-label="X value"]') as HTMLInputElement;
}

async function settle(): Promise<void> {
  for (let i = 0; i < 3; i++) {
    await Promise.resolve();
    flushSync();
  }
}

describe("the X picker under a printed ceiling", () => {
  it("opens at the ceiling rather than the mana estimate, and says why", () => {
    preview.xMax = 4;
    mount(4);
    expect(nameOf(barPrimary()!)).toBe("Cast with X = 4");
    expect(field().getAttribute("max")).toBe("4");
    expect(sheetPanel()!.textContent).toContain("X can't be greater than 4 right now");
  });

  it("holds a typed X under the ceiling", () => {
    preview.xMax = 4;
    mount(4, 1);
    const input = field();
    input.value = "9";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    expect(nameOf(barPrimary()!)).toBe("Cast with X = 4");
  });

  it("takes the preview's fresher count when it is smaller", async () => {
    preview.xMax = 2;
    mount(4);
    await settle();
    expect(nameOf(barPrimary()!)).toBe("Cast with X = 2");
    expect(field().getAttribute("max")).toBe("2");
  });

  it("allows only X = 0 at a ceiling of 0", async () => {
    preview.xMax = 0;
    mount(0);
    await settle();
    expect(nameOf(barPrimary()!)).toBe("Cast with X = 0");
  });

  it("leaves a card with no ceiling alone", () => {
    mount(undefined);
    expect(nameOf(barPrimary()!)).toBe("Cast with X = 7");
    expect(field().hasAttribute("max")).toBe(false);
    expect(sheetPanel()!.textContent).not.toContain("can't be greater than");
  });
});
