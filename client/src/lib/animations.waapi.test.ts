// @vitest-environment jsdom
//
// #3076: GSAP was replaced by the Web Animations API. jsdom has none, so
// these stub Element.prototype.animate and pin what each helper asks
// for (keyframes, timing, easing), and that every one of them stays
// silent when combat / motion is off.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  IMPACT_GLOW,
  STRIKE_BACK_MS,
  STRIKE_HOLD_MS,
  STRIKE_MS,
  STRIKE_OUT_MS,
  animateTap,
  crumble,
  etbPulse,
  flyTo,
  impactShake,
  lunge,
  setAnimationConfig,
  streak,
} from "./animations";
import { EASE, cancelAnimations } from "./motion";
import { crumbleShards } from "./combatStrikes";

interface Call {
  el: Element;
  keyframes: Keyframe[];
  options: KeyframeAnimationOptions;
  anim: FakeAnim;
}
class FakeAnim {
  onfinish: (() => void) | null = null;
  oncancel: (() => void) | null = null;
  cancelled = false;
  finish() {
    this.onfinish?.();
  }
  cancel() {
    this.cancelled = true;
    this.oncancel?.();
  }
}

let calls: Call[] = [];
const proto = Element.prototype as unknown as { animate?: unknown };
let had: unknown;

beforeEach(() => {
  calls = [];
  had = proto.animate;
  proto.animate = function (
    this: Element,
    keyframes: Keyframe[],
    options: KeyframeAnimationOptions,
  ) {
    const anim = new FakeAnim();
    calls.push({ el: this, keyframes, options, anim });
    return anim;
  };
});

afterEach(() => {
  proto.animate = had;
  setAnimationConfig({
    enabled: true,
    combat: true,
    speed: 1,
    cardTap: true,
    cardUntap: true,
    particlesEtb: true,
    cardPlay: true,
  });
  document.body.innerHTML = "";
});

function div(): HTMLDivElement {
  const el = document.createElement("div");
  document.body.appendChild(el);
  return el;
}

describe("lunge", () => {
  it("flies out, holds, and comes home as one animation at the ADR 0134 timings", () => {
    const el = div();
    void lunge(el, { x: 120, y: -40 });
    expect(calls).toHaveLength(1);
    const [c] = calls;
    expect(c.options.duration).toBe(STRIKE_MS);
    expect(c.keyframes.map((k) => k.translate)).toEqual([
      "0px 0px",
      "120px -40px",
      "120px -40px",
      "0px 0px",
    ]);
    expect(c.keyframes.map((k) => k.offset)).toEqual([
      0,
      STRIKE_OUT_MS / STRIKE_MS,
      (STRIKE_OUT_MS + STRIKE_HOLD_MS) / STRIKE_MS,
      1,
    ]);
    expect(c.keyframes[0].easing).toBe(EASE.power2In);
    expect(c.keyframes[2].easing).toBe(EASE.power3Out);
    expect(STRIKE_OUT_MS + STRIKE_HOLD_MS + STRIKE_BACK_MS).toBe(STRIKE_MS);
  });

  it("scales every phase by the speed setting", () => {
    setAnimationConfig({ speed: 2 });
    void lunge(div(), { x: 1, y: 1 });
    expect(calls[0].options.duration).toBe(STRIKE_MS * 2);
  });

  it("starts the flash at contact, red for a lethal one", () => {
    const el = div();
    const face = div();
    void lunge(el, { x: 5, y: 5 }, { flash: face, lethal: true });
    expect(calls).toHaveLength(2);
    const flash = calls[1];
    expect(flash.el).toBe(face);
    expect(flash.options.delay).toBe(STRIKE_OUT_MS);
    expect(flash.options.duration).toBe(240);
    expect(flash.keyframes).toEqual([{ "--impact-glow": IMPACT_GLOW }, { "--impact-glow": 0 }]);
    expect(face.style.getPropertyValue("--impact-wash")).toBe("#e5484d");
  });

  it("resolves when the copy is home, and when it is cancelled", async () => {
    const el = div();
    let done = false;
    void lunge(el, { x: 1, y: 1 }).then(() => (done = true));
    await Promise.resolve();
    expect(done).toBe(false);
    cancelAnimations(el);
    await new Promise((r) => setTimeout(r, 0));
    expect(done).toBe(true);
    expect(calls[0].anim.cancelled).toBe(true);
  });

  it("plays nothing with combat motion off", async () => {
    setAnimationConfig({ combat: false });
    await lunge(div(), { x: 1, y: 1 }, { flash: div() });
    setAnimationConfig({ combat: true, enabled: false });
    await lunge(div(), { x: 1, y: 1 });
    expect(calls).toHaveLength(0);
  });
});

describe("impactShake", () => {
  it("swings in three half-cycles through the registered shake variable", async () => {
    const el = div();
    const p = impactShake(el);
    expect(el.style.getPropertyValue("--impact-x")).toBe("var(--impact-shake)");
    const [shake, glow] = calls;
    expect(shake.options.duration).toBe(160);
    expect(shake.keyframes.map((k) => k["--impact-shake"])).toEqual([
      "0px",
      "4px",
      "-4px",
      "4px",
      "0px",
    ]);
    expect(shake.keyframes.map((k) => k.offset)).toEqual([0, 1 / 6, 1 / 2, 5 / 6, 1]);
    expect(glow.options.duration).toBe(160);
    shake.anim.finish();
    glow.anim.finish();
    await p;
    // Cleared, not zeroed.
    expect(el.style.getPropertyValue("--impact-x")).toBe("");
  });

  it("is bigger and longer for a lethal hit", () => {
    void impactShake(div(), { lethal: true });
    expect(calls[0].options.duration).toBe(240);
    expect(calls[0].keyframes[1]["--impact-shake"]).toBe("8px");
  });

  it("plays nothing with combat motion off", async () => {
    setAnimationConfig({ combat: false });
    const el = div();
    await impactShake(el);
    expect(calls).toHaveLength(0);
    expect(el.style.getPropertyValue("--impact-x")).toBe("");
  });
});

