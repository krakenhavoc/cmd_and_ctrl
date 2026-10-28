// S09 polish-pass animation helpers, backed by the Web Animations API
// (motion.ts; GSAP was dropped in #3076).
//
// The Card.svelte component composes its `transform` from CSS custom
// properties so multiple effects can stack without fighting each
// other:
//
//   transform: rotate(var(--tap-rot, 0deg)) translateY(50%) scale(var(--hover-scale, 1)) translateY(-50%)
//
// CSS owns the static state (hover lift, base rotation). The animated
// transitions tween the relevant custom property (registered with
// @property in app.css so it interpolates), so a CSS hover and an
// in-flight tap animation compose cleanly. This also means a
// reduced-motion fallback can short-circuit the animation and just
// snap the var to its target value.
//
// Motion that is not a custom property uses the individual CSS
// transform properties (`translate`, `rotate`, `scale`), which apply
// before the element's own `transform`, so a card or popup that keeps
// a CSS transform of its own (translateX(-50%) on the damage popup)
// is moved on top of it, never over it.

import { backOut, cubicIn, cubicOut } from "svelte/easing";
import { EASE, backOut as backOutCss, cancelAnimations, canAnimate, run } from "./motion";
import type { TransitionConfig } from "svelte/transition";
import type { Action } from "svelte/action";
import type { EffectsMode } from "./tableEffects";

// Per-effect gating, settable from the S11.5 settings panel via
// setAnimationConfig. Each animation function checks the relevant
// per-effect boolean (gated under `enabled` as the master). When
// disabled, the function snaps to its end state without a tween so
// downstream UI (taps, deals, ETB pulses) still reach the right
// final visual — the user just doesn't see motion.
//
// `speed` multiplies durations: speed=2 → twice as slow; speed=0.5
// → twice as fast. Read each call instead of pre-computed at
// settings-change time so a flip mid-animation is felt by the next
// fire without needing every duration constant to be re-derived.
type AnimationConfig = {
  enabled: boolean;
  speed: number;
  cardDraw: boolean;
  cardPlay: boolean;
  cardTap: boolean;
  cardUntap: boolean;
  cardFlip: boolean;
  particlesEtb: boolean;
  damagePopups: boolean;
  dice: boolean;
  // ADR 0134 §5: attackers lunge and hits land.
  combat: boolean;
  // ADR 0119 amendment 2026-10-16 §D: the stack's arrival trace
  // (StackArrival.svelte reads the setting itself; kept here so the
  // whole group is mirrored).
  stackArrival: boolean;
  // ADR 0209 §1: the generic table effects (destroy, exile, board wipe ...).
  // "off" plays none of their motion; their outcome badge still shows.
  effects: EffectsMode;
};
const cfg: AnimationConfig = {
  enabled: true,
  speed: 1,
  cardDraw: true,
  cardPlay: true,
  cardTap: true,
  cardUntap: true,
  cardFlip: true,
  particlesEtb: true,
  damagePopups: true,
  dice: true,
  combat: true,
  stackArrival: true,
  effects: "full",
};
export function setAnimationConfig(next: Partial<AnimationConfig>): void {
  Object.assign(cfg, next);
}

// gatedDuration applies the speed multiplier and forces a snap
// (~1ms) when the master switch or per-effect flag is off. The 1ms
// floor — instead of 0 — is because some Svelte transitions skip
// rendering the final frame on duration: 0.
function gatedDuration(baseMs: number, perEffect: boolean): number {
  if (!cfg.enabled || !perEffect) return 1;
  return Math.max(1, baseMs * cfg.speed);
}

// TAP_DURATION is short enough that a click-to-tap feels responsive
// (under the 100ms perception threshold for "instant") but long enough
// that the rotation reads as motion rather than a jump cut.
const TAP_DURATION = 0.18;

