// settingsSyncFields.test.ts — ADR 0110 §4 item 3 and owner answer 5:
// every settings field is classified as per-person (synced to the
// account) or per-device (never leaves the browser), and the split is
// the one the owner chose.
//
// The first test is the guard. Adding a field to Settings without
// deciding where it lives fails here (and in `npm run check`, since
// SYNCED_FIELDS' type is exhaustive). Removing one without removing its
// line fails here too.

import { describe, expect, it } from "vitest";
import {
  SYNCED_FIELDS,
  applySyncedCopy,
  canonicalJSON,
  defaultSettings,
  syncedPaths,
  syncedSubset,
  SETTINGS_VERSION,
  type Settings,
} from "./settings";

function fieldsOf(s: Settings): string[] {
  const out: string[] = [];
  for (const [group, value] of Object.entries(s)) {
    if (group === "__version") continue;
    for (const key of Object.keys(value as object)) out.push(`${group}.${key}`);
  }
  return out.sort();
}

function classified(): string[] {
  const out: string[] = [];
  for (const [group, scopes] of Object.entries(SYNCED_FIELDS)) {
    for (const key of Object.keys(scopes)) out.push(`${group}.${key}`);
  }
  return out.sort();
}

describe("SYNCED_FIELDS", () => {
  it("classifies every settings field, and nothing that is not one", () => {
    const fields = fieldsOf(defaultSettings());
    const unclassified = fields.filter((f) => !classified().includes(f));
    const stale = classified().filter((f) => !fields.includes(f));
    expect(
      unclassified,
      "add these to SYNCED_FIELDS in settings.ts as 'synced' or 'device'",
    ).toEqual([]);
    expect(stale, "these are in SYNCED_FIELDS but no longer in Settings").toEqual([]);
    for (const scopes of Object.values(SYNCED_FIELDS)) {
      for (const scope of Object.values(scopes)) expect(["synced", "device"]).toContain(scope);
    }
  });

  it("keeps owner answer 5's per-device list exactly", () => {
    const device = classified().filter((f) => {
      const [group, key] = f.split(".");
      return (SYNCED_FIELDS as Record<string, Record<string, string>>)[group][key] === "device";
    });
    expect(device).toEqual(
      [
        // Volumes and mute: speakers versus headphones.
        "audio.muted",
        "audio.masterVolume",
        "audio.effectsVolume",
        "audio.musicVolume",
        // Screen-shaped (#956).
        "display.cardSize",
        "display.handLayout",
        "display.tableLayout",
        "display.opponentDetail",
        "display.expandActivePlayer",
        "display.expandStyle",
        "accessibility.textScale",
        // Follow the OS reduced-motion signal at runtime.
        "animations.enabled",
        "accessibility.reduceMotion",
      ].sort(),
    );
  });

  it("syncs the art-tile choices (#1954, #2209) with the other display tastes", () => {
    expect(SYNCED_FIELDS.display.battlefieldArt).toBe("synced");
    expect(SYNCED_FIELDS.display.handArt).toBe("synced");
    expect(SYNCED_FIELDS.display.theme).toBe("synced");
    expect(SYNCED_FIELDS.display.stackStyle).toBe("synced");
  });

  it("syncs the two forced practice-table fields that are per person, and not the two that are per device", () => {
    expect(SYNCED_FIELDS.gameplay.strictMana).toBe("synced");
    expect(SYNCED_FIELDS.gameplay.autoPassPriority).toBe("synced");
    expect(SYNCED_FIELDS.display.tableLayout).toBe("device");
    expect(SYNCED_FIELDS.display.cardSize).toBe("device");
  });
});

describe("syncedSubset and applySyncedCopy", () => {
  it("the subset carries every synced field and no device field", () => {
    const s = defaultSettings();
    const subset = syncedSubset(s);
    const sent = Object.entries(subset).flatMap(([g, v]) => Object.keys(v).map((k) => `${g}.${k}`));
    expect(sent.sort()).toEqual(
      syncedPaths()
        .map(([g, k]) => `${g}.${k}`)
        .sort(),
    );
    expect(subset).not.toHaveProperty("audio");
  });

  it("the subset fits the server's caps: depth 4, well under 32 KiB", () => {
    const subset = syncedSubset(defaultSettings());
    const depth = (v: unknown): number =>
      v !== null && typeof v === "object"
        ? 1 + Math.max(0, ...Object.values(v as object).map(depth))
        : 0;
    expect(depth(subset)).toBeLessThanOrEqual(4);
    expect(JSON.stringify(subset).length).toBeLessThan(32 * 1024);
  });

  it("applying a copy changes only synced fields, through migrate", () => {
    const base = defaultSettings();
    base.display.cardSize = "large";
    base.audio.masterVolume = 7;
    const next = applySyncedCopy(
      base,
      {
        display: { theme: "light", cardSize: "small" },
        audio: { masterVolume: 99 },
        bogus: { x: 1 },
      },
      SETTINGS_VERSION,
    );
    expect(next.display.theme).toBe("light");
    // Per-device fields stay the device's even if a copy names them.
    expect(next.display.cardSize).toBe("large");
    expect(next.audio.masterVolume).toBe(7);
    expect(next).not.toHaveProperty("bogus");
    // base is not mutated.
    expect(base.display.theme).toBe("dark");
  });

  it("a stack style migrate does not know falls back, as a stored one does", () => {
    const next = applySyncedCopy(
      defaultSettings(),
      { display: { stackStyle: "removed-style" } },
      SETTINGS_VERSION,
    );
    expect(next.display.stackStyle).toBe("compact");
  });

  it("canonicalJSON ignores key order", () => {
    expect(canonicalJSON({ b: 1, a: { d: 2, c: [1, { f: 1, e: 2 }] } })).toBe(
      canonicalJSON({ a: { c: [1, { e: 2, f: 1 }], d: 2 }, b: 1 }),
    );
  });
});
