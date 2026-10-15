import { describe, expect, it } from "vitest";

import {
  acceptParams,
  clampLoopCount,
  defaultShorten,
  loopControl,
  loopProposalFor,
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
  it("offers Record only to the priority holder with no prompt open", () => {
    expect(loopControl(view(), "me", true).mode).toBe("record");
    expect(loopControl(view(), "me", false).mode).toBe("none");
    expect(loopControl(view(), null, true).mode).toBe("none");
    const prompted = view(undefined, {
      pending_choices: [{ id: "c", kind: "pay_unless", chooser: "me" }],
    } as Partial<GameView>);
    expect(loopControl(prompted, "me", true).mode).toBe("none");
  });

  it("follows the recorder's own recording", () => {
    const rec = { recorder: "me", status: "recording", steps: 3, max_iterations: 1000 };
    expect(loopControl(view(rec), "me", false)).toEqual({ mode: "recording", steps: 3, text: "" });
    expect(loopControl(view({ ...rec, status: "ready" }), "me", true).mode).toBe("ready");
    expect(loopControl(view({ ...rec, status: "proposed" }), "me", false).mode).toBe("proposed");
  });

  it("tells every other seat what is being built (CR 732.1a)", () => {
    const rec = { recorder: "ann", status: "recording", steps: 2, max_iterations: 1000 };
    expect(loopControl(view(rec), "me", true)).toEqual({
      mode: "watching",
      steps: 2,
      text: "Ann is recording a loop",
    });
    const proposed = { ...rec, status: "proposed", name: "Basalt Monolith", count: 50 };
    expect(loopControl(view(proposed), "me", false).text).toBe(
      "Ann proposes repeating Basalt Monolith 50 more times",
    );
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
