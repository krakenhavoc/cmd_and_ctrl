<script lang="ts">
  import { FILTER_FROM, filterRows, type PickerRow } from "../deckPicker";

  // DeckRadioList is the list half of both deck pickers (YourDecksPicker,
  // PrebuiltDeckPicker): one line per deck (a coverage figure only when the deck has gaps), a height cap so a long
  // library scrolls inside the panel instead of pushing the paste box a
  // page down, and a filter box once the list is longer than the cap
  // shows. What each row says comes from deckPicker.ts; the selected
  // deck's longer story is the picker's to tell, under this list.

  interface Props {
    rows: PickerRow[];
    /** The radio group's name, which the tests and the e2e specs select on. */
    name: string;
    selected: string;
    /** The filter box's accessible name. */
    filterLabel: string;
    /** Fires when the player picks a row themselves. */
    onpick?: () => void;
  }

  let { rows, name, selected = $bindable(), filterLabel, onpick }: Props = $props();

  let query = $state("");
  let list = $state<HTMLUListElement | undefined>();

  const shown = $derived(filterRows(rows, query));

  // Keep the selected row in view inside the list — the preselected last
  // deck can sit well below the cap. Scrolls the list only, never the
  // page: scrollIntoView would yank the whole lobby on mount.
  $effect(() => {
    if (!list || !selected) return;
    void shown;
    const row = [...list.querySelectorAll<HTMLElement>("[data-id]")].find(
      (el) => el.dataset.id === selected,
    );
    if (!row) return;
    // The list is position: relative, so offsetTop is from its top.
    const top = row.offsetTop;
    const bottom = top + row.offsetHeight;
    if (top < list.scrollTop) list.scrollTop = top;
    else if (bottom > list.scrollTop + list.clientHeight)
      list.scrollTop = bottom - list.clientHeight;
  });
</script>

{#if rows.length > FILTER_FROM}
  <input
    class="filter"
    type="search"
    placeholder={`filter ${rows.length} decks — name, commander`}
    aria-label={filterLabel}
    bind:value={query}
  />
{/if}

<ul class="rows" bind:this={list}>
  {#each shown as row (row.id)}
    <li data-id={row.id}>
      <label class="row" class:sel={selected === row.id} title={row.sub}>
        <input type="radio" {name} value={row.id} bind:group={selected} onchange={onpick} />
        <span class="name">{row.name}</span>
        {#if row.colors.length > 0}
          <span class="pips">
            {#each row.colors as c (c)}
              <i class="pip pip-{c}">{c}</i>
            {/each}
          </span>
        {/if}
        <span class="sub">{row.sub}</span>
        {#if row.cov && row.tone !== "full"}
          <span class="cov"><i class="dot dot-{row.tone}"></i>{row.cov}</span>
        {/if}
      </label>
    </li>
  {:else}
    <li class="none">No deck matches “{query.trim()}”.</li>
  {/each}
</ul>

<style>
  .filter {
    display: block;
    width: 100%;
    box-sizing: border-box;
    margin: 0 0 8px;
    padding: 6px 10px;
    border: 1px solid var(--border-strong);
    border-radius: 8px;
    background: var(--surface-sunken);
    color: var(--fg);
    font: inherit;
    font-size: 12.5px;
  }
  .filter:focus {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-soft);
  }
  .filter::placeholder {
    color: var(--fg-dim);
  }
  /* About seven rows, then scroll. The half row peeking out at the
     bottom is the cue that there is more. */
  .rows {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 4px;
    max-height: 17.5rem;
    overflow-y: auto;
    overscroll-behavior: contain;
    position: relative;
  }
  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    column-gap: 8px;
    row-gap: 2px;
    padding: 6px 10px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: var(--surface-sunken);
    cursor: pointer;
  }
  .row:hover {
    border-color: var(--border-strong);
  }
  .row.sel {
    border-color: var(--accent);
    background: var(--accent-soft);
  }
  .row input {
    margin: 0;
    flex: 0 0 auto;
  }
  .name {
    font-weight: 600;
    font-size: 13px;
    color: var(--fg);
    white-space: nowrap;
  }
  .pips {
    display: inline-flex;
    gap: 3px;
  }
  /* Same tints and the same "colour on the text, not the fill"
     treatment as the catalogue page's colour filter, so a player who
     has seen one recognises the other. */
  .pip {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    border: 1px solid currentColor;
    background: color-mix(in srgb, currentColor 16%, transparent);
    font-size: 8px;
    font-style: normal;
    font-weight: 700;
  }
  .pip-W {
    color: #e8dcc0;
  }
  .pip-U {
    color: #7fb4e8;
  }
  .pip-B {
    color: #b09ac4;
  }
  .pip-R {
    color: #e88b7f;
  }
  .pip-G {
    color: #7fc79a;
  }
  /* Takes the slack, and ellipsises before it pushes the coverage
     figure onto a second line; on a phone it wraps under the name. */
  .sub {
    flex: 1 1 8rem;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 11.5px;
    color: var(--fg-dim);
  }
  .cov {
    margin-left: auto;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    color: var(--fg-muted);
    white-space: nowrap;
    font-variant-numeric: tabular-nums;
  }
  .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--accent);
    flex: 0 0 auto;
  }
  /* Green only for a deck that really does play as printed, so the
     colour keeps meaning something. */
  .dot-full {
    background: var(--mint);
  }
  .dot-gap {
    background: var(--danger);
  }
  .none {
    padding: 8px 10px;
    font-size: 12px;
    color: var(--fg-dim);
  }
</style>
