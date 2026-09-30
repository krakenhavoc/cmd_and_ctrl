// @vitest-environment jsdom
//
// ADR 0103 — a Room's door strip, the fused face-picker choice, and the
// helpers behind them. What only the markup can say: both doors are on
// the tile with their lock state, a locked door the server offers an
// unlock for is a button for the Room's controller and nobody else, a
// greyed row is shown rather than hidden, and clicking asks the Board
// for the unlock through the roomDoors store. The picker lists a
// fused cast and hands it back as fused.

import { get } from "svelte/store";
import { afterEach, describe, expect, it } from "vitest";

import RoomDoorStrip from "./components/board/RoomDoorStrip.svelte";
import FacePickerModal from "./components/board/FacePickerModal.svelte";
import { buildMenuSections } from "./contextMenu.logic";
import { cardAsFused, displayName, faceCastableFrom, faceOptions, needsFacePicker } from "./faces";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { doorRows, unlockParams, unlockPrice, unlockRequest } from "./roomDoors";
import { applyCastChoices } from "./targeting";
import { cleanup, click, render } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  unlockRequest.set(null);
});

function room(extra: Partial<CardView> = {}): CardView {
  return {
    instance_id: "room1",
    name: "Dim Hall",
    owner: "me",
    controller: "me",
    type_line: "Enchantment — Room",
    layout: "split",
    faces: [
      { name: "Dim Hall", type_line: "Enchantment — Room", mana_cost: "{1}{B}" },
      { name: "Deep Cellar", type_line: "Enchantment — Room", mana_cost: "{4}{B}" },
    ],
    doors: { left: true, right: false },
    special_actions: [
      {
        kind: "unlock",
        label: "Unlock Deep Cellar {4}{B}",
        cost: "{4}{B}",
        available: true,
        door: "right",
      },
    ],
    ...extra,
  };
}

describe("roomDoors helpers", () => {
  it("builds one row per door with its unlock row", () => {
    const rows = doorRows(room());
    expect(rows.map((r) => [r.door, r.name, r.unlocked])).toEqual([
      ["left", "Dim Hall", true],
      ["right", "Deep Cellar", false],
    ]);
    expect(rows[0].unlock).toBeUndefined();
    expect(rows[1].unlock?.door).toBe("right");
    expect(unlockPrice(rows[1])).toBe("{4}{B}");
  });

  it("prices a discounted unlock at its charged cost, and a free one as {0}", () => {
    const r = doorRows(
      room({
        special_actions: [
          {
            kind: "unlock",
            label: "",
            cost: "{4}{B}",
            charged_cost: "",
            available: true,
            door: "right",
          },
        ],
      }),
    )[1];
    expect(unlockPrice(r)).toBe("{0}");
  });

  it("is empty for anything that is not a Room on the battlefield", () => {
    expect(doorRows(room({ doors: undefined }))).toEqual([]);
  });

  it("sends the door with the unlock", () => {
    expect(unlockParams("room1", "left")).toEqual({
      card_id: "room1",
      kind: "unlock",
      door: "left",
      strict: true,
      auto_tap: true,
    });
  });

  it("labels a fully locked Room from its halves", () => {
    expect(displayName(room({ name: "" }))).toBe("Dim Hall // Deep Cellar");
    expect(displayName(room())).toBe("Dim Hall");
  });
});

describe("RoomDoorStrip", () => {
  it("shows both doors and an unlock button for the controller", () => {
    const r = render(RoomDoorStrip, { card: room(), canUnlock: true });
    const doors = r.container.querySelectorAll(".door");
    expect(doors).toHaveLength(2);
    expect(doors[0].classList.contains("unlocked")).toBe(true);
    const btn = r.container.querySelector("button.unlock") as HTMLButtonElement;
    expect(btn).not.toBeNull();
    expect(btn.textContent).toContain("Unlock {4}{B}");
    click(btn);
    expect(get(unlockRequest)).toEqual({ cardID: "room1", door: "right" });
  });

  it("offers no button to anyone but the controller", () => {
    const r = render(RoomDoorStrip, { card: room(), canUnlock: false });
    expect(r.container.querySelectorAll(".door")).toHaveLength(2);
    expect(r.container.querySelector("button.unlock")).toBeNull();
  });

  it("greys an unlock the server says is not available now", () => {
    const card = room({
      special_actions: [
        { kind: "unlock", label: "", cost: "{4}{B}", available: false, door: "right" },
      ],
    });
    const r = render(RoomDoorStrip, { card, canUnlock: true });
    const btn = r.container.querySelector("button.unlock") as HTMLButtonElement;
    expect(btn.disabled).toBe(true);
  });
});

