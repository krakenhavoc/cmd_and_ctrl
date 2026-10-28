import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// music.ts keeps module-level state (armed, muted, the Audio element),
// so each test loads a fresh copy of it.
class FakeAudio {
  static instances: FakeAudio[] = [];
  src: string;
  loop = false;
  preload = "";
  volume = 1;
  paused = true;
  play = vi.fn(() => {
    this.paused = false;
    return Promise.resolve();
  });
  pause = vi.fn(() => {
    this.paused = true;
  });
  constructor(src: string) {
    this.src = src;
    FakeAudio.instances.push(this);
  }
}

async function load() {
  vi.resetModules();
  return import("./music");
}

// The suite runs in Node, so window is a bare EventTarget for the gesture listeners.
let win: EventTarget;

function gesture() {
  win.dispatchEvent(new Event("pointerdown"));
}

describe("music", () => {
  beforeEach(() => {
    FakeAudio.instances = [];
    win = new EventTarget();
    vi.stubGlobal("window", win);
    vi.stubGlobal("Audio", FakeAudio);
  });
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("does not create the track before a gesture arms it", async () => {
    const m = await load();
    m.setMusicVolumeMultiplier(0.5);
    expect(FakeAudio.instances).toHaveLength(0);
  });

  it("does not create the track on a gesture while muted", async () => {
    const m = await load();
    m.setMusicVolumeMultiplier(0.5);
    m.setMusicMuted(true);
    m.armMusicOnFirstGesture();
    gesture();
    expect(FakeAudio.instances).toHaveLength(0);
  });

  it("does not create the track on a gesture at volume 0", async () => {
    const m = await load();
    m.setMusicVolumeMultiplier(0);
    m.armMusicOnFirstGesture();
    gesture();
    expect(FakeAudio.instances).toHaveLength(0);
  });

  it("plays on the first gesture when audible", async () => {
    const m = await load();
    m.setMusicVolumeMultiplier(0.5);
    m.armMusicOnFirstGesture();
    gesture();
    expect(FakeAudio.instances).toHaveLength(1);
    const a = FakeAudio.instances[0];
    expect(a.src).toBe("/sounds/ambient_bronze_thunder.mp3");
    expect(a.loop).toBe(true);
    expect(a.volume).toBe(0.5);
    expect(a.play).toHaveBeenCalledTimes(1);
  });

  // ADR 0209 §2: an interface cue turns the music down while it plays.
  it("ducks a playing track and restores it", async () => {
    const m = await load();
    m.setMusicVolumeMultiplier(0.5);
    m.armMusicOnFirstGesture();
    gesture();
    const a = FakeAudio.instances[0];
    m.setMusicDuck(0.5);
    expect(a.volume).toBe(0.25);
    m.setMusicDuck(1);
    expect(a.volume).toBe(0.5);
  });

  it("starts a track that is created while ducked at the ducked volume", async () => {
    const m = await load();
    m.setMusicVolumeMultiplier(0.5);
    m.setMusicDuck(0.5);
    m.armMusicOnFirstGesture();
    gesture();
    expect(FakeAudio.instances[0].volume).toBe(0.25);
  });

  it("creates the track on unmute after arming, then plays", async () => {
    const m = await load();
    m.setMusicVolumeMultiplier(0.5);
    m.setMusicMuted(true);
    m.armMusicOnFirstGesture();
    gesture();
    expect(FakeAudio.instances).toHaveLength(0);
    m.setMusicMuted(false);
    expect(FakeAudio.instances).toHaveLength(1);
    expect(FakeAudio.instances[0].play).toHaveBeenCalledTimes(1);
  });

  it("pauses an audible track when muted", async () => {
    const m = await load();
    m.setMusicVolumeMultiplier(0.5);
    m.armMusicOnFirstGesture();
    gesture();
    m.setMusicMuted(true);
    expect(FakeAudio.instances[0].pause).toHaveBeenCalledTimes(1);
  });
});
