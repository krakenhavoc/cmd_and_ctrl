// S09 polish-pass animation helpers, all backed by GSAP.
//
// The Card.svelte component composes its `transform` from CSS custom
// properties so multiple effects can stack without fighting each
// other:
//
//   transform: rotate(var(--tap-rot, 0deg)) translateY(var(--hover-lift, 0px))
//
// CSS owns the static state (hover lift, base rotation). GSAP drives
// the animated transitions by tweening the relevant custom property,
// so a CSS hover and an in-flight tap animation compose cleanly. This
// also means a reduced-motion fallback can short-circuit the GSAP
// call and just snap the var to its target value.

import { gsap } from "gsap";
import { backOut, cubicIn, cubicOut } from "svelte/easing";
import type { TransitionConfig } from "svelte/transition";
import type { Action } from "svelte/action";

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
// by tweening the --tap-rot CSS var. Idempotent — calling it with the
// already-current state just no-ops the in-flight tween. Returns the
// gsap tween so callers can chain or kill if needed; most callers can
// ignore the return value.
export function animateTap(el: HTMLElement, tapped: boolean): gsap.core.Tween {
  // Per-effect gate distinguishes tap (cardTap) from untap
  // (cardUntap) so a user who likes the tap motion but finds untap-
  // all visually noisy at end-of-turn can disable just that side.
  const allowed = tapped ? cfg.cardTap : cfg.cardUntap;
  const dur = cfg.enabled && allowed ? TAP_DURATION * cfg.speed : 0;
  return gsap.to(el, {
    "--tap-rot": tapped ? "90deg" : "0deg",
    duration: dur,
    ease: "power2.out",
    overwrite: "auto",
  });
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
// end). gsap.set is fire-and-forget here — Svelte owns the timing,
// gsap just paints the per-frame values.
const DEAL_IN_DURATION = 360;
const DEAL_OUT_DURATION = 220;
export function dealIn(node: HTMLElement): TransitionConfig {
  return {
    duration: gatedDuration(DEAL_IN_DURATION, cfg.cardDraw),
    easing: backOut,
    tick: (t: number) => {
      const u = 1 - t;
      gsap.set(node, {
        y: u * 48,
        scale: 1 - u * 0.18,
        opacity: t,
      });
    },
  };
}

// dealOut is the partner transition used when a card leaves the hand
// — typically because the player cast it. Lifts upward + fades + a
// small scale-down so it reads as "off the hand" rather than vanishing.
// Played-card-to-battlefield travel is intentionally NOT modelled here
// because the new hand DOM node and the new battlefield DOM node are
// in different parents; a true cross-zone fly-in would need a portal /
// FLIP step. Saving that for a follow-up if it's missed.
export function dealOut(node: HTMLElement): TransitionConfig {
  return {
    // Played-from-hand uses cardPlay; pure draw discards (less
    // common) also flow through dealOut, but cardPlay is the more
    // visible action so it's the right gate to bind to.
    duration: gatedDuration(DEAL_OUT_DURATION, cfg.cardPlay),
    easing: cubicIn,
    tick: (t: number) => {
      const u = 1 - t;
      gsap.set(node, {
        y: -u * 32,
        scale: 1 - u * 0.12,
        opacity: t,
      });
    },
  };
}

// etbPulse is a Svelte action applied to a wrapper around each
// battlefield card. The action fires when the keyed each-block mounts
// a new DOM node — i.e. exactly when a fresh permanent enters the
// battlefield (or, on initial snapshot, when the existing board first
// renders). gsap drives a one-shot scale punch with backOut overshoot
// so the card "pops" into existence rather than appearing flat.
//
// Action (vs. transition) chosen because there's no symmetric
// out-animation: a permanent leaving the battlefield doesn't get a
// dedicated effect (yet); the card just disappears as the next
// snapshot rebuilds the row. A future death-pulse could wrap this
// in `out:` if/when desired.
const ETB_DURATION = 0.32;
export const etbPulse: Action<HTMLElement> = (node) => {
  if (!cfg.enabled || !cfg.particlesEtb) {
    // Skip the punch — but still snap to the final state so the
    // permanent doesn't render at scale 0.55 / opacity 0.
    gsap.set(node, { scale: 1, opacity: 1 });
    return;
  }
  gsap.fromTo(
    node,
    { scale: 0.55, opacity: 0 },
    {
      scale: 1,
      opacity: 1,
      duration: ETB_DURATION * cfg.speed,
      ease: "back.out(2.2)",
      // overwrite so a snapshot-driven re-mount mid-animation cleanly
      // restarts the punch instead of layering on top.
      overwrite: "auto",
    },
  );
};

// floatUp / fadeOut drive the damage / heal popup over a player's
// header. floatUp slides in from below + scales + fades; fadeOut
// drifts upward + fades, so the popup appears to lift off the player
// like a damage indicator in MTG Arena. Both are short — the popup
// auto-clears on a setTimeout in PlayerHeader so it doesn't linger
// after the next change.
const POPUP_IN_DURATION = 220;
const POPUP_OUT_DURATION = 320;
export function floatUp(node: HTMLElement): TransitionConfig {
  return {
    duration: gatedDuration(POPUP_IN_DURATION, cfg.damagePopups),
    easing: backOut,
    tick: (t: number) => {
      const u = 1 - t;
      gsap.set(node, {
        y: u * 14,
        scale: 0.7 + t * 0.3,
        opacity: t,
      });
    },
  };
}
export function fadeOut(node: HTMLElement): TransitionConfig {
  return {
    duration: gatedDuration(POPUP_OUT_DURATION, cfg.damagePopups),
    easing: cubicOut,
    tick: (t: number) => {
      const u = 1 - t;
      gsap.set(node, {
        y: -u * 22,
        opacity: t,
      });
    },
  };
}
