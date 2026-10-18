// @vitest-environment jsdom
//
// #4106 — reading the other side of a two-faced card in the hover zoom.
// The info panel always lists the face that is NOT up under a "Back
// face" (or "Front face") heading, and holding Shift swaps the big
// image with the inset. Split cards stay one side (#3518) and a
// face-down card shows nothing more.

import { describe, it, expect, beforeEach, afterEach } from "vitest";

import HoverZoomOverlay from "./components/board/HoverZoomOverlay.svelte";
import { hoveredCard } from "./cardTypes";
import { L } from "./labels";
import type { CardView, GameView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

const view = {
  id: "g",
  state: "active",
  seats: [],
  battlefield: { kind: "battlefield", count: 0, cards: [] },
  stack: { kind: "stack", count: 0, cards: [] },
  exile: { kind: "exile", count: 0, cards: [] },
  turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
} as unknown as GameView;

const TOVOLAR_ID = "925dee92-9c8d-4eef-a326-2fdae72b4328";
const VALKI_ID = "86528590-6ff1-4dfa-adf2-6de05ff7b89c";
const AKKI_ID = "6ee6cd34-c117-4d7e-97d1-8f8464bfaac8";

const FRONT_TEXT =
  "Whenever a Wolf or Werewolf you control deals combat damage to a player, draw a card.";
const BACK_TEXT =
  "Whenever a Wolf or Werewolf you control deals combat damage to a player, you may reveal the top card of your library.";

const tovolar = (extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: "tov",
    name: "Tovolar, Dire Overlord",
    owner: "me",
    controller: "me",
    known_by_you: true,
    scryfall_id: TOVOLAR_ID,
    layout: "transform",
    active_face: 0,
    type_line: "Legendary Creature — Human Werewolf",
    power: 3,
    toughness: 3,
    faces: [
      {
        name: "Tovolar, Dire Overlord",
        type_line: "Legendary Creature — Human Werewolf",
        mana_cost: "{1}{R}{G}",
        oracle_text: FRONT_TEXT,
        power: 3,
        toughness: 3,
        image: `/cards/${TOVOLAR_ID}/image?face=0`,
      },
      {
        name: "Tovolar, the Midnight Scourge",
        type_line: "Legendary Creature — Werewolf",
        oracle_text: BACK_TEXT,
        power: 4,
        toughness: 4,
        image: `/cards/${TOVOLAR_ID}/image?face=1`,
      },
    ],
    ...extra,
  }) as unknown as CardView;

// Scryfall's view of the same printings: the top level carries the
// joined name and NO rules text, the faces carry the rest.
const META: Record<string, unknown> = {
  [TOVOLAR_ID]: {
    id: TOVOLAR_ID,
    name: "Tovolar, Dire Overlord // Tovolar, the Midnight Scourge",
    type_line: "Legendary Creature — Human Werewolf // Legendary Creature — Werewolf",
    card_faces: [
      { name: "Tovolar, Dire Overlord", oracle_text: FRONT_TEXT, power: "3", toughness: "3" },
      { name: "Tovolar, the Midnight Scourge", oracle_text: BACK_TEXT, power: "4", toughness: "4" },
    ],
  },
  [VALKI_ID]: {
    id: VALKI_ID,
    name: "Valki, God of Lies // Tibalt, Cosmic Impostor",
    card_faces: [
      { name: "Valki, God of Lies", power: "2", toughness: "1", mana_cost: "{1}{B}" },
      { name: "Tibalt, Cosmic Impostor", loyalty: "5", mana_cost: "{5}{B}{R}" },
    ],
  },
};

