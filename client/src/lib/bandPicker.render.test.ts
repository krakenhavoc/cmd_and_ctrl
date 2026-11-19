// @vitest-environment jsdom
//
// ADR 0224 (#5723) — the Choose attackers picker's Band toggle. Shown on
// every row when a creature there has banding, off by default, sent only
// when two or more checked rows are on, and a band with two creatures
// without banding says why and holds the confirm (CR 702.22c).

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

const creature = (id: string, name: string, abilities: string[] = []): CardView => ({
  instance_id: id,
  name,
  owner: "me",
  controller: "me",
  type_line: "Creature — Test",
  abilities,
});

function snap(withHero = true): GameView {
  const cards = [creature("bear", "Bear"), creature("giant", "Giant")];
  if (withHero) cards.unshift(creature("hero", "Benalish Hero", ["banding"]));
  const sources: Record<string, unknown> = {};
  for (const c of cards)
    sources[c.instance_id] = { kinds: ["attack"], moves: 1, attack_targets: ["bob"] };
  return {
    id: "g1",
    state: "active",
    seats: [seat("me", "Me", 0), seat("bob", "Bob", 1)],
    battlefield: zone("battlefield", undefined, cards),
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
    legal_actions: { pass: true, sources },
  } as GameView;
}

function mount(view: GameView) {
  const confirmed: Array<{ attackers: string[]; band: string[] }> = [];
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
          _enlist: Record<string, string>,
          band: string[],
        ) => confirmed.push({ attackers, band }),
        onCancel: () => {},
      },
    } as never,
  );
  return { container: r.container, confirmed };
}

const bandToggle = (c: HTMLElement, name: string): HTMLButtonElement =>
  c.querySelector<HTMLButtonElement>(`[aria-label="${L.bandAttacker(name)}"]`)!;
const attackButton = (c: HTMLElement): HTMLButtonElement =>
  [...c.querySelectorAll<HTMLButtonElement>(".dock-bar button.request-primary")].find((b) =>
    (b.textContent ?? "").includes("Attack with"),
  )!;

describe("the picker's Band toggle (ADR 0224)", () => {
  it("is on every row when a creature there has banding, and off", () => {
    const { container } = mount(snap());
    for (const name of ["Benalish Hero", "Bear", "Giant"]) {
      const t = bandToggle(container, name);
      expect(t).not.toBeNull();
      expect(t.getAttribute("aria-checked")).toBe("false");
    }
  });

  it("is not offered when nothing there has banding", () => {
    const { container } = mount(snap(false));
    expect(bandToggle(container, "Bear")).toBeNull();
  });

  it("sends no band unless two rows are in it", () => {
    const { container, confirmed } = mount(snap());
    click(bandToggle(container, "Benalish Hero"));
    click(attackButton(container));
    expect(confirmed).toEqual([{ attackers: ["hero", "bear", "giant"], band: [] }]);
  });

  it("sends the band of the Hero and one creature without banding", () => {
    const { container, confirmed } = mount(snap());
    click(bandToggle(container, "Benalish Hero"));
    click(bandToggle(container, "Giant"));
    click(attackButton(container));
    expect(confirmed).toEqual([{ attackers: ["hero", "bear", "giant"], band: ["hero", "giant"] }]);
  });

  it("holds a band with two creatures without banding and says why", () => {
    const { container, confirmed } = mount(snap());
    click(bandToggle(container, "Benalish Hero"));
    click(bandToggle(container, "Bear"));
    click(bandToggle(container, "Giant"));
    expect(container.textContent).toContain("only one creature without banding");
    expect(attackButton(container).disabled).toBe(true);
    click(attackButton(container));
    expect(confirmed).toEqual([]);
  });
});
