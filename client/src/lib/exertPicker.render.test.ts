// @vitest-environment jsdom
//
// ADR 0130 §7 (owner decision 1) — the Choose attackers picker's Exert
// toggle. Shown only on a row whose creature the server says may be
// exerted as it attacks, off by default, and sent with the declaration
// only when on.

import { afterEach, beforeEach, describe, expect, it } from "vitest";

import AttackDeclarationModal from "./components/board/AttackDeclarationModal.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { L } from "./labels";
import { legalActionsOf } from "./legalActions";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { cleanup, click, render } from "./test/render.svelte";

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

const creature = (id: string, name: string): CardView => ({
  instance_id: id,
  name,
  owner: "me",
  controller: "me",
  type_line: "Creature — Human Warrior",
});

function snap(): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat("me", "Me", 0), seat("bob", "Bob", 1)],
    battlefield: zone("battlefield", undefined, [
      creature("avenger", "Oketra's Avenger"),
      creature("bear", "Bear"),
    ]),
    stack: zone("stack", undefined, []),
    exile: zone("exile", undefined, []),
    turn: {
      seq: 1,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "combat",
      step: "declare_attackers",
    },
    mulligans_open: false,
    legal_actions: {
      pass: true,
      sources: {
        avenger: { kinds: ["attack"], moves: 2, attack_targets: ["bob"], exert_on_attack: true },
        bear: { kinds: ["attack"], moves: 1, attack_targets: ["bob"] },
      },
    },
  };
}

function mount(view: GameView) {
  const confirmed: Array<{ attackers: string[]; exert: string[] }> = [];
  const r = render(
    DockHarness as never,
    {
      component: AttackDeclarationModal,
      view,
      props: {
        view,
        viewerID: "me",
        defenderSeatID: "bob",
        legalGate: legalActionsOf(view),
        onConfirm: (attackers: string[], _locked: string[], exert: string[]) =>
          confirmed.push({ attackers, exert }),
        onCancel: () => {},
      },
    } as never,
  );
  return { container: r.container, confirmed };
}

const toggles = (c: HTMLElement): HTMLButtonElement[] => [
  ...c.querySelectorAll<HTMLButtonElement>('[role="switch"]'),
];
const attackButton = (c: HTMLElement): HTMLButtonElement =>
  [...c.querySelectorAll<HTMLButtonElement>(".dock-bar button.request-primary")].find((b) =>
    (b.textContent ?? "").includes("Attack with"),
  )!;

describe("the picker's Exert toggle (ADR 0130 §7)", () => {
  it("is on the exertable row only, named by the contract label, and off", () => {
    const { container } = mount(snap());
    const t = toggles(container);
    expect(t).toHaveLength(1);
    expect(t[0]!.getAttribute("aria-label")).toBe(L.exertAttacker("Oketra's Avenger"));
    expect(t[0]!.getAttribute("aria-checked")).toBe("false");
  });

  it("sends no exert unless it is turned on", () => {
    const { container, confirmed } = mount(snap());
    click(attackButton(container));
    expect(confirmed).toEqual([{ attackers: ["avenger", "bear"], exert: [] }]);
  });

  it("sends the exert when it is on", () => {
    const { container, confirmed } = mount(snap());
    click(toggles(container)[0]!);
    expect(toggles(container)[0]!.getAttribute("aria-checked")).toBe("true");
    click(attackButton(container));
    expect(confirmed).toEqual([{ attackers: ["avenger", "bear"], exert: ["avenger"] }]);
  });

  it("can't exert a row that is not attacking", () => {
    const { container, confirmed } = mount(snap());
    click(toggles(container)[0]!);
    const avengerRow = [...container.querySelectorAll<HTMLButtonElement>('[role="checkbox"]')].find(
      (b) => (b.textContent ?? "").includes("Oketra"),
    )!;
    click(avengerRow);
    expect(toggles(container)[0]!.disabled).toBe(true);
    click(attackButton(container));
    expect(confirmed).toEqual([{ attackers: ["bear"], exert: [] }]);
  });
});
