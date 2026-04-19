<script lang="ts">
  import { onMount } from "svelte";
  import { createGame, listGames, logout as apiLogout, startGame, type GameMeta } from "../lib/api";
  import { inviteURL, navigate } from "../lib/router";
  import { session, LobbyApiError } from "../lib/session";
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

  // Track the freshly-created game's invite token client-side — the
  // List endpoint strips invite tokens, so we remember them per
  // session so admins can copy the link without re-fetching /games/{id}.
  const recentInvites = new Map<string, string>();

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
              {#if canStart(g)}
                <button onclick={() => onStart(g.id)}>start</button>
              {:else if g.state === "lobby" && g.players.length >= 2}
                <button disabled title="waiting for all seats to upload a deck">start</button>
              {/if}
              <button onclick={() => openGame(g.id)}>open</button>
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
    max-width: 720px;
    margin: 2rem auto;
    padding: 1.5rem;
  }
  header {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
  }
  .games {
    list-style: none;
    padding: 0;
  }
  .games > li {
    border: 1px solid #ddd;
    padding: 0.75rem;
    border-radius: 4px;
    margin-bottom: 0.5rem;
  }
  .row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .row-actions {
    display: flex;
    gap: 0.5rem;
  }
  .seats {
    margin: 0.5rem 0 0 1rem;
    color: #555;
    font-size: 0.9em;
  }
  .muted {
    color: #666;
  }
  .error {
    color: #c00;
  }
  .badge-ok {
    color: #060;
    margin-left: 0.5rem;
  }
  .badge-pending {
    color: #a60;
    margin-left: 0.5rem;
  }
  .deck-upload {
    margin-top: 0.75rem;
    padding: 0.5rem;
    background: #f7f7f7;
    border-radius: 3px;
  }
  .deck-upload summary {
    cursor: pointer;
    font-weight: 600;
  }
  form {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }
  input {
    flex: 1;
    padding: 0.5rem;
  }
  .linkish {
    background: none;
    border: none;
    color: #06c;
    text-decoration: underline;
    cursor: pointer;
    padding: 0;
    margin-left: 0.5rem;
  }
</style>
