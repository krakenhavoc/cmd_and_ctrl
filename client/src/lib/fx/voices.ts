// Planeswalker voice lines (ADR 0209 §5, #4398). The wiring ships; the lines
// do not. VOICE_MANIFEST is EMPTY until Luke OKs shipping any (see
// client/public/voices/README.md), and an empty manifest is a silent no-op:
// no catalog fetch, no audio element, nothing.
//
// Rules this module enforces, so a caller cannot get them wrong:
//   - a walker speaks at most once every VOICE_COOLDOWN_MS;
//   - one voice line plays at a time across the whole table;
//   - lines play on the `voice` bus (master x the voice volume, from
//     sounds.ts), and the mute toggle silences them;
//   - a file that is not there is silent, never an error.

import { busGain, isMuted } from "../sounds";

/** The moments a walker can speak at. */
export type VoiceMoment = "enter" | "leave" | "ultimate";

export const VOICE_MOMENTS: readonly VoiceMoment[] = ["enter", "leave", "ultimate"];

/** Takes per moment: URLs under /voices/. A moment may be left out. */
export type VoiceLines = Partial<Record<VoiceMoment, readonly string[]>>;

/** A walker's lines, keyed by the card's Scryfall `oracle_id`. */
export type VoiceManifest = Record<string, VoiceLines>;

// To add a line: one entry here plus the file under client/public/voices/.
// The key is the card's oracle_id (the catalog page shows it); name each
// file <walker>-<moment>-<take>.mp3. Append a take to give a moment variety.
//
//   "00000000-0000-0000-0000-000000000000": {
//     // Example Walker (a placeholder id, not a real card)
//     enter: ["/voices/example-enter-1.mp3", "/voices/example-enter-2.mp3"],
//     leave: ["/voices/example-leave-1.mp3"],
//     ultimate: ["/voices/example-ultimate-1.mp3"],
//   },
export const VOICE_MANIFEST: VoiceManifest = {};

/** A walker speaks at most once in this long. */
export const VOICE_COOLDOWN_MS = 20_000;
// A line is at most 3 s (ADR 0209 §6); the lock frees itself if its `ended`
// never fires (a stalled load, a suspended tab).
const VOICE_MAX_MS = 3500;

/** Plays one file; calls `done(false)` when it ends, `done(true)` when it could not play. */
export type VoicePlayer = (url: string, volume: number, done: (failed: boolean) => void) => void;

export interface VoiceDeps {
  manifest: VoiceManifest;
  player: VoicePlayer;
  now: () => number;
  muted: () => boolean;
  volume: () => number;
  random: () => number;
}

export class VoiceBus {
  private lastSpoke = new Map<string, number>();
  private missing = new Set<string>();
  private busy = false;
  private readonly deps: VoiceDeps;

  constructor(deps: Partial<VoiceDeps> = {}) {
    this.deps = {
      manifest: VOICE_MANIFEST,
      player: audioPlayer,
      now: () => Date.now(),
      muted: isMuted,
      volume: () => busGain("voice"),
      random: Math.random,
      ...deps,
    };
  }

  /** True when no walker has any line at all: every call is a no-op. */
  get empty(): boolean {
    return Object.keys(this.deps.manifest).length === 0;
  }

  /** Whether `oracleID` has a line for `moment`. */
  has(oracleID: string, moment: VoiceMoment): boolean {
    return (this.deps.manifest[oracleID]?.[moment]?.length ?? 0) > 0;
  }

  /** Speaks a line if the walker has one, is off cooldown and nobody else is talking. */
  speak(oracleID: string, moment: VoiceMoment): boolean {
    const takes = (this.deps.manifest[oracleID]?.[moment] ?? []).filter(
      (u) => !this.missing.has(u),
    );
    if (takes.length === 0) return false;
    if (this.busy || this.deps.muted()) return false;
    const now = this.deps.now();
    const last = this.lastSpoke.get(oracleID);
    if (last !== undefined && now - last < VOICE_COOLDOWN_MS) return false;
    const url = takes[Math.min(takes.length - 1, Math.floor(this.deps.random() * takes.length))]!;
    this.busy = true;
    this.lastSpoke.set(oracleID, now);
    let finished = false;
    const finish = (failed: boolean) => {
      if (finished) return;
      finished = true;
      this.busy = false;
      if (failed) {
        // The file is not there: remember it, and let the walker try
        // again with another take rather than waiting out the cooldown.
        this.missing.add(url);
        if (this.lastSpoke.get(oracleID) === now) this.lastSpoke.delete(oracleID);
      }
    };
    try {
      this.deps.player(url, this.deps.volume(), finish);
    } catch {
      finish(true);
    }
    setTimeout(() => finish(false), VOICE_MAX_MS);
    return true;
  }
}

function audioPlayer(url: string, volume: number, done: (failed: boolean) => void): void {
  if (typeof Audio === "undefined") {
    done(true);
    return;
  }
  const audio = new Audio(url);
  audio.volume = Math.min(1, Math.max(0, volume));
  audio.addEventListener?.("ended", () => done(false));
  audio.addEventListener?.("error", () => done(true));
  // A 404 or an autoplay block rejects; both are silent.
  void audio.play().catch(() => done(true));
}

/** The table's one voice bus. */
export const voices = new VoiceBus();
