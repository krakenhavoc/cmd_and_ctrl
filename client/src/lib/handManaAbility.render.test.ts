// @vitest-environment jsdom
//
// #1228: a card in hand can offer a MANA ability that functions there
// — "Exile this card from your hand: Add {R}" (the Spirit Guides,
// CR 113.6). It reaches the client on `zone_mana_abilities` rather
// than `mana_abilities`, for the reason cycling rides `zone_abilities`
// rather than `activated_abilities`: the server filters by the zone
// the card is in, and a hand is not public.
//
// What only a render test can see: that the mana popover opens on a
// HAND card at all (it was gated on `mana_abilities` alone, which the
// server now leaves empty there), that the row is labelled with the
// printed clause, and that clicking it calls back with the ability's
// own index — the index `activate_mana_ability` carries.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import type { CardView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(cleanup);

const spiritGuide = (): CardView =>
  ({
    instance_id: "c1",
    name: "Simian Spirit Guide",
    owner: "me",
    controller: "me",
    type_line: "Creature — Ape Spirit",
    zone_mana_abilities: [
      {
        index: 0,
        label: "Exile this card from your hand: Add {R}",
        produced: "{R}",
        exile_self: true,
      },
    ],
  }) as unknown as CardView;

function mountHandCard(card: CardView): {
  container: HTMLElement;
  root: HTMLElement;
  fired: number[];
} {
  const fired: number[] = [];
  const view = render(
    Card as never,
    {
      card,
      onActivateManaAbility: (index: number) => fired.push(index),
    } as never,
  );
  return {
    container: view.container,
    root: view.container.querySelector<HTMLElement>(".card")!,
    fired,
  };
}

function openMenu(root: HTMLElement): void {
  root.dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
  flushSync();
}

describe("hand mana abilities", () => {
  it("opens the mana popover on a hand card and fires the Spirit Guide row", () => {
    const tile = mountHandCard(spiritGuide());
    expect(tile.container.querySelector('[role="menu"]')).toBeNull();

    openMenu(tile.root);
    const menu = tile.container.querySelector<HTMLElement>('[role="menu"]');
    expect(menu, "the popover should open for a card with hand mana abilities").toBeTruthy();

    const rows = menu!.querySelectorAll<HTMLButtonElement>('[role="menuitem"]');
    expect(rows.length).toBe(1);
    expect(rows[0].textContent).toContain("Exile this card from your hand");

    click(rows[0]);
    flushSync();
    expect(tile.fired).toEqual([0]);
  });

  it("offers nothing when the card has no hand mana abilities", () => {
    const bare = { ...spiritGuide(), zone_mana_abilities: undefined } as CardView;
    const tile = mountHandCard(bare);
    openMenu(tile.root);
    expect(tile.container.querySelector('[role="menu"]')).toBeNull();
  });

  it("prefers the battlefield list when a card carries one", () => {
    // The two lists are disjoint by construction — the server filters
    // by zone (CR 113.6) — but if one ever arrived with both, the
    // permanent's menu is the one the board is rendering.
    const onBoard = {
      ...spiritGuide(),
      mana_abilities: [{ index: 1, label: "{T}: Add {G}", produced: "{G}", tap_cost: true }],
    } as unknown as CardView;
    const tile = mountHandCard(onBoard);
    openMenu(tile.root);
    const rows = tile.container.querySelectorAll<HTMLButtonElement>('[role="menuitem"]');
    expect(rows.length).toBe(1);
    expect(rows[0].textContent).toContain("{T}: Add {G}");
  });

  it("does not grey the row: a Spirit Guide has no tap cost to be blocked by", () => {
    const tapped = { ...spiritGuide(), tapped: true, summoning_sick: true } as CardView;
    const tile = mountHandCard(tapped);
    openMenu(tile.root);
    const row = tile.container.querySelector<HTMLButtonElement>('[role="menuitem"]');
    expect(row).toBeTruthy();
    expect(row!.disabled).toBe(false);
  });
});
