// A thin layer over the Web Animations API (element.animate), which
// replaced GSAP (#3076; C2 is AGPL and GSAP's license is not free).
//
// It carries the three things every caller needs and element.animate
// does not give on its own:
//
//   - the named easings the old GSAP calls used, as CSS easing
//     strings (so "power2.out" keeps its curve);
//   - a registry per element, so a caller that tears down can cancel
//     whatever it started (the job gsap.killTweensOf did), and so a
//     new animation can replace one still running on the same
//     element (gsap's overwrite: "auto");
//   - a promise that always settles, finished or cancelled, because
//     the sequencing code waits on it the way it waited on a gsap
//     onComplete / onInterrupt pair.
//
// Where the environment has no element.animate (jsdom, an ancient
// browser) nothing is animated and the promise resolves at once: the
// callers set their end state themselves, so that is a snap, the same
// outcome a disabled-motion setting produces.

// GSAP's power1 / power2 / power3 are the quad / cubic / quart Penner
// curves; these are their standard cubic-bezier forms. sine.inOut is
// easeInOutSine.
export const EASE = {
  power1In: "cubic-bezier(0.55, 0.085, 0.68, 0.53)",
  power1Out: "cubic-bezier(0.25, 0.46, 0.45, 0.94)",
  power2In: "cubic-bezier(0.55, 0.055, 0.675, 0.19)",
  power2Out: "cubic-bezier(0.215, 0.61, 0.355, 1)",
  power2InOut: "cubic-bezier(0.645, 0.045, 0.355, 1)",
  power3Out: "cubic-bezier(0.165, 0.84, 0.44, 1)",
  sineInOut: "cubic-bezier(0.445, 0.05, 0.55, 0.95)",
  linear: "linear",
} as const;

// backOut is GSAP's back.out(overshoot): t' = (t-1)^2 ((s+1)(t-1) + s) + 1.
// A cubic-bezier cannot hold an overshoot of 2.2, so it is sampled into
// a CSS linear() easing. A browser without linear() gets the stock
// ease-out-back bezier (overshoot about 1.7), which is close.
export function backOut(overshoot: number, samples = 32): string {
  const supported =
    typeof CSS !== "undefined" &&
    typeof CSS.supports === "function" &&
    CSS.supports("animation-timing-function", "linear(0, 1)");
  if (!supported) return "cubic-bezier(0.34, 1.56, 0.64, 1)";
  const pts: string[] = [];
  for (let i = 0; i <= samples; i++) {
    const t = i / samples - 1;
    const v = t * t * ((overshoot + 1) * t + overshoot) + 1;
    pts.push(String(Math.round(v * 1000) / 1000));
  }
  return `linear(${pts.join(", ")})`;
}

export interface Run {
  // Stops the animation where it stands. Its promise settles with false.
  cancel(): void;
  // true when it played out, false when it was cancelled.
  done: Promise<boolean>;
}

const live = new WeakMap<Element, Set<Animation>>();

function track(el: Element, a: Animation): void {
  let set = live.get(el);
  if (!set) live.set(el, (set = new Set()));
  set.add(a);
}

function untrack(el: Element, a: Animation): void {
  const set = live.get(el);
  if (!set) return;
  set.delete(a);
  if (set.size === 0) live.delete(el);
}

// cancelAnimations stops every animation run() started on `el`: the
// teardown call, and the "replace what is running" call.
export function cancelAnimations(el: Element): void {
  const set = live.get(el);
  if (!set) return;
  for (const a of [...set]) a.cancel();
}

export function canAnimate(el: Element): boolean {
  return typeof (el as { animate?: unknown }).animate === "function";
}

// settled resolves once every finite animation running on `el` or on
// one of its ancestors has finished or been cancelled. A layout box read
// while one of them is mid-flight carries its transform (an entrance pop
// scales the card down), so a caller that measures `el` waits on this
// first. Endless animations are skipped: they would never settle. Where
// the environment has no document.getAnimations it resolves at once.
export function settled(el: Element): Promise<void> {
  if (typeof document === "undefined" || typeof document.getAnimations !== "function") {
    return Promise.resolve();
  }
  const waits: Promise<unknown>[] = [];
  for (const a of document.getAnimations()) {
    const target = (a.effect as KeyframeEffect | null)?.target;
    if (!(target instanceof Element) || !target.contains(el)) continue;
    if (!Number.isFinite(a.effect?.getComputedTiming().endTime ?? Infinity)) continue;
    waits.push(a.finished.catch(() => undefined));
  }
  return Promise.all(waits).then(() => undefined);
}

// run starts one animation on `el`. `fill: "forwards"` is up to the
// caller: an animation that has to leave `el` in its end state passes
// it (the element is removed by its owner), one that returns to the
// element's own style does not.
export function run(el: Element, keyframes: Keyframe[], options: KeyframeAnimationOptions): Run {
  if (!canAnimate(el)) return { cancel() {}, done: Promise.resolve(true) };
  const a = el.animate(keyframes, options);
  track(el, a);
  const done = new Promise<boolean>((resolve) => {
    a.onfinish = () => {
      untrack(el, a);
      resolve(true);
    };
    a.oncancel = () => {
      untrack(el, a);
      resolve(false);
    };
  });
  return { cancel: () => a.cancel(), done };
}
