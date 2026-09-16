import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  ART_RETRY_DELAY_MS,
  createArtRetry,
  describeArtURL,
  flattensChildren,
  markerFor,
  reportArtFailure,
  resetArtFailureReports,
  withIDRef,
  withoutIDRef,
  type ArtState,
} from "./cardArtRetry";
import { recentClientErrors, resetClientErrors } from "./clientErrors";

const URL_A = "/cards/aaaa/image?size=small";
const URL_B = "/cards/aaaa/image?size=small&face=1";

// harness wires a controller to spies so each test reads as the event
// sequence an <img> would produce.
function harness(url = URL_A) {
  const reload = vi.fn<(url: string) => void>();
  const report = vi.fn<(url: string) => void>();
  const states: ArtState[] = [];
  const art = createArtRetry(url, { reload, report, render: (s) => states.push(s) });
  return { art, reload, report, states };
}

describe("createArtRetry", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("does nothing on a clean load", () => {
    const { art, reload, report } = harness();
    art.load();
    expect(art.state).toBe("loaded");
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS * 2);
    expect(reload).not.toHaveBeenCalled();
    expect(report).not.toHaveBeenCalled();
  });

  it("retries the plain URL once, after the delay", () => {
    const { art, reload } = harness();
    art.error();
    expect(art.state).toBe("scheduled");
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS - 1);
    expect(reload).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1);
    expect(reload).toHaveBeenCalledTimes(1);
    // The same string, not a cache-busted variant: the service worker
    // keys card art on the full query string.
    expect(reload).toHaveBeenCalledWith(URL_A);
    expect(art.state).toBe("retrying");
  });

  it("recovers silently when the automatic retry loads", () => {
    const { art, report, states } = harness();
    art.error();
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS);
    art.load();
    expect(art.state).toBe("loaded");
    expect(report).not.toHaveBeenCalled();
    expect(states.map(markerFor)).not.toContain("ready");
  });

  it("fails and reports once the automatic retry also errors, and stops retrying", () => {
    const { art, reload, report } = harness();
    art.error();
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS);
    art.error();
    expect(art.state).toBe("failed");
    expect(report).toHaveBeenCalledWith(URL_A);
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS * 10);
    expect(reload).toHaveBeenCalledTimes(1);
  });

  it("ignores a duplicate error while the retry is still scheduled", () => {
    const { art, reload } = harness();
    art.error();
    art.error();
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS);
    expect(reload).toHaveBeenCalledTimes(1);
    expect(art.state).toBe("retrying");
  });

  // An image that loaded can still fire `error` later: the browser
  // re-fetches when src is set again, and that fetch can fail. Art that
  // was on screen gets the same automatic retry as a first load.
  it("an error after a load schedules the automatic retry", () => {
    const { art, reload, report } = harness();
    art.load();
    art.error();
    expect(art.state).toBe("scheduled");
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS);
    expect(reload).toHaveBeenCalledTimes(1);
    expect(reload).toHaveBeenCalledWith(URL_A);
    expect(art.state).toBe("retrying");
    expect(report).not.toHaveBeenCalled();
  });

  // Once failed, the attempt has been counted and reported; a second
  // `error` for it must neither report again nor restart the cycle.
  it("ignores a duplicate error once failed", () => {
    const { art, reload, report, states } = harness();
    art.error();
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS);
    art.error();
    expect(art.state).toBe("failed");
    const before = states.length;

    art.error();
    expect(art.state).toBe("failed");
    expect(states.length).toBe(before);
    expect(report).toHaveBeenCalledTimes(1);
    expect(vi.getTimerCount()).toBe(0);
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS * 10);
    expect(reload).toHaveBeenCalledTimes(1);
  });

  it("click-retry is a no-op until the art has actually failed", () => {
    const { art, reload } = harness();
    art.retry();
    art.error();
    art.retry();
    expect(reload).not.toHaveBeenCalled();
    expect(art.state).toBe("scheduled");
  });

  it("click-retry re-requests immediately, as often as the player clicks", () => {
    const { art, reload, report } = harness();
    art.error();
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS);
    art.error();

    art.retry();
    expect(art.state).toBe("manual-retrying");
    expect(reload).toHaveBeenCalledTimes(2);
    // A second click while the first is in flight does not stack.
    art.retry();
    expect(reload).toHaveBeenCalledTimes(2);

    art.error();
    expect(art.state).toBe("failed");
    art.retry();
    art.load();
    expect(art.state).toBe("loaded");
    expect(reload).toHaveBeenCalledTimes(3);
    expect(reload).toHaveBeenLastCalledWith(URL_A);
    // No automatic retry is ever scheduled after a manual one fails.
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS * 10);
    expect(reload).toHaveBeenCalledTimes(3);
    // The controller reports on every visible failure; the dedupe is
    // reportArtFailure's job, not the state machine's.
    expect(report).toHaveBeenCalledTimes(2);
  });

  it("a load cancels the pending retry", () => {
    const { art, reload } = harness();
    art.error();
    art.load();
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS);
    expect(reload).not.toHaveBeenCalled();
    expect(art.state).toBe("loaded");
  });

  // A DFC transforming, or the hover panel swapping cards, changes the
  // URL under a mounted <img>. The new art gets its own full cycle.
  it("resets on a URL change and cancels the old URL's retry", () => {
    const { art, reload } = harness();
    art.error();
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS);
    art.error();
    expect(art.state).toBe("failed");

    art.setURL(URL_B);
    expect(art.state).toBe("loading");
    expect(art.url).toBe(URL_B);

    art.error();
    art.setURL(URL_A);
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS);
    // Only the first cycle's retry ever fired.
    expect(reload).toHaveBeenCalledTimes(1);
    expect(art.state).toBe("loading");
  });

  it("the same URL does not restart the cycle", () => {
    const { art, states } = harness();
    art.error();
    const before = states.length;
    art.setURL(URL_A);
    expect(art.state).toBe("scheduled");
    expect(states.length).toBe(before);
  });

  it("destroy clears the timer and ignores late events", () => {
    const { art, reload, report, states } = harness();
    art.error();
    art.destroy();
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS * 2);
    expect(reload).not.toHaveBeenCalled();
    expect(vi.getTimerCount()).toBe(0);

    const before = states.length;
    art.error();
    art.load();
    art.retry();
    art.setURL(URL_B);
    expect(states.length).toBe(before);
    expect(report).not.toHaveBeenCalled();
  });

  it("renders every transition", () => {
    const { art, states } = harness();
    art.error();
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS);
    art.error();
    art.retry();
    art.load();
    expect(states).toEqual(["scheduled", "retrying", "failed", "manual-retrying", "loaded"]);
  });

  it("defaults to reportArtFailure", () => {
    resetClientErrors();
    resetArtFailureReports();
    const art = createArtRetry(URL_A, { reload: () => {}, render: () => {} });
    art.error();
    vi.advanceTimersByTime(ART_RETRY_DELAY_MS);
    art.error();
    expect(recentClientErrors().map((e) => e.text)).toEqual([
      "card art failed to load: id=aaaa face=0 size=small",
    ]);
  });
});

