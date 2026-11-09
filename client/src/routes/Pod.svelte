<script lang="ts">
  // Pod — #/pods/<id>, one pod (ADR 0220 §11, slice 3 of #4880).
  //
  // The name as the title, the owner and the member count as its meta.
  // Members with a Sharing or Not sharing tag; the owner's actions on a
  // row sit in a ⋯ menu (Remove, Make owner). Invite a friend opens a
  // picker of your own friends who are not in the pod or already
  // invited. The owner also has the invite link (shown once, when made),
  // Rename and Delete pod; everyone has Leave pod. Leaving, removing and
  // declining are silent (§3). The sharing switch and the pod's stats
  // are slice 5's: joining shares nothing.
  import { onMount } from "svelte";
  import Page from "../lib/components/Page.svelte";
  import MoreMenu from "../lib/components/MoreMenu.svelte";
  import { fetchFriends, type FriendEntry } from "../lib/friends";
  import { L } from "../lib/labels";
  import { signedInUserID } from "../lib/myGames";
  import {
    cancelPodInvite,
    deletePod,
    fetchPod,
    handOverPod,
    invitePodFriend,
    joinedLine,
    linkLine,
    memberCountLine,
    mintPodLink,
    pickableFriends,
    podErrorLine,
    podLinkURL,
    PODS_HASH,
    removePodMember,
    renamePod,
    turnOffPodLink,
    type PodDetail,
    type PodLink,
    type PodMember,
  } from "../lib/pods";
  import { navigate } from "../lib/router";
  import { LobbyApiError, session } from "../lib/session";

  interface Props {
    id: string;
  }
  const { id }: Props = $props();

  const isPerson = $derived(
    signedInUserID($session) !== null && $session?.principal.role !== "admin",
  );

  let pod = $state<PodDetail | null>(null);
  let loading = $state(true);
  let gone = $state(false);
  let error = $state("");
  let busy = $state(false);
  let note = $state<{ ok: boolean; text: string } | null>(null);

  const isOwner = $derived(pod?.role === "owner");

  // What is open: a member's Remove, Leave, Delete, Rename, the picker.
  let removing = $state<PodMember | null>(null);
  let alsoNewLink = $state(true);
  let leaving = $state(false);
  let deleting = $state(false);
  let renaming = $state(false);
  let newName = $state("");
  let picking = $state(false);
  let friends = $state<FriendEntry[]>([]);
  let pickedFriend = $state("");
  let minted = $state<PodLink | null>(null);
  let copied = $state(false);

  const pickable = $derived(pod ? pickableFriends(friends, pod) : []);
  const full = $derived(pod ? pod.member_count >= pod.max_members : false);

  async function load(spinner = true): Promise<void> {
    if (spinner) loading = true;
    error = "";
    try {
      pod = await fetchPod(id);
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 404) gone = true;
      else error = podErrorLine(err);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    if (isPerson) void load();
    else loading = false;
  });

  async function act(fn: () => Promise<unknown>, done: string, then?: () => void): Promise<void> {
    if (busy) return;
    busy = true;
    note = null;
    try {
      await fn();
      note = { ok: true, text: done };
      then?.();
      await load(false);
    } catch (err) {
      note = { ok: false, text: podErrorLine(err) };
    } finally {
      busy = false;
    }
  }

  async function openPicker(): Promise<void> {
    picking = true;
    pickedFriend = "";
    try {
      friends = (await fetchFriends()).friends;
    } catch (err) {
      note = { ok: false, text: podErrorLine(err) };
      picking = false;
    }
  }

  const doInvite = () => {
    const f = pickable.find((x) => x.user_id === pickedFriend);
    if (!f) return Promise.resolve();
    return act(
      () => invitePodFriend(id, f.user_id),
      `Invited ${f.display_name}. They can accept or decline.`,
      () => {
        picking = false;
      },
    );
  };

  const doRename = () =>
    act(
      () => renamePod(id, newName.trim()),
      "Renamed.",
      () => {
        renaming = false;
      },
    );

  const doHandOver = (m: PodMember) =>
    act(() => handOverPod(id, m.user_id), `${m.display_name} is the owner now.`);

  async function doRemove(m: PodMember): Promise<void> {
    const wantLink = alsoNewLink && !!pod?.link?.live;
    await act(
      async () => {
        const link = await removePodMember(id, m.user_id, wantLink);
        if (link) minted = link;
      },
      `Removed ${m.display_name}. They aren’t told.`,
      () => {
        removing = null;
      },
    );
  }

  async function doLeave(): Promise<void> {
    if (busy) return;
    busy = true;
    try {
      await removePodMember(id, "me");
      navigate(PODS_HASH);
    } catch (err) {
      note = { ok: false, text: podErrorLine(err) };
      busy = false;
    }
  }

  async function doDelete(): Promise<void> {
    if (busy) return;
    busy = true;
    try {
      await deletePod(id);
      navigate(PODS_HASH);
    } catch (err) {
      note = { ok: false, text: podErrorLine(err) };
      busy = false;
    }
  }

  const doMint = () =>
    act(async () => {
      minted = await mintPodLink(id);
      copied = false;
    }, "New invite link made. Copy it now; it isn’t shown again.");

  const doTurnOff = () =>
    act(async () => {
      await turnOffPodLink(id);
      minted = null;
    }, "Invite link turned off.");

  const doCancelInvite = (userID: string, name: string) =>
    act(() => cancelPodInvite(id, userID), `Cancelled the invite to ${name}.`);

  async function copyLink(): Promise<void> {
    if (!minted) return;
    try {
      await navigator.clipboard.writeText(podLinkURL(minted.path));
      copied = true;
    } catch {
      copied = false;
      note = { ok: false, text: "Couldn’t copy. Select the link and copy it by hand." };
    }
  }

  function menuFor(m: PodMember) {
    if (m.blocked) return [{ label: L.removePodMember, danger: true, onclick: () => ask(m) }];
    return [
      { label: L.makePodOwner, onclick: () => void doHandOver(m) },
      { label: L.removePodMember, danger: true, onclick: () => ask(m) },
    ];
  }

  function ask(m: PodMember): void {
    removing = m;
    alsoNewLink = true;
  }

  function initial(name: string): string {
    return name.trim().slice(0, 1).toUpperCase() || "?";
  }
