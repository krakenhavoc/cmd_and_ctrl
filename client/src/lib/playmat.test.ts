// playmat.test.ts — ADR 0128's pure rules: which seats draw a mat under
// the per-device setting, what the board will load, and the setting's
// place in the settings schema.

import { afterEach, describe, expect, it } from "vitest";
import {
  DEFAULT_PLAYMATS_MODE,
  PLAYMATS_MODES,
  isPlaymatPath,
  isPlaymatsMode,
  playmatShownFor,
  playmatSrc,
} from "./playmat";
import {
  SYNCED_FIELDS,
  defaultSettings,
  importSettings,
  exportSettings,
  settings,
} from "./settings";
import { session, type Session } from "./session";
import { get } from "svelte/store";

const URL1 = "/playmats/6f1c2a9e-1b2c-4d3e-8f40-0123456789ab";

afterEach(() => {
  session.set(null);
});

describe("isPlaymatPath", () => {
  it("accepts only the route the server mints", () => {
    expect(isPlaymatPath(URL1)).toBe(true);
    for (const bad of [
      undefined,
      "",
      "https://evil.example/mat.png",
      "//evil.example/playmats/6f1c2a9e-1b2c-4d3e-8f40-0123456789ab",
      "/playmats/../etc/passwd",
      "/playmats/6f1c2a9e-1b2c-4d3e-8f40-0123456789ab/extra",
      "/playmats/6f1c2a9e-1b2c-4d3e-8f40-0123456789ab?x=1",
      "/playmats/6F1C2A9E-1B2C-4D3E-8F40-0123456789AB",
      "/avatars/1/2.png",
      "javascript:alert(1)",
      42,
    ]) {
      expect(isPlaymatPath(bad), String(bad)).toBe(false);
    }
  });
});

describe("playmatShownFor", () => {
  const seat = { playmat_url: URL1 };
  it("all draws every seat's mat", () => {
    expect(playmatShownFor("all", seat, true)).toBe(URL1);
    expect(playmatShownFor("all", seat, false)).toBe(URL1);
  });
  it("mine draws only the viewer's own", () => {
    expect(playmatShownFor("mine", seat, true)).toBe(URL1);
    expect(playmatShownFor("mine", seat, false)).toBeNull();
  });
  it("off draws none", () => {
    expect(playmatShownFor("off", seat, true)).toBeNull();
    expect(playmatShownFor("off", seat, false)).toBeNull();
  });
  it("draws nothing for a seat with none, or with a URL that is not ours", () => {
    expect(playmatShownFor("all", {}, true)).toBeNull();
    expect(playmatShownFor("all", { playmat_url: "https://evil.example/x.png" }, false)).toBeNull();
  });
});

describe("playmatSrc", () => {
  const signedIn = {
    token: "t0k/en+1",
    expiresAt: "2999-01-01T00:00:00Z",
    principal: { role: "identified" },
  } as unknown as Session;

  it("is the path itself with no session", () => {
    expect(playmatSrc(URL1)).toBe(URL1);
  });
  it("carries the session token, as an avatar does, because an <img> cannot send a header", () => {
    session.set(signedIn);
    expect(playmatSrc(URL1)).toBe(`${URL1}?token=${encodeURIComponent("t0k/en+1")}`);
  });
  it("refuses a URL that is not a playmat path", () => {
    session.set(signedIn);
    expect(playmatSrc("https://evil.example/x.png")).toBeNull();
    expect(playmatSrc(undefined)).toBeNull();
  });
});

describe("display.playmats in the settings schema", () => {
  it("defaults to all and is per device, never synced", () => {
    expect(defaultSettings().display.playmats).toBe(DEFAULT_PLAYMATS_MODE);
    expect(DEFAULT_PLAYMATS_MODE).toBe("all");
    expect(SYNCED_FIELDS.display.playmats).toBe("device");
  });
  it("knows exactly three modes", () => {
    expect([...PLAYMATS_MODES]).toEqual(["all", "mine", "off"]);
    expect(isPlaymatsMode("mine")).toBe(true);
    expect(isPlaymatsMode("everyone")).toBe(false);
    expect(isPlaymatsMode(undefined)).toBe(false);
  });
  it("a stored blob from before playmats, or with a bad value, reads as the default", () => {
    const prior = JSON.parse(exportSettings());
    delete prior.display.playmats;
    importSettings(JSON.stringify(prior));
    expect(get(settings).display.playmats).toBe("all");

    prior.display.playmats = "banana";
    importSettings(JSON.stringify(prior));
    expect(get(settings).display.playmats).toBe("all");

    prior.display.playmats = "off";
    importSettings(JSON.stringify(prior));
    expect(get(settings).display.playmats).toBe("off");
  });
});
