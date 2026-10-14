// playingAs.test.ts — ADR 0167 §9: which seat the board acts for while
// one player plays another's whole turn.

import { describe, expect, it } from "vitest";

import {
  NOT_PLAYING_AS,
  actingSeatID,
  actsForSeat,
  controlledByViewer,
  controllerOfViewer,
  decisionOwer,
  nextPlayingAs,
  playingAsBanner,
  togglePlayingAs,
  viewForSeat,
  type PlayingAsState,
} from "./playingAs";
import { legalActionsOf } from "./legalActions";
import type { GameView, LegalMoveView, PendingChoiceView, PlayerControlView } from "./protocol";

const ALICE = "alice";
const BOB = "bob";
const CAROL = "carol";

const EMRAKUL: PlayerControlView = {
  controller: ALICE,
  controlled: BOB,
  source_name: "Emrakul, the Promised End",
  until: "turn",
};

const move = (player: string, kind: LegalMoveView["kind"], source = ""): LegalMoveView =>
  ({
    type: kind === "pass" ? "pass_priority" : "cast_spell",
    player,
    kind,
    label: kind,
    source,
  }) as LegalMoveView;

function frame(extra: Partial<GameView> = {}, priority = 1): GameView {
  return {
    id: "g",
    state: "active",
    seats: [
      { id: ALICE, name: "Alice", seat: 0 },
      { id: BOB, name: "Bob", seat: 1 },
      { id: CAROL, name: "Carol", seat: 2 },
    ],
    turn: { seq: 7, number: 7, active_seat: 1, priority_holder: priority, step: "precombat_main" },
    ...extra,
  } as unknown as GameView;
}

const controlled = (extra: Partial<GameView> = {}, priority = 1) =>
  frame({ player_controls: [EMRAKUL], ...extra }, priority);

describe("who controls whom", () => {
  it("finds the turn the viewer plays and the one played for them", () => {
    const v = controlled();
    expect(controlledByViewer(v, ALICE)).toEqual(EMRAKUL);
    expect(controllerOfViewer(v, ALICE)).toBeNull();
    expect(controllerOfViewer(v, BOB)).toEqual(EMRAKUL);
    expect(controlledByViewer(v, BOB)).toBeNull();
    expect(controlledByViewer(v, CAROL)).toBeNull();
    expect(controllerOfViewer(v, CAROL)).toBeNull();
    expect(controlledByViewer(v, null)).toBeNull();
  });

  it("is not Word of Command's control, which lasts one resolution", () => {
    const v = frame({ player_controls: [{ ...EMRAKUL, until: "resolution" }] });
    expect(controlledByViewer(v, ALICE)).toBeNull();
    expect(controllerOfViewer(v, BOB)).toBeNull();
  });
});

describe("actingSeatID", () => {
  it("is the controlled seat on its board, the viewer otherwise", () => {
    const v = controlled();
    expect(actingSeatID(v, ALICE, "controlled")).toBe(BOB);
    expect(actingSeatID(v, ALICE, "self")).toBe(ALICE);
    expect(actingSeatID(v, BOB, "controlled")).toBe(BOB);
    expect(actingSeatID(frame(), ALICE, "controlled")).toBe(ALICE);
    expect(actingSeatID(v, null, "controlled")).toBeNull();
  });
});

describe("decisionOwer", () => {
  const choice = (extra: Partial<PendingChoiceView>): PendingChoiceView =>
    ({ id: "c", kind: "confirm", chooser: ALICE, ...extra }) as PendingChoiceView;

  it("reads a prompt first: Bob's when Alice answers it for him, else hers", () => {
    expect(
      decisionOwer(controlled({ pending_choices: [choice({ on_behalf_of: BOB })] }), ALICE, BOB),
    ).toBe(BOB);
    expect(decisionOwer(controlled({ pending_choices: [choice({})] }), ALICE, BOB)).toBe(ALICE);
  });

  it("reads the cleanup discard, then priority", () => {
    expect(decisionOwer(controlled({ discard_pending: { [BOB]: 2 } }, 2), ALICE, BOB)).toBe(BOB);
    expect(decisionOwer(controlled({}, 1), ALICE, BOB)).toBe(BOB);
    expect(decisionOwer(controlled({}, 0), ALICE, BOB)).toBe(ALICE);
    expect(decisionOwer(controlled({}, 2), ALICE, BOB)).toBeNull();
  });

  it("reads a block declaration Alice owes", () => {
    const v = controlled({}, 2);
    (v.turn as unknown as { block_pending_seats: number[] }).block_pending_seats = [0];
    expect(decisionOwer(v, ALICE, BOB)).toBe(ALICE);
  });
});

