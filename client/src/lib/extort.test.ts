// #4544 — the rules of the Scales of Souls effect, pure: what an extort
// drain is in the log, how a run of drains queues, the clock and the
// geometry, and when a life popup stands down. The draw is
// extort.render.test.ts.

import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";
import {
  EXTORT_BEATS,
  EXTORT_HIT_MS,
  EXTORT_MS,
  SCALE_TILT_DEG,
  arrive,
  darkPanAt,
  emptyExtortQueue,
  emptyExtortTracker,
  extortBatches,
  extortK,
  extortMotion,
  extortMs,
  finish,
  flight,
  foldExtort,
  isExtortLabel,
  lightPanAt,
  mergeDrains,
  popupIsExtort,
  scaleBox,
  type ExtortDrain,
} from "./extort";
import { EFFECT_SHORT_CAP_MS, EFFECT_STILL_HOLD_MS } from "./tableEffects";
import type { GameView, LogEvent } from "./protocol";

let seq = 0;
const resolve = (batch: number, label = "Crypt Ghast — extort", seat = 0): LogEvent => ({
  seq: ++seq,
  kind: "resolve",
  seat,
  card_id: "ghast",
  label,
  batch,
  text: "",
});
const life = (batch: number, seat: number, amount: number): LogEvent => ({
  seq: ++seq,
  kind: "life",
  seat,
  card_id: "ghast",
  amount,
  batch,
  text: "",
});

