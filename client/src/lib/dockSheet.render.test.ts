// @vitest-environment jsdom
//
// dockSheet.render.test.ts — ADR 0111 Delivery PR 6 (S56, #1958). The
// sheet shell: a picker that needs room is drawn in a panel that grows
// up out of the action dock (owner decision 2), inside the request's
// one non-modal dialog, with its confirm and cancel in the dock's action
// bar. These pin the shell itself, through the X picker (a flow sheet):
// the dialog's name, the bar's buttons, Enter / Escape through the
// dock's one handler, focus into the sheet, minimise and restore, and
// that no backdrop is left.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import { get } from "svelte/store";

vi.mock("./api", async (orig) => ({
  ...(await orig<typeof import("./api")>()),
  fetchAutoTapPreview: vi.fn(async () => ({ ok: true, cost: "{X}{R}", plan: [] })),
}));

import XCostModal from "./components/board/XCostModal.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { _resetForTests as resetDock, pushDockRequest } from "./dock";
import { _resetForTests as resetModals, foreignModalOpen, modalOpen } from "./modalLayers";
import { defaultSettings, settings } from "./settings";
import type { CardView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";
import {
  barButton,
  barPrimary,
  barSecondaries,
  dockDialog,
  dockRegion,
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
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  settings.set(defaultSettings());
  resetDock();
  resetModals();
});
afterEach(cleanup);

const blaze = {
  instance_id: "blaze-1",
  name: "Blaze",
  owner: "me",
  controller: "me",
  mana_cost: "{X}{R}",
} as CardView;

function mount(card: CardView | null = blaze) {
  const calls = { confirm: [] as number[], cancel: 0 };
  const r = render(
    DockHarness as never,
    {
      component: XCostModal,
      props: {
        gameID: "g1",
        card,
        suggestedMax: 3,
        onConfirm: (x: number) => calls.confirm.push(x),
        onCancel: () => calls.cancel++,
      },
    } as never,
  );
  return { ...r, calls };
}

describe("a dock sheet (the X picker)", () => {
  it("opens as a sheet inside the dock, in a non-modal dialog named as the modal was", () => {
    mount();
    const dialog = dockDialog();
    expect(dialog).not.toBeNull();
    expect(dockRegion()!.contains(dialog)).toBe(true);
    expect(dialog!.getAttribute("aria-label")).toBe("Choose X for Blaze");
    expect(dialog!.hasAttribute("aria-modal")).toBe(false);
    const sheet = sheetPanel();
    expect(sheet).not.toBeNull();
    expect(dialog!.contains(sheet)).toBe(true);
    // The body is the picker's.
    expect(sheet!.querySelector('input[aria-label="X value"]')).not.toBeNull();
    // No backdrop and no centred modal anywhere.
    expect(document.querySelector(".prompt-backdrop")).toBeNull();
    expect(document.querySelector(".prompt-modal")).toBeNull();
    expect(document.querySelector('[aria-modal="true"]')).toBeNull();
  });

  it("puts the confirm in the corner and Cancel on the left of the dock's bar", () => {
    const { calls } = mount();
    const primary = barPrimary()!;
    expect(nameOf(primary)).toBe("Cast with X = 3");
    expect(primary.getAttribute("aria-keyshortcuts")).toBe("Enter");
    const [cancel] = barSecondaries();
    expect(nameOf(cancel)).toBe("Cancel");
    expect(cancel.getAttribute("aria-keyshortcuts")).toBe("Escape");
    // next and Pass turn give way to it.
    expect(document.querySelector(".dock-btn.next")).toBeNull();
    expect(document.querySelector(".dock-btn.pass-turn")).toBeNull();

    click(primary);
    expect(calls.confirm).toEqual([3]);
    click(barButton("Cancel")!);
    expect(calls.cancel).toBe(1);
  });

  it("takes focus into the sheet when it opens", () => {
    mount();
    const sheet = sheetPanel()!;
    // The X field is the sheet's marked focus.
    expect(document.activeElement).toBe(sheet.querySelector('input[aria-label="X value"]'));
  });

  it("answers Enter and Escape through the dock's one handler", () => {
    const { calls } = mount();
    (document.activeElement as HTMLElement | null)?.blur();
    pressKey("Enter");
    expect(calls.confirm).toEqual([3]);
    pressKey("Escape");
    expect(calls.cancel).toBe(1);
  });

  it("answers Enter and Escape in its own field, where the dock's handler stands down", () => {
    const { calls } = mount();
    const input = sheetPanel()!.querySelector<HTMLInputElement>('input[aria-label="X value"]')!;
    pressKey("Enter", input);
    expect(calls.confirm).toEqual([3]);
    pressKey("Escape", input);
    expect(calls.cancel).toBe(1);
  });

  it("stands the global shortcuts down, but not the dock's keys", () => {
    mount();
    expect(get(modalOpen)).toBe(true);
    expect(get(foreignModalOpen)).toBe(false);
  });

  it("minimises to a restore chip and keeps the request open, and restores", () => {
    const { calls } = mount();
    const min = dockDialog()!.querySelector<HTMLButtonElement>("button.sheet-min")!;
    expect(min.getAttribute("aria-label")).toBe("minimise");
    click(min);
    flushSync();
    // The sheet is folded away; the dialog and its bar stay.
    expect(sheetPanel()).toBeNull();
    const chip = dockDialog()!.querySelector<HTMLButtonElement>("button.sheet-restore")!;
    expect(chip).not.toBeNull();
    expect(nameOf(chip)).toBe("restore: Choose X for Blaze");
    expect(nameOf(barPrimary()!)).toBe("Cast with X = 3");
    // The bar still answers while it is minimised.
    pressKey("Escape");
    expect(calls.cancel).toBe(1);
    click(chip);
    flushSync();
    expect(sheetPanel()).not.toBeNull();
    expect(dockDialog()!.querySelector("button.sheet-restore")).toBeNull();
  });

  it("gives way to a stronger request and comes back with its body", () => {
    mount();
    const input = sheetPanel()!.querySelector('input[aria-label="X value"]');
    expect(input).not.toBeNull();
    // A pending choice outranks a flow: its request is drawn instead,
    // and the X picker's body leaves the dock.
    const choice = pushDockRequest({
      rank: "choice",
      label: "Reclamation Sage — destroy target artifact?",
      primary: { id: "yes", label: "Yes", onPress: () => {} },
      secondary: [],
    });
    flushSync();
    expect(dockDialog()!.getAttribute("aria-label")).toBe(
      "Reclamation Sage — destroy target artifact?",
    );
    expect(dockRegion()!.querySelector('input[aria-label="X value"]')).toBeNull();
    choice.close();
    flushSync();
    // The same node, back in the sheet.
    expect(sheetPanel()!.querySelector('input[aria-label="X value"]')).toBe(input);
  });

  it("closes with the picker, and next comes back", () => {
    const r = mount();
    const live = r.props as unknown as { props: Record<string, unknown> };
    r.setProps({ props: { ...live.props, card: null } } as never);
    expect(dockDialog()).toBeNull();
    expect(document.querySelector(".dock-btn.next")).not.toBeNull();
    expect(get(modalOpen)).toBe(false);
  });
});
