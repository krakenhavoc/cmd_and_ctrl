// @vitest-environment jsdom
//
// #4198 — the attack dock's "mana or life per Phyrexian symbol" question,
// as markup: the picker's inline stepper and the Pay the attack tax
// sheet. The pure rules are in attackLife.test.ts. Both are sheets in the
// action dock (ADR 0111 PR 6), so each is mounted beside a real
// ActionDock (DockHarness) and confirmed with the dock bar's primary.

import { afterEach, beforeEach, describe, expect, it } from "vitest";

import AttackDeclarationModal from "./components/board/AttackDeclarationModal.svelte";
import AttackLifeModal from "./components/board/AttackLifeModal.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import type { AttackTargetView, CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { cleanup, click, render } from "./test/render.svelte";
import { nameOf } from "./test/dockView";
import { L } from "./labels";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}
beforeEach(() => {
  (globalThis as Record<string, unknown>).ResizeObserver ??= FakeObserver;
  resetDock();
  resetModals();
});
afterEach(cleanup);

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, name: string, n: number): PlayerView {
  return {
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id, []),
    hand: zone("hand", id, []),
    graveyard: zone("graveyard", id, []),
    command: zone("command", id, []),
    commander_damage: {},
    life_history: [],
  };
}

const NORN: AttackTargetView = { kind: "player", id: "bob", tax: "{W/P}", tax_phyrexian: 1 };

function snap(rows: AttackTargetView[]): GameView {
  const bf: CardView[] = ["Bear A", "Bear B", "Bear C"].map((name, i) => ({
    instance_id: `bear${i + 1}`,
    name,
    owner: "me",
    controller: "me",
    type_line: "Creature — Bear",
  }));
  return {
    id: "g1",
    state: "active",
    seats: [seat("me", "Me", 0), seat("bob", "Bob", 1)],
    battlefield: zone("battlefield", undefined, bf),
    stack: zone("stack", undefined, []),
    exile: zone("exile", undefined, []),
    turn: {
      seq: 1,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "combat",
      step: "declare_attackers",
      attack_targets: rows,
    },
    mulligans_open: false,
  };
}

const text = (el: Element | null | undefined): string =>
  el?.textContent?.replace(/\s+/g, " ").trim() ?? "";

const primary = (c: HTMLElement): HTMLButtonElement =>
  c.querySelector<HTMLButtonElement>(".dock-bar button.request-primary")!;
const stepper = (c: HTMLElement): HTMLElement | null =>
  c.querySelector(`[aria-label="${L.attackPayWithLife}"]`);
const plus = (c: HTMLElement): HTMLButtonElement =>
  c.querySelector<HTMLButtonElement>('button[aria-label="one more symbol"]')!;
const minus = (c: HTMLElement): HTMLButtonElement =>
  c.querySelector<HTMLButtonElement>('button[aria-label="one fewer symbol"]')!;

describe("the attackers picker under a Phyrexian tax", () => {
  type Sent = { attackers: string[]; life: number };
  function mount(rows: AttackTargetView[], life = 40) {
    const sent: Sent[] = [];
    const view = snap(rows);
    view.seats[0]!.life = life;
    const r = render(
      DockHarness as never,
      {
        component: AttackDeclarationModal,
        view,
        props: {
          view,
          viewerID: "me",
          defenderSeatID: "bob",
          life,
          onConfirm: (attackers: string[], _locked: string[], _exert: string[], l: number) =>
            sent.push({ attackers, life: l }),
          onCancel: () => {},
        },
      } as never,
    );
    return { container: r.container, sent };
  }

  it("draws the stepper with every checked attacker's symbol, defaulting to mana", () => {
    const { container } = mount([NORN]);
    expect(stepper(container)).not.toBeNull();
    expect(text(container.querySelector(".count"))).toBe("0 of 3");
    expect(nameOf(primary(container))).toBe("Attack with 3");
  });

  it("sends the symbols the player buys with life", () => {
    const { container, sent } = mount([NORN]);
    click(plus(container));
    click(plus(container));
    expect(text(container.querySelector(".count"))).toBe("2 of 3");
    expect(text(container.querySelector(".price"))).toBe("4 life");
    click(primary(container));
    expect(sent).toEqual([{ attackers: ["bear1", "bear2", "bear3"], life: 2 }]);
  });

  it("caps the claim at what the life total buys (CR 119.4)", () => {
    const { container } = mount([NORN], 3);
    click(plus(container));
    expect(text(container.querySelector(".count"))).toBe("1 of 3");
    expect(plus(container).disabled).toBe(true);
    expect(text(container.querySelector(".status"))).toContain("You have 3 life");
  });

  it("claims nothing by default, and nothing for a plain tax", () => {
    const first = mount([NORN]);
    click(primary(first.container));
    expect(first.sent[0]!.life).toBe(0);
    cleanup();
    resetDock();
    resetModals();
    const plain = mount([{ kind: "player", id: "bob", tax: "{2}" }]);
    expect(stepper(plain.container)).toBeNull();
    click(primary(plain.container));
    expect(plain.sent[0]!.life).toBe(0);
  });

  it("shrinks the claim with the checked set", () => {
    const { container, sent } = mount([NORN]);
    click(plus(container));
    click(plus(container));
    click(plus(container));
    const rows = [...container.querySelectorAll<HTMLButtonElement>('[role="checkbox"]')];
    click(rows[0]!);
    click(rows[1]!);
    expect(text(container.querySelector(".count"))).toBe("1 of 1");
    click(primary(container));
    expect(sent).toEqual([{ attackers: ["bear3"], life: 1 }]);
  });
});

describe("the Pay the attack tax sheet", () => {
  function mount(symbols: number, life = 40) {
    const answers: number[] = [];
    const r = render(
      DockHarness as never,
      {
        component: AttackLifeModal,
        view: snap([NORN]),
        props: {
          prompt: { symbols, each: "{W/P}", attackers: symbols, verb: `Attack with ${symbols}` },
          life,
          onConfirm: (n: number) => answers.push(n),
          onCancel: () => {},
        },
      } as never,
    );
    return { container: r.container, answers };
  }

  it("asks, and the confirm names the life it spends", () => {
    const { container, answers } = mount(2);
    expect(nameOf(primary(container))).toBe("Attack with 2");
    click(plus(container));
    expect(nameOf(primary(container))).toBe("Attack with 2 for 2 life");
    click(primary(container));
    expect(answers).toEqual([1]);
  });

  it("confirms with mana when the stepper is left alone", () => {
    const { container, answers } = mount(1);
    click(primary(container));
    expect(answers).toEqual([0]);
    click(minus(container));
  });
});
