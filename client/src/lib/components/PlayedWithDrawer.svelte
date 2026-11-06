<script lang="ts">
  // PlayedWithDrawer — the people you have shared a table with, in a
  // drawer from the site header (ADR 0213 §1 and §9, slice 2).
  //
  // Slice 2 is the Played with section only. Requests and Friends join
  // it in slices 3 and 4 (the drawer's shape is theirs: one scrolling
  // list of sections), so this component owns the frame, the keyboard
  // and the fetch, and the section is the part that will move.
  //
  // Fetched when the drawer opens and never again: no polling, no
  // heartbeat (that is slice 4). Opening it again is a new fetch,
  // because the header unmounts the drawer on close.
  //
  // Each row has one button. It opens a chooser of your tables that are
  // still in the lobby with an empty seat, plus New table, which makes a
  // table from your last setup and invites into it. Delivery depends on
  // the server: with a bot token the invite goes as a Discord DM
  // (POST /games/{id}/invites/dm), without one (every environment but
  // production) the row says Copy invite and puts the table's link on
  // the clipboard.
  //
  // At tablet width and below the drawer covers the page; above it it
  // sits at the right edge with a scrim over the rest.
  import { onMount, tick } from "svelte";
  import { createGame, fetchTablemates, getGame, listGames, sendInviteDM } from "../api";
  import type { GameMeta } from "../api";
  import { community } from "../community";
  import { L } from "../labels";
  import { inviteURL } from "../router";
  import { LobbyApiError, session } from "../session";
  import type { Tablemate } from "../tablemates";
  import {
    inviteDone,
    inviteHeading,
    inviteTables,
    tableChoiceLabel,
    togetherLine,
  } from "../playedWith";
  import { TABLES_CHANGED_EVENT } from "../tablesChanged";
  import Icon from "./Icon.svelte";
  import ModalLayer from "./ModalLayer.svelte";

  interface Props {
    onclose: () => void;
  }
  const { onclose }: Props = $props();

  const dm = $derived($community.dm_invites === true);

  let drawerEl = $state<HTMLElement | null>(null);
  let mates = $state<Tablemate[]>([]);
  let tables = $state<GameMeta[]>([]);
  let loading = $state(true);
  let loadError = $state("");

  // The row whose chooser is open, the one an invite is in flight for,
  // and the last thing that happened to a row.
  let openFor = $state<string | null>(null);
  let busy = $state<string | null>(null);
  let result = $state<{ userID: string; ok: boolean; text: string; link?: string } | null>(null);

  const offered = $derived(inviteTables(tables, $session));

  async function load(): Promise<void> {
    loading = true;
    loadError = "";
    try {
      // Your tables are for the chooser; a failure there leaves New table.
      const [list, games] = await Promise.all([
        fetchTablemates(),
        listGames().catch(() => [] as GameMeta[]),
      ]);
      mates = list;
      tables = games;
    } catch (err) {
      loadError =
        err instanceof LobbyApiError ? err.message : "Couldn’t load the people you’ve played with.";
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void load();
    void tick().then(() => drawerEl?.querySelector<HTMLElement>(".close")?.focus());
  });

  function toggle(mate: Tablemate): void {
    result = null;
    openFor = openFor === mate.user_id ? null : mate.user_id;
  }

  // The link to a table: the one a just-made table came back with, or the
  // one the server serves to its creator and its seats.
  async function tableLink(g: GameMeta): Promise<string | null> {
    const token = g.invite_token || (await getGame(g.id)).invite_token;
    return token ? inviteURL(g.id, token) : null;
  }

  async function deliver(mate: Tablemate, choice: GameMeta | "new"): Promise<void> {
    if (busy) return;
    busy = mate.user_id;
    result = null;
    try {
      let table: GameMeta;
      if (choice === "new") {
        // A blank name lets the server pick one; the last setup brings
        // the settings and bots you usually play with.
        table = await createGame("", { setup: "last" });
        window.dispatchEvent(new Event(TABLES_CHANGED_EVENT));
        tables = [...tables, table];
      } else {
        table = choice;
      }
      if (dm) {
        await sendInviteDM(table.id, mate.user_id);
        result = {
          userID: mate.user_id,
          ok: true,
          text: inviteDone(mate.display_name, table.name, true),
        };
      } else {
        const link = await tableLink(table);
        if (!link) {
          result = {
            userID: mate.user_id,
            ok: false,
            text: `No invite link is held for ${table.name}. Open it on the Tables page and copy one there.`,
          };
        } else {
          let copied = true;
          try {
            await navigator.clipboard.writeText(link);
          } catch {
            copied = false; // insecure origin or an unfocused tab
          }
          result = copied
            ? {
                userID: mate.user_id,
                ok: true,
                text: inviteDone(mate.display_name, table.name, false),
              }
            : {
                userID: mate.user_id,
                ok: false,
                text: "Couldn’t copy to the clipboard. Copy the link by hand:",
                link,
              };
        }
      }
      openFor = null;
    } catch (err) {
      result = {
        userID: mate.user_id,
        ok: false,
        text: err instanceof LobbyApiError ? err.message : "Couldn’t send that invite. Try again.",
      };
    } finally {
      busy = null;
    }
  }

  // Escape closes. Tab stays inside, since the scrim hides the page.
  function onKey(e: KeyboardEvent): void {
    if (e.key === "Escape") {
      e.preventDefault();
      onclose();
      return;
    }
    if (e.key !== "Tab" || !drawerEl) return;
    const items = [
      ...drawerEl.querySelectorAll<HTMLElement>("button:not(:disabled), a[href], input"),
    ];
    if (items.length === 0) return;
    const first = items[0];
    const last = items[items.length - 1];
    const at = document.activeElement;
    if (e.shiftKey && (at === first || !drawerEl.contains(at))) {
      e.preventDefault();
      last.focus();
    } else if (!e.shiftKey && at === last) {
      e.preventDefault();
      first.focus();
    }
  }

  function initial(m: Tablemate): string {
    return m.display_name.trim().slice(0, 1).toUpperCase() || "?";
  }
