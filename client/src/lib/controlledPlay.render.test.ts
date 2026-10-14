// @vitest-environment jsdom
//
// ADR 0166 §4 (#2164): Word of Command's forced play. The controller is
// shown the chosen card and the server's ways of playing it (the legal
// moves naming the prompt), and each one is sent exactly as the server
// wrote it, as the controlled player. With one way it is the bar's
// primary; with none, "It can't be played" is.

import { describe, it, expect, afterEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import type { ActionType, CardView, GameView, LegalMoveView } from "./protocol";
import { barPrimary, nameOf } from "./test/dockView";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const bolt = {
  instance_id: "bolt",
  name: "Lightning Bolt",
  owner: "them",
  controller: "them",
  type_line: "Instant",
  known_by_you: true,
} as CardView;

const castAt = (target: string, label: string): LegalMoveView =>
  ({
    type: "cast_spell",
    player: "them",
    kind: "cast",
    label: `Word of Command: ${label}`,
    source: "bolt",
    params: {
      instance_id: "bolt",
      targets: [{ kind: "player", id: target }],
      controlled_play: "play-1",
    },
    controlled_play: "play-1",
  }) as LegalMoveView;

const cantPlay = {
  type: "resolve_choice",
  player: "me",
  kind: "choice",
  label: "Word of Command: it can't be played",
  params: { choice_id: "play-1" },
  controlled_play: "play-1",
} as LegalMoveView;

const land = (id: string, name: string, extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: id,
    name,
    owner: "them",
    controller: "them",
    type_line: "Basic Land — Mountain",
    ...extra,
  }) as CardView;

const theirLands = [
  land("m1", "Mountain"),
  land("m2", "Mountain"),
  land("m3", "Tapped Mountain", { tapped: true }),
];

const snapWith = (moves: LegalMoveView[]): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      { id: "me", name: "Me" },
      { id: "them", name: "Bob" },
    ],
    battlefield: { kind: "battlefield", count: 3, cards: theirLands },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: -1, phase: "main1", step: "main" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "play-1",
        kind: "controlled_play",
        chooser: "me",
        from_player: "them",
        count: 1,
        reason: "Word of Command",
        options: [bolt],
        eligible: ["bolt"],
      },
    ],
    legal_moves: moves,
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
  player?: string;
}

function mount(moves: LegalMoveView[]): { container: HTMLElement; sent: Sent[] } {
  const sent: Sent[] = [];
  const view = render(
    ChoiceDockHarness as never,
    {
      snap: snapWith(moves),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown, player?: string) =>
        sent.push({ type, params, player }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

describe("ControlledPlaySheet — Word of Command's forced play (ADR 0166)", () => {
  it("names the card and the player you control", () => {
    mount([castAt("them", "Cast Lightning Bolt targeting Bob")]);
    const dialog = document.querySelector('[role="dialog"][aria-label^="Word of Command"]');
    expect(dialog).not.toBeNull();
    expect(nameOf(dialog!)).toBe("Word of Command: play Lightning Bolt as Bob");
  });

  it("makes a lone way the primary, sent as the controlled player", () => {
    const { sent } = mount([castAt("them", "Cast Lightning Bolt targeting Bob")]);
    const primary = barPrimary();
    expect(primary && nameOf(primary)).toBe("Cast Lightning Bolt targeting Bob");
    click(primary!);
    expect(sent).toEqual([
      {
        type: "cast_spell",
        player: "them",
        params: {
          instance_id: "bolt",
          targets: [{ kind: "player", id: "them" }],
          controlled_play: "play-1",
        },
      },
    ]);
  });

  it("lists several ways, and sends the one clicked", () => {
    const { sent } = mount([
      castAt("them", "Cast Lightning Bolt targeting Bob"),
      castAt("me", "Cast Lightning Bolt targeting Me"),
    ]);
    const rows = [...document.querySelectorAll<HTMLButtonElement>("ul.plays button")];
    expect(rows.map((b) => nameOf(b))).toEqual([
      "Cast Lightning Bolt targeting Bob",
      "Cast Lightning Bolt targeting Me",
    ]);
    click(rows[0]);
    expect(sent).toHaveLength(1);
    expect((sent[0].params as { targets: { id: string }[] }).targets[0].id).toBe("them");
  });

  it("lets you pick their lands, locking the rest out of the auto-tapper", () => {
    const { sent } = mount([castAt("them", "Cast Lightning Bolt targeting Bob")]);
    const chips = [...document.querySelectorAll<HTMLButtonElement>("fieldset.lands button")];
    // Untapped lands only.
    expect(chips.map((b) => nameOf(b))).toEqual(["Mountain", "Mountain"]);
    click(chips[1]);
    expect(chips[1].getAttribute("aria-pressed")).toBe("true");
    click(barPrimary()!);
    expect((sent[0].params as { locked_sources: string[] }).locked_sources).toEqual(["m1"]);
  });

  it("sends no lock when you let the game choose", () => {
    const { sent } = mount([castAt("them", "Cast Lightning Bolt targeting Bob")]);
    click(barPrimary()!);
    expect(sent[0].params).not.toHaveProperty("locked_sources");
  });

  it('offers "It can\'t be played" when there is no way', () => {
    const { sent } = mount([cantPlay]);
    const primary = barPrimary();
    expect(primary && nameOf(primary)).toBe("It can't be played");
    click(primary!);
    expect(sent).toEqual([
      { type: "resolve_choice", player: "me", params: { choice_id: "play-1" } },
    ]);
  });
});