// animateTap rotates the given element to 90° (tapped) or 0° (untapped)
// by animating the --tap-rot CSS var from where it stands now. A tap
// that arrives mid-rotation replaces the one in flight and starts from
// the angle it had reached. Returns the animation, or null when it
// snapped (motion off, already there, or no Web Animations); most
// callers can ignore the return value.
const tapping = new WeakMap<HTMLElement, Animation>();
export function animateTap(el: HTMLElement, tapped: boolean): Animation | null {
  // Per-effect gate distinguishes tap (cardTap) from untap
  // (cardUntap) so a user who likes the tap motion but finds untap-
  // all visually noisy at end-of-turn can disable just that side.
  const allowed = tapped ? cfg.cardTap : cfg.cardUntap;
  const dur = cfg.enabled && allowed ? TAP_DURATION * cfg.speed : 0;
  const target = tapped ? "90deg" : "0deg";
  // Read where the rotation stands before the old animation is dropped.
  const from = canAnimate(el) ? getComputedStyle(el).getPropertyValue("--tap-rot").trim() : "";
  tapping.get(el)?.cancel();
  tapping.delete(el);
  el.style.setProperty("--tap-rot", target);
  if (dur <= 0 || from === "" || from === target) return null;
  const a = el.animate([{ "--tap-rot": from }, { "--tap-rot": target }], {
    duration: dur * 1000,
    easing: EASE.power2Out,
  });
  tapping.set(el, a);
  const clear = () => {
    if (tapping.get(el) === a) tapping.delete(el);
  };
  a.onfinish = clear;
  a.oncancel = clear;
  return a;
}

// pose paints one frame of a Svelte transition's tick. Svelte owns the
// timing; this only writes the values, on the individual transform
// properties so the element's own `transform` is left alone. At rest
// (t = 1) it clears them, so a settled element carries no inline style
// from the transition.
function pose(node: HTMLElement, t: number, y: number, scale: number): void {
  if (t >= 1) {
    node.style.removeProperty("translate");
    node.style.removeProperty("scale");
    node.style.removeProperty("opacity");
    return;
  }
  node.style.translate = `0px ${y}px`;
  node.style.scale = String(scale);
  node.style.opacity = String(t);
}

// dealIn is the Svelte transition applied to a card slot the first
// time it appears in the hand fan: card slides up from below + scales
// up + fades in, with a small overshoot at the end (back.out easing)
// so it reads as "dealt to the player" rather than "popped into
// existence." Drives the inner wrapper, not the .hand-slot itself,
// because the slot already owns the fan rotation transform — separate
// elements, separate transform stacks.
//
// Implementation note: Svelte transitions with a `tick` callback
// receive t in [0, 1] (0 at the start of an in-transition, 1 at the
// end). pose() is fire-and-forget here — Svelte owns the timing,
// pose just paints the per-frame values.
const DEAL_IN_DURATION = 360;
const DEAL_OUT_DURATION = 220;
export function dealIn(node: HTMLElement): TransitionConfig {
  return {
    duration: gatedDuration(DEAL_IN_DURATION, cfg.cardDraw),
    easing: backOut,
    tick: (t: number) => {
      const u = 1 - t;
      pose(node, t, u * 48, 1 - u * 0.18);
    },
  };
}

// dealOut is the partner transition used when a card leaves the hand
// — typically because the player cast it. Lifts upward + fades + a
// small scale-down so it reads as "off the hand" rather than vanishing.
// Played-card-to-battlefield travel is intentionally NOT modelled here
// because the new hand DOM node and the new battlefield DOM node are
// in different parents; a true cross-zone fly-in would need a portal /
// FLIP step. ADR 0119 §3 built that step for one case, a card leaving
// the stack (`flyTo` below, driven by StackLinger.svelte); the hand
// still has none.
export function dealOut(node: HTMLElement): TransitionConfig {
  return {
    // Played-from-hand uses cardPlay; pure draw discards (less
    // common) also flow through dealOut, but cardPlay is the more
    // visible action so it's the right gate to bind to.
    duration: gatedDuration(DEAL_OUT_DURATION, cfg.cardPlay),
    easing: cubicIn,
    tick: (t: number) => {
      const u = 1 - t;
      pose(node, t, -u * 32, 1 - u * 0.12);
    },
  };
}

