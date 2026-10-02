<script lang="ts">
  import { onMount } from "svelte";
  import { deleteMyDeck, fetchMyDeckCoverage, fetchMyDecks, renameMyDeck } from "../lib/api";
  import {
    BUCKET_LABELS,
    BUCKET_ORDER,
    groupCardsByBucket,
    type CoverageReport,
  } from "../lib/deckcheck";
  import {
    coverageDetail,
    coverageLine,
    deckCheckHref,
    deckSubtitle,
    isSignedIn,
    sourceHost,
    type MyDeckInfo,
  } from "../lib/myDecks";
  import { LobbyApiError, session } from "../lib/session";
  import Icon from "../lib/components/Icon.svelte";
  import SiteHeader from "../lib/components/SiteHeader.svelte";

  // "My decks" (ADR 0110 section 6): the saved library, with how much of
  // each deck the engine plays as printed. The coverage is computed by
  // the server when the list is read, so it is right after every deploy.
  // Rename and delete are the owner's only; the report for a deck opens
  // in place from GET /me/decks/{id}/coverage.

  const signedIn = $derived(isSignedIn($session?.principal.user_id));

  let decks = $state<MyDeckInfo[]>([]);
  let loading = $state(false);
  let error = $state("");

  let renaming = $state("");
  let renameValue = $state("");
  let renameError = $state("");
  let confirmDelete = $state("");
  let busy = $state("");

  let reportFor = $state("");
  let report = $state<CoverageReport | null>(null);
  let reportError = $state("");

  onMount(() => {
    void load();
  });

  function message(err: unknown, fallback: string): string {
    return err instanceof LobbyApiError ? err.message : fallback;
  }

  async function load(): Promise<void> {
    if (!isSignedIn($session?.principal.user_id)) return;
    loading = true;
    error = "";
    try {
      decks = (await fetchMyDecks()).decks ?? [];
    } catch (err) {
      error = message(err, "couldn't load your decks");
    } finally {
      loading = false;
    }
  }

  function startRename(d: MyDeckInfo): void {
    renaming = d.id;
    renameValue = d.name;
    renameError = "";
    confirmDelete = "";
  }

  async function saveRename(d: MyDeckInfo): Promise<void> {
    const name = renameValue.trim();
    if (!name || busy) return;
    if (name === d.name) {
      renaming = "";
      return;
    }
    busy = d.id;
    renameError = "";
    try {
      const updated = await renameMyDeck(d.id, name);
      decks = decks.map((x) => (x.id === d.id ? { ...x, name: updated.name } : x));
      renaming = "";
    } catch (err) {
      renameError = message(err, "couldn't rename that deck");
    } finally {
      busy = "";
    }
  }

  async function remove(d: MyDeckInfo): Promise<void> {
    if (busy) return;
    busy = d.id;
    error = "";
    try {
      await deleteMyDeck(d.id);
      decks = decks.filter((x) => x.id !== d.id);
      if (reportFor === d.id) closeReport();
      confirmDelete = "";
    } catch (err) {
      error = message(err, "couldn't delete that deck");
    } finally {
      busy = "";
    }
  }

  function closeReport(): void {
    reportFor = "";
    report = null;
    reportError = "";
  }

  async function toggleReport(d: MyDeckInfo): Promise<void> {
    if (reportFor === d.id) {
      closeReport();
      return;
    }
    reportFor = d.id;
    report = null;
    reportError = "";
    try {
      const r = await fetchMyDeckCoverage(d.id);
      if (reportFor === d.id) report = r;
    } catch (err) {
      if (reportFor === d.id) reportError = message(err, "couldn't build the report");
    }
  }

  const grouped = $derived(report ? groupCardsByBucket(report.cards) : null);
</script>

<SiteHeader />

