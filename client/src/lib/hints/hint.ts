// hint.ts — what a first-use hint is (ADR 0125 §3.1).
//
// A hint is one small, dismissible card that says where something is
// and how to use it, the first time someone meets it. Each lives in its
// own `<Feature>.hint.ts` beside the component that owns the feature
// (§3.2), with a `default` export of one `Hint`; lib/hints/index.ts
// collects them. Nothing here touches the DOM.

import type { Anchor, Copy } from "../tutorial";
import type { GameView } from "../protocol";
import type { TableMoment } from "./tableMoment";

/**
 * Where a hint is offered. A route's place ("lobby", "decks", …),
 * "settings" (inside the Settings dialog, through its HintSlot), "site"
 * (any site page, after that page's own hints) or "table" (a game, in a
 * quiet moment only).
 */
export const HINT_PLACES = [
  "site",
  "settings",
  "table",
  "home",
  "lobby",
  "decks",
  "catalog",
  "roadmap",
  "admin",
  "my-games",
] as const;

export type HintPlace = (typeof HINT_PLACES)[number];

/** A site page's place: everything but the dialog, the table and "site" itself. */
export type PagePlace = Exclude<HintPlace, "site" | "settings" | "table">;

/** `<place>.<feature>`, e.g. "table.stack". Stable; never reused (retired.ts). */
export type HintID = `${HintPlace}.${string}`;

/** What a hint's `anchor` and `when` may read, and nothing else (§3.1). */
export interface HintContext {
  /** The place being offered hints: the page, the Settings dialog or the table. */
  place: HintPlace;
  /** The route's name (router.ts), for a hint that cares about more than its place. */
  route: string;
  /** A Discord-signed-in person (a session with a user). */
  signedIn: boolean;
  /** An allowlisted person with admin mode on (ADR 0112 §2). */
  adminMode: boolean;
  /** The shared admin token (no person behind it). */
  adminToken: boolean;
  /**
   * Whether the person has a finished game, or null while unknown. Read
   * by the tutorial offer (ADR 0125 §3.7, `lobby.practice`).
   */
  hasEndedGame: boolean | null;
  /** At the table: the moment (§3.5). Null everywhere else. */
  moment: TableMoment | null;
  /** At the table: the game as the viewer sees it. Null everywhere else. */
  view: GameView | null;
  /** At the table: the viewer's player id, or null for a spectator. */
  viewerID: string | null;
}

export interface Hint {
  /** `<place>.<feature>`; its place prefix is `place`. Stable; never reused. */
  id: HintID;
  /** Starts at 1. Bumped only when a returning player would now be wrong (§3.3). */
  version: number;
  place: HintPlace;
  /** Which of a place's hints comes first, lowest first. */
  order: number;
  /** A contract label (labels.ts), a card or a seat; or a function of the context. */
  anchor: Anchor | ((c: HintContext) => Anchor | null);
  /** At most 32 characters (§3.4). */
  title: Copy;
  /** At most 140 characters, one or two sentences, ending in a full stop (§3.4). */
  body: Copy;
  /** Extra conditions: signed in, admin mode, three or more seats, … */
  when?: (c: HintContext) => boolean;
  /** One link button. With one, "Got it" reads "Not now". */
  action?: { label: string; href: string };
  /**
   * The anchor is always mounted but has no area until its feature has
   * something to show (the attention strip while it is empty). The hint
   * waits for it to show and, unlike a hint whose anchor is gone, its
   * absence is not logged as "has no anchor".
   */
  waitsForAnchor?: boolean;
}

/** anchorOf is the hint's anchor in this context, or null when it names none. */
export function anchorOf(h: Hint, c: HintContext): Anchor | null {
  return typeof h.anchor === "function" ? h.anchor(c) : h.anchor;
}

/** holds reports whether a hint's `when` allows it here; a `when` that throws does not. */
export function holds(h: Hint, c: HintContext): boolean {
  if (!h.when) return true;
  try {
    return h.when(c);
  } catch {
    return false;
  }
}

/** emptyContext is a context with nothing known, for a place. Tests and defaults. */
export function emptyContext(place: HintPlace, route: string = place): HintContext {
  return {
    place,
    route,
    signedIn: false,
    adminMode: false,
    adminToken: false,
    hasEndedGame: null,
    moment: null,
    view: null,
    viewerID: null,
  };
}

/**
 * placeOfRoute is the hint place a route offers, or null for a route
 * that offers none: sign-in, the invite and reclaim doors, the practice
 * door and the OAuth hand-off each do one thing and say so (§3.8), and
 * the game route is the table.
 */
export function placeOfRoute(name: string): HintPlace | null {
  switch (name) {
    case "home":
    case "lobby":
    case "decks":
    case "catalog":
    case "roadmap":
      return name;
    case "adminViews":
      return "admin";
    case "myGames":
      return "my-games";
    case "game":
      return "table";
    default:
      return null;
  }
}
