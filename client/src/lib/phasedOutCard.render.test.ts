// @vitest-environment jsdom
//
// #1199 / ADR 0084, the client half: what a phased-out permanent
// looks like on the board.
//
// Board.svelte folds GameView.phased_out onto its controller's row,
// so the player can see WHERE their four permanents were. What only
// exists once the markup is rendered — and so is what this file is
// for — is the BADGE and the withheld click: CR 702.26b says a
// phased-out permanent "can't affect or be affected by anything else
// in the game", and a card that still answers a click is a card the
// server is about to refuse.

import { describe, it, expect, afterEach, vi } from "vitest";

import Card from "./components/board/Card.svelte";
import type { CardView } from "./protocol";
import { render, cleanup, click } from "./test/render.svelte";

afterEach(cleanup);

function mount(card: CardView, onClick?: () => void) {
  return render(Card as never, { card, onClick } as Record<string, unknown>);
}

const phasedBear = (): CardView =>
  ({
    instance_id: "phased",
    name: "Grizzly Bears",
    owner: "me",
    controller: "me",
    scryfall_id: "aaaa-bbbb",
    type_line: "Creature — Bear",
    power: 2,
    toughness: 2,
    known_by_you: true,
    phased_out: true,
  }) as unknown as CardView;

const presentBear = (): CardView =>
  ({
    instance_id: "present",
    name: "Grizzly Bears",
    owner: "me",
    controller: "me",
    scryfall_id: "aaaa-bbbb",
    type_line: "Creature — Bear",
    power: 2,
    toughness: 2,
    known_by_you: true,
  }) as unknown as CardView;

describe("a phased-out permanent on the board", () => {
  it("wears a PHASED badge, because the board otherwise reads as a death", () => {
    const { container } = mount(phasedBear());
    const badge = container.querySelector(".badge.phased");
    expect(badge).not.toBeNull();
    expect(badge?.textContent?.trim()).toBe("PHASED");
  });

  it("carries the dimming class", () => {
    const { container } = mount(phasedBear());
    expect(container.querySelector(".card")?.classList.contains("phased-out")).toBe(true);
  });

  it("is not clickable — CR 702.26b, it can't be affected by anything", () => {
    const onClick = vi.fn();
    const { container } = mount(phasedBear(), onClick);
    const el = container.querySelector(".card") as HTMLElement;
    expect(el.classList.contains("clickable")).toBe(false);
    expect(el.getAttribute("role")).toBe("img");
    expect(el.getAttribute("tabindex")).toBeNull();
    click(el);
    expect(onClick).not.toHaveBeenCalled();
  });

  it("leaves an ordinary permanent alone", () => {
    const onClick = vi.fn();
    const { container } = mount(presentBear(), onClick);
    const el = container.querySelector(".card") as HTMLElement;
    expect(container.querySelector(".badge.phased")).toBeNull();
    expect(el.classList.contains("phased-out")).toBe(false);
    expect(el.classList.contains("clickable")).toBe(true);
    click(el);
    expect(onClick).toHaveBeenCalled();
  });
});
