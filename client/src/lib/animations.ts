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
  return gsap.to(el, {
    "--tap-rot": tapped ? "90deg" : "0deg",
    duration: TAP_DURATION,
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
    duration: DEAL_IN_DURATION,
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
    duration: DEAL_OUT_DURATION,
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
  gsap.fromTo(
    node,
    { scale: 0.55, opacity: 0 },
    {
      scale: 1,
      opacity: 1,
      duration: ETB_DURATION,
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
    duration: POPUP_IN_DURATION,
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
    duration: POPUP_OUT_DURATION,
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
