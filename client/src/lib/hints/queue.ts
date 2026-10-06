// queue.ts — which hint shows, and when (ADR 0125 §3.5).
//
// One hint on screen at a time, everywhere; a second unseen hint waits.
//
// Site pages (and the Settings dialog, which is its own "visit" each
// time it opens): at most one hint per visit. It is the first unseen
// hint for the place (then, on a page, the "site" hints) in `order`
// whose `when` holds and whose anchor resolves within SITE_WINDOW_MS of
// the visit starting. A hint whose anchor does not resolve is skipped
// for the visit, stays unseen, and logs `hint: <id> has no anchor on the
// page`. A hint never makes the page wait.
//
// The table: a hint shows only in a quiet moment (tableMoment.ts). If
// the moment stops being quiet the hint steps aside at once and stays
// unseen, and comes back at the next quiet moment.
//
// "Hide tips" (settings.help.tipsOff) stops every offer. A replay (the
// Help menu's "Tips for this page", ADR 0125 §6) shows the place's hints
// one after another, ignoring tipsOff and the one-per-visit rule,
// because the person asked.
//
// Nothing here touches the DOM or a store: HintLayer.svelte feeds the
// engine the context, the seen map and a `resolves` function every poll,
// and draws whatever it answers.

import { holds, type Hint, type HintContext, type HintPlace } from "./hint";
import { isUnseen, type SeenMap } from "./seen";
import { isQuiet } from "./tableMoment";

/** How long after a visit starts a site hint's anchor may take to appear. */
export const SITE_WINDOW_MS = 3_000;

export interface Offer {
  hints: readonly Hint[];
  ctx: HintContext;
  seen: SeenMap;
  tipsOff: boolean;
  /** A replay in progress for this place: tipsOff does not apply to these. */
  replay?: ReadonlySet<string> | null;
  /** Offer only the replay's hints (a site page's replay shows just those). */
  only?: boolean;
}

function byOrder(a: Hint, b: Hint): number {
  return a.order - b.order || (a.id < b.id ? -1 : a.id > b.id ? 1 : 0);
}

/** placesFor is where a place's hints come from, strongest first. */
export function placesFor(place: HintPlace): HintPlace[] {
  if (place === "table" || place === "settings" || place === "site") return [place];
  return [place, "site"];
}

/**
 * candidates is every hint that could be offered here now, in the order
 * they would be: unseen, allowed by `when`, and not held back by
 * tipsOff (a replay's own hints ignore it). An anchor is not checked.
 */
export function candidates(o: Offer): Hint[] {
  const out: Hint[] = [];
  for (const place of placesFor(o.ctx.place)) {
    const here = o.hints.filter((h) => h.place === place).sort(byOrder);
    for (const h of here) {
      const replayed = o.replay?.has(h.id) === true;
      if (o.only && !replayed) continue;
      if (o.tipsOff && !replayed) continue;
      if (!isUnseen(o.seen, h.id, h.version)) continue;
      if (!holds(h, o.ctx)) continue;
      out.push(h);
    }
  }
  return out;
}

export interface SitePick {
  /** The hint to show, or null. */
  hint: Hint | null;
  /** Hints passed over because their anchor never came: log them. */
  missing: Hint[];
  /** True once nothing more can come of this visit. */
  settled: boolean;
}

/**
 * pickSiteHint applies the site rule at `elapsed` ms into a visit. The
 * first candidate is shown as soon as its anchor resolves; a later one
 * only once every earlier one has had the whole window and not come.
 */
export function pickSiteHint(
  cands: readonly Hint[],
  resolves: (h: Hint) => boolean,
  elapsed: number,
  windowMs: number = SITE_WINDOW_MS,
): SitePick {
  const missing: Hint[] = [];
  for (const h of cands) {
    if (resolves(h)) return { hint: h, missing, settled: true };
    if (elapsed < windowMs) return { hint: null, missing: [], settled: false };
    missing.push(h);
  }
  return { hint: null, missing, settled: elapsed >= windowMs || cands.length === 0 };
}

export interface EngineEnv {
  hints: readonly Hint[];
  ctx: HintContext;
  /**
   * Which visit this is: a new value is a new visit (a route change,
   * the Settings dialog opening). Returning to an earlier visit's key
   * resumes it, so closing Settings does not give the page a second hint.
   */
  visit: string;
  seen: SeenMap;
  tipsOff: boolean;
  now: number;
  /** Whether a hint's anchor is on the page with some area right now. */
  resolves: (h: Hint) => boolean;
}

