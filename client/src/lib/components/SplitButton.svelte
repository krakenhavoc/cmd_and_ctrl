<script lang="ts">
  // SplitButton — one control, two actions (#3499): the wide part does
  // the everyday thing (copy a table's link) and the narrow ↻ part the
  // rare one (mint a new link, which the caller confirms first). The
  // ↻ part is optional: someone who may copy a link but not replace it
  // gets a plain button of the same shape.
  import Icon from "./Icon.svelte";
  import type { IconName } from "../icons";

  interface Props {
    /** The main part's visible text. */
    label: string;
    /** Shown in place of `label` while `done` is set, e.g. "Invite copied". */
    doneLabel?: string;
    done?: boolean;
    icon?: IconName;
    title?: string;
    disabled?: boolean;
    onmain: () => void;
    /** The ↻ part's accessible name; with `onrotate`, the part is shown. */
    rotateLabel?: string;
    rotateTitle?: string;
    onrotate?: () => void;
  }

  const {
    label,
    doneLabel,
    done = false,
    icon = "link",
    title,
    disabled = false,
    onmain,
    rotateLabel,
    rotateTitle,
    onrotate,
  }: Props = $props();
</script>

<span class="split" class:solo={!onrotate}>
  <button type="button" class="main" class:on={done} {title} {disabled} onclick={onmain}>
    {#if done}
      <Icon name="check" size={13} /> {doneLabel ?? label}
    {:else}
      <Icon name={icon} size={13} /> {label}
    {/if}
  </button>
  {#if onrotate}
    <button
      type="button"
      class="rot"
      aria-label={rotateLabel}
      title={rotateTitle ?? rotateLabel}
      {disabled}
      onclick={onrotate}
    >
      <Icon name="cycle" size={13} />
    </button>
  {/if}
</span>

<style>
  /* Matches .btn.ghost.sm (app.css), so it sits in a row of them. */
  .split {
    display: inline-flex;
    align-items: stretch;
    height: 32px;
    box-sizing: border-box;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    background: transparent;
    overflow: hidden;
  }
  button {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    border: 0;
    background: transparent;
    color: var(--fg);
    font: 700 12.5px/1 var(--font-ui);
    letter-spacing: 0.005em;
    padding: 0 12px;
    cursor: pointer;
    white-space: nowrap;
  }
  button:hover:not(:disabled) {
    background: var(--overlay);
  }
  button:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }
  button:disabled {
    opacity: 0.55;
    cursor: default;
  }
  .main {
    flex: 1 1 auto;
  }
  .main.on {
    color: var(--mint);
  }
  .rot {
    flex: 0 0 auto;
    padding: 0 10px;
    border-left: 1px solid var(--border-strong);
    color: var(--fg-muted);
  }
  .rot:hover:not(:disabled) {
    color: var(--fg);
  }
</style>
