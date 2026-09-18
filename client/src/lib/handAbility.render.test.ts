// @vitest-environment jsdom
//
// #660: a card in hand offers the abilities that function THERE
// (cycling, CR 702.29a). They reach the client on `hand_abilities`
// rather than `activated_abilities` — a hand is not public, so the
// server strips the list for every seat but the owner — and Card
// opens the same popover for either list.
//
// What only a render test can see: that the popover appears at all on
// a hand card (it is gated on the card having abilities AND the parent
// wiring the callback), that the row is labelled with the cycling
// text, and that clicking it calls back with the ability's own index —
// the index the engine validates against.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import type { CardView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(cleanup);

const triome = (): CardView =>
  ({
    instance_id: "c1",
    name: "Ketria Triome",
    owner: "me",
    controller: "me",
    type_line: "Land — Forest Island Mountain",
    hand_abilities: [
      {
        index: 0,
        label: "Cycling {3} ({3}, Discard this card: Draw a card.)",
        mana_cost: "{3}",
        discard_self: true,
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
      onActivateAbility: (index: number) => fired.push(index),
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

describe("hand abilities", () => {
  it("opens the ability popover on a hand card and fires the cycling row", () => {
    const tile = mountHandCard(triome());
    expect(tile.container.querySelector('[role="menu"]')).toBeNull();

    openMenu(tile.root);
    const menu = tile.container.querySelector<HTMLElement>('[role="menu"]');
    expect(menu, "the popover should open for a card with hand abilities").toBeTruthy();

    const rows = menu!.querySelectorAll<HTMLButtonElement>('[role="menuitem"]');
    expect(rows.length).toBe(1);
    expect(rows[0].textContent).toContain("Cycling {3}");

    click(rows[0]);
    flushSync();
    expect(tile.fired).toEqual([0]);
  });

  it("offers nothing when the card has no hand abilities", () => {
    const bare = { ...triome(), hand_abilities: undefined } as CardView;
    const tile = mountHandCard(bare);
    openMenu(tile.root);
    expect(tile.container.querySelector('[role="menu"]')).toBeNull();
  });

  it("prefers the battlefield list when a card carries one", () => {
    // The two lists are disjoint by construction — the server filters
    // by zone — but if one ever arrived with both, the permanent's
    // menu is the one the board is rendering.
    const onBoard = {
      ...triome(),
      activated_abilities: [{ index: 1, label: "{T}: Add {G}" }],
    } as unknown as CardView;
    const tile = mountHandCard(onBoard);
    openMenu(tile.root);
    const rows = tile.container.querySelectorAll<HTMLButtonElement>('[role="menuitem"]');
    expect(rows.length).toBe(1);
    expect(rows[0].textContent).toContain("{T}: Add {G}");
  });
});
