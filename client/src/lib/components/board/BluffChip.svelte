<script lang="ts">
  // BluffChip — ADR 0111 §5. The bluff control is a split button that is
  // always on the table. The main part arms and disarms in one click
  // (bluff.ts pressBluff, shared with the `b` key); the ▾ part opens a
  // small popover for which bluff and how. Both write the same
  // gameplay.bluff* settings the Settings panel writes, so the two never
  // disagree. The pause range stays in Settings, linked from here.
  //
  // Disabled, not hidden, while smart auto-pass is off: with it off you
  // stop at every opponent spell, so a pause tells nobody anything.

  import { bluffArmed, pressBluff } from "../../bluff";
  import { openSettings, settings, updateSettings } from "../../settings";

  interface Props {
    // " (b)" or "", from the action dock's key-hint helper.
    keyHint?: string;
    // The same key for aria-keyshortcuts ("B"), or undefined.
    keyShortcuts?: string;
    // Disabled whatever the settings say: the opening roll, before
    // there is a turn to bluff in (ADR 0121 §6).
    locked?: boolean;
  }
  const { keyHint = "", keyShortcuts, locked = false }: Props = $props();

  const gp = $derived($settings.gameplay);
  const enabled = $derived(gp.smartAutoPass && !locked);
  let open = $state(false);
  let root: HTMLElement | undefined = $state();

  // "bluff ✓ counter" names what was asked for when only one kind is on;
  // with both (or neither) the bare "bluff ✓" says enough.
  const kindSuffix = $derived(
    gp.bluffCounterspell === gp.bluffInstant ? "" : gp.bluffCounterspell ? " counter" : " instant",
  );
  const label = $derived($bluffArmed ? `bluff ✓${kindSuffix}` : "bluff");

  const title = $derived(
    (locked
      ? "nothing to bluff before the first turn"
      : !enabled
        ? "bluffing needs smart auto-pass: with it off you stop at every opponent spell, so a pause gives nothing away"
        : $bluffArmed
          ? "bluff ON — when you have no answer, pause anyway so a pause gives nothing away; click to stop bluffing this game"
          : "bluff OFF — windows you can't answer pass instantly; click to bluff for the rest of this game") +
      (enabled ? keyHint : ""),
  );

  function onWindowPointerDown(e: PointerEvent): void {
    if (open && root && !root.contains(e.target as Node)) open = false;
  }
  function onKeydown(e: KeyboardEvent): void {
    if (e.key !== "Escape" || !open) return;
    e.stopPropagation();
    open = false;
    root?.querySelector<HTMLButtonElement>("button.bluff-caret")?.focus();
  }
  function toSettings(): void {
    open = false;
    openSettings("gameplay");
  }
</script>

<svelte:window onpointerdown={onWindowPointerDown} />

<div class="bluff-chip" bind:this={root} role="presentation" onkeydown={onKeydown}>
  <div class="split">
    <button
      type="button"
      class="action hold bluff bluff-main"
      class:on={$bluffArmed && enabled}
      aria-pressed={$bluffArmed}
      aria-keyshortcuts={enabled ? keyShortcuts : undefined}
      disabled={!enabled}
      onclick={pressBluff}
      {title}
    >
      {label}
    </button>
    <button
      type="button"
      class="action hold bluff bluff-caret"
      class:on={$bluffArmed && enabled}
      aria-label="bluff options"
      aria-haspopup="true"
      aria-expanded={open}
      disabled={!enabled}
      onclick={() => (open = !open)}
      title={enabled ? "choose which bluff, and how" : title}
    >
      ▾
    </button>
  </div>

  {#if open && enabled}
    <div class="pop" role="group" aria-label="bluff settings">
      <label>
        <input
          type="checkbox"
          checked={gp.bluffCounterspell}
          onchange={(e) => updateSettings("gameplay", "bluffCounterspell", e.currentTarget.checked)}
        />
        Represent a counterspell
      </label>
      <label>
        <input
          type="checkbox"
          checked={gp.bluffInstant}
          onchange={(e) => updateSettings("gameplay", "bluffInstant", e.currentTarget.checked)}
        />
        Represent an instant
      </label>
      <div class="style" role="radiogroup" aria-label="bluff style">
        <label>
          <input
            type="radio"
            name="bluff-mode"
            value="timed"
            checked={gp.bluffMode === "timed"}
            onchange={() => updateSettings("gameplay", "bluffMode", "timed")}
          />
          Timed
        </label>
        <label>
          <input
            type="radio"
            name="bluff-mode"
            value="manual"
            checked={gp.bluffMode === "manual"}
            onchange={() => updateSettings("gameplay", "bluffMode", "manual")}
          />
          Manual
        </label>
      </div>
      <button type="button" class="link" onclick={toSettings}>Pause range in Settings…</button>
    </div>
  {/if}
</div>

<style>
  .bluff-chip {
    position: relative;
    flex: 1 1 auto;
    min-width: 0;
    display: flex;
  }
  .split {
    display: flex;
    flex: 1 1 auto;
    min-width: 0;
  }
  /* The two halves share one outline: the main part keeps the row's
     button look, the caret is a narrow tail. The toggles row's .action rules
     live in ActionDock; these are the chip's own. */
  .action {
    flex: 1 1 auto;
    min-width: 0;
    white-space: nowrap;
    padding: 4px 8px;
    border: 1px solid var(--border);
    background: var(--overlay-faint);
    color: var(--fg);
    font-size: 0.85em;
    font-weight: 600;
    letter-spacing: 0.02em;
    cursor: pointer;
  }
  .bluff-main {
    border-radius: 6px 0 0 6px;
  }
  .bluff-caret {
    flex: 0 0 auto;
    padding: 4px 6px;
    border-left-width: 0;
    border-radius: 0 6px 6px 0;
  }
  .action:hover:not(:disabled) {
    background: color-mix(in srgb, var(--overlay-ink) 8%, transparent);
  }
  .action:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  /* Armed reads in the "user override" colour hold uses. */
  .action.on {
    background: color-mix(in srgb, var(--magenta) 18%, transparent);
    color: var(--magenta);
    font-weight: 700;
    border-color: color-mix(in srgb, var(--magenta) 55%, transparent);
  }
  .action.on:hover:not(:disabled) {
    background: color-mix(in srgb, var(--magenta) 28%, transparent);
  }
  .action:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }
  /* Opens upward and right-aligned: the widget sits in the bottom-right
     corner, so down or left would leave the screen. */
  .pop {
    position: absolute;
    right: 0;
    bottom: calc(100% + 4px);
    z-index: 30;
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 200px;
    max-width: min(260px, calc(100vw - 32px));
    color: var(--fg);
    padding: 8px 10px;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--surface-raised);
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.5);
    font-size: 0.85em;
  }
  .pop label {
    display: flex;
    align-items: center;
    gap: 6px;
    cursor: pointer;
  }
  .style {
    display: flex;
    gap: 12px;
  }
  .link {
    align-self: flex-start;
    padding: 0;
    border: 0;
    background: none;
    color: var(--accent);
    font: inherit;
    cursor: pointer;
    text-decoration: underline;
  }
</style>
