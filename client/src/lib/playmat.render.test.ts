// @vitest-environment jsdom
//
// playmat.render.test.ts — ADR 0128. A seat's playmat is drawn behind
// its part of the board for every viewer: the panel carries the image
// and its wash as custom properties for app.css. Your own always shows;
// other players' follow display.showPlaymats. A path that is not one of
// the server's playmats is never put in a CSS url().

import { describe, it, expect, afterEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import { playmatURL } from "./api";
import { resetSettings, updateSettings } from "./settings";
import type { GameView, PlayerView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  resetSettings();
  localStorage.clear();
});

const zone = (kind: string, owner?: string) => ({ kind, owner, count: 0, cards: [] });

const seat = (id: string, idx: number, playmat?: { path: string; wash: number }): PlayerView =>
  ({
    id,
    name: id,
    seat: idx,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
    playmat,
  }) as unknown as PlayerView;

const MAT = { path: "/playmats/0123456789abcdef0123456789abcdef.webp", wash: 45 };

function mount(who: PlayerView, isSelf: boolean): HTMLElement {
  const view = {
    id: "g1",
    state: "active",
    seats: [seat("me", 0), who],
    battlefield: zone("battlefield"),
    stack: zone("stack"),
    exile: zone("exile"),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
  } as unknown as GameView;
  const r = render(
    PlayerPanel as never,
    {
      seat: who,
      isSelf,
      isActive: false,
      hasPriority: false,
      viewerID: "me",
      isAdmin: false,
      sendAction: () => {},
      isMonarch: false,
      isInitiative: false,
      view,
      controlledCards: [],
      exile: zone("exile"),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
      onTapToggle: () => {},
      onPlayCard: () => {},
      onDrawCard: () => {},
      onActivateAbility: () => {},
      onManaAbilityCost: () => {},
    } as never,
  );
  return r.container.querySelector<HTMLElement>(".seat-panel")!;
}

describe("a seat's playmat (ADR 0128)", () => {
  it("draws another player's playmat with its wash", () => {
    const panel = mount(seat("alice", 1, MAT), false);
    expect(panel.classList.contains("has-playmat")).toBe(true);
    expect(panel.style.getPropertyValue("--playmat")).toContain(MAT.path);
    expect(panel.style.getPropertyValue("--playmat-wash")).toBe("45%");
  });

  it("draws nothing for a seat without one", () => {
    const panel = mount(seat("alice", 1), false);
    expect(panel.classList.contains("has-playmat")).toBe(false);
    expect(panel.style.getPropertyValue("--playmat")).toBe("");
  });

  it("hides other players' playmats when the viewer turns them off, never their own", () => {
    updateSettings("display", "showPlaymats", false);
    expect(mount(seat("alice", 1, MAT), false).classList.contains("has-playmat")).toBe(false);
    cleanup();
    expect(mount(seat("me", 0, MAT), true).classList.contains("has-playmat")).toBe(true);
  });

  it("only ever builds a url for the server's playmat paths", () => {
    expect(playmatURL(MAT.path)).toMatch(/^\/playmats\//);
    expect(playmatURL("https://evil.example/x.png")).toBeNull();
    expect(playmatURL('/playmats/x") ; background: red')).toBeNull();
    expect(playmatURL("/playmats/../me/settings")).toBeNull();
    expect(playmatURL(undefined)).toBeNull();
  });
});
