<script lang="ts">
  // One table (ADR 0124 §3.3, §7, call 2): its row, its seats (an
  // account opens its detail), its live connections at a loaded table,
  // and the linked actions, each the existing route with the Lobby's
  // confirmation: Open table (#/games/<id>, where an admin's socket gets
  // the full view), Archive or Unarchive, and Replay on an ended table.
  // Delete is not offered here; it stays on the Lobby's admin list.
  import { onMount } from "svelte";
  import { archiveGame, fetchAdminGame, replayURL, unarchiveGame } from "../../api";
  import {
    archiveConfirm,
    canArchive,
    canReplay,
    canUnarchive,
    gamesHash,
    outcomeLine,
    seatNumber,
    stateLabel,
    type AdminGameResponse,
  } from "../../adminViews";
  import { LobbyApiError } from "../../session";
  import Icon from "../Icon.svelte";
  import AccountLink from "./AccountLink.svelte";
  import RelTime from "./RelTime.svelte";
  import SeatList from "./SeatList.svelte";

  interface Props {
    id: string;
    onforbidden: () => void;
  }
  const { id, onforbidden }: Props = $props();

  let g = $state<AdminGameResponse | null>(null);
  let loading = $state(false);
  let error = $state("");
  let confirming = $state(false);
  let busy = $state(false);
  let done = $state("");

  async function load(): Promise<void> {
    loading = true;
    error = "";
    try {
      g = await fetchAdminGame(id);
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 403) {
        onforbidden();
        return;
      }
      error =
        err instanceof LobbyApiError
          ? err.status === 404
            ? "No such table."
            : err.message
          : "couldn't load this table";
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void load();
  });

  async function act(fn: () => Promise<unknown>, ok: string, failure: string): Promise<void> {
    busy = true;
    error = "";
    done = "";
    try {
      await fn();
      confirming = false;
      done = ok;
      await load();
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 403) {
        onforbidden();
        return;
      }
      error = err instanceof LobbyApiError ? `${failure}: ${err.message}` : failure;
    } finally {
      busy = false;
    }
  }

  const now = $derived(g?.generated_at ?? Date.now());
  const replay = $derived(g && canReplay(g) ? replayURL(g.id) : null);
</script>

<div class="view">
  <p class="back">
    <a href={gamesHash()}><Icon name="chevronLeft" size={12} /> All tables</a>
  </p>

  {#if error}
    <p class="notice err" role="alert">{error}</p>
  {/if}

  {#if g}
    <div class="card">
      <div class="thead">
        <h2>{g.name || "untitled table"}</h2>
        <span class="chip" class:live={g.state === "active"}>{stateLabel(g.state)}</span>
        {#if g.practice}<span class="chip">practice</span>{/if}
        {#if g.archived_at}<span class="chip dim">archived</span>{/if}
        {#if !g.loaded && (g.state === "lobby" || g.state === "active")}
          <span class="chip warn" title="this table's room is not in memory">not loaded</span>
        {/if}
      </div>

      <dl class="facts">
        <div>
          <dt>Outcome</dt>
          <dd>{outcomeLine(g)}</dd>
        </div>
        <div>
          <dt>Created</dt>
          <dd><RelTime ms={g.created_at} {now} /></dd>
        </div>
        <div>
          <dt>Started</dt>
          <dd><RelTime ms={g.started_at} {now} /></dd>
        </div>
        <div>
          <dt>Ended</dt>
          <dd><RelTime ms={g.ended_at} {now} /></dd>
        </div>
        <div>
          <dt>Archived</dt>
          <dd><RelTime ms={g.archived_at} {now} /></dd>
        </div>
        <div>
          <dt>Created by</dt>
          <dd>
            {#if g.creator}<AccountLink account={g.creator} />{:else}<span class="dim"
                >the admin token</span
              >{/if}
          </dd>
        </div>
        {#if g.spectators_connected !== undefined}
          <div>
            <dt>Spectators</dt>
            <dd>{g.spectators_connected}</dd>
          </div>
        {/if}
      </dl>

      <div class="actions">
        {#if g.loaded}
          <a class="btn-link" href={`#/games/${g.id}`}
            >Open table <Icon name="chevronRight" size={12} /></a
          >
        {/if}
        {#if replay}
          <a class="btn-link" href={replay} download={`${g.id}.jsonl`}>
            <Icon name="draw" size={13} /> Replay
          </a>
        {/if}
        {#if canUnarchive(g)}
          <button
            type="button"
            disabled={busy}
            onclick={() => act(() => unarchiveGame(g!.id), "Unarchived.", "unarchive failed")}
          >
            Unarchive
          </button>
        {:else if canArchive(g)}
          <button type="button" class="ghost" disabled={busy} onclick={() => (confirming = true)}>
            Archive
          </button>
        {/if}
      </div>

      {#if confirming}
        <div class="confirm" role="alert">
          <p>{archiveConfirm(g)}</p>
          <div class="confirm-actions">
            <button
              type="button"
              class="primary"
              disabled={busy}
              onclick={() => act(() => archiveGame(g!.id), "Archived.", "archive failed")}
            >
              {g.state === "active" ? "archive and end" : "archive table"}
            </button>
            <button type="button" class="ghost" disabled={busy} onclick={() => (confirming = false)}
              >cancel</button
            >
          </div>
        </div>
      {/if}
      {#if done}
        <p class="notice ok" role="status">{done}</p>
      {/if}
    </div>

    <div class="card">
      <h3>Seats</h3>
      <SeatList seats={g.seats} {now} winnerSeat={g.winner_seat} label="seats" />
    </div>

    {#if g.connections}
      <div class="card">
        <h3>Connected now · {g.connections.length}</h3>
        {#if g.connections.length === 0}
          <p class="none">Nobody is connected.</p>
        {:else}
          <ul class="people" aria-label="connections">
            {#each g.connections as c, i (i)}
              <li>
                <span class="chip">{c.kind}</span>
                {#if c.seat !== undefined}<span>{seatNumber(c.seat)}</span>{/if}
                {#if c.account}
                  <AccountLink account={c.account} />
                {:else}
                  <span class="dim">{c.kind === "admin" ? "admin token" : "guest"}</span>
                {/if}
                <span class="dim"><RelTime ms={c.since} {now} prefix="since" /></span>
              </li>
            {/each}
          </ul>
        {/if}
      </div>
    {/if}
  {:else if loading}
    <p class="empty" aria-live="polite">Loading the table…</p>
  {/if}
</div>

<style>
  h2 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 20px;
    font-weight: 800;
    color: var(--fg);
    text-transform: none;
    letter-spacing: -0.01em;
    overflow-wrap: anywhere;
  }
  h3 {
    margin: 0 0 8px;
    font-family: var(--font-mono);
    font-size: 10.5px;
    font-weight: 600;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 12px;
  }
  .people {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 13px;
  }
  .people li {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px 10px;
  }
  .none {
    margin: 0;
    font-size: 12.5px;
    color: var(--fg-muted);
  }
</style>
