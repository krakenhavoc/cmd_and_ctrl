// @vitest-environment jsdom
//
// #4449, the rendered half: a player leaving the game plays the
// player_eliminated sound once and shatters their avatar (EliminationFx),
// and a tombstone with their name and commander stays on their panel
// (Tombstone) whoever they are and however many have left. The rules are
// pinned in elimination.test.ts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";

vi.mock("./sounds", () => ({ play: vi.fn() }));

import Board from "./components/board/Board.svelte";
import EliminationFx from "./components/board/EliminationFx.svelte";
import Tombstone from "./components/board/Tombstone.svelte";
import { play as playSound } from "./sounds";
import { L } from "./labels";
import { resetCommanderMemory } from "./elimination";
import type { CardView, GameView, LogEvent, PlayerView } from "./protocol";
import { resetSettings, updateSettings } from "./settings";
import { cleanup, render } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

// ---- the table ------------------------------------------------------------

const IDS = ["me", "bob", "cat", "dee"];
const NAMES = ["Me", "Bob", "Cat", "Dee"];

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const commanderOf = (id: string, name: string): CardView =>
  ({
    instance_id: `${id}-cmdr`,
    name,
    owner: id,
    controller: id,
    is_commander: true,
    type_line: "Legendary Creature",
  }) as unknown as CardView;

const COMMANDERS = ["Atraxa", "Kenrith", "Edea", "Aang"];

function seatView(n: number, out: boolean, commanderShown: boolean): PlayerView {
  const id = IDS[n];
  return {
    id,
    name: NAMES[n],
    seat: n,
    life: out ? 0 : 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    // An eliminated player's objects have left the game with them.
    command: zone("command", id, commanderShown ? [commanderOf(id, COMMANDERS[n])] : []),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
    ...(out ? { eliminated: true } : {}),
  } as unknown as PlayerView;
}

const eliminatedE = (seq: number, seat: number, cause = "life"): LogEvent =>
  ({ seq, kind: "eliminated", seat, cause, turn: 4, text: "" }) as LogEvent;

// A four-player game in which `gone` have left, in the order they left.
function gameView(gone: number[] = [], extraLog: LogEvent[] = [], active = 0): GameView {
  return {
    id: "g1",
    state: "active",
    seats: IDS.map((_, n) => seatView(n, gone.includes(n), !gone.includes(n))),
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      number: 4,
      active_seat: active,
      priority_holder: 0,
      phase: "main1",
      step: "precombat_main",
    },
    mulligans_open: false,
    log: [...gone.map((s, i) => eliminatedE(10 + i, s)), ...extraLog],
  } as unknown as GameView;
}

const soundCalls = () => vi.mocked(playSound).mock.calls.filter(([n]) => n === "player_eliminated");

function rectEl(
  attr: string,
  value: string,
  r: { left: number; top: number; width: number; height: number },
): HTMLElement {
  const el = document.createElement("div");
  el.setAttribute(attr, value);
  el.getBoundingClientRect = () => new DOMRect(r.left, r.top, r.width, r.height);
  return el;
}

// A board with each seat's avatar at its own spot.
function makeBoard(): HTMLElement {
  const board = document.createElement("div");
  board.getBoundingClientRect = () => new DOMRect(0, 0, 1200, 800);
  IDS.forEach((id, n) => {
    board.appendChild(
      rectEl("data-seat-id", id, {
        left: 100 + n * 250,
        top: n === 0 ? 700 : 40,
        width: 60,
        height: 60,
      }),
    );
  });
  document.body.appendChild(board);
  return board;
}

function advance(ms: number) {
  vi.advanceTimersByTime(ms);
  flushSync();
}

const bursts = (root: ParentNode = document) => [
  ...root.querySelectorAll<HTMLElement>("[data-burst]"),
];

function stubOSReduceMotion(on: boolean) {
  vi.stubGlobal("matchMedia", (query: string) => ({
    matches: on && query.includes("prefers-reduced-motion"),
    media: query,
    addEventListener: () => {},
    removeEventListener: () => {},
    addListener: () => {},
    removeListener: () => {},
    onchange: null,
    dispatchEvent: () => false,
  }));
}

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  vi.useFakeTimers({
    toFake: ["Date", "setTimeout", "clearTimeout", "setInterval", "clearInterval"],
  });
  vi.setSystemTime(1_000_000);
  resetSettings();
  updateSettings("animations", "enabled", true);
  updateSettings("animations", "effects", "full");
  updateSettings("animations", "speed", 1);
  updateSettings("accessibility", "reduceMotion", false);
  vi.mocked(playSound).mockClear();
  resetCommanderMemory();
  stubOSReduceMotion(false);
});

