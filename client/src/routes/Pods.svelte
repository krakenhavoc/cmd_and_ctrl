<script lang="ts">
  // Pods — #/pods, my pods (ADR 0220 §11, slice 3 of #4880): the invites
  // waiting on you first, then Suggested, then your pods, one `.card` row each, and New
  // pod, the page's one primary. Between them, Suggested (slice 4): people
  // you keep sitting with, each card with Make this pod and Not now.
  //
  // Signed-in only (the router sends a signed-out visitor away). A guest
  // seat or the admin token is not a person and gets a line saying so; a
  // server with no pods store says pods are off.
  import { onMount } from "svelte";
  import Page from "../lib/components/Page.svelte";
  import { L } from "../lib/labels";
  import { signedInUserID } from "../lib/myGames";
  import {
    answerPodInvite,
    createPod,
    dismissPodSuggestion,
    fetchMyPods,
    memberCountLine,
    namesLine,
    podErrorLine,
    podHash,
    suggestionLine,
    type MyPods,
  } from "../lib/pods";
  import { navigate } from "../lib/router";
  import { LobbyApiError, session } from "../lib/session";

  const isPerson = $derived(
    signedInUserID($session) !== null && $session?.principal.role !== "admin",
  );

  let data = $state<MyPods | null>(null);
  let loading = $state(true);
  let error = $state("");
  let off = $state(false);
  let busy = $state<string | null>(null);
  let note = $state<{ ok: boolean; text: string } | null>(null);

  let naming = $state(false);
  let name = $state("");

  async function load(spinner = true): Promise<void> {
    if (spinner) loading = true;
    error = "";
    try {
      data = await fetchMyPods();
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 503) off = true;
      else error = podErrorLine(err);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    if (isPerson) void load();
    else loading = false;
  });

  async function answer(podID: string, podName: string, accept: boolean): Promise<void> {
    if (busy) return;
    busy = podID;
    note = null;
    try {
      await answerPodInvite(podID, accept);
      note = {
        ok: true,
        text: accept ? `You joined ${podName}.` : `Declined. ${podName} isn’t told.`,
      };
      await load(false);
    } catch (err) {
      note = { ok: false, text: podErrorLine(err) };
      await load(false);
    } finally {
      busy = null;
    }
  }

  // The suggestion being named, by key.
  let suggesting = $state<string | null>(null);
  let suggestedName = $state("");

  function startSuggestion(key: string, people: string[]): void {
    suggesting = key;
    suggestedName = people.length > 0 ? `${people[0]} & co` : "";
  }

  async function makeSuggested(e: SubmitEvent, key: string): Promise<void> {
    e.preventDefault();
    if (busy) return;
    busy = key;
    note = null;
    try {
      const row = await createPod(suggestedName.trim(), key);
      navigate(podHash(row.id));
    } catch (err) {
      note = { ok: false, text: podErrorLine(err) };
      await load(false);
    } finally {
      busy = null;
    }
  }

  async function notNow(key: string): Promise<void> {
    if (busy) return;
    busy = key;
    note = null;
    try {
      await dismissPodSuggestion(key);
      if (suggesting === key) suggesting = null;
      note = { ok: true, text: "Okay. We won’t suggest that group for a while." };
      await load(false);
    } catch (err) {
      note = { ok: false, text: podErrorLine(err) };
      await load(false);
    } finally {
      busy = null;
    }
  }

  async function make(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    if (busy) return;
    busy = "new";
    note = null;
    try {
      const row = await createPod(name.trim());
      navigate(podHash(row.id));
    } catch (err) {
      note = { ok: false, text: podErrorLine(err) };
    } finally {
      busy = null;
    }
  }

  function initial(s: string): string {
    return s.trim().slice(0, 1).toUpperCase() || "?";
  }
</script>

<Page
  eyebrow="Pods"
  title="Your pods"
  lede="A pod is a group you play with. Make one, then invite your friends."
