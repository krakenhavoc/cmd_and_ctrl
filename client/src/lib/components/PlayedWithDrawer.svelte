<script lang="ts">
  // PlayedWithDrawer — your friends and the people you have shared a
  // table with, in a drawer from the site header (ADR 0213 §1 and §9).
  //
  // Slice 2 was the Played with section. Slice 3 adds, above it,
  // Requests (people who asked you: Accept or Ignore) and Friends (each
  // with Invite), and gives each Played with row Add friend beside
  // Invite. A friend shows once, under Friends, and so does someone whose
  // request is waiting on you, under Requests. The link at the foot goes
  // to #/friends, which manages requests sent, blocks and the settings.
  //
  // Fetched when the drawer opens, and again after an answer changes
  // the lists: no polling, no heartbeat (that is slice 4). Opening it
  // again is a new fetch, because the header unmounts the drawer on
  // close. A server with no friends store (no database) answers 503 to
  // GET /me/friends; the drawer is then Played with alone, as in slice 2.
  //
  // Invite opens a chooser of your tables still in the lobby with an
  // empty seat, plus New table, which makes a table from your last setup
  // and invites into it. Delivery depends on the server: with a bot token
  // the invite goes as a Discord DM (POST /games/{id}/invites/dm),
  // without one (every environment but production) the row says Copy
  // invite and puts the table's link on the clipboard.
  //
  // At tablet width and below the drawer covers the page; above it it
  // sits at the right edge with a scrim over the rest.
  import { onMount, tick } from "svelte";
  import { createGame, fetchTablemates, getGame, listGames, sendInviteDM } from "../api";
  import type { GameMeta } from "../api";
  import { community } from "../community";
  import {
    acceptFriend,
    askFriend,
    fetchFriends,
    FRIENDS_HASH,
    friendControlText,
    friendErrorLine,
    friendStateOf,
    ignoreFriend,
    outcomeLine,
    sinceLine,
    stateAfter,
    type FriendEntry,
    type FriendState,
    type FriendsResponse,
  } from "../friends";
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

  /** Whoever a row is about: a tablemate or a friend. */
  interface Person {
    user_id: string;
    display_name: string;
    avatar_url?: string;
  }

  const dm = $derived($community.dm_invites === true);

  let drawerEl = $state<HTMLElement | null>(null);
  let mates = $state<Tablemate[]>([]);
  let social = $state<FriendsResponse | null>(null);
  let tables = $state<GameMeta[]>([]);
  let loading = $state(true);
  let loadError = $state("");

  // The row whose chooser is open, the one a call is in flight for, and
  // the last thing that happened to a row.
  let openFor = $state<string | null>(null);
  let busy = $state<string | null>(null);
  let result = $state<{ userID: string; ok: boolean; text: string; link?: string } | null>(null);

  const offered = $derived(inviteTables(tables, $session));
  const incoming = $derived(social?.incoming ?? []);
  const friends = $derived(social?.friends ?? []);
  // Each person once: a friend under Friends, an asker under Requests.
  const asking = $derived(new Set(incoming.map((e) => e.user_id)));
  const rest = $derived(mates.filter((m) => !m.friend && !asking.has(m.user_id)));

  async function load(spinner = true): Promise<void> {
    if (spinner) loading = true;
    loadError = "";
    try {
      // Your tables are for the chooser; a failure there leaves New table.
      // A server with no friends store leaves Played with alone.
      const [list, f, games] = await Promise.all([
        fetchTablemates(),
        fetchFriends().catch(() => null),
        spinner ? listGames().catch(() => [] as GameMeta[]) : Promise.resolve(tables),
      ]);
      mates = list;
      social = f;
      tables = games;
    } catch (err) {
      loadError =
        err instanceof LobbyApiError ? err.message : "Couldn’t load your friends and Played with.";
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void load();
    void tick().then(() => drawerEl?.querySelector<HTMLElement>(".close")?.focus());
  });

  function toggle(p: Person): void {
    result = null;
    openFor = openFor === p.user_id ? null : p.user_id;
  }

  // The link to a table: the one a just-made table came back with, or the
  // one the server serves to its creator and its seats.
  async function tableLink(g: GameMeta): Promise<string | null> {
    const token = g.invite_token || (await getGame(g.id)).invite_token;
    return token ? inviteURL(g.id, token) : null;
  }

  async function deliver(p: Person, choice: GameMeta | "new"): Promise<void> {
    if (busy) return;
    busy = p.user_id;
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
        await sendInviteDM(table.id, p.user_id);
        result = {
          userID: p.user_id,
          ok: true,
          text: inviteDone(p.display_name, table.name, true),
        };
      } else {
        const link = await tableLink(table);
        if (!link) {
          result = {
            userID: p.user_id,
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
            ? { userID: p.user_id, ok: true, text: inviteDone(p.display_name, table.name, false) }
            : {
                userID: p.user_id,
                ok: false,
                text: "Couldn’t copy to the clipboard. Copy the link by hand:",
                link,
              };
        }
      }
      openFor = null;
    } catch (err) {
      result = {
        userID: p.user_id,
        ok: false,
        text: err instanceof LobbyApiError ? err.message : "Couldn’t send that invite. Try again.",
      };
    } finally {
      busy = null;
    }
  }

  // Add friend on a Played with row, or Accept where they asked you.
  async function befriend(m: Tablemate): Promise<void> {
    if (busy) return;
    const state = friendStateOf(m);
    busy = m.user_id;
    result = null;
    try {
      const res = state === "incoming" ? await acceptFriend(m.user_id) : await askFriend(m.user_id);
      const next: FriendState = stateAfter(res.outcome);
      result = { userID: m.user_id, ok: true, text: outcomeLine(res.outcome, m.display_name) };
      if (next === "friend") {
        await load(false);
      } else {
        mates = mates.map((x) => (x.user_id === m.user_id ? { ...x, request: "outgoing" } : x));
      }
    } catch (err) {
      result = { userID: m.user_id, ok: false, text: friendErrorLine(err) };
    } finally {
      busy = null;
    }
  }

  async function answer(e: FriendEntry, yes: boolean): Promise<void> {
    if (busy) return;
    busy = e.user_id;
    result = null;
    try {
      if (yes) {
        await acceptFriend(e.user_id);
        result = { userID: e.user_id, ok: true, text: outcomeLine("friends", e.display_name) };
        await load(false);
      } else {
        await ignoreFriend(e.user_id);
        if (social) {
          social = { ...social, incoming: social.incoming.filter((x) => x.user_id !== e.user_id) };
        }
      }
    } catch (err) {
      result = { userID: e.user_id, ok: false, text: friendErrorLine(err) };
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

  function initial(p: Person): string {
    return p.display_name.trim().slice(0, 1).toUpperCase() || "?";
  }
</script>

<svelte:window onkeydown={onKey} />

<ModalLayer />

<div class="scrim" role="presentation" onclick={onclose}></div>

<div
  class="drawer"
  role="dialog"
  aria-modal="true"
  aria-label={L.friendsDrawer}
  bind:this={drawerEl}
>
  <header class="head">
    <h2>Friends</h2>
    <button type="button" class="ghost close" aria-label="Close" onclick={onclose}>
      <Icon name="x" size={18} />
    </button>
  </header>

  <div class="body">
    {#if loading}
      <p class="note" role="status">Loading…</p>
    {:else if loadError}
      <div class="note error" role="alert">
        <p>{loadError}</p>
        <button type="button" onclick={() => void load()}>Try again</button>
      </div>
    {:else}
      {#if incoming.length > 0}
        <section class="sect" aria-labelledby="fr-requests">
          <h3 id="fr-requests">Requests <span class="count">{incoming.length}</span></h3>
          <ul class="mates">
            {#each incoming as e (e.user_id)}
              <li class="mate asker">
                <div class="line">
                  {@render avatar(e)}
                  <span class="who">
                    <span class="name">{e.display_name}</span>
                    <span class="sub">{sinceLine("incoming", e.since)}</span>
                  </span>
                  <span class="acts">
                    <button
                      type="button"
                      class="primary"
                      disabled={busy !== null}
                      onclick={() => void answer(e, true)}>{L.acceptFriend}</button
                    >
                    <button
                      type="button"
                      class="ghost"
                      disabled={busy !== null}
                      onclick={() => void answer(e, false)}>{L.ignoreFriend}</button
                    >
                  </span>
                </div>
                {@render outcome(e.user_id)}
              </li>
            {/each}
          </ul>
        </section>
      {/if}

      {#if social}
        <section class="sect" aria-labelledby="fr-friends">
          <h3 id="fr-friends">
            Friends{#if friends.length > 0}
              <span class="count">{friends.length}</span>{/if}
          </h3>
          {#if friends.length === 0}
            <p class="note">
              No friends yet. Add someone you’ve played with below, or from their seat at your
              table. They’ll be asked first.
            </p>
          {:else}
            <ul class="mates">
              {#each friends as f (f.user_id)}
                <li class="mate" class:open={openFor === f.user_id}>
                  <div class="line">
                    {@render avatar(f)}
                    <span class="who">
                      <span class="name">{f.display_name}</span>
                      <span class="sub">{sinceLine("friend", f.since)}</span>
                    </span>
                    {@render inviteButton(f)}
                  </div>
                  {@render chooser(f)}
                  {@render outcome(f.user_id)}
                </li>
              {/each}
            </ul>
          {/if}
        </section>
      {/if}

      <section class="sect" aria-labelledby="fr-played">
        <h3 id="fr-played">{L.playedWith}</h3>
        {#if mates.length === 0}
          <p class="note">
            Nobody yet. The people you sit at a table with will show up here, so you can invite them
            again.
          </p>
        {:else if rest.length === 0}
          <p class="note">Everyone you’ve played with is above.</p>
        {:else}
          <ul class="mates">
            {#each rest as mate (mate.user_id)}
              {@const state = friendStateOf(mate)}
              <li class="mate" class:open={openFor === mate.user_id}>
                <div class="line">
                  {@render avatar(mate)}
                  <span class="who">
                    <span class="name">{mate.display_name}</span>
                    <span class="sub">{togetherLine(mate)}</span>
                  </span>
                  <span class="acts">
                    {#if social}
                      {#if state === "outgoing"}
                        <span class="sent">{friendControlText(state)}</span>
                      {:else}
                        <button
                          type="button"
                          class="add"
                          disabled={busy !== null}
                          onclick={() => void befriend(mate)}
                        >
                          <Icon name="plus" size={13} />
                          {friendControlText(state)}
                        </button>
                      {/if}
                    {/if}
                    {@render inviteButton(mate)}
                  </span>
                </div>
                {@render chooser(mate)}
                {@render outcome(mate.user_id)}
              </li>
            {/each}
          </ul>
        {/if}
      </section>

      {#if social}
        <p class="foot">
          <a class="link-arrow" href={FRIENDS_HASH} onclick={onclose}>{L.manageFriends}</a>
        </p>
      {/if}
    {/if}
  </div>
</div>

{#snippet avatar(p: Person)}
  {#if p.avatar_url}
    <img class="avatar" src={p.avatar_url} alt="" width="36" height="36" />
  {:else}
    <span class="avatar fallback" aria-hidden="true">{initial(p)}</span>
  {/if}
{/snippet}

{#snippet inviteButton(p: Person)}
  <button
    type="button"
    class="invite"
    class:primary={openFor === p.user_id}
    aria-expanded={openFor === p.user_id}
    aria-controls="choose-{p.user_id}"
    disabled={busy !== null}
    onclick={() => toggle(p)}
  >
    {#if dm}Invite{:else}{L.copyInvite}{/if}
  </button>
{/snippet}

{#snippet chooser(p: Person)}
  {#if openFor === p.user_id}
    <div
      class="choose"
      id="choose-{p.user_id}"
      role="group"
      aria-label={inviteHeading(p.display_name, dm)}
    >
      <p class="ask">{inviteHeading(p.display_name, dm)}</p>
      <ul class="choices">
        {#each offered as g (g.id)}
          <li>
            <button type="button" disabled={busy !== null} onclick={() => void deliver(p, g)}>
              {tableChoiceLabel(g)}
            </button>
          </li>
        {/each}
        <li>
          <button type="button" disabled={busy !== null} onclick={() => void deliver(p, "new")}>
            <Icon name="plus" size={14} /> New table
          </button>
        </li>
      </ul>
      {#if offered.length === 0}
        <p class="hint">
          You have no table in the lobby with an empty seat. New table makes one from your last
          setup.
        </p>
      {/if}
    </div>
  {/if}
{/snippet}

{#snippet outcome(userID: string)}
  {#if result && result.userID === userID}
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
{/snippet}

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
    width: min(420px, 100vw);
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
    padding: 6px 18px 24px;
  }
  .sect {
    margin-top: 14px;
  }
  h3 {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 8px;
    font-size: 0.72rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  .count {
    font-family: var(--font-mono);
    font-size: 0.7rem;
    padding: 1px 7px;
    border-radius: 999px;
    background: var(--surface-raised);
    color: var(--fg);
    letter-spacing: 0;
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
  .mate.asker {
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
  .acts {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .acts button,
  .invite {
    flex: 0 0 auto;
    min-height: 36px;
  }
  .add {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }
  .sent {
    font-family: var(--font-mono);
    font-size: 0.68rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--fg-muted);
    padding: 0 4px;
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
  .foot {
    margin: 18px 0 0;
    font-size: 0.85rem;
  }
</style>
