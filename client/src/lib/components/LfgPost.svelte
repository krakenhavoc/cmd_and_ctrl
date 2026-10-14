<script lang="ts">
  // The table's post on the CMD & CTRL Discord's #lfg board (ADR 0161),
  // as a section of the lobby's table settings. Everyone who can see the
  // settings sees whether the table is posted; the host (or an admin)
  // also gets the box. The server enforces all of it: a disabled box is
  // only so the click never happens, and its hint says why.

  import type { GameMeta } from "../api";
  import { lfgBlocked, lfgStatusLine } from "../lfgPost";

  interface Props {
    game: GameMeta;
    // Host or admin (canManageTableFor).
    canManage: boolean;
    ontoggle: (enabled: boolean) => void;
    busy?: boolean;
    error?: string | null;
  }

  const { game, canManage, ontoggle, busy = false, error = null }: Props = $props();

  const status = $derived(lfgStatusLine(game));
  const blocked = $derived(lfgBlocked(game));
</script>

<section class="lfg-post" aria-labelledby={`lfg-title-${game.id}`}>
  <h3 id={`lfg-title-${game.id}`}>#lfg</h3>
  {#if status}
    <p class="lfg-status" data-testid="lfg-status">{status}</p>
  {:else if !canManage}
    <p class="lfg-status" data-testid="lfg-status">This table is not posted to #lfg.</p>
  {/if}

  {#if canManage}
    <label class="lfg-toggle">
      <input
        type="checkbox"
        checked={game.lfg === true}
        disabled={busy || blocked !== null}
        onchange={(e) => ontoggle((e.currentTarget as HTMLInputElement).checked)}
      />
      <span>
        Post to the CMD & CTRL Discord (#lfg)
        <span class="lfg-sum"
          >the table's name, you as host and the open seats, with a Join button</span
        >
      </span>
    </label>
    {#if blocked}
      <p class="lfg-hint" data-testid="lfg-hint">{blocked}</p>
    {/if}
  {/if}
  {#if error}
    <p class="error" role="alert">{error}</p>
  {/if}
</section>

<style>
  .lfg-post {
    margin-top: 1rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--border, rgba(127, 127, 127, 0.3));
  }
  h3 {
    margin: 0 0 0.35rem;
    font-size: 0.95rem;
  }
  .lfg-status,
  .lfg-hint {
    margin: 0.25rem 0;
  }
  .lfg-hint {
    opacity: 0.85;
    font-size: 0.9em;
  }
  .lfg-toggle {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }
  .lfg-sum {
    display: block;
    opacity: 0.75;
    font-size: 0.85em;
  }
</style>
