// counterTypes.ts mirrors the server's named counter registry from
// server/internal/game/counter_types.go. The server-side map is the
// source of truth — keep these strings byte-equal to its constants.
//
// Each entry carries an optional pinned colour + glyph for the
// pip-overlay UI. Counter names not in the map render as text-only
// pips with a neutral colour.

// Card-level counter type identifiers.
export const COUNTER_PLUS_ONE = "+1/+1";
export const COUNTER_MINUS_ONE = "-1/-1";
export const COUNTER_LOYALTY = "loyalty";
export const COUNTER_DEFENSE = "defense";
export const COUNTER_CHARGE = "charge";
export const COUNTER_STUN = "stun";
export const COUNTER_SHIELD = "shield";
export const COUNTER_LORE = "lore";

// Player-level counter type identifiers.
export const COUNTER_POISON = "poison";
export const COUNTER_ENERGY = "energy";
export const COUNTER_EXPERIENCE = "experience";
export const COUNTER_RAD = "rad";

// PoisonLethal mirrors the server's PoisonLethal const — the
// poison-counter total at which a player loses (CR 704.5c).
export const POISON_LETHAL = 10;

export interface CounterStyle {
  // Short label for the pip chip (typically 2-3 chars).
  abbr: string;
  // CSS colour for the pip background.
  color: string;
  // Optional emoji / icon glyph for the player-counter panel.
  glyph?: string;
}

// COUNTER_STYLES is the iconography pin map. Unknown counter names
// fall back to neutral styling (see counterStyle() below).
export const COUNTER_STYLES: Record<string, CounterStyle> = {
  [COUNTER_PLUS_ONE]: { abbr: "+1", color: "#5fbf67", glyph: "▲" },
  [COUNTER_MINUS_ONE]: { abbr: "-1", color: "#c2536b", glyph: "▼" },
  [COUNTER_LOYALTY]: { abbr: "L", color: "#7a5fbf", glyph: "★" },
  [COUNTER_DEFENSE]: { abbr: "D", color: "#3f6e8a", glyph: "🛡" },
  [COUNTER_CHARGE]: { abbr: "C", color: "#bf9a5f", glyph: "⚡" },
  [COUNTER_STUN]: { abbr: "S", color: "#a5a5a5", glyph: "💫" },
  [COUNTER_SHIELD]: { abbr: "Sh", color: "#5f9abf", glyph: "🛡" },
  [COUNTER_LORE]: { abbr: "Lo", color: "#bfa55f", glyph: "📜" },
  [COUNTER_POISON]: { abbr: "Poi", color: "#7ab85f", glyph: "🟢" },
  [COUNTER_ENERGY]: { abbr: "E", color: "#e0c050", glyph: "⚡" },
  [COUNTER_EXPERIENCE]: { abbr: "Exp", color: "#d0a050", glyph: "⭐" },
  [COUNTER_RAD]: { abbr: "Rad", color: "#a8c050", glyph: "☢" },
};

const NEUTRAL_STYLE: CounterStyle = { abbr: "?", color: "#6c7a99" };

// counterStyle resolves a counter name to its pin style, falling
// back to a neutral style with the first 3 characters as the abbr.
export function counterStyle(name: string): CounterStyle {
  const known = COUNTER_STYLES[name];
  if (known) return known;
  return { ...NEUTRAL_STYLE, abbr: name.slice(0, 3) };
}

// PINNED_PLAYER_COUNTERS is the ordered list of counters that the
// player-counter panel (PlayerHeader) renders as always-visible
// rows. Other counters appear only when non-zero.
export const PINNED_PLAYER_COUNTERS: readonly string[] = [
  COUNTER_POISON,
  COUNTER_ENERGY,
  COUNTER_EXPERIENCE,
  COUNTER_RAD,
];