describe("reading an extort drain from the log", () => {
  it("knows an extort label by its end", () => {
    expect(isExtortLabel("Pontiff of Blight — extort")).toBe(true);
    expect(isExtortLabel("Sorin of House Markov — extort")).toBe(true);
    expect(isExtortLabel("Gray Merchant of Asphodel — drain")).toBe(false);
    expect(isExtortLabel("extort")).toBe(false);
    expect(isExtortLabel(undefined)).toBe(false);
    expect(isExtortLabel("")).toBe(false);
  });

  it("a drain of one opponent: their loss and the player's gain", () => {
    const { drains } = foldExtort(emptyExtortTracker(), [
      resolve(7),
      life(7, 1, -1),
      life(7, 0, 1),
    ]);
    expect(drains).toEqual([
      { victims: [{ seat: 1, amount: 1 }], gainers: [{ seat: 0, amount: 1 }], count: 1 },
    ]);
  });

  it("a drain of two, and of three, opponents: one soul each", () => {
    const two = foldExtort(emptyExtortTracker(), [
      resolve(7),
      life(7, 1, -1),
      life(7, 3, -1),
      life(7, 0, 2),
    ]).drains;
    expect(two).toHaveLength(1);
    expect(two[0].victims.map((v) => v.seat)).toEqual([1, 3]);
    expect(two[0].gainers).toEqual([{ seat: 0, amount: 2 }]);

    const three = foldExtort(emptyExtortTracker(), [
      resolve(7),
      life(7, 1, -1),
      life(7, 2, -1),
      life(7, 3, -1),
      life(7, 0, 3),
    ]).drains;
    expect(three[0].victims.map((v) => v.seat)).toEqual([1, 2, 3]);
    expect(three[0].gainers).toEqual([{ seat: 0, amount: 3 }]);
  });

  it("plays nothing for life lost any other way", () => {
    // Gray Merchant, Blood Artist, a fetch land: life entries, no extort resolve.
    const other: LogEvent = { ...resolve(4, "Gray Merchant of Asphodel — drain"), seat: 0 };
    expect(foldExtort(emptyExtortTracker(), [life(4, 1, -1), life(4, 0, 1)]).drains).toEqual([]);
    expect(foldExtort(emptyExtortTracker(), [other, life(4, 1, -2), life(4, 0, 2)]).drains).toEqual(
      [],
    );
    // Combat damage is not a life entry at all.
    const dmg: LogEvent = {
      seq: ++seq,
      kind: "damage",
      seat: 0,
      target_seat: 1,
      amount: 3,
      text: "",
    };
    expect(foldExtort(emptyExtortTracker(), [dmg]).drains).toEqual([]);
  });

  it("plays nothing when the {W/B} was declined or every loss was replaced away", () => {
    expect(foldExtort(emptyExtortTracker(), [resolve(9)]).drains).toEqual([]);
    expect(foldExtort(emptyExtortTracker(), [resolve(9), life(9, 0, 0)]).drains).toEqual([]);
  });

  it("does not take life of another batch for the drain's", () => {
    const { drains } = foldExtort(emptyExtortTracker(), [
      life(6, 2, -5),
      resolve(7),
      life(7, 1, -1),
      life(7, 0, 1),
      life(8, 3, -4),
    ]);
    expect(drains).toHaveLength(1);
    expect(drains[0].victims).toEqual([{ seat: 1, amount: 1 }]);
  });

  it("finds the drain when the life comes before its resolve line (a prompt held the line back)", () => {
    const { drains } = foldExtort(emptyExtortTracker(), [
      life(7, 1, -1),
      life(7, 0, 1),
      resolve(7),
    ]);
    expect(drains).toHaveLength(1);
  });

  it("finds the drain when the life comes in a later frame than the resolve line", () => {
    const first = foldExtort(emptyExtortTracker(), [resolve(7)]);
    expect(first.drains).toEqual([]);
    const second = foldExtort(first.tracker, [life(7, 1, -1), life(7, 0, 1)]);
    expect(second.drains).toHaveLength(1);
  });

  it("drains each batch once", () => {
    const first = foldExtort(emptyExtortTracker(), [resolve(7), life(7, 1, -1), life(7, 0, 1)]);
    // The same batch's gain turning up again does not play a second scale.
    const again = foldExtort(first.tracker, [life(7, 0, 1)]);
    expect(again.drains).toEqual([]);
  });

  it("two extort triggers in one frame are two drains, each its own", () => {
    const { drains } = foldExtort(emptyExtortTracker(), [
      resolve(7),
      life(7, 1, -1),
      life(7, 0, 1),
      resolve(8, "Pontiff of Blight — extort"),
      life(8, 1, -1),
      life(8, 0, 1),
    ]);
    expect(drains).toHaveLength(2);
  });

  it("names the extort player from the resolve line when the gain was replaced away", () => {
    const { drains } = foldExtort(emptyExtortTracker(), [resolve(7, undefined, 2), life(7, 1, -1)]);
    expect(drains[0].gainers).toEqual([{ seat: 2, amount: 0 }]);
  });

  it("ignores an extort resolve with no batch (an entry from before batches)", () => {
    const old: LogEvent = { ...resolve(1), batch: undefined };
    expect(foldExtort(emptyExtortTracker(), [old, life(1, 1, -1)]).drains).toEqual([]);
  });

  it("does not mutate the tracker it was given", () => {
    const t = emptyExtortTracker();
    foldExtort(t, [resolve(7), life(7, 1, -1)]);
    expect(t.pending).toEqual([]);
  });

  it("remembers only a few resolutions", () => {
    let t = emptyExtortTracker();
    for (let b = 1; b <= 20; b++) t = foldExtort(t, [resolve(b)]).tracker;
    expect(t.pending.length).toBeLessThanOrEqual(8);
    expect(t.pending.at(-1)?.batch).toBe(20);
  });
});

