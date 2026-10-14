<script lang="ts">
  // Catalog — the list of every card this engine automates.
  //
  // Public (ADR 0042). The server mounts GET /catalog without a session
  // and signedInHome.ts lists "catalog" as a public route: sign-up needs
  // a role on the C2 Discord (ADR 0156), and Scryfall's guidelines rule
  // out putting card data behind a chat server (S70, #2985).
  //
  // The page's one job is to be honest. A card is shown as Complete
  // only where a human declared it so in the card file; anything
  // nobody has audited shows as Unreviewed, in its own bucket, never
  // folded in with the working ones.
  //
  // Layout (ADR 0158 §3): a sticky filter sidebar beside the grid; under
  // 1000px the sidebar becomes a sheet behind a Filters button. The
  // search box is rendered exactly once (in the sidebar, or in the bar
  // at tablet width) so its contract label names one control.

  import { onMount, tick } from "svelte";
  import Icon from "../lib/components/Icon.svelte";
  import Page from "../lib/components/Page.svelte";
  import ShopLinks from "../lib/components/ShopLinks.svelte";
  import ShopNote from "../lib/components/ShopNote.svelte";
  import { cardArt } from "../lib/cardArt";
  import { route } from "../lib/router";
  import { L } from "../lib/labels";
  import {
    fetchCatalog,
    filterCatalog,
    facetCounts,
    activeChips,
    removeChip,
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

  // #/catalog?q=<text> pre-fills the search (the roadmap links its
  // example card names here). Re-applied when the hash changes while
  // the page is open, so a second link lands on its own card.
  $effect(() => {
    const r = $route;
    if (r.name === "catalog" && r.query !== undefined) {
      query = r.query;
    }
  });
  let colors = $state<ColorFilter[]>([]);
  let types = $state<TypeFilter[]>([]);
  let completeness = $state<Completeness[]>([]);

  const buckets: Completeness[] = ["full", "caveats", "unreviewed"];

  const filter = $derived({ query, colors, types, completeness });
  const shown = $derived(filterCatalog(cards, filter));
  const facets = $derived(facetCounts(cards, filter));
  const chips = $derived(activeChips(filter));

  function reset() {
    query = "";
    colors = [];
    types = [];
    completeness = [];
  }

  // The catalogue's rows carry no "added" order (the server sorts by
  // name), so Name is the only sort and there is no Newest to offer.

  // --- tablet: the sidebar as a sheet ------------------------------
  let narrow = $state(false);
  let sheetOpen = $state(false);
  let sheetEl = $state<HTMLElement | null>(null);
  let filtersBtn = $state<HTMLButtonElement | null>(null);

  async function openSheet() {
    sheetOpen = true;
    await tick();
    sheetEl?.querySelector<HTMLElement>("button, input")?.focus();
  }
  function closeSheet() {
    if (!sheetOpen) return;
    sheetOpen = false;
    filtersBtn?.focus();
  }
  function sheetKeydown(ev: KeyboardEvent) {
    if (ev.key === "Escape") {
      ev.stopPropagation();
      closeSheet();
      return;
    }
    if (ev.key !== "Tab" || !sheetEl) return;
    const els = [...sheetEl.querySelectorAll<HTMLElement>("button, input, [href]")].filter(
      (e) => !e.hasAttribute("disabled"),
    );
    if (els.length === 0) return;
    const first = els[0];
    const last = els[els.length - 1];
    const active = document.activeElement;
    if (ev.shiftKey && (active === first || !sheetEl.contains(active))) {
      ev.preventDefault();
      last.focus();
    } else if (!ev.shiftKey && (active === last || !sheetEl.contains(active))) {
      ev.preventDefault();
      first.focus();
    }
  }

  onMount(() => {
    const mq = window.matchMedia("(max-width: 999px)");
    narrow = mq.matches;
    const onChange = (e: MediaQueryListEvent) => {
      narrow = e.matches;
      if (!e.matches) sheetOpen = false;
    };
    mq.addEventListener("change", onChange);

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
    return () => {
      ac.abort();
      mq.removeEventListener("change", onChange);
    };
  });
</script>

{#snippet searchBox()}
  <input
    type="text"
    class="search"
    placeholder="search name, type or rules text"
    aria-label={L.searchCatalogue}
    disabled={loading}
    bind:value={query}
  />
{/snippet}

{#snippet facetPanel()}
  <fieldset class="facet">
    <legend>Completeness</legend>
    <ul class="rows">
      {#each buckets as b (b)}
        <li>
          <button
            type="button"
            class="row"
            class:on={completeness.includes(b)}
            aria-pressed={completeness.includes(b)}
            title={COMPLETENESS_BLURBS[b]}
            onclick={() => (completeness = toggle(completeness, b))}
          >
            <span class="dot k-{b}" aria-hidden="true"></span>
            <span class="row-name">{COMPLETENESS_LABELS[b]}</span>
            <span class="row-n">{facets.completeness[b]}</span>
          </button>
        </li>
      {/each}
    </ul>
  </fieldset>

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
    <ul class="rows">
      {#each TYPE_FILTERS as t (t)}
        <li>
          <button
            type="button"
            class="row"
            class:on={types.includes(t)}
            aria-pressed={types.includes(t)}
            onclick={() => (types = toggle(types, t))}
          >
            <span class="row-name cap">{t}</span>
            <span class="row-n">{facets.types[t]}</span>
          </button>
        </li>
      {/each}
    </ul>
  </fieldset>
{/snippet}

<Page
  eyebrow="Cards"
  title="Card catalogue"
  lede="Every card wired into the rules engine, and how completely each one is implemented. Cards not listed here still work in the sandbox — you just resolve them by hand."
  width="wide"
>
  {#if loading}
    <!-- The search box is the catalog.search hint's anchor: it must exist before the data
         does (#3305). The facets need the data, so only the box shows, disabled. -->
    {#if narrow}
      <div class="bar">{@render searchBox()}</div>
    {:else}
      <div class="layout">
        <aside class="sidebar" aria-label="Filters">
          {@render searchBox()}
        </aside>
        <div class="main">
          <p class="help" aria-live="polite">Loading the catalogue…</p>
        </div>
      </div>
    {/if}
    {#if narrow}<p class="help" aria-live="polite">Loading the catalogue…</p>{/if}
  {:else if error}
    <p class="notice err" role="alert">
      <Icon name="x" size={14} />
      <span>Couldn't load the catalogue ({error}).</span>
    </p>
  {:else}
    <div class="layout">
      {#if !narrow}
        <aside class="sidebar" aria-label="Filters">
          {@render searchBox()}
          {@render facetPanel()}
        </aside>
      {/if}

      <div class="main">
        <div class="bar">
          {#if narrow}
            <button
              type="button"
              class="filters-btn"
              bind:this={filtersBtn}
              aria-haspopup="dialog"
              aria-expanded={sheetOpen}
              onclick={openSheet}
            >
              Filters{#if chips.length > 0}<span class="filters-n">{chips.length}</span>{/if}
            </button>
            {@render searchBox()}
          {/if}
          <p class="count" aria-live="polite">
            <strong>{shown.length}</strong>
            {shown.length === 1 ? "card" : "cards"}
            {#if chips.length > 0 && counts}<span class="dim">of {counts.total}</span>{/if}
          </p>
          <p class="sort">Sorted by name</p>
        </div>

        {#if chips.length > 0}
          <ul class="active" aria-label="Active filters">
            {#each chips as chip (chip.kind + ":" + chip.value)}
              <li>
                <button
                  type="button"
                  class="achip"
                  onclick={() => {
                    const f = removeChip(filter, chip);
                    query = f.query;
                    colors = f.colors;
                    types = f.types;
                    completeness = f.completeness;
                  }}
                >
                  <span class:cap={chip.kind === "types"}>{chip.label}</span>
                  <Icon name="x" size={11} />
                  <span class="sr-only">remove filter</span>
                </button>
              </li>
            {/each}
            <li>
              <button type="button" class="ghost sm" onclick={reset}>clear filters</button>
            </li>
          </ul>
        {/if}

        {#if shown.length === 0}
          <p class="help">No card matches those filters.</p>
        {:else}
          <!-- ADR 0179: one disclosure for the Buy links on every tile.
               Renders nothing with no shop configured or links hidden. -->
          <ShopNote />
          <ul class="grid">
            {#each shown as c (c.oracle_id)}
              {@const img = catalogImageURL(c, "normal")}
              {@const cav = c.caveats ?? []}
              <!-- A card with caveats is focusable so the note shows on
                   keyboard focus, and on tap (a tap focuses it). -->
              <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
              <li class="tile" tabindex={cav.length > 0 ? 0 : undefined}>
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
                  {#if cav.length > 0}
                    <ul class="caveats">
                      {#each cav as note (note)}
                        <li>{note}</li>
                      {/each}
                    </ul>
                  {/if}
                </div>
                <div class="meta">
                  <p class="tname">{c.name}</p>
                  <span class="cchip k-{c.completeness}">{COMPLETENESS_LABELS[c.completeness]}</span
                  >
                </div>
                <ShopLinks card={c.name} compact note={false} />
              </li>
            {/each}
          </ul>
        {/if}
      </div>
    </div>

    {#if narrow && sheetOpen}
      <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
      <div class="backdrop" onclick={closeSheet}></div>
      <div
        class="sheet"
        role="dialog"
        aria-modal="true"
        aria-label="Filters"
        tabindex="-1"
        bind:this={sheetEl}
        onkeydown={sheetKeydown}
      >
        <div class="sheet-head">
          <h2>Filters</h2>
          <button type="button" class="ghost sm" onclick={closeSheet}>
            <Icon name="x" size={12} /> Done
          </button>
        </div>
        <div class="sheet-body">
          {@render facetPanel()}
        </div>
      </div>
    {/if}
  {/if}
</Page>

<style>
  .help {
    margin: 0;
    color: var(--fg-muted);
    font-size: 13px;
  }
  .dim {
    color: var(--fg-dim);
  }
  .cap {
    text-transform: capitalize;
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

  /* --- layout ---------------------------------------------------- */
  .layout {
    display: grid;
    grid-template-columns: 250px minmax(0, 1fr);
    gap: 28px;
    align-items: start;
  }
  .main {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .sidebar {
    position: sticky;
    top: 16px;
    display: flex;
    flex-direction: column;
    gap: 18px;
    max-height: calc(100vh - 32px);
    overflow-y: auto;
    padding: 14px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
  }
  .search {
    width: 100%;
    box-sizing: border-box;
  }

  /* --- badges ---------------------------------------------------- */
  .cchip {
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
  .cchip.k-full {
    color: var(--mint);
    border-color: color-mix(in srgb, var(--mint) 45%, transparent);
    background: color-mix(in srgb, var(--mint) 12%, transparent);
  }
  .cchip.k-caveats {
    color: var(--accent-strong);
    border-color: color-mix(in srgb, var(--accent) 45%, transparent);
    background: var(--accent-soft);
  }
  .cchip.k-unreviewed {
    color: var(--fg-dim);
  }

  /* --- facets ---------------------------------------------------- */
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
  .rows {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .row {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 8px;
    height: 30px;
    padding: 0 8px;
    border: 1px solid transparent;
    border-radius: var(--radius);
    background: transparent;
    color: var(--fg-muted);
    font-size: 13px;
    text-align: left;
    cursor: pointer;
  }
  .row:hover {
    background: var(--surface-hover);
    color: var(--fg);
  }
  .row.on {
    background: var(--accent-soft);
    border-color: var(--accent);
    color: var(--accent-strong);
  }
  .row-name {
    flex: 1;
    min-width: 0;
  }
  .row-n {
    font-family: var(--font-mono);
    font-size: 11.5px;
    color: var(--fg-dim);
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--fg-dim);
  }
  .dot.k-full {
    background: var(--mint);
  }
  .dot.k-caveats {
    background: var(--accent);
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
  /* Colour pips: the universal WUBRG tints, as before. Colour identity
     is the one facet where colour IS the information. */
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

  /* --- bar and active chips -------------------------------------- */
  .bar {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }
  .bar .search {
    flex: 1 1 200px;
    width: auto;
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
  .sort {
    margin: 0 0 0 auto;
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .filters-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .filters-n {
    min-width: 18px;
    height: 18px;
    padding: 0 5px;
    box-sizing: border-box;
    border-radius: 999px;
    background: var(--accent);
    color: var(--bg, #000);
    font-family: var(--font-mono);
    font-size: 10.5px;
    line-height: 18px;
    text-align: center;
  }
  .active {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
  }
  .achip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 26px;
    padding: 0 9px;
    border-radius: 999px;
    border: 1px solid var(--accent);
    background: var(--accent-soft);
    color: var(--accent-strong);
    font-size: 12px;
    cursor: pointer;
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
    /* six across at the 1200px page width */
    grid-template-columns: repeat(auto-fill, minmax(128px, 1fr));
    gap: 14px;
  }
  .tile {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
    outline: none;
  }
  .art {
    /* positioned for the failed-art pip (#33) and the caveat note */
    position: relative;
    --art-error-top: 8px;
    --art-error-right: 8px;
    aspect-ratio: 488 / 680;
    border-radius: 10px;
    overflow: hidden;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
  }
  .tile:focus-visible .art {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
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
    padding: 10px;
    box-sizing: border-box;
    text-align: center;
  }
  .noart-name {
    font-family: var(--font-display);
    font-size: 13px;
    color: var(--fg);
  }
  .noart-why {
    font-family: var(--font-mono);
    font-size: 9.5px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .meta {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
    min-width: 0;
  }
  .tname {
    margin: 0;
    font-size: 12.5px;
    font-weight: 600;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  /* Caveats: laid over the foot of the art on hover, keyboard focus or
     tap. Kept in the DOM (just transparent) so a screen reader reads
     them with the card. */
  .caveats {
    position: absolute;
    inset: auto 0 0 0;
    margin: 0;
    padding: 24px 8px 8px;
    list-style: none;
    display: flex;
    flex-direction: column;
    gap: 4px;
    background: linear-gradient(
      to top,
      color-mix(in srgb, var(--surface-sunken) 96%, transparent) 70%,
      transparent
    );
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.12s ease;
  }
  .tile:hover .caveats,
  .tile:focus-within .caveats,
  .tile:focus .caveats {
    opacity: 1;
  }
  .caveats li {
    font-size: 11px;
    line-height: 1.4;
    color: var(--fg);
    border-left: 2px solid var(--accent);
    padding-left: 6px;
  }

  /* --- sheet (tablet) -------------------------------------------- */
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    z-index: 60;
  }
  .sheet {
    position: fixed;
    z-index: 61;
    top: 0;
    bottom: 0;
    left: 0;
    width: min(340px, 90vw);
    display: flex;
    flex-direction: column;
    background: var(--surface);
    border-right: 1px solid var(--border-strong);
    box-shadow: var(--shadow);
    outline: none;
  }
  .sheet-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px;
    border-bottom: 1px solid var(--border);
  }
  .sheet-head h2 {
    margin: 0;
    font-size: 16px;
    font-family: var(--font-display);
  }
  .sheet-body {
    display: flex;
    flex-direction: column;
    gap: 18px;
    padding: 16px;
    overflow-y: auto;
  }

  @media (max-width: 999px) {
    .layout {
      grid-template-columns: minmax(0, 1fr);
    }
    .grid {
      grid-template-columns: repeat(4, minmax(0, 1fr));
    }
  }
  @media (max-width: 600px) {
    .grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }
</style>
