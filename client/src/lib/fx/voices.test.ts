import { describe, expect, it, vi } from "vitest";
import { VOICE_COOLDOWN_MS, VOICE_MANIFEST, VoiceBus, type VoiceManifest } from "./voices";
import { speakCue } from "./voiceWiring";

const A = "walker-a";
const B = "walker-b";
const manifest: VoiceManifest = {
  [A]: { enter: ["/voices/a-enter-1.mp3"], leave: ["/voices/a-leave-1.mp3"] },
  [B]: { enter: ["/voices/b-enter-1.mp3"] },
};

function rig(m: VoiceManifest = manifest) {
  let t = 1_000_000;
  const done: Array<(failed: boolean) => void> = [];
  const played: string[] = [];
  const bus = new VoiceBus({
    manifest: m,
    now: () => t,
    muted: () => false,
    volume: () => 0.5,
    random: () => 0,
    player: (url, _vol, d) => {
      played.push(url);
      done.push(d);
    },
  });
  return {
    bus,
    played,
    finish: (failed = false) => done.shift()?.(failed),
    advance: (ms: number) => {
      t += ms;
    },
  };
}

describe("VoiceBus", () => {
  it("ships an empty manifest, and an empty manifest is a silent no-op", async () => {
    expect(VOICE_MANIFEST).toEqual({});
    const bus = new VoiceBus({
      player: () => {
        throw new Error("must not play");
      },
    });
    expect(bus.empty).toBe(true);
    expect(bus.speak("anything", "enter")).toBe(false);
    const load = vi.fn();
    await speakCue({ name: "Anyone", moment: "enter" }, bus, load);
    expect(load).not.toHaveBeenCalled(); // no catalog fetch either
  });

  it("gives each walker a 20 s cooldown", () => {
    const r = rig();
    expect(r.bus.speak(A, "enter")).toBe(true);
    r.finish();
    r.advance(VOICE_COOLDOWN_MS - 1);
    expect(r.bus.speak(A, "leave")).toBe(false);
    r.advance(1);
    expect(r.bus.speak(A, "leave")).toBe(true);
    expect(r.played).toEqual(["/voices/a-enter-1.mp3", "/voices/a-leave-1.mp3"]);
  });

  it("plays one line at a time across the table", () => {
    const r = rig();
    expect(r.bus.speak(A, "enter")).toBe(true);
    expect(r.bus.speak(B, "enter")).toBe(false); // A is still talking
    r.finish();
    expect(r.bus.speak(B, "enter")).toBe(true);
  });

  it("stays silent for a moment with no take, and when muted", () => {
    const r = rig();
    expect(r.bus.speak(B, "leave")).toBe(false);
    const muted = new VoiceBus({ manifest, muted: () => true, player: () => {} });
    expect(muted.speak(A, "enter")).toBe(false);
  });

  it("treats a missing file as silent, frees the bus and refunds the cooldown", () => {
    const r = rig();
    expect(r.bus.speak(A, "enter")).toBe(true);
    r.finish(true); // 404
    expect(r.bus.speak(B, "enter")).toBe(true); // the lock was released
    r.finish();
    expect(r.bus.speak(A, "enter")).toBe(false); // that file is now known missing
    expect(r.bus.speak(A, "leave")).toBe(true); // the walker is not on cooldown
  });

  it("swallows a player that throws", () => {
    const bus = new VoiceBus({
      manifest,
      player: () => {
        throw new Error("boom");
      },
    });
    expect(() => bus.speak(A, "enter")).not.toThrow();
    expect(bus.speak(B, "enter")).toBe(true); // not stuck busy
  });
});