// etbPulse is a Svelte action applied to a wrapper around each
// battlefield card. The action fires when the keyed each-block mounts
// a new DOM node — i.e. exactly when a fresh permanent enters the
// battlefield (or, on initial snapshot, when the existing board first
// renders). A Web Animation drives a one-shot scale punch with backOut overshoot
// so the card "pops" into existence rather than appearing flat.
//
// Action (vs. transition) chosen because there's no symmetric
// out-animation: a permanent leaving the battlefield doesn't get a
// dedicated effect (yet); the card just disappears as the next
// snapshot rebuilds the row. A future death-pulse could wrap this
// in `out:` if/when desired.
const ETB_DURATION = 0.32;
export const etbPulse: Action<HTMLElement> = (node) => {
  // Replace a punch still running, so a snapshot-driven re-mount
  // mid-animation cleanly restarts it instead of layering on top.
  cancelAnimations(node);
  // Skipping the punch leaves the permanent at its own style (scale 1,
  // opacity 1): the animation is the only thing that ever moves it off.
  if (cfg.enabled && cfg.particlesEtb) {
    run(
      node,
      [
        { scale: 0.55, opacity: 0 },
        { scale: 1, opacity: 1 },
      ],
      { duration: ETB_DURATION * cfg.speed * 1000, easing: backOutCss(2.2) },
    );
  }
  return { destroy: () => cancelAnimations(node) };
};

// floatUp / fadeOut drive the damage / heal popup over a player's
// identity disc. floatUp slides in from below + scales + fades;
// fadeOut drifts upward + fades, so the popup appears to lift off
// the player like a damage indicator in MTG Arena. Both are short —
// the popup auto-clears on a setTimeout in PlayerIdentity so it
// doesn't linger after the next change.
const POPUP_IN_DURATION = 220;
const POPUP_OUT_DURATION = 320;
export function floatUp(node: HTMLElement): TransitionConfig {
  return {
    duration: gatedDuration(POPUP_IN_DURATION, cfg.damagePopups),
    easing: backOut,
    tick: (t: number) => {
      const u = 1 - t;
      pose(node, t, u * 14, 0.7 + t * 0.3);
    },
  };
}
export function fadeOut(node: HTMLElement): TransitionConfig {
  return {
    duration: gatedDuration(POPUP_OUT_DURATION, cfg.damagePopups),
    easing: cubicOut,
    tick: (t: number) => {
      const u = 1 - t;
      pose(node, t, -u * 22, 1);
    },
  };
}

// flyTo is the FLIP step for a card leaving the stack (ADR 0119 §3).
// The caller draws a copy of the card where it was, in a layer over the
// whole board (the portal), and passes the translation and scale that
// put it on its landing (stackLinger.ts flightTransform). The copy
// travels there and fades as it lands. Gated by cardPlay under the
// master switch and scaled by speed, like every other card motion;
// with motion off it settles at once. Resolves when the flight is over.
export function flyTo(
  el: HTMLElement,
  to: { x: number; y: number; scale: number },
  baseMs: number,
): Promise<void> {
  const ms = gatedDuration(baseMs, cfg.cardPlay);
  // Replace a flight already under way; the old one's promise settles.
  cancelAnimations(el);
  el.style.transformOrigin = "0 0";
  // One keyframe: the flight starts from wherever the copy is.
  const kf = { translate: `${to.x}px ${to.y}px`, scale: String(to.scale), opacity: 0.35 };
  if (!canAnimate(el)) {
    Object.assign(el.style, { translate: kf.translate, scale: kf.scale, opacity: "0.35" });
    return Promise.resolve();
  }
  // forwards: the copy stays where it landed until its owner removes it.
  return run(el, [kf], { duration: ms, easing: EASE.power2InOut, fill: "forwards" }).done.then(
    () => undefined,
  );
}

