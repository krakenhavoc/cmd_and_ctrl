// statsPrefs.test.ts — the pure half of the advanced-stats setting
// (ADR 0217 §2, owner answer 1): the wire shape, the words beside the
// switch, and the admin listing's merge.

import { describe, expect, it } from "vitest";
import {
  STATS_PREFS_OFF,
  accountStatsOf,
  accountStatsText,
  normalizeStatsPrefs,
  statsPrefsNote,
  withAccount,
  withDefault,
  type AdminAdvancedStats,
} from "./statsPrefs";

describe("normalizeStatsPrefs", () => {
  it("reads anything unexpected as off and following the default", () => {
    expect(normalizeStatsPrefs(null)).toEqual(STATS_PREFS_OFF);
    expect(normalizeStatsPrefs({ advanced: "yes", choice: "maybe", set_by: "bot" })).toEqual(
      STATS_PREFS_OFF,
    );
  });
  it("keeps who set it only when someone chose", () => {
    expect(
      normalizeStatsPrefs({ available: true, advanced: true, choice: "on", set_by: "admin" }),
    ).toEqual({
      available: true,
      advanced: true,
      choice: "on",
      set_by: "admin",
      default: false,
      share_pods: false,
    });
    expect(normalizeStatsPrefs({ choice: "default", set_by: "self" }).set_by).toBeUndefined();
  });
  it("reads the pod-sharing switch as off unless it is exactly true", () => {
    expect(normalizeStatsPrefs({ share_pods: true }).share_pods).toBe(true);
    expect(normalizeStatsPrefs({ share_pods: "yes" }).share_pods).toBe(false);
    expect(normalizeStatsPrefs({}).share_pods).toBe(false);
  });
});

describe("statsPrefsNote", () => {
  const base = { ...STATS_PREFS_OFF, available: true };
  it("says it is off until the person turns it on, by default", () => {
    expect(statsPrefsNote(base)).toBe("Off until you turn it on.");
  });
  it("says when the server's default turned it on", () => {
    expect(statsPrefsNote({ ...base, advanced: true, default: true })).toMatch(/everyone/);
  });
  it("says an admin's switch can be switched off", () => {
    expect(statsPrefsNote({ ...base, advanced: true, choice: "on", set_by: "admin" })).toBe(
      "An admin turned this on for you. You can switch it off.",
    );
  });
  it("says the person's own answer", () => {
    expect(statsPrefsNote({ ...base, choice: "off", set_by: "self", default: true })).toBe(
      "You turned this off.",
    );
  });
});

describe("the admin listing", () => {
  const data: AdminAdvancedStats = {
    default: { on: false, stored: false },
    accounts: [{ user_id: "a", choice: "on", set_by: "admin", advanced: true, updated_at: 5 }],
  };
  it("gives an unlisted account the default", () => {
    expect(accountStatsOf(data, "b")).toEqual({ user_id: "b", choice: "default", advanced: false });
    expect(accountStatsOf(data, "a").set_by).toBe("admin");
  });
  it("replaces one account's line after a switch", () => {
    const next = withAccount(data, {
      user_id: "a",
      choice: "off",
      set_by: "admin",
      advanced: false,
    });
    expect(next.accounts).toHaveLength(1);
    expect(accountStatsOf(next, "a").advanced).toBe(false);
  });
  it("moves only the accounts following the default when it flips", () => {
    const next = withDefault(data, { on: true, stored: true, updated_at: 9 });
    expect(next.default.on).toBe(true);
    expect(accountStatsOf(next, "b").advanced).toBe(true);
    expect(accountStatsOf(next, "a")).toEqual(data.accounts[0]);
  });
  it("words each state", () => {
    expect(accountStatsText({ user_id: "x", choice: "default", advanced: false })).toBe(
      "Default (off)",
    );
    expect(accountStatsText({ user_id: "x", choice: "on", set_by: "admin", advanced: true })).toBe(
      "On, set by an admin",
    );
    expect(accountStatsText({ user_id: "x", choice: "off", set_by: "self", advanced: false })).toBe(
      "Off, set by them",
    );
  });
});
