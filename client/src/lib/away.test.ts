import { describe, expect, it } from "vitest";

import {
  awayOfferFor,
  awayOfferRequest,
  canDecideAway,
  describeAwayAfter,
  minutesSince,
  presenceChip,
  standInSelf,
} from "./away";
import { L } from "./labels";
import type { GameView, PlayerView } from "./protocol";

const NOW = Date.parse("2026-10-14T21:10:00Z");
const ago = (mins: number) => new Date(NOW - mins * 60000).toISOString();

function seat(id: string, extra: Partial<PlayerView> = {}): PlayerView {
  return { id, name: id, seat: 0, ...extra } as PlayerView;
}

function table(seats: PlayerView[]): GameView {
  return { id: "g", state: "active", seats } as unknown as GameView;
}

describe("presenceChip", () => {
  it("shows offline at once, with the minutes", () => {
    expect(
      presenceChip(seat("bo", { presence: { state: "offline", since: ago(2) } }), NOW),
    ).toEqual(expect.objectContaining({ kind: "offline", text: "offline 2m" }));
  });
  it("shows a hidden tab only once it has lasted two minutes", () => {
    const hidden = (m: number) => seat("bo", { presence: { state: "hidden", since: ago(m) } });
    expect(presenceChip(hidden(1), NOW)).toBeNull();
    expect(presenceChip(hidden(4), NOW)?.text).toBe("away 4m");
  });
  it("shows nothing for a present seat, a bot or a seat that has left", () => {
    expect(presenceChip(seat("bo"), NOW)).toBeNull();
    const p = { state: "offline" as const, since: ago(5) };
    expect(presenceChip(seat("b", { presence: p, is_bot: true }), NOW)).toBeNull();
    expect(presenceChip(seat("b", { presence: p, eliminated: true }), NOW)).toBeNull();
  });
  it("never counts backwards or reads garbage", () => {
    expect(minutesSince(ago(-3), NOW)).toBe(0);
    expect(minutesSince("not a time", NOW)).toBe(0);
  });
});

describe("who decides (ADR 0177 §3)", () => {
  const away = seat("bo", { presence: { state: "offline", since: ago(4), movable: true } });

  it("is the host while the host is at the table", () => {
    const v = table([seat("ana", { is_host: true }), away, seat("cy")]);
    expect(canDecideAway(v, "ana", away, false)).toBe(true);
    expect(canDecideAway(v, "cy", away, false)).toBe(false);
    expect(canDecideAway(v, "bo", away, false)).toBe(false);
    expect(awayOfferFor(v, "cy", false)).toBeNull();
    expect(awayOfferFor(v, "ana", false)?.id).toBe("bo");
  });
  it("is any present player when the host is gone too", () => {
    const v = table([
      seat("ana", { is_host: true, presence: { state: "offline", since: ago(9) } }),
      away,
      seat("cy"),
      seat("dee", { eliminated: true }),
    ]);
    expect(canDecideAway(v, "cy", away, false)).toBe(true);
    expect(canDecideAway(v, "dee", away, false)).toBe(false);
  });
  it("is a present player when the away seat is the host", () => {
    const host = seat("bo", {
      is_host: true,
      presence: { state: "offline", since: ago(4), movable: true },
    });
    const v = table([host, seat("cy")]);
    expect(canDecideAway(v, "cy", host, false)).toBe(true);
  });
  it("is the admin always, and never a spectator", () => {
    const v = table([seat("ana", { is_host: true }), away]);
    expect(canDecideAway(v, null, away, true)).toBe(true);
    expect(canDecideAway(v, null, away, false)).toBe(false);
  });
  it("offers nothing for a seat that is not movable or a bot already plays", () => {
    const notYet = seat("bo", { presence: { state: "offline", since: ago(1) } });
    expect(awayOfferFor(table([seat("ana", { is_host: true }), notYet]), "ana", false)).toBeNull();
    const played = seat("bo", {
      presence: { state: "offline", since: ago(9), movable: true },
      stand_in: { tier: "heuristic", since: ago(5) },
    });
    expect(awayOfferFor(table([seat("ana", { is_host: true }), played]), "ana", false)).toBeNull();
  });
});

describe("the offer sheet", () => {
  it("asks about the seat by name and sends each answer", () => {
    const sent: string[] = [];
    const bo = seat("bo", {
      name: "Bo",
      presence: { state: "offline", since: ago(3), movable: true },
    });
    const r = awayOfferRequest({
      seat: bo,
      now: NOW,
      busy: false,
      error: null,
      onAnswer: (a) => sent.push(a),
    });
    expect(r.label).toBe(L.awayOffer);
    expect(r.question).toBe("Bo has been offline for 3 minutes.");
    expect(r.row?.map((b) => b.label)).toEqual([
      L.awayStandIn("Bo"),
      L.awayConcede("Bo"),
      L.awayWait,
    ]);
    for (const b of r.row ?? []) b.onPress();
    expect(sent).toEqual(["stand_in", "concede", "wait"]);
  });
  it("disables the answers while one is in flight and shows a refusal", () => {
    const bo = seat("bo", { presence: { state: "offline", since: ago(3), movable: true } });
    const r = awayOfferRequest({
      seat: bo,
      now: NOW,
      busy: true,
      error: "someone has already answered",
      onAnswer: () => {},
    });
    expect(r.row?.every((b) => b.disabled)).toBe(true);
    expect(r.hint).toBe("someone has already answered");
    expect(r.hintWarn).toBe(true);
  });
});

describe("the take-back banner and the setting", () => {
  it("is the viewer's own seat while a bot plays it", () => {
    const me = seat("bo", { stand_in: { tier: "heuristic", since: ago(5) } });
    expect(standInSelf(table([me, seat("ana")]), "bo")?.id).toBe("bo");
    expect(standInSelf(table([me, seat("ana")]), "ana")).toBeNull();
  });
  it("describes the away time", () => {
    expect(describeAwayAfter(0)).toMatch(/^Never/);
    expect(describeAwayAfter(1)).toContain("1 minute,");
    expect(describeAwayAfter(undefined)).toContain("3 minutes");
  });
});
