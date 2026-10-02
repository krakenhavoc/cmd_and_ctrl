// ADR 0111 §5: the one press the button and the `b` key share.

import { describe, it, expect, beforeEach } from "vitest";
import { get } from "svelte/store";

import { _resetForTests, bluffArmed, pressBluff } from "./bluff";
import { defaultSettings, settings } from "./settings";

beforeEach(() => {
  settings.set(defaultSettings());
  _resetForTests();
});

describe("pressBluff", () => {
  it("arms and turns on the counterspell bluff when nothing is set up", () => {
    expect(pressBluff()).toBe(true);
    expect(get(bluffArmed)).toBe(true);
    expect(get(settings).gameplay.bluffCounterspell).toBe(true);
    expect(get(settings).gameplay.bluffInstant).toBe(false);
  });

  it("disarms on the next press", () => {
    pressBluff();
    expect(pressBluff()).toBe(false);
    expect(get(bluffArmed)).toBe(false);
  });

  it("changes no setting when a kind is already chosen", () => {
    settings.update((s) => ({ ...s, gameplay: { ...s.gameplay, bluffInstant: true } }));
    pressBluff();
    expect(get(bluffArmed)).toBe(true);
    expect(get(settings).gameplay.bluffCounterspell).toBe(false);
  });

  it("does nothing while smart autopass is off", () => {
    settings.update((s) => ({ ...s, gameplay: { ...s.gameplay, smartAutoPass: false } }));
    expect(pressBluff()).toBe(false);
    expect(get(bluffArmed)).toBe(false);
    expect(get(settings).gameplay.bluffCounterspell).toBe(false);
  });
});
