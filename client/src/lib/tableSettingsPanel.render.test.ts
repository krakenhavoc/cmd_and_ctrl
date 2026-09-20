// @vitest-environment jsdom
//
// The table-settings panel's two rendering rules (ADR 0075 §2.5),
// neither of which can be asked without a DOM:
//
// 1. A non-manager sees the WHOLE panel with every control disabled.
//    The settings are public on purpose — hiding them would be the
//    easy mistake, and it would leave the table unable to find out
//    whether spawning is on.
// 2. Starting life is disabled once the game is active, because the
//    server answers 422 there. The control is the only warning the
//    player gets before the click.
//
// What the panel EMITS is pinned here too: a patch carries only the
// field that moved, since a whole-object write would reset the four
// knobs the host did not touch.

import { describe, it, expect, afterEach, vi } from "vitest";

import TableSettingsPanel from "./components/TableSettingsPanel.svelte";
import type { TableSettingsView } from "./protocol";
import { DEFAULT_TABLE_SETTINGS, UNDO_UNLIMITED } from "./tableSettings";
import { render, cleanup, click, flushSync } from "./test/render.svelte";

afterEach(cleanup);

function mount(over: {
  settings?: Partial<TableSettingsView>;
  canManage?: boolean;
  gameState?: string;
  onpatch?: (p: Partial<TableSettingsView>) => void;
}) {
  return render(
    TableSettingsPanel as never,
    {
      settings: { ...DEFAULT_TABLE_SETTINGS, ...(over.settings ?? {}) },
      canManage: over.canManage ?? true,
      gameState: over.gameState ?? "lobby",
      onpatch: over.onpatch ?? (() => {}),
    } as never,
  );
}

function field(container: HTMLElement, name: string): HTMLElement {
  const el = container.querySelector<HTMLElement>(`[data-field="${name}"]`);
  if (!el) throw new Error(`no settings field "${name}" in the panel`);
  return el;
}

function control(container: HTMLElement, name: string): HTMLInputElement {
  const el = field(container, name).querySelector<HTMLInputElement>("input");
  if (!el) throw new Error(`no input in settings field "${name}"`);
  return el;
}

describe("the table settings panel", () => {
  it("renders every v1 setting", () => {
    const { container } = mount({});
    for (const name of [
      "undo_limit",
      "undo_scope",
      "starting_life",
      "commander_damage",
      "bot_pace",
      "allow_spawn",
    ]) {
      expect(field(container, name)).toBeTruthy();
    }
  });

  it("shows a non-manager the same panel, with everything disabled", () => {
    const { container } = mount({ canManage: false });
    const controls = container.querySelectorAll<HTMLInputElement | HTMLButtonElement>(
      ".tsp-field input, .tsp-field button",
    );
    expect(controls.length).toBeGreaterThan(0);
    for (const c of controls) expect(c.disabled).toBe(true);
    // Still readable — the values are the point of showing it.
    expect(control(container, "starting_life").value).toBe("40");
  });

  it("sends nothing at all when a non-manager's control is reached anyway", () => {
    const onpatch = vi.fn();
    const { container } = mount({ canManage: false, onpatch });
    click(field(container, "bot_pace").querySelectorAll("button")[0]);
    expect(onpatch).not.toHaveBeenCalled();
  });

  it("disables starting life once the game is active, and says why", () => {
    const { container } = mount({ gameState: "active" });
    expect(control(container, "starting_life").disabled).toBe(true);
    expect(field(container, "starting_life").textContent).toMatch(/already been dealt/i);
    // Every other control is still live mid-game — settings change at
    // any time (§2.3); starting life is the one exception.
    expect(control(container, "commander_damage").disabled).toBe(false);
    expect(control(container, "allow_spawn").disabled).toBe(false);
  });

  it("leaves starting life editable in the lobby", () => {
    const { container } = mount({ gameState: "lobby" });
    expect(control(container, "starting_life").disabled).toBe(false);
  });

  it("renders an unlimited budget as ∞ rather than as -1 take-backs", () => {
    const { container } = mount({ settings: { undo_limit: UNDO_UNLIMITED } });
    expect(field(container, "undo_limit").textContent).toMatch(/Unlimited/);
    const chip = field(container, "undo_limit").querySelector("button");
    expect(chip?.textContent?.trim()).toBe("∞");
    expect(chip?.className).toContain("on");
  });

  it("reads 0 as undo being off, not as unlimited", () => {
    const { container } = mount({ settings: { undo_limit: 0 } });
    expect(field(container, "undo_limit").textContent).toMatch(/Undo is off/i);
    expect(field(container, "undo_limit").querySelector("button")?.className).not.toContain("on");
  });

  it("the ∞ chip patches the undo limit to -1", () => {
    const onpatch = vi.fn();
    const { container } = mount({ settings: { undo_limit: 1 }, onpatch });
    click(field(container, "undo_limit").querySelector("button")!);
    expect(onpatch).toHaveBeenCalledWith({ undo_limit: UNDO_UNLIMITED });
  });

  it("patches only the field that moved", () => {
    const onpatch = vi.fn();
    const { container } = mount({ settings: { bot_pace: "normal" }, onpatch });
    const slow = [...field(container, "bot_pace").querySelectorAll("button")].find(
      (b) => b.textContent?.trim() === "Slow",
    );
    click(slow!);
    expect(onpatch).toHaveBeenCalledWith({ bot_pace: "slow" });
  });

  it("sends nothing when the control is set to what it already is", () => {
    const onpatch = vi.fn();
    const { container } = mount({ settings: { bot_pace: "fast" }, onpatch });
    const fast = [...field(container, "bot_pace").querySelectorAll("button")].find(
      (b) => b.textContent?.trim() === "Fast",
    );
    click(fast!);
    expect(onpatch).not.toHaveBeenCalled();
  });

  it("toggles the spawn switch with a boolean patch", () => {
    const onpatch = vi.fn();
    const { container } = mount({ settings: { allow_spawn: false }, onpatch });
    const box = control(container, "allow_spawn");
    box.checked = true;
    box.dispatchEvent(new Event("change", { bubbles: true }));
    flushSync();
    expect(onpatch).toHaveBeenCalledWith({ allow_spawn: true });
  });

  it("warns that lowering commander damage can end a game at the next check", () => {
    const { container } = mount({ gameState: "active", settings: { commander_damage: 21 } });
    const box = control(container, "commander_damage");
    box.value = "10";
    box.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    expect(field(container, "commander_damage").textContent).toMatch(/next check/i);
  });

  it("catches an out-of-range value under the control instead of sending it", () => {
    const onpatch = vi.fn();
    const { container } = mount({ gameState: "lobby", onpatch });
    const box = control(container, "starting_life");
    box.value = "5000";
    box.dispatchEvent(new Event("input", { bubbles: true }));
    box.dispatchEvent(new Event("change", { bubbles: true }));
    flushSync();
    expect(onpatch).not.toHaveBeenCalled();
    expect(container.querySelector(".tsp-error")?.textContent).toMatch(/1–999/);
  });

  it("renders the server's own refusal verbatim", () => {
    const { container } = render(
      TableSettingsPanel as never,
      {
        settings: DEFAULT_TABLE_SETTINGS,
        canManage: true,
        gameState: "active",
        onpatch: () => {},
        error: "only the table host or the admin may change settings",
      } as never,
    );
    expect(container.querySelector(".tsp-error")?.textContent).toContain("only the table host");
  });
});
