// @vitest-environment jsdom
//
// adventureCast.test.ts — #992, the client half.
//
// Casting the Adventure half of an adventure card (CR 715.3) walked
// the Board's cast chain with every announce prompt missing, because
// `cardAsFace` CLEARED the announce fields when the player picked
// face 1 and the server published none to put back. So Stomp reached
// `sendAction("cast_spell", …)` with no target picker ever opening and
// the announce was refused.
//
// The chain is a sequence of seams, each asking one question of the
// card it was handed, and this file follows them in order the way
// impulseCast.test.ts does for #874:
//
//   handlePlayCard  needsFacePicker            -> the face picker
//   confirmFace     cardAsFace(card, face)     -> the chosen half
//   afterFace       alternativeCostsOf         -> the cost picker
//   continueCast    card.target_mode           -> targeting
//   begin           card.legal_targets/clauses -> the picker's legal set
//
// Nothing in the chain changed. What changed is that the card it is
// handed after the face pick is now the FACE, announce block and all.

import { describe, it, expect, afterEach, vi } from "vitest";

// Hand plays a deal-in sound from an action on every card it mounts,
// and jsdom's HTMLMediaElement.play returns undefined rather than a
// promise. Stubbed here so the render below is about the markup.
vi.mock("./sounds", () => ({ play: () => {}, arm: () => {}, preload: () => {} }));

import { cardAsFace, needsFacePicker } from "./faces";
import { alternativeCostsOf, begin, isModal, stepsFor, targeting } from "./targeting";
import type { CardView, ZoneView } from "./protocol";
import Hand from "./components/board/Hand.svelte";
import { render, cleanup } from "./test/render.svelte";

afterEach(() => {
  targeting.set(null);
  cleanup();
});

// bonecrusher is the card #992 ships as its proof, as the wire now
// sends it: the creature half announces nothing and the Adventure half
// carries its own target clause.
function bonecrusher(): CardView {
  return {
    instance_id: "bg1",
    name: "Bonecrusher Giant",
    owner: "me",
    controller: "me",
    scryfall_id: "bg",
    type_line: "Creature — Giant",
    mana_cost: "{2}{R}",
    layout: "adventure",
    faces: [
      { name: "Bonecrusher Giant", type_line: "Creature — Giant", mana_cost: "{2}{R}" },
      {
        name: "Stomp",
        type_line: "Instant — Adventure",
        mana_cost: "{1}{R}",
        target_mode: "any",
        legal_targets: { cards: ["bear"], players: ["them"], min: 1, max: 1 },
      },
    ],
  };
}

describe("casting the Adventure half", () => {
  const card = bonecrusher();

  it("asks which half first — CR 715.3", () => {
    expect(needsFacePicker(card)).toBe(true);
  });

  it("hands the chain the Adventure, announce block and all", () => {
    const stomp = cardAsFace(card, 1);
    expect(stomp.name).toBe("Stomp");
    expect(stomp.type_line).toBe("Instant — Adventure");
    expect(stomp.mana_cost).toBe("{1}{R}");
    expect(stomp.active_face).toBe(1);
    // The seam that used to see nothing.
    expect(stomp.target_mode).toBe("any");
  });

  it("opens the target picker on the CHOSEN face's legal set", () => {
    const stomp = cardAsFace(card, 1);
    expect(isModal(stomp)).toBe(false);
    expect(alternativeCostsOf(stomp)).toEqual([]);

    begin(stomp, "any", { face: 1 });
    const state = get(targeting);
    expect(state).not.toBeNull();
    expect(state?.card.name).toBe("Stomp");
    expect(state?.mode).toBe("any");
    expect([...(state?.legal?.cards ?? [])]).toEqual(["bear"]);
    expect([...(state?.legal?.players ?? [])]).toEqual(["them"]);
    // The face rides through to the announcement, so the server casts
    // the half the picker was opened for.
    expect(state?.choices?.face).toBe(1);
  });

  it("opens no picker for the creature half, which targets nothing", () => {
    const giant = cardAsFace(card, 0);
    expect(giant.name).toBe("Bonecrusher Giant");
    expect(giant.target_mode).toBeUndefined();
    // stepsFor is what `begin` would walk; with no clause and no legal
    // set the chain never reaches it, and continueCast fires
    // cast_spell straight away.
    expect(stepsFor("any", giant.legal_targets, giant.clauses, 0)).toEqual([
      { mode: "any", min: 1, max: 1, modeIndex: 0, slot: 0, distinct: false },
    ]);
  });
});

// #1166, the client side. A revealed card in ANOTHER seat's hand
// arrives without `legal_targets` and `clauses` now — they are the
// hand owner's answer to "what may YOU target" — and the reader that
// renders it must not need them.
describe("the revealed-hand reader", () => {
  function revealed(): CardView {
    return {
      instance_id: "revealed-1",
      name: "Lightning Bolt",
      owner: "them",
      controller: "them",
      type_line: "Instant",
      mana_cost: "{R}",
      known_by_you: true,
      // Public, because it is printed on a card the viewer has been
      // shown. What is absent is the owner's legal set beside it.
      target_mode: "any",
    };
  }

  function handZone(cards: CardView[], count: number): ZoneView {
    return { kind: "hand", owner: "them", count, cards };
  }

  it("renders a revealed opponent card by name with no announce block", () => {
    const { container } = render(
      Hand as never,
      {
        hand: handZone([revealed()], 4),
        isSelf: false,
        snap: null,
        viewerID: "me",
      } as never,
    );
    expect(container.textContent).toContain("Lightning Bolt");
    // Never greyed: an opponent's hand carries no move list and no
    // legal set, and the reader has always treated it as somebody
    // else's business rather than deriving a verdict from fields it
    // does not have.
    expect(container.querySelectorAll(".timing-disabled").length).toBe(0);
  });
});

// get reads a store's current value without subscribing for the life
// of the test, which is all these assertions need.
function get<T>(store: { subscribe(run: (v: T) => void): () => void }): T {
  let out!: T;
  store.subscribe((v) => {
    out = v;
  })();
  return out;
}
