<script lang="ts">
  // ModalLayer renders nothing. It exists so a modal can say "I own
  // the keyboard now" by mounting one line inside its own {#if}:
  //
  //     {#if card && offer}
  //       <ModalLayer />
  //       …the dialog…
  //     {/if}
  //
  // While it is mounted the global shortcut layer stands down. See
  // lib/modalLayers.ts for why this is a registration rather than a
  // list of open-flags the dispatcher checks.
  //
  // ADDING A MODAL? Add this line. The failure mode if you forget is
  // visible (shortcuts keep firing behind your dialog) rather than
  // invisible (a stale flag that disables the keymap forever).
  //
  // ADR 0111 PR 6: an action-dock sheet mounts
  // `<ModalLayer kind="sheet" />`. The global shortcuts stand down for
  // it as for any modal, but the dock's own Enter / Escape handler does
  // not (lib/modalLayers.ts, foreignModalOpen).
  import { onMount } from "svelte";
  import { pushModalLayer, type ModalLayerKind } from "../modalLayers";

  const { kind = "modal" }: { kind?: ModalLayerKind } = $props();

  onMount(() => pushModalLayer(kind));
</script>
