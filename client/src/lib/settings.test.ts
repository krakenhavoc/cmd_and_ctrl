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
    expect(s.display.stackStyle).toBe("pile");
    expect(s.gameplay.confirmExit).toBe(true);
  });

  // #956 / ADR 0077. Pinned as its own test rather than folded into
  // the shape check above, because these three are the defaults that
  // decide what a new player's table LOOKS like, and the ADR's
  // consequences section is written assuming exactly these values.
  // If one of them flips, that document is wrong and should move with
  // the code.
  // #2433: full opponent boards, the overlay and the row layout.
  it("defaults an opponent's board to full, expansion to the overlay, and the table to a row", async () => {
    const { defaultSettings } = await freshModule();
    const d = defaultSettings();
    expect(d.display.opponentDetail).toBe("full");
    expect(d.display.expandActivePlayer).toBe(true);
    expect(d.display.expandStyle).toBe("overlay");
    expect(d.display.tableLayout).toBe("row");
  });

  // ADR 0121 §7: dice animate by default, and a stored blob from before
  // the field existed gains it from the merge, with no version bump.
  it("defaults animations.dice on and fills it into an older blob", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 17, animations: { enabled: true, speed: 1.5, cardDraw: false } }),
    );
    const { settings, defaultSettings, SYNCED_FIELDS } = await freshModule();
    expect(defaultSettings().animations.dice).toBe(true);
    const s = get(settings);
    expect(s.animations.dice).toBe(true);
    expect(s.animations.cardDraw).toBe(false);
    expect(s.animations.speed).toBe(1.5);
    expect(SYNCED_FIELDS.animations.dice).toBe("synced");
  });

  // ADR 0134 §5: combat motion is on by default, synced, and an older
  // blob gains it from the merge, with no version bump.
  it("defaults animations.combat on, syncs it, and fills it into an older blob", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 21, animations: { enabled: true, dice: false } }),
    );
    const { settings, defaultSettings, SYNCED_FIELDS } = await freshModule();
    expect(defaultSettings().animations.combat).toBe(true);
    const s = get(settings);
    expect(s.animations.combat).toBe(true);
    expect(s.animations.dice).toBe(false);
    expect(SYNCED_FIELDS.animations.combat).toBe("synced");
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

  // #956 / ADR 0077. Unlike the eight migrations before it, this one
  // changes what the table LOOKS like for a user who has touched
  // nothing: opponentDetail arrives as "summary" and their opponents
  // stop being drawn as cards. The test says so explicitly so nobody
  // later "fixes" the default thinking it was an oversight.
  it("v10 → v11 seeds the opponent-detail settings without disturbing stored choices", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({
        __version: 10,
        display: { cardSize: "large", tableLayout: "row", showOpponentHandCount: false },
        gameplay: { adminOverrides: true },
        shortcuts: { enabled: true, bindings: { undo: "z" } },
      }),
    );
    const { settings, SETTINGS_VERSION } = await freshModule();
    const s = get(settings);
    expect(s.__version).toBe(SETTINGS_VERSION);
    // The three new fields arrive at their defaults (today's: #2433).
    expect(s.display.opponentDetail).toBe("full");
    expect(s.display.expandActivePlayer).toBe(true);
    expect(s.display.expandStyle).toBe("overlay");
    // Everything the user had actually chosen is untouched.
    expect(s.display.cardSize).toBe("large");
    expect(s.display.tableLayout).toBe("row");
    expect(s.display.showOpponentHandCount).toBe(false);
    expect(s.gameplay.adminOverrides).toBe(true);
    expect(s.shortcuts.bindings).toEqual({ undo: "z" });
  });

  // #2209: art tiles on the battlefield by default, full cards in the
  // hand, and each choice survives a reload.
  it("battlefieldArt defaults on, handArt off, and both persist once changed", async () => {
    let m = await freshModule();
    expect(get(m.settings).display.battlefieldArt).toBe(true);
    expect(get(m.settings).display.handArt).toBe(false);
    expect(m.defaultSettings().display.battlefieldArt).toBe(true);
    expect(m.defaultSettings().display.handArt).toBe(false);
    m.updateSettings("display", "battlefieldArt", false);
    m.updateSettings("display", "handArt", true);
    m = await freshModule();
    expect(get(m.settings).display.battlefieldArt).toBe(false);
    expect(get(m.settings).display.handArt).toBe(true);
  });

  // #2209 migration. A v15 blob always carries artOnlyCards, default
  // included, so only `true` is known to be a choice.
  it("v16 keeps a v15 artOnlyCards: true as art in both places", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 15, display: { artOnlyCards: true, cardSize: "large" } }),
    );
    const { settings, SETTINGS_VERSION } = await freshModule();
    const s = get(settings);
    expect(s.__version).toBe(SETTINGS_VERSION);
    expect(s.display.battlefieldArt).toBe(true);
    expect(s.display.handArt).toBe(true);
    expect(s.display.cardSize).toBe("large");
    expect("artOnlyCards" in s.display).toBe(false);
  });

  it("v16 gives a v15 artOnlyCards: false (or none) the new defaults", async () => {
    for (const display of [{ artOnlyCards: false }, {}]) {
      localStorage.clear();
      localStorage.setItem("cmdctrl.settings.v1", JSON.stringify({ __version: 15, display }));
      const { settings } = await freshModule();
      const s = get(settings);
      expect(s.display.battlefieldArt).toBe(true);
      expect(s.display.handArt).toBe(false);
      expect("artOnlyCards" in s.display).toBe(false);
    }
  });

  it("from v16 on, the stored art choices are honoured", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 16, display: { battlefieldArt: false, handArt: true } }),
    );
    const { settings } = await freshModule();
    const s = get(settings);
    expect(s.display.battlefieldArt).toBe(false);
    expect(s.display.handArt).toBe(true);
  });

  it("an account copy from a v15 client migrates the same way (ADR 0110 §4)", async () => {
    const m = await freshModule();
    const base = m.defaultSettings();
    base.display.battlefieldArt = false;
    const on = m.applySyncedCopy(base, { display: { artOnlyCards: true } }, 15);
    expect(on.display.battlefieldArt).toBe(true);
    expect(on.display.handArt).toBe(true);
    const off = m.applySyncedCopy(base, { display: { artOnlyCards: false } }, 15);
    expect(off.display.battlefieldArt).toBe(true);
    expect(off.display.handArt).toBe(false);
    const v16 = m.applySyncedCopy(
      m.defaultSettings(),
      { display: { battlefieldArt: false, handArt: true } },
      16,
    );
    expect(v16.display.battlefieldArt).toBe(false);
    expect(v16.display.handArt).toBe(true);
    expect(m.syncedSubset(v16).display).not.toHaveProperty("artOnlyCards");
  });

  it("v11 keeps an explicit opponentDetail: full across a load", async () => {
    // The escape hatch has to survive a reload, or a player who
    // deliberately went back to full boards gets summaries again on
    // next launch — which would read as the setting not working.
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({
        __version: 11,
        display: { opponentDetail: "full", expandActivePlayer: false, expandStyle: "overlay" },
      }),
    );
    const { settings } = await freshModule();
    const s = get(settings);
    expect(s.display.opponentDetail).toBe("full");
    expect(s.display.expandActivePlayer).toBe(false);
    expect(s.display.expandStyle).toBe("overlay");
  });

  // #1307. The new fields arrive at their defaults, and an existing
  // smartAutoPass: false is left alone — turning smart autopass off is
  // how a player keeps the old "every opponent stack item stops".
  it("v11 → v12 seeds the response categories without disturbing stored choices", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({
        __version: 11,
        gameplay: { smartAutoPass: false, autoPassOwnStack: false },
      }),
    );
    const { settings, SETTINGS_VERSION } = await freshModule();
    const s = get(settings);
    expect(s.__version).toBe(SETTINGS_VERSION);
    expect(s.gameplay.respondCounterspells).toBe(true);
    expect(s.gameplay.respondInstants).toBe(true);
    expect(s.gameplay.respondAbilities).toBe(true);
    expect(s.gameplay.respondSpecialActions).toBe(true);
    // v24 (ADR 0143): smart auto-pass off is now Careful.
    expect(s.gameplay.passMode).toBe("careful");
    expect(s.gameplay.autoPassOwnStack).toBe(false);
  });

  it("v12 → v13 seeds the bluff settings off, and keeps a stored choice", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 12, gameplay: { respondInstants: false } }),
    );
    let mod = await freshModule();
    let s = get(mod.settings);
    expect(s.__version).toBe(mod.SETTINGS_VERSION);
    expect(s.gameplay.bluffCounterspell).toBe(false);
    expect(s.gameplay.bluffInstant).toBe(false);
    expect(s.gameplay.bluffMode).toBe("timed");
    expect(s.gameplay.bluffDelayMinMs).toBe(1500);
    expect(s.gameplay.bluffDelayMaxMs).toBe(4000);
    expect(s.gameplay.respondInstants).toBe(false);

    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({
        __version: 13,
        gameplay: { bluffInstant: true, bluffMode: "manual", bluffDelayMaxMs: 6000 },
      }),
    );
    mod = await freshModule();
    s = get(mod.settings);
    expect(s.gameplay.bluffInstant).toBe(true);
    expect(s.gameplay.bluffMode).toBe("manual");
    expect(s.gameplay.bluffDelayMaxMs).toBe(6000);
  });

  // ADR 0119 §2: the stack hold arrives on, at 2 s, for everyone.
  it("v17 → v18 seeds the stack hold at 2 s, and keeps a stored choice", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 17, gameplay: { bluffInstant: true } }),
    );
    let mod = await freshModule();
    let s = get(mod.settings);
    expect(s.__version).toBe(mod.SETTINGS_VERSION);
    expect(mod.SETTINGS_VERSION).toBeGreaterThanOrEqual(18);
    expect(s.gameplay.stackHoldMs).toBe(2000);
    expect(s.gameplay.bluffInstant).toBe(true);

    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 18, gameplay: { stackHoldMs: 0 } }),
    );
    mod = await freshModule();
    s = get(mod.settings);
    expect(s.gameplay.stackHoldMs).toBe(0);
  });

  // ADR 0118 §1, owner decision 5: strict payment is the default, and
  // every existing player is moved to it once. A stored false from
  // before v19 is the old default materialised, so it moves; from v19
  // on it is a choice and is kept.
  it("defaults strictMana on for a new player", async () => {
    const { settings, defaultSettings } = await freshModule();
    expect(defaultSettings().gameplay.strictMana).toBe(true);
    expect(get(settings).gameplay.strictMana).toBe(true);
  });

  it("v18 → v19 writes strictMana true whatever was stored", async () => {
    for (const version of [4, 15, 18]) {
      for (const stored of [false, true, "no", undefined]) {
        localStorage.setItem(
          "cmdctrl.settings.v1",
          JSON.stringify({
            __version: version,
            gameplay: { strictMana: stored, smartAutoPass: false },
          }),
        );
        const { settings, SETTINGS_VERSION } = await freshModule();
        const s = get(settings);
        expect(SETTINGS_VERSION).toBeGreaterThanOrEqual(21);
        expect(s.__version).toBe(SETTINGS_VERSION);
        expect(s.gameplay.strictMana, `v${version} stored ${String(stored)}`).toBe(true);
        expect(s.gameplay.passMode).toBe("careful");
      }
    }
  });

  it("v19 keeps a player's choice to turn strict payment off", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 19, gameplay: { strictMana: false } }),
    );
    const mod = await freshModule();
    expect(get(mod.settings).gameplay.strictMana).toBe(false);
    // It survives the save and the next load: the move happens once.
    mod.updateSettings("gameplay", "confirmExit", false);
    const again = await freshModule();
    expect(get(again.settings).gameplay.strictMana).toBe(false);
  });

  // v19 → v20: display.theme names a skin. The old "dark" (the only
  // value the disabled select ever stored) and anything unknown become
  // the default skin; a real skin id is kept.
  it("v20 maps the old dark theme and unknown values to the default skin", async () => {
    for (const stored of ["dark", "neon", 7, undefined]) {
      localStorage.setItem(
        "cmdctrl.settings.v1",
        JSON.stringify({ __version: 19, display: { theme: stored } }),
      );
      const { settings } = await freshModule();
      expect(get(settings).display.theme, String(stored)).toBe("warroom");
    }
    for (const stored of ["classic", "light", "high-contrast", "prism"]) {
      localStorage.setItem(
        "cmdctrl.settings.v1",
        JSON.stringify({ __version: 19, display: { theme: stored } }),
      );
      const { settings } = await freshModule();
      expect(get(settings).display.theme).toBe(stored);
    }
  });

  it("keeps a #rrggbb accent, lower-cased, and drops anything else", async () => {
    for (const [stored, want] of [
      ["#FF7A1A", "#ff7a1a"],
      ["#abc", ""],
      ["orange", ""],
      [42, ""],
      [undefined, ""],
    ] as const) {
      localStorage.setItem(
        "cmdctrl.settings.v1",
        JSON.stringify({ __version: 20, display: { accent: stored } }),
      );
      const { settings } = await freshModule();
      expect(get(settings).display.accent, String(stored)).toBe(want);
    }
  });

  it("moves a synced copy written before v19 to strict, and keeps a v19 copy's off", async () => {
    const m = await freshModule();
    const base = { ...m.defaultSettings(), gameplay: { ...m.defaultSettings().gameplay } };
    base.gameplay.strictMana = false;
    for (const version of [15, 18]) {
      const moved = m.applySyncedCopy(base, { gameplay: { strictMana: false } }, version);
      expect(moved.gameplay.strictMana, `v${version} copy`).toBe(true);
    }
    const kept = m.applySyncedCopy(base, { gameplay: { strictMana: false } }, 19);
    expect(kept.gameplay.strictMana).toBe(false);
  });

  it("v19 falls back to the default for a strictMana that is not a boolean", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 19, gameplay: { strictMana: "off" } }),
    );
    const { settings } = await freshModule();
    expect(get(settings).gameplay.strictMana).toBe(true);
  });

  // #1467 seeded stackStyle at v14; since v17 the seeded value is the
  // pile (ADR 0119 §1), and nothing else stored moves.
  it("a v13 blob gets the default stack style without disturbing stored choices", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({
        __version: 13,
        display: { tableLayout: "row", handLayout: "stacked" },
        gameplay: { bluffInstant: true },
      }),
    );
    const { settings, SETTINGS_VERSION } = await freshModule();
    const s = get(settings);
    expect(s.__version).toBe(SETTINGS_VERSION);
    expect(s.display.stackStyle).toBe("pile");
    expect(s.display.tableLayout).toBe("row");
    expect(s.display.handLayout).toBe("stacked");
    expect(s.gameplay.bluffInstant).toBe(true);
  });

  // ADR 0105 owner decision 4: the legal-action highlights start on
  // for everyone, existing players included, whatever the stored blob
  // says — and from v15 on the player's own choice is kept.
  // #2336: tableLayout gained "focus" without a version bump, and from
  // then on the value is checked: an unknown one is the quadrant.
  it("keeps every table layout and resets an unknown one to the default row", async () => {
    for (const [stored, want] of [
      ["quadrant", "quadrant"],
      ["row", "row"],
      ["focus", "focus"],
      ["sideways", "row"],
      [7, "row"],
    ] as const) {
      localStorage.setItem(
        "cmdctrl.settings.v1",
        JSON.stringify({ __version: 20, display: { tableLayout: stored, cardSize: "large" } }),
      );
      const { settings } = await freshModule();
      const s = get(settings);
      expect(s.display.tableLayout, `stored ${String(stored)}`).toBe(want);
      expect(s.display.cardSize).toBe("large");
    }
  });

  it("v14 → v15 writes highlightLegalActions true whatever was stored", async () => {
    for (const stored of [false, true, "no", undefined]) {
      localStorage.setItem(
        "cmdctrl.settings.v1",
        JSON.stringify({
          __version: 14,
          display: { stackStyle: "fan" },
          gameplay: { highlightLegalActions: stored, smartAutoPass: false },
        }),
      );
      const { settings, SETTINGS_VERSION } = await freshModule();
      const s = get(settings);
      expect(s.__version).toBe(SETTINGS_VERSION);
      expect(s.gameplay.highlightLegalActions, `stored ${String(stored)}`).toBe(true);
      expect(s.display.stackStyle).toBe("fan");
      expect(s.gameplay.passMode).toBe("careful");
    }
  });

  it("v15 keeps a player's choice to turn the highlights off", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 15, gameplay: { highlightLegalActions: false } }),
    );
    const { settings } = await freshModule();
    expect(get(settings).gameplay.highlightLegalActions).toBe(false);
  });

  it("defaults highlightLegalActions on for a new player", async () => {
    const { settings } = await freshModule();
    expect(get(settings).gameplay.highlightLegalActions).toBe(true);
  });

  it("keeps a chosen fan, spotlight or ribbon across the v17 upgrade", async () => {
    for (const version of [14, 15, 16]) {
      for (const style of ["fan", "spotlight", "ribbon"] as const) {
        localStorage.setItem(
          "cmdctrl.settings.v1",
          JSON.stringify({ __version: version, display: { stackStyle: style } }),
        );
        const { settings } = await freshModule();
        expect(get(settings).display.stackStyle, `v${version} ${style}`).toBe(style);
      }
    }
  });

  // ADR 0119 §1: before v17 an untouched compact and a chosen one look
  // the same (the shallow merge wrote the default in), so a stored
  // compact moves to the new default. From v17 on it is a choice.
  it("v16 → v17 moves a stored compact to the pile", async () => {
    for (const version of [14, 15, 16]) {
      localStorage.setItem(
        "cmdctrl.settings.v1",
        JSON.stringify({ __version: version, display: { stackStyle: "compact" } }),
      );
      const { settings } = await freshModule();
      expect(get(settings).display.stackStyle, `v${version}`).toBe("pile");
    }
  });

  it("keeps a compact chosen at v17", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 17, display: { stackStyle: "compact" } }),
    );
    const { settings } = await freshModule();
    expect(get(settings).display.stackStyle).toBe("compact");
  });

  it("keeps every style across a load from v17", async () => {
    for (const style of ["pile", "compact", "fan", "spotlight", "ribbon"] as const) {
      localStorage.setItem(
        "cmdctrl.settings.v1",
        JSON.stringify({ __version: 17, display: { stackStyle: style } }),
      );
      const { settings } = await freshModule();
      expect(get(settings).display.stackStyle).toBe(style);
    }
  });

  it("falls back to the pile for a style it does not know", async () => {
    // A style that was tried and removed, or a hand-edited blob: the
    // board must still draw a stack, and the default is the one that
    // is always there (ADR 0119 §1).
    for (const version of [14, 17]) {
      for (const bad of ["carousel", 7, null]) {
        localStorage.setItem(
          "cmdctrl.settings.v1",
          JSON.stringify({ __version: version, display: { stackStyle: bad } }),
        );
        const { settings } = await freshModule();
        expect(get(settings).display.stackStyle).toBe("pile");
      }
    }
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

// ---- #2853: untargeted abilities split out of respondAbilities ----

describe("respondUntargetedAbilities (v21 → v22)", () => {
  beforeEach(() => localStorage.clear());

  const load = async (blob: unknown) => {
    localStorage.setItem("cmdctrl.settings.v1", JSON.stringify(blob));
    const mod = await freshModule();
    return { mod, s: get(mod.settings) };
  };

  it("is off for a new player, and the other four categories are on", async () => {
    const { defaultSettings } = await freshModule();
    const gp = defaultSettings().gameplay;
    expect(gp.respondUntargetedAbilities).toBe(false);
    expect(gp.respondCounterspells).toBe(true);
    expect(gp.respondInstants).toBe(true);
    expect(gp.respondAbilities).toBe(true);
    expect(gp.respondSpecialActions).toBe(true);
  });

  it("moves a player still on the old all-on default to the new default", async () => {
    const { mod, s } = await load({
      __version: 21,
      gameplay: {
        respondCounterspells: true,
        respondInstants: true,
        respondAbilities: true,
        respondSpecialActions: true,
        smartAutoPass: true,
      },
    });
    expect(mod.SETTINGS_VERSION).toBeGreaterThanOrEqual(22);
    expect(s.__version).toBe(mod.SETTINGS_VERSION);
    expect(s.gameplay.respondAbilities).toBe(true);
    expect(s.gameplay.respondUntargetedAbilities).toBe(false);
  });

  it("moves a blob from before the categories existed to the new default", async () => {
    const { s } = await load({ __version: 11, gameplay: { smartAutoPass: true } });
    expect(s.gameplay.respondAbilities).toBe(true);
    expect(s.gameplay.respondUntargetedAbilities).toBe(false);
  });

  it("keeps the old meaning for a player who changed the Stop for list", async () => {
    // Counters off, abilities left on: they looked at the list and kept
    // "any ability", so untargeted abilities still stop them.
    let { s } = await load({
      __version: 21,
      gameplay: {
        respondCounterspells: false,
        respondInstants: true,
        respondAbilities: true,
        respondSpecialActions: true,
      },
    });
    expect(s.gameplay.respondCounterspells).toBe(false);
    expect(s.gameplay.respondAbilities).toBe(true);
    expect(s.gameplay.respondUntargetedAbilities).toBe(true);

    // Abilities off: both halves stay off.
    ({ s } = await load({
      __version: 21,
      gameplay: {
        respondCounterspells: true,
        respondInstants: true,
        respondAbilities: false,
        respondSpecialActions: true,
      },
    }));
    expect(s.gameplay.respondAbilities).toBe(false);
    expect(s.gameplay.respondUntargetedAbilities).toBe(false);
  });

  it("from v22 on, a stored choice stands and survives a save", async () => {
    const { mod } = await load({
      __version: 22,
      gameplay: { respondUntargetedAbilities: true },
    });
    expect(get(mod.settings).gameplay.respondUntargetedAbilities).toBe(true);
    mod.updateSettings("gameplay", "confirmExit", false);
    const again = await freshModule();
    expect(get(again.settings).gameplay.respondUntargetedAbilities).toBe(true);
  });

  it("a v22 blob with a malformed value falls back to the default", async () => {
    const { s } = await load({
      __version: 22,
      gameplay: { respondUntargetedAbilities: "yes" },
    });
    expect(s.gameplay.respondUntargetedAbilities).toBe(false);
  });
});

// ---- ADR 0125 §4: the help group (first-use hints) ----

describe("the help group (v20 → v21)", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.doUnmock("./hints/retired");
  });

  it("adds the group to a v20 blob with nothing seen and tips on", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 20, display: { theme: "light" } }),
    );
    const { settings, SETTINGS_VERSION } = await freshModule();
    const s = get(settings);
    expect(SETTINGS_VERSION).toBeGreaterThanOrEqual(21);
    expect(s.__version).toBe(SETTINGS_VERSION);
    expect(s.help).toEqual({ seen: {}, tipsOff: false });
    expect(s.display.theme).toBe("light");
  });

  it("keeps a v21 blob's seen hints and tipsOff, unknown ids included", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({
        __version: 21,
        help: { seen: { "table.stack": 2, "lobby.from-a-newer-client": 1 }, tipsOff: true },
      }),
    );
    const { settings } = await freshModule();
    expect(get(settings).help).toEqual({
      seen: { "table.stack": 2, "lobby.from-a-newer-client": 1 },
      tipsOff: true,
    });
  });

  it("checks the stored map rather than trusting it", async () => {
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({
        __version: 21,
        help: { seen: { "table.stack": "2", "lobby.create": 0, "decks.check": 1 }, tipsOff: "yes" },
      }),
    );
    const { settings } = await freshModule();
    expect(get(settings).help).toEqual({ seen: { "decks.check": 1 }, tipsOff: false });
    for (const seen of [null, [1], "x"]) {
      localStorage.setItem(
        "cmdctrl.settings.v1",
        JSON.stringify({ __version: 21, help: { seen } }),
      );
      const again = await freshModule();
      expect(get(again.settings).help.seen, JSON.stringify(seen)).toEqual({});
    }
  });

  it("drops retired ids", async () => {
    vi.doMock("./hints/retired", () => ({ RETIRED_HINT_IDS: ["table.gone"] }));
    localStorage.setItem(
      "cmdctrl.settings.v1",
      JSON.stringify({ __version: 21, help: { seen: { "table.gone": 1, "table.stack": 1 } } }),
    );
    const { settings, applySyncedCopy, defaultSettings } = await freshModule();
    expect(get(settings).help.seen).toEqual({ "table.stack": 1 });
    // An account copy goes through the same check.
    const fromAccount = applySyncedCopy(
      defaultSettings(),
      { help: { seen: { "table.gone": 1, "decks.check": 1 } } },
      21,
    );
    expect(fromAccount.help.seen).toEqual({ "decks.check": 1 });
  });

  it("syncs both fields, and an account copy written before v21 brings none", async () => {
    const { SYNCED_FIELDS, syncedSubset, defaultSettings, applySyncedCopy } = await freshModule();
    expect(SYNCED_FIELDS.help).toEqual({ seen: "synced", tipsOff: "synced" });
    const s = defaultSettings();
    s.help = { seen: { "table.stack": 1 }, tipsOff: true };
    expect(syncedSubset(s).help).toEqual({ seen: { "table.stack": 1 }, tipsOff: true });
    const old = applySyncedCopy(s, { display: { theme: "light" } }, 20);
    expect(old.help).toEqual({ seen: {}, tipsOff: false });
  });

  // §4's size estimate: about 30 bytes an entry, so the 18 planned hints
  // come to well under a kilobyte and the 32 KiB cap would take about a
  // thousand. The body's depth is 3 (help.seen.<id>), inside the limit of 4.
  it("stays small: 18 hints are under 1 KiB, and the depth is 3", async () => {
    const { syncedSubset, defaultSettings } = await freshModule();
    const ids = [
      "site.help",
      "lobby.practice",
      "lobby.create",
      "decks.check",
      "decks.library",
      "settings.display",
      "catalog.search",
      "roadmap.search",
      "admin.views",
      "table.dock",
      "table.opening-roll",
      "table.stack",
      "table.right-click",
      "table.commander",
      "table.attention",
      "table.more",
      "table.expand",
      "table.shortcuts",
    ];
    const s = defaultSettings();
    s.help.seen = Object.fromEntries(ids.map((id) => [id, 1]));
    const help = JSON.stringify(syncedSubset(s).help);
    expect(help.length).toBeLessThan(1024);
    expect(help.length / ids.length).toBeLessThan(32);
    const depth = (v: unknown): number =>
      v !== null && typeof v === "object"
        ? 1 + Math.max(0, ...Object.values(v as object).map(depth))
        : 0;
    expect(depth(syncedSubset(s))).toBe(3);
    // The whole synced body, with a thousand hints seen, still fits 32 KiB.
    s.help.seen = Object.fromEntries(
      Array.from({ length: 1000 }, (_, i) => [`table.hint-${i}`, 1]),
    );
    expect(JSON.stringify(syncedSubset(s)).length).toBeLessThan(32 * 1024);
  });
});

