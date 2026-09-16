// cardArt.ts — the `use:cardArt` action every card-art <img> wears (#33).
//
//   <img src={url} alt="" use:cardArt={url} />
//
// It adds one automatic retry, a corner marker once that retry has
// also failed, click-to-retry on the marker, and a deduplicated entry
// in clientErrors. The decisions live in cardArtRetry.ts, which is
// unit-tested; this file is only the DOM wiring.
//
// An action rather than a wrapper component, because a component's
// <img> would not carry its parent's scoped styles: each call site
// sizes its art with a local `img { … }` rule, and a wrapper would
// mean rewriting all seven as :global().
//
// The marker is created here, as the img's next sibling, and styled
// by `.card-art-error` in app.css (global for the same scoping
// reason). It is absolutely positioned, so the img's parent must be a
// positioned box; `--art-error-top` / `--art-error-right` move it off
// a corner a site already uses.
//
// The URL passed in must be the same string as the img's `src`. On a
// change (a DFC transforming, a face flip) the action resets to a
// fresh cycle for the new URL, so a failure on one face never marks
// the other. Face-down and card-back images must not wear this: the
// back is a bundled static asset, not card art.

import type { Action } from "svelte/action";
import {
  ART_FAILED_TITLE,
  createArtRetry,
  markerFor,
  type ArtState,
  type MarkerState,
} from "./cardArtRetry";

export type CardArtParam =
  | string
  | {
      url: string;
      // interactive: false renders the marker as a plain signal with
      // no click-retry and no tab stop. For art inside a surface the
      // pointer can never reach and assistive tech is told to skip —
      // the hover-zoom panel is pointer-events: none and aria-hidden.
      interactive?: boolean;
    };

function normalise(p: CardArtParam): { url: string; interactive: boolean } {
  return typeof p === "string"
    ? { url: p, interactive: true }
    : { url: p.url, interactive: p.interactive ?? true };
}

export const cardArt: Action<HTMLImageElement, CardArtParam> = (img, param) => {
  let opts = normalise(param);
  let marker: HTMLElement | null = null;
  let markerState: MarkerState = "hidden";

  const art = createArtRetry(opts.url, {
    // Setting the attribute re-runs the image fetch even when the
    // value is unchanged, and a failed image is never served from the
    // browser's memory cache — so the plain URL is a real retry.
    reload: (url) => img.setAttribute("src", url),
    render: (state: ArtState) => renderMarker(markerFor(state)),
  });

  // Events are checked against the URL the controller is tracking: a
  // late event for an image whose src has since moved on belongs to a
  // cycle that setURL already reset.
  function onError(): void {
    if (img.getAttribute("src") === art.url) art.error();
  }
  function onLoad(): void {
    if (img.getAttribute("src") === art.url) art.load();
  }

  // The marker sits inside tiles that are themselves click targets —
  // a board card plays, taps or selects; a face-picker option selects
  // and double-click confirms; a command-zone card casts on
  // double-click. None of those may fire from the marker, so it stops
  // propagation of everything it handles. Keyboard activation matters
  // for the same reason: Space on a focused marker must not reach the
  // card's own Enter/Space handler or the pass-priority shortcut.
  function activate(ev: Event): void {
    ev.preventDefault();
    ev.stopPropagation();
    art.retry();
  }
  function onMarkerKeydown(ev: KeyboardEvent): void {
    if (ev.key !== "Enter" && ev.key !== " ") return;
    activate(ev);
  }
  function swallow(ev: Event): void {
    ev.stopPropagation();
  }

  function createMarker(): HTMLElement {
    const el = document.createElement("span");
    el.className = "card-art-error";
    el.title = ART_FAILED_TITLE;
    el.textContent = "!";
    if (opts.interactive) {
      // The tooltip stays exactly ART_FAILED_TITLE; the accessible
      // name also says what activating it does.
      el.setAttribute("aria-label", `${ART_FAILED_TITLE}, retry`);
      el.setAttribute("role", "button");
      el.tabIndex = 0;
      el.addEventListener("click", activate);
      el.addEventListener("keydown", onMarkerKeydown);
      el.addEventListener("dblclick", swallow);
    } else {
      el.setAttribute("aria-label", ART_FAILED_TITLE);
      el.setAttribute("role", "img");
      el.classList.add("static");
    }
    return el;
  }

  function renderMarker(next: MarkerState): void {
    markerState = next;
    if (next === "hidden") {
      marker?.remove();
      marker = null;
      return;
    }
    if (!marker) {
      marker = createMarker();
      img.after(marker);
    }
    const busy = next === "busy";
    marker.classList.toggle("busy", busy);
    if (opts.interactive) marker.setAttribute("aria-disabled", busy ? "true" : "false");
  }

  img.addEventListener("error", onError);
  img.addEventListener("load", onLoad);
  // An image can fail before the action attaches — its fetch starts
  // when `src` is set, and a fast 502 can land first. `complete` with
  // no pixels is a broken image; one still loading, or deferred by
  // loading="lazy", reports complete === false and waits for events.
  if (img.complete && img.naturalWidth === 0) onError();

  return {
    update(next: CardArtParam) {
      const prev = opts;
      opts = normalise(next);
      if (prev.interactive !== opts.interactive && marker) {
        marker.remove();
        marker = null;
        renderMarker(markerState);
      }
      art.setURL(opts.url);
    },
    destroy() {
      art.destroy();
      img.removeEventListener("error", onError);
      img.removeEventListener("load", onLoad);
      marker?.remove();
      marker = null;
    },
  };
};
