// @vitest-environment jsdom
//
// rulings.render.test.ts — #3831, on the REAL Board. A card's right-click
// menu ends in "Rulings", a real link to the card's official rulings that
// opens in a new tab: the ability popover by default, the override menu
// with admin overrides on. It is offered for any card the viewer can see,
// whoever controls it, and never for a card the viewer cannot see, a
// face-down card or a token.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import Board from "./components/board/Board.svelte";
import { resetSettings, updateSettings } from "./settings";
import { closeAbilityPopover } from "./abilityPopover";
import { closeCardMenu } from "./contextMenu";
import type { CardView, GameView, PlayerView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

const ME = "me";
const THEM = "them";

// GET /cards/{id}, by Scryfall id. cardMetaCache is module-level and keeps
// what it fetched, so each test names its own printing.
const META: Record<string, unknown> = {
  "sc-gatherer": {
    id: "sc-gatherer",
    name: "Rampant Growth",
    set: "lea",
    collector_number: "1",
    multiverse_ids: [4321],
  },
  "sc-scryfall": { id: "sc-scryfall", name: "Sol Ring", set: "ecc", collector_number: "57" },
  "sc-theirs": {
    id: "sc-theirs",
    name: "Swords to Plowshares",
    set: "ice",
    collector_number: "9",
    multiverse_ids: [777],
  },
  "sc-back": {
    id: "sc-back",
    name: "Delver of Secrets // Insectile Aberration",
    set: "isd",
    collector_number: "51",
    multiverse_ids: [226749, 226755],
  },
};

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string) => {
      const id = String(url).replace(/^\/cards\//, "");
      const meta = META[id];
      return meta
        ? ({ ok: true, json: async () => meta } as Response)
        : ({ ok: false, json: async () => ({}) } as Response);
    }),
  );
});

afterEach(() => {
  cleanup();
  closeAbilityPopover();
  closeCardMenu();
  resetSettings();
  vi.unstubAllGlobals();
});

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const card = (id: string, name: string, owner: string, extras: Partial<CardView> = {}): CardView =>
  ({
    instance_id: id,
    name,
    owner,
    controller: owner,
    known_by_you: true,
    type_line: "Sorcery",
    mana_cost: "{G}",
    ...extras,
  }) as CardView;

const seat = (id: string, idx: number, hand: CardView[]): PlayerView =>
  ({
    id,
    name: id === ME ? "Me" : "Them",
    seat: idx,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id, hand),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

function gameView(o: { hand?: CardView[]; battlefield?: CardView[] }): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat(ME, 0, o.hand ?? []), seat(THEM, 1, [])],
    battlefield: zone("battlefield", undefined, o.battlefield ?? []),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    legal_moves: [{ type: "pass_priority", player: ME, kind: "pass", label: "Pass" }],
  } as unknown as GameView;
}

function mountBoard(o: Parameters<typeof gameView>[0]) {
  const r = render(
    Board as never,
    {
      view: gameView(o),
      viewerID: ME,
      isAdmin: false,
      sendAction: () => {},
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
    } as never,
  );
  return {
    ...r,
    cardNamed: (name: string) =>
      r.container.querySelector<HTMLElement>(`.card[aria-label='${name}']`),
  };
}

async function rightClick(el: HTMLElement): Promise<void> {
  el.dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
  flushSync();
  // metaFor's fetch resolves a few microtasks later.
  await new Promise((r) => setTimeout(r, 0));
  flushSync();
}

const link = (c: HTMLElement) => c.querySelector<HTMLAnchorElement>("a[data-rulings]");

