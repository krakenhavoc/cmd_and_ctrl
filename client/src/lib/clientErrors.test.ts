import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  CLIENT_ERROR_LIMIT,
  installErrorCapture,
  recentClientErrors,
  recordClientError,
  resetClientErrors,
} from "./clientErrors";

// fakeWindow records listeners so tests can fire events without a DOM.
function fakeWindow() {
  const listeners: Record<string, ((ev: Event) => void)[]> = {};
  return {
    listeners,
    addEventListener: (type: string, fn: EventListenerOrEventListenerObject) => {
      listeners[type] = [...(listeners[type] ?? []), fn as (ev: Event) => void];
    },
    removeEventListener: (type: string, fn: EventListenerOrEventListenerObject) => {
      listeners[type] = (listeners[type] ?? []).filter((f) => f !== fn);
    },
    fire: (type: string, ev: unknown) => {
      for (const fn of listeners[type] ?? []) fn(ev as Event);
    },
  };
}

describe("recordClientError", () => {
  beforeEach(resetClientErrors);

  it("stamps entries and keeps insertion order", () => {
    recordClientError("first");
    recordClientError("second");
    const got = recentClientErrors();
    expect(got.map((e) => e.text)).toEqual(["first", "second"]);
    expect(got[0].at).toBeGreaterThan(0);
  });

  // A broken render loop emits thousands of near-identical errors; the
  // buffer must stay bounded rather than eating the tab.
  it("drops the oldest past the limit", () => {
    for (let i = 0; i < CLIENT_ERROR_LIMIT + 10; i++) recordClientError(`e${i}`);
    const got = recentClientErrors();
    expect(got).toHaveLength(CLIENT_ERROR_LIMIT);
    expect(got[got.length - 1].text).toBe(`e${CLIENT_ERROR_LIMIT + 9}`);
    expect(got[0].text).toBe("e10");
  });

  it("clips a very long entry", () => {
    recordClientError("x".repeat(5000));
    expect(recentClientErrors()[0].text.length).toBeLessThanOrEqual(400);
  });
});

describe("installErrorCapture", () => {
  // Teardown runs even when an assertion throws mid-test: the module
  // holds an installed flag, so a leaked install would silently make
  // every later test in this block a no-op.
  let teardowns: (() => void)[] = [];
  const install = (...args: Parameters<typeof installErrorCapture>) => {
    const off = installErrorCapture(...args);
    teardowns.push(off);
    return off;
  };

  beforeEach(() => {
    resetClientErrors();
    teardowns = [];
  });
  afterEach(() => {
    for (const off of teardowns.reverse()) off();
  });

  it("captures uncaught errors with their source location", () => {
    const win = fakeWindow();
    const con = { error: vi.fn(), warn: vi.fn() };
    install(win, con);
    win.fire("error", { message: "x is not a function", filename: "a.js", lineno: 4, colno: 9 });
    expect(recentClientErrors()[0].text).toBe("uncaught x is not a function (a.js:4:9)");
  });

  // Rejection reasons can be anything at all, and the handler must not
  // throw while handling an error.
  it("describes any rejection reason", () => {
    const win = fakeWindow();
    install(win, { error: vi.fn(), warn: vi.fn() });
    win.fire("unhandledrejection", { reason: new TypeError("bad cast") });
    win.fire("unhandledrejection", { reason: "plain string" });
    win.fire("unhandledrejection", { reason: { code: 500 } });
    win.fire("unhandledrejection", { reason: undefined });
    const texts = recentClientErrors().map((e) => e.text);
    expect(texts[0]).toBe("unhandled rejection: TypeError: bad cast");
    expect(texts[1]).toBe("unhandled rejection: plain string");
    expect(texts[2]).toBe('unhandled rejection: {"code":500}');
    expect(texts[3]).toContain("unhandled rejection:");
  });

  // Capturing must never cost the developer their devtools output.
  it("chains console.error and console.warn through to the original", () => {
    const con = { error: vi.fn(), warn: vi.fn() };
    // Hold the originals: install REPLACES these properties, so
    // asserting on con.error afterwards would inspect the wrapper.
    const originalError = con.error;
    const originalWarn = con.warn;
    install(fakeWindow(), con);
    con.error("boom", new Error("inner"));
    con.warn("careful");
    expect(recentClientErrors().map((e) => e.text)).toEqual([
      "console.error: boom Error: inner",
      "console.warn: careful",
    ]);
    expect(originalError).toHaveBeenCalledWith("boom", expect.any(Error));
    expect(originalWarn).toHaveBeenCalledWith("careful");
  });

  // A hot reload calling install twice must not stack wrappers, or one
  // console.error would record three entries.
  it("is idempotent", () => {
    const win = fakeWindow();
    const con = { error: vi.fn(), warn: vi.fn() };
    install(win, con);
    install(win, con);
    con.error("once");
    expect(recentClientErrors()).toHaveLength(1);
  });

  it("teardown restores the console and unhooks the listeners", () => {
    const win = fakeWindow();
    const original = vi.fn();
    const con = { error: original, warn: vi.fn() };
    const teardown = install(win, con);
    teardown();
    expect(con.error).toBe(original);
    win.fire("error", { message: "after teardown" });
    expect(recentClientErrors()).toHaveLength(0);
  });
});
