// @vitest-environment jsdom
//
// ADR 0104 (owner decision 6): a permanent another player controls
// wears a TAKEN FROM badge naming its owner — a stolen creature, or the
// permanent a stolen spell became.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import type { CardView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const permanent = {
  instance_id: "perm",
  name: "Grizzly Bears",
  owner: "alice",
  controller: "bob",
  known_by_you: true,
  type_line: "Creature — Bear",
} as unknown as CardView;

describe("the TAKEN FROM badge", () => {
  it("names the owner when the prop is set", () => {
    const { container } = render(
      Card as never,
      {
        card: permanent,
        takenFrom: "Alice",
      } as Record<string, unknown>,
    );
    const badge = container.querySelector<HTMLElement>(".badge.taken");
    expect(badge?.textContent?.trim()).toBe("TAKEN FROM Alice");
    expect(badge?.getAttribute("aria-label")).toBe("taken from Alice");
  });

  it("is absent for a permanent its owner controls", () => {
    const { container } = render(Card as never, { card: permanent } as Record<string, unknown>);
    expect(container.querySelector(".badge.taken")).toBeNull();
  });
});
