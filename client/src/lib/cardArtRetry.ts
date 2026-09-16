// cardArtRetry.ts — what happens when a card-art <img> fails (#33).
//
// `GET /cards/{id}/image` makes exactly one upstream request to the
// Scryfall CDN and answers any failure with a JSON 502. An <img>
// pointed at that fires `error` and paints nothing, permanently, with
// no signal to the player and no way back short of a reload. A cold
// cache on launch is exactly when the CDN hiccups, so one transient
// failure was costing a tile its art for the whole session.
//
// This module is the DOM-free half: the retry state machine and the
// deduplicated report. The DOM half — listeners, the marker element,
// re-requesting the image — is the `cardArt` action in cardArt.ts.
// The split exists because there is no Svelte component harness in
// this repo, and every decision worth testing lives here.
//
// The retry re-requests the PLAIN URL. No cache-busting parameter:
// nothing caches the failure (the server sets Cache-Control only on
// success, the service worker refuses to store a non-200, and the
// disk cache renames into place only after a clean copy), while the
// service worker keys card art on the full query string — a busted
// URL that succeeded would sit in the size-trimmed cache under a key
// nothing ever requests again.

import { recordClientError } from "./clientErrors";

// ART_RETRY_DELAY_MS is the wait before the one automatic retry. Long
// enough to ride out a CDN blip, short enough that the player has not
// yet decided the tile is broken.
export const ART_RETRY_DELAY_MS = 2000;

// ART_FAILED_TITLE is the marker's tooltip and accessible name. Not
// the card name: every tile already shows that on hover.
export const ART_FAILED_TITLE = "Art failed to load";

// ArtState walks one URL through its attempts:
//
//   loading ──error──▶ scheduled ──2 s──▶ retrying ──error──▶ failed
//                                                               │  ▲
//                                                          retry()  error
//                                                               ▼  │
//                                                        manual-retrying
//
// `load` from any state lands on `loaded`. Only an `error` while
// loading gets the automatic retry; every later failure waits for a
// click.
export type ArtState =
  | "loading"
  | "scheduled"
  | "retrying"
  | "failed"
  | "manual-retrying"
  | "loaded";

// MarkerState is what the tile shows for an ArtState. The marker is
// hidden until the automatic retry has also failed: two seconds of
// blank tile is ordinary loading, and a pip that appears and vanishes
// on a blip is noise. "busy" keeps it in place while a click-retry is
// in flight, so a fast second failure doesn't flicker it away and back.
export type MarkerState = "hidden" | "ready" | "busy";

export function markerFor(state: ArtState): MarkerState {
  if (state === "failed") return "ready";
  if (state === "manual-retrying") return "busy";
  return "hidden";
}

export interface ArtRetryHooks {
  // reload re-requests the current URL. Called for the automatic retry
  // and for every click-retry.
  reload: (url: string) => void;
  // render is called after every state change.
  render: (state: ArtState) => void;
  // report records a failure the player can see. Defaults to
  // reportArtFailure; injectable for tests.
  report?: (url: string) => void;
  delayMs?: number;
}

export interface ArtRetry {
  readonly state: ArtState;
  readonly url: string;
  // error / load feed the <img> events in.
  error(): void;
  load(): void;
  // retry is the marker's click. A no-op unless the art has failed.
  retry(): void;
  // setURL resets to a fresh `loading` for a different URL — a DFC
  // transforming, a face flip — and cancels any pending retry. The
  // same URL is a no-op, so a re-render cannot restart a retry cycle.
  setURL(url: string): void;
  // destroy cancels the pending retry and ignores every later event.
  // Hand cards mount and unmount constantly; a timer outliving its
  // <img> would re-request art for a card that has left the hand.
  destroy(): void;
}

export function createArtRetry(initialURL: string, hooks: ArtRetryHooks): ArtRetry {
  const delayMs = hooks.delayMs ?? ART_RETRY_DELAY_MS;
  const report = hooks.report ?? reportArtFailure;
  let url = initialURL;
  let state: ArtState = "loading";
  let timer: ReturnType<typeof setTimeout> | null = null;
  let destroyed = false;

  function cancelTimer(): void {
    if (timer !== null) {
      clearTimeout(timer);
      timer = null;
    }
  }

  function set(next: ArtState): void {
    state = next;
    hooks.render(next);
  }

  return {
    get state() {
      return state;
    },
    get url() {
      return url;
    },
    error() {
      if (destroyed) return;
      switch (state) {
        case "loading":
        case "loaded":
          set("scheduled");
          timer = setTimeout(() => {
            timer = null;
            if (destroyed) return;
            set("retrying");
            hooks.reload(url);
          }, delayMs);
          return;
        case "retrying":
        case "manual-retrying":
          report(url);
          set("failed");
          return;
        // scheduled / failed: a duplicate event for an attempt that has
        // already been counted.
      }
    },
    load() {
      if (destroyed || state === "loaded") return;
      cancelTimer();
      set("loaded");
    },
    retry() {
      if (destroyed || state !== "failed") return;
      set("manual-retrying");
      hooks.reload(url);
    },
    setURL(next) {
      if (destroyed || next === url) return;
      cancelTimer();
      url = next;
      set("loading");
    },
    destroy() {
      destroyed = true;
      cancelTimer();
    },
  };
}

// --- reporting -------------------------------------------------------

// reported holds every URL already sent to clientErrors this session.
// clientErrors is a 50-entry ring buffer that rides along on bug
// reports; a CDN outage fails every tile on the table, and without the
// dedupe it would push the one TypeError that actually matters out of
// the buffer inside a single frame. Bounded by the number of distinct
// art URLs a session can request, which is small.
const reported = new Set<string>();

// reportArtFailure records a failed art URL once per session.
export function reportArtFailure(url: string): void {
  if (reported.has(url)) return;
  reported.add(url);
  recordClientError(`card art failed to load: ${describeArtURL(url)}`);
}

// resetArtFailureReports forgets what has been reported. For tests.
export function resetArtFailureReports(): void {
  reported.clear();
}

// describeArtURL renders an art URL as the card id, face and size a
// bug triager needs, e.g. "id=abc face=1 size=small". Covers both the
// session route (/cards/{id}/image) and the public catalogue's
// (/catalog/image/{id}). A face-0 URL carries no `face` parameter
// (cardImage.ts explains why), so absence reads as face 0. Anything
// unrecognised comes back verbatim rather than being dropped.
export function describeArtURL(url: string): string {
  let parsed: URL;
  try {
    parsed = new URL(url, "http://art.invalid");
  } catch {
    return url;
  }
  const m =
    /^\/cards\/([^/]+)\/image$/.exec(parsed.pathname) ??
    /^\/catalog\/image\/([^/]+)$/.exec(parsed.pathname);
  if (!m) return url;
  const face = parsed.searchParams.get("face") ?? "0";
  const size = parsed.searchParams.get("size") ?? "normal";
  return `id=${decodeURIComponent(m[1])} face=${face} size=${size}`;
}
