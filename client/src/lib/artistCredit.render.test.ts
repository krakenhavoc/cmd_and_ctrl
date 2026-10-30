// @vitest-environment jsdom
//
// #2982: Scryfall's art_crop image drops the artist line and their image
// guidelines ask that it not be hidden, so every surface where an art
// crop is the main view of a card prints "Illus. <artist>" itself, from
// GET /cards/{id}. A surface that shows the full card image (the hover
// zoom, the hand, the catalogue) prints the line on the card and is
// never given an overlay.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import BattlefieldRow from "./components/board/BattlefieldRow.svelte";
import Hand from "./components/board/Hand.svelte";
import HoverZoomOverlay from "./components/board/HoverZoomOverlay.svelte";
import StackLaneFan from "./components/board/StackLaneFan.svelte";
import DeckCard from "./components/decks/DeckCard.svelte";
import DeckGrid from "./components/DeckGrid.svelte";
import SeatTile from "./components/SeatTile.svelte";
import { botDeckTile, libraryTile, prebuiltTile } from "./deckPicker";
import { artistOf, creditText } from "./artistCredit.svelte";
import type { CardMeta } from "./cardMetaCache";
import { hoveredCard } from "./cardTypes";
import type { CardView, GameView, PlayerView, StackItemView } from "./protocol";
import { buildStackLane, type StackLaneControls, type StackLaneItem } from "./stackLane";
import { cleanup, flushSync, render } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

// One printing per test: cardMetaCache is module-level and keeps what it
// has fetched, so a reused id would answer from an earlier test.
let nextID = 0;
const newID = () => `00000000-0000-4000-8000-${String(++nextID).padStart(12, "0")}`;

const metaByID = new Map<string, Partial<CardMeta>>();
const requested: string[] = [];
let realFetch: unknown;

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  realFetch = g.fetch;
  g.fetch = (url: string) => {
    requested.push(url);
    const id = url.replace(/^\/cards\//, "").split("?")[0];
    const meta = metaByID.get(id);
    return Promise.resolve({
      ok: meta !== undefined,
      json: () => Promise.resolve({ id, name: "x", ...meta }),
    } as unknown as Response);
  };
  requested.length = 0;
});
afterEach(() => {
  cleanup();
  hoveredCard.set(null);
  (globalThis as Record<string, unknown>).fetch = realFetch;
  metaByID.clear();
});

async function settle(): Promise<void> {
  await new Promise((r) => setTimeout(r, 0));
  flushSync();
}

const bear = (scryfall_id: string, extra: Partial<CardView> = {}): CardView => ({
  instance_id: "bear",
  name: "Grizzly Bears",
  owner: "me",
  controller: "me",
  type_line: "Creature — Bear",
  power: 2,
  toughness: 2,
  scryfall_id,
  known_by_you: true,
  ...extra,
});

function row(cards: CardView[]) {
  return render(
    BattlefieldRow as never,
    {
      label: "creatures",
      cards,
      viewerID: "me",
      attachmentsByHost: {},
      onCardClick: () => {},
    } as never,
  );
}

describe("artistOf and creditText", () => {
  it("reads the top-level artist, preferring the face shown", () => {
    expect(artistOf({ id: "a", name: "n", artist: "Top" })).toBe("Top");
    const dfc = {
      id: "a",
      name: "n",
      artist: "Front",
      card_faces: [
        { name: "F", artist: "Front" },
        { name: "B", artist: "Back" },
      ],
    };
    expect(artistOf(dfc, 0)).toBe("Front");
    expect(artistOf(dfc, 1)).toBe("Back");
    // A face that names none falls back to the card's.
    expect(artistOf({ ...dfc, card_faces: [{ name: "F" }, { name: "B" }] }, 1)).toBe("Front");
  });

  it("is null when there is no metadata or no artist", () => {
    expect(artistOf(null)).toBeNull();
    expect(artistOf({ id: "a", name: "n" })).toBeNull();
    expect(artistOf({ id: "a", name: "n", artist: "" })).toBeNull();
  });

  it("prints one string everywhere", () => {
    expect(creditText("Rebecca Guay")).toBe("Illus. Rebecca Guay");
  });
});