</script>

<Page eyebrow="Pod" title={pod?.name ?? "Pod"}>
  {#snippet lede()}
    {#if pod}
      {pod.owner ? `Owned by ${pod.owner.display_name} · ` : ""}{memberCountLine(
        pod.member_count,
      )}{#if pod.invites.length > 0}
        · {pod.invites.length} invited{/if}
    {:else}
      A group you play with.
    {/if}
  {/snippet}
  {#snippet actions()}
    <a class="btn ghost" href={PODS_HASH}>All pods</a>
  {/snippet}

  {#if !isPerson}
    <p class="card note" role="status">Sign in with Discord to see your pods.</p>
  {:else if loading}
    <p class="card note" role="status">Loading…</p>
  {:else if gone}
    <div class="card note" role="alert">
      <p>That pod isn’t there. It may have been deleted, or you may have left it.</p>
      <a class="btn" href={PODS_HASH}>Your pods</a>
    </div>
  {:else if error}
    <div class="card note" role="alert">
      <p>{error}</p>
      <button type="button" class="btn sm" onclick={() => void load()}>Try again</button>
    </div>
  {:else if pod}
    {#if note}
      <p class="flash" class:bad={!note.ok} role={note.ok ? "status" : "alert"}>{note.text}</p>
    {/if}

    <div class="grid">
      <section class="card" aria-labelledby="pd-members">
        <h2 id="pd-members">
          Members <span class="count">{pod.member_count}/{pod.max_members}</span>
        </h2>
        <ul class="rows">
          {#each pod.members as m (m.user_id)}
            <li>
              <span class="who">
                {#if m.avatar_url}
                  <img class="avatar" src={m.avatar_url} alt="" width="32" height="32" />
                {:else}
                  <span class="avatar fallback" aria-hidden="true">{initial(m.display_name)}</span>
                {/if}
                <span class="text">
                  <span class="name">{m.display_name}</span>
                  <span class="sub">{m.blocked ? "Blocked" : joinedLine(m.joined_at)}</span>
                </span>
              </span>
              <span class="acts">
                {#if m.owner}<span class="tag">Owner</span>{/if}
                {#if !m.blocked}
                  <span class="tag" class:on={m.sharing}
                    >{m.sharing ? "Sharing" : "Not sharing"}</span
                  >
                {/if}
                {#if isOwner && !m.owner}
                  <MoreMenu label={`actions for ${m.display_name}`} items={menuFor(m)} />
                {/if}
              </span>
              {#if removing?.user_id === m.user_id}
                <span class="confirm" role="group" aria-label="confirm">
                  <span class="ask">Remove {m.display_name}? They aren’t told.</span>
                  {#if pod.link?.live}
                    <label class="choice">
                      <input type="checkbox" bind:checked={alsoNewLink} />
                      Also make a new invite link
                    </label>
                  {/if}
                  <span class="acts">
                    <button
                      type="button"
                      class="btn sm danger"
                      disabled={busy}
                      onclick={() => void doRemove(m)}>{L.removePodMember}</button
                    >
                    <button
                      type="button"
                      class="btn sm ghost"
                      disabled={busy}
                      onclick={() => (removing = null)}>Keep</button
                    >
                  </span>
                </span>
              {/if}
            </li>
          {/each}
        </ul>
        <p class="hint">
          Joining a pod shares nothing. A member who isn’t sharing their stats shows as Not sharing,
          with no numbers.
        </p>
      </section>

      <div class="col">
        <section class="card" aria-labelledby="pd-invite">
          <h2 id="pd-invite">Invite</h2>
          {#if picking}
            {#if pickable.length === 0}
              <p class="empty">
                None of your friends are free to invite: they’re in this pod already, invited, or
                not available.
              </p>
            {:else}
              <div class="choices" role="radiogroup" aria-label="Friends to invite">
                {#each pickable as f (f.user_id)}
                  <label class="choice">
                    <input type="radio" name="friend" value={f.user_id} bind:group={pickedFriend} />
                    {f.display_name}
                  </label>
                {/each}
              </div>
            {/if}
            <div class="acts">
              <button
                type="button"
                class="btn sm primary"
                disabled={busy || pickedFriend === ""}
                onclick={() => void doInvite()}>Send invite</button
              >
              <button type="button" class="btn sm ghost" onclick={() => (picking = false)}
                >Cancel</button
              >
            </div>
          {:else}
            <button
              type="button"
              class="btn"
              disabled={busy || full}
              onclick={() => void openPicker()}>{L.inviteToPod}</button
            >
            {#if full}<p class="hint">This pod is full.</p>{/if}
          {/if}
          {#if pod.invites.length > 0}
            <h3>Waiting for an answer</h3>
            <ul class="rows">
              {#each pod.invites as i (i.user_id ?? `${i.display_name}-${i.since}`)}
                <li>
                  <span class="text">
                    <span class="name">{i.display_name}</span>
                    {#if i.invited_by}<span class="sub">invited by {i.invited_by.display_name}</span
                      >{/if}
                  </span>
                  {#if i.can_cancel && i.user_id}
                    {@const uid = i.user_id}
                    <button
                      type="button"
                      class="btn sm ghost"
                      disabled={busy}
                      onclick={() => void doCancelInvite(uid, i.display_name)}>Cancel</button
                    >
                  {/if}
                </li>
              {/each}
            </ul>
          {/if}
        </section>

        {#if isOwner}
          <section class="card" aria-labelledby="pd-link">
            <h2 id="pd-link">Invite link</h2>
            <p class="empty">{linkLine(pod.link)}</p>
            {#if minted}
              <p class="link-box">
                <input
                  type="text"
                  readonly
                  aria-label="Invite link"
                  value={podLinkURL(minted.path)}
                  onfocus={(e) => e.currentTarget.select()}
                />
                <button type="button" class="btn sm" onclick={() => void copyLink()}
                  >{copied ? "Copied" : L.copyPodLink}</button
                >
              </p>
            {/if}
            <div class="acts">
              <button
                type="button"
                class="btn sm"
                disabled={busy || full}
                onclick={() => void doMint()}>{L.makePodLink}</button
              >
              {#if pod.link?.live}
                <button
                  type="button"
                  class="btn sm ghost"
                  disabled={busy}
                  onclick={() => void doTurnOff()}>Turn off</button
                >
              {/if}
            </div>
            <p class="hint">
              A link works for 7 days. Anyone who is signed in can use it, so make a new one if it
              gets around.
            </p>
          </section>

          <section class="card" aria-labelledby="pd-manage">
            <h2 id="pd-manage">Manage</h2>
            {#if renaming}
              <form
                class="rename"
                onsubmit={(e) => {
                  e.preventDefault();
                  void doRename();
                }}
              >
                <label for="pod-rename">Pod name</label>
                <input id="pod-rename" type="text" maxlength="40" required bind:value={newName} />
                <span class="acts">
                  <button
                    type="submit"
                    class="btn sm primary"
                    disabled={busy || newName.trim() === ""}>Save</button
                  >
                  <button type="button" class="btn sm ghost" onclick={() => (renaming = false)}
                    >Cancel</button
                  >
                </span>
              </form>
            {:else if deleting}
              <p class="ask">Delete {pod.name}? Its members and invites go with it.</p>
              <div class="acts">
                <button
                  type="button"
                  class="btn sm danger"
                  disabled={busy}
                  onclick={() => void doDelete()}>Delete pod</button
                >
                <button type="button" class="btn sm ghost" onclick={() => (deleting = false)}
                  >Keep</button
                >
              </div>
            {:else}
              <div class="acts">
                <button
                  type="button"
                  class="btn sm"
                  onclick={() => {
                    newName = pod?.name ?? "";
                    renaming = true;
                  }}>Rename</button
                >
                <button type="button" class="btn sm ghost" onclick={() => (deleting = true)}
                  >Delete pod</button
                >
              </div>
            {/if}
          </section>
        {/if}

        <section class="card" aria-labelledby="pd-leave">
          <h2 id="pd-leave">Leave</h2>
          {#if leaving}
            <p class="ask">Leave {pod.name}? Nobody is told.</p>
            <div class="acts">
              <button
                type="button"
                class="btn sm danger"
                disabled={busy}
                onclick={() => void doLeave()}>{L.leavePod}</button
              >
              <button type="button" class="btn sm ghost" onclick={() => (leaving = false)}
                >Stay</button
              >
            </div>
          {:else}
            <button type="button" class="btn sm ghost" onclick={() => (leaving = true)}
              >{L.leavePod}</button
            >
            {#if isOwner}
              <p class="hint">
                If you leave, the member who joined first becomes the owner. The last one out
                deletes the pod.
              </p>
            {/if}
          {/if}
        </section>
      </div>
    </div>
  {/if}
</Page>

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 18px;
    align-items: start;
  }
  .col {
    display: grid;
    gap: 18px;
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
  .count {
    font-family: var(--font-mono);
    font-size: 0.72rem;
    font-weight: 500;
    padding: 1px 8px;
    border-radius: 999px;
    background: var(--surface-raised);
    color: var(--fg-muted);
  }
  .empty,
  .hint {
    margin: 0 0 10px;
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .hint {
    margin: 12px 0 0;
  }
  .rows {
    list-style: none;
    margin: 0;
    padding: 0;
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
    flex: 1 1 160px;
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
  }
  .acts {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
  }
  .tag {
    font-size: 0.66rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 2px 8px;
    border-radius: 999px;
    background: var(--surface-raised);
    color: var(--fg-muted);
  }
  .tag.on {
    color: var(--mint);
  }
  .confirm {
    display: flex;
    flex-direction: column;
    gap: 8px;
    flex: 1 1 100%;
  }
  .ask {
    margin: 0 0 8px;
    font-size: 0.85rem;
    line-height: 1.45;
  }
  .confirm .ask {
    margin: 0;
  }
  .choices {
    display: grid;
    gap: 6px;
    margin-bottom: 12px;
    max-height: 260px;
    overflow-y: auto;
  }
  .choice {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.88rem;
    min-height: 32px;
    cursor: pointer;
  }
  .link-box {
    display: flex;
    gap: 6px;
    margin: 0 0 10px;
  }
  .link-box input {
    flex: 1;
    min-width: 0;
    font-family: var(--font-mono);
    font-size: 0.78rem;
  }
  .rename {
    display: grid;
    gap: 8px;
  }
  .rename label {
    font-size: 0.8rem;
    font-weight: 600;
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
  .btn.danger {
    border-color: var(--danger-line);
    background: var(--danger-soft);
    color: var(--fg);
  }
  .btn.danger:hover:not(:disabled) {
    border-color: var(--danger);
  }
</style>
