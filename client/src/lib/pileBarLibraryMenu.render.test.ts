// @vitest-environment jsdom
//
// pileBarLibraryMenu.render.test.ts — #2962. One click on your own
// library used to draw a card, and a misclick cost you one. The click
// now opens a menu; only a row sends anything.

import { describe, it, expect, afterEach } from "vitest";

import PileBar from "./components/board/PileBar.svelte";
import type { PlayerView, ZoneView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const zone = (kind: string, count = 0): ZoneView => ({ kind, count, cards: [] });

function mount(isSelf = true) {
  const sent: string[] = [];
  let draws = 0;
  const view = render(
    PileBar as never,
    {
      seat: {
        id: "me",
        name: "Me",
        seat: 0,
        life: 40,
        library: zone("library", 60),
        hand: zone("hand"),
        graveyard: zone("graveyard"),
        command: zone("command"),
        commander_damage: {},
        life_history: [],
      } as unknown as PlayerView,
      exile: zone("exile"),
      isSelf,
      sendAction: (t: string) => sent.push(t),
      onDrawCard: () => draws++,
    } as never,
  );
  const pile = view.container.querySelector<HTMLElement>('[data-pile="library"]')!;
  return { c: view.container, pile, sent, draws: () => draws };
}

const items = (c: HTMLElement) => [...c.querySelectorAll<HTMLElement>("[role=menuitem]")];

describe("the library pile menu (#2962)", () => {
  it("does not draw on a single click; it opens a menu", () => {
    const m = mount();
    click(m.pile);
    expect(m.draws()).toBe(0);
    expect(m.c.querySelector("[role=menu]")).not.toBeNull();
    expect(items(m.c).map((i) => i.textContent?.trim())).toEqual([
      "Draw a card",
      "Shuffle library",
    ]);
  });

  it("draws once from the Draw a card row and closes the menu", () => {
    const m = mount();
    click(m.pile);
    click(items(m.c)[0]);
    expect(m.draws()).toBe(1);
    expect(m.c.querySelector("[role=menu]")).toBeNull();
  });

  it("shuffles from the Shuffle row", () => {
    const m = mount();
    click(m.pile);
    click(items(m.c)[1]);
    expect(m.sent).toEqual(["shuffle_library"]);
    expect(m.draws()).toBe(0);
  });

  it("a second click on the pile closes the menu without drawing", () => {
    const m = mount();
    click(m.pile);
    click(m.pile);
    expect(m.c.querySelector("[role=menu]")).toBeNull();
    expect(m.draws()).toBe(0);
  });

  it("is inert on another seat's library", () => {
    const m = mount(false);
    click(m.pile);
    expect(m.c.querySelector("[role=menu]")).toBeNull();
  });
});