</script>

<svelte:window onkeydown={onKey} />

<ModalLayer />

<div class="scrim" role="presentation" onclick={onclose}></div>

<div class="drawer" role="dialog" aria-modal="true" aria-label={L.playedWith} bind:this={drawerEl}>
  <header class="head">
    <h2>{L.playedWith}</h2>
    <button type="button" class="ghost close" aria-label="Close" onclick={onclose}>
      <Icon name="x" size={18} />
    </button>
  </header>

  <div class="body">
    <p class="lede">
      People you have shared a table with, newest first. Invite one to a table of yours.
    </p>

    {#if loading}
      <p class="note" role="status">Loading…</p>
    {:else if loadError}
      <div class="note error" role="alert">
        <p>{loadError}</p>
        <button type="button" onclick={() => void load()}>Try again</button>
      </div>
    {:else if mates.length === 0}
      <p class="note">
        Nobody yet. The people you sit at a table with will show up here, so you can invite them
        again.
      </p>
    {:else}
      <ul class="mates">
        {#each mates as mate (mate.user_id)}
          <li class="mate" class:open={openFor === mate.user_id}>
            <div class="line">
              {#if mate.avatar_url}
                <img class="avatar" src={mate.avatar_url} alt="" width="36" height="36" />
              {:else}
                <span class="avatar fallback" aria-hidden="true">{initial(mate)}</span>
              {/if}
              <span class="who">
                <span class="name">{mate.display_name}</span>
                <span class="sub">{togetherLine(mate)}</span>
              </span>
              <button
                type="button"
                class="invite"
                class:primary={openFor === mate.user_id}
                aria-expanded={openFor === mate.user_id}
                aria-controls="choose-{mate.user_id}"
                disabled={busy !== null}
                onclick={() => toggle(mate)}
              >
                {#if dm}Invite{:else}{L.copyInvite}{/if}
              </button>
            </div>

            {#if openFor === mate.user_id}
              <div
                class="choose"
                id="choose-{mate.user_id}"
                role="group"
                aria-label={inviteHeading(mate.display_name, dm)}
              >
                <p class="ask">{inviteHeading(mate.display_name, dm)}</p>
                <ul class="choices">
                  {#each offered as g (g.id)}
                    <li>
                      <button
                        type="button"
                        disabled={busy !== null}
                        onclick={() => void deliver(mate, g)}
                      >
                        {tableChoiceLabel(g)}
                      </button>
                    </li>
                  {/each}
                  <li>
                    <button
                      type="button"
                      disabled={busy !== null}
                      onclick={() => void deliver(mate, "new")}
                    >
                      <Icon name="plus" size={14} /> New table
                    </button>
                  </li>
                </ul>
                {#if offered.length === 0}
                  <p class="hint">
                    You have no table in the lobby with an empty seat. New table makes one from your
                    last setup.
                  </p>
                {/if}
              </div>
            {/if}

            {#if result && result.userID === mate.user_id}
              <p class="result" class:bad={!result.ok} role={result.ok ? "status" : "alert"}>
                {#if result.ok}<Icon name="check" size={13} />{/if}
                {result.text}
              </p>
              {#if result.link}
                <input
                  class="link"
                  type="text"
                  readonly
                  value={result.link}
                  aria-label="Invite link"
                  onfocus={(e) => e.currentTarget.select()}
                />
              {/if}
            {/if}
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>

<style>
  .scrim {
    position: fixed;
    inset: 0;
    z-index: 180;
    background: var(--scrim);
  }
  .drawer {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    z-index: 190;
    width: min(400px, 100vw);
    display: flex;
    flex-direction: column;
    background: var(--surface);
    border-left: 1px solid var(--border-strong);
    box-shadow: 0 0 60px rgba(0, 0, 0, 0.5);
  }
  /* The design floor is a tablet (820px): below it the drawer covers the page. */
  @media (max-width: 820px) {
    .drawer {
      width: 100vw;
      border-left: 0;
    }
  }
  .head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 18px 12px;
    border-bottom: 1px solid var(--border);
  }
  h2 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 1.15rem;
    font-weight: 600;
    color: var(--fg);
  }
  .close {
    padding: 0.4rem;
    min-width: 36px;
    min-height: 36px;
  }
  .body {
    flex: 1 1 auto;
    overflow-y: auto;
    overscroll-behavior: contain;
    padding: 14px 18px 24px;
  }
  .lede {
    margin: 0 0 14px;
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .note {
    margin: 0;
    padding: 12px 14px;
    border-radius: var(--radius);
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .note.error {
    border-color: var(--danger-line);
    color: var(--fg);
  }
  .note.error p {
    margin: 0 0 10px;
  }
  .mates,
  .choices {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .mates {
    display: grid;
    gap: 8px;
  }
  .mate {
    padding: 10px 12px;
    border-radius: var(--radius);
    border: 1px solid var(--border);
    background: var(--surface-sunken);
  }
  .mate.open {
    border-color: var(--accent-line);
  }
  .line {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .avatar {
    flex: 0 0 auto;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    object-fit: cover;
  }
  .avatar.fallback {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: var(--surface-raised);
    color: var(--fg-muted);
    font-weight: 600;
  }
  .who {
    flex: 1 1 auto;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .name {
    font-weight: 600;
    font-size: 0.92rem;
    color: var(--fg);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .sub {
    overflow-wrap: anywhere;
    font-size: 0.75rem;
    line-height: 1.35;
    color: var(--fg-muted);
  }
  .invite {
    flex: 0 0 auto;
    min-height: 36px;
  }
  .choose {
    margin-top: 10px;
    padding-top: 10px;
    border-top: 1px solid var(--border);
  }
  .ask {
    margin: 0 0 8px;
    font-size: 0.8rem;
    color: var(--fg-muted);
  }
  .choices {
    display: grid;
    gap: 6px;
  }
  .choices button {
    min-width: 0;
    overflow-wrap: anywhere;
    width: 100%;
    justify-content: flex-start;
    min-height: 36px;
  }
  .hint {
    margin: 8px 0 0;
    font-size: 0.75rem;
    line-height: 1.45;
    color: var(--fg-muted);
  }
  .result {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 10px 0 0;
    font-size: 0.8rem;
    line-height: 1.4;
    color: var(--mint);
  }
  .result.bad {
    color: var(--fg);
  }
  .link {
    width: 100%;
    margin-top: 6px;
    box-sizing: border-box;
    font-size: 0.75rem;
  }
</style>
