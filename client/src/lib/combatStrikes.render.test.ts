// @vitest-environment jsdom
//
// ADR 0134, the rendered half: a combat damage beat mounts an art-only
// copy for each creature that lunges, hides its live tile for the
// flight, shakes what it hit at contact, and draws a creature that died
// from the cache. The rules are pinned in combatStrikes.test.ts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));
vi.mock("./animations", async (importOriginal) => {
  const real = await importOriginal<typeof import("./animations")>();
  return {
    ...real,
    impactShake: vi.fn(() => Promise.resolve()),
    lunge: vi.fn(() => Promise.resolve()),
    crumble: vi.fn(() => Promise.resolve()),
    streak: vi.fn(() => Promise.resolve()),
  };
});

import CombatStrikes from "./components/board/CombatStrikes.svelte";
import CombatCuesCard from "./test/CombatCuesCard.svelte";
import { crumble, impactShake, lunge, streak } from "./animations";
import { CombatCues } from "./combatCues.svelte";
import type { CardView, GameView, LogEvent } from "./protocol";
import { resetSettings, settings, updateSettings } from "./settings";
import { dockTestView } from "./test/dockView";
import { cleanup, render } from "./test/render.svelte";

// ---- the table ----------------------------------------------------------

const creature = (id: string, controller: string, extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: id,
    name: id,
    owner: controller,
    controller,
    known_by_you: true,
    type_line: "Creature — Ogre",
    scryfall_id: `sf-${id}`,
    ...extra,
  }) as unknown as CardView;

function viewWith(log: LogEvent[], cards: CardView[], step = "combat_damage"): GameView {
  const v = dockTestView({ log } as Partial<GameView>);
  v.battlefield = { kind: "battlefield", count: cards.length, cards } as GameView["battlefield"];
  v.turn = { ...v.turn, step } as GameView["turn"];
  return v;
}

const stepE = (seq: number, step: string): LogEvent => ({
  seq,
  kind: "step",
  step,
  turn: 3,
  seat: 0,
  text: step,
});
const blockE = (seq: number, blocker: string, attacker: string): LogEvent => ({
  seq,
  kind: "block",
  turn: 3,
  seat: 1,
  card_id: blocker,
  target: attacker,
  text: "",
});
const hit = (seq: number, seat: number, src: string, to: string | number): LogEvent => ({
  seq,
  kind: "damage",
  turn: 3,
  seat,
  card_id: src,
  ...(typeof to === "number" ? { target_seat: to } : { target: to }),
  amount: 3,
  combat: true,
  text: "",
});
const died = (seq: number, id: string): LogEvent => ({
  seq,
  kind: "zone",
  turn: 3,
  seat: -1,
  card_id: id,
  old_zone: "battlefield",
  new_zone: "graveyard",
  text: "",
});

let boardSize = { w: 1200, h: 800 };

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

// A board with the attacker's tile (tapped), a blocker's tile and the
// defender's avatar.
function makeBoard(): { board: HTMLElement; tiles: Record<string, HTMLElement> } {
  const board = document.createElement("div");
  board.getBoundingClientRect = () => new DOMRect(0, 0, boardSize.w, boardSize.h);
  const ogre = rectEl("data-instance-id", "ogre", { left: 500, top: 560, width: 112, height: 80 });
  ogre.dataset.tapped = "true";
  ogre.appendChild(Object.assign(document.createElement("img"), { src: "/ogre.jpg" }));
  const bear = rectEl("data-instance-id", "bear", { left: 520, top: 200, width: 80, height: 112 });
  bear.dataset.tapped = "false";
  const avatar = rectEl("data-seat-id", "opp", { left: 560, top: 20, width: 60, height: 60 });
  board.append(ogre, bear, avatar);
  document.body.appendChild(board);
  return { board, tiles: { ogre, bear, avatar } };
}

let cues: CombatCues;

function advance(ms: number) {
  vi.advanceTimersByTime(ms);
  flushSync();
}

function copies(root: ParentNode = document): HTMLElement[] {
  return [...root.querySelectorAll<HTMLElement>("[data-strike-copy]")];
}

