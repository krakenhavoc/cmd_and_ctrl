<script lang="ts">
  // One account (ADR 0124 §3.2, §7): its row, its sign-in state, its
  // tables, its saved decks and its deck requests. Sessions are HMAC
  // tokens with no row on the server (ADR 0044 decision 3), so there is
  // no list of them: the view shows the last sign-in and the revocation
  // time, and links the existing revoke (POST /admin/users/{id}/
  // revoke-sessions) behind a confirmation that names the person.
  import { onMount } from "svelte";
  import { fetchAdminAccount, revokeUserSessions } from "../../api";
  import {
    accountsHash,
    avatarSrc,
    externalLink,
    gameHash,
    gamesHash,
    outcomeLine,
    revokeConfirm,
    seatNumber,
    stateLabel,
    type AdminAccountResponse,
  } from "../../adminViews";
  import { LobbyApiError, session } from "../../session";
  import Icon from "../Icon.svelte";
  import RelTime from "./RelTime.svelte";

  interface Props {
    id: string;
    onforbidden: () => void;
  }
  const { id, onforbidden }: Props = $props();

  let d = $state<AdminAccountResponse | null>(null);
  let loading = $state(false);
  let error = $state("");
  let confirming = $state(false);
  let busy = $state(false);
  let done = $state("");

  async function load(): Promise<void> {
    loading = true;
    error = "";
    try {
      d = await fetchAdminAccount(id);
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 403) {
        onforbidden();
        return;
      }
      error =
        err instanceof LobbyApiError
          ? err.status === 404
            ? "No such account."
            : err.message
          : "couldn't load this account";
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void load();
  });

  async function revoke(): Promise<void> {
    if (!d) return;
    busy = true;
    error = "";
    done = "";
    try {
      const res = await revokeUserSessions(d.account.id);
      confirming = false;
      const n = res.sockets_closed;
      done = `Signed out everywhere${n > 0 ? `; ${n} ${n === 1 ? "connection" : "connections"} closed` : ""}.`;
      await load();
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 403) {
        onforbidden();
        return;
      }
      error = err instanceof LobbyApiError ? `revoke failed: ${err.message}` : "revoke failed";
    } finally {
      busy = false;
    }
  }

  const now = $derived(d?.generated_at ?? Date.now());
  const avatar = $derived(d ? avatarSrc(d.account.avatar_url, $session?.token) : null);
</script>