describe("crumble", () => {
  it("drops each shard on its own delay, all landing at the end of the crumble", () => {
    const el = div();
    el.appendChild(document.createElement("img"));
    const shards = crumbleShards("bear", 80, 112);
    void crumble(el, shards);
    expect(calls).toHaveLength(shards.length);
    shards.forEach((sh, i) => {
      const c = calls[i];
      expect(c.options.delay).toBeCloseTo(sh.delay * 420);
      expect((c.options.delay as number) + (c.options.duration as number)).toBeCloseTo(
        Math.max(420, sh.delay * 420 + 1),
      );
      expect(c.options.fill).toBe("forwards");
      expect(c.keyframes[1]).toEqual({
        translate: `${sh.dx}px ${sh.dy}px`,
        rotate: `${sh.rot}deg`,
        opacity: 0,
      });
    });
  });

  it("makes and plays nothing with motion off", async () => {
    setAnimationConfig({ enabled: false });
    const el = div();
    await crumble(el, crumbleShards("bear", 80, 112));
    expect(calls).toHaveLength(0);
  });
});

describe("streak", () => {
  it("draws along x during the hold, then fades", () => {
    void streak(div());
    const [c] = calls;
    expect(c.options.duration).toBe(60 + 220);
    expect(c.keyframes.map((k) => [k.scale, k.opacity])).toEqual([
      ["0 1", 1],
      ["1 1", 1],
      ["1 1", 0],
    ]);
    expect(c.keyframes[1].offset).toBeCloseTo(60 / 280);
    expect(c.options.fill).toBe("forwards");
  });

  it("plays nothing with combat motion off", async () => {
    setAnimationConfig({ combat: false });
    await streak(div());
    expect(calls).toHaveLength(0);
  });
});

describe("flyTo", () => {
  it("travels to the landing and fades, and replaces a flight under way", async () => {
    const el = div();
    const first = flyTo(el, { x: 10, y: 20, scale: 0.5 }, 400);
    const second = flyTo(el, { x: 30, y: 40, scale: 0.25 }, 400);
    expect(calls[0].anim.cancelled).toBe(true);
    await first;
    const c = calls[1];
    expect(c.keyframes).toEqual([{ translate: "30px 40px", scale: "0.25", opacity: 0.35 }]);
    expect(c.options).toMatchObject({ duration: 400, easing: EASE.power2InOut, fill: "forwards" });
    expect(el.style.transformOrigin).toMatch(/^0(px)? 0(px)?$/);
    c.anim.finish();
    await second;
  });

  it("settles at once with motion off", async () => {
    setAnimationConfig({ enabled: false });
    const el = div();
    void flyTo(el, { x: 1, y: 2, scale: 1 }, 400);
    expect(calls[0].options.duration).toBe(1);
  });
});

describe("etbPulse", () => {
  it("punches in from 0.55 over 320 ms and cleans up on destroy", () => {
    const el = div();
    const act = etbPulse(el) as { destroy(): void };
    const [c] = calls;
    expect(c.keyframes).toEqual([
      { scale: 0.55, opacity: 0 },
      { scale: 1, opacity: 1 },
    ]);
    expect(c.options.duration).toBe(320);
    act.destroy();
    expect(c.anim.cancelled).toBe(true);
  });

  it("restarts rather than layers on a re-mount", () => {
    const el = div();
    etbPulse(el);
    etbPulse(el);
    expect(calls[0].anim.cancelled).toBe(true);
    expect(calls[1].anim.cancelled).toBe(false);
  });

  it("plays nothing with the effect off", () => {
    setAnimationConfig({ particlesEtb: false });
    etbPulse(div());
    expect(calls).toHaveLength(0);
  });
});

describe("animateTap", () => {
  it("snaps with no animation when the property is not at a known angle", () => {
    const el = div();
    animateTap(el, true);
    expect(el.style.getPropertyValue("--tap-rot")).toBe("90deg");
    animateTap(el, false);
    expect(el.style.getPropertyValue("--tap-rot")).toBe("0deg");
  });

  it("eases from the angle it stands at when the browser reports one", () => {
    const el = div();
    const gcs = vi.spyOn(window, "getComputedStyle").mockReturnValue({
      getPropertyValue: () => "0deg",
    } as unknown as CSSStyleDeclaration);
    animateTap(el, true);
    gcs.mockRestore();
    const [c] = calls;
    expect(c.keyframes).toEqual([{ "--tap-rot": "0deg" }, { "--tap-rot": "90deg" }]);
    expect(c.options).toMatchObject({ duration: 180, easing: EASE.power2Out });
    expect(el.style.getPropertyValue("--tap-rot")).toBe("90deg");
  });

  it("snaps with the per-effect flag off, and never animates an untap when untap is off", () => {
    const el = div();
    const gcs = vi.spyOn(window, "getComputedStyle").mockReturnValue({
      getPropertyValue: () => "90deg",
    } as unknown as CSSStyleDeclaration);
    setAnimationConfig({ cardUntap: false });
    animateTap(el, false);
    gcs.mockRestore();
    expect(calls).toHaveLength(0);
    expect(el.style.getPropertyValue("--tap-rot")).toBe("0deg");
  });
});