export interface HintEngine {
  /** One poll: the hint to draw now, or null. */
  tick(env: EngineEnv): Hint | null;
  /**
   * The hint on screen closes ("Got it", Escape, "Hide tips"). Returns
   * it, so the caller can mark it seen, or null if none was up.
   * `onScreen` is the hint the card shows. It is the one dismissed when
   * the visit no longer holds one (it stepped aside after it was
   * drawn), so a card someone closed never comes straight back.
   */
  dismiss(now: number, onScreen?: Hint | null): Hint | null;
  /**
   * Show these hints of `place` again, one after another (the Help
   * menu). The caller clears them from the seen map first.
   */
  replay(place: HintPlace, ids: readonly string[]): void;
  /** The hint the current visit holds, drawn or not. */
  current(): Hint | null;
}

export interface EngineOptions {
  log?: (msg: string) => void;
  windowMs?: number;
}

interface Visit {
  start: number;
  done: boolean;
  logged: Set<string>;
  shown: Hint | null;
}

const MAX_VISITS = 16;

/** missingAnchorMessage is the console line for a hint whose anchor never came. */
export function missingAnchorMessage(id: string): string {
  return `hint: ${id} has no anchor on the page`;
}

export function createHintEngine(opts: EngineOptions = {}): HintEngine {
  const log = opts.log ?? ((m: string) => console.warn(m));
  const windowMs = opts.windowMs ?? SITE_WINDOW_MS;
  const visits = new Map<string, Visit>();
  let key: string | null = null;
  let lastClosedAt: number | null = null;
  let replay: { place: HintPlace; ids: Set<string>; restart: boolean } | null = null;

  const visit = (): Visit | null => (key === null ? null : (visits.get(key) ?? null));

  function enter(env: EngineEnv): Visit {
    if (key !== env.visit) {
      key = env.visit;
      if (replay && replay.place !== env.ctx.place) replay = null;
    }
    let v = visits.get(env.visit);
    if (!v) {
      v = { start: env.now, done: false, logged: new Set(), shown: null };
      visits.set(env.visit, v);
      while (visits.size > MAX_VISITS) visits.delete(visits.keys().next().value as string);
    }
    if (replay?.restart && replay.place === env.ctx.place) {
      replay.restart = false;
      v.start = env.now;
      v.done = false;
      v.shown = null;
    }
    return v;
  }

  function logOnce(v: Visit, h: Hint): void {
    // An anchor that is empty by design is not missing (Hint.waitsForAnchor).
    if (h.waitsForAnchor) return;
    if (v.logged.has(h.id)) return;
    v.logged.add(h.id);
    log(missingAnchorMessage(h.id));
  }

  return {
    tick(env) {
      const v = enter(env);
      const replayIDs = replay && replay.place === env.ctx.place ? replay.ids : null;
      const cands = candidates({
        hints: env.hints,
        ctx: env.ctx,
        seen: env.seen,
        tipsOff: env.tipsOff,
        replay: replayIDs,
        only: replayIDs !== null && env.ctx.place !== "table",
      });
      // Seen in another tab, or no longer allowed: it goes.
      if (v.shown && !cands.some((h) => h.id === v.shown!.id)) v.shown = null;

      if (env.ctx.place === "table") {
        const m = env.ctx.moment;
        if (!m || !isQuiet(m, env.now, lastClosedAt)) {
          // Steps aside at once and stays unseen.
          v.shown = null;
          return null;
        }
        if (v.shown && env.resolves(v.shown)) return v.shown;
        v.shown = null;
        for (const h of cands) {
          if (env.resolves(h)) {
            v.shown = h;
            return h;
          }
          logOnce(v, h);
        }
        return null;
      }

      // A site page or the Settings dialog: the visit's one hint. While
      // its anchor is gone the card is not drawn; it is still the visit's.
      if (v.shown) return env.resolves(v.shown) ? v.shown : null;
      if (v.done) return null;
      const pick = pickSiteHint(cands, env.resolves, env.now - v.start, windowMs);
      for (const h of pick.missing) {
        logOnce(v, h);
        replayIDs?.delete(h.id);
      }
      if (pick.hint) {
        v.shown = pick.hint;
        return pick.hint;
      }
      if (pick.settled) {
        v.done = true;
        if (replayIDs) replay = null;
      }
      return null;
    },

    dismiss(now, onScreen = null) {
      const v = visit();
      const h = v?.shown ?? onScreen;
      if (!h) return null;
      lastClosedAt = now;
      // A replay goes on to its next hint; anything else ends the visit.
      let visitDone = true;
      if (replay?.ids.has(h.id)) {
        replay.ids.delete(h.id);
        if (replay.ids.size === 0) replay = null;
        else visitDone = false;
      }
      // Only the visit that held the hint is spent by closing it.
      if (v?.shown) {
        v.shown = null;
        if (visitDone) v.done = true;
      }
      return h;
    },

    replay(place, ids) {
      replay = ids.length > 0 ? { place, ids: new Set(ids), restart: true } : null;
    },

    current() {
      return visit()?.shown ?? null;
    },
  };
}