let realFetch: unknown;
beforeEach(() => {
  hoveredCard.set(null);
  const g = globalThis as Record<string, unknown>;
  realFetch = g.fetch;
  g.fetch = (url: string) => {
    const id = String(url).split("/").pop() ?? "";
    const meta = META[id] ?? { id, name: "?" };
    return Promise.resolve({ ok: true, json: () => Promise.resolve(meta) } as unknown as Response);
  };
});
afterEach(() => {
  hoveredCard.set(null);
  // Leave no Shift held for the next test.
  window.dispatchEvent(new Event("blur"));
  cleanup();
  document.body.querySelectorAll("input").forEach((el) => el.remove());
  (globalThis as Record<string, unknown>).fetch = realFetch;
});

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await Promise.resolve();
    flushSync();
  }
}

function show(card: CardView): HTMLElement {
  const { container } = render(HoverZoomOverlay as never, { view } as never);
  hoveredCard.set(card);
  flushSync();
  return container;
}

const big = (c: HTMLElement) => c.querySelector<HTMLImageElement>(".scan > img");
const inset = (c: HTMLElement) => c.querySelector<HTMLImageElement>(".inset-col .other-face img");
const otherSide = (c: HTMLElement) => c.querySelector<HTMLElement>(".other-side");

function key(
  type: "keydown" | "keyup",
  target: EventTarget = window,
  init: KeyboardEventInit = {},
) {
  target.dispatchEvent(
    new KeyboardEvent(type, { key: "Shift", shiftKey: type === "keydown", bubbles: true, ...init }),
  );
  flushSync();
}

