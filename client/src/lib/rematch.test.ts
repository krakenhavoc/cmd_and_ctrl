import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { acceptRematch, leaveRematch, offerRematch } from "./api";
import { gameOverRequest, type GameOverRematch } from "./choiceDock";
import type { GameView, PlayerView, RematchView } from "./protocol";
import { rematchHostSeat, rematchPlan } from "./rematch";
import { currentSession, setSession } from "./session";

// rematch.test.ts: #3285 (ADR 0164). What a finished table offers each
// viewer, the dock request that carries it, and the three calls.

const seat = (id: string, n: number, over: Partial<PlayerView> = {}): PlayerView =>
  ({ id, name: id, seat: n, ...over }) as PlayerView;

const view = (seats: PlayerView[], over: Partial<GameView> = {}): GameView =>
  ({ id: "g", state: "ended", seats, ...over }) as GameView;

const offer = (over: Partial<RematchView> = {}): RematchView => ({
  game_id: "g2",
  host: "a",
  host_name: "Ann",
  open: true,
  ...over,
});

describe("rematchHostSeat", () => {
  it("is the table's host", () => {
    const v = view([seat("a", 0), seat("b", 1, { is_host: true })]);
    expect(rematchHostSeat(v)?.id).toBe("b");
  });

  it("is the first human seat when nobody hosts (the bots won)", () => {
    const v = view([seat("bot", 0, { is_bot: true }), seat("c", 2), seat("b", 1)]);
    expect(rematchHostSeat(v)?.id).toBe("b");
  });

  it("never names a bot or an agent, even one flagged host", () => {
    const v = view([
      seat("bot", 0, { is_bot: true, is_host: true }),
      seat("x", 1, { is_agent: true }),
    ]);
    expect(rematchHostSeat(v)).toBeNull();
  });
});

describe("rematchPlan", () => {
  const seats = [seat("a", 0, { is_host: true }), seat("b", 1), seat("bot", 2, { is_bot: true })];

  it("offers the host a rematch once the game is over", () => {
    expect(rematchPlan(view(seats), "a")).toEqual({ kind: "offer" });
  });

  it("offers nothing before the end, to a spectator, a bot, a non-host, or on the practice table", () => {
    expect(rematchPlan(view(seats, { state: "active" }), "a").kind).toBe("none");
    expect(rematchPlan(view(seats), null).kind).toBe("none");
    expect(rematchPlan(view(seats), "bot").kind).toBe("none");
    expect(rematchPlan(view(seats), "b").kind).toBe("none");
    expect(rematchPlan(view(seats), "stranger").kind).toBe("none");
    expect(rematchPlan(view(seats), "a", true).kind).toBe("none");
    expect(rematchPlan(null, "a").kind).toBe("none");
  });

  it("asks an undecided human while the offer is open", () => {
    const v = view(seats, { rematch: offer({ you: "undecided" }) });
    expect(rematchPlan(v, "b")).toEqual({ kind: "ask", hostName: "Ann", gameID: "g2" });
  });

  it("stops asking once the seat answered or the offer closed", () => {
    expect(rematchPlan(view(seats, { rematch: offer({ you: "declined" }) }), "b").kind).toBe(
      "none",
    );
    expect(rematchPlan(view(seats, { rematch: offer({ you: "accepted" }) }), "b").kind).toBe(
      "none",
    );
    expect(
      rematchPlan(view(seats, { rematch: offer({ you: "undecided", open: false }) }), "b").kind,
    ).toBe("none");
  });

  it("gives the host a way back to the new table, not a second offer", () => {
    const v = view(seats, { rematch: offer({ you: "host" }) });
    expect(rematchPlan(v, "a")).toEqual({ kind: "open-table", gameID: "g2" });
  });
});

