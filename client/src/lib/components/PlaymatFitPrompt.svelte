<script lang="ts">
  // The best-size prompt (ADR 0128 §11): shown right after a playmat is
  // uploaded or linked, and again from the slot's "Fit to best size"
  // action. It shows the stored image with the part a fit would keep
  // outlined and the rest dimmed, says what size playmats look best at,
  // and offers "Fit to best size" or "Keep as is".
  //
  // The window starts where the server centred it and the person can move
  // it along the one axis being cropped, to choose which part of the art
  // is kept: by dragging, or with the arrow keys once it has focus (Home
  // and End jump to an end, Shift takes bigger steps). Enter accepts and
  // Escape keeps the image as it is.

  import { onMount } from "svelte";
  import type { PlaymatSlot } from "../api";
  import {
    clampCropOrigin,
    cropAxis,
    cropBoxStyle,
    dragCrop,
    fitPromptText,
    fitResultText,
    nudgeCrop,
    playmatSrc,
    type CropOrigin,
  } from "../playmat";

  interface Props {
    mat: PlaymatSlot;
    idealWidth: number;
    idealHeight: number;
    busy?: boolean;
    onaccept: (origin: CropOrigin) => void;
    onkeep: () => void;
  }
  let { mat, idealWidth, idealHeight, busy = false, onaccept, onkeep }: Props = $props();

  // The suggestion is present whenever the prompt is opened; a slot with
  // none has nothing to offer and never reaches here.
  const sg = $derived(mat.suggestion!);
  const axis = $derived(cropAxis(mat.width, mat.height, sg.crop));

  // The window's origin: the server's centred default until it is moved.
  let moved = $state<CropOrigin | null>(null);
  const origin = $derived(
    clampCropOrigin(mat.width, mat.height, sg.crop, moved ?? { x: sg.crop.x, y: sg.crop.y }),
  );
  const box = $derived(cropBoxStyle(mat.width, mat.height, sg.crop, origin));

  const src = $derived(playmatSrc(mat.url));
  let failed = $state(false);

  let frame = $state<HTMLDivElement | null>(null);
  let windowEl = $state<HTMLDivElement | null>(null);
  let drag: { x: number; y: number; from: CropOrigin; id: number } | null = null;

  onMount(() => windowEl?.focus({ preventScroll: true }));

  function onPointerDown(e: PointerEvent): void {
    if (busy || axis === null || e.button !== 0) return;
    drag = { x: e.clientX, y: e.clientY, from: origin, id: e.pointerId };
    (e.currentTarget as HTMLElement).setPointerCapture?.(e.pointerId);
    windowEl?.focus({ preventScroll: true });
    e.preventDefault();
  }
  function onPointerMove(e: PointerEvent): void {
    if (!drag || !frame) return;
    const r = frame.getBoundingClientRect();
    moved = dragCrop(
      mat.width,
      mat.height,
      sg.crop,
      drag.from,
      e.clientX - drag.x,
      e.clientY - drag.y,
      r.width,
      r.height,
    );
  }
  function onPointerUp(e: PointerEvent): void {
    if (!drag) return;
    (e.currentTarget as HTMLElement).releasePointerCapture?.(drag.id);
    drag = null;
  }

  function nudge(move: -1 | 1 | "home" | "end", big: boolean): void {
    moved = nudgeCrop(mat.width, mat.height, sg.crop, origin, move, big ? 0.2 : 0.05);
  }

  function onKey(e: KeyboardEvent): void {
    if (busy) return;
    switch (e.key) {
      case "ArrowLeft":
      case "ArrowUp":
        nudge(-1, e.shiftKey);
        break;
      case "ArrowRight":
      case "ArrowDown":
        nudge(1, e.shiftKey);
        break;
      case "Home":
        nudge("home", false);
        break;
      case "End":
        nudge("end", false);
        break;
      case "Enter":
        onaccept(origin);
        break;
      case "Escape":
        onkeep();
        break;
      default:
        return;
    }
    e.preventDefault();
    e.stopPropagation();
  }

  function onSectionKey(e: KeyboardEvent): void {
    if (e.key === "Escape" && !busy) {
      e.preventDefault();
      onkeep();
    }
  }

  // What a screen reader hears for the window: where it is on its axis.
  const position = $derived.by(() => {
    if (axis === null) return "the whole image is kept";
    const room = axis === "x" ? mat.width - sg.crop.width : mat.height - sg.crop.height;
    const at = axis === "x" ? origin.x : origin.y;
    const pct = room > 0 ? Math.round((at / room) * 100) : 0;
    return axis === "x"
      ? `${pct}% of the way from left to right`
      : `${pct}% of the way from top to bottom`;
  });
