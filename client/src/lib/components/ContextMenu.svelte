<script lang="ts">
  // ContextMenu — the site's own right-click menu (#3499), opened by
  // lib/menuTrigger.ts at a point on the screen. The items are a ⋯
  // menu's (MoreMenu's MoreItem), so a surface can hand the same list
  // to either. A table's Invite menu (ADR 0212 §4) is one too, opened
  // under its button with `align="end"`; its items carry a glyph, a
  // line that says what each does, and a heading over the last group.
  //
  // Fixed to the viewport and clamped inside it once it knows its own
  // size. It takes focus when it opens, moves with the arrow keys, Home
  // and End, and closes on Escape (focus goes back to whatever opened
  // it), Tab, a choice, a click outside, a scroll or a resize.
  import { onMount, tick } from "svelte";
  import Icon from "./Icon.svelte";
  import type { MoreItem } from "./MoreMenu.svelte";

  interface Props {
    /** The menu's accessible name, e.g. "actions for Friday Pod". */
    label: string;
    x: number;
    y: number;
    /** "end": x is where the menu's right edge goes, as under a button at the right. */
    align?: "start" | "end";
    items: MoreItem[];
    /** Called once; `restoreFocus` is true when Escape closed it. */
    onclose: (restoreFocus: boolean) => void;
  }

  const { label, x, y, align = "start", items, onclose }: Props = $props();

  // The items in runs that share a heading, in order; an item with no
  // group is a run of its own with no heading.
  const sections = $derived.by(() => {
    const out: { group?: string; items: MoreItem[] }[] = [];
    for (const item of items) {
      const last = out[out.length - 1];
      if (last && item.group && last.group === item.group) last.items.push(item);
      else out.push({ group: item.group, items: [item] });
    }
    return out;
  });

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
    const want = align === "end" ? x - w : x;
    left = Math.max(EDGE, Math.min(want, window.innerWidth - w - EDGE));
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
  {#each sections as section, i (i)}
    {#if section.group}
      {#if i > 0}<div class="msep" role="separator"></div>{/if}
      <div role="group" aria-label={section.group}>
        <div class="mhead" aria-hidden="true">{section.group}</div>
        {#each section.items as item (item.label)}
          {@render entry(item)}
        {/each}
      </div>
    {:else}
      {#each section.items as item (item.label)}
        {@render entry(item)}
      {/each}
    {/if}
  {/each}
</div>

{#snippet entry(item: MoreItem)}
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
  {:else if item.icon || item.hint}
    <!-- The hint is read as the title, not as part of the name, so the
         item is still called by its label. -->
    <button
      type="button"
      class="mi rich"
      class:danger={item.danger}
      role="menuitem"
      aria-label={item.hint ? item.label : undefined}
      disabled={item.disabled}
      title={item.title ?? item.hint}
      onclick={() => choose(item)}
    >
      {#if item.icon}<Icon name={item.icon} size={14} />{/if}
      <span class="mtext">
        <span>{item.label}</span>
        {#if item.hint}<small>{item.hint}</small>{/if}
      </span>
    </button>
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
{/snippet}

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
  /* An item with a glyph and a line under it (the Invite menu). */
  .mi.rich {
    display: flex;
    align-items: flex-start;
    justify-content: flex-start;
    gap: 10px;
    white-space: normal;
  }
  .mi.rich :global(.icon) {
    flex: none;
    margin-top: 2px;
    color: var(--fg-muted);
  }
  .mtext {
    display: flex;
    flex-direction: column;
    gap: 1px;
    min-width: 0;
  }
  .mtext small {
    font: 400 11.5px/1.35 var(--font-ui);
    color: var(--fg-muted);
  }
  .msep {
    height: 1px;
    margin: 5px 4px;
    background: var(--border);
  }
  .mhead {
    padding: 6px 10px 2px;
    font: 700 9.5px/1 var(--font-mono);
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
</style>
