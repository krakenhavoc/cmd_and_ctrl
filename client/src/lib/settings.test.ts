import { describe, it, expect, beforeEach, vi } from "vitest";
import { get } from "svelte/store";
import { stopKeyFor, hasOwnStop } from "./turn";

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

  it("v8 → v9 seeds showBotReasoning off without touching the rest", async () => {
    // A blob from just before S31 sub-PR 8: everything the v8 schema
    // had, and no bot settings, because bot seats did not exist.
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({
        __version: 8,
        gameplay: { adminOverrides: true, strictMana: true, autoPassPriority: false },
        display: { cardSize: "large" },
      }),
    );
    const { settings, SETTINGS_VERSION } = await freshModule();
    const s = get(settings);
    expect(s.__version).toBe(SETTINGS_VERSION);
    // The new field arrives at its default...
    expect(s.gameplay.showBotReasoning).toBe(false);
    // ...and the migration rescues nothing and breaks nothing.
    expect(s.gameplay.adminOverrides).toBe(true);
    expect(s.gameplay.strictMana).toBe(true);
    expect(s.display.cardSize).toBe("large");
  });

  it("v9 → v10 seeds the shortcuts section with the keymap enabled and no overrides", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({
        __version: 9,
        gameplay: { showBotReasoning: true, adminOverrides: true },
        display: { cardSize: "small" },
      }),
    );
    const { settings, SETTINGS_VERSION } = await freshModule();
    const s = get(settings);
    expect(s.__version).toBe(SETTINGS_VERSION);
    expect(s.shortcuts.enabled).toBe(true);
    // No overrides: a v9 user has never expressed an opinion about a
    // key, so every row resolves against today's default and a future
    // retune reaches them.
    expect(s.shortcuts.bindings).toEqual({});
    // Nothing else moved.
    expect(s.gameplay.showBotReasoning).toBe(true);
    expect(s.gameplay.adminOverrides).toBe(true);
    expect(s.display.cardSize).toBe("small");
  });

  it("v10 keeps an explicit rebind and an explicit unbind across a load", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({
        __version: 10,
        shortcuts: { enabled: false, bindings: { undo: "z", drawCard: "" } },
      }),
    );
    const { settings } = await freshModule();
    const s = get(settings);
    expect(s.shortcuts.enabled).toBe(false);
    expect(s.shortcuts.bindings).toEqual({ undo: "z", drawCard: "" });
  });

  it("v10 scrubs a hostile or stale bindings blob without dropping the good rows", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({
        __version: 10,
        shortcuts: {
          enabled: true,
          bindings: {
            undo: "CTRL+Z", // normalised
            passPriority: "Escape", // reserved — refused
            retiredAction: "q", // no such action any more
            passTurn: "Hyper+k", // unparseable
            openSettings: ",", // equal to the default — not a choice
          },
        },
      }),
    );
    const { settings } = await freshModule();
    expect(get(settings).shortcuts.bindings).toEqual({ undo: "Ctrl+z" });
  });

  it("v10 treats a non-object bindings value as no overrides rather than throwing", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 10, shortcuts: { enabled: true, bindings: "nope" } }),
    );
    const { settings } = await freshModule();
    expect(get(settings).shortcuts.bindings).toEqual({});
  });

  it("falls back to defaults when stored blob is corrupt", async () => {
    localStorage.setItem("cmdctrl.settings.v1", "{not valid json");
    const { settings, SETTINGS_VERSION } = await freshModule();
    // No exception; defaults loaded.
    expect(get(settings).__version).toBe(SETTINGS_VERSION);
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
    const { settings, updateSettings, resetSettings, SETTINGS_VERSION } = await freshModule();
    updateSettings("audio", "muted", true);
    expect(get(settings).audio.muted).toBe(true);
    resetSettings();
    expect(get(settings).audio.muted).toBe(false);
    expect(get(settings).__version).toBe(SETTINGS_VERSION);
  });

  it("export → import is a round-trip", async () => {
    const { settings, updateSettings, exportSettings, importSettings } = await freshModule();
    // Mutate so the exported blob carries non-default values.
    updateSettings("audio", "muted", true);
    updateSettings("display", "cardSize", "large");
    updateSettings("accessibility", "textScale", 1.5);
    const exported = exportSettings();

    // Reset and confirm the values are gone, then re-import.
    updateSettings("audio", "muted", false);
    updateSettings("display", "cardSize", "small");
    updateSettings("accessibility", "textScale", 0.9);

    const res = importSettings(exported);
    expect(res.ok).toBe(true);
    expect(res.changed).toBe(true);
    expect(get(settings).audio.muted).toBe(true);
    expect(get(settings).display.cardSize).toBe("large");
    expect(get(settings).accessibility.textScale).toBe(1.5);
  });

  it("importSettings rejects non-JSON and non-object blobs", async () => {
    const { settings, importSettings } = await freshModule();
    const before = JSON.stringify(get(settings));

    expect(importSettings("not json").ok).toBe(false);
    expect(importSettings("[1,2,3]").ok).toBe(false);
    expect(importSettings("null").ok).toBe(false);
    expect(importSettings('"just a string"').ok).toBe(false);

    // Store untouched after every rejection.
    expect(JSON.stringify(get(settings))).toBe(before);
  });

  it("importSettings of identical state reports changed: false", async () => {
    const { exportSettings, importSettings } = await freshModule();
    const blob = exportSettings();
    const res = importSettings(blob);
    expect(res.ok).toBe(true);
    expect(res.changed).toBe(false);
  });

  it("fingerprintSettings is deterministic for the same state and changes when state does", async () => {
    const { updateSettings, fingerprintSettings } = await freshModule();
    const a1 = fingerprintSettings();
    const a2 = fingerprintSettings();
    expect(a1).toBe(a2);
    updateSettings("audio", "muted", true);
    const b = fingerprintSettings();
    expect(b).not.toBe(a1);
  });
});

// ---- #717: the two combat damage steps share one stop ----

describe("per-step stops and the first-strike damage step", () => {
  it("gives the first-strike damage step no row of its own", async () => {
    const { defaultStepStops } = await freshModule();
    const stops = defaultStepStops();
    expect(Object.keys(stops)).not.toContain("first_strike_damage");
    expect(Object.keys(stops)).toContain("combat_damage");
    // Untap and Cleanup are out for the older reason: no priority.
    expect(Object.keys(stops)).not.toContain("untap");
    expect(Object.keys(stops)).not.toContain("cleanup");
  });

  it("routes the first-strike damage step's stop to combat_damage", () => {
    expect(stopKeyFor("first_strike_damage")).toBe("combat_damage");
    expect(stopKeyFor("combat_damage")).toBe("combat_damage");
    expect(stopKeyFor("declare_blockers")).toBe("declare_blockers");
    expect(hasOwnStop("first_strike_damage")).toBe(false);
    expect(hasOwnStop("combat_damage")).toBe(true);
    expect(hasOwnStop("untap")).toBe(false);
  });
});
