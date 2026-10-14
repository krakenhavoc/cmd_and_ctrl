// ADR 0159 §6: what the game bar and the voice panel say about a table's
// voice room, and the two calls that ask the server to change it.

import { afterEach, describe, expect, it, vi } from "vitest";

import type { VoiceView } from "./protocol";
import {
  refreshVoiceRoom,
  setVoiceRoom,
  voiceChip,
  discordAppURL,
  voiceSeatBadge,
  voiceStatusLine,
  voiceYouLine,
} from "./voiceRoom";

const URL = "https://discord.com/channels/1/2";

function v(over: Partial<VoiceView>): VoiceView {
  return { enabled: true, status: "ready", you: "granted", ...over };
}

describe("voiceChip", () => {
  it("is nothing without a room, with voice off, or for a spectator", () => {
    expect(voiceChip(undefined)).toBeNull();
    expect(voiceChip(v({ enabled: false }))).toBeNull();
    expect(voiceChip(v({ you: "spectator", url: URL }))).toBeNull();
  });

  it("links straight into the room for a player who was let in", () => {
    const chip = voiceChip(v({ url: URL }));
    expect(chip?.label).toBe("Join voice");
    expect(chip?.href).toBe(URL);
  });

  it("never links without the URL, even for a granted player", () => {
    expect(voiceChip(v({ status: "pending" }))?.href).toBeUndefined();
    expect(voiceChip(v({ status: "ready" }))?.href).toBeUndefined();
  });

  it("tells a player the step they still have to take", () => {
    expect(voiceChip(v({ you: "not_in_server" }))?.label).toBe("Voice: join the Discord");
    expect(voiceChip(v({ you: "no_discord" }))?.label).toBe("Voice: sign in");
    expect(voiceChip(v({ status: "idle" }))?.label).toBe("Voice: reopen");
    expect(voiceChip(v({ status: "full" }))?.label).toBe("Voice: full");
    expect(voiceChip(v({ status: "unavailable" }))?.label).toBe("Voice: unavailable");
    expect(voiceChip(v({ status: "pending", you: "pending" }))?.label).toBe("Voice: starting");
    for (const chip of [voiceChip(v({ you: "not_in_server" })), voiceChip(v({ status: "idle" }))]) {
      expect(chip?.href).toBeUndefined();
    }
  });
});

describe("voiceStatusLine and voiceYouLine", () => {
  it("word every status", () => {
    for (const status of ["off", "pending", "ready", "idle", "full", "unavailable"] as const) {
      expect(voiceStatusLine(v({ status }))).not.toBe("");
    }
    expect(voiceStatusLine(v({ enabled: false }))).toBe("No voice room for this table.");
  });

  it("say what this viewer can do", () => {
    expect(voiceYouLine(v({ url: URL }))).toMatch(/You're in/);
    expect(voiceYouLine(v({ you: "not_in_server" }))).toMatch(/Join it, then check again/);
    expect(voiceYouLine(v({ you: "no_discord" }))).toMatch(/Sign in with Discord/);
    expect(voiceYouLine(v({ you: "spectator" }))).toBeNull();
    expect(voiceYouLine(v({ enabled: false }))).toBeNull();
  });
});

describe("the voice calls", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("PUT the host's choice and POST a refresh", async () => {
    const f = vi.fn().mockResolvedValue(new Response(null, { status: 202 }));
    vi.stubGlobal("fetch", f);
    await setVoiceRoom("g-1", true);
    await refreshVoiceRoom("g-1");
    expect(f.mock.calls[0][0]).toBe("/games/g-1/voice");
    expect(f.mock.calls[0][1].method).toBe("PUT");
    expect(JSON.parse(f.mock.calls[0][1].body)).toEqual({ enabled: true });
    expect(f.mock.calls[1][0]).toBe("/games/g-1/voice/refresh");
    expect(f.mock.calls[1][1].method).toBe("POST");
  });

  it("surface the server's refusal", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ error: "voice rooms are off on this server" }), {
          status: 503,
        }),
      ),
    );
    await expect(setVoiceRoom("g-1", true)).rejects.toThrow(/voice rooms are off/);
  });
});

describe("voice connection state (#3311)", () => {
  it("counts who is in the room on the chip", () => {
    const chip = voiceChip(v({ url: URL, granted: 4, in_room: 3 }));
    expect(chip?.label).toBe("Voice · 3 of 4");
    expect(voiceChip(v({ url: URL, granted: 4 }))?.label).toBe("Join voice · 0 of 4");
  });

  it("badges a seat that is connected, and one that is muted", () => {
    expect(voiceSeatBadge("live")?.glyph).toBe("🎧");
    expect(voiceSeatBadge("muted")?.glyph).toBe("🔇");
    expect(voiceSeatBadge(undefined)).toBeNull();
  });

  it("builds the desktop app link from the room link only", () => {
    expect(discordAppURL("https://discord.com/channels/1/2")).toBe("discord://-/channels/1/2");
    expect(discordAppURL("https://evil.example/channels/1/2")).toBeNull();
    expect(discordAppURL("https://discord.com/channels/1/2?x=y")).toBeNull();
  });
});
