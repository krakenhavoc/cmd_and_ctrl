// @vitest-environment jsdom
//
// ADR 0209 §4 (#4397), the rendered half of signatures: a named card's
// scene replaces the generic cue for its batch once it has loaded, the
// cards it took still show their own (lighter) cues, it loads the first
// time the card appears, it is capped (2.5 s, Short 0.8 s), Off and reduced
// motion keep the generic still cues, and a skip clears it. The rules are
// pinned in fx/signatures.test.ts, the scenes in fx/scenes/kit.test.ts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";

vi.mock("./sounds", () => ({ play: vi.fn(), preloadSound: vi.fn() }));
vi.mock("./animations", async (importOriginal) => {
  const real = await importOriginal<typeof import("./animations")>();
  return { ...real, crumble: vi.fn(() => Promise.resolve()) };
});
vi.mock("./motion", async (importOriginal) => {
  const real = await importOriginal<typeof import("./motion")>();
  return {
    ...real,
    run: vi.fn(() => ({ cancel() {}, done: Promise.resolve(true) })),
  };
});

import TableEffects from "./components/board/TableEffects.svelte";
import { play as playSound } from "./sounds";
import type { CardView, GameView, LogEvent } from "./protocol";
import { resetSettings, updateSettings } from "./settings";
import { dockTestView } from "./test/dockView";
import { cleanup, render } from "./test/render.svelte";
import { SignatureLoader, type SceneID } from "./fx/signatures";
import type { SceneContext, SceneModule } from "./fx/scenes/kit";

// ---- the table ----------------------------------------------------------

const card = (id: string, name: string, extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: id,
    name,
    owner: "opp",
    controller: "opp",
    known_by_you: true,
    type_line: "Creature — Ogre",
    scryfall_id: `sf-${id}`,
    ...extra,
  }) as unknown as CardView;

const IDS = ["a", "b", "c", "d"];
const creatures = IDS.map((id) => card(id, `Ogre ${id}`));
const ACT = card("act", "Blasphemous Act", { type_line: "Sorcery" });
const TRIUMPH = card("bt", "Bitter Triumph", { type_line: "Instant" });
const BOLAS = card("bolas", "Nicol Bolas, God-Pharaoh", {
  type_line: "Legendary Planeswalker — Bolas",
});

interface Zones {
  battlefield?: CardView[];
  stack?: CardView[];
  graveyard?: CardView[];
  hand?: CardView[];
}

function viewWith(log: LogEvent[], z: Zones): GameView {
  const v = dockTestView({ log } as Partial<GameView>);
  const zone = (kind: string, cards: CardView[]) =>
    ({ kind, count: cards.length, cards }) as unknown as GameView["battlefield"];
  v.battlefield = zone("battlefield", z.battlefield ?? []);
  v.stack = zone("stack", z.stack ?? []);
  v.seats[1].graveyard = zone("graveyard", z.graveyard ?? []);
  v.seats[0].hand = zone("hand", z.hand ?? []);
  return v;
}

let seq = 100;
const entry = (kind: LogEvent["kind"], extra: Partial<LogEvent> = {}): LogEvent => ({
  seq: ++seq,
  kind,
  turn: 3,
  seat: 1,
  text: "",
  ...extra,
});
const died = (id: string, source: string, batch = 9): LogEvent =>
  entry("zone", {
    card_id: id,
    old_zone: "battlefield",
    new_zone: "graveyard",
    cause: "sba",
    batch,
    source,
  });

const boardSize = { w: 1200, h: 800 };

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

function makeBoard(ids: string[], stack: string[] = []): HTMLElement {
  const board = document.createElement("div");
  board.getBoundingClientRect = () => new DOMRect(0, 0, boardSize.w, boardSize.h);
  ids.forEach((id, i) => {
    const el = rectEl("data-instance-id", id, {
      left: 100 + i * 100,
      top: 300,
      width: 80,
      height: 112,
    });
    el.dataset.tapped = "false";
    el.appendChild(Object.assign(document.createElement("img"), { src: `/${id}.jpg` }));
    board.appendChild(el);
  });
  for (const id of stack) {
    const el = rectEl("data-stack-item-id", id, { left: 40, top: 200, width: 150, height: 210 });
    el.appendChild(Object.assign(document.createElement("img"), { src: `/${id}-stack.jpg` }));
    board.appendChild(el);
  }
  document.body.appendChild(board);
  return board;
}

