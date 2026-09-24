<script lang="ts">
  // Catalog — the public list of every card this engine automates.
  //
  // Public on purpose: no session, no dev gate, and the server route
  // behind it is mounted outside auth.Middleware. App.svelte's auth
  // effect lists "catalog" as a public route; without that entry a
  // signed-out visitor would be bounced to /login.
  //
  // The page's one job is to be honest. A card is shown as Complete
  // only where a human declared it so in the card file; anything
  // nobody has audited shows as Unreviewed, in its own bucket, never
  // folded in with the working ones.

  import { onMount } from "svelte";
  import Icon from "../lib/components/Icon.svelte";
  import SiteHeader from "../lib/components/SiteHeader.svelte";
  import { cardArt } from "../lib/cardArt";
  import {
    fetchCatalog,
    filterCatalog,
    toggle,
    catalogImageURL,
    COLOR_FILTERS,
    COLOR_LABELS,
    TYPE_FILTERS,
    COMPLETENESS_LABELS,
    COMPLETENESS_BLURBS,
    type CatalogEntry,
    type CatalogCounts,
    type Completeness,
    type ColorFilter,
    type TypeFilter,
  } from "../lib/catalog";

  let cards = $state<CatalogEntry[]>([]);
  let counts = $state<CatalogCounts | null>(null);
  let loading = $state(true);
  let error = $state("");

  let query = $state("");
  let colors = $state<ColorFilter[]>([]);
  let types = $state<TypeFilter[]>([]);
  let completeness = $state<Completeness[]>([]);

  const buckets: Completeness[] = ["full", "caveats", "unreviewed"];

  const shown = $derived(filterCatalog(cards, { query, colors, types, completeness }));
  const filtered = $derived(
    query.trim() !== "" || colors.length > 0 || types.length > 0 || completeness.length > 0,
  );

  function reset() {
    query = "";
    colors = [];
    types = [];
    completeness = [];
  }

  function bucketCount(b: Completeness): number {
    if (!counts) return 0;
    return b === "full" ? counts.full : b === "caveats" ? counts.caveats : counts.unreviewed;
  }

  onMount(() => {
    const ac = new AbortController();
    fetchCatalog(ac.signal)
      .then((res) => {
        cards = res.cards;
        counts = res.counts;
      })
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === "AbortError") return;
        error = err instanceof Error ? err.message : "couldn't load the catalogue";
      })
      .finally(() => {
        loading = false;
      });
    return () => ac.abort();
  });
</script>

<SiteHeader />

