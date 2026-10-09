import { describe, it, expect, beforeEach, vi } from "vitest";
import { get } from "svelte/store";

// settingsTriggerOrder.test.ts — #1968: #1530's "Always ask me to order
// my triggers" checkbox (gameplay.alwaysAskTriggerOrder) becomes the
// three-way gameplay.triggerOrder, for a stored blob and for an account
// copy from an older client (ADR 0110 §4).

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

async function freshModule() {
  vi.resetModules();
  return await import("./settings");
}

function store(gameplay: Record<string, unknown>) {
  localStorage.setItem("cmdctrl.settings.v1", JSON.stringify({ __version: 21, gameplay }));
}

describe("#1968: the trigger-order setting", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("defaults to asking only when the order matters", async () => {
    const { settings } = await freshModule();
    expect(get(settings).gameplay.triggerOrder).toBe("when_it_matters");
  });

  it("maps #1530's ticked checkbox to always, and drops the old key", async () => {
    store({ alwaysAskTriggerOrder: true });
    const { settings } = await freshModule();
    const s = get(settings);
    expect(s.gameplay.triggerOrder).toBe("always");
    expect("alwaysAskTriggerOrder" in s.gameplay).toBe(false);
  });

  it("maps an unticked checkbox to the default", async () => {
    store({ alwaysAskTriggerOrder: false });
    const { settings } = await freshModule();
    expect(get(settings).gameplay.triggerOrder).toBe("when_it_matters");
  });

  it("keeps a stored mode over the old key, and refuses an unknown one", async () => {
    store({ triggerOrder: "never", alwaysAskTriggerOrder: true });
    let m = await freshModule();
    expect(get(m.settings).gameplay.triggerOrder).toBe("never");
    store({ triggerOrder: "sometimes" });
    m = await freshModule();
    expect(get(m.settings).gameplay.triggerOrder).toBe("when_it_matters");
  });

  it("an account copy from an older client migrates the same way (ADR 0110 §4)", async () => {
    const m = await freshModule();
    const ticked = m.applySyncedCopy(
      m.defaultSettings(),
      { gameplay: { alwaysAskTriggerOrder: true } },
      21,
    );
    expect(ticked.gameplay.triggerOrder).toBe("always");
    const never = m.applySyncedCopy(
      m.defaultSettings(),
      { gameplay: { triggerOrder: "never" } },
      21,
    );
    expect(never.gameplay.triggerOrder).toBe("never");
    expect(m.syncedSubset(never).gameplay).toHaveProperty("triggerOrder", "never");
    expect(m.syncedSubset(never).gameplay).not.toHaveProperty("alwaysAskTriggerOrder");
  });
});
