// @vitest-environment jsdom
//
// #781, the client half. A colour or creature type chosen as a
// permanent entered is announced at the table (CR 105.4) and is what
// makes the rest of the card readable: CR 607.2d links "choose a
// color" to "creatures you control of the chosen color", and that set
// cannot be computed without the answer. Before this the engine stored
// the choice and told nobody — not even the player who made it.
//
// The chip is rendered from `chosenValueChips`, the one module that
// turns "G" into "Green", so the card tile, the zone browser (which
// renders the same tile) and the hover panel cannot disagree.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import { chosenValueChips } from "./chosenValues";
import type { CardView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const permanent = (extra: Record<string, unknown>): CardView =>
  ({
    instance_id: "perm",
    name: "Coldsteel Heart",
    owner: "me",
    controller: "me",
    known_by_you: true,
    type_line: "Artifact",
    ...extra,
  }) as unknown as CardView;

function chips(container: HTMLElement): { text: string; title: string }[] {
  return [...container.querySelectorAll<HTMLElement>(".kw-chosen")].map((el) => ({
    text: el.textContent?.trim() ?? "",
    title: el.getAttribute("title") ?? "",
  }));
}

describe("chosenValueChips", () => {
  it("is empty for the card that chose nothing", () => {
    expect(chosenValueChips({})).toEqual([]);
    expect(chosenValueChips({ chosen_color: "", named_tribe: "" })).toEqual([]);
  });

  it("names the colour rather than shipping the letter", () => {
    expect(chosenValueChips({ chosen_color: "G" })[0].label).toBe("Green");
    expect(chosenValueChips({ chosen_color: "U" })[0].label).toBe("Blue");
    expect(chosenValueChips({ chosen_color: "B" })[0].label).toBe("Black");
  });

  it("quotes the linked clause in the tooltip (CR 607.2d)", () => {
    expect(chosenValueChips({ chosen_color: "R" })[0].title).toContain('"the chosen color"');
    expect(chosenValueChips({ named_tribe: "Elf" })[0].title).toContain('"the chosen type"');
  });

  it("shows a letter it does not recognise rather than dropping the answer", () => {
    // A future server that learns a sixth colour must not make the
    // client silently forget that a choice was made.
    expect(chosenValueChips({ chosen_color: "Z" })[0].label).toBe("Z");
  });

  it("puts the colour first when a permanent chose both", () => {
    const both = chosenValueChips({ chosen_color: "W", named_tribe: "Elf" });
    expect(both.map((c) => c.kind)).toEqual(["color", "tribe"]);
  });
});

describe("the chosen-value chip on a card", () => {
  it("is absent on a permanent that chose nothing", () => {
    const { container } = mountCard(permanent({ name: "Sol Ring" }));
    expect(chips(container)).toEqual([]);
  });

  it("shows the chosen colour by name", () => {
    const { container } = mountCard(permanent({ chosen_color: "G" }));
    expect(chips(container).map((c) => c.text)).toEqual(["Green"]);
    expect(chips(container)[0].title).toMatch(/Green/);
  });

  it("shows the named creature type", () => {
    const { container } = mountCard(permanent({ name: "Adaptive Automaton", named_tribe: "Elf" }));
    expect(chips(container).map((c) => c.text)).toEqual(["Elf"]);
  });

  it("shows both, colour first, on a permanent that chose both", () => {
    const { container } = mountCard(permanent({ chosen_color: "R", named_tribe: "Goblin" }));
    expect(chips(container).map((c) => c.text)).toEqual(["Red", "Goblin"]);
  });

  it("renders alongside the keyword badges rather than replacing them", () => {
    const { container } = mountCard(
      permanent({
        name: "Adaptive Automaton",
        type_line: "Artifact Creature — Construct",
        named_tribe: "Elf",
        abilities: ["flying"],
      }),
    );
    const row = container.querySelector(".keyword-row");
    expect(row, "the badge row should be rendered").toBeTruthy();
    expect(row!.querySelectorAll(".kw-badge").length).toBe(2);
  });

  // The server clears both fields for a viewer who is not a knower of
  // the card — "Elf" names Cavern of Souls as loudly as loyalty says
  // "planeswalker". This pins the client half: it renders what it is
  // given and invents nothing for a card it cannot read.
  it("renders nothing for a card the viewer does not know", () => {
    const { container } = mountCard(
      permanent({ known_by_you: false, name: "", type_line: "", chosen_color: undefined }),
    );
    expect(chips(container)).toEqual([]);
  });
});

function mountCard(card: CardView) {
  return render(Card as never, { card } as Record<string, unknown>);
}
