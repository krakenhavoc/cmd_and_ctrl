// #5704: the colour edge of a draft card.

import { describe, expect, it } from "vitest";

import { cardEdge, edgeBackground, edgeColours } from "./draftEdge";

describe("the colours of an edge", () => {
  it("is the card's colour for a mono-coloured card", () => {
    expect(edgeColours({ colors: ["R"] })).toEqual(["R"]);
  });

  it("is the colours in WUBRG order, whatever order they came in", () => {
    expect(edgeColours({ colors: ["G", "W", "B"] })).toEqual(["W", "B", "G"]);
    expect(edgeColours({ colors: ["R", "U"] })).toEqual(["U", "R"]);
  });

  it("is empty for a colourless card", () => {
    expect(edgeColours({ colors: [] })).toEqual([]);
    expect(edgeColours({})).toEqual([]);
  });

  it("uses the colour identity, which is how a land has a colour", () => {
    expect(edgeColours({ colors: [], color_identity: ["G"] })).toEqual(["G"]);
    // An identity wider than the card's colours wins: an ability can add one.
    expect(edgeColours({ colors: ["W"], color_identity: ["W", "U"] })).toEqual(["W", "U"]);
  });

  it("falls back to the colours when the wire sent no identity", () => {
    expect(edgeColours({ colors: ["B"], color_identity: [] })).toEqual(["B"]);
    expect(edgeColours({ colors: ["B"], color_identity: undefined })).toEqual(["B"]);
  });

  it("drops a letter that is not a colour and counts one twice only once", () => {
    expect(edgeColours({ colors: ["C", "U", "U", "x"] })).toEqual(["U"]);
  });
});

describe("the background of an edge", () => {
  it("is one flat identity token for one colour", () => {
    expect(edgeBackground(["U"])).toBe("linear-gradient(var(--identity-u), var(--identity-u))");
  });

  it("is the neutral token for no colour", () => {
    expect(edgeBackground([])).toBe("linear-gradient(var(--identity-c), var(--identity-c))");
  });

  it("splits two colours in halves on a diagonal", () => {
    expect(edgeBackground(["W", "B"])).toBe(
      "linear-gradient(135deg, var(--identity-w) 0% 50%, var(--identity-b) 50% 100%)",
    );
  });

  it("splits three or five colours in equal bands, each colour once, in order", () => {
    const three = edgeBackground(["U", "R", "G"]);
    expect(three).toBe(
      "linear-gradient(135deg, var(--identity-u) 0% 33.33%, var(--identity-r) 33.33% 66.67%, var(--identity-g) 66.67% 100%)",
    );
    const five = edgeBackground(["W", "U", "B", "R", "G"]);
    expect(five.match(/--identity-/g)).toHaveLength(5);
    expect(five).toContain("var(--identity-w) 0% 20%");
    expect(five).toContain("var(--identity-g) 80% 100%");
  });

  it("uses only the colour-identity tokens, never a literal colour", () => {
    for (const c of [[], ["W"], ["W", "U"], ["W", "U", "B", "R", "G"]] as const) {
      expect(edgeBackground(c).replace(/var\(--identity-[wubrgc]\)/g, "")).not.toMatch(/#|rgb/);
    }
  });
});

describe("the edge of a card", () => {
  it("names the colours in a key, empty for colourless", () => {
    expect(cardEdge({ colors: ["U", "R"] }).key).toBe("UR");
    expect(cardEdge({ colors: [] }).key).toBe("");
    expect(cardEdge({ colors: [], color_identity: ["G"] }).key).toBe("G");
  });

  it("gives a Forest the green edge", () => {
    expect(cardEdge({ colors: [], color_identity: ["G"] }).background).toContain("--identity-g");
  });
});
