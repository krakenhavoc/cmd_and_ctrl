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
  //
  // Paging (#6120): the catalogue holds tens of thousands of cards, and
  // drawing every match as a tile was what made the page slow to open.
  // The filters, the search and the facet counts still run over the
  // whole list; only the current page's tiles are drawn, so only their
  // art loads. The search, the page and the page size live in the URL
  // (#/catalog?q=…&page=3&size=100): the pager's links and the size
  // selector push a history entry, so back and forward step through
  // pages, and typing or ticking a filter rewrites the entry in place
  // (back to page 1) rather than leaving one per keystroke.

  import { onMount, tick, untrack } from "svelte";
  import Icon from "../lib/components/Icon.svelte";
  import Page from "../lib/components/Page.svelte";
  import ShopLinks from "../lib/components/ShopLinks.svelte";
  import ShopNote from "../lib/components/ShopNote.svelte";
  import SetPicker from "../lib/components/SetPicker.svelte";
  import { cardArt } from "../lib/cardArt";
  import { navigate, route } from "../lib/router";
  import { L } from "../lib/labels";
  import {
    fetchCatalog,
    filterCatalog,
    facetCounts,
    activeChips,
    removeChip,
    toggle,
    findSet,
    setSummary,
    catalogImageURL,
    catalogHash,
    clampPage,
    pageCount,
    pageRange,
    pageSlice,
    pagerItems,
    parsePageSize,
    DEFAULT_PAGE_SIZE,
    PAGE_SIZES,
    COLOR_FILTERS,
    COLOR_LABELS,
    TYPE_FILTERS,
    COMPLETENESS_LABELS,
    COMPLETENESS_BLURBS,
    type CatalogEntry,
    type CatalogCounts,
    type CatalogSet,
    type Completeness,
    type ColorFilter,
    type PageSize,
    type TypeFilter,
  } from "../lib/catalog";

  let cards = $state<CatalogEntry[]>([]);
  let counts = $state<CatalogCounts | null>(null);
  // The set picker's data, newest set first; empty when the server has no
  // card data, in which case the picker is not drawn (#4790).
  let sets = $state<CatalogSet[]>([]);
  let loading = $state(true);
  let error = $state("");

  let query = $state("");
  let colors = $state<ColorFilter[]>([]);
  let types = $state<TypeFilter[]>([]);
  let completeness = $state<Completeness[]>([]);
  let setCode = $state("");

  const buckets: Completeness[] = ["full", "caveats", "unreviewed"];

  const filter = $derived({ query, colors, types, completeness, set: setCode });
  const shown = $derived(filterCatalog(cards, filter));
  const facets = $derived(facetCounts(cards, filter));
  const chips = $derived(activeChips(filter, sets));
  const pickedSet = $derived(setCode ? findSet(sets, setCode) : undefined);

  // --- paging (#6120) ----------------------------------------------
  let pageSize = $state<PageSize>(DEFAULT_PAGE_SIZE);
  // The page asked for, and the filter it was asked under: a page only
  // means something for the filter that made it, so a changed filter,
  // search or page size is page 1.
  let pageAt = $state({ page: 1, key: "" });
  const filterKey = $derived(JSON.stringify(filter));
  const page = $derived(
    pageAt.key === filterKey ? clampPage(pageAt.page, shown.length, pageSize) : 1,
  );
  const pages = $derived(pageCount(shown.length, pageSize));
  const visible = $derived(pageSlice(shown, page, pageSize));
  const range = $derived(pageRange(page, shown.length, pageSize));
  const pager = $derived(pagerItems(page, pages));
  const fmt = new Intl.NumberFormat("en");

  const hrefFor = (p: number, size: number = pageSize) => catalogHash({ query, page: p, size });

  // The URL is the record. #/catalog?q=<text> pre-fills the search (the
  // roadmap links its example card names here), and ?page= and ?size=
  // pick the page. Re-applied whenever the hash changes while the page
  // is open: a second roadmap link, a pager link, back and forward.
  $effect(() => {
    const r = $route;
    if (r.name !== "catalog") return;
    untrack(() => {
      query = r.query ?? "";
      pageSize = parsePageSize(r.size === undefined ? null : String(r.size));
      pageAt = { page: r.page ?? 1, key: JSON.stringify(filter) };
    });
  });

  // Keep the URL in step with what is on screen without adding history:
  // a filter change drops back to page 1, and a link past the last page
  // lands on the last page, so the address says so too. Not while the
  // data is loading, when every page clamps to 1.
  $effect(() => {
    if (loading) return;
    if (pageAt.key !== filterKey || pageAt.page !== page) {
      pageAt = { page, key: filterKey };
    }
    const want = hrefFor(page);
    if (!location.hash.startsWith("#/catalog")) return;
    if (decodeURIComponent(location.hash) !== decodeURIComponent(want)) {
      history.replaceState(history.state, "", want);
    }
  });

  function pickSize(ev: Event) {
    const size = parsePageSize((ev.currentTarget as HTMLSelectElement).value);
    navigate(hrefFor(1, size));
  }

  // After a pager link: put the top of the results back in view, and if
  // the link that had focus is gone (Next on the last page), hand focus
  // to the current page's number so the keyboard is not dropped.
  let barEl = $state<HTMLElement | null>(null);
  let pagerEl = $state<HTMLElement | null>(null);
  let fromPager = false;
  function pagerClick() {
    fromPager = true;
  }
  $effect(() => {
    void page;
    if (!fromPager) return;
    fromPager = false;
    // The page can change inside another effect's run (the route's), before
    // the pager has re-rendered; wait for it, or the old page's number is
    // the one still marked current.
    void tick().then(() => {
      if (barEl && barEl.getBoundingClientRect().top < 0) barEl.scrollIntoView({ block: "start" });
      if (pagerEl && !pagerEl.contains(document.activeElement)) {
        pagerEl.querySelector<HTMLElement>('[aria-current="page"]')?.focus();
      }
    });
  });

  function reset() {
    query = "";
    colors = [];
    types = [];
    completeness = [];
    setCode = "";
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
    sheetEl?.querySelector<HTMLElement>("button, input, select")?.focus();
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
    const els = [...sheetEl.querySelectorAll<HTMLElement>("button, input, select, [href]")].filter(
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
        sets = res.sets ?? [];
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
  {#if sets.length > 0}
    <fieldset class="facet">
      <legend>Set</legend>
      <SetPicker {sets} bind:value={setCode} />
    </fieldset>
  {/if}

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
  {#if !loading && error}
    <p class="notice err" role="alert">
      <Icon name="x" size={14} />
      <span>Couldn't load the catalogue ({error}).</span>
    </p>
  {:else}
    <!-- The search box is the catalog.search hint's anchor: it must exist before the data
         does (#3305), and it must be the SAME element before and after, or the tip's anchor
         is torn out from under it when the data lands (#3343: the hints e2e saw "Element is
         not attached"). So the layout, the box and the bar it sits in are one block for the
         whole load; the facets, the count and the grid come and go inside it, and while
         loading the box is disabled. -->
    <div class="layout">
      {#if !narrow}
        <aside class="sidebar" aria-label="Filters">
          {@render searchBox()}
          {#if !loading}{@render facetPanel()}{/if}
        </aside>
      {/if}

      <div class="main">
        <div class="bar" bind:this={barEl}>
          {#if narrow}
            {#if !loading}
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
            {/if}
            {@render searchBox()}
          {/if}
          {#if !loading}
            <p class="count" aria-live="polite">
              {#if pages > 1}
                Showing <strong>{fmt.format(range.from)}–{fmt.format(range.to)}</strong> of
                <strong>{fmt.format(shown.length)}</strong> cards
              {:else}
                <strong>{fmt.format(shown.length)}</strong>
                {shown.length === 1 ? "card" : "cards"}
              {/if}
              {#if chips.length > 0 && counts}<span class="dim">of {fmt.format(counts.total)}</span
                >{/if}
            </p>
            <p class="sort">Sorted by name</p>
            <label class="size">
              <span>Per page</span>
              <select aria-label={L.cataloguePageSize} value={pageSize} onchange={pickSize}>
                {#each PAGE_SIZES as n (n)}
                  <option value={n}>{n}</option>
                {/each}
              </select>
            </label>
          {/if}
        </div>

        {#if loading}
          <p class="help" aria-live="polite">Loading the catalogue…</p>
        {/if}

        {#if !loading && chips.length > 0}
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
                    setCode = f.set;
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

        {#if !loading && pickedSet}
          <!-- The picked set's answer: how much of it plays, in words and in a bar. The
               grid below lists only the cards that do, so the unsupported count is the
               one place the cards that are missing show up. -->
          <div class="setline">
            <p class="setline-text">{setSummary(pickedSet)}</p>
            {#if pickedSet.total > 0}
              <div class="setbar" aria-hidden="true">
                <span class="seg k-full" style:flex-grow={pickedSet.full}></span>
                <span class="seg k-caveats" style:flex-grow={pickedSet.caveats}></span>
                <span class="seg k-unreviewed" style:flex-grow={pickedSet.unreviewed}></span>
                <span class="seg k-none" style:flex-grow={pickedSet.unsupported}></span>
              </div>
            {/if}
          </div>
        {/if}

        {#if loading}
          <!-- the line above is the whole body until the data lands -->
        {:else if shown.length === 0}
          <p class="help">No card matches those filters.</p>
        {:else}
          <!-- ADR 0179: one disclosure for the Buy links on every tile.
               Renders nothing with no shop configured or links hidden. -->
          <ShopNote />
          <ul class="grid">
            {#each visible as c (c.oracle_id)}
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
          {#if pages > 1}
            <nav class="pager" aria-label={L.cataloguePager} bind:this={pagerEl}>
              {#if page > 1}
                <a
                  class="pg step"
                  href={hrefFor(page - 1)}
                  rel="prev"
                  aria-label={L.cataloguePrevPage}
                  onclick={pagerClick}
                  ><span aria-hidden="true">‹</span><span class="word"> Previous</span></a
                >
              {:else}
                <span class="pg step off" aria-hidden="true"
                  >‹<span class="word"> Previous</span></span
                >
              {/if}
              <ul class="pages">
                {#each pager as item, i (item === "gap" ? `gap${i}` : item)}
                  <li>
                    {#if item === "gap"}
                      <span class="gap" aria-hidden="true">…</span>
                    {:else}
                      <a
                        class="pg"
                        class:on={item === page}
                        href={hrefFor(item)}
                        aria-label={L.cataloguePage(item)}
                        aria-current={item === page ? "page" : undefined}
                        onclick={pagerClick}>{fmt.format(item)}</a
                      >
                    {/if}
                  </li>
                {/each}
              </ul>
              {#if page < pages}
                <a
                  class="pg step"
                  href={hrefFor(page + 1)}
                  rel="next"
                  aria-label={L.catalogueNextPage}
                  onclick={pagerClick}
                  ><span class="word">Next </span><span aria-hidden="true">›</span></a
                >
              {:else}
                <span class="pg step off" aria-hidden="true"><span class="word">Next </span>›</span>
              {/if}
            </nav>
          {/if}
        {/if}
      </div>
    </div>

    {#if !loading && narrow && sheetOpen}
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
  .setline {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 10px 12px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
  }
  .setline-text {
    margin: 0;
    font-size: 13px;
    color: var(--fg);
  }
  .setbar {
    display: flex;
    height: 6px;
    gap: 2px;
    border-radius: 3px;
    overflow: hidden;
  }
  .seg {
    flex-basis: 0;
    flex-shrink: 0;
    min-width: 0;
  }
  .seg.k-full {
    background: var(--mint);
  }
  .seg.k-caveats {
    background: var(--accent);
  }
  .seg.k-unreviewed {
    background: var(--fg-dim);
  }
  .seg.k-none {
    background: var(--surface-sunken);
    box-shadow: inset 0 0 0 1px var(--border-strong);
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

  .size {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: var(--fg-muted);
  }

  /* --- pager (#6120) --------------------------------------------- */
  .pager {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-wrap: wrap;
    gap: 6px;
    margin: 0 0 40px;
  }
  .pages {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 4px;
  }
  .pg {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    min-width: 32px;
    height: 32px;
    padding: 0 8px;
    box-sizing: border-box;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--surface);
    color: var(--fg-muted);
    font-family: var(--font-mono);
    font-size: 12px;
    text-decoration: none;
  }
  a.pg:hover {
    background: var(--surface-hover);
    color: var(--fg);
  }
  a.pg:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }
  .pg.on {
    background: var(--accent-soft);
    border-color: var(--accent);
    color: var(--accent-strong);
    font-weight: 600;
  }
  .pg.step {
    font-family: inherit;
    font-size: 12.5px;
    padding: 0 12px;
  }
  .pg.off {
    opacity: 0.45;
  }
  .gap {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 20px;
    height: 32px;
    color: var(--fg-dim);
  }

  /* --- grid ------------------------------------------------------ */
  .grid {
    list-style: none;
    margin: 0 0 20px;
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
    transition:
      transform 180ms var(--ease),
      box-shadow 180ms var(--ease);
  }
  /* A card answers the pointer the way one in hand does: it lifts and
     tilts a little toward you. Pointer devices only; the global
     reduced-motion rule stills it. */
  @media (hover: hover) {
    .tile:hover .art {
      transform: translateY(-4px) rotate(-1deg);
      box-shadow: 0 18px 36px -16px var(--shadow-ink);
    }
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
    /* The pager keeps to one row on a phone: Previous and Next drop to
       their arrows (their accessible names are unchanged). */
    .pg.step {
      padding: 0 8px;
    }
    .pg.step .word {
      display: none;
    }
  }
</style>
