// Interface cues (ADR 0209 §2, #4394): the sounds that tell the local seat
// the game is waiting on them, before they read anything.
//
//   your_turn      your turn begins (the untap step reaches your seat)
//   priority       you hold priority and it is a real decision
//   prompt         a choice is waiting for you
//   timer_warning  a turn clock is running out (no clock exists yet; the
//                  name is in the manifest for the day one does)
//
// This module is pure and dependency-free apart from the SoundName type:
// the table feeds it one frame per snapshot and it says which cue, if any,
// to play. Game.svelte owns the wiring.
//
// "Real decision" means the table is really stopped on the viewer. A frame
// the autopass is about to pass, a timed bluff that will pass by itself, a
// busy table (mulligans, a finished game, a controlled turn) and a frame
// where a prompt is open (that one cues as a prompt) are not. That is the
// whole point of the cue: it must never fire on auto-pass.

import type { SoundName } from "./sounds";

/** The verdict of `autopassDecision`, narrowed to what the cue needs. */
export type CueVerdict = "hold" | "pass" | "clear-toggle" | { kind: "bluff" };

/** What the table knows about the local seat on one snapshot. */
export interface CueFrame {
  /** The game is being played: false in the lobby, once it has ended, or while no seat is the viewer's. */
  playing: boolean;
  /** The viewer's seat holds priority (`AutopassGates.viewerHasPriority`). */
  hasPriority: boolean;
  /** Mulligans open, game over, viewer out, or a controlled turn (`AutopassGates.tableBusy`). */
  tableBusy: boolean;
  /** The viewer has an open choice (`AutopassGates.hasPendingChoice`). */
  hasPendingChoice: boolean;
  /** What the autopass would do with this frame. */
  verdict: CueVerdict;
  /** The ids of the choices waiting on the viewer. */
  promptIDs: readonly string[];
}

/**
 * isRealPriority is true only when the viewer holds priority and the table
 * will wait for them: the autopass holds (or clears its toggle and holds),
 * nothing else owns the frame, and no choice is open.
 */
export function isRealPriority(f: CueFrame): boolean {
  if (!f.playing || !f.hasPriority || f.tableBusy || f.hasPendingChoice) return false;
  return f.verdict === "hold" || f.verdict === "clear-toggle";
}

/**
 * turnBeginCue is the sound for a step change: `your_turn` when the untap
 * step begins on the viewer's turn, `turn_change` when it begins on anyone
 * else's, and null for every other step change.
 */
export function turnBeginCue(
  prevStep: string | null,
  step: string,
  viewerIsActive: boolean,
): SoundName | null {
  if (prevStep === null || step === prevStep || step !== "untap") return null;
  return viewerIsActive ? "your_turn" : "turn_change";
}

/**
 * createCueTracker returns a function that takes each frame and returns the
 * cues to play for it. Cues are edges: priority sounds when the viewer goes
 * from not having a real decision to having one, a prompt sounds once for
 * each choice id it has not announced. The first frame, and the first frame
 * after `reset`, is the baseline and plays nothing, so loading or
 * reconnecting into a game that is already waiting on you is silent.
 */
export function createCueTracker(): {
  observe(frame: CueFrame): SoundName[];
  reset(): void;
} {
  let baselined = false;
  let hadPriority = false;
  let announced = new Set<string>();
  return {
    observe(f: CueFrame): SoundName[] {
      const real = isRealPriority(f);
      const ids = new Set(f.promptIDs);
      if (!baselined) {
        baselined = true;
        hadPriority = real;
        announced = ids;
        return [];
      }
      const cues: SoundName[] = [];
      // A choice you have not heard about yet. One cue however many arrive
      // at once; ids that are gone are forgotten so a repeat can sound.
      let freshPrompt = false;
      for (const id of ids) if (!announced.has(id)) freshPrompt = true;
      announced = ids;
      if (freshPrompt && f.playing) cues.push("prompt");
      if (real && !hadPriority) cues.push("priority");
      hadPriority = real;
      return cues;
    },
    reset(): void {
      baselined = false;
      hadPriority = false;
      announced = new Set();
    },
  };
}
