<script lang="ts">
  // Shown while a lazily loaded route's chunk downloads (App.svelte, #3345),
  // and when it cannot (offline with a cold cache, or a deploy replaced the
  // chunk): a reload fetches the current shell, which is the right fix.
  let { failed = false }: { failed?: boolean } = $props();
</script>

<div class="route-loading" role="status" aria-live="polite">
  {#if failed}
    <p>
      This page could not load.
      <button type="button" onclick={() => location.reload()}>Reload</button>
    </p>
  {:else}
    <p>Loading…</p>
  {/if}
</div>

<style>
  .route-loading {
    display: grid;
    place-items: center;
    min-height: 40vh;
    color: var(--text-muted, inherit);
  }
</style>
