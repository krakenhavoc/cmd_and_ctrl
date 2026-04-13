<script lang="ts">
  import { onMount } from "svelte";
  import { createGame, listGames, logout as apiLogout, startGame, type GameMeta } from "../lib/api";
  import { inviteURL, navigate } from "../lib/router";
  import { session, LobbyApiError } from "../lib/session";

  // Lobby is the admin + player landing page. Admins see a create-
  // game form and the invite token for each game they've created;
  // players see the game they're seated in and a button to jump
  // into the game view.
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
              {#if g.state === "lobby" && g.players.length >= 2}
                <button onclick={() => onStart(g.id)}>start</button>
              {/if}
              <button onclick={() => openGame(g.id)}>open</button>
            </div>
          </div>
          <ul class="seats">
            {#each g.players as p (p.player_id)}
              <li>seat {p.seat}: {p.name}</li>
            {/each}
          </ul>
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
