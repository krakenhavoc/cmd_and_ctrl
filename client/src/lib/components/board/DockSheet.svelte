<script lang="ts">
  // DockSheet renders nothing where it is mounted. It opens a SHEET in
  // the action dock (ADR 0111 §3, Delivery PR 6, owner decision 2): a
  // panel that grows up out of the dock, holding the picker's body,
  // with the picker's confirm and cancel in the dock's action bar.
  //
  //     {#if card}
  //       <DockSheet
  //         label="Choose X for Blaze"
  //         title="Choose X for Blaze"
  //         src="{X}{R}"
  //         primary={{ id: "confirm", label: "Cast with X = 3", keyShortcuts: "Enter", cap: "⏎", onPress: confirm }}
  //         secondary={[{ id: "cancel", label: "Cancel", keyShortcuts: "Escape", cap: "Esc", onPress: onCancel }]}
  //       >
  //         …the picker's body: hint, rows, grid…
  //       </DockSheet>
  //     {/if}
  //
  // The body stays the picker's: its state, its validation and its
  // handlers live in the component that mounts this, exactly as they
  // did in its modal, and so does its DOM, which is only MOVED into the
  // dock's panel (see `attach` below). What changes is where it is drawn, and that its
  // footer is gone: the confirm and cancel are `primary` / `secondary`,
  // and a key presses the one that advertises it (`keyShortcuts:
  // "Enter"` / `"Escape"`) through the dock's one key handler. So the
  // picker's own Enter / Escape listener goes too.
  //
  // `label` is the dialog's accessible name: the name the modal had (its
  // heading, without the aria-hidden source tag), because the e2e suite
  // and the tutorial find the picker by it.
  //
  // It also registers a "sheet" modal layer: the global shortcuts stand
  // down while it is open (Space must not pass priority under an open
  // cost picker), and the dock's Enter / Escape do not.
  import { onDestroy, type Snippet } from "svelte";
  import type { DockAction, DockRank, DockRefusal } from "../../dock";
  import DockRequest from "./DockRequest.svelte";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    label: string;
    // "flow" for a picker the player opened (a cost, the auto-tap
    // preview, the attack picker); "choice" for one the game waits on
    // (a pending choice, the mulligan, the discard to hand size).
    rank?: Extract<DockRank, "choice" | "flow">;
    title?: string;
    src?: string;
    count?: string;
    width?: number;
    // A new key is a new question (a minimised sheet comes back up).
    sheetKey?: string;
    primary?: DockAction | null;
    secondary?: DockAction[];
    refusal?: DockRefusal | null;
    children: Snippet;
  }

  const {
    label,
    rank = "flow",
    title,
    src,
    count,
    width,
    sheetKey,
    primary = null,
    secondary = [],
    refusal = null,
    children,
  }: Props = $props();

  // The body is rendered HERE, in the picker's own component tree, and
  // its DOM is moved into the dock's sheet panel while this request is
  // the one the dock draws (`attach`). It is not handed to the dock as a
  // snippet: a snippet rendered inside ActionDock runs in the dock's
  // effect tree, which in Game.svelte updates BEFORE the picker's own
  // `{#if}` tears it down, so a body reading `active.count` threw on the
  // frame its prompt closed (the scry e2e caught it). Rendered here, the
  // body lives and dies with the `{#if}` that guards it, exactly as the
  // modal's did; only its nodes sit in the dock.
  let home: HTMLElement | null = $state(null);
  let content: HTMLElement | null = $state(null);
  //
  // This is a portal, so it moves one element by hand. That is safe for
  // the Svelte runtime: it only ever moves `content`, a whole element
  // whose children Svelte keeps anchored inside it, and puts it back (or
  // removes it, on destroy) itself; Svelte never inserts beside it.
  function attach(host: HTMLElement): () => void {
    const el = content;
    if (!el) return () => {};
    host.appendChild(el);
    return () => {
      // eslint-disable-next-line svelte/no-dom-manipulating -- the portal above
      if (el.parentNode === host) home?.appendChild(el);
    };
  }
  // eslint-disable-next-line svelte/no-dom-manipulating -- the portal above
  onDestroy(() => content?.remove());

  const request = $derived({
    rank,
    label,
    primary,
    secondary,
    refusal,
    sheet: { title: title ?? label, src, count, width, key: sheetKey, attach },
  });
</script>

<ModalLayer kind="sheet" />
<DockRequest {request} />
<div class="sheet-home" hidden bind:this={home}>
  <div class="sheet-content" bind:this={content}>
    {@render children()}
  </div>
</div>

<style>
  .sheet-home {
    display: none;
  }
  /* Moved into the dock's `.sheet-body`; stacks the body as the
     modal's panel did. */
  .sheet-content {
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-width: 0;
  }
</style>
