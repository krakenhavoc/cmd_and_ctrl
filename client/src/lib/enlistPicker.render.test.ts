// @vitest-environment jsdom
//
// #5658 — the Choose attackers picker's enlist choice (CR 702.154). Shown
// only on a row whose creature the server says may enlist another, empty
// by default, offering only creatures that are not checked to attack, and
// sent with the declaration only when chosen.

import { afterEach, beforeEach, describe, expect, it } from "vitest";

import AttackDeclarationModal from "./components/board/AttackDeclarationModal.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { L } from "./labels";
import { legalActionsOf } from "./legalActions";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

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

const creature = (id: string, name: string, power: number): CardView => ({
  instance_id: id,
  name,
  owner: "me",
  controller: "me",
  type_line: "Creature — Elk",
  power,
});

function snap(): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat("me", "Me", 0), seat("bob", "Bob", 1)],
    battlefield: zone("battlefield", undefined, [
      creature("crusher", "Barkweave Crusher", 2),
      creature("elk", "Elk", 3),
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
        crusher: {
          kinds: ["attack"],
          moves: 2,
          attack_targets: ["bob"],
          enlist_on_attack: ["elk"],
        },
        elk: { kinds: ["attack"], moves: 1, attack_targets: ["bob"] },
      },
    },
  };
}

function mount(view: GameView) {
  const confirmed: Array<{ attackers: string[]; enlist: Record<string, string> }> = [];
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
        onConfirm: (
          attackers: string[],
          _locked: string[],
          _exert: string[],
          _life: number,
          enlist: Record<string, string>,
        ) => confirmed.push({ attackers, enlist }),
        onCancel: () => {},
      },
    } as never,
  );
  return { container: r.container, confirmed };
}

const picks = (c: HTMLElement): HTMLSelectElement[] => [
  ...c.querySelectorAll<HTMLSelectElement>("select.enlist-pick"),
];
const row = (c: HTMLElement, name: string): HTMLButtonElement =>
  [...c.querySelectorAll<HTMLButtonElement>('[role="checkbox"]')].find((b) =>
    (b.textContent ?? "").includes(name),
  )!;
const attackButton = (c: HTMLElement): HTMLButtonElement =>
  [...c.querySelectorAll<HTMLButtonElement>(".dock-bar button.request-primary")].find((b) =>
    (b.textContent ?? "").includes("Attack with"),
  )!;
const optionTexts = (s: HTMLSelectElement): string[] =>
  [...s.options].map((o) => (o.textContent ?? "").trim());

describe("the picker's enlist choice (#5658)", () => {
  it("is on the enlist row only, named by the contract label, and empty", () => {
    const { container } = mount(snap());
    const p = picks(container);
    expect(p).toHaveLength(1);
    expect(p[0]!.getAttribute("aria-label")).toBe(L.enlistFor("Barkweave Crusher"));
    expect(p[0]!.value).toBe("");
  });

  it("offers nothing while every candidate is checked to attack, and sends no enlist", () => {
    const { container, confirmed } = mount(snap());
    expect(optionTexts(picks(container)[0]!)).toEqual(["No enlist"]);
    expect(picks(container)[0]!.disabled).toBe(true);
    click(attackButton(container));
    expect(confirmed).toEqual([{ attackers: ["crusher", "elk"], enlist: {} }]);
  });

  it("offers a creature left at home and sends the choice", () => {
    const { container, confirmed } = mount(snap());
    click(row(container, "Elk"));
    flushSync();
    const p = picks(container)[0]!;
    expect(optionTexts(p)).toEqual(["No enlist", "Enlist Elk (+3)"]);
    p.value = "elk";
    p.dispatchEvent(new Event("change", { bubbles: true }));
    flushSync();
    click(attackButton(container));
    expect(confirmed).toEqual([{ attackers: ["crusher"], enlist: { crusher: "elk" } }]);
  });

  it("drops the choice when the enlisted creature is sent to attack after all", () => {
    const { container, confirmed } = mount(snap());
    click(row(container, "Elk"));
    flushSync();
    const p = picks(container)[0]!;
    p.value = "elk";
    p.dispatchEvent(new Event("change", { bubbles: true }));
    flushSync();
    click(row(container, "Elk"));
    flushSync();
    click(attackButton(container));
    expect(confirmed).toEqual([{ attackers: ["crusher", "elk"], enlist: {} }]);
  });
});
