// @vitest-environment jsdom
//
// manaChosenPlayer.render.test.ts — #6101. "{T}: Choose a player. That
// player adds one mana of any color they choose." (Spectral
// Searchlight) opens a sheet with one button per player still in the
// game, you first; the confirm names who gets the mana and sends that
// player's id.

import { describe, it, expect, afterEach, beforeEach } from "vitest";

import ManaChosenPlayerModal from "./components/board/ManaChosenPlayerModal.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { defaultSettings, settings } from "./settings";
import { chosenPlayerOptions } from "./manaAbilityCost";
import { L } from "./labels";
import type { CardView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";
import { barPrimary, nameOf, sheetPanel } from "./test/dockView";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  settings.set(defaultSettings());
  resetDock();
  resetModals();
});
afterEach(cleanup);

const searchlight = {
  instance_id: "light-1",
  name: "Spectral Searchlight",
  owner: "me",
  controller: "me",
} as CardView;

const seats = [
  { id: "bea", name: "Bea" },
  { id: "me", name: "Me" },
  { id: "cy", name: "Cy", eliminated: true },
];

function mount(onConfirm: (id: string) => void) {
  return render(
    DockHarness as never,
    {
      component: ManaChosenPlayerModal,
      props: {
        card: searchlight,
        players: chosenPlayerOptions(seats, "me"),
        onConfirm,
        onCancel: () => {},
      },
    } as never,
  );
}

describe("the chosen-player sheet", () => {
  it("offers you first and the other players still in the game", () => {
    mount(() => {});
    const group = document.querySelector(`[aria-label="${L.manaChosenPlayer}"]`);
    expect(group).not.toBeNull();
    const names = [...sheetPanel()!.querySelectorAll(".player-opt")].map((b) =>
      b.textContent?.trim(),
    );
    expect(names).toEqual(["You", "Bea"]);
    expect(nameOf(barPrimary()!)).toBe("Add mana yourself");
  });

  it("sends the player picked", () => {
    let sent = "";
    mount((id) => (sent = id));
    const buttons = sheetPanel()!.querySelectorAll<HTMLButtonElement>(".player-opt");
    buttons[1].click();
    flushSync();
    expect(nameOf(barPrimary()!)).toBe("Give the mana to Bea");
    barPrimary()!.click();
    flushSync();
    expect(sent).toBe("bea");
  });
});
