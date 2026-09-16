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
// positioned box; `--art-error-top` / `-right` / `-left` / `-z` move
// it off a corner a site already uses.
//
// The URL passed in must be the same string as the img's `src`. On a
// change (a DFC transforming, a face flip) the action resets to a
// fresh cycle for the new URL, so a failure on one face never marks
// the other. Face-down and card-back images must not wear this: the
// back is a bundled static asset, not card art.
//
// The marker comes in three kinds:
//
//   button     — role="button", a tab stop, Enter/Space retry. When
//                nothing above the img is a control or an image (the
//                catalogue, the mulligan grid, the reveal banner).
//   described  — inside a control or image (board card, face-picker
//                option, targetable stack item), whose children
//                assistive tech treats as presentational: a nested
//                button would be a nameless tab stop. The marker is
//                aria-hidden and pointer-only; the outermost such
//                ancestor — the pile <button> around a role="img"
//                Card, not the Card — gets "Art failed to load" as its
//                accessible description, and keyboard focus on it is
//                the retry. cardArtRetry.ts has the full reasoning.
//   static     — `interactive: false`: a signal with no retry at all.

import type { Action } from "svelte/action";
import {
  ART_FAILED_TITLE,
  createArtRetry,
  describedHost,
  markerFor,
  withIDRef,
  withoutIDRef,
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

type MarkerKind = "button" | "described" | "static";

function normalise(p: CardArtParam): { url: string; interactive: boolean } {
  return typeof p === "string"
    ? { url: p, interactive: true }
    : { url: p.url, interactive: p.interactive ?? true };
}

// Description ids only need to be unique within the document.
let nextDescriptionID = 0;

export const cardArt: Action<HTMLImageElement, CardArtParam> = (img, param) => {
  let opts = normalise(param);
  let marker: HTMLElement | null = null;
  let markerKind: MarkerKind | null = null;
  // The control a "described" marker is attached to, and the id of
  // the description element it references.
  let host: HTMLElement | null = null;
  let descriptionID = "";
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

  // Keyboard focus landing on a control whose art has failed retries
  // it — the keyboard's equivalent of clicking the marker, which is
  // not a tab stop there. Only keyboard focus: a mouse click focuses
  // the tile too, and a click on the tile means play / tap / select,
  // while the pointer has the marker for retrying.
  function onHostFocus(): void {
    if (host && focusIsVisible(host)) art.retry();
  }

  function ancestors(): HTMLElement[] {
    const out: HTMLElement[] = [];
    for (let el = img.parentElement; el && el !== document.body; el = el.parentElement) {
      out.push(el);
    }
    return out;
  }

  // flatteningHost is the ancestor a "described" marker attaches to —
  // the outermost one whose children assistive tech treats as
  // presentational (describedHost has why) — or null.
  function flatteningHost(): HTMLElement | null {
    const chain = ancestors();
    const i = describedHost(
      chain.map((el) => ({ tagName: el.tagName, role: el.getAttribute("role") })),
    );
    return i < 0 ? null : chain[i];
  }

  // A role can change under a marker that is already showing: a stack
  // item is role="button" only while it is a legal target, and a hand
  // card is role="button" only while it is castable. While the marker
  // shows, the ancestors' roles are watched and the kind re-decided.
  let roleWatch: MutationObserver | null = null;
  function watchRoles(): void {
    if (roleWatch || typeof MutationObserver === "undefined") return;
    roleWatch = new MutationObserver(() => {
      if (marker) renderMarker(markerState);
    });
    for (const el of ancestors()) {
      roleWatch.observe(el, { attributes: true, attributeFilter: ["role"] });
    }
  }
  function unwatchRoles(): void {
    roleWatch?.disconnect();
    roleWatch = null;
  }

  function createMarker(kind: MarkerKind, control: HTMLElement | null): HTMLElement {
    const el = document.createElement("span");
    el.className = "card-art-error";
    el.title = ART_FAILED_TITLE;
    el.textContent = "!";
    switch (kind) {
      case "button":
        // The tooltip stays exactly ART_FAILED_TITLE; the accessible
        // name also says what activating it does.
        el.setAttribute("aria-label", `${ART_FAILED_TITLE}, retry`);
        el.setAttribute("role", "button");
        el.tabIndex = 0;
        el.addEventListener("click", activate);
        el.addEventListener("keydown", onMarkerKeydown);
        el.addEventListener("dblclick", swallow);
        break;
      case "described": {
        el.setAttribute("aria-hidden", "true");
        el.addEventListener("click", activate);
        el.addEventListener("dblclick", swallow);
        // A description element that is itself hidden is still read
        // when referenced by id — aria-describedby is the documented
        // exception to hidden content being skipped.
        const desc = document.createElement("span");
        descriptionID = `card-art-error-${++nextDescriptionID}`;
        desc.id = descriptionID;
        desc.className = "card-art-error-desc";
        desc.textContent = ART_FAILED_TITLE;
        el.append(desc);
        host = control;
        if (host) {
          host.setAttribute(
            "aria-describedby",
            withIDRef(host.getAttribute("aria-describedby"), descriptionID),
          );
          host.addEventListener("focus", onHostFocus);
        }
        break;
      }
      case "static":
        el.setAttribute("aria-label", ART_FAILED_TITLE);
        el.setAttribute("role", "img");
        el.classList.add("static");
        break;
    }
    return el;
  }

  function removeMarker(): void {
    if (host) {
      host.removeEventListener("focus", onHostFocus);
      const rest = withoutIDRef(host.getAttribute("aria-describedby"), descriptionID);
      if (rest === null) host.removeAttribute("aria-describedby");
      else host.setAttribute("aria-describedby", rest);
      host = null;
    }
    marker?.remove();
    marker = null;
    markerKind = null;
  }

  function renderMarker(next: MarkerState): void {
    markerState = next;
    if (next === "hidden") {
      unwatchRoles();
      removeMarker();
      return;
    }
    // Re-decided on every render, not only at creation, and on any
    // ancestor role change while the marker shows (watchRoles).
    if (opts.interactive) watchRoles();
    else unwatchRoles();
    const control = opts.interactive ? flatteningHost() : null;
    const kind: MarkerKind = !opts.interactive ? "static" : control ? "described" : "button";
    if (marker && (kind !== markerKind || control !== host)) removeMarker();
    if (!marker) {
      marker = createMarker(kind, control);
      markerKind = kind;
      img.after(marker);
    }
    const busy = next === "busy";
    marker.classList.toggle("busy", busy);
    if (kind === "button") marker.setAttribute("aria-disabled", busy ? "true" : "false");
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
      opts = normalise(next);
      // Rebuilds the marker if `interactive` changed its kind.
      if (marker) renderMarker(markerState);
      art.setURL(opts.url);
    },
    destroy() {
      art.destroy();
      img.removeEventListener("error", onError);
      img.removeEventListener("load", onLoad);
      unwatchRoles();
      removeMarker();
    },
  };
};

// focusIsVisible is `:focus-visible`, treated as true where the
// selector is unsupported so the keyboard route never silently
// disappears.
function focusIsVisible(el: HTMLElement): boolean {
  try {
    return el.matches(":focus-visible");
  } catch {
    return true;
  }
}
