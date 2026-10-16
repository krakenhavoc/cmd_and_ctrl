<script lang="ts">
  // BugReportDraftPill — the minimised bug report (#4029).
  //
  // Shown by whichever screen opens the report (the site header, the
  // table's command bar) whenever a draft exists and its modal is not
  // open. It sits in that screen's own top-right control cluster rather
  // than floating over the page: the table has no free corner (hand
  // below, dock bottom-right, stack and commanders left, hover zoom
  // top-right), and the command bar's icon row is the one strip that is
  // never covering a card. It is a plain button, so the page under it
  // stays fully usable and nothing here holds a modal layer or traps
  // focus.
  //
  // Below 600px it drops its words and keeps the icons, so it costs
  // little of a phone's row.
  import Icon from "./Icon.svelte";

  interface Props {
    onrestore: () => void;
  }
  const { onrestore }: Props = $props();

  let el = $state<HTMLButtonElement | null>(null);

  // The host calls this after minimising, so focus lands on the pill
  // that replaced the form instead of falling to <body>.
  export function focus(): void {
    el?.focus();
  }
</script>

<button
  bind:this={el}
  type="button"
  class="draft-pill"
  aria-label="Restore the bug report draft"
  title="Restore your bug report draft"
  onclick={onrestore}
>
  <Icon name="bug" size={15} />
  <span class="words">Bug report draft</span>
  <Icon name="chevron-up" size={14} />
</button>

<style>
  .draft-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 28px;
    padding: 0 10px;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--accent) 50%, transparent);
    background: color-mix(in srgb, var(--accent) 14%, transparent);
    color: var(--accent);
    font-size: 12px;
    font-weight: 600;
    white-space: nowrap;
    cursor: pointer;
  }
  .draft-pill:hover {
    background: color-mix(in srgb, var(--accent) 24%, transparent);
  }
  .draft-pill:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  @media (max-width: 599px) {
    .words {
      display: none;
    }
    .draft-pill {
      min-width: 36px;
      height: 36px;
      padding: 0 8px;
      justify-content: center;
    }
  }
</style>
