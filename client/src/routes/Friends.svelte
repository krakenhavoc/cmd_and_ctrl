<script lang="ts">
  // Friends — #/friends (ADR 0213 §9, slice 3 of #3313): what the header's
  // drawer leaves out. Your friends (Remove friend, Block), the requests
  // you sent (Cancel request), the requests waiting on you, the people you
  // have played with (Block), the people you blocked (Unblock), and the
  // two settings: who can send you friend requests, and who can see what
  // you are doing on CMD & CTRL (stored now; slice 4 shows it to others).
  //
  // Removing a friend and blocking someone ask first, in the row, and say
  // what they end. Nobody is told about either (ADR 0213 §2, §3).
  //
  // Signed-in only (the router sends a signed-out visitor away). A guest
  // seat or the admin token is not a person and gets a line saying so; a
  // server with no friends store (no database) says friends are off.
  import { onMount } from "svelte";
  import Page from "../lib/components/Page.svelte";
  import { fetchTablemates } from "../lib/api";
  import {
    acceptFriend,
    blockPerson,
    cancelFriendRequest,
    fetchBlocks,
    fetchFriends,
    fetchSocialSettings,
    friendErrorLine,
    ignoreFriend,
    PRESENCE_CHOICES,
    putSocialSettings,
    removeFriend,
    REQUESTS_CHOICES,
    sinceLine,
    unblockPerson,
    type FriendEntry,
    type FriendsResponse,
    type PresenceSetting,
    type RequestsSetting,
    type SocialSettings,
  } from "../lib/friends";
  import { L } from "../lib/labels";
  import { signedInUserID } from "../lib/myGames";
  import { togetherLine } from "../lib/playedWith";
  import { LobbyApiError, session } from "../lib/session";
  import type { Tablemate } from "../lib/tablemates";

  const isPerson = $derived(
    signedInUserID($session) !== null && $session?.principal.role !== "admin",
  );

  let lists = $state<FriendsResponse | null>(null);
  let blocks = $state<FriendEntry[]>([]);
  let mates = $state<Tablemate[]>([]);
  let settings = $state<SocialSettings | null>(null);
  let loading = $state(true);
  let error = $state("");
  let off = $state(false);

  // The row asking "are you sure", the one a call is in flight for, and
  // the last thing that happened.
  let confirming = $state<{ id: string; what: "remove" | "block" } | null>(null);
  let busy = $state<string | null>(null);
  let note = $state<{ ok: boolean; text: string } | null>(null);
  let saved = $state("");

  const strangers = $derived(mates.filter((m) => !m.friend));

  async function load(spinner = true): Promise<void> {
    if (spinner) loading = true;
    error = "";
    try {
      const [l, b, m, s] = await Promise.all([
        fetchFriends(),
        fetchBlocks(),
        fetchTablemates(),
        fetchSocialSettings(),
      ]);
      lists = l;
      blocks = b;
      mates = m;
      settings = s;
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 503) off = true;
      else error = friendErrorLine(err);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    if (isPerson) void load();
    else loading = false;
  });

  async function act(id: string, fn: () => Promise<unknown>, done: string): Promise<void> {
    if (busy) return;
    busy = id;
    note = null;
    try {
      await fn();
      confirming = null;
      note = { ok: true, text: done };
      await load(false);
    } catch (err) {
      note = { ok: false, text: friendErrorLine(err) };
    } finally {
      busy = null;
    }
  }

  const doRemove = (e: FriendEntry) =>
    act(e.user_id, () => removeFriend(e.user_id), `${e.display_name} is no longer your friend.`);
  const doBlock = (p: { user_id: string; display_name: string }) =>
    act(p.user_id, () => blockPerson(p.user_id), `You blocked ${p.display_name}.`);
  const doUnblock = (e: FriendEntry) =>
    act(e.user_id, () => unblockPerson(e.user_id), `You unblocked ${e.display_name}.`);
  const doCancel = (e: FriendEntry) =>
    act(
      e.user_id,
      () => cancelFriendRequest(e.user_id),
      `Your request to ${e.display_name} is cancelled.`,
    );
  const doAccept = (e: FriendEntry) =>
    act(e.user_id, () => acceptFriend(e.user_id), `You and ${e.display_name} are friends now.`);
  const doIgnore = (e: FriendEntry) =>
    act(e.user_id, () => ignoreFriend(e.user_id), `Ignored. ${e.display_name} isn’t told.`);

  async function choose(change: Partial<SocialSettings>): Promise<void> {
    saved = "";
    try {
      settings = await putSocialSettings(change);
      saved = "Saved";
    } catch (err) {
      saved = friendErrorLine(err);
    }
  }

  function initial(name: string): string {
    return name.trim().slice(0, 1).toUpperCase() || "?";
  }
</script>

<Page
  eyebrow="Friends"
  title="Friends and blocks"
  lede="Your friends, the requests you sent, the people you blocked, and who can reach you."