<div class="view">
  <p class="back">
    <a href={accountsHash()}><Icon name="chevronLeft" size={12} /> All accounts</a>
  </p>

  {#if error}
    <p class="notice err" role="alert">{error}</p>
  {/if}

  {#if d}
    {@const a = d.account}
    <div class="card">
      <div class="thead">
        {#if avatar}<img class="avatar" src={avatar} alt="" width="36" height="36" />{/if}
        <h2>{a.name}</h2>
        {#if a.playing_now}<span class="chip live">playing now</span>{/if}
      </div>
      <dl class="facts">
        <div>
          <dt>First seen</dt>
          <dd><RelTime ms={a.first_seen_at} {now} /></dd>
        </div>
        <div>
          <dt>Games played</dt>
          <dd>{a.games_played}</dd>
        </div>
        <div>
          <dt>Last played</dt>
          <dd><RelTime ms={a.last_played_at} {now} /></dd>
        </div>
      </dl>
    </div>

    <div class="card">
      <h3>Sign-in</h3>
      <dl class="facts">
        <div>
          <dt>Last sign-in</dt>
          <dd><RelTime ms={d.sign_in.last_sign_in_at} {now} /></dd>
        </div>
        <div>
          <dt>Discord linked</dt>
          <dd><RelTime ms={d.sign_in.discord_linked_at} {now} /></dd>
        </div>
        <div>
          <dt>Sessions revoked</dt>
          <dd>
            {#if d.sign_in.sessions_invalid_before}
              <RelTime ms={d.sign_in.sessions_invalid_before} {now} />
            {:else}
              <span class="dim">never</span>
            {/if}
          </dd>
        </div>
      </dl>
      <p class="help">
        Sessions are signed tokens with no list on the server, so there are none to show. Revoking
        refuses every session issued before now.
      </p>
      <div class="actions">
        <button type="button" class="ghost" disabled={busy} onclick={() => (confirming = true)}>
          Revoke sessions
        </button>
      </div>
      {#if confirming}
        <div class="confirm" role="alert">
          <p>{revokeConfirm(a.name)}</p>
          <div class="confirm-actions">
            <button type="button" class="primary" disabled={busy} onclick={revoke}>
              sign them out everywhere
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
      <div class="thead">
        <h3>Tables · {d.games.length}{d.games_truncated ? "+" : ""}</h3>
        <a class="more" href={gamesHash({ user: a.id })}>in Games</a>
      </div>
      {#if d.games.length === 0}
        <p class="none">No tables yet.</p>
      {:else}
        <div class="scroll">
          <table aria-label="their tables">
            <thead>
              <tr>
                <th scope="col">Table</th>
                <th scope="col">Seat</th>
                <th scope="col">State</th>
                <th scope="col">Outcome</th>
                <th scope="col">Ended</th>
              </tr>
            </thead>
            <tbody>
              {#each d.games as g (g.id)}
                <tr>
                  <td class="tcell">
                    <div>
                      <a class="tname" href={gameHash(g.id)}>{g.name || "untitled table"}</a>
                      {#if g.archived_at}<span class="chip dim">archived</span>{/if}
                    </div>
                  </td>
                  <td>{g.their_seat === undefined ? "—" : seatNumber(g.their_seat)}</td>
                  <td>
                    <span class="chip" class:live={g.state === "active"}>{stateLabel(g.state)}</span
                    >
                  </td>
                  <td>{outcomeLine(g)}</td>
                  <td><RelTime ms={g.ended_at ?? g.archived_at} {now} /></td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
        {#if d.games_truncated}
          <p class="help">The newest 500 only.</p>
        {/if}
      {/if}
    </div>

    <div class="card">
      <h3>Saved decks · {d.decks.length}</h3>
      {#if d.decks.length === 0}
        <p class="none">No saved decks.</p>
      {:else}
        <ul class="rows" aria-label="their decks">
          {#each d.decks as deck (deck.id)}
            {@const link = externalLink(deck.source_url)}
            <li>
              <span class="rname">
                {#if link}
                  <a href={link} target="_blank" rel="noopener noreferrer">{deck.name}</a>
                {:else}
                  {deck.name}
                {/if}
              </span>
              <span class="dim">
                {deck.commanders.join(" & ") || "no commander"} · {deck.card_count} cards · {deck.format}
                · updated <RelTime ms={deck.updated_at} {now} />
              </span>
            </li>
          {/each}
        </ul>
      {/if}
    </div>

    <div class="card">
      <h3>Deck requests · {d.deck_requests.length}{d.deck_requests_truncated ? "+" : ""}</h3>
      {#if d.deck_requests.length === 0}
        <p class="none">No deck requests.</p>
      {:else}
        <ul class="rows" aria-label="their deck requests">
          {#each d.deck_requests as r, i (i)}
            {@const issue = externalLink(r.issue_url)}
            <li>
              <span class="rname">{r.deck_key}</span>
              <span class="dim">
                asked <RelTime ms={r.asked_at} {now} />
                {#if issue}
                  · <a href={issue} target="_blank" rel="noopener noreferrer"
                    >#{r.issue_number ?? "issue"}</a
                  >
                {/if}
              </span>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  {:else if loading}
    <p class="empty" aria-live="polite">Loading the account…</p>
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
  .thead h3 {
    margin: 0;
  }
  .more {
    margin-left: auto;
    font-size: 12px;
    color: var(--accent-strong);
  }
  .avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: var(--surface-raised);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 10px;
  }
  .rows {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
  }
  .rows li {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 8px 0;
    border-top: 1px solid var(--border);
    font-size: 13px;
  }
  .rows li:first-child {
    border-top: none;
    padding-top: 0;
  }
  .rname {
    font-weight: 600;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .rname a {
    color: var(--fg);
  }
  .rows .dim {
    font-size: 12px;
  }
  .none {
    margin: 0;
    font-size: 12.5px;
    color: var(--fg-muted);
  }
  .help {
    margin: 8px 0 0;
    font-size: 12px;
    color: var(--fg-dim);
  }
</style>