describe("nextPlayingAs: the toggle", () => {
  const fold = (frames: GameView[], start: PlayingAsState = NOT_PLAYING_AS) =>
    frames.reduce((s, v) => nextPlayingAs(s, v, ALICE), start);

  it("swaps as the controlled turn begins", () => {
    const s = fold([frame(), controlled({}, 1)]);
    expect(s.control).toBe(BOB);
    expect(actingSeatID(controlled(), ALICE, s.pref)).toBe(BOB);
  });

  it("switches by itself at a handoff, and only then", () => {
    let s = fold([controlled({}, 1)]);
    expect(s.pref).toBe("controlled");
    // Alice gets priority for herself (CR 723.8): her board.
    s = nextPlayingAs(s, controlled({}, 0), ALICE);
    expect(s.pref).toBe("self");
    // Carol decides: nothing moves.
    s = nextPlayingAs(s, controlled({}, 2), ALICE);
    expect(s.pref).toBe("self");
    // Back to Bob: his board.
    s = nextPlayingAs(s, controlled({}, 1), ALICE);
    expect(s.pref).toBe("controlled");
  });

  it("keeps a manual toggle while the same seat is still deciding", () => {
    let s = fold([controlled({}, 1)]);
    s = togglePlayingAs(s);
    expect(s.pref).toBe("self");
    const again = nextPlayingAs(s, controlled({ stack_items: [] }, 1), ALICE);
    expect(again).toBe(s);
    expect(again.pref).toBe("self");
  });

  it("returns the same state for an unchanged frame, and resets when the control ends", () => {
    const s = fold([controlled({}, 1)]);
    expect(nextPlayingAs(s, controlled({}, 1), ALICE)).toBe(s);
    expect(nextPlayingAs(s, frame(), ALICE)).toBe(NOT_PLAYING_AS);
    expect(nextPlayingAs(NOT_PLAYING_AS, frame(), ALICE)).toBe(NOT_PLAYING_AS);
  });

  it("toggles nothing with no control", () => {
    expect(togglePlayingAs(NOT_PLAYING_AS)).toBe(NOT_PLAYING_AS);
  });
});

describe("viewForSeat and the legal-action lookup", () => {
  const v = controlled({
    legal_moves: [
      move(BOB, "pass"),
      move(BOB, "cast", "bobs-bolt"),
      move(ALICE, "cast", "alices-bolt"),
    ],
    legal_actions: { pass: false, sources: { "alices-bolt": { kinds: ["cast"], moves: 1 } } },
    controlled_legal_actions: {
      [BOB]: { pass: true, sources: { "bobs-bolt": { kinds: ["cast"], moves: 1 } } },
    },
  } as unknown as Partial<GameView>);

  it("gives each seat its own moves and digest", () => {
    const asBob = viewForSeat(v, BOB, ALICE);
    expect(asBob.legal_moves?.map((m) => m.player)).toEqual([BOB, BOB]);
    expect(asBob.legal_actions?.pass).toBe(true);
    expect(asBob.controlled_legal_actions).toBeUndefined();
    const asAlice = viewForSeat(v, ALICE, ALICE);
    expect(asAlice.legal_moves?.map((m) => m.source)).toEqual(["alices-bolt"]);
    expect(asAlice.legal_actions?.pass).toBe(false);
  });

  it("feeds legalActionsOf(view, actingID)", () => {
    expect(legalActionsOf(v, BOB).isReady("bobs-bolt")).toBe(true);
    expect(legalActionsOf(v, BOB).isReady("alices-bolt")).toBe(false);
    expect(legalActionsOf(v, BOB).pass).toBe(true);
    expect(legalActionsOf(v, ALICE).isReady("alices-bolt")).toBe(true);
    expect(legalActionsOf(v, ALICE).isReady("bobs-bolt")).toBe(false);
  });

  it("leaves a frame with no control alone", () => {
    const plain = frame({ legal_moves: [move(ALICE, "pass")] });
    expect(viewForSeat(plain, ALICE, ALICE)).toBe(plain);
  });

  it("makes the prompts Alice answers for Bob his own on his board, and nobody else's", () => {
    const p = controlled({
      pending_choices: [
        { id: "for-bob", kind: "confirm", chooser: ALICE, on_behalf_of: BOB },
        { id: "hers", kind: "confirm", chooser: ALICE },
      ] as PendingChoiceView[],
    });
    const asBob = viewForSeat(p, BOB, ALICE);
    expect(asBob.pending_choices?.map((c) => c.chooser)).toEqual([BOB, ALICE]);
    // Bob's own frame: the prompt stays Alice's, so his board asks him nothing.
    expect(viewForSeat(p, BOB, BOB).pending_choices?.map((c) => c.chooser)).toEqual([ALICE, ALICE]);
    expect(legalActionsOf(p, BOB).known).toBe(false);
  });
});