// flyFrom is flyTo run backwards, for a card ARRIVING on the stack
// (#3381): the copy is drawn where it lands, starts at `from` (the
// translation and scale that put it on where it came from) and travels
// home, growing as it comes. Same gate and speed as flyTo. Resolves
// when the flight is over.
export function flyFrom(
  el: HTMLElement,
  from: { x: number; y: number; scale: number },
  baseMs: number,
): Promise<void> {
  const ms = gatedDuration(baseMs, cfg.cardPlay);
  cancelAnimations(el);
  el.style.transformOrigin = "0 0";
  if (!canAnimate(el)) return Promise.resolve();
  return run(
    el,
    [
      { translate: `${from.x}px ${from.y}px`, scale: String(from.scale), opacity: 0.5 },
      { translate: "0px 0px", scale: "1", opacity: 1 },
    ],
    { duration: ms, easing: EASE.power2InOut },
  ).done.then(() => undefined);
}

// ---- Combat strikes (ADR 0134 §2) ----
//
// An attacker's art-only copy lunges to contact with what it hit, holds
// there for the impact, and snaps back. The phases are at speed 1 and
// every one of them is multiplied by animations.speed; nothing here is
// reading time, so nothing is exempt (ADR 0134 §5).
export const STRIKE_OUT_MS = 180;
export const STRIKE_HOLD_MS = 60;
export const STRIKE_BACK_MS = 260;
export const STRIKE_MS = STRIKE_OUT_MS + STRIKE_HOLD_MS + STRIKE_BACK_MS;
// The impact: a horizontal shake of ±IMPACT_SHAKE_PX over
// IMPACT_SHAKE_MS, with a brightness flash that settles back over the
// same time.
export const IMPACT_SHAKE_MS = 160;
export const IMPACT_SHAKE_PX = 4;
// The flash's peak, as brightness above 1 (0.5 is brightness 1.5).
export const IMPACT_GLOW = 0.5;
// The lethal hit (ADR 0134 PR 2): what a creature that dies, or a
// player the hit eliminates, feels. A bigger, longer shake, and the
// flash is red instead of white.
export const IMPACT_LETHAL_SHAKE_PX = 8;
export const IMPACT_LETHAL_SHAKE_MS = 240;
export const IMPACT_LETHAL_WASH = "#e5484d";
// A creature that died crumbles (ADR 0134 question 7, PR 2): its copy
// breaks into shards that fall and fade over this long. A dead blocker
// crumbles after its lethal shake; a dead attacker once its copy is
// home.
export const STRIKE_CRUMBLE_MS = 420;
// Trample's streak (ADR 0134 question 5, PR 2): from the blockers to
// the player, planeswalker or battle that took the excess. It draws
// during the contact hold and fades after.
export const STREAK_DRAW_MS = STRIKE_HOLD_MS;
export const STREAK_FADE_MS = 220;

function strikeSpeed(): number {
  return Number.isFinite(cfg.speed) && cfg.speed > 0 ? cfg.speed : 1;
}

function strikeAllowed(): boolean {
  return cfg.enabled && cfg.combat;
}

// effectsAllowed gates the generic table effects (ADR 0209): the master
// switch and animations.effects. Reduced motion is the caller's, as for
// combat (accessibility.reduceMotion is never pushed in here).
export function effectsAllowed(): boolean {
  return cfg.enabled && cfg.effects !== "off";
}

