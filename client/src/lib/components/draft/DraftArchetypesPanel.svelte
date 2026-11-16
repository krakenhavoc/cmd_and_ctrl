<script lang="ts">
  // DraftArchetypesPanel — a set's colour pairs as a collapsible panel
  // for the draft room (ADR 0222 §7, #5703): a heading with a collapse
  // button, over the pair list. The room puts it to the left of the pack
  // on a wide screen and under everything on a narrow one; that is the
  // room's layout, not this panel's.
  //
  // Collapsed (#5848 §4) it is a narrow vertical strip on a wide screen,
  // not a hidden body in a column that stays 240px wide: the room's grid
  // gives the column back to the pack. The strip has the label turned on
  // its side, an expand button, and a pip for each colour the viewer has
  // three or more Main picks in, so the hint survives while it is shut.
  // On a narrow screen collapsed is one short bar.
  //
  // It starts open and remembers the viewer's choice in this browser
  // (draftArchetypes.ts, guarded). The pairs are fetched when it is
  // first open, so a viewer who keeps it closed never asks. With `main`
  // each pair says how many of the viewer's Main picks fit it.
  import type { DraftCard } from "../../draft";
  import {
    loadArchetypesOpen,
    pairName,
    railColours,
    saveArchetypesOpen,
  } from "../../draftArchetypes";
  import { L } from "../../labels";
  import ManaSymbol from "../board/ManaSymbol.svelte";
  import Icon from "../Icon.svelte";
  import DraftArchetypes from "./DraftArchetypes.svelte";

  interface Props {
    /** The set code being drafted. */
    set: string;
    /** The set's name, for the panel's description. */
    setName?: string;
    /**
     * Open or closed. The room binds it, to size its column; unbound it is
     * the panel's own state, starting as the viewer remembered it.
     */
    open?: boolean;
    /** The viewer's Main picks, for the fit counts and the strip's pips. */
    main?: DraftCard[];
  }
  let { set, setName = "", open = $bindable(loadArchetypesOpen()), main }: Props = $props();

  const pips = $derived(railColours(main ?? []));

  function toggle(next: boolean): void {
    open = next;
    saveArchetypesOpen(next);
  }
</script>

<section
  class="archetypes-panel"
  class:collapsed={!open}
  data-testid="archetypes-panel"
  data-open={open}
  aria-label={open ? undefined : L.draftArchetypesTitle}
  aria-labelledby={open ? "draft-archetypes-h" : undefined}
>
  {#if open}
    <header class="head">
      <div class="title">
        <h2 id="draft-archetypes-h">{L.draftArchetypesTitle}</h2>
        {#if setName}<span class="set">{setName}</span>{/if}
      </div>
      <button
        type="button"
        class="iconbtn"
        aria-label={L.draftArchetypesHide}
        aria-expanded="true"
        aria-controls="draft-archetypes-body"
        title={L.draftArchetypesHide}
        onclick={() => toggle(false)}
      >
        <Icon name="chevronLeft" size={14} />
      </button>
    </header>
    <div class="body" id="draft-archetypes-body"><DraftArchetypes {set} {main} /></div>
  {:else}
    <button
      type="button"
      class="rail"
      aria-label={L.draftArchetypesShow}
      aria-expanded="false"
      title={L.draftArchetypesShow}
      onclick={() => toggle(true)}
    >
      <span class="expand" aria-hidden="true"><Icon name="chevronRight" size={14} /></span>
      <span class="vlabel" aria-hidden="true">{L.draftArchetypesTitle}</span>
      {#if pips.length > 0}
        <span class="dots" data-testid="archetypes-rail-pips">
          {#each pips as k (k)}
            <ManaSymbol symbol={k} size={16} label={pairName(k)} />
          {/each}
        </span>
      {/if}
    </button>
  {/if}
</section>

<style>
  .archetypes-panel {
    min-width: 0;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--surface);
  }
  .head {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 10px 10px 12px;
  }
  .title {
    display: flex;
    align-items: baseline;
    flex-wrap: wrap;
    gap: 4px 10px;
    flex: 1;
    min-width: 0;
  }
  h2 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 15px;
    font-weight: 800;
    text-transform: none;
    color: var(--fg);
  }
  .set {
    font-size: 12px;
    color: var(--fg-muted);
  }
  .iconbtn {
    flex: none;
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    padding: 0;
    border-radius: 6px;
    border: 1px solid var(--border-strong);
    background: var(--surface-raised);
    color: var(--fg-muted);
    cursor: pointer;
  }
  .iconbtn:hover {
    color: var(--fg);
    background: var(--surface-hover);
  }
  .iconbtn:focus-visible,
  .rail:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }
  .body {
    padding: 0 8px 8px;
  }

  /* Collapsed, on a wide screen: a 44px strip with the label on its side. */
  .archetypes-panel.collapsed {
    width: 44px;
  }
  .rail {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    width: 100%;
    height: min(420px, calc(100vh - 140px));
    padding: 8px 0;
    border: 0;
    border-radius: var(--radius);
    background: none;
    font: inherit;
    color: var(--fg-muted);
    cursor: pointer;
  }
  .rail:hover {
    color: var(--fg);
    background: var(--surface-hover);
  }
  .expand {
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    border-radius: 6px;
    border: 1px solid var(--border-strong);
    background: var(--surface-raised);
  }
  .vlabel {
    writing-mode: vertical-rl;
    transform: rotate(180deg);
    font-family: var(--font-display);
    font-size: 14px;
    font-weight: 800;
    letter-spacing: 0.01em;
    white-space: nowrap;
  }
  .dots {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  /* Collapsed, on a narrow screen: one short bar, the label upright. */
  @media (max-width: 1099px) {
    .archetypes-panel.collapsed {
      width: auto;
    }
    .rail {
      flex-direction: row;
      height: auto;
      padding: 8px 12px;
    }
    .vlabel {
      writing-mode: horizontal-tb;
      transform: none;
    }
    .dots {
      flex-direction: row;
    }
  }
</style>
