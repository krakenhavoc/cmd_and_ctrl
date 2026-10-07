// playmatMode.ts — the per-device playmat setting's type and its guard
// (ADR 0128), in a module with no imports so settings.ts can use it
// without pulling the session store in behind it. playmat.ts re-exports
// all of this.

/**
 * PlaymatsMode is the per-device setting (display.playmats).
 *
 *   all   every seat's mat (the default)
 *   mine  only your own
 *   off   none: for anyone who finds other people's art distracting, or
 *         is on a low-power device
 */
export type PlaymatsMode = "all" | "mine" | "off";

export const PLAYMATS_MODES: readonly PlaymatsMode[] = ["all", "mine", "off"];
export const DEFAULT_PLAYMATS_MODE: PlaymatsMode = "all";

export function isPlaymatsMode(v: unknown): v is PlaymatsMode {
  return typeof v === "string" && (PLAYMATS_MODES as readonly string[]).includes(v);
}