// Mount the layer on a primed frame, then deliver `next` as a new frame.
function play(
  before: LogEvent[],
  next: LogEvent[],
  cardsBefore: CardView[],
  cardsAfter: CardView[],
) {
  const { board, tiles } = makeBoard();
  const r = render(
    CombatStrikes as never,
    {
      view: viewWith(before, cardsBefore, "declare_blockers"),
      boardEl: board,
      cues,
    } as never,
  );
  cues.frame(before, get(settings));
  // Let the layer measure the tiles as they stand before the damage.
  advance(20);
  r.setProps({ view: viewWith(next, cardsAfter) } as never);
  cues.frame(next, get(settings));
  return { r, board, tiles };
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
  boardSize = { w: 1200, h: 800 };
  resetSettings();
  updateSettings("animations", "enabled", true);
  updateSettings("animations", "combat", true);
  updateSettings("animations", "speed", 1);
  updateSettings("accessibility", "reduceMotion", false);
  vi.mocked(impactShake).mockClear();
  vi.mocked(lunge).mockClear();
  vi.mocked(crumble).mockClear();
  vi.mocked(streak).mockClear();
  cues = new CombatCues();
});

afterEach(() => {
  cues.dispose();
  cleanup();
  document.body.innerHTML = "";
  vi.useRealTimers();
});

const ogre = creature("ogre", "me", { tapped: true });
const bear = creature("bear", "opp");

