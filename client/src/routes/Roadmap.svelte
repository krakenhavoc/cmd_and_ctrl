<script lang="ts">
  // Roadmap — what the rules engine supports, keyword by keyword and
  // gap by gap (#1386, ADR 0092).
  //
  // Public on purpose: App.svelte lists "roadmap" as a public route and
  // the server mounts GET /roadmap without auth.Middleware. That is safe
  // because of what the body is: card NAMES and caveat sentences this
  // repo wrote, never art or oracle text (ADR 0092 Decision 1). The
  // example card names link into the catalogue, which shows the art.
  //
  // The data is server/internal/roadmap's registry, checked against the
  // live catalog in CI, so the page says nothing a test does not hold.

  import { onMount } from "svelte";
  import Icon from "../lib/components/Icon.svelte";
  import Page from "../lib/components/Page.svelte";
  import { L } from "../lib/labels";
  import { toggle } from "../lib/catalog";
  import {
    fetchRoadmap,
    filterRoadmap,
    groupByStatus,
    nextItems,
    impactLine,
    issueURL,
    adrURL,
    adrLabel,
    catalogSearchHash,
    parseRoadmapHash,
    roadmapHash,
    KINDS,
    STATUSES,
    KIND_LABELS,
    STATUS_LABELS,
    STATUS_BLURBS,
    type RoadmapResponse,
    type RoadmapKind,
    type RoadmapStatus,
  } from "../lib/roadmap";

  let data = $state<RoadmapResponse | null>(null);
  let loading = $state(true);
  let error = $state("");

  // The search and the chips live in the hash (#/roadmap?q=…&kind=…&status=…),
  // so a filtered view can be linked. They are read back on load and when
  // the hash changes under us; every change is written with replaceState,
  // so typing does not fill the history and the router is not re-run.
  const fromHash = parseRoadmapHash(typeof location === "undefined" ? "" : location.hash);
  let query = $state(fromHash.query);
  let kinds = $state<RoadmapKind[]>(fromHash.kinds);
  let statuses = $state<RoadmapStatus[]>(fromHash.statuses);

  $effect(() => {
    const want = roadmapHash({ query, kinds, statuses });
    if (location.hash !== want && location.hash.startsWith("#/roadmap")) {
      history.replaceState(history.state, "", want);
    }
  });

  function onHashChange(): void {
    if (!location.hash.startsWith("#/roadmap")) return;
    const f = parseRoadmapHash(location.hash);
    query = f.query;
    kinds = f.kinds;
    statuses = f.statuses;
  }

  const items = $derived(data?.items ?? []);
  const upNext = $derived(data ? nextItems(data) : []);
  const shown = $derived(filterRoadmap(items, { query, kinds, statuses }));
  const groups = $derived(groupByStatus(shown));
  const filtered = $derived(query.trim() !== "" || kinds.length > 0 || statuses.length > 0);

  const SECTION_NOTES: Record<RoadmapStatus, string> = {
    implemented: "Works end to end. Each lists a few cards that use it.",
    partial: "Works, with a named gap you resolve by hand.",
    missing: "Not built yet, and the cards known to be waiting on it.",
  };

  // Each status section starts closed: its heading and count are the page
  // until the reader opens it (ADR 0156 §7). "Up next" is not a status
  // section and stays open. While a search or chip filter is active every
  // section with a match opens, so a result is never hidden behind a
  // closed heading; the reader can still close one by hand afterwards.
  let open = $state<Record<RoadmapStatus, boolean>>({
    implemented: false,
    partial: false,
    missing: false,
  });

  $effect(() => {
    if (!filtered) return;
    for (const s of STATUSES) {
      if (groups[s].length > 0) open[s] = true;
    }
  });

  function reset() {
    query = "";
    kinds = [];
    statuses = [];
  }

  function statusCount(s: RoadmapStatus): number {
    return data ? data.counts[s] : 0;
  }

  const numberFormat = new Intl.NumberFormat("en-US");
  function fmt(n: number): string {
    return numberFormat.format(n);
  }

  onMount(() => {
    const ac = new AbortController();
    fetchRoadmap(ac.signal)
      .then((res) => {
        data = res;
      })
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === "AbortError") return;
        error = err instanceof Error ? err.message : "couldn’t load the roadmap";
      })
      .finally(() => {
        loading = false;
      });
    return () => ac.abort();
  });
</script>