</script>

<!-- Escape means "keep as is" from anywhere inside the prompt, not only from the window. -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<section class="prompt" aria-label="Fit playmat to the best size" onkeydown={onSectionKey}>
  <p class="said">{fitPromptText(mat.width, mat.height, idealWidth, idealHeight)}</p>

  <div class="frame" bind:this={frame} style:aspect-ratio={`${mat.width} / ${mat.height}`}>
    {#if src && !failed}
      <img
        {src}
        alt="Your playmat, with the part that would be kept outlined"
        referrerpolicy="no-referrer"
        draggable="false"
        onerror={() => (failed = true)}
      />
    {/if}
    <!-- The kept area. Its huge box-shadow is the dimming of the rest. -->
    <div
      class="window"
      class:movable={axis !== null}
      bind:this={windowEl}
      role="slider"
      tabindex="0"
      aria-label={axis === "y"
        ? "Kept area. Arrow keys move it up and down"
        : "Kept area. Arrow keys move it left and right"}
      aria-orientation={axis === "y" ? "vertical" : "horizontal"}
      aria-valuemin="0"
      aria-valuemax="100"
      aria-valuenow={axis === null
        ? 0
        : Math.round(
            ((axis === "x" ? origin.x : origin.y) /
              Math.max(1, axis === "x" ? mat.width - sg.crop.width : mat.height - sg.crop.height)) *
              100,
          )}
      aria-valuetext={position}
      style:left={box.left}
      style:top={box.top}
      style:width={box.width}
      style:height={box.height}
      onpointerdown={onPointerDown}
      onpointermove={onPointerMove}
      onpointerup={onPointerUp}
      onpointercancel={onPointerUp}
      onkeydown={onKey}
    ></div>
  </div>

  <p class="help">
    {fitResultText(sg)}
    {#if axis !== null}
      Drag the outlined area, or use the arrow keys, to choose which part of the art is kept.
    {/if}
  </p>

  <div class="actions">
    <button type="button" class="primary" disabled={busy} onclick={() => onaccept(origin)}>
      Fit to best size
    </button>
    <button type="button" disabled={busy} onclick={onkeep}>Keep as is</button>
  </div>
  <p class="help keys">Enter fits it. Escape keeps it as it is.</p>
</section>

<style>
  .prompt {
    margin: 12px 0;
    padding: 12px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--surface-sunken);
  }
  .said {
    margin: 0 0 10px;
    font-size: 13px;
    font-weight: 600;
    color: var(--fg);
  }
  .frame {
    position: relative;
    width: 100%;
    max-width: 480px;
    max-height: 60vh;
    overflow: hidden;
    border-radius: 8px;
    background: var(--surface);
    user-select: none;
  }
  .frame img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: fill;
    pointer-events: none;
  }
  .window {
    position: absolute;
    box-sizing: border-box;
    border: 2px solid var(--accent);
    border-radius: 3px;
    /* Everything outside the window is dimmed by its own shadow. */
    box-shadow: 0 0 0 9999px color-mix(in srgb, #000 58%, transparent);
    touch-action: none;
  }
  .window.movable {
    cursor: grab;
  }
  .window.movable:active {
    cursor: grabbing;
  }
  .window:focus-visible {
    outline: 2px solid var(--fg);
    outline-offset: 2px;
  }
  .help {
    color: var(--fg-dim);
    font-size: 11.5px;
    line-height: 1.45;
    margin: 8px 0;
  }
  .keys {
    margin-bottom: 0;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
</style>