describe("the game-over dock request with a rematch", () => {
  const actions = () => ({
    onOffer: vi.fn(),
    onOpen: vi.fn(),
    onAccept: vi.fn(),
    onLeave: vi.fn(),
  });
  const withPlan = (plan: GameOverRematch["plan"], a = actions(), busy = false) => ({
    a,
    req: gameOverRequest(vi.fn(), true, { plan, busy, ...a }),
  });

  it("is unchanged without a rematch, and for a plan of none", () => {
    expect(gameOverRequest(vi.fn()).secondary).toEqual([]);
    expect(withPlan({ kind: "none" }).req.secondary).toEqual([]);
    expect(withPlan({ kind: "none" }).req.row).toBeUndefined();
  });

  it("puts Rematch beside Back to lobby for the host", () => {
    const { a, req } = withPlan({ kind: "offer" });
    expect(req.primary?.label).toBe("Back to lobby");
    expect(req.secondary?.map((s) => s.label)).toEqual(["Rematch"]);
    req.secondary![0].onPress();
    expect(a.onOffer).toHaveBeenCalledOnce();
  });

  it("asks the others with Join and No thanks", () => {
    const { a, req } = withPlan({ kind: "ask", hostName: "Ann", gameID: "g2" });
    expect(req.question).toBe("Ann offered a rematch");
    expect(req.row?.map((r) => r.label)).toEqual(["Join the rematch", "No thanks"]);
    req.row![0].onPress();
    req.row![1].onPress();
    expect(a.onAccept).toHaveBeenCalledOnce();
    expect(a.onLeave).toHaveBeenCalledOnce();
  });

  it("opens the new table for a host who already offered", () => {
    const { a, req } = withPlan({ kind: "open-table", gameID: "g2" });
    req.secondary![0].onPress();
    expect(a.onOpen).toHaveBeenCalledWith("g2");
  });

  it("refuses a press while one is in flight, and before the grace", () => {
    const busy = withPlan({ kind: "offer" }, actions(), true);
    expect(busy.req.secondary![0].disabled).toBe(true);
    busy.req.secondary![0].onPress();
    expect(busy.a.onOffer).not.toHaveBeenCalled();

    const a = actions();
    const early = gameOverRequest(vi.fn(), false, { plan: { kind: "offer" }, ...a });
    expect(early.secondary![0].disabled).toBe(true);
    early.secondary![0].onPress();
    expect(a.onOffer).not.toHaveBeenCalled();
  });
});

describe("the rematch calls", () => {
  let calls: { url: string; method: string }[] = [];
  const stub = (status: number, body: unknown) => {
    calls = [];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string, init: RequestInit = {}) => {
        calls.push({ url, method: init.method ?? "GET" });
        return new Response(status === 204 ? null : JSON.stringify(body), {
          status,
          headers: { "Content-Type": "application/json" },
        });
      }),
    );
  };
  beforeEach(() => {
    setSession({
      token: "old",
      expiresAt: new Date(Date.now() + 3_600_000).toISOString(),
      principal: {
        role: "player",
        game_id: "g",
        player_id: "a",
        issued_at: "2026-10-13T00:00:00Z",
        expires_at: "2026-10-14T00:00:00Z",
      },
      gameID: "g",
      playerID: "a",
    });
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    setSession(null);
  });

  const newSeat = {
    token: "new",
    expires_at: "2026-10-14T00:00:00Z",
    principal: {
      role: "player",
      game_id: "g2",
      player_id: "a2",
      issued_at: "2026-10-13T00:00:00Z",
      expires_at: "2026-10-14T00:00:00Z",
    },
    game: { id: "g2" },
    player_id: "a2",
  };

  it("offering posts to the old table and installs the seat at the new one", async () => {
    stub(201, newSeat);
    const s = await offerRematch("g");
    expect(calls).toEqual([{ url: "/games/g/rematch", method: "POST" }]);
    expect(s.gameID).toBe("g2");
    expect(currentSession()).toMatchObject({ token: "new", gameID: "g2", playerID: "a2" });
  });

  it("accepting does the same on /accept", async () => {
    stub(200, newSeat);
    await acceptRematch("g");
    expect(calls).toEqual([{ url: "/games/g/rematch/accept", method: "POST" }]);
    expect(currentSession()?.gameID).toBe("g2");
  });

  it("leaving keeps the session where it is", async () => {
    stub(204, null);
    await leaveRematch("g");
    expect(calls).toEqual([{ url: "/games/g/rematch/leave", method: "POST" }]);
    expect(currentSession()?.token).toBe("old");
  });

  it("a refusal throws the server's sentence and keeps the session", async () => {
    stub(409, { error: "the rematch offer is closed" });
    await expect(acceptRematch("g")).rejects.toThrow("the rematch offer is closed");
    expect(currentSession()?.token).toBe("old");
  });
});
