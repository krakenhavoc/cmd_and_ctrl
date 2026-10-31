import { describe, expect, it } from "vitest";

import {
  acceptParams,
  clampLoopCount,
  defaultShorten,
  loopControl,
  loopProposalFor,
  recordLoopEntry,
  shortenAllowed,
  shortenParams,
} from "./loopShortcut";
import type { GameView, LoopProposalView, LoopRecordingView } from "./protocol";

function view(loop?: LoopRecordingView, extra: Partial<GameView> = {}): GameView {
  return {
    state: "active",
    seats: [
      { id: "me", name: "Me" },
      { id: "ann", name: "Ann" },
    ],
    loop_recording: loop,
    ...extra,
  } as unknown as GameView;
}

const proposal: LoopProposalView = {
  proposer: "ann",
  name: "Basalt Monolith",
  count: 10,
  end: { iteration: 10, step: 4 },
  steps: [
    { actor: "ann", label: "Ann activates Basalt Monolith" },
    { actor: "ann", label: "Ann passes" },
    { actor: "me", label: "Me passes" },
    { actor: "ann", label: "Ann passes" },
  ],
  choice_points: [2],
};

describe("loopControl", () => {
  it("shows nothing on the dock while no loop stands (Record loop is the ⋯ menu's)", () => {
    expect(loopControl(view(), "me", true).mode).toBe("none");
    expect(loopControl(view(), null, true).mode).toBe("none");
  });

  it("follows the recorder's own recording", () => {
    const rec = { recorder: "me", status: "recording", steps: 3, max_iterations: 1000 };
    expect(loopControl(view(rec), "me", false)).toEqual({
      mode: "recording",
      steps: 3,
      text: "",
      stopping: false,
      canPropose: false,
    });
    expect(loopControl(view({ ...rec, status: "proposed" }), "me", false).mode).toBe("proposed");
  });

  it("shows an armed recording and a stop that waits for priority (#4736)", () => {
    const armed = { recorder: "me", status: "armed", steps: 0, max_iterations: 1000 };
    expect(loopControl(view(armed), "me", false).mode).toBe("armed");
    const stopping = {
      recorder: "me",
      status: "recording",
      steps: 4,
      stopping: true,
      max_iterations: 1000,
    };
    expect(loopControl(view(stopping), "me", false)).toMatchObject({
      mode: "recording",
      stopping: true,
    });
  });

  it("lets a ready loop be proposed only by its recorder holding priority with nothing open (CR 732.2a)", () => {
    const ready = { recorder: "me", status: "ready", steps: 2, max_iterations: 1000 };
    expect(loopControl(view(ready), "me", true)).toMatchObject({ mode: "ready", canPropose: true });
    expect(loopControl(view(ready), "me", false)).toMatchObject({
      mode: "ready",
      canPropose: false,
    });
    const prompted = view(ready, {
      pending_choices: [{ id: "c", kind: "pay_unless", chooser: "ann" }],
    } as Partial<GameView>);
    expect(loopControl(prompted, "me", true).canPropose).toBe(false);
  });

  it("tells every other seat what is being built (CR 732.1a)", () => {
    const rec = { recorder: "ann", status: "recording", steps: 2, max_iterations: 1000 };
    expect(loopControl(view(rec), "me", true)).toMatchObject({
      mode: "watching",
      steps: 2,
      text: "Ann is recording a loop",
    });
    expect(loopControl(view({ ...rec, status: "armed" }), "me", true).text).toBe(
      "Ann will record a loop from their next priority",
    );
    const proposed = { ...rec, status: "proposed", name: "Basalt Monolith", count: 50 };
    expect(loopControl(view(proposed), "me", false).text).toBe(
      "Ann proposes repeating Basalt Monolith 50 more times",
    );
  });
});

