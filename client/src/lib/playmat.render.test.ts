// @vitest-environment jsdom
//
// playmat.render.test.ts — ADR 0128. Two surfaces:
//
//   - the board: a seat's panel draws its owner's playmat behind the
//     zones, under the per-device all / mine / off setting, with no
//     broken-image box when the file will not load, and never a URL
//     that is not the server's own route;
//   - the Settings "Playmat" tab: the per-device choice for everyone,
//     and an account's preview, upload, link and Remove with the
//     server's message when it refuses.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { flushSync } from "svelte";

vi.mock("./sounds", () => ({ play: () => {} }));

const api = vi.hoisted(() => ({
  fetchMyPlaymat: vi.fn(),
  uploadMyPlaymat: vi.fn(),
  linkMyPlaymat: vi.fn(),
  removeMyPlaymat: vi.fn(),
  setMyPlaymatWash: vi.fn(),
}));
vi.mock("./api", async (orig) => ({ ...((await orig()) as object), ...api }));

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import PlaymatSettings from "./components/PlaymatSettings.svelte";
import type { GameView, PlayerView } from "./protocol";
import { resetSettings, updateSettings } from "./settings";
import { LobbyApiError, session, type Session } from "./session";
import { cleanup, click, render } from "./test/render.svelte";

const ME = "me";
const BOB = "bob";
const MINE = "/playmats/11111111-1111-4111-8111-111111111111";
const THEIRS = "/playmats/22222222-2222-4222-8222-222222222222";