describe("the ability popover (the default right-click)", () => {
  it("ends in a Rulings link to Gatherer, in a new tab, for a card in hand", async () => {
    const b = mountBoard({
      hand: [card("g", "Rampant Growth", ME, { scryfall_id: "sc-gatherer" })],
    });
    await rightClick(b.cardNamed("Rampant Growth")!);
    const a = link(b.container)!;
    expect(a, "a plain spell's right-click opens the menu with Rulings").toBeTruthy();
    expect(a.getAttribute("role")).toBe("menuitem");
    expect(a.closest("[role='menu']")).toBeTruthy();
    expect(a.textContent).toContain("Rulings");
    expect(a.getAttribute("href")).toBe(
      "https://gatherer.wizards.com/Pages/Card/Details.aspx?multiverseid=4321",
    );
    expect(a.getAttribute("target")).toBe("_blank");
    expect(a.getAttribute("rel")).toBe("noopener noreferrer");
  });

  it("falls back to the Scryfall rulings section when Gatherer has no id", async () => {
    const b = mountBoard({ hand: [card("s", "Sol Ring", ME, { scryfall_id: "sc-scryfall" })] });
    await rightClick(b.cardNamed("Sol Ring")!);
    expect(link(b.container)!.getAttribute("href")).toBe(
      "https://scryfall.com/card/ecc/57#rulings",
    );
  });

  it("is on an opponent's permanent too", async () => {
    const b = mountBoard({
      battlefield: [
        card("sw", "Swords to Plowshares", THEM, {
          scryfall_id: "sc-theirs",
          type_line: "Instant",
        }),
      ],
    });
    await rightClick(b.cardNamed("Swords to Plowshares")!);
    expect(link(b.container)!.getAttribute("href")).toBe(
      "https://gatherer.wizards.com/Pages/Card/Details.aspx?multiverseid=777",
    );
  });

  it("opens the side of a double-faced permanent that is up", async () => {
    const b = mountBoard({
      battlefield: [
        card("dv", "Insectile Aberration", ME, {
          scryfall_id: "sc-back",
          type_line: "Creature — Human Insect",
          active_face: 1,
        }),
      ],
    });
    await rightClick(b.cardNamed("Insectile Aberration")!);
    expect(link(b.container)!.getAttribute("href")).toBe(
      "https://gatherer.wizards.com/Pages/Card/Details.aspx?multiverseid=226755",
    );
  });

  it("is not offered for a card the viewer cannot see", async () => {
    // A redacted card arrives with no scryfall_id.
    const b = mountBoard({
      battlefield: [card("hid", "hidden", THEM, { known_by_you: false, type_line: "" })],
    });
    const el = b.container.querySelector<HTMLElement>(".card");
    expect(el).toBeTruthy();
    await rightClick(el!);
    expect(link(b.container)).toBeNull();
    expect(b.container.querySelector("[role='menu']")).toBeNull();
  });

  it("is not offered for a face-down card, even the viewer's own", async () => {
    const b = mountBoard({
      battlefield: [
        card("fd", "Hooded Hydra", ME, {
          scryfall_id: "sc-gatherer",
          face_down: true,
          face_down_kind: "manifested",
          face_visible: true,
        }),
      ],
    });
    const el = b.container.querySelector<HTMLElement>(".card");
    await rightClick(el!);
    expect(link(b.container)).toBeNull();
  });

  it("is not offered for a token", async () => {
    const b = mountBoard({
      battlefield: [
        card("tk", "Soldier", ME, {
          scryfall_id: "sc-scryfall",
          is_token: true,
          type_line: "Token Creature — Soldier",
        }),
      ],
    });
    await rightClick(b.cardNamed("Soldier")!);
    expect(link(b.container)).toBeNull();
  });
});

describe("the override menu (admin overrides on)", () => {
  beforeEach(() => {
    updateSettings("gameplay", "adminOverrides", true);
  });

  it("ends in the same Rulings link, on a card the viewer controls", async () => {
    const b = mountBoard({
      hand: [card("g", "Rampant Growth", ME, { scryfall_id: "sc-gatherer" })],
    });
    await rightClick(b.cardNamed("Rampant Growth")!);
    const menu = b.container.querySelector("[role='menu'][aria-label='card overrides']");
    expect(menu).toBeTruthy();
    const a = link(menu as HTMLElement)!;
    expect(a.getAttribute("href")).toBe(
      "https://gatherer.wizards.com/Pages/Card/Details.aspx?multiverseid=4321",
    );
    expect(a.getAttribute("target")).toBe("_blank");
    expect(a.getAttribute("rel")).toBe("noopener noreferrer");
  });

  it("is the one row on a card the viewer cannot override", async () => {
    const b = mountBoard({
      battlefield: [
        card("sw", "Swords to Plowshares", THEM, {
          scryfall_id: "sc-theirs",
          type_line: "Instant",
        }),
      ],
    });
    await rightClick(b.cardNamed("Swords to Plowshares")!);
    const items = b.container.querySelectorAll("[role='menu'] [role='menuitem']");
    expect(items.length).toBe(1);
    expect(items[0].textContent).toContain("Rulings");
  });

  it("is not offered for a face-down card or a token", async () => {
    const b = mountBoard({
      battlefield: [
        card("fd", "Hooded Hydra", ME, { scryfall_id: "sc-gatherer", face_down: true }),
        card("tk", "Soldier", ME, { scryfall_id: "sc-scryfall", is_token: true }),
      ],
    });
    for (const el of b.container.querySelectorAll<HTMLElement>(".card")) {
      await rightClick(el);
      expect(link(b.container)).toBeNull();
      closeCardMenu();
      flushSync();
    }
  });
});