describe("markerFor", () => {
  it("shows the marker only once the art has failed", () => {
    const want: Record<ArtState, string> = {
      loading: "hidden",
      scheduled: "hidden",
      retrying: "hidden",
      failed: "ready",
      "manual-retrying": "busy",
      loaded: "hidden",
    };
    for (const [state, marker] of Object.entries(want)) {
      expect(markerFor(state as ArtState)).toBe(marker);
    }
  });
});

// A marker inside one of these must not be a button of its own: the
// ancestor's children are presentational, so a nested button is a tab
// stop with no role or name.
describe("flattensChildren", () => {
  it("is true for the hosts card art actually sits in", () => {
    // board Card: role="button" with a click, role="img" without
    expect(flattensChildren("DIV", "button")).toBe(true);
    expect(flattensChildren("DIV", "img")).toBe(true);
    // face-picker option: a native button
    expect(flattensChildren("BUTTON", null)).toBe(true);
  });

  it("is false for plain containers", () => {
    // catalogue tile, mulligan grid (role="listitem"), reveal banner,
    // a stack item that is not currently targetable
    expect(flattensChildren("DIV", null)).toBe(false);
    expect(flattensChildren("DIV", "listitem")).toBe(false);
    expect(flattensChildren("SPAN", "")).toBe(false);
    expect(flattensChildren("LI", null)).toBe(false);
  });

  it("lets an explicit role override the tag", () => {
    expect(flattensChildren("BUTTON", "listitem")).toBe(false);
    expect(flattensChildren("SPAN", "switch")).toBe(true);
    expect(flattensChildren("div", "Option")).toBe(true);
  });

  it("reads only the first role token, the rest being fallbacks", () => {
    expect(flattensChildren("DIV", " img button")).toBe(true);
    expect(flattensChildren("DIV", "group button")).toBe(false);
  });
});

