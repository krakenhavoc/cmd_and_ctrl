<script lang="ts">
  // The banner a player sees on their own seat while a bot plays it
  // (ADR 0177 §6). The bot keeps playing until they press Take it back,
  // so a player who only opens the table to look never takes over by
  // accident.
  import { L } from "../../labels";
  import { standInSinceText } from "../../away";

  interface Props {
    since: string | undefined;
    busy: boolean;
    error: string | null;
    onTakeBack: () => void;
  }

  const { since, busy, error, onTakeBack }: Props = $props();
</script>

<div class="stand-in-banner" role="status">
  <span class="text">A bot has been playing your seat since {standInSinceText(since)}.</span>
  <button type="button" class="take-back" disabled={busy} onclick={onTakeBack}
    >{L.takeBackSeat}</button
  >
  {#if error}
    <span class="error" role="alert">{error}</span>
  {/if}
</div>

<style>
  .stand-in-banner {
    /* Above the table and anything staged over it: while a bot plays
       the seat, this is the one control the player has. */
    position: relative;
    z-index: 40;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem 0.75rem;
    padding: 0.45rem 0.75rem;
    border: 1px solid rgba(255, 208, 122, 0.4);
    border-radius: 8px;
    background: rgba(255, 208, 122, 0.12);
    color: var(--fg);
    font-size: 0.9rem;
  }
  .take-back {
    font-weight: 600;
  }
  .error {
    flex-basis: 100%;
    color: var(--danger);
    font-size: 0.85rem;
  }
</style>