// ---- #2871: stop at a ticked step only when I can act ----

describe("stepStopsOnlyWhenCanAct (v22 → v23)", () => {
  beforeEach(() => localStorage.clear());

  const load = async (blob: unknown) => {
    localStorage.setItem("cmdctrl.settings.v1", JSON.stringify(blob));
    const mod = await freshModule();
    return { mod, s: get(mod.settings) };
  };

  it("is on for a new player", async () => {
    const { defaultSettings, SETTINGS_VERSION } = await freshModule();
    expect(SETTINGS_VERSION).toBeGreaterThanOrEqual(23);
    expect(defaultSettings().gameplay.stepStopsOnlyWhenCanAct).toBe(true);
  });

  it("copies smart auto-pass from a v22 blob", async () => {
    let { mod, s } = await load({ __version: 22, gameplay: { smartAutoPass: true } });
    expect(s.__version).toBe(mod.SETTINGS_VERSION);
    expect(s.gameplay.stepStopsOnlyWhenCanAct).toBe(true);

    // Smart auto-pass off meant every ticked step stops; it still does.
    ({ mod, s } = await load({ __version: 22, gameplay: { smartAutoPass: false } }));
    expect(s.gameplay.passMode).toBe("careful");
    expect(s.gameplay.stepStopsOnlyWhenCanAct).toBe(false);
  });

  it("takes the default for an older blob that never stored smart auto-pass", async () => {
    const { s } = await load({ __version: 4, gameplay: {} });
    expect(s.gameplay.stepStopsOnlyWhenCanAct).toBe(true);
  });

  it("from v23 on, a stored choice stands and survives a save", async () => {
    const { mod } = await load({
      __version: 23,
      gameplay: { smartAutoPass: true, stepStopsOnlyWhenCanAct: false },
    });
    expect(get(mod.settings).gameplay.stepStopsOnlyWhenCanAct).toBe(false);
    mod.updateSettings("gameplay", "confirmExit", false);
    const again = await freshModule();
    expect(get(again.settings).gameplay.stepStopsOnlyWhenCanAct).toBe(false);
    expect(get(again.settings).gameplay.passMode).toBe("smart");
    expect("smartAutoPass" in get(again.settings).gameplay).toBe(false);
  });

  it("a v23 blob with a malformed value falls back to the default", async () => {
    const { s } = await load({ __version: 23, gameplay: { stepStopsOnlyWhenCanAct: "yes" } });
    expect(s.gameplay.stepStopsOnlyWhenCanAct).toBe(true);
  });
});

