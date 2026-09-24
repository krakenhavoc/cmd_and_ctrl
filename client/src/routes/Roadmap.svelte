<script lang="ts">
  // Roadmap — what the rules engine supports, keyword by keyword and
  // gap by gap (#1386, ADR 0092).
  //
  // Public on purpose: App.svelte lists "roadmap" as a public route and
  // the server mounts GET /roadmap without auth.Middleware. That is safe
  // because of what the body is: card NAMES and caveat sentences this
  // repo wrote, never art or oracle text (ADR 0092 Decision 1). A
  // signed-in viewer's example card names link into the catalogue,
  // which shows the art behind its own session gate; a signed-out
  // viewer sees plain text.
  //
  // The data is server/internal/roadmap's registry, checked against the
  // live catalog in CI, so the page says nothing a test does not hold.

  import { onMount } from "svelte";
  import Icon from "../lib/components/Icon.svelte";
  import SiteHeader from "../lib/components/SiteHeader.svelte";
  import { session } from "../lib/session";
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

  let query = $state("");
  let kinds = $state<RoadmapKind[]>([]);
  let statuses = $state<RoadmapStatus[]>([]);

  const isSignedIn = $derived($session !== null);

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

  function reset() {
    query = "";
    kinds = [];
    statuses = [];
  }

  function statusCount(s: RoadmapStatus): number {
    return data ? data.counts[s] : 0;
  }

  function fmt(n: number): string {
    return n.toLocaleString("en-US");
  }

  onMount(() => {
    const ac = new AbortController();
    fetchRoadmap(ac.signal)
      .then((res) => {
        data = res;
      })
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === "AbortError") return;
        error = err instanceof Error ? err.message : "couldn't load the roadmap";
      })
      .finally(() => {
        loading = false;
      });
    return () => ac.abort();
  });
</script>

<!-- A card name: a catalogue link for a signed-in viewer, plain text
     otherwise (the catalogue needs a session; see the header comment). -->
{#snippet cardName(name: string)}
  {#if isSignedIn}
    <a class="cname" href={catalogSearchHash(name)}>{name}</a>
  {:else}
    <span class="cname">{name}</span>
  {/if}
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

<SiteHeader />

<section class="entry">
  <div class="stack">
    <div class="head">
      <p class="eyebrow">Engine roadmap</p>
      <h1>What the engine plays, and what's next</h1>
      <p class="lede">
        Every keyword, mechanic and engine gap the rules engine knows about.
        <strong class="t-implemented">Implemented</strong> means it resolves on its own.
        <strong class="t-partial">Partial</strong> means most of it works, and the entry says which
        part you still do by hand.
        <strong class="t-missing">Missing</strong> means it isn't built yet: cards that need it still
        play, and you resolve that part yourself.
      </p>
    </div>

    {#if loading}
      <p class="help" aria-live="polite">Loading the roadmap…</p>
    {:else if error}
      <p class="notice err" role="alert">
        <Icon name="x" size={14} />
        <span>Couldn't load the roadmap ({error}).</span>
      </p>
    {:else if data}
      <ul class="legend">
        {#each STATUSES as s (s)}
          <li class="lg-item {s}">
            <span class="badge {s}">{STATUS_LABELS[s]}</span>
            <span class="lg-n">{statusCount(s)}</span>
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
                      <strong>{it.unblocks}</strong>
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

      <div class="filters">
        <input
          type="text"
          class="search"
          placeholder="search a mechanic, keyword or card name"
          aria-label="search the roadmap"
          bind:value={query}
        />
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
            <strong>{shown.length}</strong>
            {shown.length === 1 ? "entry" : "entries"}
            {#if filtered}<span class="dim">of {items.length}</span>{/if}
          </p>
          {#if filtered}
            <button type="button" class="ghost sm" onclick={reset}>
              <Icon name="x" size={12} /> clear filters
            </button>
          {/if}
        </div>
      </div>

      {#if shown.length === 0}
        <p class="help">Nothing on the roadmap matches those filters.</p>
      {/if}

      {#each STATUSES as s (s)}
        {#if groups[s].length > 0}
          <section class="status-sec {s}" aria-labelledby="rm-{s}">
            <div class="sec-head">
              <h2 id="rm-{s}">
                <span class="dot {s}" aria-hidden="true"></span>
                {STATUS_LABELS[s]}
                <span class="sec-n">{groups[s].length}</span>
              </h2>
              <p class="sec-note">{SECTION_NOTES[s]}</p>
            </div>

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
                      <span class="lbl">What's missing</span>
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
          </section>
        {/if}
      {/each}
    {/if}
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
    width: min(1180px, 100%);
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 22px;
    margin: clamp(8px, 2vh, 24px) 0 48px;
  }

  .eyebrow {
    margin: 0 0 6px;
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--gold-strong);
  }
  h1 {
    margin: 0;
    font-family: var(--font-display);
    font-weight: 800;
    letter-spacing: -0.01em;
  }
  .lede {
    margin: 10px 0 0;
    max-width: 68ch;
    color: var(--fg-muted);
    font-size: 14px;
    line-height: 1.6;
  }
  .lede strong {
    font-weight: 700;
  }
  .t-implemented {
    color: var(--mint);
  }
  .t-partial {
    color: var(--gold-strong);
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
    --tone: color-mix(in srgb, var(--gold) 60%, transparent);
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
    color: var(--gold-strong);
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
    color: var(--gold-strong);
    border-color: color-mix(in srgb, var(--gold) 45%, transparent);
    background: var(--gold-soft);
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
    background: var(--gold);
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
    --tone: var(--gold);
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
    color: var(--gold-strong);
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
  .search {
    width: 100%;
    box-sizing: border-box;
    margin-top: 10px;
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
    border-color: var(--gold);
    color: var(--gold-strong);
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
    --tone: var(--gold);
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
    text-decoration-color: color-mix(in srgb, var(--gold) 50%, transparent);
    text-underline-offset: 2px;
  }
  a.cname:hover {
    color: var(--gold-strong);
    text-decoration-color: var(--gold);
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
    border-left: 2px solid color-mix(in srgb, var(--gold) 50%, transparent);
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
    color: var(--gold-strong);
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
