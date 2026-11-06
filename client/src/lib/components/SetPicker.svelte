<script lang="ts">
  // SetPicker — the catalogue's set filter (#4790): a text box over a
  // filtered list of sets, in place of a <select> of several hundred
  // sets nobody could find their way through.
  //
  // An editable combobox with list autocomplete (WAI-ARIA APG): the input
  // is role=combobox and keeps focus the whole time; the options are
  // pointed at with aria-activedescendant, never focused. Typing filters
  // by name or set code (searchSets in lib/catalog.ts, where the ranking
  // has its tests). ArrowDown/ArrowUp move, Enter picks, Escape closes
  // and puts the box back to the picked set, and the clear button goes
  // back to All sets.
  //
  // The list opens inline under the box, not floating over the page: the
  // sidebar scrolls and clips, and the tablet sheet scrolls too, so a list
  // in the flow is never cut off and needs no rung on the z-index ladder.
  //
  // Options keep the input focused on a press (mousedown is prevented), so
  // a tap on a phone picks the set rather than blurring the box first.

  import { tick } from "svelte";
  import Icon from "./Icon.svelte";
  import { L } from "../labels";
  import {
    findSet,
    searchSets,
    setOptionLabel,
    setYear,
    supportedCount,
    type CatalogSet,
  } from "../catalog";

  let {
    sets,
    value = $bindable(""),
  }: {
    /** Newest first, as the server sends them. */
    sets: CatalogSet[];
    /** The picked set's code, or "" for all sets. */
    value?: string;
  } = $props();

  const uid = $props.id();
  const listId = `${uid}-sets`;

  const picked = $derived(value ? findSet(sets, value) : undefined);

  let text = $state("");
  let open = $state(false);
  // Whether the box holds what the person typed, rather than the picked
  // set's name: opening on a picked set shows every set, not just it.
  let typed = $state(false);
  let active = $state(-1);
  let inputEl = $state<HTMLInputElement | null>(null);

  // The box shows the picked set's name, and follows the filter when a
  // chip or "clear filters" changes it from outside.
  $effect(() => {
    const name = picked?.name ?? "";
    if (!open) {
      text = name;
      typed = false;
    }
  });

  type Opt = { kind: "all" } | { kind: "set"; set: CatalogSet };

  const options = $derived.by<Opt[]>(() => {
    const q = typed ? text : "";
    const matched = searchSets(sets, q).map((set) => ({ kind: "set" as const, set }));
    return q.trim() ? matched : [{ kind: "all" }, ...matched];
  });

  function optId(o: Opt): string {
    return o.kind === "all" ? `${uid}-all` : `${uid}-set-${o.set.code}`;
  }
  function isSelected(o: Opt): boolean {
    return o.kind === "all" ? !value : o.set.code === value;
  }

  async function scrollActive() {
    await tick();
    const o = options[active];
    if (!o) return;
    document.getElementById(optId(o))?.scrollIntoView?.({ block: "nearest" });
  }

  function openList() {
    if (open) return;
    open = true;
    const i = options.findIndex(isSelected);
    active = i;
    void scrollActive();
  }

  function close() {
    open = false;
    active = -1;
    text = picked?.name ?? "";
    typed = false;
  }

  function pick(o: Opt) {
    value = o.kind === "all" ? "" : o.set.code;
    open = false;
    active = -1;
    typed = false;
    text = o.kind === "all" ? "" : o.set.name;
  }

  function clear() {
    value = "";
    text = "";
    typed = false;
    open = false;
    active = -1;
    inputEl?.focus();
  }

  function oninput() {
    typed = true;
    open = true;
    // The first match is the one Enter takes.
    active = text.trim() && options.length > 0 ? 0 : -1;
  }

  function onkeydown(ev: KeyboardEvent) {
    switch (ev.key) {
      case "ArrowDown":
        ev.preventDefault();
        if (!open || ev.altKey) {
          openList();
          return;
        }
        if (options.length > 0) active = Math.min(active + 1, options.length - 1);
        void scrollActive();
        return;
      case "ArrowUp":
        ev.preventDefault();
        if (!open) {
          openList();
          return;
        }
        if (options.length > 0) active = Math.max(active - 1, 0);
        void scrollActive();
        return;
      case "Enter":
        if (open && active >= 0 && options[active]) {
          ev.preventDefault();
          pick(options[active]);
        }
        return;
      case "Escape":
        if (open) {
          // Close the list only; the tablet sheet that holds it stays open.
          ev.preventDefault();
          ev.stopPropagation();
          close();
        }
        return;
      case "Tab":
        if (open) close();
        return;
    }
  }

  function onfocusout(ev: FocusEvent) {
    const next = ev.relatedTarget as Node | null;
    if (next && (ev.currentTarget as HTMLElement).contains(next)) return;
    if (open) close();
  }

  const status = $derived(
    !open
      ? ""
      : options.filter((o) => o.kind === "set").length === 1
        ? "1 set"
        : `${options.filter((o) => o.kind === "set").length} sets`,
  );