// #4736 (owner report, 2026-10-30): "I was in the middle of resolving an
// ability and could not demo my loop as it was hidden."
describe("recordLoopEntry", () => {
  const seat = { seated: true, eliminated: false, controlled: false, viewerHasPriority: true };

  it("is enabled for a seat holding priority with nothing open, with no note", () => {
    expect(recordLoopEntry(view(), "me", seat)).toEqual({ enabled: true, note: "" });
  });

  it("stays enabled partway through a resolution, and says when it begins", () => {
    const resolving = view(undefined, {
      pending_choices: [{ id: "c", kind: "confirm", chooser: "me" }],
    } as Partial<GameView>);
    expect(recordLoopEntry(resolving, "me", seat)).toEqual({
      enabled: true,
      note: "Starts when you next hold priority",
    });
    expect(recordLoopEntry(view(), "me", { ...seat, viewerHasPriority: false })).toEqual({
      enabled: true,
      note: "Starts when you next hold priority",
    });
    const triggers = view(undefined, { pending_triggers: [{ id: "t" }] } as Partial<GameView>);
    expect(recordLoopEntry(triggers, "me", seat).note).toBe("Starts when you next hold priority");
  });

  it("is disabled, with its one-line reason, where nobody here could record", () => {
    const off = (v: GameView, id: string | null, ctx = seat) => recordLoopEntry(v, id, ctx);
    expect(off(view(), null, { ...seat, seated: false })).toEqual({
      enabled: false,
      note: "Only a player at the table can record a loop",
    });
    expect(off(view(), "me", { ...seat, eliminated: true }).note).toBe("You have left the game");
    expect(off(view(undefined, { state: "ended" }), "me").note).toBe("The game is over");
    expect(off(view(undefined, { state: "lobby" }), "me").note).toBe("The game has not started");
    expect(off(view(undefined, { mulligans_open: true }), "me").note).toBe(
      "Not until every opening hand is kept",
    );
    expect(off(view(), "me", { ...seat, controlled: true }).note).toBe(
      "Not while another player makes your decisions",
    );
    const theirs = view({ recorder: "ann", status: "recording", steps: 1, max_iterations: 1000 });
    expect(off(theirs, "me")).toEqual({ enabled: false, note: "Ann is recording a loop" });
    const mine = view({ recorder: "me", status: "armed", steps: 0, max_iterations: 1000 });
    expect(off(mine, "me")).toEqual({
      enabled: false,
      note: "You have a loop going: its controls are on the dock",
    });
  });
});

describe("the proposal", () => {
  it("clamps a count to 1..max", () => {
    expect(clampLoopCount(0, 1000)).toBe(1);
    expect(clampLoopCount(5000, 1000)).toBe(1000);
    expect(clampLoopCount(12.7, 1000)).toBe(12);
    expect(clampLoopCount(Number.NaN, 1000)).toBe(1);
  });

  it("finds the viewer's prompt", () => {
    const v = view(undefined, {
      pending_choices: [{ id: "q", kind: "loop_proposal", chooser: "me", loop_proposal: proposal }],
    } as Partial<GameView>);
    expect(loopProposalFor(v, "me")?.id).toBe("q");
    expect(loopProposalFor(v, "ann")).toBeNull();
  });

  it("shortens only to the chooser's own step, no later than the end", () => {
    expect(defaultShorten(proposal)).toEqual({ iteration: 1, step: 2 });
    expect(defaultShorten({ ...proposal, choice_points: [] })).toBeNull();
    expect(shortenAllowed(proposal, { iteration: 3, step: 2 })).toBe(true);
    expect(shortenAllowed(proposal, { iteration: 3, step: 1 })).toBe(false);
    expect(shortenAllowed(proposal, { iteration: 11, step: 2 })).toBe(false);
    expect(shortenAllowed(proposal, { iteration: 0, step: 2 })).toBe(false);
    const shortened = { ...proposal, end: { iteration: 4, step: 1 } };
    expect(shortenAllowed(shortened, { iteration: 4, step: 2 })).toBe(false);
    expect(shortenAllowed(shortened, { iteration: 3, step: 2 })).toBe(true);
  });

  it("builds the answer params", () => {
    expect(acceptParams("q")).toEqual({ choice_id: "q", accept: true });
    expect(shortenParams("q", { iteration: 2, step: 2 })).toEqual({
      choice_id: "q",
      shorten: { iteration: 2, step: 2 },
    });
  });
});
