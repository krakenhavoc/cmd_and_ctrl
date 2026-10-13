<script lang="ts" module>
  /** One entry of a ⋯ menu: a button, or a link (a download, say). */
  export interface MoreItem {
    label: string;
    onclick?: () => void;
    href?: string;
    download?: string;
    danger?: boolean;
    disabled?: boolean;
    title?: string;
  }
</script>

<script lang="ts">
  // MoreMenu — the ⋯ button that holds a row's less-used actions (ADR
  // 0158 §1: a Now card's and a History row's). A small popup menu,
  // closed by choosing an item, Escape, or a click outside it.
  import Icon from "./Icon.svelte";

  interface Props {
    /** The button's accessible name, e.g. "more for Friday Pod". */
    label: string;
    items: MoreItem[];
  }

  const { label, items }: Props = $props();

  let open = $state(false);
  let root = $state<HTMLElement | null>(null);
  let button = $state<HTMLButtonElement | null>(null);
  let menu = $state<HTMLElement | null>(null);

  function toggle(): void {
    open = !open;
    if (open) queueMicrotask(() => menu?.querySelector<HTMLElement>("[role=menuitem]")?.focus());
  }

  function close(focusButton = false): void {
    open = false;
    if (focusButton) button?.focus();
  }

  function choose(item: MoreItem): void {
    close();
    item.onclick?.();
  }

  function onWindowClick(e: MouseEvent): void {
    if (open && root && !root.contains(e.target as Node)) close();
  }

  function onKey(e: KeyboardEvent): void {
    if (!open) return;
    if (e.key === "Escape") {
      e.preventDefault();
      close(true);
      return;
    }
    if (e.key !== "ArrowDown" && e.key !== "ArrowUp") return;
    const all = [...(menu?.querySelectorAll<HTMLElement>("[role=menuitem]") ?? [])];
    if (all.length === 0) return;
    e.preventDefault();
    const at = all.indexOf(document.activeElement as HTMLElement);
    const next = e.key === "ArrowDown" ? at + 1 : at - 1;
    all[(next + all.length) % all.length].focus();
  }
</script>

<svelte:window onclick={onWindowClick} />

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="more" bind:this={root} onkeydown={onKey}>
  <button
    type="button"
    class="btn ghost sm more-btn"
    aria-label={label}
    aria-haspopup="menu"
    aria-expanded={open}
    bind:this={button}
    onclick={toggle}
  >
    <Icon name="more" size={16} />
  </button>
  {#if open}
    <div class="more-menu" role="menu" aria-label={label} bind:this={menu}>
      {#each items as item (item.label)}
        {#if item.href}
          <a
            class="mi"
            class:danger={item.danger}
            role="menuitem"
            href={item.href}
            download={item.download}
            title={item.title}
            onclick={() => close()}>{item.label}</a
          >
        {:else}
          <button
            type="button"
            class="mi"
            class:danger={item.danger}
            role="menuitem"
            disabled={item.disabled}
            title={item.title}
            onclick={() => choose(item)}>{item.label}</button
          >
        {/if}
      {/each}
    </div>
  {/if}
</div>

<style>
  .more {
    position: relative;
    display: inline-flex;
  }
  .more-btn {
    width: 34px;
    padding: 0;
  }
  .more-menu {
    position: absolute;
    top: calc(100% + 6px);
    right: 0;
    z-index: 30;
    min-width: 200px;
    display: flex;
    flex-direction: column;
    padding: 6px;
    border-radius: var(--radius-lg);
    border: 1px solid var(--border-strong);
    background: var(--surface-raised);
    box-shadow: var(--shadow);
  }
  .mi {
    display: block;
    width: 100%;
    box-sizing: border-box;
    padding: 8px 10px;
    border: 0;
    border-radius: var(--radius);
    background: transparent;
    color: var(--fg);
    font: 600 13px/1.3 var(--font-ui);
    text-align: left;
    text-decoration: none;
    white-space: nowrap;
    cursor: pointer;
  }
  .mi:hover:not(:disabled),
  .mi:focus-visible {
    background: var(--surface-hover);
    outline: none;
  }
  .mi:disabled {
    opacity: 0.5;
    cursor: default;
  }
  .mi.danger {
    color: var(--danger);
  }
</style>