beforeEach(() => {
  resetSettings();
  session.set(null);
  for (const f of Object.values(api)) f.mockReset();
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

describe("the Settings Playmat tab", () => {
  const person = (): Session =>
    ({
      token: "tok",
      expiresAt: "2999-01-01T00:00:00Z",
      principal: { role: "identified", user_id: "u1" },
    }) as unknown as Session;

  async function settle(): Promise<void> {
    for (let i = 0; i < 4; i++) await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }

  const radio = (c: HTMLElement, v: string) =>
    c.querySelector<HTMLInputElement>(`input[name="playmats-mode"][value="${v}"]`)!;
  const button = (c: HTMLElement, text: string) =>
    [...c.querySelectorAll<HTMLButtonElement>("button")].find(
      (b) => b.textContent?.trim() === text,
    );

  it("always offers the per-device choice, defaulting to everyone's, and saves it", async () => {
    const r = render(PlaymatSettings as never, {} as never);
    await settle();
    expect(radio(r.container, "all").checked).toBe(true);
    click(radio(r.container, "off"));
    const stored = JSON.parse(localStorage.getItem("cmdctrl.settings.v1") ?? "{}");
    expect(stored.display.playmats).toBe("off");
    expect(radio(r.container, "off").checked).toBe(true);
  });

  it("hides the account controls from a guest, and does not ask the server", async () => {
    const r = render(PlaymatSettings as never, {} as never);
    await settle();
    expect(api.fetchMyPlaymat).not.toHaveBeenCalled();
    expect(button(r.container, "Upload an image")).toBeUndefined();
    expect(r.container.querySelector('input[type="url"]')).toBeNull();
    expect(r.container.textContent).toContain("Sign in with Discord");
  });

  it("hides them when the server says 403 (no account to own a playmat)", async () => {
    session.set(person());
    api.fetchMyPlaymat.mockRejectedValue(new LobbyApiError(403, "not signed in as a person"));
    const r = render(PlaymatSettings as never, {} as never);
    await settle();
    expect(button(r.container, "Upload an image")).toBeUndefined();
    expect(radio(r.container, "all")).not.toBeNull();
  });

  it("hides them when the server has nowhere to store one", async () => {
    session.set(person());
    api.fetchMyPlaymat.mockResolvedValue({ enabled: false });
    const r = render(PlaymatSettings as never, {} as never);
    await settle();
    expect(button(r.container, "Upload an image")).toBeUndefined();
  });

  it("shows the preview and Remove for an account that has a playmat", async () => {
    session.set(person());
    api.fetchMyPlaymat.mockResolvedValue({ enabled: true, url: MINE, width: 100, height: 60 });
    const r = render(PlaymatSettings as never, {} as never);
    await settle();
    expect(r.container.querySelector<HTMLImageElement>(".preview img")!.getAttribute("src")).toBe(
      `${MINE}?token=tok`,
    );
    api.removeMyPlaymat.mockResolvedValue({ enabled: true });
    click(button(r.container, "Remove")!);
    await settle();
    expect(api.removeMyPlaymat).toHaveBeenCalledTimes(1);
    expect(r.container.querySelector(".preview")).toBeNull();
    expect(button(r.container, "Remove")).toBeUndefined();
    expect(r.container.textContent).toContain("You have no playmat");
  });

  it("shows no broken image when the preview will not load", async () => {
    session.set(person());
    api.fetchMyPlaymat.mockResolvedValue({ enabled: true, url: MINE });
    const r = render(PlaymatSettings as never, {} as never);
    await settle();
    r.container.querySelector(".preview img")!.dispatchEvent(new Event("error"));
    flushSync();
    expect(r.container.querySelector(".preview")).toBeNull();
  });

  it("uses a pasted link, sends it to the server once, and clears the field", async () => {
    session.set(person());
    api.fetchMyPlaymat.mockResolvedValue({ enabled: true });
    api.linkMyPlaymat.mockResolvedValue({ enabled: true, url: MINE, width: 8, height: 8 });
    const r = render(PlaymatSettings as never, {} as never);
    await settle();
    const use = button(r.container, "Use this image")!;
    expect(use.disabled).toBe(true); // nothing pasted yet
    const input = r.container.querySelector<HTMLInputElement>('input[type="url"]')!;
    input.value = " https://example.com/mat.jpg ";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    r.container
      .querySelector("form")!
      .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();
    expect(api.linkMyPlaymat).toHaveBeenCalledTimes(1);
    expect(api.linkMyPlaymat).toHaveBeenCalledWith("https://example.com/mat.jpg");
    expect(input.value).toBe("");
    // The preview is OUR url, not the pasted one.
    expect(r.container.querySelector(".preview img")!.getAttribute("src")).toContain("/playmats/");
    expect(r.container.querySelector(".preview")!.innerHTML).not.toContain("example.com");
  });

  it("shows the server's message when it refuses a link", async () => {
    session.set(person());
    api.fetchMyPlaymat.mockResolvedValue({ enabled: true });
    api.linkMyPlaymat.mockRejectedValue(
      new LobbyApiError(422, "could not fetch that link: only https links are accepted"),
    );
    const r = render(PlaymatSettings as never, {} as never);
    await settle();
    const input = r.container.querySelector<HTMLInputElement>('input[type="url"]')!;
    input.value = "http://example.com/a.png";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    r.container
      .querySelector("form")!
      .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();
    expect(r.container.querySelector('[role="alert"]')!.textContent).toContain("only https links");
    // The field keeps what was typed, so it can be fixed.
    expect(input.value).toBe("http://example.com/a.png");
  });

  it("uploads a chosen file", async () => {
    session.set(person());
    api.fetchMyPlaymat.mockResolvedValue({ enabled: true });
    api.uploadMyPlaymat.mockResolvedValue({ enabled: true, url: THEIRS, width: 4, height: 4 });
    const r = render(PlaymatSettings as never, {} as never);
    await settle();
    const file = new File([new Uint8Array(10)], "mat.png", { type: "image/png" });
    const input = r.container.querySelector<HTMLInputElement>('input[type="file"]')!;
    expect(input.getAttribute("accept")).toBe("image/png,image/jpeg,image/webp");
    Object.defineProperty(input, "files", { value: [file], configurable: true });
    input.dispatchEvent(new Event("change", { bubbles: true }));
    await settle();
    expect(api.uploadMyPlaymat).toHaveBeenCalledTimes(1);
    expect(api.uploadMyPlaymat).toHaveBeenCalledWith(file);
    expect(r.container.querySelector(".preview img")).not.toBeNull();
  });

  it("refuses an oversized file before sending it, and shows the server's message when it refuses one", async () => {
    session.set(person());
    api.fetchMyPlaymat.mockResolvedValue({ enabled: true });
    const r = render(PlaymatSettings as never, {} as never);
    await settle();
    const input = r.container.querySelector<HTMLInputElement>('input[type="file"]')!;

    const big = new File([new Uint8Array(1)], "big.png", { type: "image/png" });
    Object.defineProperty(big, "size", { value: 11 * 1024 * 1024 });
    Object.defineProperty(input, "files", { value: [big], configurable: true });
    input.dispatchEvent(new Event("change", { bubbles: true }));
    await settle();
    expect(api.uploadMyPlaymat).not.toHaveBeenCalled();
    expect(r.container.querySelector('[role="alert"]')!.textContent).toContain("10 MB");

    api.uploadMyPlaymat.mockRejectedValue(
      new LobbyApiError(415, "that file is not a PNG, JPEG or WebP image"),
    );
    const bad = new File([new Uint8Array(4)], "x.png", { type: "image/png" });
    Object.defineProperty(input, "files", { value: [bad], configurable: true });
    input.dispatchEvent(new Event("change", { bubbles: true }));
    await settle();
    expect(r.container.querySelector('[role="alert"]')!.textContent).toContain(
      "not a PNG, JPEG or WebP",
    );
  });
});

// ADR 0128 amendment: the owner sets how dark their playmat is, and
// every viewer draws it at that strength.
describe("the owner-set wash", () => {
  const person = (): Session =>
    ({
      token: "tok",
      expiresAt: "2999-01-01T00:00:00Z",
      principal: { role: "identified", user_id: "u1" },
    }) as unknown as Session;

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

  it("starts from the account's wash and saves once the slider rests", async () => {
    vi.useFakeTimers();
    try {
      session.set(person());
      api.fetchMyPlaymat.mockResolvedValue({ enabled: true, url: MINE, wash: 70 });
      api.setMyPlaymatWash.mockResolvedValue({ enabled: true, url: MINE, wash: 45 });
      const r = render(PlaymatSettings as never, {} as never);
      await vi.advanceTimersByTimeAsync(10);
      flushSync();
      const slider = r.container.querySelector<HTMLInputElement>(
        'input[aria-label="Playmat darkness"]',
      )!;
      expect(slider.value).toBe("70");
      expect(washOf(r.container.querySelector<HTMLElement>(".preview"))).toBe("70%");
      for (const v of ["60", "50", "45"]) {
        slider.value = v;
        slider.dispatchEvent(new Event("input", { bubbles: true }));
      }
      flushSync();
      expect(washOf(r.container.querySelector<HTMLElement>(".preview"))).toBe("45%");
      expect(api.setMyPlaymatWash).not.toHaveBeenCalled();
      await vi.advanceTimersByTimeAsync(450);
      expect(api.setMyPlaymatWash).toHaveBeenCalledTimes(1);
      expect(api.setMyPlaymatWash).toHaveBeenCalledWith(45);
    } finally {
      vi.useRealTimers();
    }
  });
});
