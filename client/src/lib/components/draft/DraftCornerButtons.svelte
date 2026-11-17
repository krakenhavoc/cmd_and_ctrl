<script lang="ts">
  // DraftCornerButtons — the draft room's buttons in the bottom-right
  // corner (#5848 §5, §8): today Deck view, which swaps the pack for the
  // deckbuilding layout of the picks so far and back (key D).
  //
  // They float over the page, so they are half transparent with a blurred
  // backdrop and fully opaque on hover or focus; the room pads its foot
  // so the last row can scroll clear of them.
  //
  // The group is a row that grows to the left: another button (Full
  // stats, key S) goes in as `children`, in a button of class "fab", and
  // takes the same look. Its own key belongs to whoever adds it.
  import type { Snippet } from "svelte";

  import { isTypingTarget } from "../../shortcuts";
  import { L } from "../../labels";
  import Icon from "../Icon.svelte";

  interface Props {
    /** The picks so far, for the Deck view button. */
    picks: number;
    /** Deck view is showing: the button says how to go back. */
    deckView: boolean;
    /** Deck view was asked for, or asked to end. */
    ondeck: () => void;
    /** More buttons, drawn before Deck view. Give each the class "fab". */
    children?: Snippet;
  }
  const { picks, deckView, ondeck, children }: Props = $props();

  // D, unless the viewer is typing, holds a modifier or the page
  // already used the key.
  function onkeydown(ev: KeyboardEvent): void {
    if (ev.defaultPrevented || ev.repeat || ev.ctrlKey || ev.metaKey || ev.altKey) return;
    if (ev.key !== "d" && ev.key !== "D") return;
    if (isTypingTarget(ev.target as Element | null)) return;
    ev.preventDefault();
    ondeck();
  }
</script>

<svelte:window {onkeydown} />

<div class="fabs" data-testid="draft-corner-buttons">
  {@render children?.()}
  <button
    type="button"
    class="fab"
    class:back={deckView}
    aria-keyshortcuts="D"
    data-testid="draft-deck-view-button"
    onclick={ondeck}
  >
    {#if deckView}
      <Icon name="chevronLeft" size={14} />
      {L.draftBackToPack}
    {:else}
      <Icon name="list" size={14} />
      {L.draftDeckView}
      <span class="n">{picks === 1 ? "1 pick" : `${picks} picks`}</span>
    {/if}
    <kbd>D</kbd>
  </button>
</div>

<style>
  .fabs {
    position: fixed;
    right: max(16px, calc(env(safe-area-inset-right, 0px) + 16px));
    bottom: max(16px, calc(env(safe-area-inset-bottom, 0px) + 16px));
    z-index: 40;
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 8px;
  }
  .fabs :global(.fab) {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 10px 16px;
    border-radius: 999px;
    border: 1px solid var(--accent);
    background: var(--accent);
    color: var(--accent-fg);
    font: inherit;
    font-weight: 600;
    cursor: pointer;
    box-shadow: 0 8px 22px rgba(0, 0, 0, 0.5);
    /* Over the cards, so half there: the page can still be read through it. */
    opacity: 0.5;
    -webkit-backdrop-filter: blur(6px);
    backdrop-filter: blur(6px);
    transition: opacity 0.15s ease;
  }
  .fabs :global(.fab:hover),
  .fabs :global(.fab:focus-visible) {
    opacity: 1;
  }
  .fabs :global(.fab:focus-visible) {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }
  .fabs :global(.fab.back) {
    background: var(--surface-raised);
    color: var(--fg);
    border-color: var(--border-strong);
  }
  .fabs :global(.fab .n) {
    font-family: var(--font-mono);
    font-size: 12px;
    opacity: 0.8;
  }
  .fabs :global(.fab kbd) {
    padding: 0 6px;
    border-radius: 4px;
    border: 1px solid currentColor;
    font-family: var(--font-mono);
    font-size: 11px;
    opacity: 0.8;
  }
  @media (prefers-reduced-motion: reduce) {
    .fabs :global(.fab) {
      transition: none;
    }
  }
</style>
