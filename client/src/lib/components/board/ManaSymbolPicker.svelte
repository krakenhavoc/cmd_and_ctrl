<script lang="ts">
  // ManaSymbolPicker — one large mana-symbol button per option (#1438).
  //
  // The one picker for "which mana?": the anchored popover a
  // multi-ability source opens on click (ManaSourcePicker), and the
  // mana_pick / choose_color prompt in ChoicePromptModal. Options come
  // in already ordered — the server's colour order for a prompt (#843,
  // commander identity first), the server's ability order for a
  // source — and are drawn in that order.
  //
  // Keyboard: 1–9 pick the matching option, Escape cancels when the
  // caller allows cancelling. A mana_pick prompt passes no onCancel:
  // by the time the server asks, the source is already tapped, and the
  // prompt holds the table until it is answered.
  import type { ManaPickOption } from "../../manaSource";
  import ManaSymbol from "./ManaSymbol.svelte";

  interface Props {
    options: ManaPickOption[];
    onPick: (option: ManaPickOption) => void;
    onCancel?: () => void;
    /** Accessible name for the group. */
    label?: string;
    /** Listen for 1–9 / Escape on the window. */
    keyboard?: boolean;
  }

  const { options, onPick, onCancel, label = "choose mana", keyboard = true }: Props = $props();

  function pick(option: ManaPickOption): void {
    if (option.disabled) return;
    onPick(option);
  }

  function onKey(ev: KeyboardEvent): void {
    if (!keyboard) return;
    if (ev.ctrlKey || ev.metaKey || ev.altKey) return;
    const t = ev.target as HTMLElement | null;
    if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable)) return;
    if (ev.key === "Escape") {
      if (!onCancel) return;
      ev.preventDefault();
      ev.stopPropagation();
      onCancel();
      return;
    }
    if (!/^[1-9]$/.test(ev.key)) return;
    const option = options[Number(ev.key) - 1];
    if (!option) return;
    ev.preventDefault();
    ev.stopPropagation();
    pick(option);
  }

  // One symbol is drawn big; a pair (a Signet's {W}{U}, Sol Ring's
  // {C}{C}) a little smaller; alternatives ("{R} or {W}") smaller
  // still, because the colour is picked in the next step.
  function symbolSize(o: ManaPickOption): number {
    if (o.choice) return o.symbols.length > 3 ? 18 : 24;
    if (o.symbols.length <= 1) return 40;
    if (o.symbols.length === 2) return 30;
    return 22;
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="mana-picker" role="group" aria-label={label}>
  {#each options as o, i (o.key)}
    <button
      type="button"
      class="mana-option"
      class:has-rider={!!o.rider}
      disabled={!!o.disabled}
      title={o.rider ? `${o.title} — ${o.rider}` : o.title}
      aria-label={o.rider ? `${o.title}, ${o.rider}` : o.title}
      aria-keyshortcuts={i < 9 ? String(i + 1) : undefined}
      data-key={o.key}
      onclick={(ev) => {
        ev.stopPropagation();
        pick(o);
      }}
    >
      {#if i < 9}
        <span class="hotkey" aria-hidden="true">{i + 1}</span>
      {/if}
      <span class="symbols" class:choice={o.choice}>
        {#each o.symbols as s, j (j)}
          <ManaSymbol symbol={s} size={symbolSize(o)} />
        {/each}
      </span>
      <span class="caption">{o.caption}</span>
      {#if o.choice}
        <span class="note">pick the color next</span>
      {/if}
      {#if o.rider}
        <span class="rider">{o.rider}</span>
      {/if}
      {#if o.disabled}
        <span class="note">{o.disabled}</span>
      {/if}
    </button>
  {/each}
</div>

<style>
  .mana-picker {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 8px;
    max-width: 100%;
  }
  .mana-option {
    position: relative;
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-start;
    gap: 4px;
    /* 44px is the touch-target floor; the symbol alone is 40. */
    min-width: 64px;
    min-height: 72px;
    max-width: 140px;
    padding: 10px 8px 8px;
    box-sizing: border-box;
    background: rgba(255, 255, 255, 0.04);
    color: #e8ecf6;
    border: 1px solid rgba(200, 168, 106, 0.35);
    border-radius: 10px;
    cursor: pointer;
    font: inherit;
    transition:
      transform 120ms ease,
      background 120ms ease,
      border-color 120ms ease;
  }
  .mana-option:hover:not(:disabled),
  .mana-option:focus-visible:not(:disabled) {
    background: rgba(200, 168, 106, 0.16);
    border-color: rgba(232, 200, 130, 0.9);
    transform: translateY(-2px);
    outline: none;
  }
  .mana-option:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .hotkey {
    position: absolute;
    top: 3px;
    left: 5px;
    font-size: 9px;
    font-weight: 700;
    opacity: 0.55;
  }
  .symbols {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    min-height: 40px;
  }
  .symbols.choice {
    flex-wrap: wrap;
    justify-content: center;
    /* Five 18px alternatives on one line ("any color"). */
    max-width: 112px;
  }
  .caption {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.02em;
    text-align: center;
    line-height: 1.2;
  }
  .rider {
    font-size: 10px;
    line-height: 1.2;
    text-align: center;
    color: #ffb3a8;
  }
  .note {
    font-size: 10px;
    line-height: 1.2;
    text-align: center;
    opacity: 0.7;
  }
</style>
