<script lang="ts">
  // ContextMenu — the site's own right-click menu (#3499), opened by
  // lib/menuTrigger.ts at a point on the screen. The items are a ⋯
  // menu's (MoreMenu's MoreItem), so a surface can hand the same list
  // to either.
  //
  // Fixed to the viewport and clamped inside it once it knows its own
  // size. It takes focus when it opens, moves with the arrow keys, Home
  // and End, and closes on Escape (focus goes back to whatever opened
  // it), Tab, a choice, a click outside, a scroll or a resize.
  import { onMount, tick } from "svelte";
  import type { MoreItem } from "./MoreMenu.svelte";

  interface Props {
    /** The menu's accessible name, e.g. "actions for Friday Pod". */
    label: string;
    x: number;
    y: number;
    items: MoreItem[];
    /** Called once; `restoreFocus` is true when Escape closed it. */
    onclose: (restoreFocus: boolean) => void;
  }

  const { label, x, y, items, onclose }: Props = $props();

  // The gap kept between the menu and the viewport's edge.
  const EDGE = 8;

  let menu = $state<HTMLElement | null>(null);
  let left = $state(0);
  let top = $state(0);
  let closed = false;
  // Where the page was scrolled when the menu opened. A scroll event can
  // land a frame after the menu mounts for a scroll that finished before
  // it (the browser bringing the right-clicked card into view), so the
  // menu closes only once the page has actually moved.
  const scrolledAt = { x: window.scrollX, y: window.scrollY };

  function onWindowScroll(): void {
    if (window.scrollX !== scrolledAt.x || window.scrollY !== scrolledAt.y) close();
  }

  function close(restoreFocus = false): void {
    if (closed) return;
    closed = true;
    onclose(restoreFocus);
  }

  function enabled(): HTMLElement[] {
    return [...(menu?.querySelectorAll<HTMLElement>("[role=menuitem]:not(:disabled)") ?? [])];
  }

  function place(): void {
    const w = menu?.offsetWidth ?? 0;
    const h = menu?.offsetHeight ?? 0;
    left = Math.max(EDGE, Math.min(x, window.innerWidth - w - EDGE));
    top = Math.max(EDGE, Math.min(y, window.innerHeight - h - EDGE));
  }

  $effect(() => {
    // Re-place when the point moves (the menu reopened somewhere else).
    void x;
    void y;
    place();
  });

  onMount(() => {
    void tick().then(() => {
      place();
      enabled()[0]?.focus({ preventScroll: true });
    });
  });

  function choose(item: MoreItem): void {
    close();
    item.onclick?.();
  }

  function onKey(e: KeyboardEvent): void {
    if (e.key === "Escape") {
      e.preventDefault();
      e.stopPropagation();
      close(true);
      return;
    }
    if (e.key === "Tab") {
      close();
      return;
    }
    const all = enabled();
    if (all.length === 0) return;
    const at = all.indexOf(document.activeElement as HTMLElement);
    let next: number | null = null;
    if (e.key === "ArrowDown") next = (at + 1) % all.length;
    else if (e.key === "ArrowUp") next = (at - 1 + all.length) % all.length;
    else if (e.key === "Home") next = 0;
    else if (e.key === "End") next = all.length - 1;
    if (next === null) return;
    e.preventDefault();
    all[next].focus({ preventScroll: true });
  }

  function onWindowPointer(e: PointerEvent): void {
    if (menu && !menu.contains(e.target as Node)) close();
  }
</script>

<svelte:window
  onpointerdown={onWindowPointer}
  onresize={() => close()}
  onscroll={onWindowScroll}
  onblur={() => close()}
/>

<div
  class="ctx-menu"
  role="menu"
  aria-label={label}
  tabindex="-1"
  style:left={`${left}px`}
  style:top={`${top}px`}
  bind:this={menu}
  onkeydown={onKey}
  oncontextmenu={(e) => e.preventDefault()}
>
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

<style>
  .ctx-menu {
    position: fixed;
    z-index: 60;
    min-width: 220px;
    max-width: min(320px, calc(100vw - 16px));
    max-height: calc(100vh - 16px);
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    padding: 6px;
    border-radius: var(--radius-lg);
    border: 1px solid var(--border-strong);
    background: var(--surface-raised);
    box-shadow: var(--shadow);
    outline: none;
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
    overflow: hidden;
    text-overflow: ellipsis;
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
