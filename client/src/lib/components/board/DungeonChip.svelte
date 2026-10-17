<script lang="ts">
  // DungeonChip — a player's dungeon on the player panel (ADR 0203,
  // #4096, CR 309). The chip names the dungeon and the room the venture
  // marker is in ("Undercity · Forge"); its card lists every room in
  // printed order, the current one marked, the rooms the next venture
  // may enter marked as next, and what each room does. A player who has
  // completed a dungeon and owns none right now still gets the chip, with
  // the completed count.
  //
  // It opens the way EmblemLevelChip does: hover or keyboard focus while
  // it lasts, a click, tap or long-press pins it, a second press, Escape
  // or a press elsewhere closes it. The card is a visual copy of the
  // chip's description (dungeon.ts), so it is aria-hidden; it is
  // portalled to <body> with position: fixed so the panel rail's clip
  // cannot cut it.

  import type { DungeonView } from "../../protocol";
  import {
    completedPhrase,
    dungeonChipName,
    dungeonDescription,
    dungeonRoomLines,
    currentRoom,
  } from "../../dungeon";
  import Icon from "../Icon.svelte";

  interface Props {
    dungeon?: DungeonView;
    completed?: string[];
  }
  const { dungeon, completed = [] }: Props = $props();

  const uid = $props.id();
  const descID = `${uid}-desc`;

  const lines = $derived(dungeon ? dungeonRoomLines(dungeon) : []);
  const room = $derived(dungeon ? currentRoom(dungeon) : undefined);
  const name = $derived(dungeon ? dungeonChipName(dungeon) : completedPhrase(completed));
  const description = $derived(dungeonDescription(dungeon, completed));
  const done = $derived(completedPhrase(completed));

  let hovered = $state(false);
  let focused = $state(false);
  let pinned = $state(false);
  const open = $derived(hovered || focused || pinned);

  let chip: HTMLButtonElement | undefined = $state();
  let place = $state<{ left: number; top?: number; bottom?: number }>({ left: 8 });
  const CARD_MAX = 320;

  function measure(): void {
    if (!chip || typeof window === "undefined") return;
    const r = chip.getBoundingClientRect();
    const vw = window.innerWidth || document.documentElement.clientWidth || 0;
    const vh = window.innerHeight || document.documentElement.clientHeight || 0;
    const width = Math.min(CARD_MAX, Math.max(0, vw - 16));
    const left = Math.max(8, Math.min(r.left + r.width / 2 - width / 2, vw - width - 8));
    place = r.top > vh / 2 ? { left, bottom: vh - r.top + 6 } : { left, top: r.bottom + 6 };
  }

  $effect(() => {
    if (!open) return;
    measure();
    const onMove = () => measure();
    window.addEventListener("resize", onMove);
    window.addEventListener("scroll", onMove, true);
    return () => {
      window.removeEventListener("resize", onMove);
      window.removeEventListener("scroll", onMove, true);
    };
  });

  function portal(node: HTMLElement): { destroy(): void } {
    document.body.appendChild(node);
    return {
      destroy() {
        node.remove();
      },
    };
  }

  function onPointerEnter(e: PointerEvent): void {
    if (e.pointerType === "mouse" || e.pointerType === "pen") hovered = true;
  }
  function onFocus(): void {
    let keyboard = true;
    try {
      keyboard = chip?.matches(":focus-visible") ?? true;
    } catch {
      // No :focus-visible: treat as keyboard.
    }
    focused = keyboard;
  }
  function onClick(e: MouseEvent): void {
    e.stopPropagation();
    pinned = !pinned;
  }
  function onContextMenu(e: MouseEvent): void {
    e.preventDefault();
    e.stopPropagation();
    pinned = true;
  }
  function onKeydown(e: KeyboardEvent): void {
    if (e.key !== "Escape" || !open) return;
    e.stopPropagation();
    pinned = false;
    hovered = false;
    focused = false;
  }
  function onWindowPointerDown(e: PointerEvent): void {
    if (pinned && chip && !chip.contains(e.target as Node)) pinned = false;
  }
</script>

<svelte:window onpointerdown={onWindowPointerDown} />

<button
  type="button"
  class="dungeon"
  bind:this={chip}
  aria-label={name}
  aria-describedby={descID}
  aria-expanded={open}
  onpointerenter={onPointerEnter}
  onpointerleave={() => (hovered = false)}
  onfocus={onFocus}
  onblur={() => {
    focused = false;
  }}
  onclick={onClick}
  oncontextmenu={onContextMenu}
  onkeydown={onKeydown}