describe("the board's art tile", () => {
  it("credits the artist under the name strip once the metadata lands", async () => {
    const id = newID();
    metaByID.set(id, { artist: "Rebecca Guay" });
    const { container } = row([bear(id)]);
    const tile = container.querySelector<HTMLElement>(".card")!;
    expect(tile.classList.contains("art-tile")).toBe(true);
    // Nothing until the metadata arrives, and no layout reserved for it.
    expect(container.querySelector(".art-credit")).toBeNull();
    expect(tile.classList.contains("has-credit")).toBe(false);

    await settle();
    const credit = container.querySelector(".art-credit");
    expect(credit?.textContent).toBe("Illus. Rebecca Guay");
    expect(tile.classList.contains("has-credit")).toBe(true);
    // Inside the tile, after the name strip, and decorative for assistive tech.
    expect(credit?.previousElementSibling?.classList.contains("art-name")).toBe(true);
    expect(credit?.getAttribute("aria-hidden")).toBe("true");
    // The tooltip carries it too, for a tile too narrow to print it.
    expect(tile.getAttribute("title")).toBe("Grizzly Bears (Illus. Rebecca Guay)");
    // One metadata fetch, for the printing.
    expect(requested).toEqual([`/cards/${id}`]);
  });

  it("credits the face that is up on a double-faced card", async () => {
    const id = newID();
    metaByID.set(id, {
      artist: "Front Artist",
      card_faces: [
        { name: "Front", artist: "Front Artist" },
        { name: "Back", artist: "Back Artist" },
      ],
    });
    const { container } = row([bear(id, { active_face: 1 })]);
    await settle();
    expect(container.querySelector(".art-credit")?.textContent).toBe("Illus. Back Artist");
  });

  it("shows no credit, and keeps the tile as it was, when the printing names no artist", async () => {
    const id = newID();
    metaByID.set(id, {});
    const { container } = row([bear(id)]);
    await settle();
    expect(container.querySelector(".art-credit")).toBeNull();
    expect(container.querySelector(".card")!.classList.contains("has-credit")).toBe(false);
    expect(container.querySelector(".card")!.getAttribute("title")).toBe("Grizzly Bears");
  });

  it("shows no credit when the metadata request fails", async () => {
    const id = newID(); // not in metaByID, so the stub answers not-ok
    const { container } = row([bear(id)]);
    await settle();
    expect(container.querySelector(".art-credit")).toBeNull();
  });
});

describe("full card images carry no credit overlay", () => {
  it("your hand's full card prints its own artist line and asks for no metadata", async () => {
    const id = newID();
    metaByID.set(id, { artist: "Rebecca Guay" });
    const cards = [bear(id)];
    const { container } = render(
      Hand as never,
      {
        hand: { kind: "hand", owner: "me", count: 1, cards },
        isSelf: true,
        viewerID: "me",
      } as never,
    );
    await settle();
    expect(container.querySelector("img")?.getAttribute("src")).toContain("size=small");
    expect(container.querySelector(".art-credit")).toBeNull();
    expect(requested).toEqual([]);
  });

  it("the hover zoom shows the full card, with no credit element laid over it", async () => {
    const id = newID();
    metaByID.set(id, { artist: "Rebecca Guay" });
    const view = {
      id: "g",
      state: "active",
      seats: [],
      battlefield: { kind: "battlefield", count: 0, cards: [] },
      stack: { kind: "stack", count: 0, cards: [] },
      exile: { kind: "exile", count: 0, cards: [] },
      turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main", step: "main" },
    } as unknown as GameView;
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set(bear(id));
    flushSync();
    await settle();
    const srcs = [...container.querySelectorAll("img")].map((i) => i.getAttribute("src"));
    expect(srcs).toContain(`/cards/${id}/image?size=normal`);
    expect(srcs.some((s) => s?.includes("art_crop"))).toBe(false);
    expect(container.querySelector(".art-credit")).toBeNull();
  });
});