describe("the strike layer", () => {
  it("mounts a copy of an unblocked attacker and hides its tile until the copy is gone", () => {
    const before = [stepE(100, "declare_attackers"), stepE(110, "declare_blockers")];
    const next = [...before, stepE(120, "combat_damage"), hit(121, 0, "ogre", 1)];
    const { r, tiles } = play(before, next, [ogre], [ogre]);

    advance(0);
    const [copy] = copies(r.container);
    expect(copy?.dataset.strikeCopy).toBe("ogre");
    expect(copy.dataset.strikeMode).toBe("lunge");
    // Drawn where the tile is: the tile's centre, its size at rest and
    // its tap rotation.
    expect(copy.style.left).toBe("556px");
    expect(copy.style.top).toBe("600px");
    const face = copy.querySelector<HTMLElement>(".face")!;
    expect(face.style.width).toBe("80px");
    expect(face.style.height).toBe("112px");
    expect(face.style.getPropertyValue("--strike-rot")).toBe("90deg");
    expect(face.querySelector("img")?.getAttribute("src")).toBe("/ogre.jpg");
    expect(cues.isStriking("ogre")).toBe(true);
    // It lunges toward the avatar (up the board).
    const [, to] = vi.mocked(lunge).mock.calls[0];
    expect(to.y).toBeLessThan(0);

    // The avatar shakes at contact, not before.
    advance(179);
    expect(impactShake).not.toHaveBeenCalled();
    advance(1);
    expect(impactShake).toHaveBeenCalledWith(tiles.avatar, { lethal: false });

    advance(320);
    expect(copies(r.container)).toHaveLength(0);
    expect(cues.isStriking("ogre")).toBe(false);
  });

  it("is aria-hidden", () => {
    const { board } = makeBoard();
    const r = render(
      CombatStrikes as never,
      {
        view: viewWith([], [], "precombat_main"),
        boardEl: board,
        cues,
      } as never,
    );
    expect(r.container.querySelector("[data-combat-strikes]")?.getAttribute("aria-hidden")).toBe(
      "true",
    );
  });

  it("mounts nothing under reduced motion", () => {
    updateSettings("accessibility", "reduceMotion", true);
    const before = [stepE(100, "declare_attackers")];
    const next = [...before, stepE(120, "combat_damage"), hit(121, 0, "ogre", 1)];
    const { r } = play(before, next, [ogre], [ogre]);
    advance(600);
    expect(copies(r.container)).toHaveLength(0);
    expect(impactShake).not.toHaveBeenCalled();
  });

  it("mounts nothing with the combat toggle off", () => {
    updateSettings("animations", "combat", false);
    const before = [stepE(100, "declare_attackers")];
    const next = [...before, stepE(120, "combat_damage"), hit(121, 0, "ogre", 1)];
    const { r } = play(before, next, [ogre], [ogre]);
    advance(600);
    expect(copies(r.container)).toHaveLength(0);
  });

  it("a priming frame removes copies in flight and unhides their tiles", () => {
    const before = [stepE(100, "declare_attackers")];
    const next = [...before, stepE(120, "combat_damage"), hit(121, 0, "ogre", 1)];
    const { r } = play(before, next, [ogre], [ogre]);
    advance(0);
    expect(copies(r.container)).toHaveLength(1);

    cues.requestReprime();
    cues.frame(next, get(settings));
    flushSync();
    expect(copies(r.container)).toHaveLength(0);
    expect(cues.isStriking("ogre")).toBe(false);
    // And the contact that was coming never lands.
    advance(600);
    expect(impactShake).not.toHaveBeenCalled();
  });

  it("draws a dead blocker from the cache: the lethal shake, then it crumbles, in its place", async () => {
    const before = [
      stepE(100, "declare_attackers"),
      stepE(110, "declare_blockers"),
      blockE(111, "bear", "ogre"),
    ];
    const next = [
      ...before,
      stepE(120, "combat_damage"),
      hit(121, 0, "ogre", "bear"),
      died(122, "bear"),
    ];
    const { r, tiles } = play(before, next, [ogre, bear], [ogre]);
    // The frame's end state: the bear's tile is gone.
    tiles.bear.remove();

    advance(0);
    expect(copies(r.container).map((c) => c.dataset.strikeCopy)).toEqual(["ogre"]);
    advance(180);
    const dead = r.container.querySelector<HTMLElement>('[data-strike-mode="die"]');
    expect(dead?.dataset.strikeCopy).toBe("bear");
    expect(dead?.style.left).toBe("560px");
    expect(dead?.style.top).toBe("256px");
    // A dead copy hides no tile: there is none left.
    expect(cues.isStriking("bear")).toBe(false);
    // Its face takes the lethal shake, then crumbles into the planned
    // shards (the mocked shake resolves at once).
    const face = dead!.querySelector<HTMLElement>(".face")!;
    expect(impactShake).toHaveBeenCalledWith(face, { lethal: true });
    await Promise.resolve();
    await Promise.resolve();
    expect(crumble).toHaveBeenCalledTimes(1);
    const [crumbled, shards] = vi.mocked(crumble).mock.calls[0];
    expect(crumbled).toBe(face);
    expect(shards).toHaveLength(12);
    // 240 ms of lethal shake and 420 of crumble from contact.
    advance(659);
    expect(r.container.querySelector('[data-strike-mode="die"]')).not.toBeNull();
    advance(1);
    expect(r.container.querySelector('[data-strike-mode="die"]')).toBeNull();
  });

  it("flies a dead attacker home with a lethal flash, then crumbles it there", async () => {
    const before = [
      stepE(100, "declare_attackers"),
      stepE(110, "declare_blockers"),
      blockE(111, "bear", "ogre"),
    ];
    const next = [
      ...before,
      stepE(120, "combat_damage"),
      hit(121, 0, "ogre", "bear"),
      hit(122, 1, "bear", "ogre"),
      died(123, "ogre"),
    ];
    const { r, tiles } = play(before, next, [ogre, bear], [bear]);
    tiles.ogre.remove();

    advance(0);
    const [copy] = copies(r.container);
    expect(copy.dataset.strikeCopy).toBe("ogre");
    const [, , opts] = vi.mocked(lunge).mock.calls[0];
    expect(opts).toMatchObject({ lethal: true });
    expect(opts?.flash).toBe(copy.querySelector(".face"));
    await Promise.resolve();
    await Promise.resolve();
    expect(crumble).toHaveBeenCalledTimes(1);
    // Home at 500 ms, then 420 ms of crumble.
    advance(919);
    expect(copies(r.container).some((c) => c.dataset.strikeCopy === "ogre")).toBe(true);
    advance(1);
    expect(copies(r.container).some((c) => c.dataset.strikeCopy === "ogre")).toBe(false);
  });

  it("does not crumble an attacker that lived", async () => {
    const before = [stepE(100, "declare_attackers")];
    const next = [...before, stepE(120, "combat_damage"), hit(121, 0, "ogre", 1)];
    play(before, next, [ogre], [ogre]);
    advance(600);
    await Promise.resolve();
    expect(crumble).not.toHaveBeenCalled();
  });

  it("shakes the avatar with the lethal hit when the beat eliminates its player", () => {
    const before = [stepE(100, "declare_attackers")];
    const next: LogEvent[] = [
      ...before,
      stepE(120, "combat_damage"),
      hit(121, 0, "ogre", 1),
      { seq: 122, kind: "eliminated", turn: 3, seat: 1, cause: "life", text: "" },
    ];
    const { tiles } = play(before, next, [ogre], [ogre]);
    advance(180);
    expect(impactShake).toHaveBeenCalledWith(tiles.avatar, { lethal: true });
  });

  it("draws trample's streak from the blocker to the avatar at contact", () => {
    const before = [
      stepE(100, "declare_attackers"),
      stepE(110, "declare_blockers"),
      blockE(111, "bear", "ogre"),
    ];
    const next = [
      ...before,
      stepE(120, "combat_damage"),
      hit(121, 0, "ogre", "bear"),
      hit(122, 0, "ogre", 1),
    ];
    const { r } = play(before, next, [ogre, bear], [ogre, bear]);
    advance(0);
    // The lunge aims at the blocker, not the player.
    expect(r.container.querySelector("[data-strike-streak]")).toBeNull();
    advance(179);
    expect(r.container.querySelector("[data-strike-streak]")).toBeNull();
    advance(1);
    const el = r.container.querySelector<HTMLElement>("[data-strike-streak]");
    expect(el).not.toBeNull();
    // From the bear's top edge (its centre is 560, 256) up toward the
    // avatar (590, 50).
    expect(parseFloat(el!.style.top)).toBeLessThan(256);
    expect(parseFloat(el!.style.top)).toBeGreaterThan(190);
    expect(parseFloat(el!.style.getPropertyValue("--streak-angle"))).toBeLessThan(-80);
    expect(streak).toHaveBeenCalledTimes(1);
    // Drawn in the 60 ms hold, faded over 220.
    advance(279);
    expect(r.container.querySelector("[data-strike-streak]")).not.toBeNull();
    advance(1);
    expect(r.container.querySelector("[data-strike-streak]")).toBeNull();
  });

  it("draws no streak for an unblocked attacker", () => {
    const before = [stepE(100, "declare_attackers")];
    const next = [...before, stepE(120, "combat_damage"), hit(121, 0, "ogre", 1)];
    const { r } = play(before, next, [ogre], [ogre]);
    advance(200);
    expect(r.container.querySelector("[data-strike-streak]")).toBeNull();
  });

  it("skips a creature whose box was cached on a board of another size", () => {
    const before = [
      stepE(100, "declare_attackers"),
      stepE(110, "declare_blockers"),
      blockE(111, "bear", "ogre"),
    ];
    const next = [
      ...before,
      stepE(120, "combat_damage"),
      hit(121, 1, "bear", "ogre"),
      died(122, "ogre"),
    ];
    // The attacker died to a first-strike-less blocker: the bear lunges
    // (its attacker dealt it nothing), and the ogre's tile is gone.
    const { r, tiles } = play(before, next, [ogre, bear], [bear]);
    tiles.ogre.remove();
    // The table reflowed between the frames.
    boardSize = { w: 1000, h: 800 };

    advance(0);
    // The bear's live tile still places its copy, but the ogre has no
    // usable box to aim at, so the strike is skipped.
    expect(copies(r.container)).toHaveLength(0);
    advance(200);
    expect(r.container.querySelector('[data-strike-mode="die"]')).toBeNull();
  });

  it("drops a beat that would start more than a second late", () => {
    const before = [stepE(100, "declare_attackers")];
    const next = [...before, stepE(120, "combat_damage"), hit(121, 0, "ogre", 1)];
    const { board } = makeBoard();
    const r = render(
      CombatStrikes as never,
      {
        view: viewWith(before, [ogre], "declare_blockers"),
        boardEl: board,
        cues,
      } as never,
    );
    cues.frame(before, get(settings));
    r.setProps({ view: viewWith(next, [ogre]) } as never);
    cues.frame(next, get(settings));
    // The timer is starved (a busy main thread) and fires 1.2 s late.
    vi.setSystemTime(Date.now() + 1_200);
    advance(0);
    expect(copies(r.container)).toHaveLength(0);
  });

  it("schedules nothing for a frame that arrives while the page is hidden", () => {
    const before = [stepE(100, "declare_attackers")];
    const next = [...before, stepE(120, "combat_damage"), hit(121, 0, "ogre", 1)];
    cues.frame(before, get(settings));
    expect(cues.frame(next, get(settings), "hidden").cues).toEqual([]);
    expect(
      cues.frame([...next, hit(122, 0, "ogre", 1)], get(settings), "visible").cues,
    ).toHaveLength(1);
  });
});

describe("a Card under the clock", () => {
  it("hides while its copy flies, and renders as before otherwise", () => {
    const r = render(CombatCuesCard as never, { cues, card: ogre } as never);
    const tile = r.container.querySelector<HTMLElement>('[data-instance-id="ogre"]')!;
    expect(tile.style.visibility).toBe("");
    // No shake at rest: nothing inline, so the rule's 0px default holds.
    expect(tile.style.getPropertyValue("--impact-x")).toBe("");

    cues.strikeStart("ogre");
    flushSync();
    expect(tile.style.visibility).toBe("hidden");
    cues.strikeEnd("ogre");
    flushSync();
    expect(tile.style.visibility).toBe("");
  });
});
