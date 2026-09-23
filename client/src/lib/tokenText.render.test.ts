// @vitest-environment jsdom
//
// ADR 0083, the client half: a token's printed ability text.
//
// A token has no Scryfall printing, so `scryfall_id` is empty, no art
// loads and the card renders through the `.name-fallback` branch —
// the name on a grey rectangle. That was enough while every token
// ability was a MANA or ACTIVATED one, because those reach the player
// as rows in the right-click menu. A token's own TRIGGER ("when this
// token dies, create a 2/2 red Dragon") has no control to hang text
// off, so without `token_text` the words are nowhere on the client at
// all and the trigger fires as a surprise.
//
// The server decides the wording; this file only pins that it is
// rendered, that it survives its printed line breaks, and that it is
// absent for everything that is not a token with text.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import type { CardView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

function mount(card: CardView) {
  return render(Card as never, { card } as Record<string, unknown>);
}

const token = (extra: Record<string, unknown>): CardView =>
  ({
    instance_id: "tok",
    name: "Dragon Egg",
    owner: "me",
    controller: "me",
    known_by_you: true,
    type_line: "Token Creature — Dragon",
    power: 0,
    toughness: 2,
    ...extra,
  }) as unknown as CardView;

function tokenText(container: HTMLElement): HTMLElement | null {
  return container.querySelector(".token-text");
}

describe("a token's printed text", () => {
  it("is rendered on a token that prints one", () => {
    const { container } = mount(
      token({ token_text: "When this token dies, create a 2/2 red Dragon creature token." }),
    );
    expect(tokenText(container)?.textContent).toBe(
      "When this token dies, create a 2/2 red Dragon creature token.",
    );
  });

  it("carries the full text in the title, so a clamped card is still readable", () => {
    const long =
      "When this token dies, create a 6/6 blue Whale creature token with " +
      '"When this token dies, create a 9/9 blue Kraken creature token."';
    const { container } = mount(token({ name: "Fish", token_text: long }));
    expect(tokenText(container)?.getAttribute("title")).toBe(long);
  });

  it("keeps the printed line breaks", () => {
    const { container } = mount(
      token({ token_text: "Defender\nWhen this token dies, create a 2/2 red Dragon." }),
    );
    // `white-space: pre-line` is what renders the break; the text
    // node has to still contain it for that to do anything.
    expect(tokenText(container)?.textContent).toContain("\n");
  });

  it("is absent on a vanilla token", () => {
    const { container } = mount(token({ name: "Soldier" }));
    expect(tokenText(container)).toBeNull();
  });

  it("is absent on a printed card", () => {
    const { container } = mount(
      token({ name: "Grizzly Bears", type_line: "Creature — Bear", scryfall_id: "abc" }),
    );
    expect(tokenText(container)).toBeNull();
  });
});
