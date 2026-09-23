// @vitest-environment jsdom
//
// #1227: ninjutsu (CR 702.49) is a HAND ability whose cost returns an
// unblocked attacker you control — so its row is unpayable for nearly
// the whole game and payable only in the declare-blockers window with
// an unblocked attacker on the board.
//
// Nothing on the wire is new: the row arrives as `zone_abilities`
// (#1221) carrying `return_label` / `return_options` (#1213), and the
// announce chain that collects `return_ids` never looks at the card's
// zone. What WAS missing is the greying: the right-click menu had the
// shortfall check inline and this popover — the one a hand card and
// the zone browser open — had none, so a ninjutsu row with no unblocked
// attacker was clickable and came back refused.
//
// One predicate now answers for both (returnShortfall), and these are
// the two ends of it.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import { returnShortfall } from "./contextMenu.logic";
import type { CardView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(cleanup);

const ninja = (returnCards: string[]): CardView =>
  ({
    instance_id: "ninja",
    name: "Ninja of the Deep Hours",
    owner: "me",
    controller: "me",
    type_line: "Creature — Human Ninja",
    zone_abilities: [
      {
        index: 0,
        label:
          "Ninjutsu {1}{U} ({1}{U}, Return an unblocked attacker you control to hand: " +
          "Put this card onto the battlefield from your hand tapped and attacking.)",
        mana_cost: "{1}{U}",
        return_label: "an unblocked attacker you control",
        return_options: { cards: returnCards, min: 1, max: 1 },
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

describe("returnShortfall", () => {
  it("is silent when the clause has a payment", () => {
    expect(returnShortfall({ cards: ["a"], min: 1, max: 1 }, "an unblocked attacker")).toBe("");
  });

  it("names the clause when nothing can pay", () => {
    expect(returnShortfall({ cards: [], min: 1, max: 1 }, "an unblocked attacker")).toBe(
      "nothing to return (an unblocked attacker)",
    );
  });

  it("falls back to a generic clause and to a count of one", () => {
    expect(returnShortfall({ cards: [] })).toBe("nothing to return (a permanent you control)");
  });

  it("says nothing about an ability with no return cost at all", () => {
    expect(returnShortfall(undefined, "a Forest you control")).toBe("");
  });
});

describe("ninjutsu row", () => {
  it("greys the row when no unblocked attacker can pay, and does not fire", () => {
    const tile = mountHandCard(ninja([]));
    openMenu(tile.root);
    const rows = tile.container.querySelectorAll<HTMLButtonElement>('[role="menuitem"]');
    expect(rows.length).toBe(1);
    expect(rows[0].disabled).toBe(true);
    expect(rows[0].title).toBe("nothing to return (an unblocked attacker you control)");

    click(rows[0]);
    flushSync();
    expect(tile.fired).toEqual([]);
  });

  it("offers the row once an unblocked attacker is on the board", () => {
    const tile = mountHandCard(ninja(["rat"]));
    openMenu(tile.root);
    const rows = tile.container.querySelectorAll<HTMLButtonElement>('[role="menuitem"]');
    expect(rows.length).toBe(1);
    expect(rows[0].disabled).toBe(false);
    expect(rows[0].textContent).toContain("Ninjutsu {1}{U}");

    click(rows[0]);
    flushSync();
    expect(tile.fired).toEqual([0]);
  });
});