>
  <Icon name="door" size={12} />
  {#if dungeon}
    <span class="label">{dungeon.name}</span>
    {#if room}
      <span class="sep" aria-hidden="true">·</span>
      <span class="room">{room.name}</span>
    {/if}
  {/if}
  {#if completed.length > 0}
    <span class="done" aria-hidden="true" title={done}>{completed.length}</span>
  {/if}
</button>
<span id={descID} class="sr-only">{description}</span>

{#if open}
  <div
    class="dungeon-card"
    aria-hidden="true"
    use:portal
    style:left="{place.left}px"
    style:top={place.top !== undefined ? `${place.top}px` : null}
    style:bottom={place.bottom !== undefined ? `${place.bottom}px` : null}
    style:width="min({CARD_MAX}px, calc(100vw - 16px))"
  >
    <div class="head">
      <Icon name="door" size={14} />
      <span class="title">{dungeon ? dungeon.name : "Dungeons"}</span>
      {#if done}<span class="count">{done}</span>{/if}
    </div>
    {#if dungeon}
      <ol class="rooms">
        {#each lines as line (line.key)}
          <li class="room-line" class:current={line.current} class:next={line.next}>
            <span class="mark">
              {#if line.current}<Icon
                  name="dot"
                  size={11}
                  strokeWidth={3}
                />{:else if line.next}<Icon name="chevronRight" size={11} strokeWidth={2.25} />{/if}
            </span>
            <span class="body">
              <span class="name">{line.name}</span>
              {#if line.current}<span class="when">you are here</span>{:else if line.next}<span
                  class="when">next</span
                >{/if}
              <span class="text">{line.text}</span>
              {#if line.leadsTo.length > 0}
                <span class="leads">Leads to {line.leadsTo.join(" or ")}</span>
              {/if}
            </span>
          </li>
        {/each}
      </ol>
    {/if}
  </div>
{/if}

<style>
  .dungeon {
    font: inherit;
    font-size: 11px;
    line-height: 1;
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 7px;
    border-radius: 999px;
    background: color-mix(in srgb, var(--magenta) 14%, transparent);
    border: 1px solid color-mix(in srgb, var(--magenta) 50%, transparent);
    color: var(--magenta);
    max-width: 100%;
    min-width: 0;
    white-space: nowrap;
    cursor: pointer;
    box-shadow: none;
    touch-action: manipulation;
    -webkit-tap-highlight-color: transparent;
  }
  .dungeon:hover,
  .dungeon[aria-expanded="true"] {
    background: color-mix(in srgb, var(--magenta) 26%, transparent);
  }
  .dungeon:focus-visible {
    outline: 2px solid var(--magenta);
    outline-offset: 1px;
  }
  .label,
  .room {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .room {
    font-weight: 700;
  }
  .sep {
    opacity: 0.7;
  }
  .done {
    display: inline-grid;
    place-items: center;
    min-width: 15px;
    height: 15px;
    padding: 0 3px;
    box-sizing: border-box;
    border-radius: 999px;
    background: var(--magenta);
    color: var(--bg, #111);
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 800;
    font-variant-numeric: tabular-nums;
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
  .dungeon-card {
    position: fixed;
    z-index: 1000;
    box-sizing: border-box;
    padding: 8px 10px 9px;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    background: var(--surface-raised);
    color: var(--fg);
    box-shadow: var(--shadow);
    font-family: var(--font-ui);
    font-size: 12px;
    line-height: 1.35;
    pointer-events: none;
  }
  .head {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--magenta);
    margin-bottom: 6px;
  }
  .title {
    font-family: var(--font-display);
    font-weight: 700;
    font-size: 13px;
  }
  .count {
    margin-left: auto;
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--fg-muted);
  }
  .rooms {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  .room-line {
    display: flex;
    align-items: flex-start;
    gap: 7px;
  }
  .mark {
    flex: 0 0 auto;
    display: inline-grid;
    place-items: center;
    width: 16px;
    height: 16px;
    margin-top: 1px;
    box-sizing: border-box;
    border-radius: 50%;
    border: 1px dashed var(--fg-dim);
    color: var(--fg-muted);
  }
  .current .mark {
    border: none;
    background: var(--magenta);
    color: var(--bg, #111);
  }
  .next .mark {
    border-style: solid;
    border-color: var(--magenta);
    color: var(--magenta);
  }
  .body {
    min-width: 0;
    display: flex;
    flex-direction: column;
  }
  .name {
    font-weight: 700;
  }
  .when {
    display: block;
    font-weight: 400;
    font-family: var(--font-mono);
    font-size: 9.5px;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  .room-line:not(.current):not(.next) .text {
    color: var(--fg-muted);
  }
  .leads {
    font-size: 10.5px;
    color: var(--fg-muted);
  }
</style>