function advance(ms: number) {
  flushSync();
  vi.advanceTimersByTime(ms);
  flushSync();
}

const layer = () => document.querySelector<HTMLElement>("[data-table-effects]")!;
const effects = (kind: string) => [
  ...layer().querySelectorAll<HTMLElement>(`[data-effect="${kind}"]`),
];
const signature = () => layer().querySelector<HTMLElement>('[data-effect="signature"]');
const badges = () => [...layer().querySelectorAll<HTMLElement>("[data-effect-badge]")];
const sounds = () => vi.mocked(playSound).mock.calls.map((c) => c[0]);

// A loader whose scenes resolve at once, recording what they are asked to draw.
const played: { scene: SceneID; ctx: SceneContext }[] = [];
function loader(ready: SceneID[] = ["blasphemous_act", "bitter_triumph", "nicol_bolas"]) {
  const load = vi.fn(
    (scene: SceneID): Promise<SceneModule> =>
      Promise.resolve({ play: (ctx: SceneContext) => void played.push({ scene, ctx }) }),
  );
  const l = new SignatureLoader({ load });
  for (const s of ready) l.want(s);
  return { l, load };
}

async function settle() {
  await vi.advanceTimersByTimeAsync(0);
  flushSync();
}

// Mount on a primed frame showing `before`, then deliver `log` with `after`.
async function deliver(
  l: SignatureLoader,
  before: Zones,
  log: LogEvent[],
  after: Zones,
  board = makeBoard(IDS),
) {
  const r = render(
    TableEffects as never,
    {
      view: viewWith([], before),
      boardEl: board,
      signatures: l,
    } as never,
  );
  await settle();
  advance(50);
  r.setProps({ view: viewWith(log, after) } as never);
  flushSync();
  await settle();
  return r;
}

beforeEach(() => {
  vi.useFakeTimers({
    toFake: [
      "Date",
      "setTimeout",
      "clearTimeout",
      "setInterval",
      "clearInterval",
      "requestAnimationFrame",
      "cancelAnimationFrame",
    ],
  });
  vi.setSystemTime(1_000_000);
  seq = 100;
  played.length = 0;
  resetSettings();
  updateSettings("animations", "enabled", true);
  updateSettings("animations", "effects", "full");
  updateSettings("animations", "speed", 1);
  updateSettings("accessibility", "reduceMotion", false);
  vi.mocked(playSound).mockClear();
});