function fuseCard(): CardView {
  return {
    instance_id: "fuse1",
    name: "Snap // Crackle",
    owner: "me",
    controller: "me",
    type_line: "Instant",
    layout: "split",
    faces: [
      { name: "Snap", type_line: "Instant", mana_cost: "{R}", oracle_text: "Snap.\nFuse (…)" },
      {
        name: "Crackle",
        type_line: "Instant",
        mana_cost: "{1}{U}",
        oracle_text: "Crackle.\nFuse (…)",
      },
    ],
    fused: { name: "Snap // Crackle", type_line: "Instant", mana_cost: "{R}{1}{U}" },
  };
}

function aftermathCard(): CardView {
  return {
    instance_id: "aft1",
    name: "Bury // Rise Again",
    owner: "me",
    controller: "me",
    type_line: "Instant Sorcery",
    layout: "split",
    faces: [
      { name: "Bury", type_line: "Instant", mana_cost: "{1}{B}", oracle_text: "Bury." },
      {
        name: "Rise Again",
        type_line: "Sorcery",
        mana_cost: "{3}{B}",
        oracle_text:
          "Aftermath (Cast this spell only from your graveyard. Then exile it.)\nRise Again.",
      },
    ],
  };
}

describe("split faces", () => {
  it("asks which half of a split card, and offers the fused cast from hand", () => {
    expect(needsFacePicker(fuseCard())).toBe(true);
    const opts = faceOptions(fuseCard());
    expect(opts.map((o) => [o.face, o.fused, o.view.name])).toEqual([
      [0, false, "Snap"],
      [1, false, "Crackle"],
      [0, true, "Snap // Crackle"],
    ]);
    expect(faceOptions(fuseCard(), "graveyard").some((o) => o.fused)).toBe(false);
    expect(cardAsFused(fuseCard()).mana_cost).toBe("{R}{1}{U}");
  });

  it("keeps an aftermath half out of the hand and in the graveyard", () => {
    const c = aftermathCard();
    expect(faceCastableFrom(c, 1, undefined)).toBe(false);
    expect(faceCastableFrom(c, 1, "graveyard")).toBe(true);
    expect(faceOptions(c).map((o) => o.face)).toEqual([0]);
    expect(faceOptions(c, "graveyard").map((o) => o.face)).toEqual([0, 1]);
  });

  it("sends fuse on the cast", () => {
    const params: Record<string, unknown> = { instance_id: "fuse1" };
    applyCastChoices(params, { fuse: true });
    expect(params.fuse).toBe(true);
    expect(params.face).toBeUndefined();
  });
});

describe("FacePickerModal", () => {
  it("lists the fused cast and confirms it as fused", () => {
    const got: Array<[number, boolean]> = [];
    const r = render(FacePickerModal, {
      card: fuseCard(),
      onConfirm: (face: number, fused: boolean) => got.push([face, fused]),
      onCancel: () => {},
    });
    const opts = r.container.querySelectorAll("button.face-opt");
    expect(opts).toHaveLength(3);
    expect(opts[2].classList.contains("fused")).toBe(true);
    click(opts[2] as HTMLElement);
    click(r.container.querySelector("button.primary") as HTMLElement);
    expect(got).toEqual([[0, true]]);
  });
});

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, n: number): PlayerView {
  return {
    id,
    name: id,
    seat: n,
    life: 40,
    library: zone("library", id, []),
    hand: zone("hand", id, []),
    graveyard: zone("graveyard", id, []),
    command: zone("command", id, []),
    commander_damage: {},
    life_history: [],
  };
}

describe("unlock in the card menu", () => {
  it("names the door in the row and in the payload", () => {
    const view = {
      id: "g1",
      state: "active",
      seats: [seat("me", 0)],
      battlefield: zone("battlefield", undefined, [
        room({
          doors: { left: false, right: false },
          special_actions: [
            {
              kind: "unlock",
              label: "Unlock Dim Hall {1}{B}",
              cost: "{1}{B}",
              available: true,
              door: "left",
            },
            {
              kind: "unlock",
              label: "Unlock Deep Cellar {4}{B}",
              cost: "{4}{B}",
              available: true,
              door: "right",
            },
          ],
        }),
      ]),
      stack: zone("stack", undefined, []),
      exile: zone("exile", undefined, []),
    } as unknown as GameView;
    const sections = buildMenuSections(view, view.battlefield.cards[0], "me", false);
    const special = sections.find((s) => s.id === "special_actions");
    expect(special?.items.map((i) => i.id)).toEqual([
      "special-unlock-left",
      "special-unlock-right",
    ]);
    expect(special?.items[1].action?.params).toMatchObject({ kind: "unlock", door: "right" });
  });
});