describe("queueing a run of drains", () => {
  const drain = (victimSeat: number, gainSeat = 0, count = 1): ExtortDrain => ({
    victims: [{ seat: victimSeat, amount: 1 }],
    gainers: [{ seat: gainSeat, amount: 1 }],
    count,
  });

  it("merging adds per-seat amounts and the trigger count", () => {
    const m = mergeDrains([drain(1), drain(1), drain(2)])!;
    expect(m.victims).toEqual([
      { seat: 1, amount: 2 },
      { seat: 2, amount: 1 },
    ]);
    expect(m.gainers).toEqual([{ seat: 0, amount: 3 }]);
    expect(m.count).toBe(3);
    expect(mergeDrains([])).toBeNull();
  });

  it("the first drain starts at once", () => {
    const a = arrive(emptyExtortQueue, drain(1));
    expect(a.start).toEqual(drain(1));
    expect(a.queue.active).toEqual(drain(1));
    expect(a.queue.waiting).toBeNull();
  });

  it("a drain that arrives during a scale waits; the rest merge into the one waiting", () => {
    let q = arrive(emptyExtortQueue, drain(1)).queue;
    const second = arrive(q, drain(2));
    expect(second.start).toBeNull();
    q = second.queue;
    for (let i = 0; i < 8; i++) q = arrive(q, drain(3)).queue;
    // Ten triggers in a row: the scale in play, and ONE waiting for nine.
    expect(q.active?.count).toBe(1);
    expect(q.waiting?.count).toBe(9);
    expect(q.waiting?.victims).toEqual([
      { seat: 2, amount: 1 },
      { seat: 3, amount: 8 },
    ]);
  });

  it("finishing starts the waiting one, then the table is clear", () => {
    let q = arrive(emptyExtortQueue, drain(1)).queue;
    q = arrive(q, drain(2)).queue;
    const next = finish(q);
    expect(next.start?.victims).toEqual([{ seat: 2, amount: 1 }]);
    expect(next.queue.waiting).toBeNull();
    const done = finish(next.queue);
    expect(done.start).toBeNull();
    expect(done.queue).toEqual(emptyExtortQueue);
  });
});

describe("the clock", () => {
  it("the coin hits 1.2 to 1.4 s in, the sound is cut to it", () => {
    expect(EXTORT_HIT_MS).toBeGreaterThanOrEqual(1200);
    expect(EXTORT_HIT_MS).toBeLessThanOrEqual(1400);
    expect(EXTORT_HIT_MS).toBe(EXTORT_BEATS.hit);
  });

  it("the beats run in order and inside the scene, about 2 to 3 s", () => {
    const b = EXTORT_BEATS;
    expect(b.soulsFrom).toBeLessThan(b.soulsTo);
    expect(b.soulsTo).toBeLessThanOrEqual(b.hit);
    expect(b.hit).toBeLessThan(b.sparkFrom);
    expect(b.sparkFrom).toBeLessThan(b.sparkTo);
    expect(b.sparkTo).toBeLessThanOrEqual(b.fadeFrom);
    expect(b.fadeFrom).toBeLessThan(EXTORT_MS);
    expect(EXTORT_MS).toBeGreaterThanOrEqual(2000);
    expect(EXTORT_MS).toBeLessThanOrEqual(3000);
  });

  it("full motion scales with animations.speed; a static scene is a hold speed does not scale", () => {
    expect(extortMs("full", 1)).toBe(EXTORT_MS);
    expect(extortMs("full", 2)).toBe(EXTORT_MS * 2);
    expect(extortK("full", 2)).toBe(2);
    expect(extortMs("short", 2)).toBe(EFFECT_SHORT_CAP_MS);
    expect(extortMs("still", 2)).toBe(EFFECT_STILL_HOLD_MS);
    expect(extortK("still", 2)).toBe(1);
    // A nonsense speed is 1.
    expect(extortMs("full", 0)).toBe(EXTORT_MS);
  });

  it("only Full has motion; Short and every reduced-motion signal show a static scale", () => {
    expect(extortMotion("full")).toBe("full");
    expect(extortMotion("short")).toBe("still");
    expect(extortMotion("still")).toBe("still");
  });
});

