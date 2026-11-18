// @vitest-environment jsdom
//
// zoneManaAbility.render.test.ts — #5896. A graveyard card can print a
// MANA ability that functions there: Jack-o'-Lantern's "{1}, Exile this
// card from your graveyard: Add one mana of any color." The server
// ships the row on `zone_mana_abilities` to the owner alone, and the
// zone browser, the one place a player looks at a graveyard, offers it
// in the card's menu.

import { describe, it, expect, afterEach } from "vitest";

import ZoneBrowserModal from "./components/board/ZoneBrowserModal.svelte";
import type { ActionType, CardView, GameView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(cleanup);

const ME = "me";

function lantern(extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: "jack",
    name: "Jack-o'-Lantern",
    owner: ME,
    controller: ME,
    type_line: "Artifact",
    zone_mana_abilities: [
      {
        index: 0,
        label: "{1}, Exile this card from your graveyard: Add one mana of any color.",
        mana_cost: "{1}",
        exile_self: true,
        produced: "{W|U|B|R|G}",
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
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 7, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
  } as unknown as GameView;
}

function mount(cards: CardView[], withHandler = true) {
  const fired: { card: CardView; index: number }[] = [];
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
        ? {
            onActivateManaAbility: (card: CardView, index: number) => fired.push({ card, index }),
          }
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

describe("zone-browser mana rows", () => {
  it("offers the graveyard mana row and hands the index up", () => {
    const m = mount([lantern()]);
    const menu = openMenuOn(m.container);
    expect(menu, "the popover should open for a card with a graveyard mana row").toBeTruthy();

    const rows = [...menu!.querySelectorAll<HTMLButtonElement>('[role="menuitem"]')];
    const row = rows.find((r) =>
      (r.textContent ?? "").includes("Exile this card from your graveyard"),
    );
    expect(row, "the mana row should be in the menu").toBeTruthy();

    click(row!);
    flushSync();
    expect(m.fired.map((f) => f.index)).toEqual([0]);
    expect(m.fired[0].card.instance_id).toBe("jack");
    // A mana ability opens no prompt of Board's, so the browser stays.
    expect(m.closed()).toBe(0);
  });

  it("offers nothing on a card the server stamped no rows for", () => {
    const m = mount([lantern({ zone_mana_abilities: undefined })]);
    expect(openMenuOn(m.container)).toBeNull();
  });

  it("offers nothing when no parent wired the callback", () => {
    const m = mount([lantern()], false);
    expect(openMenuOn(m.container)).toBeNull();
  });
});
