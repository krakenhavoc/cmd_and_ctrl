<script lang="ts">
  import { onMount } from "svelte";
  import { discordAuthEnabled, joinGame, spectateGame } from "../lib/api";
  import { navigate } from "../lib/router";
  import { LobbyApiError } from "../lib/session";

  // Props carried from the route parser (lib/router.ts). The
  // `spectator` flag flips the page from a player join (claims a
  // seat, posts /join) to a spectator join (read-only watch, posts
  // /spectate). Both flows share this route so the invite-URL
  // distinction is purely the `?spectator=1` query param.
  interface Props {
    gameID: string;
    inviteToken: string;
    spectator: boolean;
  }
  const { gameID, inviteToken, spectator }: Props = $props();

  let name = $state("");
  let busy = $state(false);
  let error = $state("");

  // discordEnabled gates the "Sign in with Discord" button. Probed
  // once on mount from /auth/discord/config — a deploy without the
  // three CMDCTRL_DISCORD_* env vars hides the button so the user
  // isn't offered a flow that would 503. Spectator joins stay
  // manual-name-only for now; the Discord flow claims a seat and
  // spectators don't need one.
  let discordEnabled = $state(false);
  onMount(() => {
    if (spectator) return; // not surfaced for spectator flow
    void discordAuthEnabled().then((on) => {
      discordEnabled = on;
    });
  });

  // discordHref is a plain link into /auth/discord/start rather
  // than a fetch — the server issues a 302 to Discord, which the
  // browser must follow at the top level (not inside an XHR) so
  // the user actually lands on the Discord consent screen.
  const discordHref = $derived(
    `/auth/discord/start?game=${encodeURIComponent(gameID)}&t=${encodeURIComponent(inviteToken)}`,
  );

  async function submit(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    if (!name.trim()) return;
    busy = true;
    error = "";
    try {
      if (spectator) {
        await spectateGame(gameID, inviteToken, name.trim());
        // Spectators bypass the lobby (no deck to import, no seat
        // to manage) and land directly on the game route.
        navigate(`#/games/${gameID}`);
      } else {
        await joinGame(gameID, inviteToken, name.trim());
        // Player flow: lobby first so they can import a deck and
        // see other seats' status before entering the game route.
        navigate("#/lobby");
      }
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "join failed";
    } finally {
      busy = false;
    }
  }
</script>

<section>
  <h1>{spectator ? "spectate game" : "join game"}</h1>
  <p class="muted">
    game: <code>{gameID}</code>
    {#if spectator}
      <span class="badge">read-only</span>
    {/if}
  </p>

  {#if !inviteToken}
    <p class="error">invite token missing from URL</p>
  {:else}
    {#if discordEnabled}
      <a class="discord-btn" href={discordHref}>
        <span class="discord-mark" aria-hidden="true">◆</span> Sign in with Discord
      </a>
      <p class="muted or-line">or enter a name manually</p>
    {/if}
    <form onsubmit={submit}>
      <input
        type="text"
        placeholder={spectator ? "your name (chat label)" : "your name"}
        bind:value={name}
        required
      />
      <button type="submit" disabled={busy || !name.trim()}>
        {busy ? "…" : spectator ? "watch" : "join"}
      </button>
    </form>
    {#if spectator}
      <p class="muted note">
        You'll see the table from a non-seated viewpoint. Opponent hands and libraries stay hidden,
        the same way they do for any other player. You can't send actions.
      </p>
    {/if}
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
  .note {
    margin-top: 0.75rem;
  }
  .discord-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    background: #5865f2;
    color: #fff;
    text-decoration: none;
    padding: 0.6rem 1rem;
    border-radius: 4px;
    font-weight: 600;
    margin-bottom: 0.5rem;
  }
  .discord-btn:hover {
    background: #4651c8;
  }
  .discord-mark {
    font-size: 1.1em;
    line-height: 1;
  }
  .or-line {
    margin: 0.25rem 0 0.75rem 0;
    font-size: 0.85em;
  }
  code {
    font-family: monospace;
  }
  .badge {
    display: inline-block;
    margin-left: 0.5rem;
    padding: 1px 6px;
    border: 1px solid #4a5270;
    border-radius: 3px;
    font-size: 0.75em;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: #b08aff;
  }
  .error {
    color: #c00;
  }
</style>
