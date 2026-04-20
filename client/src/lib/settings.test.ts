import { describe, it, expect, beforeEach, vi } from "vitest";
import { get } from "svelte/store";

// vitest runs in node by default and the client isn't configured
// for jsdom. settings.ts only needs a minimal localStorage
// polyfill — it guards window/matchMedia access, so no DOM mock
// is required. Install once at module load so the first static
// import of ./settings inside freshModule() already sees storage.
class MemoryStorage {
  private store = new Map<string, string>();
  get length() {
    return this.store.size;
  }
  clear() {
    this.store.clear();
  }
  getItem(k: string) {
    return this.store.has(k) ? this.store.get(k)! : null;
  }
  key(i: number) {
    return Array.from(this.store.keys())[i] ?? null;
  }
  removeItem(k: string) {
    this.store.delete(k);
  }
  setItem(k: string, v: string) {
    this.store.set(k, String(v));
  }
}
if (typeof globalThis.localStorage === "undefined") {
  Object.defineProperty(globalThis, "localStorage", {
    value: new MemoryStorage(),
    writable: true,
  });
}

// freshModule re-imports settings.ts so the module-scope load
// path (reading localStorage + installing subscribers) runs with
// the current localStorage state. Required because the store is
// a module singleton; without reset, every test shares one store.
async function freshModule() {
  vi.resetModules();
  return await import("./settings");
}

describe("settings", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("loads defaults when nothing is stored", async () => {
    const { settings, defaultSettings } = await freshModule();
    const s = get(settings);
    const d = defaultSettings();
    // Comparing field-for-field rather than by reference because
    // the prefers-reduced-motion branch may differ per run; check
    // only the deterministic shape keys.
    expect(s.__version).toBe(d.__version);
    expect(s.audio.masterVolume).toBe(80);
    expect(s.display.cardSize).toBe("medium");
    expect(s.gameplay.confirmExit).toBe(true);
  });

  it("absorbs the legacy cmdctrl.muted=1 key on first load", async () => {
    localStorage.setItem("cmdctrl.muted", "1");
    const { settings } = await freshModule();
    expect(get(settings).audio.muted).toBe(true);
    // One-shot migration: legacy key gone, unified key written.
    expect(localStorage.getItem("cmdctrl.muted")).toBeNull();
    expect(localStorage.getItem("cmdctrl.settings.v1")).toBeTruthy();
  });

  it("absorbs the legacy cmdctrl.muted=0 as not muted", async () => {
    // Explicit unmuted (the key exists but is "0"). Confirms we
    // distinguish "legacy user chose unmute" from "legacy key absent".
    localStorage.setItem("cmdctrl.muted", "0");
    const { settings } = await freshModule();
    expect(get(settings).audio.muted).toBe(false);
    expect(localStorage.getItem("cmdctrl.muted")).toBeNull();
  });

  it("fills missing fields from defaults when stored blob is partial", async () => {
    // Simulate a v1 blob from a previous build that only wrote audio.
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 1, audio: { muted: true } }),
    );
    const { settings } = await freshModule();
    const s = get(settings);
    expect(s.audio.muted).toBe(true);
    // Rest come from defaults.
    expect(s.audio.masterVolume).toBe(80);
    expect(s.display.cardSize).toBe("medium");
  });

  it("falls back to defaults when stored blob is corrupt", async () => {
    localStorage.setItem("cmdctrl.settings.v1", "{not valid json");
    const { settings } = await freshModule();
    // No exception; defaults loaded.
    expect(get(settings).__version).toBe(1);
  });

  it("updateSettings does a shallow path merge", async () => {
    const { settings, updateSettings } = await freshModule();
    updateSettings("audio", "masterVolume", 42);
    expect(get(settings).audio.masterVolume).toBe(42);
    // Sibling keys in the same group are preserved.
    expect(get(settings).audio.effectsVolume).toBe(100);
    // Other groups are untouched.
    expect(get(settings).display.cardSize).toBe("medium");
  });

  it("persists changes to localStorage", async () => {
    const { settings, updateSettings } = await freshModule();
    updateSettings("display", "cardSize", "large");
    const raw = localStorage.getItem("cmdctrl.settings.v1");
    expect(raw).not.toBeNull();
    const parsed = JSON.parse(raw!);
    expect(parsed.display.cardSize).toBe("large");
    // Live store matches.
    expect(get(settings).display.cardSize).toBe("large");
  });

  it("resetSettings restores defaults without losing the schema version", async () => {
    const { settings, updateSettings, resetSettings } = await freshModule();
    updateSettings("audio", "muted", true);
    expect(get(settings).audio.muted).toBe(true);
    resetSettings();
    expect(get(settings).audio.muted).toBe(false);
    expect(get(settings).__version).toBe(1);
  });
});
