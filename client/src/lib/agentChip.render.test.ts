// @vitest-environment jsdom
//
// The AI-agent chip (ADR 0122 §7) on every surface the ADR names: the
// seat at the table, the lobby list, the invite preview and chat. It
// shows to everyone, has an accessible name, and is never the bot chip.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import BotFeed from "./components/BotFeed.svelte";
import PlayerIdentity from "./components/board/PlayerIdentity.svelte";
import Join from "../routes/Join.svelte";
import Lobby from "../routes/Lobby.svelte";
import type { PlayerView } from "./protocol";
import { setSession } from "./session";
import { cleanup, flushSync, render } from "./test/render.svelte";
import type { ChatMessage } from "./ws";

const GAME = "11111111-2222-3333-4444-555555555555";
const CHIP = '[aria-label^="AI agent"]';

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  setSession(null);
  localStorage.clear();
});

const seat = (over: Partial<PlayerView>): PlayerView =>
  ({
    id: "ag",
    name: "Claude",
    seat: 1,
    life: 40,
    library: { kind: "library", count: 0, cards: [] },
    hand: { kind: "hand", count: 0, cards: [] },
    graveyard: { kind: "graveyard", count: 0, cards: [] },
    command: { kind: "command", count: 0, cards: [] },
    commander_damage: {},
    life_history: [],
    ...over,
  }) as unknown as PlayerView;

const props = (s: PlayerView, over: Record<string, unknown> = {}) => ({
  seat: s,
  isSelf: false,
  isActive: false,
  hasPriority: false,
  attackTargetable: false,
  isMonarch: false,
  isInitiative: false,
  sendAction: () => {},
  ...over,
});

describe("on the seat at the table", () => {
  it("shows 'AI · client' with a name and the ADR's tooltip, for any viewer", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(seat({ is_agent: true, agent_client: "claude-code" })) as never,
    );
    const chip = container.querySelector(CHIP);
    expect(chip?.getAttribute("aria-label")).toBe("AI agent, claude-code");
    expect(chip?.textContent).toContain("AI · claude-code");
    expect(chip?.getAttribute("title")).toBe(
      "Played by an AI agent (claude-code). It sees only what this seat sees.",
    );
  });

  it("reads thinking… while the seat holds priority, and not once it is eliminated", () => {
    const s = seat({ is_agent: true, agent_client: "claude-code" });
    const { container, setProps } = render(
      PlayerIdentity as never,
      props(s, { hasPriority: true }) as never,
    );
    expect(container.querySelector(CHIP)?.textContent).toContain("AI · thinking…");
    expect(container.querySelector(CHIP)?.getAttribute("aria-label")).toBe(
      "AI agent, claude-code, thinking",
    );
    setProps({
      seat: seat({ is_agent: true, agent_client: "claude-code", eliminated: true }),
    } as never);
    expect(container.querySelector(CHIP)?.textContent).not.toContain("thinking");
  });

  it("is never the bot chip, and a seat never shows both", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(seat({ is_agent: true, agent_client: "x", is_bot: true, bot_tier: "random" })) as never,
    );
    expect(container.querySelectorAll(".tag.bot").length).toBe(1);
    expect(container.querySelector(CHIP)).toBeNull();

    const agent = render(PlayerIdentity as never, props(seat({ is_agent: true })) as never);
    expect(agent.container.querySelector(".tag.bot")).toBeNull();
    expect(agent.container.querySelector(".bot-mark")).toBeNull();
    expect(agent.container.querySelector(CHIP)?.getAttribute("aria-label")).toBe("AI agent");
  });

  it("is absent on a human seat", () => {
    const { container } = render(PlayerIdentity as never, props(seat({})) as never);
    expect(container.querySelector(CHIP)).toBeNull();
  });
});

function stubServer(routes: Record<string, unknown>): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const payload = routes[`${init.method ?? "GET"} ${input}`] ?? {};
      return {
        ok: true,
        status: 200,
        statusText: "OK",
        json: async () => payload,
        clone() {
          return this;
        },
      } as unknown as Response;
    }),
  );
}

async function settle(): Promise<void> {
  for (let i = 0; i < 8; i++) {
    for (let j = 0; j < 5; j++) await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

const players = [
  {
    player_id: "p0",
    name: "Robo",
    seat: 0,
    deck_uploaded: true,
    is_bot: true,
    bot_tier: "random",
  },
  {
    player_id: "p1",
    name: "Claude",
    seat: 1,
    deck_uploaded: false,
    is_agent: true,
    agent_client: "claude-code",
  },
  { player_id: "p2", name: "Bob", seat: 2, deck_uploaded: false },
];
const meta = {
  id: GAME,
  name: "Friday",
  created_at: new Date().toISOString(),
  players,
  state: "lobby",
  is_creator: false,
};

describe("in the lobby list", () => {
  beforeEach(() => {
    location.hash = "#/lobby";
  });
  it("labels the agent seat, apart from the bot seat", async () => {
    setSession({
      token: "t",
      expiresAt: new Date(Date.now() + 86_400_000).toISOString(),
      principal: {
        role: "identified",
        user_id: "6f9619ff-8b86-d011-b42d-00c04fc964ff",
        name: "Alice",
        issued_at: new Date().toISOString(),
        expires_at: new Date(Date.now() + 86_400_000).toISOString(),
      },
    } as never);
    stubServer({
      "GET /games": { games: [{ ...meta, is_creator: true }] },
      "GET /me": { admin: false },
      "GET /me/games": { games: [] },
    });
    const { container } = render(Lobby as never, {} as never);
    await settle();
    const chips = container.querySelectorAll(CHIP);
    expect(chips.length).toBe(1);
    expect(chips[0].textContent).toContain("AI · claude-code");
    expect(container.querySelectorAll(".botchip").length).toBe(1);
    expect(chips[0].closest("li")?.querySelector(".botchip")).toBeNull();
  });
});

describe("in the invite preview", () => {
  it("labels the agent seat", async () => {
    stubServer({
      [`GET /games/${GAME}/preview?t=inv`]: { game: meta, invite: "player", max_seats: 4 },
      "GET /auth/discord/config": { enabled: false },
    });
    const { container } = render(
      Join as never,
      { gameID: GAME, inviteToken: "inv", spectator: false } as never,
    );
    await settle();
    const chips = container.querySelectorAll(CHIP);
    expect(chips.length).toBe(1);
    expect(chips[0].getAttribute("aria-label")).toBe("AI agent, claude-code");
  });
});

describe("in chat", () => {
  const line = (id: string, authorID: string): ChatMessage => ({
    id,
    authorID,
    authorName: authorID === "ag" ? "Claude" : "Bob",
    text: `hello from ${authorID}`,
    at: new Date(0),
    kind: "say",
    reason: "",
  });
  it("puts the chip on an agent's line and nobody else's", () => {
    const { container } = render(
      BotFeed as never,
      {
        chat: [line("1", "ag"), line("2", "bob")],
        seats: [{ id: "ag", is_agent: true, agent_client: "claude-code" }, { id: "bob" }],
      } as never,
    );
    expect(container.textContent).toContain("hello from ag");
    expect(container.textContent).not.toContain("hello from bob");
    expect(container.querySelectorAll(CHIP).length).toBe(1);
  });
});
