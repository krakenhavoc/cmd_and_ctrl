<script lang="ts">
  // Reports (#2984, ADR 0075 §2.6): what players have reported at their
  // tables, newest first. Each card names the table (a link to its admin
  // page), who reported whom, the reason, the account behind each side
  // when there is one, and the last few chat lines as the table saw them.
  // The server keeps a report for a fixed time (retention_days) and this
  // page says so. A report is a tip, not a case: Dismiss removes it, and
  // nothing else happens to anyone from here.
  //
  // Admin only, like every view under this tab strip: a 403 means the
  // admin status is stale (onforbidden), never a request to retry.
  import { onMount } from "svelte";
  import { dismissAdminReport, fetchAdminReports } from "../../api";
  import {
    accountHash,
    gameHash,
    type AdminReport,
    type AdminReportsResponse,
  } from "../../adminViews";
  import { LobbyApiError } from "../../session";
  import RelTime from "./RelTime.svelte";

  interface Props {
    onforbidden: () => void;
  }
  const { onforbidden }: Props = $props();

  let data = $state<AdminReportsResponse | null>(null);
  let loading = $state(false);
  let error = $state("");
  let busy = $state<string | null>(null);

  async function load(): Promise<void> {
    loading = true;
    error = "";
    try {
      data = await fetchAdminReports();
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 403) {
        onforbidden();
        return;
      }
      error = err instanceof LobbyApiError ? err.message : "couldn't load the reports";
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void load();
  });

  async function dismiss(r: AdminReport): Promise<void> {
    busy = r.id;
    error = "";
    try {
      await dismissAdminReport(r.id);
      if (data) data = { ...data, reports: data.reports.filter((x) => x.id !== r.id) };
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 403) {
        onforbidden();
        return;
      }
      error = err instanceof LobbyApiError ? err.message : "couldn't dismiss the report";
    } finally {
      busy = null;
    }
  }

  const now = $derived(Date.now());
</script>

<div class="view">
  <div class="thead">
    <p class="scope">
      {#if data}
        Reports are removed {data.retention_days} days after they are filed. Dismissing one only removes
        it from this list.
      {/if}
    </p>
    <button type="button" class="ghost sm" disabled={loading} onclick={() => void load()}>
      {loading ? "Refreshing…" : "Refresh"}
    </button>
  </div>

  {#if error}
    <p class="notice err" role="alert">{error}</p>
  {/if}

  {#if data}
    {#if data.reports.length === 0}
      <p class="empty">No reports.</p>
    {:else}
      <ul class="reports" aria-label="player reports">
        {#each data.reports as r (r.id)}
          <li class="card report">
            <div class="top">
              <a class="tname" href={gameHash(r.game_id)}>{r.game_name || "untitled table"}</a>
              <RelTime ms={r.created_at} {now} />
            </div>
            <p class="who">
              <strong>{r.reporter.name}</strong>
              {#if r.reporter.user_id}
                (<a href={accountHash(r.reporter.user_id)}>account</a>)
              {:else}
                <span class="chip dim">guest</span>
              {/if}
              reported
              <strong>{r.reported.name}</strong>
              {#if r.reported.user_id}
                (<a href={accountHash(r.reported.user_id)}>account</a>)
              {:else}
                <span class="chip dim">guest</span>
              {/if}
            </p>
            <p class="reason">{r.reason}</p>
            {#if r.excerpt.length > 0}
              <details>
                <summary>Recent chat ({r.excerpt.length})</summary>
                <ol class="excerpt">
                  {#each r.excerpt as line, i (i)}
                    <li class:reported={line.author === r.reported.name}>
                      <span class="author">{line.author}</span>
                      <span class="text">{line.text}</span>
                    </li>
                  {/each}
                </ol>
              </details>
            {:else}
              <p class="dim">No chat was recorded at this table.</p>
            {/if}
            <div class="actions">
              <button
                type="button"
                class="ghost sm"
                disabled={busy === r.id}
                onclick={() => void dismiss(r)}
              >
                {busy === r.id ? "Dismissing…" : "Dismiss"}
              </button>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  {:else if loading}
    <p class="empty" aria-live="polite">Loading reports…</p>
  {/if}
</div>

<style>
  .scope {
    margin: 0;
    font-size: 12.5px;
    color: var(--fg-muted);
    flex: 1 1 auto;
  }
  .reports {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .report {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .top {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    justify-content: space-between;
    gap: 8px;
  }
  .who,
  .reason {
    margin: 0;
    overflow-wrap: anywhere;
  }
  .reason {
    color: var(--fg);
    white-space: pre-wrap;
  }
  .excerpt {
    list-style: none;
    margin: 8px 0 0;
    padding: 8px 10px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    background: var(--surface-sunken);
    border-radius: 8px;
    font-size: 12.5px;
  }
  .excerpt li {
    display: flex;
    gap: 8px;
    min-width: 0;
  }
  .excerpt .author {
    flex: none;
    font-weight: 700;
    color: var(--fg-muted);
  }
  .excerpt .text {
    overflow-wrap: anywhere;
  }
  .excerpt li.reported .author {
    color: var(--danger);
  }
  summary {
    cursor: pointer;
    font-size: 12.5px;
    color: var(--fg-muted);
  }
  .actions {
    display: flex;
    justify-content: flex-end;
  }
</style>