afterEach(() => {
  cleanup();
  document.body.innerHTML = "";
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

const wipe = () => [
  entry("resolve", { card_id: "act", batch: 9 }),
  died("a", "act"),
  died("b", "act"),
  died("c", "act"),
  entry("zone", { card_id: "act", old_zone: "stack", new_zone: "graveyard" }),
];

describe("a signature", () => {
  it("replaces the board wipe's sweep and sound; each card still shows its own cue", async () => {
    const { l } = loader();
    await deliver(l, { battlefield: creatures, stack: [ACT] }, wipe(), {
      battlefield: [creatures[3]],
      graveyard: [ACT],
    });
    expect(signature()?.dataset.signature).toBe("blasphemous_act");
    expect(effects("board_wipe")).toHaveLength(0);
    const dies = effects("dies");
    expect(dies).toHaveLength(3);
    for (const d of dies) expect(d.hasAttribute("data-effect-light")).toBe(true);
    expect(sounds()).toContain("sig_blasphemous_act");
    expect(sounds()).not.toContain("board_wipe");
    // The scene was handed the cards it took, where they stood.
    expect(played).toHaveLength(1);
    expect(played[0].ctx.cards.map((c) => c.id)).toEqual(["a", "b", "c"]);
    expect(played[0].ctx.cards[0]).toMatchObject({ x: 140, y: 356, src: "/a.jpg" });
    expect(played[0].ctx.short).toBe(false);
  });

  it("falls back to the generic cues while its scene has not loaded", async () => {
    // An import still on its way (a slow network): nothing waits on it.
    const load = vi.fn(() => new Promise<SceneModule>(() => {}));
    const l = new SignatureLoader({ load });
    await deliver(l, { battlefield: creatures, stack: [ACT] }, wipe(), {
      battlefield: [creatures[3]],
      graveyard: [ACT],
    });
    expect(signature()).toBeNull();
    expect(effects("board_wipe")).toHaveLength(1);
    expect(sounds()).toContain("board_wipe");
    expect(sounds()).not.toContain("sig_blasphemous_act");
    // Its card was on the table, so the scene began to load.
    expect(load).toHaveBeenCalledWith("blasphemous_act");
    expect(l.status("blasphemous_act")).toBe("loading");
  });

  it("loads its scene the first time the card appears in a game, and not before", async () => {
    const { l, load } = loader([]);
    const board = makeBoard(IDS);
    const r = render(
      TableEffects as never,
      {
        view: viewWith([], { battlefield: creatures }),
        boardEl: board,
        signatures: l,
      } as never,
    );
    flushSync();
    expect(load).not.toHaveBeenCalled();
    // A card in this player's hand is a card that appeared.
    r.setProps({
      view: viewWith([entry("draw")], { battlefield: creatures, hand: [ACT] }),
    } as never);
    flushSync();
    expect(load).toHaveBeenCalledWith("blasphemous_act");
    expect(load).toHaveBeenCalledTimes(1);
    r.setProps({
      view: viewWith([entry("draw"), entry("draw")], { battlefield: creatures, hand: [ACT] }),
    } as never);
    flushSync();
    expect(load).toHaveBeenCalledTimes(1);
  });

  it("is gone by its 2.5 s cap", async () => {
    const { l } = loader();
    await deliver(l, { battlefield: creatures, stack: [ACT] }, wipe(), {
      battlefield: [creatures[3]],
      graveyard: [ACT],
    });
    expect(signature()).toBeTruthy();
    advance(2300);
    expect(signature()).toBeTruthy();
    advance(2500 + 100 - 2300);
    expect(signature()).toBeNull();
  });

  it("Short plays its first beat, gone within 0.8 s", async () => {
    updateSettings("animations", "effects", "short");
    const { l } = loader();
    await deliver(l, { battlefield: creatures, stack: [ACT] }, wipe(), {
      battlefield: [creatures[3]],
      graveyard: [ACT],
    });
    expect(signature()?.dataset.effectMotion).toBe("short");
    expect(played[0].ctx.short).toBe(true);
    // The scene's own first beat, 800 ms, is 800 ms of real time.
    expect(played[0].ctx.t(800)).toBeCloseTo(800);
    advance(800 + 100);
    expect(signature()).toBeNull();
  });

  it.each([
    ["Off", () => updateSettings("animations", "effects", "off")],
    ["reduced motion", () => updateSettings("accessibility", "reduceMotion", true)],
    ["animations off", () => updateSettings("animations", "enabled", false)],
  ])("%s keeps the generic still cues, with the signature's sound", async (_name, set) => {
    set();
    const { l } = loader();
    await deliver(l, { battlefield: creatures, stack: [ACT] }, wipe(), {
      battlefield: [creatures[3]],
      graveyard: [ACT],
    });
    expect(signature()).toBeNull();
    expect(played).toHaveLength(0);
    // The wipe's still form: one badge and the static tint.
    expect(badges()).toHaveLength(1);
    expect(sounds()).toContain("sig_blasphemous_act");
    expect(sounds()).not.toContain("board_wipe");
  });

  it("the OS's reduced-motion signal keeps the generic still cues too", async () => {
    vi.stubGlobal(
      "matchMedia",
      vi.fn(() => ({ matches: true, addEventListener() {}, removeEventListener() {} })),
    );
    const { l } = loader();
    await deliver(l, { battlefield: creatures, stack: [ACT] }, wipe(), {
      battlefield: [creatures[3]],
      graveyard: [ACT],
    });
    expect(signature()).toBeNull();
    expect(badges()).toHaveLength(1);
  });

  it("a click or Escape skips it", async () => {
    const { l } = loader();
    await deliver(l, { battlefield: creatures, stack: [ACT] }, wipe(), {
      battlefield: [creatures[3]],
      graveyard: [ACT],
    });
    expect(signature()).toBeTruthy();
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    flushSync();
    expect(signature()).toBeNull();
  });

  it("Bitter Triumph strikes the card it destroys, which then goes without the slash", async () => {
    const { l } = loader();
    await deliver(
      l,
      { battlefield: creatures, stack: [TRIUMPH] },
      [
        entry("resolve", { card_id: "bt", batch: 11 }),
        entry("zone", {
          card_id: "b",
          old_zone: "battlefield",
          new_zone: "graveyard",
          cause: "destroy",
          batch: 11,
          source: "bt",
        }),
      ],
      { battlefield: creatures.filter((c) => c.instance_id !== "b"), graveyard: [TRIUMPH] },
    );
    expect(signature()?.dataset.signature).toBe("bitter_triumph");
    expect(played[0].ctx.cards.map((c) => c.id)).toEqual(["b"]);
    const [destroy] = effects("destroy");
    expect(destroy.hasAttribute("data-effect-light")).toBe(true);
    expect(destroy.querySelector(".slash")).toBeNull();
    // One card, one sound: the signature's.
    expect(sounds()).toEqual(["sig_bitter_triumph"]);
  });

  it("a card with no signature plays the generic cues as before", async () => {
    const { l } = loader();
    const bolt = card("bolt", "Murder", { type_line: "Instant" });
    await deliver(
      l,
      { battlefield: creatures, stack: [bolt] },
      [
        entry("resolve", { card_id: "bolt", batch: 12 }),
        entry("zone", {
          card_id: "b",
          old_zone: "battlefield",
          new_zone: "graveyard",
          cause: "destroy",
          batch: 12,
          source: "bolt",
        }),
      ],
      { battlefield: creatures.filter((c) => c.instance_id !== "b"), graveyard: [bolt] },
    );
    expect(signature()).toBeNull();
    const [destroy] = effects("destroy");
    expect(destroy.hasAttribute("data-effect-light")).toBe(false);
    expect(sounds()).toEqual(["destroy"]);
  });
});

describe("Nicol Bolas", () => {
  it("lands on the stack when cast, and not again when he resolves", async () => {
    const { l } = loader();
    const board = makeBoard(IDS, ["bolas"]);
    const r = await deliver(
      l,
      { battlefield: creatures, hand: [BOLAS] },
      [entry("cast", { card_id: "bolas", batch: 20 })],
      { battlefield: creatures, stack: [BOLAS] },
      board,
    );
    expect(signature()?.dataset.signature).toBe("nicol_bolas");
    expect(played[0].ctx.landing).toMatchObject({
      id: "bolas",
      x: 115,
      y: 305,
      src: "/bolas-stack.jpg",
    });
    expect(sounds()).toContain("sig_nicol_bolas");
    advance(3000);
    expect(signature()).toBeNull();

    // He resolves: the generic arrival ring, no second swoop.
    const tile = rectEl("data-instance-id", "bolas", {
      left: 600,
      top: 500,
      width: 80,
      height: 112,
    });
    board.appendChild(tile);
    vi.mocked(playSound).mockClear();
    r.setProps({
      view: viewWith(
        [
          entry("cast", { card_id: "bolas", batch: 20 }),
          entry("resolve", { card_id: "bolas", batch: 21 }),
          entry("zone", {
            card_id: "bolas",
            old_zone: "stack",
            new_zone: "battlefield",
            batch: 21,
          }),
        ],
        { battlefield: [...creatures, BOLAS] },
      ),
    } as never);
    flushSync();
    await settle();
    expect(signature()).toBeNull();
    expect(played).toHaveLength(1);
    expect(effects("pw_enter")).toHaveLength(1);
    expect(sounds()).toContain("pw_enter");
  });

  it("entering uncast, he swoops and lands on his tile in place of the ring", async () => {
    const { l } = loader();
    const board = makeBoard([...IDS, "bolas"]);
    await deliver(
      l,
      { battlefield: creatures, graveyard: [BOLAS] },
      [
        entry("zone", {
          card_id: "bolas",
          old_zone: "graveyard",
          new_zone: "battlefield",
          batch: 30,
        }),
      ],
      { battlefield: [...creatures, BOLAS] },
      board,
    );
    expect(signature()?.dataset.signature).toBe("nicol_bolas");
    expect(effects("pw_enter")).toHaveLength(0);
    expect(played[0].ctx.landing).toMatchObject({ id: "bolas", x: 540, y: 356, src: "/bolas.jpg" });
    expect(sounds()).toContain("sig_nicol_bolas");
    expect(sounds()).not.toContain("pw_enter");
  });
});
