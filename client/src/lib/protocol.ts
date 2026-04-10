// Mirror of the v0 protocol types from server/internal/protocol/protocol.go
// and server/internal/protocol/view.go. When the spec in docs/protocol.md
// evolves, update both sides in lockstep.

export const PROTOCOL_VERSION = 0;

export type Kind = "ping" | "pong" | "error" | "action" | "snapshot";

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

// View types mirror server/internal/protocol/view.go.

export interface GameView {
  id: string;
  state: "lobby" | "active" | "ended";
  seats: PlayerView[];
  battlefield: ZoneView;
  stack: ZoneView;
  exile: ZoneView;
  turn: TurnView;
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
  tapped?: boolean;
  counters?: Record<string, number>;
  is_commander?: boolean;
}

export interface TurnView {
  number: number;
  active_seat: number;
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