describe("the stack lane's fan tile", () => {
  const ME = "me";
  const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
    kind,
    owner,
    count: cards.length,
    cards,
  });
  const seat = {
    id: ME,
    name: "Me",
    seat: 0,
    life: 40,
    library: zone("library", ME),
    hand: zone("hand", ME),
    graveyard: zone("graveyard", ME),
    command: zone("command", ME),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  } as unknown as PlayerView;

  function mountFan(printing: string) {
    const bolt: CardView = {
      instance_id: "bolt",
      name: "Lightning Bolt",
      owner: ME,
      controller: ME,
      known_by_you: true,
      type_line: "Instant",
      scryfall_id: printing,
    } as CardView;
    const items: StackItemView[] = [
      {
        id: "bolt",
        kind: "spell",
        controller: ME,
        owner: ME,
        source_card_id: "bolt",
        targets: [],
      } as StackItemView,
    ];
    const model = buildStackLane({
      stack: zone("stack", undefined, [bolt]),
      stackItems: items,
      pendingTriggers: [],
      seats: [seat],
      battlefield: zone("battlefield", undefined),
      exile: zone("exile", undefined),
      viewerID: ME,
      priorityHolder: ME,
      splitSecondActive: false,
    });
    const controls: StackLaneControls = {
      counter: () => {},
      targetable: () => false,
      target: () => {},
      hoverEnter: (_it: StackLaneItem) => {},
      hoverLeave: (_it: StackLaneItem) => {},
    };
    return render(StackLaneFan as never, { model, controls, styleName: "fan" } as never);
  }

  it("credits the artist over the foot of the spell's art crop", async () => {
    const id = newID();
    metaByID.set(id, { artist: "Christopher Moeller" });
    const { container } = mountFan(id);
    expect(container.querySelector(".art img")?.getAttribute("src")).toContain("size=art_crop");
    await settle();
    const credit = container.querySelector(".art .art-credit");
    expect(credit?.textContent).toBe("Illus. Christopher Moeller");
  });

  it("prints nothing when the printing names no artist", async () => {
    const id = newID();
    metaByID.set(id, {});
    const { container } = mountFan(id);
    await settle();
    expect(container.querySelector(".art-credit")).toBeNull();
  });
});

// ADR 0158 §2: a saved deck on the Decks page is led by its commander's
// art crop, so the crop carries the credit, one per partner.
describe("the Decks grid's commander art", () => {
  function mountDeck(...ids: string[]) {
    const deck = {
      id: "d1",
      name: "Partners",
      commanders: ids.map((_, i) => `Commander ${i}`),
      card_count: 100,
      updated_at: "2026-10-01T00:00:00Z",
      commander_cards: ids.map((id, i) => ({
        name: `Commander ${i}`,
        scryfall_id: id,
        color_identity: ["W"],
      })),
      color_identity: ["W"],
    };
    const noop = () => {};
    return render(
      DeckCard as never,
      {
        deck,
        reportOpen: false,
        canRequest: false,
        requesting: false,
        requestLocked: false,
        busy: false,
        onReport: noop,
        onRequest: noop,
        onRename: async () => "",
        onDelete: async () => {},
      } as never,
    );
  }

  it("credits each commander's artist on its own crop", async () => {
    const a = newID();
    const b = newID();
    metaByID.set(a, { artist: "Winona Nelson" });
    metaByID.set(b, { artist: "Jason Rainville" });
    const { container } = mountDeck(a, b);
    const crops = [...container.querySelectorAll<HTMLElement>(".band .crop")];
    expect(crops.map((c) => c.querySelector("img")?.getAttribute("src"))).toEqual([
      `/cards/${a}/image?size=art_crop`,
      `/cards/${b}/image?size=art_crop`,
    ]);
    await settle();
    expect(crops.map((c) => c.querySelector(".art-credit")?.textContent)).toEqual([
      "Illus. Winona Nelson",
      "Illus. Jason Rainville",
    ]);
    expect(crops[0].getAttribute("title")).toBe("Commander 0 (Illus. Winona Nelson)");
    expect(requested).toEqual([`/cards/${a}`, `/cards/${b}`]);
  });

  it("prints nothing on a crop whose printing names no artist", async () => {
    const id = newID();
    metaByID.set(id, {});
    const { container } = mountDeck(id);
    await settle();
    expect(container.querySelector(".band img")).not.toBeNull();
    expect(container.querySelector(".art-credit")).toBeNull();
  });
});

