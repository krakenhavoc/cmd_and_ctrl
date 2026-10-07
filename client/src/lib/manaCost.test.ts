// @vitest-environment jsdom
//
// manaCost.test.ts — #2231. A cost is drawn as mana pips: the mapping
// from brace notation to pips (pure), what is said aloud, and the hand
// card's price badge as markup.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import { costPips, costWords, pipRun } from "./manaSymbol";
import type { CardView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const run = (cost: string): string[] => costPips(cost).map((p) => p.symbol);

describe("costPips", () => {
  it("draws coloured symbols as one pip each", () => {
    expect(run("{U}{U}{R}")).toEqual(["U", "U", "R"]); // Counterflux
  });

  it("sums generic mana into one circle", () => {
    expect(run("{2}")).toEqual(["2"]); // Arcane Signet
    expect(run("{1}{1}{U}")).toEqual(["2", "U"]);
    expect(run("{10}")).toEqual(["10"]);
  });

  it("keeps a generic circle and a coloured pip apart", () => {
    expect(costPips("{2}{U}")).toEqual([
      { symbol: "2", kind: "generic" },
      { symbol: "U", kind: "colour" },
    ]); // Rhystic Study
  });

  it("keeps {C} as its own colourless pip, never folded into generic", () => {
    expect(costPips("{5}{C}")).toEqual([
      { symbol: "5", kind: "generic" },
      { symbol: "C", kind: "colourless" },
    ]);
    expect(costPips("{C}{C}").map((p) => p.kind)).toEqual(["colourless", "colourless"]);
  });

  it("handles X, hybrid, Phyrexian, snow and zero", () => {
    expect(costPips("{X}{X}{R}").map((p) => p.kind)).toEqual(["x", "x", "colour"]);
    expect(costPips("{W/U}{2/W}").map((p) => [p.symbol, p.kind])).toEqual([
      ["W/U", "hybrid"],
      ["2/W", "hybrid"],
    ]);
    expect(costPips("{1}{U/P}")).toEqual([
      { symbol: "1", kind: "generic" },
      { symbol: "U/P", kind: "phyrexian" },
    ]);
    expect(costPips("{S}{S}").map((p) => p.kind)).toEqual(["snow", "snow"]);
    expect(run("{0}")).toEqual(["0"]);
    expect(run("{0}{R}")).toEqual(["R"]);
  });

  it("is empty for no cost", () => {
    expect(costPips("")).toEqual([]);
    expect(pipRun([])).toEqual([]);
  });
});

describe("costWords", () => {
  it("says a cost the way a player would", () => {
    expect(costWords("{2}{U}")).toBe("2 generic and 1 blue");
    expect(costWords("{U}{U}{R}")).toBe("2 blue and 1 red");
    expect(costWords("{2}")).toBe("2 generic");
    expect(costWords("{C}")).toBe("1 colourless");
    expect(costWords("{W/U}")).toBe("1 white or blue");
    expect(costWords("{G/P}")).toBe("1 green or 2 life");
  });

  // ADR 0129 §8: {E} is an energy pip, said as energy.
  it("says energy", () => {
    expect(costWords("{E}{E}")).toBe("2 energy");
    expect(costPips("{E}{E}")).toEqual([
      { symbol: "E", kind: "energy" },
      { symbol: "E", kind: "energy" },
    ]);
  });
});

const card = (name: string, mana_cost: string): CardView =>
  ({
    instance_id: name,
    scryfall_id: `sf-${name}`,
    name,
    owner: "me",
    controller: "me",
    known_by_you: true,
    type_line: "Instant",
    mana_cost,
  }) as unknown as CardView;

function pipsOf(c: CardView): { badge: HTMLElement | null; pips: (string | null)[] } {
  const { container } = render(Card as never, { card: c, showManaCost: true } as never);
  const badge = container.querySelector<HTMLElement>(".badge.cost");
  const pips = Array.from(badge?.querySelectorAll(".mana-symbol") ?? [], (e) =>
    e.getAttribute("data-symbol"),
  );
  return { badge, pips };
}

describe("the hand card's price badge", () => {
  it("draws Counterflux as blue, blue, red pips with a spoken name", () => {
    const { badge, pips } = pipsOf(card("Counterflux", "{U}{U}{R}"));
    expect(pips).toEqual(["U", "U", "R"]);
    expect(badge?.textContent).not.toContain("{");
    expect(badge?.querySelector("[role=img]")?.getAttribute("aria-label")).toBe(
      "mana cost 2 blue and 1 red",
    );
  });

  it("draws Rhystic Study as a 2 circle and a blue pip", () => {
    expect(pipsOf(card("Rhystic Study", "{2}{U}")).pips).toEqual(["2", "U"]);
  });

  it("draws Arcane Signet as a single 2", () => {
    expect(pipsOf(card("Arcane Signet", "{2}")).pips).toEqual(["2"]);
  });

  it("gives {C} its own pip", () => {
    expect(pipsOf(card("Thought-Knot Seer", "{3}{C}")).pips).toEqual(["3", "C"]);
  });

  it("draws no badge for a card without a cost", () => {
    expect(pipsOf(card("Plains", "")).badge).toBeNull();
  });
});
