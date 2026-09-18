import { describe, it, expect, beforeEach } from "vitest";
import { get, readable, writable } from "svelte/store";

import { guardSubscribe, guardedDerived, guardedWritable } from "./guardedStore";
import { recentClientErrors, resetClientErrors } from "./clientErrors";

// #266 / #720. svelte/store has ONE module-global `subscriber_queue`,
// drained by a bare loop that only clears it when the loop finishes, so
// a callback that throws leaves the queue permanently non-empty and
// every later `set()` on every store in the app enqueues without
// flushing. subscriberQueue.test.ts pins that mechanism directly; this
// file pins that the guards keep callbacks out of it.
//
// The recurring assertion is `expectQueueDrains()`: a plain, unrelated
// `writable` must still notify SYNCHRONOUSLY. That is the only
// observable difference between a healthy queue and a poisoned one from
// outside svelte/store, and it is the assertion that would fail if a
// guard ever let a throw escape.

beforeEach(() => {
  resetClientErrors();
});

function expectQueueDrains(): void {
  const canary = writable(0);
  const seen: number[] = [];
  const stop = canary.subscribe((v) => seen.push(v));
  seen.length = 0;
  canary.set(1);
  stop();
  expect(seen, "svelte/store's global subscriber_queue is poisoned").toEqual([1]);
}

const errorsMatching = (re: RegExp): string[] =>
  recentClientErrors()
    .map((e) => e.text)
    .filter((t) => re.test(t));

describe("guardedWritable", () => {
  it("behaves like a writable for set, update and get", () => {
    const s = guardedWritable(1, "counter");
    const seen: number[] = [];
    const stop = s.subscribe((v) => seen.push(v));

    s.set(2);
    s.update((v) => v + 10);

    expect(seen).toEqual([1, 2, 12]);
    expect(get(s)).toBe(12);
    stop();
  });

  it("keeps feeding healthy subscribers after one throws", () => {
    const s = guardedWritable(0, "counter");
    const healthy: number[] = [];

    s.subscribe(() => {
      throw new Error("subscriber blew up");
    });
    s.subscribe((v) => healthy.push(v));

    s.set(1);
    s.set(2);

    expect(healthy).toEqual([0, 1, 2]);
  });

  it("records the throw against the store's label", () => {
    const s = guardedWritable(0, "manualStops");
    s.subscribe(() => {
      throw new TypeError("cannot read x of undefined");
    });
    s.set(1);

    expect(errorsMatching(/manualStops/)).toHaveLength(2); // initial run + set
    expect(errorsMatching(/TypeError: cannot read x of undefined/).length).toBeGreaterThan(0);
  });

  it("does not let the throw reach the caller of set", () => {
    const s = guardedWritable(0, "counter");
    s.subscribe(() => {
      throw new Error("subscriber blew up");
    });

    expect(() => s.set(1)).not.toThrow();
  });

  it("leaves the global subscriber queue drainable", () => {
    const s = guardedWritable(0, "counter");
    s.subscribe(() => {
      throw new Error("subscriber blew up");
    });
    s.set(1);

    expectQueueDrains();
  });

  it("survives a subscriber that throws something that is not an Error", () => {
    const s = guardedWritable(0, "counter");
    // A cyclic object: describeThrown's JSON.stringify path throws on
    // it, and the failure path is not allowed to fail.
    const cyclic: Record<string, unknown> = {};
    cyclic.self = cyclic;
    s.subscribe(() => {
      throw cyclic;
    });

    expect(() => s.set(1)).not.toThrow();
    expect(errorsMatching(/unknown error/).length).toBeGreaterThan(0);
    expectQueueDrains();
  });
});

describe("guardedDerived", () => {
  it("maps the parent's value", () => {
    const parent = writable(2);
    const doubled = guardedDerived(parent, (v) => v * 2, "doubled", 0);

    const seen: number[] = [];
    const stop = doubled.subscribe((v) => seen.push(v));
    parent.set(5);
    stop();

    expect(seen).toEqual([4, 10]);
  });

  it("holds its last good value when the mapping throws", () => {
    const parent = writable(1);
    const mapped = guardedDerived(
      parent,
      (v) => {
        if (v === 2) throw new Error("mapping blew up");
        return v * 10;
      },
      "mapped",
      -1,
    );

    const seen: number[] = [];
    const stop = mapped.subscribe((v) => seen.push(v));
    parent.set(2);
    parent.set(3);
    stop();

    // The failing step emits nothing at all rather than a hole:
    // `writable`'s set suppresses an unchanged primitive, so the
    // consumer simply keeps the last good value it was given.
    expect(seen).toEqual([10, 30]);
    expect(errorsMatching(/derived mapped threw: Error: mapping blew up/)).toHaveLength(1);
  });

  it("uses the fallback when the mapping fails before any good value", () => {
    const parent = writable(1);
    const mapped = guardedDerived(
      parent,
      () => {
        throw new Error("always");
      },
      "mapped",
      "fallback",
    );

    expect(get(mapped)).toBe("fallback");
  });

  it("does not let a throwing mapping poison the queue", () => {
    const parent = writable(1);
    const mapped = guardedDerived(
      parent,
      () => {
        throw new Error("mapping blew up");
      },
      "mapped",
      0,
    );
    const stop = mapped.subscribe(() => {});
    expect(() => parent.set(2)).not.toThrow();
    stop();

    expectQueueDrains();
  });

  it("does not let a throwing subscriber on the derived poison the queue", () => {
    const parent = writable(1);
    const mapped = guardedDerived(parent, (v) => v + 1, "mapped", 0);

    const healthy: number[] = [];
    mapped.subscribe(() => {
      throw new Error("subscriber blew up");
    });
    mapped.subscribe((v) => healthy.push(v));

    expect(() => parent.set(9)).not.toThrow();
    expect(healthy).toEqual([2, 10]);
    expectQueueDrains();
  });

  it("stays lazy: it drops the parent subscription with its last subscriber", () => {
    // Laziness is inherited from svelte's own `derived` and is load
    // bearing for cardMetaCache's per-card stores and env's devFeature,
    // which are created per id and never disposed of explicitly.
    let starts = 0;
    let stops = 0;
    const parent = readable(1, () => {
      starts++;
      return () => stops++;
    });
    const mapped = guardedDerived(parent, (v) => v, "mapped", 0);

    expect(starts).toBe(0);
    const stop = mapped.subscribe(() => {});
    expect(starts).toBe(1);
    stop();
    expect(stops).toBe(1);
  });
});

describe("guardSubscribe", () => {
  it("guards a hand-rolled store's subscribe", () => {
    // The `{ subscribe }` shape several modules export. Wrapping the
    // raw subscribe is the escape hatch for a store this module did not
    // build.
    const inner = writable(0);
    const exposed = { subscribe: guardSubscribe(inner.subscribe, "exposed") };

    const healthy: number[] = [];
    exposed.subscribe(() => {
      throw new Error("subscriber blew up");
    });
    exposed.subscribe((v) => healthy.push(v));

    expect(() => inner.set(7)).not.toThrow();
    expect(healthy).toEqual([0, 7]);
    expectQueueDrains();
  });
});
