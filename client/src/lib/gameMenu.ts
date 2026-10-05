// gameMenu.ts — ADR 0111 Delivery PR 7 (S56, #1958), owner decision 3.
//
// The ⋯ menu holds what a player may do that does not move the turn
// on: the sandbox tools (draw, untap all, shuffle, mulligan to N), life
// history, table settings, spawn, the vote launcher, the way back to the
// lobby, and Concede with its confirm. For a seated player it is the
// last chip on the action dock's toggles row and opens upward. A viewer
// with no dock (a spectator, an admin with no seat here) keeps a smaller
// one on the command bar, opening downward: what they can read and where
// they can go, and nothing that acts on a seat.
//
// GameMenu.svelte draws it; this file is what Game.svelte hands it and
// the two small rules it applies.

import { L } from "./labels";

export interface GameMenuOptions {
  // A seat of the viewer's own at this table: the sandbox tools, the
  // vote launcher and Concede. False for a spectator or an unseated
  // admin, whose menu is life history, the table and navigation only.
  seated: boolean;
  // Concede is drawn disabled, not hidden, once it means nothing.
  eliminated?: boolean;
  gameEnded?: boolean;
  // The draw key, formatted for the item's key cap ("D"), or "".
  drawKey?: string;
  // "draw a card (D)" — the tooltip, from the dock's key-hint helper.
  drawTitle?: string;
  // ADR 0075 §2.1: the host or the admin. Everyone else may still OPEN
  // table settings, read-only.
  canManage: boolean;
  // Both of the spawn route's gates. The entry is absent, not
  // disabled, without them (ADR 0075 §2.5).
  spawnAvailable: boolean;
  // "Link Discord" (a navigation), or null when this viewer has no seat
  // of their own to link (lib/myGames.ts canLinkDiscord).
  discordLink?: { href: string; label: string } | null;
  // A signed-in identity: "My games".
  myGames: boolean;
  // ADR 0112 §2 item 9: an allowlisted person's admin switch, which
  // the table has no header to hold. `on` is whether admin mode is on
  // now; null for everyone else (the token included: it has no mode).
  adminMode?: { on: boolean } | null;
  onAdminMode?: () => void;
  // An open vote is the dock's request (ADR 0111 PR 5), so the launcher
  // stands down while one is open.
  voteOpen: boolean;
  // ADR 0121 §5 and §8: "Roll a d6", "Roll a d20" and "Flip a coin", for
  // a seat of the viewer's own still in an active game (the opening roll
  // and the mulligan included); null or absent otherwise. `ready` is
  // false for TABLE_ROLL_COOLDOWN_MS after the viewer's own table roll,
  // matching the server's rate, and the items are disabled until then.
  tableRoll?: { ready: boolean } | null;
  onTableRoll?: (die: TableDie) => void;
  // ADR 0125 §6: the Help group, after Table. "Tips for the table" and
  // "Keyboard shortcuts" for everyone; "Replay the tutorial" only on
  // this tab's practice table (`practice`). A real table offers no
  // "Practice game": it would take the player out of their seat.
  practice?: boolean;
  onTableTips?: () => void;
  onShortcuts?: () => void;
  onReplayTutorial?: () => void;
  onDraw: () => void;
  onUntapAll: () => void;
  onShuffle: () => void;
  onMulligan: (handSize: number) => void;
  onLifeHistory: () => void;
  onTableSettings: () => void;
  onSpawn: () => void;
  onMyGames: () => void;
  onBack: () => void;
  // Called once the confirm is accepted. Concede is irreversible.
  onConcede: () => void;
  onStartVote: (topic: string, options: string[]) => void;
}

// ADR 0121 §5: what roll_table_die rolls.
export type TableDie = "d6" | "d20" | "coin";

// The server takes one table roll per seat per 2 s and refuses the next
// with "wait for your last roll to land"; the menu waits the same 2 s.
export const TABLE_ROLL_COOLDOWN_MS = 2000;

// The ⋯ menu's table-roll items, in order. Their names are a contract
// (ADR 0121 §8, AGENTS.md "Labels are a contract").
export const TABLE_ROLL_ITEMS: readonly { die: TableDie; label: string }[] = [
  { die: "d6", label: L.rollD6 },
  { die: "d20", label: L.rollD20 },
  { die: "coin", label: L.flipCoin },
];

// A table-roll item's tooltip.
export function tableRollTitle(ready: boolean): string {
  return ready
    ? "for fun, at the table — everyone sees it land, and nothing in the game can trigger on it"
    : "wait for your last roll to land";
}

// The mulligan-to field's value, as the server takes it: a whole number
// of cards between 0 and 20.
export function clampMulligan(n: number): number {
  if (!Number.isFinite(n)) return 7;
  return Math.max(0, Math.min(20, Math.floor(n)));
}

// The vote launcher's "options, comma-separated" field. A vote needs a
// topic and at least two options; null means "not yet".
export function parseVote(
  topic: string,
  optionsText: string,
): { topic: string; options: string[] } | null {
  const t = topic.trim();
  const options = optionsText
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
  if (!t || options.length < 2) return null;
  return { topic: t, options };
}

// Why Concede is disabled, or its ordinary tooltip.
export function concedeTitle(o: Pick<GameMenuOptions, "eliminated" | "gameEnded">): string {
  if (o.eliminated) return "you are already eliminated";
  if (o.gameEnded) return "the game has ended";
  return "concede the game (irreversible)";
}
