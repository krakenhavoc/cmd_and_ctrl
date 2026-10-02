<script lang="ts">
  import { onMount } from "svelte";
  import { startPractice } from "../lib/practiceTable";
  import { navigate } from "../lib/router";
  import { LobbyApiError } from "../lib/session";
  import Icon from "../lib/components/Icon.svelte";
  import SiteHeader from "../lib/components/SiteHeader.svelte";

  // Practice is the door into the tutorial's practice table (ADR 0076
  // §2.2): #/practice opens a fresh practice game for this session and
  // moves straight on to it. There is nothing to choose — the decks
  // and the bot are fixed — so, like Reclaim, it acts on mount.
  //
  // Everything that has to be put back afterwards (four settings, the
  // session) is recorded and swapped by lib/practiceTable.ts, which
  // also owns every way back out. This page only shows progress and a
  // failure.
  let error = $state("");

  let ran = false;
  onMount(() => {
    if (ran) return;
    ran = true;
    void open();
  });

  async function open(): Promise<void> {
    try {
      const id = await startPractice();
      navigate(`#/games/${id}`);
    } catch (err) {
      error =
        err instanceof LobbyApiError ? err.message : "could not open a practice game — try again";
    }
  }
</script>

<SiteHeader />

<section class="entry">
  <div class="stack">
    <div class="head">
      <p class="eyebrow">Practice game</p>
      <h1>setting the table</h1>
    </div>

    <div class="card">
      {#if error}
        <p class="notice err" role="alert">
          <Icon name="x" size={14} />
          <span>{error}</span>
        </p>
        <div class="frow">
          <a class="btn-link" href="#/lobby">Back to the lobby</a>
        </div>
      {:else}
        <p class="help" aria-live="polite">Shuffling two decks and seating a practice bot…</p>
      {/if}
    </div>
  </div>
</section>

<style>
  /* Reclaim.svelte's shell: another page that does one thing on
     mount and moves on. */
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
    margin: 0;
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
