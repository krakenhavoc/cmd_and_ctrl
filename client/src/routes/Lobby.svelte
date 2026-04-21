<script lang="ts">
  import { onMount } from "svelte";
  import {
    createGame,
    listGames,
    logout as apiLogout,
    replayURL,
    startGame,
    type GameMeta,
  } from "../lib/api";
  import { inviteURL, spectatorInviteURL, navigate } from "../lib/router";
  import { session, LobbyApiError } from "../lib/session";
  import { openSettings } from "../lib/settings";
  import DeckUploadForm from "../lib/components/DeckUploadForm.svelte";

  // Lobby is the admin + player landing page. Admins see a create-
  // game form and the invite token for each game they've created;
  // players see the game they're seated in with a deck-upload panel
  // for their own seat and a button to jump into the game view once
  // all seats are ready.
  let games = $state<GameMeta[]>([]);
  let error = $state("");
  let newName = $state("");
  let busy = $state(false);

  // Track the freshly-created game's invite tokens client-side — the
  // List endpoint strips both invites, so we remember them per
  // session so admins can copy each link without re-fetching
  // /games/{id}. The spectator invite (S11) is distinct from the
  // player invite — sharing the player one with a spectator would
  // let them claim a seat.
  const recentInvites = new Map<string, string>();
  const recentSpectatorInvites = new Map<string, string>();

  async function refresh(): Promise<void> {
    try {
      games = await listGames();
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "list failed";
    }
  }

  async function onCreate(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    if (!newName.trim()) return;
    busy = true;
    error = "";
    try {
      const meta = await createGame(newName.trim());
      if (meta.invite_token) recentInvites.set(meta.id, meta.invite_token);
      if (meta.spectator_invite) recentSpectatorInvites.set(meta.id, meta.spectator_invite);
      newName = "";
      await refresh();
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "create failed";
    } finally {
      busy = false;
    }
  }

  async function onStart(id: string): Promise<void> {
    error = "";
    try {
      await startGame(id);
      await refresh();
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "start failed";
    }
  }

  function copyInvite(id: string): void {
    const token = recentInvites.get(id);
    if (!token) {
      error = "no invite token cached for this game — re-open as admin to recover";
      return;
    }
    const url = inviteURL(id, token);
    void navigator.clipboard.writeText(url);
  }

  function copySpectatorInvite(id: string): void {
    const token = recentSpectatorInvites.get(id);
    if (!token) {
      error = "no spectator invite cached for this game — re-open as admin to recover";
      return;
    }
    const url = spectatorInviteURL(id, token);
    void navigator.clipboard.writeText(url);
  }

  function openGame(id: string): void {
    navigate(`#/games/${id}`);
  }

  async function logout(): Promise<void> {
    await apiLogout();
    navigate("#/login");
  }

  // mySeat returns the seat this user occupies in game g, or null if
  // the session isn't bound to a seat in g (admin viewing someone
  // else's game, or a RolePlayer viewing a different game entirely).
  function mySeat(g: GameMeta) {
    const s = $session;
    if (!s?.playerID || s.gameID !== g.id) return null;
    return g.players.find((p) => p.player_id === s.playerID) ?? null;
  }

  // canStart returns true when Start would succeed: at least 2 seats
  // and every seat has uploaded a real deck.
  function canStart(g: GameMeta): boolean {
    return g.state === "lobby" && g.players.length >= 2 && g.players.every((p) => p.deck_uploaded);
  }

  onMount(() => {
    void refresh();
  });
</script>

