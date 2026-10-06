// @vitest-environment jsdom
//
// #1623: a legend-rule prompt opened ChoicePromptModal as well as the
// board's targeting banner. The modal had no branch for the kind, so it
// fell through to the discard-from-a-revealed-hand copy ("Pick 1 card
// from opponent's revealed hand. opponent will discard your pick") and
// put a full-screen backdrop over the board the real answer is clicked
// on. The prompt that is answered on the board must not open the modal,
// and every OTHER choice owed to the viewer must still reach it — even
// when a board-answered one is queued ahead of it.

import { describe, it, expect, afterEach, beforeEach } from "vitest";

// ADR 0111 PR 5: the small kinds (pay_unless here) open in the action
// dock, so the prompt is mounted beside a real dock.
import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { isBoardAnsweredChoice, listFallback, showChoiceAsList } from "./boardAnsweredChoice";
import type { ActionType, GameView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);
beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= class {
    observe(): void {}
    unobserve(): void {}
    disconnect(): void {}
  };
  resetDock();
  resetModals();
  listFallback.set(new Set());
});

const legendRule = {
  id: "legend-1",
  kind: "legend_rule",
  chooser: "me",
  from_player: "me",
  count: 1,
  reason: "Legend rule — keep one Vivi Ornitier",
  pick_target_cards: ["vivi-a", "vivi-b"],
  options: [],
};

const payUnless = {
  id: "pay-1",
  kind: "pay_unless",
  chooser: "me",
  from_player: "me",
  pay_cost: "{2}",
  reason: "Pay {2} or Propaganda stops the attack",
  options: [],
};

const snapWith = (choices: unknown[]): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      { id: "me", name: "Me" },
      { id: "them", name: "Them" },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 6, active_seat: 1, priority_holder: 0, phase: "main2", step: "main" },
    mulligans_open: false,
    pending_choices: choices,
  }) as unknown as GameView;

function mount(choices: unknown[]): HTMLElement {
  return render(
    ChoiceDockHarness as never,
    {
      snap: snapWith(choices),
      viewerID: "me",
      sendAction: (_t: ActionType) => {},
      lastError: null,
    } as never,
  ).container;
}

describe("choices answered on the board", () => {
  it("names the four board-answered kinds and nothing else", () => {
    for (const k of ["pick_target", "legend_rule", "choose_protector", "retarget"]) {
      expect(isBoardAnsweredChoice(k), k).toBe(true);
    }
    for (const k of ["pay_unless", "choose_cards", "discard_from_hand", "reveal_pick"]) {
      expect(isBoardAnsweredChoice(k), k).toBe(false);
    }
  });

  it("does not open the modal for a legend-rule prompt", () => {
    const container = mount([legendRule]);
    expect(container.querySelector(".prompt-backdrop")).toBeNull();
    expect(container.textContent).not.toContain("revealed hand");
  });

  it("does not open the modal for choose_protector", () => {
    const container = mount([{ ...legendRule, id: "p-1", kind: "choose_protector" }]);
    expect(container.querySelector(".prompt-backdrop")).toBeNull();
  });

  // #2394: "untap up to five lands" is a choose_cards over tapped lands
  // on the battlefield. It is answered on the board, not in the grid.
  const land = (id: string) => ({
    instance_id: id,
    name: "Forest",
    owner: "me",
    controller: "me",
    type_line: "Basic Land — Forest",
    tapped: true,
  });
  const untapLands = {
    id: "untap-1",
    kind: "choose_cards",
    chooser: "me",
    from_player: "me",
    reason: "Finale of Revelation — untap up to five lands",
    choose_min: 0,
    choose_max: 5,
    options: [land("f1"), land("f2")],
  };
  function mountWithLands(choices: unknown[]): HTMLElement {
    const snap = snapWith(choices) as unknown as Record<string, unknown>;
    snap.battlefield = { kind: "battlefield", count: 2, cards: [land("f1"), land("f2")] };
    return render(
      ChoiceDockHarness as never,
      {
        snap: snap as unknown as GameView,
        viewerID: "me",
        sendAction: (_t: ActionType) => {},
        lastError: null,
      } as never,
    ).container;
  }

  it("does not open the grid for a card-set pick over permanents on the battlefield", () => {
    const container = mountWithLands([untapLands]);
    expect(container.querySelectorAll(".card-pick")).toHaveLength(0);
  });

  it("opens the grid for that pick once the player asks for the list", () => {
    showChoiceAsList("untap-1");
    const container = mountWithLands([untapLands]);
    expect(container.querySelectorAll(".card-pick")).toHaveLength(2);
    expect(container.textContent).toContain("Pick up to 5 of these, or none.");
  });

  it("still opens the grid when a candidate is not on the battlefield", () => {
    const container = mount([untapLands]);
    expect(container.querySelectorAll(".card-pick")).toHaveLength(2);
  });

  it("still opens the next owed choice when a board-answered one is ahead of it", () => {
    const container = mount([legendRule, payUnless]);
    expect(
      container.querySelector(
        '[role="dialog"][aria-label="Pay {2} or Propaganda stops the attack"]',
      ),
    ).not.toBeNull();
  });
});