// lunge flies a strike copy `to` (board pixels, relative to where it is
// drawn) and back: out with power2.in, a contact hold, back with
// power3.out. `flash` is for an attacker its blocker hit back in the
// same beat: that element brightens at contact instead of shaking, so
// the two motions do not fight. `lethal` makes that flash the lethal
// one: red, and as long as the lethal shake. A creature that died no
// longer fades on the way back: it crumbles once it is home (crumble,
// below). With combat motion off it settles at once. Resolves when the
// copy is home.
export function lunge(
  el: HTMLElement,
  to: { x: number; y: number },
  opts: { flash?: HTMLElement | null; lethal?: boolean } = {},
): Promise<void> {
  if (!strikeAllowed()) {
    return Promise.resolve();
  }
  const s = strikeSpeed();
  const at = `${to.x}px ${to.y}px`;
  const home = "0px 0px";
  // One animation, four keyframes: out (power2.in), the contact hold,
  // back (power3.out). A keyframe's easing shapes the stretch after it.
  const flight = run(
    el,
    [
      { translate: home, offset: 0, easing: EASE.power2In },
      { translate: at, offset: STRIKE_OUT_MS / STRIKE_MS, easing: EASE.linear },
      {
        translate: at,
        offset: (STRIKE_OUT_MS + STRIKE_HOLD_MS) / STRIKE_MS,
        easing: EASE.power3Out,
      },
      { translate: home, offset: 1 },
    ],
    { duration: STRIKE_MS * s },
  );
  const parts = [flight.done];
  if (opts.flash) {
    if (opts.lethal) opts.flash.style.setProperty("--impact-wash", IMPACT_LETHAL_WASH);
    // At contact (the end of the out phase) the flash starts and settles.
    parts.push(
      run(opts.flash, [{ "--impact-glow": IMPACT_GLOW }, { "--impact-glow": 0 }], {
        delay: STRIKE_OUT_MS * s,
        duration: (opts.lethal ? IMPACT_LETHAL_SHAKE_MS : IMPACT_SHAKE_MS) * s,
        easing: EASE.power1Out,
      }).done,
    );
  }
  return Promise.all(parts).then(() => undefined);
}

// impactShake is the hit landing on a tile or an avatar disc: --impact-x
// swings in three half-cycles and --impact-glow flashes and settles.
// Card.svelte and PlayerIdentity.svelte compose both variables with
// defaults that do nothing, so a tile at rest renders exactly as
// before. An ordinary hit is ±IMPACT_SHAKE_PX over IMPACT_SHAKE_MS with
// a white flash. A lethal hit (a creature that dies, a player it
// eliminates) is ±IMPACT_LETHAL_SHAKE_PX over IMPACT_LETHAL_SHAKE_MS,
// and --impact-wash turns its flash red. Resolves when it is over.
export function impactShake(el: HTMLElement, opts: { lethal?: boolean } = {}): Promise<void> {
  if (!strikeAllowed()) return Promise.resolve();
  const lethal = opts.lethal === true;
  const t = (lethal ? IMPACT_LETHAL_SHAKE_MS : IMPACT_SHAKE_MS) * strikeSpeed();
  const px = lethal ? IMPACT_LETHAL_SHAKE_PX : IMPACT_SHAKE_PX;
  if (lethal) el.style.setProperty("--impact-wash", IMPACT_LETHAL_WASH);
  // --impact-x is pointed at the registered, animated --impact-shake
  // for the length of the shake and cleared after: an unset --impact-x
  // is what leaves the avatar disc's `translate` at none
  // (PlayerIdentity.svelte), so it must not be registered or left set.
  el.style.setProperty("--impact-x", "var(--impact-shake)");
  const shake = run(
    el,
    [
      { "--impact-shake": "0px", offset: 0, easing: EASE.sineInOut },
      { "--impact-shake": `${px}px`, offset: 1 / 6, easing: EASE.sineInOut },
      { "--impact-shake": `${-px}px`, offset: 1 / 2, easing: EASE.sineInOut },
      { "--impact-shake": `${px}px`, offset: 5 / 6, easing: EASE.sineInOut },
      { "--impact-shake": "0px", offset: 1 },
    ],
    { duration: t },
  );
  const glow = run(el, [{ "--impact-glow": IMPACT_GLOW }, { "--impact-glow": 0 }], {
    duration: t,
    easing: EASE.power1Out,
  });
  return Promise.all([shake.done, glow.done]).then(() => {
    el.style.removeProperty("--impact-x");
    el.style.removeProperty("--impact-wash");
  });
}