describe("actsForSeat", () => {
  it("is the rules verbs, not blocks, settings or the sandbox", () => {
    for (const t of [
      "pass_priority",
      "end_turn",
      "cast_spell",
      "declare_attackers",
      "resolve_choice",
      "discard_selection",
    ]) {
      expect(actsForSeat(t), t).toBe(true);
    }
    for (const t of [
      "declare_blocker",
      "finish_blocks",
      "draw_card",
      "concede",
      "undo",
      "set_auto_answers",
    ]) {
      expect(actsForSeat(t), t).toBe(false);
    }
  });
});

describe("playingAsBanner", () => {
  it("offers the controller the toggle and the handoff", () => {
    const v = controlled({}, 0);
    const on = playingAsBanner(v, ALICE, BOB)!;
    expect(on.kind).toBe("playing");
    expect(on.text).toBe("Playing as Bob");
    expect(on.source).toBe("Emrakul, the Promised End");
    expect(on.toggle?.label).toBe("View my board");
    // Alice holds priority for herself.
    expect(on.switchTo?.label).toBe("Switch to you");

    const off = playingAsBanner(controlled({}, 1), ALICE, ALICE)!;
    expect(off.kind).toBe("viewing");
    expect(off.text).toBe("Viewing your board, Bob's turn is yours");
    expect(off.toggle?.label).toBe("Play as Bob");
    expect(off.switchTo?.label).toBe("Switch to Bob");

    expect(playingAsBanner(controlled({}, 1), ALICE, BOB)!.switchTo).toBeUndefined();
  });

  it("tells the controlled player, and who is choosing for them", () => {
    const quiet = playingAsBanner(controlled(), BOB, BOB)!;
    expect(quiet.kind).toBe("controlled");
    expect(quiet.text).toBe("Alice controls you this turn");
    expect(quiet.note).toBeUndefined();
    expect(quiet.toggle).toBeUndefined();
    const asked = playingAsBanner(
      controlled({
        pending_choices: [
          { id: "c", kind: "confirm", chooser: ALICE, on_behalf_of: BOB },
        ] as PendingChoiceView[],
      }),
      BOB,
      BOB,
    )!;
    expect(asked.note).toBe("Alice is choosing for you");
  });

  it("gives a spectator and a bystander the text only", () => {
    for (const viewer of [null, CAROL]) {
      const line = playingAsBanner(controlled(), viewer, viewer)!;
      expect(line.kind).toBe("watching");
      expect(line.text).toBe("Alice controls Bob this turn");
      expect(line.toggle).toBeUndefined();
      expect(line.switchTo).toBeUndefined();
    }
  });

  it("warns ahead from scheduled_player_controls, the last one listed applying", () => {
    const v = frame({
      scheduled_player_controls: [
        { controller: CAROL, controlled: BOB, source_name: "Emrakul, the Promised End" },
        { controller: ALICE, controlled: BOB, source_name: "Emrakul, the Promised End" },
      ],
    });
    expect(playingAsBanner(v, BOB, BOB)!.text).toBe("Alice will control your next turn");
    expect(playingAsBanner(v, ALICE, ALICE)!.text).toBe("You will control Bob's next turn");
    expect(playingAsBanner(v, null, null)!.text).toBe("Alice will control Bob's next turn");
    expect(playingAsBanner(frame(), ALICE, ALICE)).toBeNull();
  });
});
