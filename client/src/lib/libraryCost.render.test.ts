// @vitest-environment jsdom
//
// libraryCost.render.test.ts — ADR 0109 §7 (#1902) and owner decision 3.
// The two cost components the activator chooses nothing for ("Discard a
// card at random", "Exile the top N cards of your library") are
// confirmed in a dock sheet (CostConfirmModal); "Put a card from your
// hand on top of your library" is picked with the discard picker under
// its own verb and small print. Both are sheets in the action dock
// (ADR 0111 PR 6): confirm and Cancel in the bar, Enter / Escape through
// the dock's one key handler, and a short hand or library disables the
// confirm and says why (CR 118.3).

import { describe, it, expect, afterEach, beforeEach } from "vitest";

import CostConfirmModal from "./components/board/CostConfirmModal.svelte";
import DiscardCostModal from "./components/board/DiscardCostModal.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { defaultSettings, settings } from "./settings";
import type { CardView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";
import {
  barPrimary,
  barSecondaries,
  dockDialog,
  dockRegion,
  nameOf,
  pressKey,
  sheetPanel,
} from "./test/dockView";
import {
  costConfirmLines,
  costConfirmNote,
  needsCostConfirm,
  randomDiscardCount,
  randomDiscardPool,
  topCostNote,
} from "./libraryCost";

beforeEach(() => {
  settings.set(defaultSettings());
  resetDock();
  resetModals();
});
afterEach(cleanup);

const card = (id: string, name: string, extra: Partial<CardView> = {}): CardView =>
  ({ instance_id: id, name, owner: "me", controller: "me", ...extra }) as CardView;

const pyromancy = {
  discard_cost_n: 1,
  discard_cost_label: "a card at random",
  discard_cost_random: true,
};
const tactician = { library_exile_cost_n: 4 };

describe("libraryCost", () => {
  it("knows which costs are confirmed rather than picked", () => {
    expect(needsCostConfirm(pyromancy)).toBe(true);
    expect(needsCostConfirm(tactician)).toBe(true);
    expect(needsCostConfirm({ discard_cost_n: 1, discard_cost_label: "a card" })).toBe(false);
    expect(randomDiscardCount({ discard_cost_n: 2, discard_cost_random: true })).toBe(2);
    expect(randomDiscardCount({ discard_cost_n: 2 })).toBe(0);
  });

  it("writes one line per component, with the shortfall", () => {
    expect(costConfirmLines(pyromancy, 3, 40)).toEqual([{ text: "Discard a card at random." }]);
    expect(costConfirmLines(pyromancy, 0, 40)).toEqual([
      { text: "Discard a card at random.", short: "You have no cards in hand to discard." },
    ]);
    expect(costConfirmLines(tactician, 0, 40)).toEqual([
      { text: "Exile the top four cards of your library." },
    ]);
    expect(costConfirmLines(tactician, 0, 3)).toEqual([
      {
        text: "Exile the top four cards of your library.",
        short: "Your library has only three cards.",
      },
    ]);
    expect(costConfirmLines({ library_exile_cost_n: 1 }, 0, 0)).toEqual([
      { text: "Exile the top card of your library.", short: "Your library is empty." },
    ]);
    expect(costConfirmNote(pyromancy)).toContain("random");
  });

  it("counts the hand a random discard draws from", () => {
    const seat = {
      hand: { kind: "hand", count: 3, cards: [card("a", "A"), card("b", "B"), card("src", "S")] },
    };
    expect(randomDiscardPool(seat as never, "src", [])).toBe(2);
    expect(randomDiscardPool(seat as never, "src", ["a"])).toBe(1);
    expect(randomDiscardPool(undefined, "src", [])).toBe(0);
  });
});

describe("the cost confirm is a dock sheet", () => {
  function mount(lines: ReturnType<typeof costConfirmLines>) {
    const calls = { confirmed: 0, cancelled: 0 };
    render(
      DockHarness as never,
      {
        component: CostConfirmModal,
        props: {
          card: card("pyro", "Pyromancy"),
          lines,
          note: costConfirmNote(pyromancy),
          onConfirm: () => calls.confirmed++,
          onCancel: () => calls.cancelled++,
        },
      } as never,
    );
    flushSync();
    return calls;
  }

  it("opens in region actions, named for the card, and answers", () => {
    const calls = mount(costConfirmLines(pyromancy, 3, 40));
    const dialog = dockDialog()!;
    expect(dockRegion()!.contains(dialog)).toBe(true);
    expect(dialog.getAttribute("aria-label")).toBe("Pyromancy");
    expect(dialog.hasAttribute("aria-modal")).toBe(false);
    expect(sheetPanel()!.textContent).toContain("Discard a card at random.");
    const primary = barPrimary()!;
    expect(nameOf(primary)).toBe("Pay and activate");
    expect(primary.disabled).toBe(false);
    click(primary);
    expect(calls.confirmed).toBe(1);
    click(barSecondaries().find((b) => nameOf(b) === "Cancel")!);
    expect(calls.cancelled).toBe(1);
  });

  it("answers Enter and Escape through the dock", () => {
    const calls = mount(costConfirmLines(tactician, 0, 40));
    pressKey("Enter");
    expect(calls.confirmed).toBe(1);
    pressKey("Escape");
    expect(calls.cancelled).toBe(1);
  });

  it("disables the confirm and says why when the cost can't be paid", () => {
    const calls = mount(costConfirmLines(tactician, 0, 3));
    expect(barPrimary()!.disabled).toBe(true);
    expect(sheetPanel()!.querySelector(".error")?.textContent).toContain("only three cards");
    pressKey("Enter");
    expect(calls.confirmed).toBe(0);
  });
});

describe("the put-on-top picker is the discard picker with its own words", () => {
  it("asks for a card and sends the pick", () => {
    const confirmed: string[][] = [];
    render(
      DockHarness as never,
      {
        component: DiscardCostModal,
        props: {
          card: card("pen", "Penance"),
          options: [card("bolt", "Lightning Bolt"), card("ox", "Ox")],
          need: 1,
          label: "a card from your hand on top of your library",
          verb: "Put on top",
          note: topCostNote,
          onConfirm: (ids: string[]) => confirmed.push(ids),
          onCancel: () => {},
        },
      } as never,
    );
    flushSync();
    expect(dockDialog()!.getAttribute("aria-label")).toBe("Penance");
    expect(sheetPanel()!.textContent).toContain("a card from your hand on top of your library");
    expect(barPrimary()!.disabled).toBe(true);
    click(sheetPanel()!.querySelector<HTMLElement>(".prompt-options .prompt-opt")!);
    expect(nameOf(barPrimary()!)).toBe("Put on top");
    click(barPrimary()!);
    expect(confirmed).toEqual([["bolt"]]);
  });
});
