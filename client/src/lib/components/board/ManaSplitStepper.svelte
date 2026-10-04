<script lang="ts">
  // ManaSplitStepper — one −/+ row per colour, a running "N of N", and
  // Add mana (ADR 0117 §4). It answers a mana activation with two or
  // more picking slots in one dialog: Vivi Ornitier at power 3, Relic of
  // Sauron, a filter land, Gwenna, Cascading Cataracts, Selvala.
  //
  // Rows follow the server's colour order (the union of the lists by
  // first appearance, identity first, #843), each labelled by its
  // colour's name and symbol. The total is fixed: one per list. "+" is
  // enabled while one more is still feasible (Hall's condition,
  // manaStepper.ts), "−" down to a fixed slot's share. Add mana sends
  // one colour per slot, in slot order, from a matching.
  //
  // The lists are re-read from every frame. If their number changes
  // (Vivi's power moved), the counts go back to the start state with
  // the new total. Nothing is sent until Add mana; Escape sends nothing.
  //
  // Lives inside ManaSourcePicker, which owns the dialog, its name, its
  // placement and the outside-click close.
  import { tick, untrack } from "svelte";
  import {
    canAdd,
    canRemove,
    rememberSplit,
    rememberedSplit,
    splitAdjust,
    splitAssignment,
    splitColors,
    splitComplete,
    splitStart,
    splitTotal,
    type SplitCounts,
  } from "../../manaStepper";
  import { manaSymbolMeta } from "../../manaSymbol";
  import ManaSymbol from "./ManaSymbol.svelte";

  interface Props {
    /** The ability's `color_options`, one list per picking slot. */
    lists: string[][];
    /** Where this card's last confirmed split is remembered. */
    memoryKey: string;
    /** Non-empty when the ability cannot be activated right now. */
    disabled?: string;
    /** Add mana: one colour per slot, in slot order. */
    onConfirm: (colors: string[]) => void;
    onCancel: () => void;
  }

  const { lists, memoryKey, disabled = "", onConfirm, onCancel }: Props = $props();

  let counts = $state<SplitCounts>(untrack(() => splitStart(lists, rememberedSplit(memoryKey))));
  // The total the counts were set for. Plain, not $state: the effect
  // below reads it to notice a change, and writing it must not re-run.
  let shownTotal = untrack(() => lists.length);
  // A colour the new lists no longer offer has no row to step it back
  // down with, so that resets too.
  $effect(() => {
    const n = lists.length;
    const offered = splitColors(lists);
    const stale = untrack(() =>
      Object.entries(counts).some(([c, k]) => k > 0 && !offered.includes(c)),
    );
    if (n === shownTotal && !stale) return;
    shownTotal = n;
    counts = splitStart(
      lists,
      untrack(() => rememberedSplit(memoryKey)),
    );
  });

  const colors = $derived(splitColors(lists));
  const total = $derived(splitTotal(counts));
  const ready = $derived(!disabled && splitComplete(lists, counts));
  const fixedOnly = $derived(
    new Set(
      colors.filter((c) => lists.every((l) => !l.includes(c) || (l.length === 1 && l[0] === c))),
    ),
  );

  const nameOf = (c: string): string => manaSymbolMeta(c).name.toLowerCase();
  const titleOf = (c: string): string => manaSymbolMeta(c).name;

  let root: HTMLElement | null = $state(null);
  // The row the keyboard steps: the one holding focus, else the first.
  let active = $state(0);

  function step(c: string, delta: number): void {
    if (delta > 0 && !canAdd(lists, counts, c)) return;
    if (delta < 0 && !canRemove(lists, counts, c)) return;
    counts = splitAdjust(counts, c, delta);
  }

  function confirm(): void {
    if (!ready) return;
    const out = splitAssignment(lists, counts);
    if (!out) return;
    rememberSplit(memoryKey, counts);
    onConfirm(out);
  }

  async function focusRow(i: number): Promise<void> {
    active = Math.max(0, Math.min(colors.length - 1, i));
    await tick();
    const row = root?.querySelectorAll<HTMLElement>("[data-split-row]")[active];
    if (!row) return;
    const target =
      row.querySelector<HTMLButtonElement>("button[data-step='more']:not([disabled])") ??
      row.querySelector<HTMLButtonElement>("button[data-step='less']:not([disabled])") ??
      row;
    target.focus({ preventScroll: true });
  }

  function onKey(ev: KeyboardEvent): void {
    if (ev.ctrlKey || ev.metaKey || ev.altKey) return;
    const c = colors[active];
    switch (ev.key) {
      case "Escape":
        ev.preventDefault();
        ev.stopPropagation();
        onCancel();
        return;
      case "Enter":
        // Enter confirms wherever focus is, and never presses the
        // focused −/+ as well.
        ev.preventDefault();
        ev.stopPropagation();
        confirm();
        return;
      case "ArrowUp":
        ev.preventDefault();
        void focusRow(active - 1);
        return;
      case "ArrowDown":
        ev.preventDefault();
        void focusRow(active + 1);
        return;
      case "ArrowLeft":
      case "-":
        ev.preventDefault();
        if (c) step(c, -1);
        return;
      case "ArrowRight":
      case "+":
      case "=":
        ev.preventDefault();
        if (c) step(c, 1);
        return;
    }
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="split" bind:this={root}>
  {#each colors as c, i (c)}
    {@const name = nameOf(c)}
    {@const k = counts[c] ?? 0}
    <div
      class="row"
      class:active={i === active}
      role="group"
      aria-label={name}
      tabindex="-1"
      data-split-row={c}
      onfocusin={() => (active = i)}
    >
      <span class="color">
        <ManaSymbol symbol={c} size={24} />
        <span class="name">{titleOf(c)}</span>
        {#if fixedOnly.has(c)}
          <span class="fixed">fixed</span>
        {/if}
      </span>
      <button
        type="button"
        class="step"
        data-step="less"
        aria-label={`less ${name}`}
        title={`One less ${name}`}
        disabled={!canRemove(lists, counts, c)}
        onclick={(ev) => {
          ev.stopPropagation();
          active = i;
          step(c, -1);
        }}>−</button
      >
      <output class="count" aria-label={`${name} count`}>{k}</output>
      <button
        type="button"
        class="step"
        data-step="more"
        aria-label={`more ${name}`}
        title={`One more ${name}`}
        disabled={!canAdd(lists, counts, c)}
        onclick={(ev) => {
          ev.stopPropagation();
          active = i;
          step(c, 1);
        }}>+</button
      >
    </div>
  {/each}
  <div class="foot">
    <p class="status" role="status" aria-live="polite" aria-label={`${total} of ${lists.length}`}>
      {total} of {lists.length}
    </p>
    <button
      type="button"
      class="confirm"
      data-add-mana
      disabled={!ready}
      title={disabled || (ready ? "Add this mana (Enter)" : `Set ${lists.length} in all`)}
      onclick={(ev) => {
        ev.stopPropagation();
        confirm();
      }}>Add mana</button
    >
  </div>
  {#if disabled}
    <p class="note">{disabled}</p>
  {/if}
</div>

<style>
  .split {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
    max-width: 100%;
  }
  .row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 44px 32px 44px;
    align-items: center;
    gap: 6px;
    padding: 2px 4px;
    border-radius: 8px;
    outline: none;
  }
  .row.active {
    background: rgba(200, 168, 106, 0.08);
  }
  .color {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    font-size: 12px;
    font-weight: 700;
    color: #e8ecf6;
  }
  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .fixed {
    font-size: 10px;
    font-weight: 400;
    opacity: 0.7;
  }
  .step {
    width: 44px;
    height: 40px;
    background: rgba(255, 255, 255, 0.04);
    color: #e8ecf6;
    border: 1px solid rgba(200, 168, 106, 0.35);
    border-radius: 8px;
    font: inherit;
    font-size: 18px;
    line-height: 1;
    cursor: pointer;
  }
  .step:hover:not(:disabled),
  .step:focus-visible:not(:disabled) {
    background: rgba(200, 168, 106, 0.16);
    border-color: rgba(232, 200, 130, 0.9);
    outline: none;
  }
  .step:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }
  .count {
    text-align: center;
    font-size: 16px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    color: #e8ecf6;
  }
  .foot {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-top: 2px;
  }
  .status {
    margin: 0;
    font-size: 12px;
    font-variant-numeric: tabular-nums;
  }
  .confirm {
    min-height: 40px;
    padding: 0 14px;
    background: rgba(200, 168, 106, 0.22);
    color: #f4e6c4;
    border: 1px solid rgba(232, 200, 130, 0.9);
    border-radius: 8px;
    font: inherit;
    font-size: 13px;
    font-weight: 700;
    cursor: pointer;
  }
  .confirm:hover:not(:disabled),
  .confirm:focus-visible:not(:disabled) {
    background: rgba(200, 168, 106, 0.36);
    outline: none;
  }
  .confirm:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .note {
    margin: 0;
    font-size: 10px;
    opacity: 0.7;
  }
</style>
