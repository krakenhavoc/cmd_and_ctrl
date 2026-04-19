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
