// Mirror of the v0 protocol types from server/internal/protocol/protocol.go
// and server/internal/protocol/view.go. When the spec in docs/protocol.md
// evolves, update both sides in lockstep.

export const PROTOCOL_VERSION = 0;

export type Kind = "ping" | "pong" | "error" | "action" | "snapshot" | "chat";

export const ErrorCode = {
  BadVersion: "bad_version",
  BadJSON: "bad_json",
  BadRequest: "bad_request",
  Internal: "internal",
} as const;

export type ErrorCodeValue = (typeof ErrorCode)[keyof typeof ErrorCode];

export interface Frame<P = unknown> {
  v: number;
  kind: Kind;
  id: string;
  payload?: P;
}

export interface PingPayload {
  msg?: string;
}

export interface PongPayload {
  msg?: string;
  server_time: string;
}

export interface ErrorPayload {
  code: string;
  message: string;
}

// ActionPayload is sent by the client to mutate game state. See
// docs/protocol.md for the full catalog of action types and their
// params shapes.
export interface ActionPayload {
  type: string;
  player?: string;
  params?: unknown;
}

// SnapshotPayload is the server's authoritative view of the game,
// broadcast after every successful action and sent once on connect.
export interface SnapshotPayload {
  seq: number;
  game: GameView;
}

// ChatPayload is the body of a Kind == "chat" frame in either
// direction. When the client sends one, only `text` is honoured —
// the server stamps `author_id`, `author_name`, and `timestamp` from
// the connection's principal before broadcasting. Spectator chat
// arrives with `author_id` empty and `author_name` == "spectator".
export interface ChatPayload {
  author_id?: string;
  author_name: string;
  text: string;
  timestamp: string;
}

// View types mirror server/internal/protocol/view.go.

export interface GameView {
  id: string;
  state: "lobby" | "active" | "ended";
  seats: PlayerView[];
  battlefield: ZoneView;
  stack: ZoneView;
  exile: ZoneView;
  turn: TurnView;
  // True between Start and the moment all seated, non-eliminated
  // players have committed to their opening hand via the keep_hand
  // action. Drives the keep / mulligan dialog and gates the normal
  // game UI. Added in S08.
  mulligans_open: boolean;
  // Player ID of the current monarch (Conspiracy mechanic). Empty /
  // omitted when no monarch is set. Sandbox marker; the must-attack
  // and combat-damage transfer rules are not enforced server-side.
  // Added in S10.
  monarch?: string;
  // Player ID currently holding the initiative (BG3 mechanic). Empty
  // when unassigned. Same sandbox posture as monarch. Added in S10.
  initiative?: string;
  // Per-pair "I owe you" promise tally as "{from}->{to}" string keys
  // → count. Sparse: zero entries are dropped server-side. Added in
  // S10.
  promises?: Record<string, number>;
  // Currently open council's-dilemma / politics vote, or omitted when
  // none is in progress. Added in S10.
  vote?: VoteView;
  // Per-player per-turn undo budget. Refreshed on each player's untap
  // step. Drives the "undos: N" indicator and gates the undo button.
  // Added in S11.
  undo_limit?: number;
}

export interface VoteView {
  id: string;
  topic: string;
  options: string[];
  initiator: string;
  // voter player ID → option index
  ballots: Record<string, number>;
}

export interface PlayerView {
  id: string;
  name: string;
  seat: number;
  life: number;
  poison?: number;
  energy?: number;
  library: ZoneView;
  hand: ZoneView;
  graveyard: ZoneView;
  command: ZoneView;
  commander_damage: Record<string, number>;
  // Rolling per-player life-change log. Bounded server-side at
  // MaxLifeHistoryEntries (currently 50). Always emitted as an array
  // (possibly empty); each entry is server-stamped at the moment of
  // change. Added in S08.
  life_history: LifeChangeView[];
  // Set when the player has conceded (S08) or, in the future, lost
  // to a state-based action. Eliminated players still appear in the
  // seats list and may continue to spectate; the UI greys them out
  // and disables their action buttons. Omitempty on the wire — only
  // present when true.
  eliminated?: boolean;
  // True once the player has committed to their opening hand during
  // the mulligan window. Omitempty on the wire — absent means false.
  // Added in S08.
  hand_kept?: boolean;
  // Number of mulligans this player has taken in the current
  // opening-hand window. Omitempty on the wire — absent means 0.
  // Added in S08.
  mulligans_taken?: number;
  // True once the player has had a real deck installed via the
  // POST /games/{id}/decks endpoint. Absent / false means the seat
  // is still using the placeholder commander handed out at join
  // time. Drives the in-game deck-import modal. Added in S08.5.
  deck_imported?: boolean;
  // Per-turn undo budget left for this seat, refreshed on entering
  // their untap step. Drives the per-seat indicator on the toolbar.
  // Added in S11.
  undos_remaining?: number;
  // Discord identity (S12.5). Populated when the seat was claimed
  // via OAuth; absent for manual-name joins. The client builds
  // /avatars/{discord_id}/{discord_avatar_hash}.png to pull the
  // cached portrait, and prefers display_name over name for the
  // seat label.
  discord_id?: string;
  discord_avatar_hash?: string;
  display_name?: string;
}

export interface LifeChangeView {
  delta: number;
  new_total: number;
  at: string; // RFC3339
}

export interface ZoneView {
  kind: string;
  owner?: string;
  count: number;
  cards: CardView[];
}

export interface CardView {
  instance_id: string;
  name: string;
  owner: string;
  controller: string;
  scryfall_id?: string;
  // Scryfall printed type line ("Legendary Creature — Human Wizard").
  // Used by the client to filter creature-only UIs (combat panel)
  // and to label cards. Omitted for placeholder demo cards. S08.
  type_line?: string;
  // Parsed printed creature stats. Omitted (zero) for non-creatures
  // and for cards with non-numeric printed stats. S08.
  power?: number;
  toughness?: number;
  tapped?: boolean;
  counters?: Record<string, number>;
  is_commander?: boolean;
  // Normalised battlefield position in [0, 1]. Only meaningful for
  // cards on the battlefield zone; server clears to 0 on exit and
  // omits the fields for cards that have never been positioned.
  battle_x?: number;
  battle_y?: number;
  // Player ID this card is currently declared to attack. Omitted
  // when not declared as attacker. Cleared on zone exit and by
  // clear_combat. Added in S08.
  attacking_target?: string;
  // Attacker instance ID this card is currently declared to block.
  // Omitted when not declared as blocker. Cleared on zone exit and
  // by clear_combat. Added in S08.
  blocking_target?: string;
  // Player ID who goaded this creature, or omitted when not goaded.
  // Cleared on zone exit. Sandbox marker; must-attack-not-the-goader
  // is not enforced server-side. Added in S10.
  goaded_by?: string;
}

export interface TurnView {
  number: number;
  active_seat: number;
  // priority_holder is the seat index that currently holds priority
  // within the step. Equal to active_seat at every step boundary;
  // rotates on pass_priority. Added in S07.
  priority_holder: number;
  phase: string;
  step: string;
}

// uuid generates a v4 UUID. Uses crypto.randomUUID when available (all
// modern browsers). Falls back to a simple implementation for older
// environments so tests can run under Node without polyfills.
export function uuid(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  // RFC 4122 v4, not cryptographically strong — only used as a fallback.
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}