describe("the other face's text", () => {
  it("lists a transform card's back under a Back face heading", async () => {
    const c = show(tovolar());
    await settle();
    const section = otherSide(c);
    expect(section?.querySelector("h3")?.textContent).toBe("Back face");
    expect(section?.querySelector(".title")?.textContent).toBe("Tovolar, the Midnight Scourge");
    expect(section?.querySelector(".type-line")?.textContent).toContain(
      "Legendary Creature — Werewolf",
    );
    expect(section?.querySelector(".pt")?.textContent).toBe("4/4");
    expect(section?.querySelector(".oracle")?.textContent).toBe(BACK_TEXT);
    // The face that is up reads as itself, not as the joined Scryfall
    // name with no text.
    expect(c.querySelector(".info > .info-head .title")?.textContent).toBe(
      "Tovolar, Dire Overlord",
    );
    expect(c.querySelector(".info > .oracle")?.textContent).toBe(FRONT_TEXT);
  });

  it("calls the front the Front face once the back is up", async () => {
    const c = show(
      tovolar({
        active_face: 1,
        name: "Tovolar, the Midnight Scourge",
        power: 4,
        toughness: 4,
      }),
    );
    await settle();
    const section = otherSide(c);
    expect(section?.querySelector("h3")?.textContent).toBe("Front face");
    expect(section?.querySelector(".title")?.textContent).toBe("Tovolar, Dire Overlord");
    expect(section?.querySelector(".oracle")?.textContent).toBe(FRONT_TEXT);
    expect(c.querySelector(".info > .oracle")?.textContent).toBe(BACK_TEXT);
  });

  it("lists a modal DFC's other face with its cost and starting loyalty", async () => {
    const c = show({
      instance_id: "valki",
      name: "Valki, God of Lies",
      owner: "me",
      controller: "me",
      known_by_you: true,
      scryfall_id: VALKI_ID,
      layout: "modal_dfc",
      active_face: 0,
      faces: [
        { name: "Valki, God of Lies", type_line: "Legendary Creature — God", mana_cost: "{1}{B}" },
        {
          name: "Tibalt, Cosmic Impostor",
          type_line: "Legendary Planeswalker — Tibalt",
          mana_cost: "{5}{B}{R}",
          oracle_text: "+2: Exile the top card of each player's library.",
        },
      ],
    } as unknown as CardView);
    await settle();
    const section = otherSide(c);
    expect(section?.querySelector("h3")?.textContent).toBe("Back face");
    expect(section?.querySelector(".title")?.textContent).toBe("Tibalt, Cosmic Impostor");
    expect(section?.querySelector(".pt")?.textContent).toBe("◈ 5");
    expect(section?.querySelector(".cost")).not.toBeNull();
    expect(section?.querySelector(".oracle")?.textContent).toContain("+2: Exile");
  });

  it("lists a flip card's other half as Flipped, and its top half as Unflipped once flipped", () => {
    const akki = {
      instance_id: "akki",
      name: "Akki Lavarunner",
      owner: "me",
      controller: "me",
      known_by_you: true,
      scryfall_id: AKKI_ID,
      layout: "flip",
      active_face: 0,
      faces: [
        { name: "Akki Lavarunner", type_line: "Creature — Goblin Warrior", mana_cost: "{3}{R}" },
        {
          name: "Tok-Tok, Volcano Born",
          type_line: "Legendary Creature — Goblin Shaman",
          oracle_text: "Protection from red",
          power: 2,
          toughness: 2,
        },
      ],
    } as unknown as CardView;
    let c = show(akki);
    expect(otherSide(c)?.querySelector("h3")?.textContent).toBe("Flipped");
    expect(otherSide(c)?.querySelector(".title")?.textContent).toBe("Tok-Tok, Volcano Born");
    expect(otherSide(c)?.querySelector(".pt")?.textContent).toBe("2/2");
    cleanup();
    c = show({ ...akki, active_face: 1, name: "Tok-Tok, Volcano Born" } as CardView);
    expect(otherSide(c)?.querySelector("h3")?.textContent).toBe("Unflipped");
    expect(otherSide(c)?.querySelector(".title")?.textContent).toBe("Akki Lavarunner");
  });

  it("leaves a split card as it was: no other-face section, no flip hint", () => {
    const c = show({
      instance_id: "fi",
      name: "Fire // Ice",
      owner: "me",
      controller: "me",
      known_by_you: true,
      scryfall_id: "split-1",
      layout: "split",
      active_face: 0,
      faces: [
        {
          name: "Fire",
          type_line: "Instant",
          oracle_text: "Fire deals 2 damage divided as you choose",
        },
        { name: "Ice", type_line: "Instant", oracle_text: "Tap target permanent." },
      ],
    } as unknown as CardView);
    expect(otherSide(c)).toBeNull();
    expect(c.querySelector(".flip-hint")).toBeNull();
    expect(c.querySelector(".other-face")).toBeNull();
    key("keydown");
    expect(big(c)?.classList.contains("sideways")).toBe(true);
  });

  it("shows nothing more for a face-down card, even one its controller may look at", () => {
    const c = show(tovolar({ face_down: true, face_visible: true, face_down_kind: "manifested" }));
    expect(otherSide(c)).toBeNull();
    expect(c.querySelector(".flip-hint")).toBeNull();
    const before = big(c)?.getAttribute("src");
    key("keydown");
    expect(big(c)?.getAttribute("src")).toBe(before);
  });

  it("lists the permanent a meld card melds into", () => {
    const c = show({
      instance_id: "lp",
      name: "Urza, Lord Protector",
      owner: "me",
      controller: "me",
      known_by_you: true,
      scryfall_id: "urza-lp",
      melds_into: {
        name: "Urza, Planeswalker",
        type_line: "Legendary Planeswalker — Urza",
        oracle_text:
          "You may activate the loyalty abilities of Urza, Planeswalker twice each turn.",
        image: "/cards/pw/image",
      },
    } as unknown as CardView);
    expect(otherSide(c)?.querySelector("h3")?.textContent).toBe("Melds into");
    expect(otherSide(c)?.querySelector(".title")?.textContent).toBe("Urza, Planeswalker");
  });
});

