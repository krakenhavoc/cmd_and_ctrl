// @vitest-environment jsdom
//
// zoneAbility.render.test.ts — #1221. A card in a BROWSED pile can
// print a CR 602 ability that functions from that pile: unearth and
// scavenge out of a graveyard, an activation out of exile. The rows
// reach the client on `zone_abilities` — the same field a hand card's
// cycling rides since #660 — and ZoneBrowserModal has to offer them,
// because the browser is the only place a player ever looks at a
// graveyard.
//
// A render test rather than a pure one for impulseCast.render.test's
// reason: what was missing is a PROP, and there is no pure function
// between the right-click and the callback to examine.

import { describe, it, expect, afterEach } from "vitest";

import ZoneBrowserModal from "./components/board/ZoneBrowserModal.svelte";
import type { ActionType, CardView, GameView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(cleanup);

const ME = "me";
const THEM = "them";

function zombie(extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: "dregscape",
    name: "Dregscape Zombie",
    owner: ME,
    controller: ME,
    type_line: "Creature — Zombie",
    zone_abilities: [
      {
        index: 0,
        label: "Unearth {B} ({B}: Return this card from your graveyard to the battlefield.)",
        mana_cost: "{B}",
        sorcery_speed: true,
      },
    ],
    ...extras,
  } as unknown as CardView;
}

function snapWith(cards: CardView[]): GameView {
  return {
    id: "g",
    state: "active",
    seats: [
      {
        id: ME,
        name: "Me",
        hand: { kind: "hand", count: 0, cards: [] },
        graveyard: { kind: "graveyard", count: cards.length, cards },
      },
      {
        id: THEM,
        name: "Them",
        hand: { kind: "hand", count: 0, cards: [] },
        graveyard: { kind: "graveyard", count: 0, cards: [] },
      },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 7, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
  } as unknown as GameView;
}

interface Fired {
  card: CardView;
  index: number;
}

function mount(cards: CardView[], withHandler = true) {
  const fired: Fired[] = [];
  let closed = 0;
  const view = render(
    ZoneBrowserModal as never,
    {
      view: snapWith(cards),
      viewerID: ME,
      zoneKind: "graveyard",
      ownerSeat: { id: ME, name: "Me" },
      sendAction: (_type: ActionType) => {},
      onClose: () => {
        closed += 1;
      },
      ...(withHandler
        ? { onActivateAbility: (card: CardView, index: number) => fired.push({ card, index }) }
        : {}),
    } as never,
  );
  return { container: view.container, fired, closed: () => closed };
}

function openMenuOn(container: HTMLElement): HTMLElement | null {
  const tile = container.querySelector<HTMLElement>(".card")!;
  tile.dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
  flushSync();
  return container.querySelector<HTMLElement>('[role="menu"]');
}

describe("zone-browser ability rows", () => {
  it("opens the popover on a browsed graveyard card and hands the index up", () => {
    const m = mount([zombie()]);
    const menu = openMenuOn(m.container);
    expect(menu, "the popover should open for a card with zone abilities").toBeTruthy();

    const rows = menu!.querySelectorAll<HTMLButtonElement>('[role="menuitem"]');
    expect(rows.length).toBe(1);
    expect(rows[0].textContent).toContain("Unearth {B}");

    click(rows[0]);
    flushSync();
    expect(m.fired.map((f) => f.index)).toEqual([0]);
    expect(m.fired[0].card.instance_id).toBe("dregscape");
    // The browser gets out of the way first: an activation can open a
    // target picker or an X prompt, and those are Board's chain.
    expect(m.closed()).toBe(1);
  });

  it("offers nothing on a card the server stamped no rows for", () => {
    // What a bystander's frame looks like: the card is readable, the
    // answer is not (#1055).
    const m = mount([zombie({ zone_abilities: undefined })]);
    expect(openMenuOn(m.container)).toBeNull();
  });

  it("offers nothing when no parent wired the callback", () => {
    const m = mount([zombie()], false);
    expect(openMenuOn(m.container)).toBeNull();
  });
});
