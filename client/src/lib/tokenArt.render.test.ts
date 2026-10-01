// @vitest-environment jsdom
//
// ADR 0078 (#1115), the client half: once the server stamps a
// resolved Scryfall TOKEN PRINTING onto `scryfall_id`, the existing
// image path — `cardImageURL`, `Card.svelte`'s `imgSrc` branch, the
// service worker's cache — needs no change at all to render it. This
// file is that claim, pinned: a token with a resolved `scryfall_id`
// draws an `<img>`, exactly like a printed card, and a token that
// resolved to nothing keeps today's `.name-fallback` text.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import type { CardView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

function mount(card: CardView) {
  return render(Card as never, { card } as Record<string, unknown>);
}

const treasureToken = (extra: Record<string, unknown> = {}): CardView =>
  ({
    instance_id: "tok-1",
    name: "Treasure",
    owner: "me",
    controller: "me",
    known_by_you: true,
    type_line: "Token Artifact — Treasure",
    is_token: true,
    ...extra,
  }) as unknown as CardView;

describe("a token with a resolved Scryfall printing", () => {
  it("renders its art like any other card", () => {
    const { container } = mount(
      treasureToken({ scryfall_id: "11111111-1111-1111-1111-111111111111" }),
    );
    const img = container.querySelector("img:not(.back-img)") as HTMLImageElement | null;
    expect(img).not.toBeNull();
    expect(img?.src).toContain("/cards/11111111-1111-1111-1111-111111111111/image");
    expect(container.querySelector(".name-fallback")).toBeNull();
  });

  it("falls back to the name when no printing resolved", () => {
    const { container } = mount(treasureToken());
    expect(container.querySelector("img:not(.back-img)")).toBeNull();
    expect(container.querySelector(".name-fallback")?.textContent).toBe("Treasure");
  });

  it("a face-down token still shows a card back, even with art resolved", () => {
    const { container } = mount(
      treasureToken({
        scryfall_id: "11111111-1111-1111-1111-111111111111",
        face_down: true,
        face_visible: false,
        known_by_you: false,
        name: "",
        type_line: "",
      }),
    );
    expect(container.querySelector("img.back-img")).not.toBeNull();
    expect(container.querySelector("img:not(.back-img)")).toBeNull();
  });
});
