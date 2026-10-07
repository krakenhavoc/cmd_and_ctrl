// @vitest-environment jsdom
//
// playmat.render.test.ts — ADR 0128. Two surfaces:
//
//   - the board: a seat's panel draws its owner's playmat behind the
//     zones, under the per-device all / mine / off setting, with no
//     broken-image box when the file will not load, and never a URL
//     that is not the server's own route;
//   - the owner-set wash on the seat's mat.
//
// The Settings "Playmat" tab is in playmatSettings.render.test.ts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { flushSync } from "svelte";

vi.mock("./sounds", () => ({ play: () => {} }));

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import type { GameView, PlayerView } from "./protocol";
import { resetSettings, updateSettings } from "./settings";
import { session, type Session } from "./session";
import { cleanup, render } from "./test/render.svelte";

const ME = "me";
const BOB = "bob";
const MINE = "/playmats/11111111-1111-4111-8111-111111111111";
const THEIRS = "/playmats/22222222-2222-4222-8222-222222222222";

beforeEach(() => {
  resetSettings();
  session.set(null);
});

afterEach(() => {
  cleanup();
  session.set(null);
});

const zone = (kind: string, owner: string | undefined) => ({ kind, owner, count: 0, cards: [] });

const seat = (id: string, n: number, playmat?: string): PlayerView =>
  ({
    id,
    name: id === ME ? "Me" : "Bob",
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
    playmat_url: playmat,
  }) as unknown as PlayerView;

const gameView = (mine?: string, theirs?: string): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, 0, mine), seat(BOB, 1, theirs)],
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "precombat_main" },
    mulligans_open: false,
  }) as unknown as GameView;

function mountPanel(view: GameView, own: boolean, extra: Record<string, unknown> = {}) {
  return render(
    PlayerPanel as never,
    {
      seat: own ? view.seats[0] : view.seats[1],
      isSelf: own,
      isActive: own,
      hasPriority: own,
      viewerID: ME,
      isAdmin: false,
      sendAction: () => {},
      isMonarch: false,
      isInitiative: false,
      view,
      controlledCards: [],
      exile: zone("exile", undefined),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
      onTapToggle: () => {},
      onPlayCard: () => {},
      onDrawCard: () => {},
      onActivateAbility: () => {},
      onManaAbilityCost: () => {},
      ...extra,
    } as never,
  );
}

// vitest runs from client/; the component is read as source because jsdom
// does not compute a component's scoped CSS.
const panelSource = () =>
  readFileSync(resolve(process.cwd(), "src/lib/components/board/PlayerPanel.svelte"), "utf8");

const matIn = (c: HTMLElement) => c.querySelector<HTMLImageElement>(".playmat img");