<section class="entry">
  <div class="stack">
    <div class="head">
      <p class="eyebrow">Decks you've played</p>
      <h1>my decks</h1>
    </div>

    <div class="card">
      {#if !signedIn}
        <p class="help first">
          Saved decks are kept for your Discord account. Sign in with Discord to see them.
        </p>
        <div class="frow">
          <a class="btn-link" href="#/login">Go to sign in</a>
        </div>
      {:else if loading && decks.length === 0}
        <p class="help first" aria-live="polite">Loading your decks…</p>
      {:else if decks.length === 0 && !error}
        <p class="help first">
          No saved decks yet. A deck you import at a table, from a link or a pasted list, is kept
          here.
        </p>
      {:else}
        <ul class="decks" aria-label="your decks">
          {#each decks as d (d.id)}
            <li class="deck">
              <div class="line">
                {#if renaming === d.id}
                  <form
                    class="rename"
                    onsubmit={(e) => {
                      e.preventDefault();
                      void saveRename(d);
                    }}
                  >
                    <input
                      type="text"
                      bind:value={renameValue}
                      maxlength="100"
                      aria-label={`new name for ${d.name}`}
                    />
                    <button
                      type="submit"
                      class="primary"
                      disabled={busy !== "" || !renameValue.trim()}
                    >
                      Save
                    </button>
                    <button type="button" onclick={() => (renaming = "")}>Cancel</button>
                  </form>
                {:else}
                  <span class="name">{d.name}</span>
                {/if}
              </div>
              {#if renaming === d.id && renameError}
                <p class="notice err" role="alert">{renameError}</p>
              {/if}
              <div class="meta">{deckSubtitle(d)}</div>
              {#if coverageLine(d)}
                <div class="cov">{coverageLine(d)}</div>
                {#if coverageDetail(d)}
                  <div class="cov-detail">{coverageDetail(d)}</div>
                {/if}
              {/if}
              {#if d.source_url}
                <div class="meta">
                  From <a href={d.source_url} target="_blank" rel="noopener noreferrer"
                    >{sourceHost(d.source_url)}</a
                  >
                </div>
              {/if}

              <div class="actions">
                {#if d.coverage}
                  <button
                    type="button"
                    onclick={() => toggleReport(d)}
                    aria-expanded={reportFor === d.id}
                  >
                    {reportFor === d.id ? "Hide report" : "Coverage report"}
                  </button>
                {/if}
                {#if d.source_url}
                  <a class="btn-link" href={deckCheckHref(d)}>Request missing cards</a>
                {/if}
                <button type="button" onclick={() => startRename(d)} disabled={busy !== ""}>
                  Rename
                </button>
                {#if confirmDelete === d.id}
                  <span class="confirm" role="alert">
                    Delete {d.name}?
                    <button
                      type="button"
                      class="danger"
                      disabled={busy !== ""}
                      onclick={() => remove(d)}
                    >
                      {busy === d.id ? "…" : "Delete"}
                    </button>
                    <button type="button" onclick={() => (confirmDelete = "")}>Keep</button>
                  </span>
                {:else}
                  <button
                    type="button"
                    onclick={() => {
                      confirmDelete = d.id;
                      renaming = "";
                    }}
                    disabled={busy !== ""}
                    aria-label={`delete ${d.name}`}
                  >
                    Delete
                  </button>
                {/if}
              </div>

              {#if reportFor === d.id}
                <div class="report">
                  {#if reportError}
                    <p class="notice err" role="alert">{reportError}</p>
                  {:else if !grouped || !report}
                    <p class="help first">Building the report…</p>
                  {:else}
                    {#each BUCKET_ORDER as b (b)}
                      {#if grouped[b].length > 0}
                        <details open={b === "manual" || b === "unreviewed"}>
                          <summary>{BUCKET_LABELS[b]} ({grouped[b].length})</summary>
                          <ul>
                            {#each grouped[b] as c (c.oracle_id || c.name)}
                              <li>
                                {c.name}
                                {#if c.caveats && c.caveats.length > 0}
                                  <span class="caveat">— {c.caveats.join(" ")}</span>
                                {/if}
                              </li>
                            {/each}
                          </ul>
                        </details>
                      {/if}
                    {/each}
                    {#if report.unknown.length > 0}
                      <p class="help">Not found in the card index: {report.unknown.join(", ")}</p>
                    {/if}
                  {/if}
                </div>
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
      <a class="ghost-link" href="#/lobby"><Icon name="chevronLeft" size={12} /> Back</a>
    </p>
  </div>
</section>

<style>
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
  .decks {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
  }
  .deck {
    padding: 14px 0;
    border-top: 1px solid var(--border);
  }
  .deck:first-child {
    border-top: none;
    padding-top: 0;
  }
  .name {
    font-weight: 700;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .meta {
    margin-top: 3px;
    font-size: 12px;
    color: var(--fg-muted);
  }
  .cov {
    margin-top: 4px;
    font-size: 12.5px;
    font-weight: 600;
    color: var(--mint);
  }
  .cov-detail {
    margin-top: 2px;
    font-size: 11.5px;
    color: var(--fg-dim);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    margin-top: 10px;
  }
  .rename {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    width: 100%;
  }
  .rename input {
    flex: 1 1 200px;
    min-width: 0;
  }
  .confirm {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 12.5px;
    color: var(--danger);
  }
  .danger {
    color: var(--danger);
    border-color: var(--danger);
  }
  .report {
    margin-top: 12px;
    padding: 10px 12px;
    border-radius: 8px;
    border: 1px solid var(--border);
    background: var(--surface-sunken);
    font-size: 12.5px;
  }
  .report details + details {
    margin-top: 8px;
  }
  .report summary {
    cursor: pointer;
    font-weight: 600;
    color: var(--fg);
  }
  .report ul {
    margin: 6px 0 0;
    padding-left: 18px;
    color: var(--fg-muted);
  }
  .caveat {
    color: var(--fg-dim);
  }
  .notice {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    margin: 8px 0 0;
    font-size: 13px;
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
