<script lang="ts">
  import { filterTiles, type DeckTile } from "../deckPicker";
  import { auraGradient } from "../deckGrid";
  import { inView } from "../inView";
  import CommanderArt from "./decks/CommanderArt.svelte";
  import Icon from "./Icon.svelte";
  import IdentityPips from "./IdentityPips.svelte";

  // DeckGrid — the deck picker's grid of commander art (ADR 0212 §3), and
  // the Add bot picker's (owner decision 8). A radio group of tiles, as
  // DeckRadioList's rows were: the radios are real, so the arrow keys
  // move the pick and Space takes it, and the e2e specs select on the
  // group's name. A tile is the commander's art crop with its artist's
  // credit (partners side by side), or a wash in the deck's colours; the
  // deck's name, the line under it (deckPicker.ts), the pips, and a flag
  // only when the deck has cards that do not play as printed.
  //
  // The grid scrolls inside a capped height. A tile mounts its art only
  // once it scrolls near the view (lib/inView.ts), so a library of 200
  // decks fetches what is seen. What the selected deck's longer story is
  // belongs to the picker, under the grid.

  interface Props {
    tiles: DeckTile[];
    /** The radios' name, which the tests and the e2e specs select on. */
    name: string;
    /** The radio group's accessible name. */
    label: string;
    selected: string;
    /** The filter box's text; the box itself is the picker's, in its tab bar. */
    query?: string;
    /** Fires when the player picks a tile themselves. */
    onpick?: () => void;
  }

  let { tiles, name, label, selected = $bindable(), query = "", onpick }: Props = $props();

  let scroller = $state<HTMLDivElement | undefined>();
  const shown = $derived(filterTiles(tiles, query));

  // Tiles whose art may load: near the view once, then for good.
  let seen = $state<Record<string, boolean>>({});
  // Crops that failed to load, per printing; a tile with none left washes.
  let failed = $state<string[]>([]);

  // Keep the selected tile in view inside the grid: the preselected last
  // deck can sit well below the cap. Scrolls the grid only, never the
  // page, which scrollIntoView would yank on mount.
  $effect(() => {
    if (!scroller || !selected) return;
    void shown;
    const tile = [...scroller.querySelectorAll<HTMLElement>("[data-id]")].find(
      (el) => el.dataset.id === selected,
    );
    if (!tile) return;
    const top = tile.offsetTop;
    const bottom = top + tile.offsetHeight;
    if (top < scroller.scrollTop) scroller.scrollTop = top;
    else if (bottom > scroller.scrollTop + scroller.clientHeight)
      scroller.scrollTop = bottom - scroller.clientHeight;
  });
</script>

<div class="scroll" bind:this={scroller} data-scroll-root>
  <div class="grid" role="radiogroup" aria-label={label}>
    {#each shown as t (t.id)}
      {@const crops = seen[t.id] ? t.art.filter((a) => !failed.includes(a.scryfallId)) : []}
      {@const wash = auraGradient(t.identity)}
      <label class="dtile" class:sel={selected === t.id} data-id={t.id}>
        <input
          class="radio"
          type="radio"
          {name}
          value={t.id}
          title={t.flagTitle || undefined}
          bind:group={selected}
          onchange={onpick}
        />
        <span
          class="art"
          class:wash={crops.length === 0 && wash !== ""}
          style:--wash={wash || null}
          use:inView={{ onenter: () => (seen = { ...seen, [t.id]: true }) }}
        >
          {#each crops as a (a.scryfallId)}
            <CommanderArt
              scryfallId={a.scryfallId}
              name={a.name}
              onfail={() => (failed = [...failed, a.scryfallId])}
            />
          {/each}
          {#if selected === t.id}
            <span class="selmark" aria-hidden="true"
              ><Icon name="check" size={12} strokeWidth={2.6} /></span
            >
          {/if}
        </span>
        <span class="dinfo">
          <span class="dname">{t.name}</span>
          <span class="dsub">{t.sub}</span>
          <span class="drow">
            <IdentityPips colors={t.colors} />
            {#if t.flag}
              <span class="gapflag" title={t.flagTitle}><Icon name="warn" size={11} />{t.flag}</span
              >
            {/if}
          </span>
        </span>
      </label>
    {:else}
      <p class="none">No deck matches “{query.trim()}”.</p>
    {/each}
  </div>
</div>

<style>
  /* About two rows at tablet width, then the grid scrolls inside the
     picker. */
  .scroll {
    position: relative;
    max-height: 20rem;
    overflow-y: auto;
    overscroll-behavior: contain;
    /* Room for the selected tile's ring. */
    padding: 4px;
    margin: -4px;
  }
  /* About 150px tiles: four columns at 820px, seven at 1280px. */
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 10px;
  }
  .dtile {
    position: relative;
    display: flex;
    flex-direction: column;
    min-width: 0;
    overflow: hidden;
    border-radius: var(--radius);
    border: 1px solid var(--border);
    background: var(--surface);
    color: var(--fg);
    cursor: pointer;
    transition:
      border-color 120ms var(--ease),
      box-shadow 120ms var(--ease);
  }
  .dtile:hover {
    border-color: var(--border-strong);
  }
  .dtile.sel {
    border-color: var(--accent);
    box-shadow:
      0 0 0 1px var(--accent),
      0 0 0 4px var(--accent-soft);
  }
  /* The radio is real but unseen, laid over the whole tile so a press
     anywhere on it lands on the radio; its focus shows on the tile. */
  .radio {
    position: absolute;
    inset: 0;
    z-index: 2;
    width: 100%;
    height: 100%;
    margin: 0;
    opacity: 0;
    cursor: pointer;
  }
  .dtile:has(.radio:focus-visible) {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }

  .art {
    position: relative;
    display: flex;
    gap: 2px;
    aspect-ratio: 16 / 9;
    overflow: hidden;
    background: var(--surface-sunken);
  }
  .art::after {
    content: "";
    position: absolute;
    inset: 0;
    background: linear-gradient(to bottom, transparent 55%, var(--surface));
    pointer-events: none;
  }
  /* No art (a pre-built deck until phase B1, or a crop that failed): the
     deck's colours as a wash, never a broken image. */
  .art.wash {
    background:
      linear-gradient(
        color-mix(in srgb, var(--surface) 55%, transparent),
        color-mix(in srgb, var(--surface) 55%, transparent)
      ),
      var(--wash);
  }
  .selmark {
    position: absolute;
    top: 6px;
    left: 6px;
    z-index: 1;
    display: grid;
    place-items: center;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--accent);
    color: var(--accent-fg);
  }

  .dinfo {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
    padding: 6px 10px 9px;
  }
  .dname,
  .dsub {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .dname {
    font-size: 13px;
    font-weight: 700;
  }
  .dsub {
    font-size: 11.5px;
    color: var(--fg-muted);
  }
  .drow {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 4px 6px;
    min-height: 18px;
    margin-top: 2px;
  }
  /* Neutral chrome, not a signal colour: --priority and --danger mean
     other things (ADR 0212 §3). The glyph and the words carry it. */
  .gapflag {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 6px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    color: var(--fg);
    font: 600 10px/1.2 var(--font-mono);
    white-space: nowrap;
  }
  .none {
    grid-column: 1 / -1;
    margin: 0;
    padding: 8px 2px;
    font-size: 12px;
    color: var(--fg-muted);
  }
</style>