describe("hold Shift to flip", () => {
  it("swaps the big image and the inset while Shift is held, and back on release", () => {
    const c = show(tovolar());
    const front = `/cards/${TOVOLAR_ID}/image?size=normal`;
    const back = `/cards/${TOVOLAR_ID}/image?size=normal&face=1`;
    expect(big(c)?.getAttribute("src")).toBe(front);
    expect(big(c)?.getAttribute("alt")).toBe("Tovolar, Dire Overlord");
    expect(inset(c)?.getAttribute("src")).toBe(back);
    expect(c.querySelector(".flip-hint")?.textContent).toBe(L.holdShiftToFlip);

    key("keydown");
    expect(big(c)?.getAttribute("src")).toBe(back);
    expect(big(c)?.getAttribute("alt")).toBe("Tovolar, the Midnight Scourge");
    expect(inset(c)?.getAttribute("src")).toBe(front);
    expect(inset(c)?.getAttribute("alt")).toBe("Tovolar, Dire Overlord");
    // The text does not move: both sides are already listed.
    expect(otherSide(c)?.querySelector("h3")?.textContent).toBe("Back face");

    key("keyup");
    expect(big(c)?.getAttribute("src")).toBe(front);
    expect(inset(c)?.getAttribute("src")).toBe(back);
  });

  it("turns a flip card's image to whichever half it shows", () => {
    const c = show({
      instance_id: "akki",
      name: "Akki Lavarunner",
      owner: "me",
      controller: "me",
      known_by_you: true,
      scryfall_id: AKKI_ID,
      layout: "flip",
      active_face: 0,
      faces: [{ name: "Akki Lavarunner" }, { name: "Tok-Tok, Volcano Born" }],
    } as unknown as CardView);
    expect(big(c)?.classList.contains("flipped")).toBe(false);
    expect(inset(c)?.classList.contains("flipped")).toBe(true);
    key("keydown");
    expect(big(c)?.classList.contains("flipped")).toBe(true);
    expect(big(c)?.getAttribute("alt")).toBe("Tok-Tok, Volcano Born");
    expect(inset(c)?.classList.contains("flipped")).toBe(false);
  });

  it("shows a meld card's meld result big, and the meld card as the inset", () => {
    const c = show({
      instance_id: "lp",
      name: "Urza, Lord Protector",
      owner: "me",
      controller: "me",
      known_by_you: true,
      scryfall_id: "urza-lp",
      melds_into: { name: "Urza, Planeswalker", image: "/cards/pw/image" },
    } as unknown as CardView);
    key("keydown");
    expect(big(c)?.getAttribute("src")).toBe("/cards/pw/image");
    const meld = c.querySelector<HTMLImageElement>(".meld-inset img");
    expect(meld?.getAttribute("src")).toBe("/cards/urza-lp/image?size=normal");
    key("keyup");
    expect(big(c)?.getAttribute("src")).toBe("/cards/urza-lp/image?size=normal");
  });

  it("ignores Shift typed into a text field", () => {
    const c = show(tovolar());
    const input = document.createElement("input");
    input.type = "text";
    document.body.appendChild(input);
    input.focus();
    key("keydown", input);
    expect(big(c)?.getAttribute("src")).toBe(`/cards/${TOVOLAR_ID}/image?size=normal`);
  });

  it("ignores a key repeat that arrives without the press", () => {
    const c = show(tovolar());
    key("keydown", window, { repeat: true });
    expect(big(c)?.getAttribute("src")).toBe(`/cards/${TOVOLAR_ID}/image?size=normal`);
  });

  it("lets go when the window loses focus or the tab is hidden", () => {
    const c = show(tovolar());
    const front = `/cards/${TOVOLAR_ID}/image?size=normal`;
    key("keydown");
    expect(big(c)?.getAttribute("src")).not.toBe(front);
    window.dispatchEvent(new Event("blur"));
    flushSync();
    expect(big(c)?.getAttribute("src")).toBe(front);

    key("keydown");
    expect(big(c)?.getAttribute("src")).not.toBe(front);
    document.dispatchEvent(new Event("visibilitychange"));
    flushSync();
    expect(big(c)?.getAttribute("src")).toBe(front);
  });
});
