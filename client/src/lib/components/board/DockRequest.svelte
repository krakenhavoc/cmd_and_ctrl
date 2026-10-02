<script lang="ts">
  // DockRequest renders nothing. It opens a request in the action dock
  // (lib/dock.ts, ADR 0111) for as long as it is mounted, the way
  // <ModalLayer /> registers a modal:
  //
  //     {#if blocksOwed}
  //       <DockRequest request={blockRequest} />
  //     {/if}
  //
  // A change to `request` updates the open request in place and keeps
  // its place in the order; unmounting closes it.
  import { onDestroy } from "svelte";
  import { pushDockRequest, type DockHandle, type DockRequest } from "../../dock";

  const { request }: { request: DockRequest } = $props();

  let handle: DockHandle | null = null;
  $effect(() => {
    const r = request;
    if (handle) handle.update(r);
    else handle = pushDockRequest(r);
  });
  onDestroy(() => {
    handle?.close();
    handle = null;
  });
</script>