describe("a seat's playmat on the board", () => {
  it("is drawn behind the owner's zones, as an image with no alt text", () => {
    const r = mountPanel(gameView(MINE, THEIRS), true);
    const img = matIn(r.container)!;
    expect(img).not.toBeNull();
    expect(img.getAttribute("src")).toBe(MINE);
    expect(img.getAttribute("alt")).toBe("");
    const layer = r.container.querySelector(".playmat")!;
    expect(layer.getAttribute("aria-hidden")).toBe("true");
    // Under everything: the layer is the panel's first child.
    expect(layer.parentElement!.firstElementChild).toBe(layer);
  });

  it("is not drawn for a seat with none", () => {
    const r = mountPanel(gameView(undefined, THEIRS), true);
    expect(r.container.querySelector(".playmat")).toBeNull();
  });

  it("draws each seat's own mat, not the viewer's, on an opponent's panel", () => {
    const r = mountPanel(gameView(MINE, THEIRS), false);
    expect(matIn(r.container)!.getAttribute("src")).toBe(THEIRS);
  });

  it("rides the session token, since an <img> cannot send a header", () => {
    session.set({
      token: "tok",
      expiresAt: "2999-01-01T00:00:00Z",
      principal: { role: "identified" },
    } as unknown as Session);
    const r = mountPanel(gameView(MINE), true);
    expect(matIn(r.container)!.getAttribute("src")).toBe(`${MINE}?token=tok`);
  });

  describe("under the per-device setting", () => {
    it("all (the default) draws every seat's", () => {
      expect(matIn(mountPanel(gameView(MINE, THEIRS), true).container)).not.toBeNull();
      expect(matIn(mountPanel(gameView(MINE, THEIRS), false).container)).not.toBeNull();
    });

    it("mine draws the viewer's own and no one else's", () => {
      updateSettings("display", "playmats", "mine");
      const own = mountPanel(gameView(MINE, THEIRS), true);
      const other = mountPanel(gameView(MINE, THEIRS), false);
      expect(matIn(own.container)).not.toBeNull();
      expect(other.container.querySelector(".playmat")).toBeNull();
    });

    it("off draws none, and following the setting is live", () => {
      const r = mountPanel(gameView(MINE, THEIRS), true);
      expect(matIn(r.container)).not.toBeNull();
      updateSettings("display", "playmats", "off");
      flushSync();
      expect(r.container.querySelector(".playmat")).toBeNull();
      updateSettings("display", "playmats", "all");
      flushSync();
      expect(matIn(r.container)).not.toBeNull();
    });
  });

  it("falls back to the plain board when the image will not load", () => {
    const r = mountPanel(gameView(MINE, THEIRS), true);
    const img = matIn(r.container)!;
    img.dispatchEvent(new Event("error"));
    flushSync();
    expect(r.container.querySelector(".playmat")).toBeNull();
    expect(r.container.querySelector(".panel")).not.toBeNull();
  });

  it("never loads a URL that is not the server's own playmat route", () => {
    for (const bad of [
      "https://evil.example/tracker.png",
      "//evil.example/x.png",
      "/avatars/1/2.png",
      "javascript:alert(1)",
    ]) {
      const r = mountPanel(gameView(bad), true);
      expect(r.container.querySelector(".playmat"), bad).toBeNull();
      expect(r.container.innerHTML).not.toContain("evil.example");
      r.destroy();
    }
  });

  it("is a background layer: the panel is a stacking context and the mat takes no input", () => {
    // jsdom does not compute scoped component CSS, so this pins the
    // rules in the source, which is where a regression would be made.
    const src = panelSource();
    expect(src).toMatch(
      /\.playmat\s*\{[^}]*position:\s*absolute;[^}]*z-index:\s*-1;[^}]*pointer-events:\s*none;/s,
    );
    expect(src).toMatch(/\.panel\s*\{[^}]*isolation:\s*isolate;/s);
    expect(src).toMatch(/object-fit:\s*cover;/);
    // The scrim is the page's own background colour, so it follows the skin.
    expect(src).toMatch(/\.playmat::after\s*\{[^}]*color-mix\(in srgb, var\(--bg\)/s);
  });

  it("is turned with an across-table opponent's board, so it faces its owner", () => {
    const src = panelSource();
    expect(src).toMatch(
      /@media \(min-width: 600px\)\s*\{\s*\.panel\.opponent\.flipped \.playmat img\s*\{\s*transform:\s*rotate\(180deg\);/,
    );
    // And a flipped panel still mounts the layer.
    const r = mountPanel(gameView(MINE, THEIRS), false, { flipped: true });
    expect(r.container.querySelector(".panel.opponent.flipped .playmat img")).not.toBeNull();
  });
});

// ADR 0128 amendment: the owner sets how dark their playmat is, and
// every viewer draws it at that strength.
describe("the owner-set wash", () => {
  const washOf = (el: HTMLElement | null) => el!.style.getPropertyValue("--playmat-wash");

  it("is drawn on the seat's mat, and the default when the server sends none", () => {
    const view = gameView(MINE, THEIRS);
    view.seats[1] = { ...view.seats[1], playmat_wash: 80 };
    const r = mountPanel(view, false);
    expect(washOf(r.container.querySelector<HTMLElement>(".playmat"))).toBe("80%");
    const plain = mountPanel(gameView(MINE, THEIRS), true);
    expect(washOf(plain.container.querySelector<HTMLElement>(".playmat"))).toBe("");
    expect(panelSource()).toMatch(/var\(--bg\) var\(--playmat-wash, 58%\)/);
  });
});