<!-- A card name: a link into the public catalogue. -->
{#snippet cardName(name: string)}
  <a class="cname" href={catalogSearchHash(name)}>{name}</a>
{/snippet}

{#snippet links(issue: number | undefined, adr: string | undefined, rules: string[] | undefined)}
  {#if issue || adr || (rules && rules.length > 0)}
    <p class="links">
      {#if issue}
        <a href={issueURL(issue)} target="_blank" rel="noreferrer noopener">#{issue}</a>
      {/if}
      {#if adr}
        <a href={adrURL(adr)} target="_blank" rel="noreferrer noopener">{adrLabel(adr)}</a>
      {/if}
      {#if rules && rules.length > 0}
        <span class="rules">CR {rules.join(", ")}</span>
      {/if}
    </p>
  {/if}
{/snippet}

{#snippet searchBox()}
  <input
    type="search"
    name="roadmap-search"
    autocomplete="off"
    spellcheck="false"
    class="search"
    placeholder="search a mechanic, keyword or card name"
    aria-label={L.searchRoadmap}
    disabled={loading}
    bind:value={query}
  />
{/snippet}

<svelte:window onhashchange={onHashChange} />

<Page eyebrow="Cards" title="What the engine plays, and what’s next" width="wide">
  {#snippet lede()}
    Every keyword, mechanic and engine gap the rules engine knows about.
    <strong class="t-implemented">Implemented</strong> means it resolves on its own.
    <strong class="t-partial">Partial</strong> means most of it works, and the entry says which part
    you still do by hand.
    <strong class="t-missing">Missing</strong> means it isn’t built yet: cards that need it still play,
    and you resolve that part yourself.
  {/snippet}
  <div class="stack">
    {#if !loading && error}
      <p class="notice err" role="alert">
        <Icon name="x" size={14} />
        <span>Couldn’t load the roadmap ({error}). Reload the page to try again.</span>
      </p>
    {/if}
    {#if data}
      <ul class="legend">
        {#each STATUSES as s (s)}
          <li class="lg-item {s}">
            <span class="badge {s}">{STATUS_LABELS[s]}</span>
            <span class="lg-n">{fmt(statusCount(s))}</span>
            <p class="lg-blurb">{STATUS_BLURBS[s]}</p>
          </li>
        {/each}
      </ul>
      <p class="cardline">
        <span class="cl-label">Cards automated</span>
        <span><strong>{fmt(data.counts.cards.total)}</strong> in all</span>
        <span class="sep" aria-hidden="true">·</span>
        <span><strong class="t-implemented">{fmt(data.counts.cards.full)}</strong> complete</span>
        <span class="sep" aria-hidden="true">·</span>
        <span><strong class="t-partial">{fmt(data.counts.cards.caveats)}</strong> with caveats</span
        >
        <span class="sep" aria-hidden="true">·</span>
        <span><strong>{fmt(data.counts.cards.unreviewed)}</strong> unreviewed</span>
      </p>

      {#if upNext.length > 0}
        <section class="next" aria-labelledby="rm-next">
          <div class="sec-head">
            <h2 id="rm-next">Up next</h2>
            <p class="sec-note">Ranked by how many cards each one would unlock.</p>
          </div>
          <ol class="next-grid">
            {#each upNext as it, i (it.slug)}
              <li class="next-card {it.status}">
                <div class="nc-top">
                  <span class="nc-rank">{String(i + 1).padStart(2, "0")}</span>
                  <span class="badge {it.status}">{STATUS_LABELS[it.status]}</span>
                  <span class="kind">{KIND_LABELS[it.kind]}</span>
                </div>
                <h3 class="nc-name">{it.name}</h3>
                <p class="nc-text">{it.missing ?? it.summary}</p>
                <div class="nc-foot">
                  {#if (it.unblocks ?? 0) > 0}
                    <p class="nc-impact">
                      <strong>{fmt(it.unblocks ?? 0)}</strong>
                      <span>{it.unblocks === 1 ? "card" : "cards"} unblocked</span>
                    </p>
                  {:else if it.waiting && it.waiting.length > 0}
                    <p class="nc-impact">
                      <strong>{it.waiting.length}</strong>
                      <span>{it.waiting.length === 1 ? "card" : "cards"} waiting</span>
                    </p>
                  {/if}
                  {#if it.issue}
                    <a
                      class="nc-issue"
                      href={issueURL(it.issue)}
                      target="_blank"
                      rel="noreferrer noopener"
                    >
                      #{it.issue}
                      <Icon name="link" size={11} />
                    </a>
                  {/if}
                </div>
              </li>
            {/each}
          </ol>
        </section>
      {/if}
    {/if}

    <!-- The search box is the roadmap.search hint's anchor: it must exist before the
           data does, or a slow load makes the hint engine skip the page's own tip (#3305),
           and it must be the SAME element before and after, or the tip's anchor is torn
           out from under it when the data lands (#3343: the hints e2e saw "Element is not
           attached"). So it sits in one block that is open for the whole load and only the
           facets and the count come and go inside it. -->
    {#if loading || data}
      <div class="filters">
        {@render searchBox()}
        {#if data}
          <div class="facets">
            <fieldset class="facet">
              <legend>Kind</legend>
              <div class="chips">
                {#each KINDS as k (k)}
                  <button
                    type="button"
                    class="fchip"
                    class:on={kinds.includes(k)}
                    aria-pressed={kinds.includes(k)}
                    onclick={() => (kinds = toggle(kinds, k))}
                  >
                    {KIND_LABELS[k]}
                  </button>
                {/each}
              </div>
            </fieldset>
            <fieldset class="facet">
              <legend>Status</legend>
              <div class="chips">
                {#each STATUSES as s (s)}
                  <button
                    type="button"
                    class="fchip {s}"
                    class:on={statuses.includes(s)}
                    aria-pressed={statuses.includes(s)}
                    onclick={() => (statuses = toggle(statuses, s))}
                  >
                    {STATUS_LABELS[s]}
                  </button>
                {/each}
              </div>
            </fieldset>
          </div>
          <div class="countbar">
            <p class="count" aria-live="polite">
              <strong>{fmt(shown.length)}</strong>
              {shown.length === 1 ? "entry" : "entries"}
              {#if filtered}<span class="dim">of {fmt(items.length)}</span>{/if}
            </p>
            {#if filtered}
              <button type="button" class="ghost sm" onclick={reset}>
                <Icon name="x" size={12} /> clear filters
              </button>
            {/if}
          </div>
        {/if}
      </div>
    {/if}

    {#if loading}
      <p class="help" aria-live="polite">Loading the roadmap…</p>
    {/if}

    {#if data}
      {#if shown.length === 0}
        <p class="help">Nothing on the roadmap matches those filters.</p>
      {/if}

      {#each STATUSES as s (s)}
        {#if groups[s].length > 0}
          <details class="status-sec {s}" bind:open={open[s]} data-status={s}>
            <summary class="sec-head">
              <span class="chev" aria-hidden="true"></span>
              <h2 id="rm-{s}">
                <span class="dot {s}" aria-hidden="true"></span>
                {STATUS_LABELS[s]}
                <span class="sec-n">{fmt(groups[s].length)}</span>
              </h2>
              <span class="sec-note">{SECTION_NOTES[s]}</span>
            </summary>

            <ul class="items {s}">
              {#each groups[s] as it (it.slug)}
                <li class="item {s}">
                  <div class="it-head">
                    <h3 class="it-name">{it.name}</h3>
                    <span class="kind">{KIND_LABELS[it.kind]}</span>
                  </div>
                  <p class="it-summary">{it.summary}</p>

                  {#if s !== "implemented" && it.missing}
                    <p class="it-missing">
                      <span class="lbl">What’s missing</span>
                      {it.missing}
                    </p>
                  {/if}

                  {#if it.examples && it.examples.length > 0}
                    <p class="it-cards">
                      <span class="lbl">{s === "implemented" ? "Plays in" : "Works in"}</span>
                      {#each it.examples as ex (ex.oracle_id)}
                        <span class="ci">{@render cardName(ex.name)}</span>
                      {/each}
                    </p>
                  {/if}

                  {#if s !== "implemented" && it.partial_examples && it.partial_examples.length > 0}
                    <div class="it-partial">
                      <span class="lbl">Works with a gap</span>
                      <ul>
                        {#each it.partial_examples as pe (pe.name)}
                          <li>
                            {@render cardName(pe.name)} <span class="dash">—</span>
                            {pe.caveat}
                          </li>
                        {/each}
                      </ul>
                    </div>
                  {/if}

                  {#if s === "missing" && it.waiting && it.waiting.length > 0}
                    <div class="it-waiting">
                      <span class="lbl">Waiting on it</span>
                      <ul class="wchips">
                        {#each it.waiting as w (w)}
                          <li>{w}</li>
                        {/each}
                      </ul>
                    </div>
                  {/if}

                  {#if s !== "implemented" && impactLine(it)}
                    <p class="it-impact">{impactLine(it)}</p>
                  {/if}

                  {@render links(it.issue, it.adr, it.rules)}
                </li>
              {/each}
            </ul>
          </details>
        {/if}
      {/each}
    {/if}
  </div>
</Page>

<style>
  .lg-n,
  .sec-n,
  .nc-impact strong,
  .count,
  .cardline {
    font-variant-numeric: tabular-nums;
  }
  .stack {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 22px;
  }

  .t-implemented {
    color: var(--mint);
  }
  .t-partial {
    color: var(--accent-strong);
  }
  .t-missing {
    color: var(--rose);
  }
  .help {
    margin: 0;
    color: var(--fg-muted);
    font-size: 13px;
  }
  .dim {
    color: var(--fg-dim);
  }
  .notice {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0;
    padding: 10px 12px;
    border-radius: var(--radius);
    background: var(--surface);
    border: 1px solid var(--border-strong);
    font-size: 13px;
  }
  .notice.err {
    color: var(--danger);
  }

  /* --- legend ---------------------------------------------------- */
  .legend {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 10px;
  }
  .lg-item {
    background: var(--surface);
    border: 1px solid var(--border);
    border-top: 2px solid var(--tone, var(--border-strong));
    border-radius: var(--radius-lg);
    padding: 12px 14px;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }
  .lg-item.implemented {
    --tone: color-mix(in srgb, var(--mint) 60%, transparent);
  }
  .lg-item.partial {
    --tone: color-mix(in srgb, var(--accent) 60%, transparent);
  }
  .lg-item.missing {
    --tone: color-mix(in srgb, var(--rose) 60%, transparent);
  }
  .lg-n {
    font-family: var(--font-mono);
    font-size: 15px;
    font-weight: 700;
    color: var(--fg);
  }
  .lg-blurb {
    flex-basis: 100%;
    margin: 0;
    color: var(--fg-muted);
    font-size: 12px;
    line-height: 1.45;
  }
  .cardline {
    margin: -8px 0 0;
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 4px 8px;
    font-size: 12.5px;
    color: var(--fg-muted);
  }
  .cardline strong {
    font-family: var(--font-mono);
    color: var(--fg);
  }
  .cardline strong.t-implemented {
    color: var(--mint);
  }
  .cardline strong.t-partial {
    color: var(--accent-strong);
  }
  .cl-label {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
    margin-right: 4px;
  }
  .sep {
    color: var(--fg-dim);
  }

  /* --- badges + kind tags --------------------------------------- */
  .badge {
    display: inline-flex;
    align-items: center;
    height: 20px;
    padding: 0 8px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    font-weight: 600;
    color: var(--fg-muted);
    white-space: nowrap;
  }
  .badge.implemented {
    color: var(--mint);
    border-color: color-mix(in srgb, var(--mint) 45%, transparent);
    background: color-mix(in srgb, var(--mint) 12%, transparent);
  }
  .badge.partial {
    color: var(--accent-strong);
    border-color: color-mix(in srgb, var(--accent) 45%, transparent);
    background: var(--accent-soft);
  }
  .badge.missing {
    color: var(--rose);
    border-color: color-mix(in srgb, var(--rose) 45%, transparent);
    background: color-mix(in srgb, var(--rose) 12%, transparent);
  }
  .kind {
    font-family: var(--font-mono);
    font-size: 9.5px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--fg-dim);
    white-space: nowrap;
  }

  /* --- section heads -------------------------------------------- */
  .sec-head {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 4px 14px;
    margin-bottom: 12px;
  }
  summary.sec-head {
    cursor: pointer;
    list-style: none;
    margin-bottom: 0;
    padding: 12px 14px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    align-items: center;
  }
  summary.sec-head::-webkit-details-marker {
    display: none;
  }
  summary.sec-head:hover {
    border-color: var(--border-strong);
  }
  summary.sec-head:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  details[open] > summary.sec-head {
    margin-bottom: 12px;
  }
  .chev {
    width: 7px;
    height: 7px;
    border-right: 2px solid var(--fg-dim);
    border-bottom: 2px solid var(--fg-dim);
    transform: rotate(-45deg);
    transition: transform 0.12s;
    flex: none;
  }
  details[open] > summary .chev {
    transform: rotate(45deg);
  }
  .sec-head h2 {
    margin: 0;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    letter-spacing: -0.005em;
    /* app.css styles a bare h2 as a small mono label; these are the
       page's section titles, so they read as headings instead. */
    text-transform: none;
    color: var(--fg);
  }
  .sec-note {
    margin: 0;
    color: var(--fg-dim);
    font-size: 12.5px;
  }
  .sec-n {
    font-family: var(--font-mono);
    font-size: 12px;
    font-weight: 600;
    color: var(--fg-dim);
  }
  .dot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: var(--fg-dim);
  }
  .dot.implemented {
    background: var(--mint);
  }
  .dot.partial {
    background: var(--accent);
  }
  .dot.missing {
    background: var(--rose);
  }

  /* --- up next --------------------------------------------------- */
  .next-grid {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(270px, 1fr));
    gap: 12px;
  }
  .next-card {
    --tone: var(--accent);
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
    padding: 14px 16px 14px;
    border-radius: var(--radius-lg);
    border: 1px solid color-mix(in srgb, var(--tone) 30%, var(--border));
    background:
      linear-gradient(160deg, color-mix(in srgb, var(--tone) 10%, transparent), transparent 55%),
      var(--surface);
    box-shadow: var(--shadow-sm);
  }
  .next-card.missing {
    --tone: var(--rose);
  }
  .nc-top {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }
  .nc-rank {
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 700;
    color: var(--tone);
    letter-spacing: 0.06em;
  }
  .nc-name {
    margin: 0;
    font-family: var(--font-display);
    font-size: 16px;
    font-weight: 700;
    line-height: 1.25;
    overflow-wrap: anywhere;
  }
  .nc-text {
    margin: 0;
    flex: 1;
    color: var(--fg-muted);
    font-size: 12.5px;
    line-height: 1.5;
  }
  .nc-foot {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 10px;
    padding-top: 8px;
    border-top: 1px solid var(--border);
  }
  .nc-impact {
    margin: 0;
    display: flex;
    align-items: baseline;
    gap: 6px;
    color: var(--fg-muted);
    font-size: 12px;
  }
  .nc-impact strong {
    font-family: var(--font-mono);
    font-size: 22px;
    font-weight: 700;
    line-height: 1;
    color: var(--fg);
  }
  .nc-issue {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-family: var(--font-mono);
    font-size: 11.5px;
    color: var(--accent-strong);
    text-decoration: none;
    white-space: nowrap;
  }
  .nc-issue:hover {
    text-decoration: underline;
  }

  /* --- filters (Catalog.svelte's, verbatim in spirit) ----------- */
  .filters {
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding-top: 6px;
    border-top: 1px solid var(--border);
  }
  /* A type="search" field: app.css styles the text inputs, so this one
     carries the same look itself. */
  .search {
    width: 100%;
    box-sizing: border-box;
    margin-top: 10px;
    background: var(--surface-sunken);
    color: var(--fg);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    padding: 0.55rem 0.8rem;
    font-size: 0.95rem;
    font-family: inherit;
    transition:
      border-color 120ms var(--ease),
      box-shadow 120ms var(--ease),
      background 120ms var(--ease);
  }
  .search:focus {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-soft);
    background: var(--bg-2);
  }
  @media (forced-colors: active) {
    .search:focus {
      outline: 2px solid Highlight;
    }
  }
  .facets {
    display: flex;
    flex-wrap: wrap;
    gap: 18px;
  }
  .facet {
    border: 0;
    margin: 0;
    padding: 0;
    min-width: 0;
  }
  .facet legend {
    padding: 0 0 6px;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .fchip {
    height: 26px;
    padding: 0 10px;
    border-radius: 999px;
    background: var(--surface);
    border: 1px solid var(--border);
    color: var(--fg-muted);
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    font-weight: 600;
    cursor: pointer;
  }
  .fchip:hover {
    background: var(--surface-hover);
    color: var(--fg);
  }
  .fchip.on {
    background: var(--accent-soft);
    border-color: var(--accent);
    color: var(--accent-strong);
  }
  .fchip.implemented.on {
    color: var(--mint);
    border-color: var(--mint);
    background: color-mix(in srgb, var(--mint) 14%, transparent);
  }
  .fchip.missing.on {
    color: var(--rose);
    border-color: var(--rose);
    background: color-mix(in srgb, var(--rose) 14%, transparent);
  }
  .countbar {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .count {
    margin: 0;
    font-size: 13px;
    color: var(--fg-muted);
  }
  .count strong {
    color: var(--fg);
    font-family: var(--font-mono);
  }
  button.sm {
    height: 26px;
    padding: 0 10px;
    font-size: 11.5px;
  }

  /* --- status sections ------------------------------------------ */
  .items {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 12px;
    grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  }
  .items.implemented {
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 10px;
  }
  .item {
    --tone: var(--accent);
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
    padding: 14px 16px;
    border-radius: var(--radius-lg);
    background: var(--surface);
    border: 1px solid var(--border);
    border-left: 3px solid color-mix(in srgb, var(--tone) 70%, transparent);
  }
  .item.implemented {
    --tone: var(--mint);
    padding: 12px 14px;
    gap: 6px;
  }
  .item.missing {
    --tone: var(--rose);
  }
  /* Kind first, visually, as a small overline: beside the name it
     squeezed long names into three or four lines. The DOM keeps the
     name first for screen readers. */
  .it-head {
    display: flex;
    flex-direction: column-reverse;
    align-items: flex-start;
    gap: 3px;
  }
  .it-name {
    margin: 0;
    font-family: var(--font-display);
    font-size: 15px;
    font-weight: 700;
    line-height: 1.25;
    overflow-wrap: anywhere;
  }
  .it-summary {
    margin: 0;
    color: var(--fg-muted);
    font-size: 12.5px;
    line-height: 1.5;
  }
  .item.implemented .it-summary {
    font-size: 12px;
  }
  .lbl {
    display: block;
    margin-bottom: 3px;
    font-family: var(--font-mono);
    font-size: 9.5px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .it-missing {
    margin: 0;
    padding: 8px 10px;
    border-radius: var(--radius);
    background: color-mix(in srgb, var(--tone) 8%, transparent);
    font-size: 12.5px;
    line-height: 1.5;
    color: var(--fg);
  }
  .it-missing .lbl {
    color: var(--tone);
  }
  .it-cards {
    margin: 0;
    font-size: 12px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .item.implemented .it-cards .lbl {
    display: inline;
    margin: 0 6px 0 0;
  }
  .cname {
    color: var(--fg);
    font-weight: 600;
  }
  a.cname {
    color: var(--fg);
    text-decoration: underline;
    text-decoration-color: color-mix(in srgb, var(--accent) 50%, transparent);
    text-underline-offset: 2px;
  }
  a.cname:hover {
    color: var(--accent-strong);
    text-decoration-color: var(--accent);
  }
  /* A comma-separated list without whitespace fights in the markup:
     the separator is the stylesheet's, so a card name never wraps away
     from its comma. */
  .ci:not(:last-child)::after {
    content: ",";
    margin-right: 0.35em;
    color: var(--fg-muted);
  }
  .it-partial ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .it-partial li {
    font-size: 11.5px;
    line-height: 1.45;
    color: var(--fg-muted);
    border-left: 2px solid color-mix(in srgb, var(--accent) 50%, transparent);
    padding-left: 8px;
    overflow-wrap: anywhere;
  }
  .dash {
    color: var(--fg-dim);
  }
  .wchips {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
  }
  .wchips li {
    max-width: 100%;
    padding: 2px 8px;
    border-radius: 999px;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    font-size: 11.5px;
    color: var(--fg-muted);
    overflow-wrap: anywhere;
  }
  .it-impact {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.02em;
    color: var(--tone);
  }
  .links {
    margin: auto 0 0;
    padding-top: 4px;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px 12px;
    font-family: var(--font-mono);
    font-size: 11px;
  }
  .links a {
    color: var(--accent-strong);
    text-decoration: none;
  }
  .links a:hover {
    text-decoration: underline;
  }
  .rules {
    color: var(--fg-dim);
  }

  @media (max-width: 760px) {
    .legend {
      grid-template-columns: minmax(0, 1fr);
    }
    .items,
    .items.implemented,
    .next-grid {
      grid-template-columns: minmax(0, 1fr);
    }
    .facets {
      gap: 12px;
    }
  }
</style>
