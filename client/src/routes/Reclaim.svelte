<script lang="ts">
  import { onMount } from "svelte";
  import { redeemSeatReclaim } from "../lib/api";
  import { navigate } from "../lib/router";
  import { LobbyApiError } from "../lib/session";
  import Icon from "../lib/components/Icon.svelte";
  import SiteHeader from "../lib/components/SiteHeader.svelte";

  // Reclaim is the landing page for an admin-minted seat-reclaim
  // link (#/games/:id/reclaim?t=<ticket>). The host generates one
  // from the lobby when a player has lost their session mid-game —
  // closed tab, cleared storage, new device — and the invite link
  // can no longer help them, because seats cannot be claimed once
  // the table is underway.
  //
  // There is nothing to fill in: the ticket names the seat, so the
  // page redeems on mount and sends the player straight to the
  // board. That is deliberate. A confirmation step here would buy
  // nothing (the holder already decided to click) and would cost
  // the person who is trying to get back into a game in progress.
  //
  // The ticket is single use, so this runs exactly once — a refresh
  // after a successful redeem lands on "already used", which is the
  // honest answer rather than a retry loop.
  interface Props {
    gameID: string;
    ticket: string;
  }
  const { gameID, ticket }: Props = $props();

  let error = $state("");
  let done = $state(false);

  onMount(() => {
    void redeem();
  });

  let ran = false;
  async function redeem(): Promise<void> {
    if (ran) return;
    ran = true;
    if (!ticket) {
      error = "This link is missing its ticket — ask your host to send a fresh one.";
      return;
    }
    try {
      await redeemSeatReclaim(gameID, ticket);
      done = true;
      navigate(`#/games/${gameID}`);
    } catch (err) {
      error =
        err instanceof LobbyApiError
          ? err.message
          : "could not return you to your seat — ask your host for a fresh link";
    }
  }
</script>

<SiteHeader />

<section class="entry">
  <div class="stack">
    <div class="head">
      <p class="eyebrow">Returning you to your seat</p>
      <h1>welcome back</h1>
    </div>

    <div class="card">
      {#if error}
        <p class="notice err" role="alert">
          <Icon name="x" size={14} />
          <span>{error}</span>
        </p>
        <p class="help">
          Reclaim links are single use and expire quickly, on purpose — they let whoever holds one
          play as you. Ask your host to generate another from the lobby.
        </p>
        <div class="frow">
          <a class="btn-link" href="#/login">Go to sign in</a>
        </div>
      {:else if done}
        <p class="help" aria-live="polite">You're in. Taking you to the table…</p>
      {:else}
        <p class="help" aria-live="polite">Checking your link…</p>
      {/if}
    </div>
  </div>
</section>

<style>
  /* Same shell as Join.svelte — this is the other page a player can
     land on without a session, and the two should not look like
     different products. */
  .entry {
    position: relative;
    min-height: calc(100vh - 3rem - 68px);
    display: flex;
    justify-content: center;
  }
  .stack {
    width: min(860px, 100%);
    display: flex;
    flex-direction: column;
    gap: 22px;
    margin-top: clamp(24px, 6vh, 60px);
    align-items: center;
  }
  .head {
    text-align: center;
  }
  .eyebrow {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  h1 {
    margin: 6px 0 0;
    font-family: var(--font-display);
    font-size: 28px;
    font-weight: 800;
    letter-spacing: -0.02em;
    color: var(--fg);
  }
  .card {
    width: min(600px, 100%);
    box-sizing: border-box;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
    padding: 20px 22px;
  }
  .notice {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    margin: 0;
    font-size: 13.5px;
    color: var(--fg);
  }
  .notice.err {
    color: var(--danger);
  }
  .help {
    font-size: 12.5px;
    color: var(--fg-muted);
    line-height: 1.5;
    margin: 10px 0 0;
  }
  .frow {
    display: flex;
    gap: 8px;
    margin-top: 14px;
  }
  .btn-link {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 34px;
    padding: 0 14px;
    border-radius: var(--radius);
    border: 1px solid var(--border-strong);
    background: var(--surface-raised);
    color: var(--fg);
    font-size: 12.5px;
    font-weight: 600;
    text-decoration: none;
    box-sizing: border-box;
  }
  .btn-link:hover {
    background: var(--surface-hover);
  }
</style>
