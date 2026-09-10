// Replay parsing for the develop environment's scrubber (ADR 0023).
//
// GET /games/{id}/replay streams the per-game log written by
// ws/room.go: JSONL, one protocol.SnapshotPayload per successful
// Apply, in order. Every line is a complete unfiltered GameView, so
// stepping the timeline is a matter of picking a line — no
// re-simulation, and no server work beyond the route that already
// exists.
//
// "Unfiltered" is worth stating plainly: a replay shows every hand
// and every library. That is the point when you are debugging, and
// the reason the route is admin-gated for a game still in progress.

import type { GameView } from "./protocol";

export interface ReplayFrame {
  seq: number;
  game: GameView;
}

// parseReplay turns the JSONL body into frames, skipping anything it
// cannot use rather than failing the whole load.
//
// The tolerance is not defensive padding: the log is appended to
// while the game runs, so a download taken mid-write can legitimately
// end in a half-written line. Throwing there would make the scrubber
// unusable on exactly the games you most want to inspect — the live
// ones.
export function parseReplay(body: string): ReplayFrame[] {
  const out: ReplayFrame[] = [];
  for (const line of body.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    let parsed: unknown;
    try {
      parsed = JSON.parse(trimmed);
    } catch {
      continue; // torn final line, or a log that grew a comment
    }
    if (typeof parsed !== "object" || parsed === null) continue;
    const p = parsed as { seq?: unknown; game?: unknown };
    if (typeof p.seq !== "number") continue;
    if (typeof p.game !== "object" || p.game === null) continue;
    out.push({ seq: p.seq, game: p.game as GameView });
  }
  return out;
}

// frameLabel is the one-line description shown beside the scrubber.
// Falls back progressively: a frame from an older log, or one written
// before the turn cursor was meaningful, still gets a usable label.
export function frameLabel(frame: ReplayFrame | null): string {
  if (!frame) return "—";
  const t = frame.game?.turn;
  if (!t) return `seq ${frame.seq}`;
  const seat = frame.game?.seats?.[t.active_seat]?.name ?? `seat ${t.active_seat}`;
  const step = t.step || t.phase || "";
  return `T${t.number} · ${seat}${step ? ` · ${step}` : ""} · seq ${frame.seq}`;
}

// describeDelta reports what changed between two frames, at the
// coarse level the scrubber can show without re-deriving the game.
// Used for the "what did this step do" readout.
export function describeDelta(prev: ReplayFrame | null, next: ReplayFrame | null): string {
  if (!next) return "";
  if (!prev) return "first recorded state";
  const parts: string[] = [];
  const bf = (f: ReplayFrame) => f.game?.battlefield?.cards?.length ?? 0;
  const stack = (f: ReplayFrame) => f.game?.stack?.cards?.length ?? 0;
  const dBf = bf(next) - bf(prev);
  const dStack = stack(next) - stack(prev);
  if (dBf) parts.push(`${dBf > 0 ? "+" : ""}${dBf} battlefield`);
  if (dStack) parts.push(`${dStack > 0 ? "+" : ""}${dStack} stack`);
  for (const s of next.game?.seats ?? []) {
    const before = prev.game?.seats?.find((p) => p.id === s.id);
    if (!before) continue;
    if (before.life !== s.life) parts.push(`${s.name} ${before.life}→${s.life}`);
    const dh = (s.hand?.count ?? 0) - (before.hand?.count ?? 0);
    if (dh) parts.push(`${s.name} hand ${dh > 0 ? "+" : ""}${dh}`);
  }
  if (next.game?.turn?.step !== prev.game?.turn?.step) {
    parts.push(`step → ${next.game?.turn?.step}`);
  }
  return parts.length ? parts.join(", ") : "no visible change";
}
