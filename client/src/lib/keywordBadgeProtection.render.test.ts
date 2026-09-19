// @vitest-environment jsdom
//
// #662, the client half: a protection badge names the QUALITY.
//
// The keyword row used to render every ability token the same way —
// three letters off the front of the string and the raw token as the
// tooltip — which for protection meant "PRO" over "protection from
// red". Protection is the one keyword whose ability carries a
// parameter, and the parameter is the whole information: "protection
// from Demons" and "protection from red" are different cards.
//
// The parse comes from the server on `card.protection`; the client
// owns no protection grammar (server/internal/game/protection.go is
// the only one). These tests pin what the row does with it.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import type { CardView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

function mount(card: Partial<CardView>) {
  return render(
    Card as never,
    {
      card: {
        instance_id: "c1",
        name: "Baneslayer Angel",
        owner: "me",
        controller: "me",
        type_line: "Creature — Angel",
        power: 5,
        toughness: 5,
        ...card,
      } as unknown as CardView,
    } as Record<string, unknown>,
  );
}

function badges(container: HTMLElement): { text: string; title: string }[] {
  return [...container.querySelectorAll(".keyword-row .kw-badge")].map((el) => ({
    text: (el.textContent ?? "").trim(),
    title: el.getAttribute("title") ?? "",
  }));
}

describe("the protection badge", () => {
  it("names the quality in the tooltip instead of showing the raw token", () => {
    const { container } = mount({
      abilities: ["flying", "protection from red"],
      protection: [{ printed: "red", kind: "color", value: "R" }],
    });
    const got = badges(container);
    expect(got.map((b) => b.title)).toContain("Protection from red");
    expect(got.map((b) => b.title)).not.toContain("protection from red");
  });

  it("abbreviates the quality, not the word protection", () => {
    const { container } = mount({
      abilities: ["protection from Demons", "protection from Dragons"],
      protection: [
        { printed: "Demons", kind: "subtype", value: "Demon" },
        { printed: "Dragons", kind: "subtype", value: "Dragon" },
      ],
    });
    const got = badges(container);
    expect(got.map((b) => b.text)).toEqual(expect.arrayContaining(["DEM", "DRA"]));
    expect(got.map((b) => b.title)).toEqual(
      expect.arrayContaining(["Protection from Demons", "Protection from Dragons"]),
    );
  });

  it("renders protection from everything as ALL", () => {
    const { container } = mount({
      abilities: ["protection from everything"],
      protection: [{ printed: "everything", kind: "everything" }],
    });
    expect(badges(container)).toEqual([{ text: "ALL", title: "Protection from everything" }]);
  });

  it("does not badge the raw token twice", () => {
    const { container } = mount({
      abilities: ["protection from red"],
      protection: [{ printed: "red", kind: "color", value: "R" }],
    });
    expect(badges(container)).toHaveLength(1);
  });

  // CR 702.16m: two Swords of Fire and Ice on one creature grant the
  // same protection twice. It is one quality for every rules check,
  // and two identical `{#each}` keys would throw in Svelte 5.
  it("dedupes two instances of the same protection", () => {
    const { container } = mount({
      abilities: ["protection from red", "protection from red"],
      protection: [
        { printed: "red", kind: "color", value: "R" },
        { printed: "red", kind: "color", value: "R" },
      ],
    });
    expect(badges(container)).toHaveLength(1);
  });

  it("leaves the ordinary keyword badges alone", () => {
    const { container } = mount({ abilities: ["flying", "lifelink"] });
    expect(badges(container)).toHaveLength(2);
  });
});