describe("withIDRef / withoutIDRef", () => {
  it("adds without clobbering a component's own references", () => {
    expect(withIDRef(null, "art-1")).toBe("art-1");
    expect(withIDRef("hint", "art-1")).toBe("hint art-1");
    expect(withIDRef("  hint   other ", "art-1")).toBe("hint other art-1");
  });

  it("does not add the same id twice", () => {
    expect(withIDRef("hint art-1", "art-1")).toBe("hint art-1");
  });

  it("removes only its own id, and reports empty as null", () => {
    expect(withoutIDRef("hint art-1", "art-1")).toBe("hint");
    expect(withoutIDRef("art-1", "art-1")).toBeNull();
    expect(withoutIDRef(null, "art-1")).toBeNull();
    expect(withoutIDRef("art-10", "art-1")).toBe("art-10");
  });
});

describe("reportArtFailure", () => {
  beforeEach(() => {
    resetClientErrors();
    resetArtFailureReports();
  });

  // A CDN outage fails every tile on the table at once. One entry per
  // URL keeps the 50-entry buffer for errors that are not this one.
  it("records each URL once", () => {
    for (let i = 0; i < 20; i++) reportArtFailure(URL_A);
    reportArtFailure(URL_B);
    reportArtFailure(URL_B);
    expect(recentClientErrors().map((e) => e.text)).toEqual([
      "card art failed to load: id=aaaa face=0 size=small",
      "card art failed to load: id=aaaa face=1 size=small",
    ]);
  });

  it("forgets on reset", () => {
    reportArtFailure(URL_A);
    resetArtFailureReports();
    reportArtFailure(URL_A);
    expect(recentClientErrors()).toHaveLength(2);
  });
});

describe("describeArtURL", () => {
  it("reads id, face and size off the session route", () => {
    expect(describeArtURL("/cards/abc-123/image?size=normal&face=1")).toBe(
      "id=abc-123 face=1 size=normal",
    );
  });

  // Face 0 emits no parameter (cardImage.ts), so absence is face 0.
  it("treats a missing face as the front", () => {
    expect(describeArtURL("/cards/abc/image?size=art_crop")).toBe("id=abc face=0 size=art_crop");
  });

  it("reads the public catalogue route", () => {
    expect(describeArtURL("/catalog/image/abc?size=normal")).toBe("id=abc face=0 size=normal");
  });

  it("falls back to the server's default size", () => {
    expect(describeArtURL("/cards/abc/image")).toBe("id=abc face=0 size=normal");
  });

  it("returns anything unrecognised verbatim", () => {
    expect(describeArtURL("/card-back.jpg")).toBe("/card-back.jpg");
    expect(describeArtURL("https://cdn.example/x.jpg")).toBe("https://cdn.example/x.jpg");
  });
});
