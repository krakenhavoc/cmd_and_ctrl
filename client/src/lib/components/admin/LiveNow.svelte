<script lang="ts">
  // Live now (ADR 0124 §3.4, §7): who is connected, to which table, as
  // a seat, a spectator or an admin, plus the bot and agent seats of
  // every running table. The drill-down behind the Overview's Players
  // connected, Spectators and Bot seats tiles; its totals come from the
  // same function as those tiles (metrics.Tally), so they differ only by
  // the time between a scrape and this page's last ask.
  //
  // It asks GET /admin/live every 10 seconds while the tab is visible,
  // stops while it is hidden, and pauses after 10 minutes with no input
  // (LivePoller, lib/adminViews.ts). Every ask is an `admin action` log
  // line on the server, which is why the pause exists.
  import { onMount } from "svelte";
  import { fetchAdminLive } from "../../api";
  import {
    gameHash,
    LivePoller,
    seatNumber,
    stateLabel,
    type LiveNowResponse,
    type PollState,
  } from "../../adminViews";
  import { LobbyApiError } from "../../session";
  import AccountLink from "./AccountLink.svelte";
  import RelTime from "./RelTime.svelte";
  import SeatList from "./SeatList.svelte";

  interface Props {
    onforbidden: () => void;
  }
  const { onforbidden }: Props = $props();

  let data = $state<LiveNowResponse | null>(null);
  let error = $state("");
  let pollState = $state<PollState>("polling");
  let inFlight = false;

  async function load(): Promise<void> {
    if (inFlight) return;
    inFlight = true;
    try {
      data = await fetchAdminLive();
      error = "";
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 403) {
        poller.stop();
        onforbidden();
        return;
      }
      error = err instanceof LobbyApiError ? err.message : "couldn't load who is on now";
    } finally {
      inFlight = false;
    }
  }

  const poller = new LivePoller({
    load: () => void load(),
    onState: (s) => (pollState = s),
  });

  onMount(() => {
    poller.start();
    pollState = poller.state;
    return () => poller.stop();
  });

  function input(): void {
    poller.input();
  }

  const now = $derived(data?.generated_at ?? Date.now());
  const totals = $derived(data?.totals);
</script>

<svelte:window onpointerdown={input} onpointermove={input} onkeydown={input} onwheel={input} />
<svelte:document onvisibilitychange={() => poller.visibilityChanged()} />