<section>
  <header>
    <h1>cmd_and_ctrl · lobby</h1>
    <p>
      logged in as <strong>{$session?.principal.role}</strong>
      {#if $session?.principal.name}
        ({$session.principal.name})
      {/if}
      <button
        class="linkish gear"
        title="settings (press , from anywhere)"
        aria-label="open settings"
        onclick={openSettings}>⚙</button
      >
      <button class="linkish" onclick={logout}>log out</button>
    </p>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if $session?.principal.role === "admin"}
    <h2>create game</h2>
    <form onsubmit={onCreate}>
      <input type="text" placeholder="game name" bind:value={newName} />
      <button type="submit" disabled={busy || !newName.trim()}>create</button>
    </form>
  {/if}

  <h2>games</h2>
  {#if games.length === 0}
    <p class="muted">no games yet.</p>
  {:else}
    <ul class="games">
      {#each games as g (g.id)}
        {@const seat = mySeat(g)}
        <li>
          <div class="row">
            <div>
              <strong>{g.name}</strong>
              <span class="muted">· {g.state} · {g.players.length} player(s)</span>
            </div>
            <div class="row-actions">
              {#if recentInvites.has(g.id)}
                <button onclick={() => copyInvite(g.id)}>copy invite</button>
              {/if}
              {#if recentSpectatorInvites.has(g.id)}
                <button onclick={() => copySpectatorInvite(g.id)}>copy spectator link</button>
              {/if}
              {#if canStart(g)}
                <button onclick={() => onStart(g.id)}>start</button>
              {:else if g.state === "lobby" && g.players.length >= 2}
                <button disabled title="waiting for all seats to upload a deck">start</button>
              {/if}
              <button onclick={() => openGame(g.id)}>open</button>
              {#if g.state !== "lobby"}
                {@const url = replayURL(g.id)}
                {#if url}
                  <a class="linkish" href={url} download={`${g.id}.jsonl`}>download replay</a>
                {/if}
              {/if}
            </div>
          </div>
          <ul class="seats">
            {#each g.players as p (p.player_id)}
              <li>
                seat {p.seat}: {p.name}
                {#if p.deck_uploaded}
                  <span class="badge-ok">✓ {p.deck_name || "deck ready"}</span>
                {:else}
                  <span class="badge-pending">deck pending</span>
                {/if}
              </li>
            {/each}
          </ul>

          {#if seat && g.state === "lobby" && $session?.playerID}
            <details class="deck-upload" open={!seat.deck_uploaded}>
              <summary>
                {seat.deck_uploaded ? "replace your deck" : "upload your deck"}
              </summary>
              <p class="muted">
                Paste a Moxfield or Archidekt deck URL, a Moxfield JSON export, or a plain-text
                decklist. The server validates against Commander rules (100-card singleton, color
                identity, format legality).
              </p>
              <DeckUploadForm
                gameID={g.id}
                playerID={$session.playerID}
                onSuccess={() => void refresh()}
              />
            </details>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}

  <p>
    <button onclick={refresh}>refresh</button>
  </p>
</section>

<style>
  section {
    max-width: 760px;
    margin: 2rem auto;
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    padding: 0.25rem 0 0.75rem;
    border-bottom: 1px solid var(--border);
  }
  header h1 {
    margin: 0;
    font-size: 1.5rem;
    letter-spacing: -0.02em;
  }
  header p {
    margin: 0;
    color: var(--fg-muted);
    font-size: 0.9rem;
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
  }
  header strong {
    color: var(--fg);
    font-weight: 600;
  }
  .games {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
  .games > li {
    background: linear-gradient(180deg, var(--surface) 0%, var(--bg-2) 100%);
    border: 1px solid var(--border);
    padding: 1rem 1.1rem;
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow);
    transition:
      border-color 140ms var(--ease),
      transform 140ms var(--ease);
  }
  .games > li:hover {
    border-color: var(--border-strong);
  }
  .row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }
  .row > div:first-child strong {
    font-size: 1.05rem;
  }
  .row-actions {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }
  .row-actions button {
    padding: 0.4rem 0.8rem;
    font-size: 0.85rem;
  }
  .seats {
    list-style: none;
    margin: 0.75rem 0 0;
    padding: 0;
    color: var(--fg-muted);
    font-size: 0.9em;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  .seats li {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.25rem 0.5rem;
    border-radius: var(--radius-sm);
    background: var(--surface-sunken);
  }
  .muted {
    color: var(--fg-muted);
    font-weight: 500;
  }
  .error {
    color: var(--danger);
    padding: 0.6rem 0.85rem;
    border-radius: var(--radius);
    background: rgba(255, 122, 122, 0.08);
    border: 1px solid rgba(255, 122, 122, 0.3);
    margin: 0;
  }
  .badge-ok {
    display: inline-block;
    padding: 1px 8px;
    border-radius: 999px;
    color: var(--mint);
    background: rgba(122, 255, 154, 0.12);
    border: 1px solid rgba(122, 255, 154, 0.3);
    margin-left: 0.25rem;
    font-weight: 600;
    font-size: 0.8em;
  }
  .badge-pending {
    display: inline-block;
    padding: 1px 8px;
    border-radius: 999px;
    color: var(--gold);
    background: var(--gold-soft);
    border: 1px solid rgba(255, 208, 122, 0.3);
    margin-left: 0.25rem;
    font-weight: 600;
    font-size: 0.8em;
  }
  .deck-upload {
    margin-top: 0.75rem;
    padding: 0.6rem 0.75rem;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    border-radius: var(--radius);
  }
  .deck-upload summary {
    cursor: pointer;
    font-weight: 600;
    color: var(--accent-strong);
    padding: 0.25rem 0;
  }
  .deck-upload summary:hover {
    color: var(--fg);
  }
  form {
    display: flex;
    gap: 0.5rem;
  }
  input {
    flex: 1;
  }
  .linkish {
    background: none;
    border: none;
    color: var(--accent-strong);
    text-decoration: none;
    cursor: pointer;
    padding: 0.3rem 0.55rem;
    margin-left: 0.25rem;
    border-radius: var(--radius-sm);
    font-weight: 500;
    box-shadow: none;
    font-size: 0.85rem;
    transition: background 120ms var(--ease);
  }
  .linkish:hover {
    background: var(--accent-soft);
    color: var(--accent-strong);
    box-shadow: none;
  }
  .linkish.gear {
    font-size: 1.1rem;
    padding: 0.3rem 0.5rem;
    line-height: 1;
  }
</style>