>
  {#snippet actions()}
    {#if isPerson && data && !off}
      <button type="button" class="btn primary" onclick={() => (naming = true)}>{L.newPod}</button>
    {/if}
  {/snippet}

  {#if !isPerson}
    <p class="card note" role="status">Sign in with Discord to have pods.</p>
  {:else if loading}
    <p class="card note" role="status">Loading…</p>
  {:else if off}
    <p class="card note" role="status">Pods aren’t available on this server.</p>
  {:else if error}
    <div class="card note" role="alert">
      <p>{error}</p>
      <button type="button" class="btn sm" onclick={() => void load()}>Try again</button>
    </div>
  {:else if data}
    {#if note}
      <p class="flash" class:bad={!note.ok} role={note.ok ? "status" : "alert"}>{note.text}</p>
    {/if}

    {#if naming}
      <form class="card name-form" onsubmit={make}>
        <label for="pod-name">Name your pod</label>
        <div class="row">
          <input
            id="pod-name"
            type="text"
            maxlength="40"
            autocomplete="off"
            required
            bind:value={name}
          />
          <button type="submit" class="btn primary" disabled={busy !== null || name.trim() === ""}
            >Make pod</button
          >
          <button type="button" class="btn ghost" onclick={() => (naming = false)}>Cancel</button>
        </div>
        <p class="hint">
          1 to 40 characters. You’ll be its owner. You can be in {data.max_pods} pods.
        </p>
      </form>
    {/if}

    {#if data.invites.length > 0}
      <section class="card" aria-labelledby="p-invites">
        <h2 id="p-invites">Invites</h2>
        <ul class="rows">
          {#each data.invites as inv (inv.pod_id)}
            <li>
              <span class="who">
                {#if inv.invited_by.avatar_url}
                  <img
                    class="avatar"
                    src={inv.invited_by.avatar_url}
                    alt=""
                    width="32"
                    height="32"
                  />
                {:else}
                  <span class="avatar fallback" aria-hidden="true"
                    >{initial(inv.invited_by.display_name)}</span
                  >
                {/if}
                <span class="text">
                  <span class="name">{inv.pod_name}</span>
                  <span class="sub">{inv.invited_by.display_name} invited you</span>
                </span>
              </span>
              <span class="acts">
                <button
                  type="button"
                  class="btn sm primary"
                  disabled={busy !== null}
                  onclick={() => void answer(inv.pod_id, inv.pod_name, true)}
                  >{L.acceptPodInvite}</button
                >
                <button
                  type="button"
                  class="btn sm ghost"
                  disabled={busy !== null}
                  onclick={() => void answer(inv.pod_id, inv.pod_name, false)}
                  >{L.declinePodInvite}</button
                >
              </span>
            </li>
          {/each}
        </ul>
      </section>
    {/if}

    {#if data.suggestions.length > 0}
      <section class="card" aria-labelledby="p-suggested">
        <h2 id="p-suggested">Suggested</h2>
        <ul class="rows">
          {#each data.suggestions as sg (sg.key)}
            <li class="suggestion">
              <span class="who">
                <span class="faces" aria-hidden="true">
                  {#each sg.people as person (person.display_name)}
                    {#if person.avatar_url}
                      <img class="avatar" src={person.avatar_url} alt="" width="32" height="32" />
                    {:else}
                      <span class="avatar fallback">{initial(person.display_name)}</span>
                    {/if}
                  {/each}
                </span>
                <span class="text">
                  <span class="name wrap"
                    >{namesLine(sg.people.map((person) => person.display_name))}</span
                  >
                  <span class="sub">{suggestionLine(sg)}</span>
                </span>
              </span>
              {#if suggesting === sg.key}
                <form class="inline-name" onsubmit={(e) => void makeSuggested(e, sg.key)}>
                  <label for="sg-name-{sg.key}">Name your pod</label>
                  <input
                    id="sg-name-{sg.key}"
                    type="text"
                    maxlength="40"
                    autocomplete="off"
                    required
                    bind:value={suggestedName}
                  />
                  <button
                    type="submit"
                    class="btn sm primary"
                    disabled={busy !== null || suggestedName.trim() === ""}>Make pod</button
                  >
                  <button type="button" class="btn sm ghost" onclick={() => (suggesting = null)}
                    >Cancel</button
                  >
                  <span class="hint">Each of them gets an invite and chooses for themselves.</span>
                </form>
              {:else}
                <span class="acts">
                  <button
                    type="button"
                    class="btn sm primary"
                    disabled={busy !== null}
                    onclick={() =>
                      startSuggestion(
                        sg.key,
                        sg.people.map((person) => person.display_name),
                      )}>{L.makeSuggestedPod}</button
                  >
                  <button
                    type="button"
                    class="btn sm ghost"
                    disabled={busy !== null}
                    onclick={() => void notNow(sg.key)}>{L.dismissPodSuggestion}</button
                  >
                </span>
              {/if}
            </li>
          {/each}
        </ul>
      </section>
    {/if}

    <section aria-labelledby="p-pods">
      <h2 id="p-pods" class="sr-only">Your pods</h2>
      {#if data.pods.length === 0}
        <p class="card empty">
          A pod is a group you play with. Make one, then invite your friends.
        </p>
      {:else}
        <ul class="pods">
          {#each data.pods as p (p.id)}
            <li class="card pod-row">
              <a class="pod-link" href={podHash(p.id)}>
                <span class="name">{p.name}</span>
                <span class="sub">{memberCountLine(p.member_count)}</span>
              </a>
              {#if p.role === "owner"}<span class="tag">Owner</span>{/if}
            </li>
          {/each}
        </ul>
      {/if}
    </section>
  {/if}
</Page>

<style>
  section.card {
    margin-bottom: 18px;
  }
  h2 {
    margin: 0 0 12px;
    font-family: var(--font-display);
    font-size: 1.05rem;
    font-weight: 600;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }
  .empty,
  .hint {
    margin: 0;
    font-size: 0.88rem;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .hint {
    margin-top: 8px;
    font-size: 0.8rem;
  }
  .name-form {
    margin-bottom: 18px;
  }
  .name-form label {
    display: block;
    margin-bottom: 6px;
    font-size: 0.85rem;
    font-weight: 600;
  }
  .name-form .row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .name-form input {
    flex: 1 1 220px;
    min-width: 0;
  }
  .rows,
  .pods {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .pods {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }
  @media (max-width: 819px) {
    .pods {
      grid-template-columns: minmax(0, 1fr);
    }
  }
  .pod-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }
  .pod-link {
    display: flex;
    flex-direction: column;
    min-width: 0;
    flex: 1;
    color: inherit;
    text-decoration: none;
  }
  .pod-link:hover .name {
    text-decoration: underline;
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
    font-size: 0.95rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .sub {
    font-size: 0.78rem;
    color: var(--fg-muted);
  }
  .faces {
    display: flex;
    flex: 0 0 auto;
  }
  .faces .avatar + .avatar {
    margin-left: -8px;
    box-shadow: 0 0 0 2px var(--surface);
  }
  .name.wrap {
    white-space: normal;
  }
  .suggestion .inline-name {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    flex: 1 1 100%;
  }
  .inline-name label {
    flex: 1 1 100%;
    font-size: 0.8rem;
    font-weight: 600;
  }
  .inline-name input {
    flex: 1 1 200px;
    min-width: 0;
  }
  .inline-name .hint {
    flex: 1 1 100%;
    margin: 2px 0 0;
  }
  .acts {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .tag {
    flex: 0 0 auto;
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 2px 8px;
    border-radius: 999px;
    background: var(--surface-raised);
    color: var(--fg-muted);
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
</style>
