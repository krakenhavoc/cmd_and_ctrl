import { describe, expect, it } from "vitest";

import { measure, pingLabel } from "./serverPing";

// A clock that advances by the given steps, one per call.
function clock(...times: number[]): () => number {
  let i = 0;
  return () => times[Math.min(i++, times.length - 1)];
}

describe("measure", () => {
  it("times one GET /healthz round trip", async () => {
    const calls: string[] = [];
    const fetchFn = (async (url: string, init?: RequestInit) => {
      calls.push(`${url} ${init?.cache}`);
      return new Response("ok", { status: 200 });
    }) as typeof fetch;
    expect(await measure(fetchFn, clock(100, 137.6))).toEqual({ state: "online", ms: 38 });
    expect(calls).toEqual(["/healthz no-store"]);
  });

  it("is offline on a non-2xx answer or a network error", async () => {
    const down = (async () => new Response("", { status: 503 })) as unknown as typeof fetch;
    expect(await measure(down, clock(0, 5))).toEqual({ state: "offline" });
    const broken = (async () => {
      throw new TypeError("network");
    }) as unknown as typeof fetch;
    expect(await measure(broken, clock(0, 5))).toEqual({ state: "offline" });
  });

  it("never reports 0 ms", async () => {
    const fast = (async () => new Response("ok")) as unknown as typeof fetch;
    expect(await measure(fast, clock(10, 10.2))).toEqual({ state: "online", ms: 1 });
  });
});

describe("pingLabel", () => {
  it("says what the chip shows", () => {
    expect(pingLabel({ state: "online", ms: 38 })).toBe("Online · 38 ms");
    expect(pingLabel({ state: "offline" })).toBe("Offline");
    expect(pingLabel({ state: "unknown" })).toBe("");
  });
});
