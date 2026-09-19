<script lang="ts">
  import { onMount } from "svelte";
  import { fetchTablemates, sendInviteDM } from "../api";
  import { LobbyApiError } from "../session";
  import { inviteSentMessage, tablemateSubtitle, type Tablemate } from "../tablemates";
  import Icon from "./Icon.svelte";

  // TablematePicker is ADR 0051 decision 8's invite picker (S34
  // sub-PR 6): the people you have already played with, and a button
  // that has the server DM the chosen one this table's existing
  // invite link.
  //
  // Same shape as YourDecksPicker on purpose — radio list, one
  // primary button, inline result — so a player who has used one
  // already knows how to use this. It renders nothing when there is
  // nobody to offer: an empty list is not a UI, and copying the
  // invite link is always the working fallback beside it.
  //
  // The caller decides whether to mount this at all
  // (tablemates.canInviteTablemates); this component does not
  // re-derive the rule.

  interface Props {
    gameID: string;
    /** Optional hook for the page to note that an invite went out. */
    onSent?: (mate: Tablemate) => void;
  }

  const { gameID, onSent }: Props = $props();

  let mates = $state<Tablemate[]>([]);
  let loaded = $state(false);
  let selected = $state<string>("");
  let busy = $state(false);
  let errorMessage = $state("");
  let sent = $state("");

  const current = $derived(mates.find((m) => m.user_id === selected));

  // One shot on mount: the list is "people I have played with", which
  // cannot change while this panel is open.
  onMount(() => {
    void fetchTablemates()
      .then((list) => {
        mates = list;
        if (!selected && mates.length > 0) selected = mates[0].user_id;
      })
      .catch(() => {
        // Session went stale between render and fetch, or this
        // deployment has no database. Either way: no picker.
        mates = [];
      })
      .finally(() => {
        loaded = true;
      });
  });

  async function invite(): Promise<void> {
    if (!selected || busy) return;
    const mate = current;
    busy = true;
    errorMessage = "";
    sent = "";
    try {
      const res = await sendInviteDM(gameID, selected);
      sent = inviteSentMessage(res, mate?.display_name ?? "");
      if (mate) onSent?.(mate);
    } catch (err) {
      errorMessage = err instanceof LobbyApiError ? err.message : "could not send that invite";
    } finally {
      busy = false;
    }
  }
</script>

{#if loaded && mates.length > 0}
  <div class="tablemates">
    <p class="lede">Invite someone you have played with — they get this table's link by DM.</p>

    <ul class="mates">
      {#each mates as mate (mate.user_id)}
        <li>
          <label class="mate" class:sel={selected === mate.user_id}>
            <input
              type="radio"
              name="tablemate-{gameID}"
              value={mate.user_id}
              bind:group={selected}
            />
            {#if mate.avatar_url}
              <img class="avatar" src={mate.avatar_url} alt="" width="22" height="22" />
            {/if}
            <span class="body">
              <span class="name">{mate.display_name}</span>
              <span class="sub">{tablemateSubtitle(mate)}</span>
            </span>
          </label>
        </li>
      {/each}
    </ul>

    <div class="row-actions">
      {#if sent}
        <p class="success"><Icon name="check" size={13} /> {sent}</p>
      {/if}
      <button class="primary" onclick={invite} disabled={busy || !selected}>
        {busy ? "sending…" : "send invite DM"}
      </button>
    </div>

    {#if errorMessage}
      <pre class="invite-error">{errorMessage}</pre>
    {/if}
  </div>
{/if}

<style>
  /* Deliberately the same weight as YourDecksPicker: a short list of
     people, one button, no chrome. */
  .tablemates {
    margin-top: 10px;
  }
  .lede {
    margin: 0 0 10px;
    font-size: 12.5px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .mates {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 8px;
  }
  .mate {
    display: flex;
    gap: 10px;
    align-items: center;
    padding: 8px 12px;
    border-radius: 8px;
    border: 1px solid var(--border);
    background: var(--surface-sunken);
    cursor: pointer;
  }
  .mate.sel {
    border-color: var(--accent);
    background: var(--accent-soft);
  }
  .mate input {
    flex: 0 0 auto;
  }
  .avatar {
    flex: 0 0 auto;
    border-radius: 50%;
    object-fit: cover;
  }
  .body {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }
  .name {
    font-weight: 600;
    font-size: 13.5px;
    color: var(--fg);
  }
  .sub {
    font-size: 11.5px;
    color: var(--fg-dim);
  }
  .row-actions {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: flex-end;
    margin-top: 10px;
  }
  .success {
    margin: 0;
    margin-right: auto;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: var(--mint);
    font-size: 12.5px;
  }
  .invite-error {
    white-space: pre-wrap;
    margin: 10px 0 0;
    padding: 8px 10px;
    border-radius: 8px;
    background: rgba(255, 107, 107, 0.08);
    border: 1px solid rgba(255, 107, 107, 0.3);
    color: var(--fg);
    font-family: var(--font-ui);
    font-size: 12.5px;
  }
</style>
