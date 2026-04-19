<script lang="ts">
  import { joinGame } from "../lib/api";
  import { navigate } from "../lib/router";
  import { LobbyApiError } from "../lib/session";

  // Props carried from the route parser (lib/router.ts).
  interface Props {
    gameID: string;
    inviteToken: string;
  }
  const { gameID, inviteToken }: Props = $props();

  let name = $state("");
  let busy = $state(false);
  let error = $state("");

  async function submit(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    if (!name.trim()) return;
    busy = true;
    error = "";
    try {
      await joinGame(gameID, inviteToken, name.trim());
      // Land in the lobby first so the player can import a deck +
      // see other seats' status before entering the game route. The
      // in-game deck-import modal in Game.svelte (S08.5 wave 1)
      // remains as a safety net for direct navigation to /games/{id}
      // by an unimported seat, but the default flow is lobby first.
      navigate("#/lobby");
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "join failed";
    } finally {
      busy = false;
    }
  }
</script>

<section>
  <h1>join game</h1>
  <p class="muted">game: <code>{gameID}</code></p>

  {#if !inviteToken}
    <p class="error">invite token missing from URL</p>
  {:else}
    <form onsubmit={submit}>
      <input type="text" placeholder="your name" bind:value={name} required />
      <button type="submit" disabled={busy || !name.trim()}>
        {busy ? "…" : "join"}
      </button>
    </form>
  {/if}

  {#if error}
    <p class="error">{error}</p>
  {/if}
</section>

<style>
  section {
    max-width: 480px;
    margin: 2rem auto;
    padding: 1.5rem;
  }
  form {
    display: flex;
    gap: 0.5rem;
  }
  input {
    flex: 1;
    padding: 0.5rem;
  }
  .muted {
    color: #666;
    font-size: 0.9em;
  }
  code {
    font-family: monospace;
  }
  .error {
    color: #c00;
  }
</style>
