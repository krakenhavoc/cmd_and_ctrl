import { existsSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// sounds.ts keeps module-level state (gains, the missing set, the burst
// memory), so each test loads a fresh copy of it. The suite runs in Node:
// window is a bare EventTarget and Audio is a fake.
class FakeAudio {
  static instances: FakeAudio[] = [];
  static broken = new Set<string>();
  src: string;
  volume = 1;
  playbackRate = 1;
  preload = "";
  private listeners = new Map<string, (() => void)[]>();
  play = vi.fn(() => {
    if (FakeAudio.broken.has(this.src)) {
      const err = new Error("The element has no supported sources.");
      err.name = "NotSupportedError";
      return Promise.reject(err);
    }
    return Promise.resolve();
  });
  load = vi.fn();
  constructor(src: string) {
    this.src = src;
    FakeAudio.instances.push(this);
  }
  addEventListener(type: string, fn: () => void) {
    this.listeners.set(type, [...(this.listeners.get(type) ?? []), fn]);
  }
  fire(type: string) {
    for (const fn of this.listeners.get(type) ?? []) fn();
  }
  cloneNode() {
    return new FakeAudio(this.src);
  }
}

async function load() {
  vi.resetModules();
  return import("./sounds");
}

// Time moves on 100 ms between plays, past the 40 ms same-name rate limit.
function tick() {
  vi.advanceTimersByTime(100);
}

// Let a rejected play() promise run its catch handler.
async function settle() {
  await Promise.resolve();
  await Promise.resolve();
}

function played(): FakeAudio[] {
  return FakeAudio.instances.filter((a) => a.play.mock.calls.length > 0);
}

describe("sounds", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    FakeAudio.instances = [];
    FakeAudio.broken = new Set();
    vi.stubGlobal("window", new EventTarget());
    vi.stubGlobal("Audio", FakeAudio);
    vi.spyOn(Math, "random").mockReturnValue(0);
  });
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  describe("categories", () => {
    it("names a category for every sound, and a manifest entry for it", async () => {
      const m = await load();
      for (const name of Object.keys(m.SOUND_MANIFEST) as (keyof typeof m.SOUND_MANIFEST)[]) {
        expect(m.SOUND_CATEGORIES).toContain(m.SOUND_CATEGORY[name]);
      }
      expect(Object.keys(m.SOUND_CATEGORY).sort()).toEqual(Object.keys(m.SOUND_MANIFEST).sort());
    });

    it("puts the new cues on the interface bus and the old ones where they belong", async () => {
      const m = await load();
      for (const n of ["your_turn", "priority", "prompt", "timer_warning"] as const) {
        expect(m.SOUND_CATEGORY[n]).toBe("interface");
      }
      expect(m.SOUND_CATEGORY.draw).toBe("cards");
      expect(m.SOUND_CATEGORY.attack).toBe("combat");
      expect(m.SOUND_CATEGORY.combat_resolve).toBe("combat");
    });

    it("plays at master x category x the caller's hint", async () => {
      const m = await load();
      m.setMasterVolume(0.8);
      m.setCategoryVolume("cards", 0.5);
      m.setCategoryVolume("combat", 1);
      expect(m.soundGain("draw")).toBeCloseTo(0.4);
      expect(m.soundGain("attack")).toBeCloseTo(0.8);

      m.play("draw", { volume: 0.5 });
      m.play("attack");
      const [draw, attack] = played();
      expect(draw.volume).toBeCloseTo(0.2);
      expect(attack.volume).toBeCloseTo(0.8);
    });

    it("silences one category without touching the others", async () => {
      const m = await load();
      m.setCategoryVolume("combat", 0);
      m.play("attack");
      m.play("draw");
      const [attack, draw] = played();
      expect(attack.volume).toBe(0);
      expect(draw.volume).toBe(1);
    });

    it("clamps gains to 0..1", async () => {
      const m = await load();
      m.setMasterVolume(3);
      m.setCategoryVolume("cards", -1);
      expect(m.soundGain("draw")).toBe(0);
      m.setCategoryVolume("cards", 5);
      expect(m.soundGain("draw")).toBe(1);
    });
  });

  describe("burst cap", () => {
    it("plays a logical sound once per batch", async () => {
      const m = await load();
      for (let i = 0; i < 12; i++) {
        m.play("combat_resolve", { batch: "g1:7" });
        tick();
      }
      expect(played()).toHaveLength(1);
    });

    it("lets a different sound, and the same sound in the next batch, through", async () => {
      const m = await load();
      m.play("damage", { batch: "g1:7" });
      tick();
      m.play("attack", { batch: "g1:7" });
      tick();
      m.play("damage", { batch: "g1:8" });
      expect(played()).toHaveLength(3);
    });

    it("does not cap a play that names no batch", async () => {
      const m = await load();
      m.play("draw");
      tick();
      m.play("draw");
      expect(played()).toHaveLength(2);
    });

    it("forgets a batch on reset, and its oldest ones past the memory", async () => {
      const m = await load();
      m.play("draw", { batch: 1 });
      tick();
      m.resetBurstCap();
      m.play("draw", { batch: 1 });
      expect(played()).toHaveLength(2);

      m.resetBurstCap();
      m.play("draw", { batch: 0 });
      tick();
      for (let b = 1; b <= 70; b++) m.play("tap", { batch: b });
      tick();
      m.play("draw", { batch: 0 });
      expect(played().filter((a) => a.src.includes("draw"))).toHaveLength(3);
    });

    it("does not burn the rate limit on a capped call", async () => {
      const m = await load();
      m.play("draw", { batch: "a" });
      tick();
      m.play("draw", { batch: "a" });
      m.play("draw", { batch: "b" });
      expect(played()).toHaveLength(2);
    });
  });

  describe("missing files", () => {
    it("lists only takes that exist in public/sounds", async () => {
      const m = await load();
      const absent = Object.values(m.SOUND_MANIFEST)
        .flat()
        .filter((u) => !existsSync(fileURLToPath(new URL(`../../public${u}`, import.meta.url))));
      expect(absent).toEqual([]);
    });

    it("falls back to an existing sound when every take of a cue is missing", async () => {
      const m = await load();
      for (const u of m.SOUND_MANIFEST.your_turn) FakeAudio.broken.add(u);
      m.play("your_turn");
      await settle();
      const sources = played().map((a) => a.src);
      // The first try hit the missing file; the retry plays turn_change.
      expect(sources[0]).toMatch(/your_turn-/);
      expect(sources.at(-1)).toMatch(/turn_change-/);
      expect(m.isMissing(m.SOUND_MANIFEST.your_turn[0])).toBe(true);
    });

    it("stays silent and does not throw when there is no fallback", async () => {
      const m = await load();
      for (const u of m.SOUND_MANIFEST.priority) FakeAudio.broken.add(u);
      expect(() => m.play("priority")).not.toThrow();
      await settle();
      tick();
      const before = played().length;
      // Learned: the next fire does not even try.
      m.play("priority");
      await settle();
      expect(played().length).toBe(before);
    });

    it("keeps playing the takes that exist", async () => {
      const m = await load();
      FakeAudio.broken.add(m.SOUND_MANIFEST.prompt[0]);
      m.play("prompt");
      await settle();
      tick();
      m.play("prompt");
      await settle();
      expect(played().at(-1)?.src).toBe(m.SOUND_MANIFEST.prompt[1]);
    });

    it("learns a missing file from a load error before the first play", async () => {
      const m = await load();
      m.armAudioOnFirstGesture();
      (window as unknown as EventTarget).dispatchEvent(new Event("pointerdown"));
      const url = m.SOUND_MANIFEST.timer_warning[0];
      FakeAudio.instances.find((a) => a.src === url)?.fire("error");
      expect(m.isMissing(url)).toBe(true);
      expect(m.isMissing(m.SOUND_MANIFEST.timer_warning[1])).toBe(false);
    });

    it("leaves signature sounds out of the first-gesture preload, for preloadSound to fetch", async () => {
      const m = await load();
      m.armAudioOnFirstGesture();
      (window as unknown as EventTarget).dispatchEvent(new Event("pointerdown"));
      const fetched = () => new Set(FakeAudio.instances.map((a) => a.src));
      expect(fetched().has(m.SOUND_MANIFEST.destroy[0])).toBe(true);
      expect(fetched().has(m.SOUND_MANIFEST.sig_toxic_deluge[0])).toBe(false);
      m.preloadSound("sig_toxic_deluge");
      expect(fetched().has(m.SOUND_MANIFEST.sig_toxic_deluge[0])).toBe(true);
      expect(fetched().has(m.SOUND_MANIFEST.sig_nicol_bolas[0])).toBe(false);
    });

    it("plays the generic cue a signature replaces until the signature's files land", async () => {
      const m = await load();
      for (const u of m.SOUND_MANIFEST.sig_blasphemous_act) FakeAudio.broken.add(u);
      m.play("sig_blasphemous_act");
      await settle();
      await settle();
      expect(played().at(-1)?.src).toBe(m.SOUND_MANIFEST.board_wipe[0]);
    });

    it("never resolves a fallback loop", async () => {
      const m = await load();
      for (const urls of Object.values(m.SOUND_MANIFEST))
        for (const u of urls) FakeAudio.broken.add(u);
      for (const name of Object.keys(m.SOUND_MANIFEST) as (keyof typeof m.SOUND_MANIFEST)[]) {
        expect(() => m.play(name)).not.toThrow();
        tick();
        await settle();
      }
    });
  });

  describe("ducking", () => {
    it("lowers the other buses by about 6 dB while an interface cue plays", async () => {
      const m = await load();
      const factors: number[] = [];
      m.onDuck((f) => factors.push(f));
      m.play("attack");
      const attack = played()[0];
      expect(attack.volume).toBe(1);

      m.play("priority");
      const cue = played()[1];
      expect(cue.volume).toBe(1);
      expect(attack.volume).toBeCloseTo(m.DUCK_FACTOR);
      expect(m.DUCK_FACTOR).toBeCloseTo(0.5, 1);
      expect(factors.at(-1)).toBeCloseTo(m.DUCK_FACTOR);

      // A sound that starts mid-cue starts ducked.
      m.play("draw");
      expect(played()[2].volume).toBeCloseTo(m.DUCK_FACTOR);

      cue.fire("ended");
      expect(attack.volume).toBe(1);
      expect(factors.at(-1)).toBe(1);
      m.play("tap");
      expect(played()[3].volume).toBe(1);
    });

    it("releases after a timeout when the cue never ends", async () => {
      const m = await load();
      m.play("priority");
      m.play("attack");
      expect(played()[1].volume).toBeCloseTo(m.DUCK_FACTOR);
      vi.advanceTimersByTime(2500);
      expect(played()[1].volume).toBe(1);
    });

    it("releases when the cue's file is missing", async () => {
      const m = await load();
      for (const u of m.SOUND_MANIFEST.priority) FakeAudio.broken.add(u);
      const factors: number[] = [];
      m.onDuck((f) => factors.push(f));
      m.play("priority");
      await settle();
      expect(factors.at(-1)).toBe(1);
    });

    it("does not duck for a non-interface sound, and stacks cues", async () => {
      const m = await load();
      m.play("attack");
      m.play("draw");
      expect(played().every((a) => a.volume === 1)).toBe(true);

      m.play("priority");
      tick();
      m.play("prompt");
      const [priority, prompt] = played().slice(2);
      priority.fire("ended");
      // One cue still playing: the table stays ducked.
      expect(played()[0].volume).toBeCloseTo(m.DUCK_FACTOR);
      prompt.fire("ended");
      expect(played()[0].volume).toBe(1);
    });

    it("ducks on top of the category volume", async () => {
      const m = await load();
      m.setCategoryVolume("combat", 0.5);
      m.play("attack");
      m.play("prompt");
      expect(played()[0].volume).toBeCloseTo(0.5 * m.DUCK_FACTOR);
    });
  });

  it("no-ops while muted", async () => {
    const m = await load();
    m.setMuted(true);
    m.play("draw");
    expect(played()).toHaveLength(0);
  });
});