</script>

<div class="picker" {onfocusout}>
  <div class="box">
    <input
      bind:this={inputEl}
      bind:value={text}
      type="text"
      class="input"
      role="combobox"
      aria-label={L.filterCatalogueBySet}
      aria-autocomplete="list"
      aria-expanded={open}
      aria-controls={listId}
      aria-activedescendant={open && active >= 0 && options[active]
        ? optId(options[active])
        : undefined}
      placeholder="All sets — type a name or code…"
      name="set"
      autocomplete="off"
      autocapitalize="off"
      spellcheck="false"
      enterkeyhint="done"
      {oninput}
      {onkeydown}
      onclick={openList}
    />
    {#if text || value}
      <button
        type="button"
        class="clear"
        aria-label={L.clearCatalogueSet}
        title={L.clearCatalogueSet}
        onclick={clear}
      >
        <Icon name="x" size={12} />
      </button>
    {/if}
  </div>

  <ul class="list" id={listId} role="listbox" aria-label="sets" hidden={!open}>
    {#if open}
      {#each options as o, i (optId(o))}
        <!-- Keyboard is handled on the combobox, which keeps focus (aria-activedescendant). -->
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <li
          id={optId(o)}
          role="option"
          class="opt"
          class:active={i === active}
          aria-selected={isSelected(o)}
          aria-label={o.kind === "all"
            ? "All sets"
            : `${setOptionLabel(o.set)}, ${o.set.code.toUpperCase()}${setYear(o.set) ? `, ${setYear(o.set)}` : ""}`}
          onmousedown={(e) => e.preventDefault()}
          onpointermove={() => (active = i)}
          onclick={() => pick(o)}
        >
          {#if o.kind === "all"}
            <span class="name">All sets</span>
          {:else}
            <span class="name">{o.set.name}</span>
            <span class="meta">
              <span class="code">{o.set.code.toUpperCase()}</span>
              {#if setYear(o.set)}<span class="year">{setYear(o.set)}</span>{/if}
              <span class="n" title="{supportedCount(o.set)} of {o.set.total} cards supported"
                >{supportedCount(o.set)}/{o.set.total}</span
              >
            </span>
          {/if}
        </li>
      {/each}
    {/if}
  </ul>
  {#if open && options.length === 0}
    <p class="none">No set matches “{text.trim()}”.</p>
  {/if}
  <span class="sr-only" role="status" aria-live="polite">{status}</span>
</div>

<style>
  .picker {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
  }
  .box {
    position: relative;
    display: flex;
    align-items: center;
    color: var(--fg-dim);
  }
  .input {
    width: 100%;
    box-sizing: border-box;
    padding-right: 34px;
    text-overflow: ellipsis;
  }
  .clear {
    position: absolute;
    right: 4px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    padding: 0;
    border: 0;
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--fg-muted);
    cursor: pointer;
    transition:
      background-color 120ms var(--ease),
      color 120ms var(--ease);
  }
  .clear:hover {
    background: var(--surface-hover);
    color: var(--fg);
  }
  .list {
    margin: 0;
    padding: 4px;
    list-style: none;
    max-height: min(320px, 50vh);
    overflow-y: auto;
    overscroll-behavior: contain;
    background: var(--surface-sunken);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
  }
  .list[hidden] {
    display: none;
  }
  .opt {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-height: 36px;
    justify-content: center;
    padding: 5px 8px;
    border: 1px solid transparent;
    border-radius: 6px;
    color: var(--fg-muted);
    font-size: 13px;
    cursor: pointer;
    touch-action: manipulation;
    -webkit-tap-highlight-color: transparent;
    /* Several hundred sets: let the browser skip laying out the ones
       scrolled out of the list. */
    content-visibility: auto;
    contain-intrinsic-size: auto 44px;
  }
  .opt.active {
    background: var(--surface-hover);
    color: var(--fg);
  }
  .opt[aria-selected="true"] {
    background: var(--accent-soft);
    border-color: var(--accent);
    color: var(--accent-strong);
  }
  .name {
    min-width: 0;
    overflow-wrap: anywhere;
  }
  .meta {
    display: flex;
    align-items: center;
    gap: 6px;
    font-family: var(--font-mono);
    font-size: 10.5px;
    color: var(--fg-muted);
  }
  .code {
    padding: 0 5px;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-sm);
    letter-spacing: 0.06em;
    font-weight: 600;
    line-height: 16px;
  }
  .n {
    margin-left: auto;
    font-variant-numeric: tabular-nums;
  }
  .none {
    margin: 0;
    padding: 6px 2px;
    font-size: 12.5px;
    color: var(--fg-muted);
  }
  /* A finger needs a bigger target than the 28px a pointer gets. */
  @media (pointer: coarse) {
    .input {
      padding-right: 44px;
    }
    .clear {
      right: 0;
      width: 40px;
      height: 40px;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .clear {
      transition: none;
    }
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
</style>