afterEach(() => {
  cleanup();
  document.body.innerHTML = "";
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

// ---- the burst and the sound ---------------------------------------------

describe("the elimination layer", () => {
  function mountFx(view: GameView) {
    const board = makeBoard();
    const r = render(EliminationFx as never, { view, boardEl: board } as never);
    return { r, board };
  }

  it("plays player_eliminated once, and shatters the avatar, when a player leaves", () => {
    const { r } = mountFx(gameView());
    r.setProps({ view: gameView([1]) } as never);
    flushSync();
    expect(soundCalls()).toHaveLength(1);
    const [burst] = bursts();
    expect(burst.dataset.burst).toBe("bob");
    // Drawn on Bob's avatar: its centre, in board pixels, and its width.
    expect(burst.style.left).toBe("380px");
    expect(burst.style.top).toBe("70px");
    expect(burst.style.getPropertyValue("--d")).toBe("60px");
    expect(burst.querySelectorAll(".shard")).toHaveLength(8);
    expect(burst.querySelectorAll(".spark")).toHaveLength(28);
    expect(burst.querySelectorAll(".ring")).toHaveLength(1);

    // The same frame again (a re-render with the same log) makes no second cue.
    r.setProps({ view: gameView([1]) } as never);
    advance(3000);
    expect(soundCalls()).toHaveLength(1);
    // The burst is cleared away once it has played.
    expect(bursts()).toHaveLength(0);
  });

  it("names the take once per eliminated player, each batch its own", () => {
    const { r } = mountFx(gameView());
    r.setProps({ view: gameView([1]) } as never);
    r.setProps({ view: gameView([1, 2]) } as never);
    r.setProps({ view: gameView([1, 2, 0]) } as never);
    advance(3000);
    // Three players leave over a four-player game, the viewer last.
    expect(soundCalls()).toHaveLength(3);
    const batches = soundCalls().map(([, opts]) => (opts as { batch?: string }).batch);
    expect(batches).toEqual(["g1:out:bob", "g1:out:cat", "g1:out:me"]);
  });

  it("shatters the viewer's own avatar as it does an opponent's", () => {
    const { r } = mountFx(gameView());
    r.setProps({ view: gameView([0]) } as never);
    flushSync();
    expect(soundCalls()).toHaveLength(1);
    expect(bursts().map((b) => b.dataset.burst)).toEqual(["me"]);
  });

  it("hears two players leaving in one frame apart", () => {
    const { r } = mountFx(gameView());
    r.setProps({ view: gameView([1, 3]) } as never);
    flushSync();
    expect(soundCalls()).toHaveLength(1);
    advance(180);
    expect(soundCalls()).toHaveLength(2);
    expect(bursts().map((b) => b.dataset.burst)).toEqual(["bob", "dee"]);
  });

  it("plays nothing for players already gone when the page opens (a reconnect)", () => {
    mountFx(gameView([1, 2]));
    advance(3000);
    expect(soundCalls()).toHaveLength(0);
    expect(bursts()).toHaveLength(0);
  });

  it("plays nothing when a prime key changes (a replay jump) until the next frame", () => {
    const { r } = mountFx(gameView());
    r.setProps({ view: gameView([1]), beatsPrimeKey: "replay" } as never);
    flushSync();
    expect(soundCalls()).toHaveLength(0);
    r.setProps({ view: gameView([1, 2]), beatsPrimeKey: "replay" } as never);
    flushSync();
    expect(soundCalls()).toHaveLength(1);
  });

  it("draws no burst under reduced motion, and still plays the sound once", () => {
    updateSettings("accessibility", "reduceMotion", true);
    const { r } = mountFx(gameView());
    r.setProps({ view: gameView([2]) } as never);
    flushSync();
    expect(soundCalls()).toHaveLength(1);
    expect(document.querySelectorAll(".shard, .spark, .ring")).toHaveLength(0);
    expect(bursts()).toHaveLength(0);
  });

  it("draws no particles when the OS asks for reduced motion", () => {
    stubOSReduceMotion(true);
    const { r } = mountFx(gameView());
    r.setProps({ view: gameView([2]) } as never);
    flushSync();
    expect(soundCalls()).toHaveLength(1);
    expect(document.querySelectorAll(".shard, .spark")).toHaveLength(0);
  });

  it("draws no particles with animations off or the effects setting Off", () => {
    updateSettings("animations", "effects", "off");
    const a = mountFx(gameView());
    a.r.setProps({ view: gameView([1]) } as never);
    flushSync();
    expect(document.querySelectorAll(".shard, .spark")).toHaveLength(0);
    expect(soundCalls()).toHaveLength(1);
  });

  it("cuts the burst short in the Short setting", () => {
    updateSettings("animations", "effects", "short");
    const { r } = mountFx(gameView());
    r.setProps({ view: gameView([1]) } as never);
    flushSync();
    expect(bursts()[0].style.getPropertyValue("--ms")).toBe("540ms");
  });
});

// ---- the tombstone --------------------------------------------------------

describe("the tombstone", () => {
  const mount = (view: GameView, n: number, compact = false) =>
    render(Tombstone as never, { seat: view.seats[n], view, compact } as never);
  const stone = () => document.querySelector<HTMLElement>("[data-tombstone]");

  it("is not there for a player still in the game", () => {
    mount(gameView(), 1);
    expect(stone()).toBeNull();
  });

  it("shows R. I. P., the player's name and their commander, from the label", () => {
    const first = gameView();
    const r = mount(first, 1);
    flushSync();
    // The commander was seen while Bob was in; the frame he left in shows none.
    const after = gameView([1]);
    r.setProps({ seat: after.seats[1], view: after } as never);
    flushSync();
    const el = stone()!;
    expect(el.dataset.tombstone).toBe("bob");
    expect(el.getAttribute("role")).toBe("group");
    expect(el.getAttribute("aria-label")).toBe(L.tombstoneOf("Bob"));
    expect(el.textContent).toContain(L.tombstoneRip);
    expect(el.textContent).toContain("Bob");
    expect(el.textContent).toContain("Kenrith");
    expect(el.querySelector(".rip")?.textContent).toBe("R. I. P.");
  });

  it("prefers a Discord display name", () => {
    const v = gameView([1]);
    (v.seats[1] as { display_name?: string }).display_name = "Robert";
    mount(v, 1);
    expect(stone()?.querySelector(".who")?.textContent).toBe("Robert");
  });

  it("omits the commander line when no commander was ever seen", () => {
    const v = gameView([2]);
    mount(v, 2);
    expect(stone()!.querySelector(".commander")).toBeNull();
    expect(stone()!.textContent).toContain("Cat");
  });

  it("drops in after the burst when the player leaves in front of it", () => {
    updateSettings("animations", "speed", 1.5);
    const r = mount(gameView(), 1);
    expect(stone()).toBeNull();
    const after = gameView([1]);
    r.setProps({ seat: after.seats[1], view: after } as never);
    flushSync();
    const el = stone()!;
    expect(el.classList.contains("arrives")).toBe(true);
    expect(el.classList.contains("still")).toBe(false);
    expect(el.dataset.motion).toBe("full");
    // Waits for the burst, scaled by the animation speed.
    expect(el.style.getPropertyValue("--tomb-delay")).toBe("780ms");
  });

  it("only fades in under reduced motion, with no wait for a burst", () => {
    updateSettings("accessibility", "reduceMotion", true);
    const r = mount(gameView(), 1);
    const after = gameView([1]);
    r.setProps({ seat: after.seats[1], view: after } as never);
    flushSync();
    const el = stone()!;
    expect(el.classList.contains("arrives")).toBe(true);
    expect(el.classList.contains("still")).toBe(true);
    expect(el.dataset.motion).toBe("still");
    expect(el.style.getPropertyValue("--tomb-delay")).toBe("0ms");
  });

  it("only fades in when the OS asks for reduced motion", () => {
    stubOSReduceMotion(true);
    const r = mount(gameView(), 2);
    const after = gameView([2]);
    r.setProps({ seat: after.seats[2], view: after } as never);
    flushSync();
    expect(stone()!.dataset.motion).toBe("still");
  });

  it("is simply there, with no arrival, on a page opened after the player left", () => {
    mount(gameView([3]), 3);
    const el = stone()!;
    expect(el.classList.contains("arrives")).toBe(false);
    expect(el.dataset.motion).toBeUndefined();
  });

  it("goes away if the elimination is undone", () => {
    const v = gameView([1]);
    const r = mount(v, 1);
    expect(stone()).not.toBeNull();
    const back = gameView();
    r.setProps({ seat: back.seats[1], view: back } as never);
    flushSync();
    expect(stone()).toBeNull();
  });

  it("is a small stone in a collapsed seat's card", () => {
    mount(gameView([1]), 1, true);
    expect(stone()!.classList.contains("compact")).toBe(true);
  });
});

// ---- on the board ----------------------------------------------------------

describe("on the board", () => {
  function mountBoard(view: GameView) {
    return render(
      Board as never,
      {
        view,
        viewerID: "me",
        isAdmin: false,
        sendAction: () => {},
        combatMode: "idle",
        selectedCombatCardID: null,
        onSelectCombatCard: () => {},
        onDeclareAttack: () => {},
        onDeclareBlock: () => {},
      } as never,
    );
  }
  const stones = () =>
    [...document.querySelectorAll<HTMLElement>("[data-tombstone]")].map((e) => e.dataset.tombstone);
  const stoneOf = (id: string) => document.querySelector<HTMLElement>(`[data-tombstone="${id}"]`);

  it("keeps a tombstone on each eliminated player's panel, up to three in a four-player game", () => {
    updateSettings("display", "expandStyle", "overlay");
    const r = mountBoard(gameView());
    expect(stones()).toEqual([]);

    r.setProps({ view: gameView([2]) } as never);
    flushSync();
    expect(stones()).toEqual(["cat"]);
    expect(soundCalls()).toHaveLength(1);

    r.setProps({ view: gameView([2, 1]) } as never);
    flushSync();
    expect(stones().sort()).toEqual(["bob", "cat"]);
    expect(soundCalls()).toHaveLength(2);

    // The third is the viewer: their own panel gets the stone too.
    r.setProps({ view: gameView([2, 1, 0]) } as never);
    flushSync();
    expect(stones().sort()).toEqual(["bob", "cat", "me"]);
    expect(soundCalls()).toHaveLength(3);
    advance(5000);
    expect(soundCalls()).toHaveLength(3);

    // On each player's own panel, with their name and the commander they had.
    for (const [id, name, cmdr] of [
      ["cat", "Cat", "Edea"],
      ["bob", "Bob", "Kenrith"],
      ["me", "Me", "Atraxa"],
    ]) {
      const el = stoneOf(id)!;
      expect(el.closest(".panel")).not.toBeNull();
      expect(el.textContent).toContain(L.tombstoneRip);
      expect(el.textContent).toContain(name);
      expect(el.textContent).toContain(cmdr);
    }
    // Dee is still playing.
    expect(stoneOf("dee")).toBeNull();
  });

  it("shows the stone on a collapsed seat's card in the reflow layout", () => {
    updateSettings("display", "expandStyle", "reflow");
    const r = mountBoard(gameView([], [], 1));
    r.setProps({ view: gameView([2], [], 1) } as never);
    flushSync();
    const el = stoneOf("cat")!;
    expect(el).not.toBeNull();
    expect(el.closest(".avatar-card")).not.toBeNull();
    expect(el.textContent).toContain(L.tombstoneRip);
    // The card was mounted before Cat left, so it knew her commander.
    expect(el.textContent).toContain("Edea");
    expect(soundCalls()).toHaveLength(1);
  });
});

// ---- where the stone sits (#4552) ------------------------------------------

describe("where the stone sits", () => {
  function mountBoard(view: GameView) {
    return render(
      Board as never,
      {
        view,
        viewerID: "me",
        isAdmin: false,
        sendAction: () => {},
        combatMode: "idle",
        selectedCombatCardID: null,
        onSelectCombatCard: () => {},
        onDeclareAttack: () => {},
        onDeclareBlock: () => {},
      } as never,
    );
  }
  const stoneOf = (id: string) => document.querySelector<HTMLElement>(`[data-tombstone="${id}"]`);

  it("is the same on every full panel: the panel's own tomb layer, one deep, full size", () => {
    updateSettings("display", "expandStyle", "overlay");
    const r = mountBoard(gameView());
    r.setProps({ view: gameView([2, 1, 0]) } as never);
    flushSync();
    // Cat and Bob are top-row-style seats, Me is the viewer's own: all
    // three get the stone by one rule, as a direct child of their panel.
    for (const id of ["cat", "bob", "me"]) {
      const el = stoneOf(id)!;
      const panel = el.parentElement!;
      expect(panel.classList.contains("panel"), id).toBe(true);
      expect(panel.getAttribute("aria-label"), id).toBeTruthy();
      expect(el.classList.contains("tomb"), id).toBe(true);
      expect(el.classList.contains("compact"), id).toBe(false);
      // Never inside the identity or the rows it must stay clear of.
      expect(el.closest(".identity"), id).toBeNull();
      expect(el.closest(".grid-bottom, .grid-middle, .grid-creatures, .rail"), id).toBeNull();
    }
    // And the layer is one box over the panel, not offset per seat.
    expect(new Set(["cat", "bob", "me"].map((id) => stoneOf(id)!.className)).size).toBe(1);
  });

  it("is the small stone in the place of the read-out on a collapsed seat's card", () => {
    updateSettings("display", "expandStyle", "reflow");
    const r = mountBoard(gameView([], [], 1));
    r.setProps({ view: gameView([2, 0], [], 1) } as never);
    flushSync();
    const el = stoneOf("cat")!;
    const card = el.parentElement!;
    expect(card.classList.contains("avatar-card")).toBe(true);
    expect(el.classList.contains("tomb")).toBe(true);
    expect(el.classList.contains("compact")).toBe(true);
    // Under the card's identity, where the read-out would be, and in the
    // flow rather than over it.
    const identity = card.querySelector(".card-identity")!;
    expect(identity.compareDocumentPosition(el) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(card.querySelector(".readout")).toBeNull();
    expect(el.style.getPropertyValue("--tomb-top")).toBe("");
    // The viewer's own panel still has the full stone.
    expect(stoneOf("me")!.classList.contains("compact")).toBe(false);
  });

  describe("measured against the identity", () => {
    type Cb = () => void;
    let observers: Cb[] = [];
    let frames: Array<() => void> = [];

    beforeEach(() => {
      observers = [];
      frames = [];
      vi.stubGlobal(
        "ResizeObserver",
        class {
          constructor(cb: Cb) {
            observers.push(cb);
          }
          observe() {}
          unobserve() {}
          disconnect() {}
        },
      );
      vi.stubGlobal("requestAnimationFrame", (cb: () => void) => frames.push(cb));
      vi.stubGlobal("cancelAnimationFrame", () => {});
    });

    // A stone on a panel whose box and identity are at the given heights,
    // measured once the page has laid out.
    function onPanel(panel: [number, number], identity: [number, number]) {
      const view = gameView([2]);
      const r = render(Tombstone as never, { seat: view.seats[2], view } as never);
      flushSync();
      const host = r.container;
      host.getBoundingClientRect = () => new DOMRect(0, panel[0], 500, panel[1] - panel[0]);
      const id = document.createElement("div");
      id.className = "identity";
      id.getBoundingClientRect = () =>
        new DOMRect(200, identity[0], 100, identity[1] - identity[0]);
      host.appendChild(id);
      for (const run of observers) run();
      for (const f of frames.splice(0)) f();
      flushSync();
      const el = document.querySelector<HTMLElement>("[data-tombstone]")!;
      return {
        top: el.style.getPropertyValue("--tomb-top"),
        bottom: el.style.getPropertyValue("--tomb-bottom"),
        snug: el.classList.contains("snug"),
      };
    }

    it("sits under an identity at the top edge of the panel", () => {
      expect(onPanel([0, 300], [8, 130])).toEqual({ top: "140px", bottom: "10px", snug: false });
    });

    it("sits over an identity at the bottom edge of the panel", () => {
      expect(onPanel([0, 600], [480, 592])).toEqual({ top: "10px", bottom: "130px", snug: false });
    });

    it("is under the avatar on a short panel and on a tall one alike", () => {
      const short = onPanel([0, 330], [8, 130]);
      const tall = onPanel([0, 900], [8, 130]);
      expect(short.top).toBe(tall.top);
      expect(short.snug).toBe(false);
      expect(tall.snug).toBe(false);
    });

    it("steps down to the small stone when the room under the avatar is too short", () => {
      expect(onPanel([0, 217], [8, 135]).snug).toBe(true);
    });
  });
});