>
  {#if !isPerson}
    <p class="card note" role="status">Sign in with Discord to have friends.</p>
  {:else if loading}
    <p class="card note" role="status">Loading…</p>
  {:else if off}
    <p class="card note" role="status">Friends aren’t available on this server.</p>
  {:else if error}
    <div class="card note" role="alert">
      <p>{error}</p>
      <button type="button" class="btn sm" onclick={() => void load()}>Try again</button>
    </div>
  {:else if lists && settings}
    {#if note}
      <p class="flash" class:bad={!note.ok} role={note.ok ? "status" : "alert"}>{note.text}</p>
    {/if}
    <div class="grid">
      <section class="card" aria-labelledby="f-friends">
        <h2 id="f-friends">
          Your friends <span class="count">{lists.friends.length}/{lists.max_friends ?? 200}</span>
        </h2>
        {#if lists.friends.length === 0}
          <p class="empty">
            No friends yet. Add someone from the Friends button in the header, or from their seat at
            your table.
          </p>
        {:else}
          <ul class="rows">
            {#each lists.friends as f (f.user_id)}
              <li>
                {@render who(f.display_name, f.avatar_url, sinceLine("friend", f.since))}
                {#if confirming?.id === f.user_id}
                  {@render confirm(f)}
                {:else}
                  <span class="acts">
                    <button
                      type="button"
                      class="btn sm ghost"
                      disabled={busy !== null}
                      onclick={() => (confirming = { id: f.user_id, what: "remove" })}
                      >{L.removeFriend}</button
                    >
                    <button
                      type="button"
                      class="btn sm ghost"
                      disabled={busy !== null}
                      onclick={() => (confirming = { id: f.user_id, what: "block" })}
                      >{L.blockPerson}</button
                    >
                  </span>
                {/if}
              </li>
            {/each}
          </ul>
        {/if}
      </section>

      <section class="card" aria-labelledby="f-requests">
        <h2 id="f-requests">Requests</h2>
        {#if lists.incoming.length === 0 && lists.outgoing.length === 0}
          <p class="empty">No requests waiting. A request lasts 30 days, and carries no message.</p>
        {/if}
        {#if lists.incoming.length > 0}
          <h3>Waiting on you</h3>
          <ul class="rows">
            {#each lists.incoming as e (e.user_id)}
              <li>
                {@render who(e.display_name, e.avatar_url, sinceLine("incoming", e.since))}
                {#if confirming?.id === e.user_id}
                  {@render confirm(e)}
                {:else}
                  <span class="acts">
                    <button
                      type="button"
                      class="btn sm primary"
                      disabled={busy !== null}
                      onclick={() => void doAccept(e)}>{L.acceptFriend}</button
                    >
                    <button
                      type="button"
                      class="btn sm ghost"
                      disabled={busy !== null}
                      onclick={() => void doIgnore(e)}>{L.ignoreFriend}</button
                    >
                    <button
                      type="button"
                      class="btn sm ghost"
                      disabled={busy !== null}
                      onclick={() => (confirming = { id: e.user_id, what: "block" })}
                      >{L.blockPerson}</button
                    >
                  </span>
                {/if}
              </li>
            {/each}
          </ul>
        {/if}
        {#if lists.outgoing.length > 0}
          <h3>
            You sent <span class="count">{lists.outgoing.length}/{lists.max_outgoing ?? 20}</span>
          </h3>
          <ul class="rows">
            {#each lists.outgoing as e (e.user_id)}
              <li>
                {@render who(e.display_name, e.avatar_url, sinceLine("outgoing", e.since))}
                <span class="acts">
                  <button
                    type="button"
                    class="btn sm ghost"
                    disabled={busy !== null}
                    onclick={() => void doCancel(e)}>{L.cancelFriendRequest}</button
                  >
                </span>
              </li>
            {/each}
          </ul>
        {/if}
      </section>

      <section class="card" aria-labelledby="f-settings">
        <h2 id="f-settings">Settings</h2>
        <h3>{L.whoCanRequest}</h3>
        <div class="choices" role="radiogroup" aria-label={L.whoCanRequest}>
          {#each REQUESTS_CHOICES as c (c.id)}
            <label class="choice">
              <input
                type="radio"
                name="requests"
                value={c.id}
                checked={settings.requests === c.id}
                onchange={() => void choose({ requests: c.id as RequestsSetting })}
              />
              {c.label}
            </label>
          {/each}
        </div>
        <h3>{L.whoCanSee}</h3>
        <div class="choices" role="radiogroup" aria-label={L.whoCanSee}>
          {#each PRESENCE_CHOICES as c (c.id)}
            <label class="choice">
              <input
                type="radio"
                name="presence"
                value={c.id}
                checked={settings.presence === c.id}
                onchange={() => void choose({ presence: c.id as PresenceSetting })}
              />
              {c.label}
            </label>
          {/each}
        </div>
        <p class="hint">
          Where you are on CMD & CTRL isn’t shown to anyone yet; this is who will see it when it is.
        </p>
        <p class="saved" role="status">{saved}</p>
      </section>

      <section class="card" aria-labelledby="f-played">
        <h2 id="f-played">People you’ve played with</h2>
        {#if strangers.length === 0}
          <p class="empty">Nobody here who isn’t already your friend.</p>
        {:else}
          <ul class="rows">
            {#each strangers as m (m.user_id)}
              <li>
                {@render who(m.display_name, m.avatar_url, togetherLine(m))}
                {#if confirming?.id === m.user_id}
                  {@render confirm(m)}
                {:else}
                  <span class="acts">
                    <button
                      type="button"
                      class="btn sm ghost"
                      disabled={busy !== null}
                      onclick={() => (confirming = { id: m.user_id, what: "block" })}
                      >{L.blockPerson}</button
                    >
                  </span>
                {/if}
              </li>
            {/each}
          </ul>
        {/if}
      </section>

      <section class="card" aria-labelledby="f-blocked">
        <h2 id="f-blocked">Blocked</h2>
        {#if blocks.length === 0}
          <p class="empty">
            You haven’t blocked anyone. Someone you block isn’t told, leaves your lists, and can’t
            send you a request or an invite.
          </p>
        {:else}
          <ul class="rows">
            {#each blocks as b (b.user_id)}
              <li>
                {@render who(b.display_name, b.avatar_url, sinceLine("blocked", b.since))}
                <span class="acts">
                  <button
                    type="button"
                    class="btn sm ghost"
                    disabled={busy !== null}
                    onclick={() => void doUnblock(b)}>{L.unblockPerson}</button
                  >
                </span>
              </li>
            {/each}
          </ul>
        {/if}
      </section>
    </div>
  {/if}
</Page>

{#snippet who(name: string, avatar: string | undefined, sub: string)}
  <span class="who">
    {#if avatar}
      <img class="avatar" src={avatar} alt="" width="32" height="32" />
    {:else}
      <span class="avatar fallback" aria-hidden="true">{initial(name)}</span>
    {/if}
    <span class="text">
      <span class="name">{name}</span>
      <span class="sub">{sub}</span>
    </span>
  </span>
{/snippet}

{#snippet confirm(p: { user_id: string; display_name: string })}
  <span class="confirm" role="group" aria-label="confirm">
    <span class="ask">
      {#if confirming?.what === "remove"}
        Remove {p.display_name}? They aren’t told.
      {:else}
        Block {p.display_name}? It ends any friendship and request between you. They aren’t told.
      {/if}
    </span>
    <span class="acts">
      <button
        type="button"
        class="btn sm danger"
        disabled={busy !== null}
        onclick={() =>
          void (confirming?.what === "remove" ? doRemove(p as FriendEntry) : doBlock(p))}
      >
        {confirming?.what === "remove" ? L.removeFriend : L.blockPerson}
      </button>
      <button
        type="button"
        class="btn sm ghost"
        disabled={busy !== null}
        onclick={() => (confirming = null)}>Keep</button
      >
    </span>
  </span>
{/snippet}

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 18px;
    align-items: start;
  }
  @media (max-width: 819px) {
    .grid {
      grid-template-columns: minmax(0, 1fr);
    }
  }
  h2 {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 12px;
    font-family: var(--font-display);
    font-size: 1.05rem;
    font-weight: 600;
  }
  h3 {
    margin: 16px 0 8px;
    font-size: 0.72rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  h2 + h3 {
    margin-top: 0;
  }
  .count {
    font-family: var(--font-mono);
    font-size: 0.72rem;
    font-weight: 500;
    padding: 1px 8px;
    border-radius: 999px;
    background: var(--surface-raised);
    color: var(--fg-muted);
    letter-spacing: 0;
  }
  .empty,
  .hint {
    margin: 0;
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .hint {
    margin-top: 12px;
  }
  .rows {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 2px;
  }
  .rows li {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 8px 12px;
    padding: 8px 0;
    border-top: 1px solid var(--border);
  }
  .rows li:first-child {
    border-top: 0;
  }
  .who {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
    flex: 1 1 180px;
  }
  .avatar {
    flex: 0 0 auto;
    width: 32px;
    height: 32px;
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
    font-size: 0.85rem;
  }
  .text {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .name {
    font-weight: 600;
    font-size: 0.92rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .sub {
    font-size: 0.75rem;
    color: var(--fg-muted);
    overflow-wrap: anywhere;
  }
  .acts {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .confirm {
    display: flex;
    flex-direction: column;
    gap: 8px;
    flex: 1 1 100%;
  }
  .ask {
    font-size: 0.82rem;
    line-height: 1.45;
  }
  .choices {
    display: grid;
    gap: 6px;
  }
  .choice {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.88rem;
    min-height: 32px;
    cursor: pointer;
  }
  .saved {
    min-height: 1.2em;
    margin: 8px 0 0;
    font-size: 0.8rem;
    color: var(--mint);
  }
  .flash {
    margin: 0 0 14px;
    font-size: 0.88rem;
    color: var(--mint);
  }
  .flash.bad {
    color: var(--fg);
  }
  .note p {
    margin: 0 0 10px;
  }
  /* The confirm's action: a removal or a block, so the danger tone. */
  .btn.danger {
    border-color: var(--danger-line);
    background: var(--danger-soft);
    color: var(--fg);
  }
  .btn.danger:hover:not(:disabled) {
    border-color: var(--danger);
  }
</style>
