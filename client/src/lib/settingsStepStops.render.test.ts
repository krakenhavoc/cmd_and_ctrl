// @vitest-environment jsdom
//
// settingsStepStops.render.test.ts — ADR 0143 §2.3. The Gameplay tab's
// stops grid has two columns, My turn and Opponents' turns, each bound
// to its own setting.

import { get } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it } from "vitest";

import Settings from "./components/Settings.svelte";
import { defaultSettings, openSettings, settings } from "./settings";
import { _resetForTests as resetModals } from "./modalLayers";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

beforeEach(() => {
  settings.set(defaultSettings());
  resetModals();
});
afterEach(() => cleanup());

function openGameplay(): HTMLTableElement {
  render(Settings as never, {} as never);
  openSettings("gameplay");
  flushSync();
  const table = document.querySelector<HTMLTableElement>("table.step-stops-table");
  if (!table) throw new Error("no stops table");
  return table;
}

const box = (table: HTMLTableElement, step: string, column: string) => {
  const b = table.querySelector<HTMLInputElement>(`input[aria-label="${step}, ${column}"]`);
  if (!b) throw new Error(`no box for ${step}, ${column}`);
  return b;
};

describe("Settings: stops by whose turn it is", () => {
  it("has a My turn and an Opponents' turns column", () => {
    const table = openGameplay();
    const heads = [...table.querySelectorAll("thead th")].map((th) => th.textContent?.trim());
    expect(heads.slice(1)).toEqual(["My turn", "Opponents' turns"]);
    // Ten stoppable steps, two boxes each.
    expect(table.querySelectorAll("tbody tr")).toHaveLength(10);
    expect(table.querySelectorAll("tbody input[type=checkbox]")).toHaveLength(20);
  });

  it("shows the defaults: your main phases and combat declarations, nothing for opponents", () => {
    const table = openGameplay();
    const mine = [...table.querySelectorAll<HTMLInputElement>("tbody tr td:nth-of-type(1) input")];
    const theirs = [
      ...table.querySelectorAll<HTMLInputElement>("tbody tr td:nth-of-type(2) input"),
    ];
    expect(mine.filter((b) => b.checked)).toHaveLength(4);
    expect(theirs.filter((b) => b.checked)).toHaveLength(0);
  });

  it("each column writes its own setting", () => {
    const table = openGameplay();
    const upkeepTheirs = box(table, "Upkeep", "Opponents' turns");
    click(upkeepTheirs);
    flushSync();
    expect(get(settings).gameplay.stepStopsOpponents.upkeep).toBe(true);
    expect(get(settings).gameplay.stepStops.upkeep).toBe(false);

    const upkeepMine = box(table, "Upkeep", "My turn");
    click(upkeepMine);
    flushSync();
    expect(get(settings).gameplay.stepStops.upkeep).toBe(true);
  });
});
