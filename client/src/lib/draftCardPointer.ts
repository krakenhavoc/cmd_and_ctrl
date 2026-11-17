// draftCardPointer.ts — the Svelte action every draft card wears
// (#5848 §1 and §2): the large preview on hover, focus and touch hold,
// and the pointer-event drag.
//
//   <button use:draftCard={{ card, scope: "pack", drag: { from: "pack", ondrop } }}>
//
// `scope` names the preview host that draws this card's preview
// (components/draft/DraftPreview.svelte). Leave `drag` out for a card
// that cannot move (a finished draft's pool): it still previews.

import type { Action } from "svelte/action";

import type { DraftCard } from "./draft";
import { type DragSpec, cancelDragOf, pressCard } from "./draftDrag.svelte";
import {
  hidePreview,
  preview,
  previewEnter,
  previewLeave,
  previewNow,
} from "./draftPreview.svelte";

export interface DraftCardParams {
  card: DraftCard;
  /** The host that draws the preview. */
  scope: string;
  drag?: DragSpec;
}

/** Keyboard focus only: a card focused by a click needs no preview of its own. */
function focusVisible(el: HTMLElement): boolean {
  try {
    return el.matches(":focus-visible");
  } catch {
    return true; // an engine without the selector: show it, which is harmless
  }
}

export const draftCard: Action<HTMLElement, DraftCardParams> = (node, initial) => {
  let params = initial;

  // A picture dragged out of a button would start the browser's own drag
  // and cancel the pointer's; this drag is ours.
  node.style.setProperty("-webkit-touch-callout", "none");
  node.style.setProperty("user-select", "none");
  const noNative = (ev: Event): void => ev.preventDefault();

  const enter = (ev: PointerEvent): void => {
    if (ev.pointerType === "touch") return;
    previewEnter(params.card, node, params.scope);
  };
  const leave = (ev: PointerEvent): void => {
    if (ev.pointerType === "touch") return;
    previewLeave();
  };
  const down = (ev: PointerEvent): void => {
    pressCard(
      ev,
      node,
      params.drag,
      () => previewNow(params.card, node, params.scope),
      () => {
        if (ev.pointerType === "touch") hidePreview();
      },
    );
  };
  const focusin = (): void => {
    if (focusVisible(node)) previewNow(params.card, node, params.scope);
  };
  const focusout = (): void => {
    if (preview.card?.id === params.card.id) hidePreview();
  };

  node.addEventListener("pointerenter", enter);
  node.addEventListener("pointerleave", leave);
  node.addEventListener("pointerdown", down);
  node.addEventListener("focusin", focusin);
  node.addEventListener("focusout", focusout);
  node.addEventListener("dragstart", noNative);

  return {
    update(next) {
      params = next;
    },
    destroy() {
      node.removeEventListener("pointerenter", enter);
      node.removeEventListener("pointerleave", leave);
      node.removeEventListener("pointerdown", down);
      node.removeEventListener("focusin", focusin);
      node.removeEventListener("focusout", focusout);
      node.removeEventListener("dragstart", noNative);
      cancelDragOf(node);
      if (preview.card?.id === params.card.id) hidePreview();
    },
  };
};