// ---- ADR 0143 §2.1 and §5: one Auto-pass choice ----

describe("passMode (v23 → v24)", () => {
  beforeEach(() => localStorage.clear());

  const load = async (blob: unknown) => {
    localStorage.setItem("cmdctrl.settings.v1", JSON.stringify(blob));
    const mod = await freshModule();
    return { mod, s: get(mod.settings) };
  };
  const modeOf = async (blob: unknown) => (await load(blob)).s.gameplay.passMode;
  const RETIRED = ["autoPassPriority", "smartAutoPass", "alwaysStopOpponentStack"];

  it("is Smart for a new player, and the retired keys are gone", async () => {
    const { defaultSettings, SETTINGS_VERSION } = await freshModule();
    expect(SETTINGS_VERSION).toBeGreaterThanOrEqual(24);
    const gp = defaultSettings().gameplay;
    expect(gp.passMode).toBe("smart");
    for (const k of RETIRED) expect(k in gp, k).toBe(false);
  });

  // Every combination of the three retired keys, as a v23 client wrote
  // them (always all three, materialised).
  it("maps every combination of the three retired keys", async () => {
    for (const autoPassPriority of [true, false]) {
      for (const smartAutoPass of [true, false]) {
        for (const alwaysStopOpponentStack of [true, false]) {
          const { mod, s } = await load({
            __version: 23,
            gameplay: { autoPassPriority, smartAutoPass, alwaysStopOpponentStack },
          });
          const want = !autoPassPriority
            ? "manual"
            : !smartAutoPass || alwaysStopOpponentStack
              ? "careful"
              : "smart";
          const label = JSON.stringify({
            autoPassPriority,
            smartAutoPass,
            alwaysStopOpponentStack,
          });
          expect(s.__version).toBe(mod.SETTINGS_VERSION);
          expect(s.gameplay.passMode, label).toBe(want);
          for (const k of RETIRED) expect(k in s.gameplay, `${label} ${k}`).toBe(false);
        }
      }
    }
  });

  it("maps a blob that stored only some of the keys", async () => {
    expect(await modeOf({ __version: 12, gameplay: { autoPassPriority: false } })).toBe("manual");
    expect(await modeOf({ __version: 12, gameplay: { smartAutoPass: false } })).toBe("careful");
    expect(await modeOf({ __version: 12, gameplay: { alwaysStopOpponentStack: true } })).toBe(
      "careful",
    );
    expect(await modeOf({ __version: 12, gameplay: { smartAutoPass: true } })).toBe("smart");
    expect(await modeOf({ __version: 4, gameplay: {} })).toBe("smart");
  });

  it("keeps the v1 and v3 rescues of autoPassPriority", async () => {
    // v1 with an empty grid: the v1 → v2 block turned auto-pass on.
    expect(
      await modeOf({ __version: 1, gameplay: { autoPassPriority: false, stepStops: {} } }),
    ).toBe("smart");
    // v2 on the seeded grid: the v2 → v3 hotfix turned it on too.
    const { defaultStepStops } = await freshModule();
    expect(
      await modeOf({
        __version: 2,
        gameplay: { autoPassPriority: false, stepStops: defaultStepStops() },
      }),
    ).toBe("smart");
  });

  it("from v24 on, a stored mode stands and survives a save", async () => {
    for (const mode of ["smart", "careful", "manual"] as const) {
      const { mod } = await load({ __version: 24, gameplay: { passMode: mode } });
      expect(get(mod.settings).gameplay.passMode).toBe(mode);
      mod.updateSettings("gameplay", "confirmExit", false);
      const again = await freshModule();
      expect(get(again.settings).gameplay.passMode).toBe(mode);
      const stored = JSON.parse(localStorage.getItem("cmdctrl.settings.v1") ?? "{}");
      expect(stored.gameplay.passMode).toBe(mode);
      for (const k of RETIRED) expect(k in stored.gameplay, k).toBe(false);
    }
  });

  it("a bad value is Smart", async () => {
    for (const bad of ["fast", 3, null, true]) {
      expect(await modeOf({ __version: 24, gameplay: { passMode: bad } }), String(bad)).toBe(
        "smart",
      );
    }
  });

  it("a hand-seeded blob with no version keeps its passMode (the e2e seed)", async () => {
    expect(await modeOf({ gameplay: { passMode: "careful" } })).toBe("careful");
  });

  it("an older client's write wins over a passMode it carried along", async () => {
    // A v23 tab keeps the unknown passMode and writes its own keys.
    expect(
      await modeOf({
        __version: 23,
        gameplay: { passMode: "manual", autoPassPriority: true, smartAutoPass: false },
      }),
    ).toBe("careful");
  });

  it("an account copy from a v23 client moves the same way", async () => {
    const { applySyncedCopy, defaultSettings } = await freshModule();
    const base = defaultSettings();
    const copy = {
      gameplay: { autoPassPriority: true, smartAutoPass: true, alwaysStopOpponentStack: true },
    };
    const next = applySyncedCopy(base, copy, 23);
    expect(next.gameplay.passMode).toBe("careful");
    for (const k of RETIRED) expect(k in next.gameplay, k).toBe(false);
    const manual = applySyncedCopy(base, { gameplay: { autoPassPriority: false } }, 23);
    expect(manual.gameplay.passMode).toBe("manual");
  });
});
