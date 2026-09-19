<script lang="ts">
  import { onMount } from "svelte";
  import {
    discordAuthEnabled,
    joinGame,
    previewGame,
    spectateGame,
    type GamePreview,
    type SeatInfo,
  } from "../lib/api";
  import { navigate } from "../lib/router";
  import { LobbyApiError, session } from "../lib/session";
  import { signedInUserID } from "../lib/myGames";
  import { seatColor } from "../lib/colors";
  import Icon from "../lib/components/Icon.svelte";

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
    void loadPreview();
    if (spectator) return; // not surfaced for spectator flow
    void discordAuthEnabled().then((on) => {
      discordEnabled = on;
    });
  });

  // The table behind the invite — name, state, who's seated — shown
  // before asking for anything. GET /games/{id}/preview?t=<invite>
  // accepts either invite; a bad one leaves the page on the bare
  // form with the server's reason.
  let preview = $state<GamePreview | null>(null);
  let previewError = $state("");
  let previewLoading = $state(false);
  async function loadPreview(): Promise<void> {
    if (!inviteToken) return;
    previewLoading = true;
    previewError = "";
    try {
      preview = await previewGame(gameID, inviteToken);
    } catch (err) {
      preview = null;
      previewError = err instanceof LobbyApiError ? err.message : "couldn't load the table";
    } finally {
      previewLoading = false;
    }
  }

  const seated = $derived(preview ? [...preview.game.players].sort((a, b) => a.seat - b.seat) : []);
  const maxSeats = $derived(preview?.max_seats ?? 4);
  const openSeats = $derived(Math.max(0, maxSeats - seated.length));
  const tableFull = $derived(preview !== null && !spectator && openSeats === 0);
  const tableStarted = $derived(preview !== null && !spectator && preview.game.state !== "lobby");
  const seatName = (p: SeatInfo): string => p.display_name || p.name;
  const initials = (p: SeatInfo): string => seatName(p).trim().slice(0, 1).toUpperCase() || "?";

  // discordHref is a plain link into /auth/discord/start rather
  // than a fetch — the server issues a 302 to Discord, which the
  // browser must follow at the top level (not inside an XHR) so
  // the user actually lands on the Discord consent screen.
  const discordHref = $derived(
    `/auth/discord/start?game=${encodeURIComponent(gameID)}&t=${encodeURIComponent(inviteToken)}`,
  );

  // A person already signed in with Discord joins as themselves: the
  // server takes the seat's name, avatar and user from the session and
  // ignores a typed name (ADR 0051 sub-PR 4), so the page does not ask
  // for one. Player invites only — a spectator's label is just a label.
  const signedInAs = $derived.by(() => {
    const s = $session;
    if (spectator || !s) return null;
    if (s.principal.role !== "identified" && !signedInUserID(s)) return null;
    return s.principal.name || "your Discord account";
  });

  async function joinSignedIn(): Promise<void> {
    busy = true;
    error = "";
    try {
      await joinGame(gameID, inviteToken, "");
      navigate("#/lobby");
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "join failed";
    } finally {
      busy = false;
    }
  }

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

