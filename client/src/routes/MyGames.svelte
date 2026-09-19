<script lang="ts">
  import { onMount } from "svelte";
  import { fetchMyGames, rejoinMyGame } from "../lib/api";
  import {
    myGameStatus,
    othersLabel,
    playedWhen,
    signedInUserID,
    sortMyGames,
    type MyGame,
  } from "../lib/myGames";
  import { navigate } from "../lib/router";
  import { LobbyApiError, session } from "../lib/session";
  import Icon from "../lib/components/Icon.svelte";

  // "My games" (ADR 0051 decision 4, S34 sub-PR 4): every table the
  // signed-in person has sat at, from any device, without an invite
  // link. An open table has an Open button that trades the current
  // session for a seat session (POST /me/games/{id}/session) — the
  // user-proved reclaim that replaces the admin ticket for anyone who
  // signed in with Discord. Finished and archived tables are history.
  //
  // Deliberately modest: a list, a status chip, who else was there. No
  // stats, no filters; eight people's games fit on one screen.

  const userID = $derived(signedInUserID($session));

  let games = $state<MyGame[]>([]);
  let loading = $state(false);
  let error = $state("");
  let opening = $state("");

  onMount(() => {
    void load();
  });

  async function load(): Promise<void> {
    if (!signedInUserID($session)) return;
    loading = true;
    error = "";
    try {
      games = sortMyGames(await fetchMyGames());
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "couldn't load your games";
    } finally {
      loading = false;
    }
  }

  async function open(g: MyGame): Promise<void> {
    if (!g.rejoin || opening) return;
    opening = g.id;
    error = "";
    try {
      await rejoinMyGame(g.rejoin);
      // A table still in the lobby opens on the lobby page, where the
      // deck is imported; one underway opens on the board.
      navigate(g.state === "lobby" ? "#/lobby" : `#/games/${g.id}`);
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "couldn't open that table";
    } finally {
      opening = "";
    }
  }

  // Where "back" goes: the lobby for a seated session, the sign-in page
  // (with its invite box) for an identity-only one.
  const backHref = $derived($session?.principal.role === "player" ? "#/lobby" : "#/login");
</script>

<section class="entry">
  <header class="topbar">
    <a class="wordmark" href={backHref} aria-label="cmd_and_ctrl home"><i></i>CMD &amp; CTRL</a>
  </header>
  <div class="stack">
    <div class="head">
      <p class="eyebrow">Every table you've sat at</p>
      <h1>my games</h1>
    </div>

    <div class="card">
      {#if !userID}
        <p class="help first">
          Your games are kept for your Discord account. Sign in with Discord to see them — a seat
          joined with a typed name isn't linked to anyone.
        </p>
        <div class="frow">
          <a class="btn-link" href="#/login">Go to sign in</a>
        </div>
      {:else if loading && games.length === 0}
        <p class="help first" aria-live="polite">Loading your games…</p>
      {:else if games.length === 0 && !error}
        <p class="help first">
          No games yet. Once you join a table while signed in, it shows up here — and so do tables
          you sat at with Discord before accounts existed.
        </p>
      {:else}
        <ul class="games" aria-label="your games">
          {#each games as g (g.id)}
            {@const status = myGameStatus(g)}
            <li class="game">
              <div class="info">
                <div class="line">
                  <span class="name">{g.name}</span>
                  <span class="chip {status.tone}">{status.label}</span>
                </div>
                <div class="meta">
                  {othersLabel(g)} · seat {g.seat + 1} · {playedWhen(g)}
                </div>
              </div>
              {#if g.rejoin}
                <button
                  type="button"
                  class="primary"
                  disabled={opening !== ""}
                  onclick={() => open(g)}
                  aria-label={`open ${g.name}`}
                >
                  {opening === g.id ? "…" : "Open"}
                  <Icon name="chevronRight" size={13} />
                </button>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
      {#if error}
        <p class="notice err" role="alert">
          <Icon name="x" size={14} />
          <span>{error}</span>
        </p>
      {/if}
    </div>

    <p class="foot">
      <a class="ghost-link" href={backHref}><Icon name="chevronLeft" size={12} /> Back</a>
    </p>
  </div>
</section>

<style>
  /* Same shell as Reclaim.svelte and Join.svelte. */
  .entry {
    position: relative;
    min-height: calc(100vh - 3rem);
    display: flex;
    justify-content: center;
    margin: -1.5rem;
    padding: 1.5rem;
  }
  .topbar {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 44px;
    display: flex;
    align-items: center;
    padding: 0 14px;
  }
  .wordmark {
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 14px;
    letter-spacing: 0.18em;
    color: var(--fg);
    display: inline-flex;
    align-items: center;
    gap: 8px;
    text-decoration: none;
  }
  .wordmark i {
    display: inline-block;
    width: 14px;
    height: 14px;
    border: 2px solid var(--gold);
    transform: rotate(45deg);
    border-radius: 3px;
    box-sizing: border-box;
  }
  .stack {
    width: min(860px, 100%);
    display: flex;
    flex-direction: column;
    gap: 22px;
    margin-top: clamp(60px, 12vh, 110px);
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
  .games {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
  }
  .game {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 0;
    border-top: 1px solid var(--border);
  }
  .game:first-child {
    border-top: none;
    padding-top: 0;
  }
  .info {
    flex: 1;
    min-width: 0;
  }
  .line {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }
  .name {
    font-weight: 700;
    color: var(--fg);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .meta {
    margin-top: 3px;
    font-size: 12px;
    color: var(--fg-muted);
  }
  .chip {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 2px 7px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    color: var(--fg-muted);
  }
  .chip.live {
    color: var(--gold);
    border-color: var(--gold);
  }
  .chip.won {
    color: var(--fg);
    border-color: var(--gold);
    background: rgba(217, 180, 92, 0.14);
  }
  .chip.archived {
    opacity: 0.7;
  }
  .notice {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    margin: 12px 0 0;
    font-size: 13.5px;
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
  .help.first {
    margin-top: 0;
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
  .foot {
    margin: 0;
    font-size: 12px;
  }
  .ghost-link {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: var(--fg-muted);
    text-decoration: none;
  }
  .ghost-link:hover {
    color: var(--fg);
  }
</style>
