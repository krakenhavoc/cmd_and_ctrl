// @vitest-environment jsdom
//
// payLifeForManaPayUnless.render.test.ts — ADR 0131 §2 (#2531), PR 2: a
// mana pay_unless (ward {B}) answered in the real action dock offers
// "Pay with 2 life" for a {B} the viewer's K'rrik lets them pay for, and
// the answer goes out as resolve_choice {apply: true, phyrexian_life: 1}.
// The plain Pay still sends no claim, and a viewer with 1 life is not
// offered the row (CR 119.4).

import { describe, it, expect, afterEach, beforeEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import type { ActionType, GameView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}
beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  resetDock();
  resetModals();
});
afterEach(cleanup);

const snap = (life: number): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      { id: "me", name: "Me", life },
      { id: "them", name: "Them", life: 40 },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-1",
        kind: "pay_unless",
        chooser: "me",
        from_player: "me",
        pay_cost: "{B}",
        phyrexian_symbols: 1,
        phyrexian_granted: 1,
        reason: "Ward {B}",
        options: [],
      },
    ],
  }) as unknown as GameView;

function mount(life: number) {
  const sent: { type: ActionType; params?: unknown }[] = [];
  const r = render(
    ChoiceDockHarness as never,
    {
      snap: snap(life),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  const button = (re: RegExp) =>
    [...r.container.querySelectorAll("button")].find((b) => re.test(b.textContent ?? ""));
  return { ...r, sent, button };
}

describe("pay_unless in the dock — Pay with life", () => {
  it("sends the claim", () => {
    const p = mount(20);
    click(p.button(/Pay with 2 life/)!);
    expect(p.sent).toHaveLength(1);
    expect(p.sent[0].type).toBe("resolve_choice");
    expect(p.sent[0].params).toEqual({ choice_id: "choice-1", apply: true, phyrexian_life: 1 });
  });

  it("the plain Pay sends no claim", () => {
    const p = mount(20);
    click(p.button(/^\s*Pay \{B\}/)!);
    expect(p.sent[0].params).toEqual({ choice_id: "choice-1", apply: true });
  });

  it("offers no life answer at 1 life", () => {
    const p = mount(1);
    expect(p.button(/Pay with 2 life/)).toBeUndefined();
    expect(p.button(/Don't pay/)).toBeDefined();
  });
});
