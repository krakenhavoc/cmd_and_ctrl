// playmat.test.ts — ADR 0128's pure rules: which seats draw a mat under
// the per-device setting, what the board will load, and the setting's
// place in the settings schema.

import { afterEach, describe, expect, it } from "vitest";
import {
  clampCropOrigin,
  cropAxis,
  cropBoxStyle,
  dragCrop,
  fitPromptText,
  fitResultText,
  nudgeCrop,
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

// ---- the best-size suggestion's crop window (ADR 0128 §11) ----

describe("the crop window", () => {
  const tall = { x: 0, y: 125, width: 3000, height: 1750 }; // of a 3000 x 2000 image
  const wide = { x: 423, y: 0, width: 1714, height: 1000 }; // of a 2560 x 1000 image

  it("slides along the one axis being cropped", () => {
    expect(cropAxis(3000, 2000, tall)).toBe("y");
    expect(cropAxis(2560, 1000, wide)).toBe("x");
    expect(cropAxis(2400, 1400, { x: 0, y: 0, width: 2400, height: 1400 })).toBeNull();
  });

  it("is kept inside the image and on its axis", () => {
    expect(clampCropOrigin(3000, 2000, tall, { x: 50, y: -40 })).toEqual({ x: 0, y: 0 });
    expect(clampCropOrigin(3000, 2000, tall, { x: 50, y: 9999 })).toEqual({ x: 0, y: 250 });
    expect(clampCropOrigin(3000, 2000, tall, { x: 0, y: 100.4 })).toEqual({ x: 0, y: 100 });
    expect(clampCropOrigin(2560, 1000, wide, { x: 9999, y: 70 })).toEqual({ x: 846, y: 0 });
    // Nothing to choose: the server's origin stands.
    const whole = { x: 0, y: 0, width: 2400, height: 1400 };
    expect(clampCropOrigin(2400, 1400, whole, { x: 30, y: 30 })).toEqual({ x: 0, y: 0 });
  });

  it("steps by a fraction of the room, at least one pixel, and jumps to the ends", () => {
    expect(nudgeCrop(3000, 2000, tall, { x: 0, y: 125 }, 1)).toEqual({ x: 0, y: 138 });
    expect(nudgeCrop(3000, 2000, tall, { x: 0, y: 125 }, -1, 0.2)).toEqual({ x: 0, y: 75 });
    expect(nudgeCrop(3000, 2000, tall, { x: 0, y: 125 }, "home")).toEqual({ x: 0, y: 0 });
    expect(nudgeCrop(3000, 2000, tall, { x: 0, y: 125 }, "end")).toEqual({ x: 0, y: 250 });
    expect(nudgeCrop(3000, 2000, tall, { x: 0, y: 250 }, 1)).toEqual({ x: 0, y: 250 });
    // A sliver of room still moves.
    const sliver = { x: 0, y: 1, width: 100, height: 58 };
    expect(nudgeCrop(100, 60, sliver, { x: 0, y: 1 }, 1, 0.01)).toEqual({ x: 0, y: 2 });
  });

  it("turns a drag in screen pixels into image pixels", () => {
    // 10 px of a 200 px frame over a 2000 px image is 100 px.
    expect(dragCrop(3000, 2000, tall, { x: 0, y: 125 }, 80, 10, 300, 200)).toEqual({
      x: 0,
      y: 225,
    });
    expect(dragCrop(3000, 2000, tall, { x: 0, y: 125 }, 0, -999, 300, 200)).toEqual({ x: 0, y: 0 });
    // A frame with no size yet moves nothing.
    expect(dragCrop(3000, 2000, tall, { x: 0, y: 125 }, 0, 50, 0, 0)).toEqual({ x: 0, y: 125 });
  });

  it("is placed over the shown image in percent", () => {
    expect(cropBoxStyle(3000, 2000, tall, { x: 0, y: 125 })).toEqual({
      left: "0.000%",
      top: "6.250%",
      width: "100.000%",
      height: "87.500%",
    });
  });
});

describe("the prompt's words", () => {
  it("says the image's size and the best one, as the owner asked", () => {
    expect(fitPromptText(3000, 2000, 2400, 1400)).toBe(
      "This image is 3000\u00d72000. Playmats look best at 2400\u00d71400 (the shape of a paper playmat).",
    );
  });

  it("warns only when the result will be smaller than the best size", () => {
    const big = {
      target_width: 2400,
      target_height: 1400,
      crop: { x: 0, y: 0, width: 3000, height: 1750 },
      smaller: false,
    };
    expect(fitResultText(big)).not.toContain("soft");
    expect(fitResultText(big)).toContain("2400\u00d71400");
    expect(
      fitResultText({ ...big, target_width: 800, target_height: 467, smaller: true }),
    ).toContain("may look soft");
  });
});