<section class="live" aria-label="live now">
  <div class="toolbar">
    <p class="status" aria-live="polite">
      {#if pollState === "paused"}
        Paused.
        <button type="button" class="ghost sm" onclick={() => poller.resume()}>Resume</button>
      {:else if pollState === "hidden"}
        Not refreshing while this tab is hidden.
      {:else if data}
        Refreshes every 10 s · updated <RelTime ms={data.generated_at} now={Date.now()} />
      {:else}
        Loading…
      {/if}
    </p>
  </div>

  {#if error}
    <p class="notice err" role="alert">{error}</p>
  {/if}

  {#if totals}
    <dl class="totals">
      <div>
        <dt>Players connected</dt>
        <dd>{totals.players_connected}</dd>
      </div>
      <div>
        <dt>Spectators</dt>
        <dd>{totals.spectators}</dd>
      </div>
      <div>
        <dt>Bot seats</dt>
        <dd>{totals.bot_seats}</dd>
      </div>
      <div>
        <dt>Practice tables</dt>
        <dd>{totals.practice_tables}</dd>
      </div>
      <div>
        <dt>Admin views</dt>
        <dd>{totals.admin_views}</dd>
      </div>
    </dl>
    <p class="help">
      Players connected counts running tables only, as the Overview's tile does; people waiting at a
      lobby table are listed below but not counted.
    </p>
  {/if}

  {#if data && data.unbound_sockets > 0}
    <p class="notice warn" role="status">
      {data.unbound_sockets}
      {data.unbound_sockets === 1 ? "socket is" : "sockets are"} bound to a table the lobby does not hold.
    </p>
  {/if}

  {#if data}
    {#if data.tables.length === 0}
      <p class="empty">Nobody is connected, and no table is running.</p>
    {:else}
      <ul class="tables">
        {#each data.tables as t (t.id)}
          <li class="card">
            <div class="thead">
              <a class="tname" href={gameHash(t.id)}>{t.name || "untitled table"}</a>
              <span class="chip" class:live={t.state === "active"}>{stateLabel(t.state)}</span>
              {#if t.practice}<span class="chip">practice</span>{/if}
              {#if t.archived}<span class="chip dim">archived</span>{/if}
            </div>
            <SeatList seats={t.seats} {now} label={`seats at ${t.name}`} />
            {#if t.spectators.length > 0}
              <div class="group">
                <h4>Spectators · {t.spectators.length}</h4>
                <ul class="people">
                  {#each t.spectators as sp, i (i)}
                    <li>
                      {#if sp.account}<AccountLink account={sp.account} />{:else}<span class="anon"
                          >guest spectator</span
                        >{/if}
                      <span class="since"><RelTime ms={sp.since} {now} prefix="since" /></span>
                    </li>
                  {/each}
                </ul>
              </div>
            {/if}
            {#if t.admins.length > 0}
              <div class="group">
                <h4>Admins · {t.admins.length}</h4>
                <ul class="people">
                  {#each t.admins as a, i (i)}
                    <li>
                      {#if a.account}<AccountLink account={a.account} />{:else}<span class="anon"
                          >admin token</span
                        >{/if}
                      {#if a.as_seat !== undefined}<span class="tag"
                          >as {seatNumber(a.as_seat)}</span
                        >{/if}
                      <span class="since"><RelTime ms={a.since} {now} prefix="since" /></span>
                    </li>
                  {/each}
                </ul>
              </div>
            {/if}
          </li>
        {/each}
      </ul>
    {/if}
  {/if}
</section>

<style>
  .live {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    flex-wrap: wrap;
  }
  .status {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0;
    font-size: 12.5px;
    color: var(--fg-muted);
  }
  .totals {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
    gap: 10px;
    margin: 0;
  }
  .totals div {
    padding: 10px 12px;
    border: 1px solid var(--border);
    border-radius: 12px;
    background: var(--surface);
  }
  .totals dt {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .totals dd {
    margin: 4px 0 0;
    font-family: var(--font-display);
    font-size: 24px;
    font-weight: 800;
    color: var(--fg);
  }
  .help {
    margin: -4px 0 0;
    font-size: 12px;
    color: var(--fg-dim);
  }
  .tables {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
    padding: 14px 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
  }
  .thead {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }
  .tname {
    font-weight: 700;
    font-size: 15px;
    color: var(--fg);
    text-decoration: none;
    overflow-wrap: anywhere;
  }
  .tname:hover {
    text-decoration: underline;
  }
  .chip,
  .tag {
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
    color: var(--accent);
    border-color: var(--accent);
  }
  .chip.dim {
    opacity: 0.7;
  }
  .group h4 {
    margin: 4px 0 4px;
    font-family: var(--font-mono);
    font-size: 10.5px;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .people {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 13px;
  }
  .people li {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px 10px;
  }
  .anon {
    color: var(--fg-muted);
  }
  .since {
    font-size: 12px;
    color: var(--fg-muted);
  }
  .empty {
    margin: 0;
    padding: 18px;
    text-align: center;
    font-size: 13px;
    color: var(--fg-muted);
    border: 1px dashed var(--border-strong);
    border-radius: 12px;
  }
  .notice {
    margin: 0;
    font-size: 13px;
  }
  .notice.err {
    color: var(--danger);
  }
  .notice.warn {
    color: var(--priority-strong);
  }
  button.ghost.sm {
    height: 26px;
    padding: 0 10px;
    font-size: 12px;
  }
</style>