<section class="entry">
  <div class="stack">
    <div class="head">
      <p class="eyebrow">What the engine plays</p>
      <h1>Card catalogue</h1>
      <p class="lede">
        Every card wired into the rules engine, and how completely each one is implemented. Cards
        not listed here still work in the sandbox — you just resolve them by hand.
      </p>
    </div>

    {#if loading}
      <p class="help" aria-live="polite">Loading the catalogue…</p>
    {:else if error}
      <p class="notice err" role="alert">
        <Icon name="x" size={14} />
        <span>Couldn't load the catalogue ({error}).</span>
      </p>
    {:else}
      <!-- The legend is the honest part of the page and deliberately
           sits above the grid rather than behind a tooltip. -->
      <ul class="legend">
        {#each buckets as b (b)}
          <li class="lg-item">
            <span class="badge {b}">{COMPLETENESS_LABELS[b]}</span>
            <span class="lg-n">{bucketCount(b)}</span>
            <p class="lg-blurb">{COMPLETENESS_BLURBS[b]}</p>
          </li>
        {/each}
      </ul>

      <div class="filters">
        <input
          type="text"
          class="search"
          placeholder="search name, type or rules text"
          aria-label="search the catalogue"
          bind:value={query}
        />

        <div class="facets">
          <fieldset class="facet">
            <legend>Colour identity</legend>
            <div class="chips">
              {#each COLOR_FILTERS as c (c)}
                <button
                  type="button"
                  class="fchip pip p{c}"
                  class:on={colors.includes(c)}
                  aria-pressed={colors.includes(c)}
                  title={COLOR_LABELS[c]}
                  onclick={() => (colors = toggle(colors, c))}
                >
                  {c}<span class="sr-only"> {COLOR_LABELS[c]}</span>
                </button>
              {/each}
            </div>
          </fieldset>

          <fieldset class="facet">
            <legend>Type</legend>
            <div class="chips">
              {#each TYPE_FILTERS as t (t)}
                <button
                  type="button"
                  class="fchip"
                  class:on={types.includes(t)}
                  aria-pressed={types.includes(t)}
                  onclick={() => (types = toggle(types, t))}
                >
                  {t}
                </button>
              {/each}
            </div>
          </fieldset>

          <fieldset class="facet">
            <legend>Completeness</legend>
            <div class="chips">
              {#each buckets as b (b)}
                <button
                  type="button"
                  class="fchip {b}"
                  class:on={completeness.includes(b)}
                  aria-pressed={completeness.includes(b)}
                  onclick={() => (completeness = toggle(completeness, b))}
                >
                  {COMPLETENESS_LABELS[b]}
                </button>
              {/each}
            </div>
          </fieldset>
        </div>

        <div class="countbar">
          <p class="count" aria-live="polite">
            <strong>{shown.length}</strong>
            {shown.length === 1 ? "card" : "cards"}
            {#if filtered && counts}<span class="dim">of {counts.total}</span>{/if}
          </p>
          {#if filtered}
            <button type="button" class="ghost sm" onclick={reset}>
              <Icon name="x" size={12} /> clear filters
            </button>
          {/if}
        </div>
      </div>

      {#if shown.length === 0}
        <p class="help">No card matches those filters.</p>
      {:else}
        <ul class="grid">
          {#each shown as c (c.oracle_id)}
            {@const img = catalogImageURL(c, "normal")}
            <li class="tile">
              <div class="art">
                {#if img}
                  <img
                    src={img}
                    alt={c.name}
                    loading="lazy"
                    decoding="async"
                    width="488"
                    height="680"
                    use:cardArt={img}
                  />
                {:else}
                  <div class="noart">
                    <span class="noart-name">{c.name}</span>
                    <span class="noart-why">no printing in the card data</span>
                  </div>
                {/if}
              </div>
              <div class="meta">
                <p class="tname">{c.name}</p>
                {#if c.type_line}<p class="ttype">{c.type_line}</p>{/if}
                <span class="badge {c.completeness}">{COMPLETENESS_LABELS[c.completeness]}</span>
                {#if c.caveats && c.caveats.length > 0}
                  <ul class="caveats">
                    {#each c.caveats as cav (cav)}
                      <li>{cav}</li>
                    {/each}
                  </ul>
                {/if}
              </div>
            </li>
          {/each}
        </ul>
      {/if}
    {/if}
  </div>
</section>

<style>
  /* Shell shared with Join.svelte and Reclaim.svelte via SiteHeader;
     the centering wrapper below is otherwise unchanged. */
  .entry {
    position: relative;
    min-height: calc(100vh - 3rem - 68px);
    display: flex;
    justify-content: center;
  }
  .stack {
    width: min(1180px, 100%);
    display: flex;
    flex-direction: column;
    gap: 22px;
    margin-top: clamp(8px, 2vh, 24px);
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
  }
  .lede {
    margin: 10px 0 0;
    max-width: 62ch;
    color: var(--fg-muted);
    font-size: 14px;
    line-height: 1.55;
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
    border-radius: var(--radius-lg);
    padding: 12px 14px;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
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

  /* --- badges ---------------------------------------------------- */
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
  .badge.full {
    color: var(--mint);
    border-color: color-mix(in srgb, var(--mint) 45%, transparent);
    background: color-mix(in srgb, var(--mint) 12%, transparent);
  }
  .badge.caveats {
    color: var(--gold-strong);
    border-color: color-mix(in srgb, var(--gold) 45%, transparent);
    background: var(--gold-soft);
  }
  .badge.unreviewed {
    color: var(--fg-dim);
  }

  /* --- filters --------------------------------------------------- */
  .filters {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .search {
    width: 100%;
    box-sizing: border-box;
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
  /* Colour pips. Same geometry and typography as every other chip —
     only the tint differs, so this stays inside the existing visual
     language rather than starting a new one. The tints are the
     universal WUBRG palette rather than something invented here;
     colour identity is the one facet where colour IS the information,
     and a card browser whose colour filter is monochrome is harder to
     use than one that is not.

     Unselected pips keep the plain chip surface and only pick up
     their colour on the text, so a row of six does not shout. */
  .fchip.pip {
    width: 30px;
    padding: 0;
    justify-content: center;
  }
  .fchip.pW {
    color: #e8dcc0;
  }
  .fchip.pU {
    color: #7fb4e8;
  }
  .fchip.pB {
    color: #b09ac4;
  }
  .fchip.pR {
    color: #e88b7f;
  }
  .fchip.pG {
    color: #7fc79a;
  }
  .fchip.pC {
    color: var(--fg-muted);
  }
  .fchip.pip.on {
    border-color: currentColor;
    background: color-mix(in srgb, currentColor 16%, transparent);
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

  /* --- grid ------------------------------------------------------ */
  .grid {
    list-style: none;
    margin: 0 0 40px;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
    gap: 16px;
  }
  .tile {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
  }
  .art {
    /* positioned for the failed-art pip (#33) */
    position: relative;
    --art-error-top: 8px;
    --art-error-right: 8px;
    aspect-ratio: 488 / 680;
    border-radius: 12px;
    overflow: hidden;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
  }
  .art img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .noart {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 14px;
    box-sizing: border-box;
    text-align: center;
  }
  .noart-name {
    font-family: var(--font-display);
    font-size: 14px;
    color: var(--fg);
  }
  .noart-why {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .meta {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 5px;
    min-width: 0;
  }
  .tname {
    margin: 0;
    font-size: 13.5px;
    font-weight: 600;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .ttype {
    margin: 0;
    font-size: 11.5px;
    color: var(--fg-dim);
    overflow-wrap: anywhere;
  }
  .caveats {
    list-style: none;
    margin: 2px 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .caveats li {
    font-size: 11.5px;
    line-height: 1.45;
    color: var(--fg-muted);
    border-left: 2px solid color-mix(in srgb, var(--gold) 50%, transparent);
    padding-left: 8px;
  }

  @media (max-width: 760px) {
    .legend {
      grid-template-columns: minmax(0, 1fr);
    }
    .grid {
      grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
      gap: 12px;
    }
    .facets {
      gap: 12px;
    }
  }
</style>