describe("the geometry", () => {
  it("is compact and centred, between 150 and 230 px wide", () => {
    const big = scaleBox({ w: 1600, h: 900 });
    expect(big.width).toBe(230);
    expect(big.x + big.width / 2).toBeCloseTo(800, 0);
    expect(big.y + big.height / 2).toBeCloseTo(450, 0);
    const phone = scaleBox({ w: 390, h: 700 });
    expect(phone.width).toBe(150);
    const mid = scaleBox({ w: 1000, h: 700 });
    expect(mid.width).toBe(200);
    // The drawing's own proportions.
    expect(big.height).toBe(Math.round((230 * 320) / 360));
  });

  it("the dark pan hangs lower and further in than the light pan before the coin", () => {
    const box = scaleBox({ w: 1200, h: 800 });
    const dark = darkPanAt(box);
    const light = lightPanAt(box);
    expect(SCALE_TILT_DEG).toBeLessThan(0);
    expect(dark.x).toBeLessThan(light.x);
    // The orb's own y when level is 185 of 320; the lean puts the dark pan below that.
    const levelY = box.y + (185 / 320) * box.height;
    expect(dark.y).toBeGreaterThan(levelY);
    expect(light.y).toBeCloseTo(levelY, 5);
    // Both pans are over the scale's own box.
    for (const p of [dark, light]) {
      expect(p.x).toBeGreaterThan(box.x);
      expect(p.x).toBeLessThan(box.x + box.width);
    }
  });

  it("a flight starts at the player, relative to the pan, and staggers", () => {
    const f0 = flight({ x: 100, y: 40 }, { x: 600, y: 420 }, 0);
    expect(f0).toEqual({ dx: -500, dy: -380, delay: 0 });
    expect(flight({ x: 0, y: 0 }, { x: 0, y: 0 }, 2).delay).toBeGreaterThan(0);
  });
});

describe("a life popup that shows an extort drain", () => {
  const viewOf = (log: LogEvent[]): GameView => ({ log }) as unknown as GameView;

  it("is one the cue covers: the opponent's loss and the player's gain", () => {
    const log = [resolve(7), life(7, 1, -1), life(7, 0, 1)];
    expect(popupIsExtort(viewOf(log), 1, [-1])).toBe(true);
    expect(popupIsExtort(viewOf(log), 0, [1])).toBe(true);
    expect(extortBatches(log)).toEqual(new Set([7]));
  });

  it("is not another seat's, nor life lost another way", () => {
    const log = [resolve(7), life(7, 1, -1), life(7, 0, 1), life(9, 2, -3)];
    expect(popupIsExtort(viewOf(log), 2, [-3])).toBe(false);
    expect(popupIsExtort(viewOf(log), 3, [-1])).toBe(false);
    expect(popupIsExtort(viewOf([life(9, 1, -1)]), 1, [-1])).toBe(false);
  });

  it("holds only when the popup lines up with the log, as lifeCause does", () => {
    const log = [resolve(7), life(7, 1, -1)];
    expect(popupIsExtort(viewOf(log), 1, [-2])).toBe(false);
    expect(popupIsExtort(viewOf(log), 1, [-1, -1])).toBe(false);
    expect(popupIsExtort(viewOf(log), -1, [-1])).toBe(false);
    expect(popupIsExtort(viewOf(log), 1, [])).toBe(false);
  });

  it("covers a popup that also holds other changes, by the newest ones", () => {
    const dmg: LogEvent = {
      seq: ++seq,
      kind: "damage",
      seat: 0,
      target_seat: 1,
      amount: 2,
      text: "",
    };
    const log = [dmg, resolve(7), life(7, 1, -1)];
    expect(popupIsExtort(viewOf(log), 1, [-2, -1])).toBe(true);
    expect(popupIsExtort(viewOf(log), 1, [-2])).toBe(false);
  });
});

describe("the source", () => {
  const src = readFileSync(
    new URL("./components/board/ExtortScales.svelte", import.meta.url),
    "utf8",
  );

  it("sits at z 57 on the ladder, takes no pointer events and is hidden from assistive tech", () => {
    expect(src).toMatch(/z-index:\s*57;/);
    expect(src).toMatch(/pointer-events:\s*none;/);
    expect(src).toMatch(/aria-hidden="true"/);
  });

  it("carries its own prefers-reduced-motion block", () => {
    expect(src).toMatch(/@media \(prefers-reduced-motion: reduce\)/);
  });

  it("levels its arms at the beats the clock names", () => {
    const pct = (ms: number) => ((ms / EXTORT_MS) * 100).toFixed(2);
    // The beam starts to level a beat before the souls land and is level at the coin.
    expect(src).toContain(`${pct(EXTORT_BEATS.soulsTo - 50)}%`);
    expect(src).toContain(`${pct(EXTORT_BEATS.hit)}%`);
    expect(src).toContain(`${pct(EXTORT_BEATS.fadeFrom)}%`);
  });
});
