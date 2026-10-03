// @vitest-environment jsdom
//
// ADR 0109 §10 decision 7, the client half: a permanent that took riot's
// haste shows the haste chip labelled "Riot", and the riot and unleash
// keywords read by name.

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
        name: "Zhur-Taa Goblin",
        owner: "me",
        controller: "me",
        type_line: "Creature — Goblin Berserker",
        power: 2,
        toughness: 2,
        ...card,
      } as unknown as CardView,
    } as Record<string, unknown>,
  );
}

function titles(container: HTMLElement): string[] {
  return [...container.querySelectorAll(".keyword-row .kw-badge")].map(
    (el) => el.getAttribute("title") ?? "",
  );
}

describe("the riot haste chip", () => {
  it("labels haste that came from riot", () => {
    const { container } = mount({ abilities: ["riot", "haste"], riot_haste: true });
    expect(titles(container)).toEqual(["Riot", "Riot: haste"]);
  });

  it("leaves any other haste as Haste", () => {
    const { container } = mount({ abilities: ["riot", "haste"] });
    expect(titles(container)).toEqual(["Riot", "Haste"]);
  });

  it("names unleash", () => {
    const { container } = mount({ abilities: ["unleash"] });
    expect(titles(container)).toEqual(["Unleash"]);
  });
});
