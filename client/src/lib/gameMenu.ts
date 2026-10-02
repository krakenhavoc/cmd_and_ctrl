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
  // An open vote is the dock's request (ADR 0111 PR 5), so the launcher
  // stands down while one is open.
  voteOpen: boolean;
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
