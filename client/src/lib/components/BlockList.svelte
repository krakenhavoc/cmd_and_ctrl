<script lang="ts">
  // Who a table has blocked, with an Unblock beside each (#3113, ADR 0075
  // §2.6). On the waiting table it is the host's and the creator's; the
  // admin has it on any table, started or not, from the lobby and the
  // Games view. The server is the gate (GET /games/{id}/blocks answers
  // 403 / 409 to anyone it is not for), so this component only has to
  // stay quiet when it is refused: a 403, 404 or 409 renders nothing.
  //
  // An entry is one removal: an account's display name, or a guest's
  // folded name. Never a Discord id, and the server never sends one.
  import { fetchBlocks, unblock, type BlockEntry } from "../api";
  import {
    BLOCK_LIST_EMPTY,
    BLOCK_LIST_HINT,
    BLOCK_LIST_TITLE,
    UNBLOCK_LABEL,
    blockErrorMessage,
    blockKindLabel,
  } from "../moderation";
  import { LobbyApiError } from "../session";

  interface Props {
    gameID: string;
    /** Bump it to read the list again (the host just removed and blocked someone). */
    refreshKey?: number;
    /** Draw the empty state. Off: an empty list draws nothing, so a waiting table stays quiet. */
    showEmpty?: boolean;
  }
  const { gameID, refreshKey = 0, showEmpty = false }: Props = $props();

  let blocks = $state<BlockEntry[] | null>(null);
  let error = $state("");
  let busy = $state<string | null>(null);
  // The server said this list is not ours to see: draw nothing at all.
  let refused = $state(false);

  async function load(): Promise<void> {
    try {
      blocks = await fetchBlocks(gameID);
      error = "";
      refused = false;
    } catch (err) {
      if (err instanceof LobbyApiError && [403, 404, 409].includes(err.status)) {
        refused = true;
        blocks = null;
        return;
      }
      error = blockErrorMessage(err);
    }
  }

  $effect(() => {
    void refreshKey;
    void gameID;
    void load();
  });

  async function lift(entry: BlockEntry): Promise<void> {
    busy = entry.id;
    error = "";
    try {
      blocks = await unblock(gameID, entry.id);
    } catch (err) {
      error = blockErrorMessage(err);
      if (err instanceof LobbyApiError && err.status === 404) void load();
    } finally {
      busy = null;
    }
  }
</script>

{#if !refused && (error || (blocks && (blocks.length > 0 || showEmpty)))}
  <section class="blocks" aria-label={BLOCK_LIST_TITLE}>
    <h4>{BLOCK_LIST_TITLE}</h4>
    {#if blocks && blocks.length > 0}
      <ul>
        {#each blocks as b (b.id)}
          <li>
            <span class="who">
              <span class="name">{b.label}</span>
              <span class="kind" class:account={b.kind === "account"}>{blockKindLabel(b.kind)}</span
              >
            </span>
            <button
              type="button"
              class="unblock"
              aria-label={`${UNBLOCK_LABEL} ${b.label}`}
              disabled={busy !== null}
              onclick={() => lift(b)}
            >
              {busy === b.id ? "…" : UNBLOCK_LABEL}
            </button>
          </li>
        {/each}
      </ul>
      <p class="hint">{BLOCK_LIST_HINT}</p>
    {:else if blocks}
      <p class="hint">{BLOCK_LIST_EMPTY}</p>
    {/if}
    {#if error}
      <p class="err" role="alert">{error}</p>
    {/if}
  </section>
{/if}

<style>
  .blocks {
    margin-top: 10px;
    padding: 10px 12px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--surface-raised, transparent);
  }
  h4 {
    margin: 0 0 6px;
    font-size: 12px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 5px 0;
    border-top: 1px solid var(--border);
    font-size: 13px;
  }
  li:first-child {
    border-top: none;
  }
  .who {
    display: inline-flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }
  .name {
    font-weight: 600;
    overflow-wrap: anywhere;
  }
  .kind {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.06em;
    padding: 1px 6px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    color: var(--fg-muted);
    white-space: nowrap;
  }
  .kind.account {
    color: var(--accent-strong, var(--fg));
  }
  .unblock {
    flex: none;
    font: inherit;
    font-size: 12px;
    padding: 3px 10px;
    border-radius: 8px;
    border: 1px solid var(--border-strong);
    background: transparent;
    color: var(--fg);
    cursor: pointer;
  }
  .unblock:disabled {
    opacity: 0.5;
    cursor: default;
  }
  .hint {
    margin: 8px 0 0;
    font-size: 11.5px;
    color: var(--fg-muted);
  }
  .err {
    margin: 8px 0 0;
    font-size: 12px;
    color: var(--danger, #e5484d);
  }
</style>
