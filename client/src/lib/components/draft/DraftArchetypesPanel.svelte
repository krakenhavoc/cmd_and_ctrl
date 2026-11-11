<script lang="ts">
  // DraftArchetypesPanel — a set's colour pairs as a collapsible panel
  // for the draft room (ADR 0222 §7, #5703): a heading that opens and
  // closes, over the pair list. The room puts it to the left of the pack
  // on a wide screen and under it on a narrow one; that is the room's
  // layout, not this panel's.
  //
  // It starts open and keeps its own state. The pairs are fetched when
  // it is first open, so a viewer who keeps it closed never asks.
  import { L } from "../../labels";
  import DraftArchetypes from "./DraftArchetypes.svelte";

  interface Props {
    /** The set code being drafted. */
    set: string;
    /** The set's name, for the panel's description. */
    setName?: string;
    open?: boolean;
  }
  const { set, setName = "", open = true }: Props = $props();

  // The panel opens as asked and then follows the viewer; the prop is
  // only its starting state.
  // svelte-ignore state_referenced_locally
  let shown = $state(open);
</script>

<details
  class="archetypes-panel"
  data-testid="archetypes-panel"
  open={shown}
  ontoggle={(e) => (shown = e.currentTarget.open)}
>
  <summary>
    <h2>{L.draftArchetypesTitle}</h2>
    {#if setName}<span class="set">{setName}</span>{/if}
  </summary>
  {#if shown}
    <div class="body"><DraftArchetypes {set} /></div>
  {/if}
</details>

<style>
  .archetypes-panel {
    min-width: 0;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--surface);
  }
  summary {
    display: flex;
    align-items: baseline;
    gap: 4px 10px;
    flex-wrap: wrap;
    padding: 10px 12px;
    cursor: pointer;
    list-style: none;
    border-radius: var(--radius);
  }
  summary::-webkit-details-marker {
    display: none;
  }
  summary::before {
    content: "";
    align-self: center;
    width: 0;
    height: 0;
    border-left: 5px solid var(--fg-muted);
    border-top: 4px solid transparent;
    border-bottom: 4px solid transparent;
  }
  details[open] > summary::before {
    transform: rotate(90deg);
  }
  summary:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: -2px;
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
  .body {
    padding: 0 8px 8px;
  }
</style>