<section class="entry">
  <header class="topbar">
    <a class="wordmark" href="#/login" aria-label="cmd_and_ctrl home"><i></i>CMD &amp; CTRL</a>
  </header>
  <div class="stack">
    <div class="head">
      <p class="eyebrow">{spectator ? "You have a spectator link" : "You're invited to a table"}</p>
      <h1>
        {#if preview}
          <span class="table-name">{preview.game.name}</span>
          <span class="sr-only">— {spectator ? "spectate game" : "join game"}</span>
        {:else}
          {spectator ? "spectate game" : "join game"}
        {/if}
      </h1>
      <div class="chips">
        {#if preview?.game.state === "active"}
          <span class="chip live"><i class="dot"></i> in progress</span>
        {:else if preview?.game.state === "ended"}
          <span class="chip">ended</span>
        {:else}
          <span class="chip">lobby</span>
        {/if}
        {#if spectator}
          <span class="chip ro"><Icon name="exile" size={11} /> read-only</span>
        {/if}
        {#if preview}
          <span class="meta">
            {seated.length} of {maxSeats} seats{preview.game.state === "lobby"
              ? " · starts when everyone's deck is in"
              : ""}
          </span>
        {:else}
          <span class="meta">game <code>{gameID.slice(0, 8)}</code></span>
        {/if}
      </div>
    </div>

    {#if preview}
      <ul class="seats" aria-label="seats at this table">
        {#each seated as p (p.seat)}
          <li class="seat">
            <span class="sav" style:border-color={seatColor(p.seat)}>
              <i style:background={seatColor(p.seat)}>{initials(p)}</i>
            </span>
            <div class="sinfo">
              <div class="sname">{seatName(p)}</div>
              <div class="sstat">
                {#if p.deck_uploaded}{p.deck_name || "deck ready"}{:else}deck pending{/if}
              </div>
            </div>
          </li>
        {/each}
        {#each Array.from({ length: openSeats }, (_, i) => i) as i (i)}
          <li class="seat open" class:yours={!spectator && i === 0 && !tableStarted}>
            <span class="sav open"></span>
            <div class="sinfo">
              <div class="sname dim">Open seat</div>
              <div class="sstat you">
                {!spectator && i === 0 && !tableStarted ? "This one's yours" : "—"}
              </div>
            </div>
          </li>
        {/each}
      </ul>
    {/if}

    <div class="card">
      {#if !inviteToken}
        <p class="notice err" role="alert">
          <Icon name="x" size={14} />
          <span>invite token missing from URL — ask your host to resend the link.</span>
        </p>
      {:else if tableFull}
        <p class="notice err" role="alert">
          <Icon name="x" size={14} />
          <span
            ><strong>This table is full.</strong> All {maxSeats} seats were claimed before you got here.</span
          >
        </p>
        <p class="help">
          If a seat frees up, this same link will let you in — check again in a bit. To watch
          instead, ask your host for a spectator link.
        </p>
        <div class="frow">
          <button type="button" class="lg" onclick={loadPreview} disabled={previewLoading}>
            <Icon name="untap" size={14} /> Check again
          </button>
        </div>
      {:else if tableStarted}
        <p class="notice err" role="alert">
          <Icon name="x" size={14} />
          <span
            ><strong>This table has already started.</strong> Seats can't be claimed once the game is
            underway.</span
          >
        </p>
        <p class="help">Ask your host for a spectator link to watch.</p>
      {:else}
        {#if signedInAs}
          <button type="button" class="primary lg" disabled={busy} onclick={joinSignedIn}>
            {busy ? "…" : `Join as ${signedInAs}`}
            <Icon name="chevronRight" size={14} />
          </button>
        {:else}
          {#if discordEnabled}
            <a class="primary discord-btn" href={discordHref}>
              Continue with Discord <Icon name="chevronRight" size={14} />
            </a>
            <p class="help centered">
              Your Discord name and avatar become your seat, and the bot can ping you when it's your
              turn.
            </p>
            <div class="divider" aria-hidden="true"><span>or</span></div>
          {/if}
          <form class="frow" onsubmit={submit}>
            <input
              type="text"
              placeholder={spectator ? "your name (chat label)" : "your name"}
              bind:value={name}
              required
            />
            <button
              type="submit"
              class="lg"
              class:primary={spectator || !discordEnabled}
              disabled={busy || !name.trim()}
            >
              {#if busy}
                …
              {:else if spectator}
                watch <Icon name="chevronRight" size={14} />
              {:else if discordEnabled}
                join with a name
              {:else}
                join <Icon name="chevronRight" size={14} />
              {/if}
            </button>
          </form>
        {/if}
        {#if spectator}
          <p class="help">
            You'll see the table from a non-seated viewpoint. Opponent hands and libraries stay
            hidden, the same way they do for any other player. You can't send actions.
          </p>
        {:else}
          <p class="help">
            Joining claims an open seat. You'll add your deck in the lobby — a Moxfield or Archidekt
            link is enough.
          </p>
        {/if}
      {/if}

      {#if previewError && !error}
        <p class="help dim">
          Couldn't load the table ({previewError}) — you can still try to join.
        </p>
      {/if}
      {#if error}
        <p class="error" role="alert">{error}</p>
        {#if /full/i.test(error)}
          <p class="help">
            If a seat frees up, this same link will let you in — check again in a bit. To watch
            instead, ask your host for a spectator link.
          </p>
        {/if}
      {/if}
    </div>
  </div>
</section>

<style>
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
  }
  .card {
    width: min(600px, 100%);
    box-sizing: border-box;
    align-self: center;
    margin-top: 6px;
  }
  .table-name {
    overflow-wrap: anywhere;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
  .seats {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 8px;
  }
  @media (max-width: 760px) {
    .seats {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }
  .seat {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 12px;
    border-radius: 10px;
    background: var(--surface);
    border: 1px solid var(--border);
    min-width: 0;
  }
  .seat.open {
    border: 1px dashed var(--border-strong);
    background: transparent;
  }
  .seat.open.yours {
    border-color: rgba(217, 180, 92, 0.5);
    background: var(--gold-soft);
  }
  .sav {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    border: 2px solid var(--border-strong);
    padding: 2px;
    box-sizing: border-box;
    background: var(--surface-raised);
    flex: 0 0 auto;
    overflow: hidden;
  }
  .sav i {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    border-radius: 50%;
    font-style: normal;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 12px;
    color: #1c1503;
  }
  .sav.open {
    border: 2px dashed var(--border-strong);
    background: transparent;
  }
  .seat.open.yours .sav.open {
    border-color: rgba(217, 180, 92, 0.6);
  }
  .sinfo {
    min-width: 0;
  }
  .sname {
    font-size: 12px;
    font-weight: 700;
    color: var(--fg);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .sname.dim {
    color: var(--fg-dim);
    font-weight: 500;
  }
  .seat.open.yours .sname.dim {
    color: var(--gold-strong);
    font-weight: 700;
  }
  .sstat {
    font-size: 10.5px;
    color: var(--fg-muted);
    line-height: 1.25;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .sstat.you {
    color: var(--fg-dim);
  }
  .seat.open.yours .sstat.you {
    color: var(--gold-strong);
  }
  .chip.live {
    color: var(--gold-strong);
    border-color: rgba(217, 180, 92, 0.5);
    background: var(--gold-soft);
  }
  .dot {
    display: inline-block;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--gold);
  }
  .help.dim {
    color: var(--fg-dim);
  }
  .head {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .eyebrow {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--gold-strong);
    font-weight: 600;
  }
  .head h1 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 34px;
    font-weight: 800;
    letter-spacing: -0.02em;
    color: var(--fg);
    line-height: 1.05;
  }
  .chips {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    height: 20px;
    padding: 0 8px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
    font-weight: 600;
  }
  .meta {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--fg-dim);
    letter-spacing: 0.04em;
  }
  .meta code {
    color: var(--fg-muted);
  }
  .card {
    background: var(--surface);
    border: 1px solid var(--border-strong);
    border-radius: 16px;
    padding: 22px 24px 24px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    box-shadow: var(--shadow-lg);
  }
  .discord-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    width: 100%;
    height: 40px;
    box-sizing: border-box;
    border-radius: var(--radius);
    background: var(--accent);
    border: 1px solid var(--accent-strong);
    color: var(--accent-fg);
    font-size: 13.5px;
    font-weight: 600;
    text-decoration: none;
  }
  .discord-btn:hover {
    background: var(--accent-strong);
    color: var(--accent-fg);
  }
  .divider {
    display: flex;
    align-items: center;
    gap: 10px;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .divider::before,
  .divider::after {
    content: "";
    flex: 1;
    height: 1px;
    background: var(--border);
  }
  .frow {
    display: flex;
    gap: 8px;
  }
  .frow input {
    flex: 1;
    min-width: 0;
    margin: 0;
    height: 40px;
    padding: 0 12px;
    box-sizing: border-box;
    font-size: 13.5px;
  }
  .lg {
    height: 40px;
    padding: 0 16px;
    font-size: 13.5px;
    flex: 0 0 auto;
  }
  .help {
    margin: 0;
    font-size: 12px;
    color: var(--fg-muted);
    line-height: 1.5;
  }
  .help.centered {
    text-align: center;
    margin-top: -4px;
  }
  .notice {
    margin: 0;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 14px;
    border-radius: 10px;
    font-size: 12.5px;
  }
  .notice.err {
    background: rgba(255, 107, 107, 0.08);
    border: 1px solid rgba(255, 107, 107, 0.35);
    color: var(--fg);
  }
  .error {
    margin: 0;
    color: var(--danger);
    font-size: 12.5px;
  }
</style>
