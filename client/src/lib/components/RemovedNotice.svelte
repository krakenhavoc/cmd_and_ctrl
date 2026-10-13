<script lang="ts">
  // The one place a removed player is told why (#2785).
  //
  // Mounted on the table, where the socket closes with the reason, and on
  // the lobby, where the next request comes back 401. Both feed the
  // `removedNotice` store, so the message is the same wherever the player
  // happens to be, and it stays until they dismiss it: a one-second
  // toast over a screen that just changed under them would be missed.
  import { removedNotice } from "../session";
  import Icon from "./Icon.svelte";

  interface Props {
    /** Show the way out. The lobby is already there, so it leaves this off. */
    withLink?: boolean;
  }
  let { withLink = false }: Props = $props();
</script>

{#if $removedNotice}
  <div class="removed-notice" role="alert">
    <Icon name="x" size={14} />
    <span class="text">{$removedNotice}</span>
    {#if withLink}
      <a class="go" href="#/lobby">Back to the lobby</a>
    {/if}
    <button type="button" class="dismiss" onclick={() => removedNotice.set("")}>dismiss</button>
  </div>
{/if}

<style>
  .removed-notice {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 60;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 16px;
    background: var(--surface-raised, #1d1f26);
    border-bottom: 2px solid var(--danger, #e5484d);
    color: var(--fg, #eee);
    font-size: 13px;
  }
  .text {
    flex: 1 1 auto;
    min-width: 0;
  }
  .go,
  .dismiss {
    flex: 0 0 auto;
    font: inherit;
    font-size: 12px;
    padding: 4px 10px;
    border-radius: 8px;
    border: 1px solid var(--border-strong, #444);
    background: transparent;
    color: var(--fg, #eee);
    cursor: pointer;
    text-decoration: none;
  }
  .go:hover,
  .dismiss:hover {
    border-color: var(--accent, #7aa2ff);
  }
</style>