// ADR 0212: the deck picker's tiles and your seat's tile are led by the
// commander's art crop too, so each crop carries the credit.
describe("the deck picker's tiles and your seat's tile", () => {
  // An observer that reports every node in view at once, and one that
  // never does (a tile scrolled far out of the grid).
  class SeesAll {
    constructor(private cb: (e: { isIntersecting: boolean }[]) => void) {}
    observe(): void {
      this.cb([{ isIntersecting: true }]);
    }
    unobserve(): void {}
    disconnect(): void {}
  }
  let saved: unknown;
  beforeEach(() => {
    saved = (globalThis as Record<string, unknown>).IntersectionObserver;
  });
  afterEach(() => {
    (globalThis as Record<string, unknown>).IntersectionObserver = saved;
  });

  function tile(...ids: string[]) {
    return libraryTile({
      id: "d1",
      name: "Partners",
      commanders: ids.map((_, i) => `Commander ${i}`),
      card_count: 100,
      updated_at: "2026-10-01T00:00:00Z",
      commander_cards: ids.map((id, i) => ({
        name: `Commander ${i}`,
        scryfall_id: id,
        color_identity: ["W"],
      })),
      color_identity: ["W"],
    });
  }

  it("credits each crop on a picker tile once it scrolls into view", async () => {
    (globalThis as Record<string, unknown>).IntersectionObserver = SeesAll;
    const a = newID();
    const b = newID();
    metaByID.set(a, { artist: "Winona Nelson" });
    metaByID.set(b, { artist: "Jason Rainville" });
    const { container } = render(
      DeckGrid as never,
      {
        tiles: [tile(a, b)],
        name: "your-deck",
        label: "your decks",
        selected: "d1",
      } as never,
    );
    await settle();
    const crops = [...container.querySelectorAll<HTMLElement>(".dtile .crop")];
    expect(crops.map((c) => c.querySelector(".art-credit")?.textContent)).toEqual([
      "Illus. Winona Nelson",
      "Illus. Jason Rainville",
    ]);
  });

  // ADR 0212 phase B1: a pre-built deck's tile, and the same deck in the
  // Add bot grid, are led by the commander's crop and credit it too.
  function preDeck(id: string) {
    return {
      id: "fury",
      name: "Forge and Fury",
      commander: "Purphoros, God of the Forge",
      colors: ["R"],
      card_count: 100,
      commander_cards: [
        { name: "Purphoros, God of the Forge", scryfall_id: id, color_identity: ["R"] },
      ],
      coverage: { cards: 73, full: 73, caveats: 0, unreviewed: 0, basics: 1, unregistered: 0 },
    };
  }

  it("credits a pre-built tile's commander crop, in the picker and in the Add bot grid", async () => {
    (globalThis as Record<string, unknown>).IntersectionObserver = SeesAll;
    const id = newID();
    metaByID.set(id, { artist: "Slawomir Maniak" });
    const grids = [
      [prebuiltTile(preDeck(id))],
      [botDeckTile({ id: "fury", name: "Fury", commander: "Purphoros" }, preDeck(id))],
    ];
    for (const tiles of grids) {
      const { container } = render(
        DeckGrid as never,
        { tiles, name: "prebuilt-deck", label: "pre-built decks", selected: "fury" } as never,
      );
      await settle();
      const crop = container.querySelector<HTMLElement>(".dtile .crop")!;
      expect(crop.querySelector("img")?.getAttribute("src")).toBe(
        `/cards/${id}/image?size=art_crop`,
      );
      expect(crop.querySelector("img")?.getAttribute("loading")).toBe("lazy");
      expect(crop.querySelector(".art-credit")?.textContent).toBe("Illus. Slawomir Maniak");
      cleanup();
    }
  });

  it("washes a pre-built tile whose commander has no printing, and asks for nothing", async () => {
    (globalThis as Record<string, unknown>).IntersectionObserver = SeesAll;
    const { container } = render(
      DeckGrid as never,
      {
        tiles: [
          prebuiltTile({ ...preDeck(""), commander_cards: [{ name: "X", color_identity: ["R"] }] }),
        ],
        name: "prebuilt-deck",
        label: "pre-built decks",
        selected: "",
      } as never,
    );
    await settle();
    expect(container.querySelector(".dtile img")).toBeNull();
    expect(container.querySelector(".dtile .art.wash")).not.toBeNull();
    expect(requested).toEqual([]);
  });

  it("credits the crop on a bot seat's tile", async () => {
    const id = newID();
    metaByID.set(id, { artist: "Slawomir Maniak" });
    const { container } = render(
      SeatTile as never,
      {
        seat: 3,
        p: {
          player_id: "b1",
          name: "Bot 1",
          seat: 3,
          deck_uploaded: true,
          is_bot: true,
          bot_tier: "heuristic",
        },
        text: {
          what: "Purphoros, God of the Forge",
          second: "Pre-built · Forge and Fury",
          colors: ["R"],
          ready: true,
        },
        art: [{ name: "Purphoros, God of the Forge", scryfallId: id }],
        botChip: "Medium",
      } as never,
    );
    await settle();
    const li = container.querySelector("li.tile")!;
    expect(li.querySelector(".crop img")?.getAttribute("src")).toBe(
      `/cards/${id}/image?size=art_crop`,
    );
    expect(li.querySelector(".crop .art-credit")?.textContent).toBe("Illus. Slawomir Maniak");
    expect(li.textContent).toMatch(/Purphoros, God of the Forge/);
    expect(li.textContent).toMatch(/Ready/);
  });

  it("loads nothing for a tile not yet in view", async () => {
    (globalThis as Record<string, unknown>).IntersectionObserver = FakeObserver;
    const id = newID();
    metaByID.set(id, { artist: "Winona Nelson" });
    const { container } = render(
      DeckGrid as never,
      {
        tiles: [tile(id)],
        name: "your-deck",
        label: "your decks",
        selected: "",
      } as never,
    );
    await settle();
    expect(container.querySelector(".dtile img")).toBeNull();
    expect(requested).toEqual([]);
  });

  it("credits the crop in your seat's preview, clear of the ×", async () => {
    const id = newID();
    metaByID.set(id, { artist: "Kieran Yanner" });
    const { container } = render(
      SeatTile as never,
      {
        seat: 0,
        p: { player_id: "p1", name: "Luke", seat: 0, deck_uploaded: false },
        text: {
          what: "Kenrith, the Returned King",
          second: "Kenrith's court",
          colors: ["W"],
          ready: false,
        },
        art: [{ name: "Kenrith, the Returned King", scryfallId: id }],
        preview: true,
        you: true,
        remove: { label: "leave your seat", title: "", disabled: false, onclick: () => {} },
      } as never,
    );
    await settle();
    const li = container.querySelector("li.tile")!;
    expect(li.classList.contains("has-x")).toBe(true);
    expect(li.querySelector(".crop .art-credit")?.textContent).toBe("Illus. Kieran Yanner");
    expect(li.textContent).toMatch(/Preview/);
    expect(li.textContent).toMatch(/Not seated yet/);
  });
});