// One shard of a crumble, as combatStrikes.ts crumbleShards plans it.
export interface CrumbleShardMotion {
  // The shard's outline: a CSS clip-path polygon over the face.
  clip: string;
  // Where it falls to, in the face's own (unrotated) pixels, its spin
  // in degrees, and when it lets go, as a share of the crumble.
  dx: number;
  dy: number;
  rot: number;
  delay: number;
}

// crumble breaks a dead creature's copy into shards that fall and fade
// (ADR 0134 PR 2). `face` is the copy's face: each shard is the face's
// image clipped to one piece, and the face's own art and frame are
// hidden under them (the `crumbling` class). The shards are made here,
// inside the copy, and go when the copy does. Resolves when they are
// gone.
export function crumble(
  face: HTMLElement,
  shards: readonly CrumbleShardMotion[],
  // The table effects (ADR 0209 §3) reuse the crumble for a death by
  // state-based action: they are gated by animations.effects, not
  // animations.combat, and pass their own (capped, scaled) length, and
  // their delay: a card in a board wipe or a signature's batch holds whole
  // and still until its own turn comes (#4416).
  opts: { gate?: "combat" | "effects"; ms?: number; delay?: number } = {},
): Promise<void> {
  if (!(opts.gate === "effects" ? effectsAllowed() : strikeAllowed())) return Promise.resolve();
  const total = opts.ms ?? STRIKE_CRUMBLE_MS * strikeSpeed();
  const wait = Math.max(0, opts.delay ?? 0);
  const src = face.querySelector("img")?.getAttribute("src") ?? null;
  const made: HTMLElement[] = [];
  for (const sh of shards) {
    const el = document.createElement("div");
    el.className = "shard";
    el.dataset.strikeShard = "";
    el.style.clipPath = sh.clip;
    if (src) {
      const img = document.createElement("img");
      img.src = src;
      img.alt = "";
      img.draggable = false;
      el.appendChild(img);
    }
    face.appendChild(el);
    made.push(el);
  }
  face.classList.add("crumbling");
  // Each shard lets go at its own moment and all land together at the
  // end of the crumble. forwards: a fallen shard stays gone.
  return Promise.all(
    made.map((el, i) => {
      const sh = shards[i];
      const at = sh.delay * total;
      return run(
        el,
        [
          { translate: "0px 0px", rotate: "0deg", opacity: 1 },
          { translate: `${sh.dx}px ${sh.dy}px`, rotate: `${sh.rot}deg`, opacity: 0 },
        ],
        {
          delay: wait + at,
          duration: Math.max(1, total - at),
          easing: EASE.power2In,
          fill: "forwards",
        },
      ).done;
    }),
  ).then(() => undefined);
}

// streak draws trample's streak: `el` is a bar laid along the segment,
// anchored at its start, scaled 0 → 1 along x during the contact hold,
// then faded. Resolves when it is gone.
export function streak(el: HTMLElement): Promise<void> {
  if (!strikeAllowed()) return Promise.resolve();
  const s = strikeSpeed();
  // Draw (scale 0 → 1 along x, power1.out) then fade (power1.in), as
  // one animation: a keyframe's easing shapes the stretch after it.
  const draw = STREAK_DRAW_MS / (STREAK_DRAW_MS + STREAK_FADE_MS);
  return run(
    el,
    [
      { scale: "0 1", opacity: 1, offset: 0, easing: EASE.power1Out },
      { scale: "1 1", opacity: 1, offset: draw, easing: EASE.power1In },
      { scale: "1 1", opacity: 0, offset: 1 },
    ],
    { duration: (STREAK_DRAW_MS + STREAK_FADE_MS) * s, fill: "forwards" },
  ).done.then(() => undefined);
}
