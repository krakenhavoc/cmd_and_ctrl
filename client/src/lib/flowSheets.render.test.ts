// @vitest-environment jsdom
//
// flowSheets.render.test.ts — ADR 0111 Delivery PR 6 (S56, #1958). Two
// pickers a player opens from the dock become sheets that grow up out
// of it: the auto-tap preview (from the insufficient-mana request's
// "Auto-tap & cast") and the attack picker (the attack row's "Choose
// attackers…"). For each: a sheet inside region "actions", in a
// non-modal dialog that keeps the modal's name; its confirm in the
// bar's corner and Cancel on its left; Enter confirms and Escape
// cancels through the dock's one handler; no backdrop is left.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";

vi.mock("./api", async (orig) => ({
  ...(await orig<typeof import("./api")>()),
  fetchAutoTapPreview: vi.fn(async () => ({
    ok: true,
    cost: "{1}{G}",
    plan: ["forest-1", "forest-2"],
    sources: [],
  })),
}));

import AutoTapPreviewModal from "./components/board/AutoTapPreviewModal.svelte";
import AttackDeclarationModal from "./components/board/AttackDeclarationModal.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { defaultSettings, settings } from "./settings";
import type { CardView, GameView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";
import {
  barPrimary,
  barSecondaries,
  dockDialog,
  dockRegion,
  dockTestView,
  nameOf,
  pressKey,
  sheetPanel,
} from "./test/dockView";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

beforeEach(() => {
  (globalThis as Record<string, unknown>).ResizeObserver ??= FakeObserver;
  settings.set(defaultSettings());
  resetDock();
  resetModals();
});
afterEach(cleanup);

const tick = () => new Promise((r) => setTimeout(r, 0));

function expectFlowSheet(name: string): HTMLElement {
  const dlg = dockDialog()!;
  expect(dlg).not.toBeNull();
  expect(dockRegion()!.contains(dlg)).toBe(true);
  expect(dlg.getAttribute("aria-label")).toBe(name);
  expect(dlg.dataset.rank).toBe("flow");
  expect(dlg.hasAttribute("aria-modal")).toBe(false);
  expect(sheetPanel()).not.toBeNull();
  expect(document.querySelector(".prompt-backdrop")).toBeNull();
  expect(document.querySelector(".prompt-modal")).toBeNull();
  return dlg;
}

const forest = (id: string): CardView =>
  ({ instance_id: id, name: "Forest", owner: "me", controller: "me" }) as CardView;

describe("the auto-tap preview, as a sheet", () => {
  async function mount() {
    const calls = { confirm: [] as string[][], cancel: 0 };
    const view = dockTestView({
      battlefield: {
        kind: "battlefield",
        count: 2,
        cards: [forest("forest-1"), forest("forest-2")],
      },
    } as Partial<GameView>);
    render(
      DockHarness as never,
      {
        component: AutoTapPreviewModal,
        view,
        props: {
          gameID: "g1",
          snap: view,
          cardID: "bears",
          onConfirm: (locked: string[]) => calls.confirm.push(locked),
          onCancel: () => calls.cancel++,
        },
      } as never,
    );
    await tick();
    flushSync();
    return calls;
  }

  it("opens as a sheet named 'Auto-tap & cast' with the plan in it", async () => {
    await mount();
    expectFlowSheet("Auto-tap & cast");
    expect(sheetPanel()!.querySelectorAll(".src-row")).toHaveLength(2);
    expect(sheetPanel()!.querySelector(".prompt-src")?.textContent).toBe("{1}{G}");
  });

  it("has Cast in the corner and Cancel on its left, and each answers", async () => {
    const calls = await mount();
    expect(nameOf(barPrimary()!)).toBe("Cast");
    expect(barSecondaries().map(nameOf)).toEqual(["Cancel"]);
    // Lock a source: the confirm sends it.
    click(sheetPanel()!.querySelector<HTMLButtonElement>(".lock-btn")!);
    await tick();
    flushSync();
    click(barPrimary()!);
    expect(calls.confirm).toEqual([["forest-1"]]);
    click(barSecondaries()[0]);
    expect(calls.cancel).toBe(1);
  });

  it("casts on Enter and cancels on Escape", async () => {
    const calls = await mount();
    (document.activeElement as HTMLElement | null)?.blur();
    pressKey("Enter");
    expect(calls.confirm).toEqual([[]]);
    pressKey("Escape");
    expect(calls.cancel).toBe(1);
  });
});

describe("the attack picker, as a sheet", () => {
  const bear = (id: string): CardView =>
    ({
      instance_id: id,
      name: `Bear ${id}`,
      owner: "me",
      controller: "me",
      type_line: "Creature — Bear",
      power: 2,
      toughness: 2,
    }) as CardView;

  function mount() {
    const calls = { confirm: [] as string[][], cancel: 0 };
    const base = dockTestView();
    const view = {
      ...base,
      battlefield: { kind: "battlefield", count: 2, cards: [bear("a"), bear("b")] },
      turn: {
        ...base.turn,
        phase: "combat",
        step: "declare_attackers",
        attack_targets: [{ kind: "player", id: "opp" }],
      },
    } as unknown as GameView;
    render(
      DockHarness as never,
      {
        component: AttackDeclarationModal,
        view,
        props: {
          view,
          viewerID: "me",
          defenderSeatID: "opp",
          onConfirm: (ids: string[]) => calls.confirm.push(ids),
          onCancel: () => calls.cancel++,
        },
      } as never,
    );
    return calls;
  }

  it("opens as a sheet named 'Choose attackers', every eligible creature checked", () => {
    mount();
    expectFlowSheet("Choose attackers");
    const rows = sheetPanel()!.querySelectorAll('[role="checkbox"][aria-checked="true"]');
    expect(rows).toHaveLength(2);
    expect(dockDialog()!.querySelector(".prompt-count")?.textContent?.trim()).toBe(
      "2 / 2 attacking",
    );
  });

  it("confirms the picked set from the bar, on Enter too, and cancels on Escape", () => {
    const calls = mount();
    expect(nameOf(barPrimary()!)).toBe("Attack with 2");
    expect(barSecondaries().map(nameOf)).toEqual(["Cancel"]);
    click(sheetPanel()!.querySelectorAll<HTMLButtonElement>('[role="checkbox"]')[1]);
    expect(nameOf(barPrimary()!)).toBe("Attack with 1");
    (document.activeElement as HTMLElement | null)?.blur();
    pressKey("Enter");
    expect(calls.confirm).toEqual([["a"]]);
    pressKey("Escape");
    expect(calls.cancel).toBe(1);
  });
});
