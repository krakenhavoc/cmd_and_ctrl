<script lang="ts">
  import { onMount } from "svelte";
  import { startPractice } from "../lib/practiceTable";
  import { navigate } from "../lib/router";
  import { session } from "../lib/session";
  import { lobbyErrorMessage } from "../lib/serverFull";
  import Icon from "../lib/components/Icon.svelte";
  import Page from "../lib/components/Page.svelte";
  import { markHintSeen } from "../lib/hints/runtime";
  import practiceOffer from "./LobbyPractice.hint";

  // Practice is the door into the tutorial's practice table (ADR 0076
  // §2.2): #/practice opens a fresh practice game for this session and
  // moves straight on to it. There is nothing to choose — the decks
  // and the bot are fixed — so, like Reclaim, it acts on mount.
  //
  // Everything that has to be put back afterwards (four settings, the
  // session) is recorded and swapped by lib/practiceTable.ts, which
  // also owns every way back out. This page only shows progress and a
  // failure.
  //
  // It is public (ADR 0156 §2): a visitor with no session gets a guest
  // practice game, which startPractice opens through POST
  // /games/practice/guest. The front page, the gate page and Help link
  // here for everyone.
  let error = $state("");

  let ran = false;
  onMount(() => {
    if (ran) return;
    ran = true;
    void open();
  });

  async function open(): Promise<void> {
    // Opening a practice table from anywhere answers the tutorial offer
    // (ADR 0125 §5.3), so the Lobby does not make it again.
    markHintSeen(practiceOffer);
    try {
      const id = await startPractice();
      navigate(`#/games/${id}`);
    } catch (err) {
      error = lobbyErrorMessage(err, "could not open a practice game — try again");
    }
  }
</script>

<Page eyebrow="Practice game" title="Setting the table" width="prose">
  <div class="stack">
    <div class="card">
      {#if error}
        <p class="notice err" role="alert">
          <Icon name="x" size={14} />
          <span>{error}</span>
        </p>
        <div class="frow">
          {#if $session}
            <a class="btn-link" href="#/lobby">Back to the lobby</a>
          {:else}
            <a class="btn-link" href="#/">Back to the front page</a>
          {/if}
        </div>
      {:else}
        <p class="help" aria-live="polite">Shuffling two decks and seating a practice bot…</p>
      {/if}
    </div>
  </div>
</Page>

<style>
  /* Reclaim.svelte's shell: another page that does one thing on
     mount and moves on. */
  .stack {
    display: flex;
    flex-direction: column;
    gap: 22px;
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
