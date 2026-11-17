import { describe, expect, it } from "vitest";

import { closedReasonLabel, idleCloseNotice, idleCloseWhen } from "./idleClose";
import { L } from "./labels";

// idleClose.test.ts: ADR 0211 on the client. The table's notice reads
// the warning's closing time in the viewer's own clock, and a closed
// table's chip says why the server closed it.

// Local times, so the test reads the same in every time zone.
const at = (y: number, mo: number, d: number, h: number, mi: number) =>
  new Date(y, mo - 1, d, h, mi).getTime();
const clock = (ms: number) =>
  new Date(ms).toLocaleTimeString(undefined, { hour: "2-digit", minute: "2-digit" });

describe("idleCloseWhen", () => {
  // Wednesday 11 November 2026, 09:00.
  const now = at(2026, 11, 11, 9, 0);

  it("says today, tomorrow, or the weekday", () => {
    const today = at(2026, 11, 11, 21, 14);
    const tomorrow = at(2026, 11, 12, 3, 5);
    const friday = at(2026, 11, 13, 21, 14);
    expect(idleCloseWhen(today, now)).toBe(`at ${clock(today)} today`);
    expect(idleCloseWhen(tomorrow, now)).toBe(`at ${clock(tomorrow)} tomorrow`);
    const weekday = new Date(friday).toLocaleDateString(undefined, { weekday: "long" });
    expect(idleCloseWhen(friday, now)).toBe(`at ${clock(friday)} on ${weekday}`);
  });

  it("never says a day that has gone", () => {
    const past = at(2026, 11, 11, 8, 0);
    expect(idleCloseWhen(past, now)).toBe(`at ${clock(past)} today`);
  });
});

describe("idleCloseNotice", () => {
  const now = at(2026, 11, 11, 9, 0);

  it("is the registered line for a warned frame", () => {
    const closes = at(2026, 11, 11, 21, 14);
    const line = idleCloseNotice(new Date(closes).toISOString(), now);
    expect(line).toBe(L.idleCloseNotice(idleCloseWhen(closes, now)));
    expect(line).toMatch(/^This game closes at .+ today unless someone moves$/);
  });

  it("is nothing for an unwarned frame or a value that does not parse", () => {
    expect(idleCloseNotice(undefined, now)).toBeNull();
    expect(idleCloseNotice("", now)).toBeNull();
    expect(idleCloseNotice("soon", now)).toBeNull();
  });
});

describe("closedReasonLabel", () => {
  it("names each of the server's reasons", () => {
    expect(closedReasonLabel("idle_lobby")).toBe(L.closedIdleLobby);
    expect(closedReasonLabel("idle_game")).toBe(L.closedIdleGame);
    expect(closedReasonLabel("idle_bots")).toBe(L.closedIdleBots);
  });

  it("has nothing for an admin's archive or a reason it does not know", () => {
    expect(closedReasonLabel(undefined)).toBeNull();
    expect(closedReasonLabel(null)).toBeNull();
    expect(closedReasonLabel("")).toBeNull();
    expect(closedReasonLabel("idle_something_new")).toBeNull();
  });
});
