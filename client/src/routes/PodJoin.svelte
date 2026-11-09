<script lang="ts">
  // PodJoin — #/pods/join/<token> (ADR 0220 §3b): a pod's invite link.
  //
  // Signed out, the page says one thing: sign in with Discord to join,
  // and shows nothing about the pod. Signed in, it shows the pod's name,
  // its owner and its member count with Join and Not now; Join is the
  // yes. Every dead link (unknown, rotated, expired, turned off, a full
  // pod, a refusal for a block) gets the same line, so a probe learns
  // nothing.
  import { onMount } from "svelte";
  import Page from "../lib/components/Page.svelte";
  import { L } from "../lib/labels";
  import { signedInUserID } from "../lib/myGames";
  import {
    DEAD_LINK,
    joinPodByLink,
    memberCountLine,
    podErrorLine,
    podHash,
    previewPodLink,
    PODS_HASH,
    type JoinPreview,
  } from "../lib/pods";
  import { navigate } from "../lib/router";
  import { LobbyApiError, session } from "../lib/session";

  interface Props {
    token: string;
  }
  const { token }: Props = $props();

  const isPerson = $derived(
    signedInUserID($session) !== null && $session?.principal.role !== "admin",
  );

  let preview = $state<JoinPreview | null>(null);
  let loading = $state(true);
  let dead = $state(false);
  let error = $state("");
  let busy = $state(false);

  onMount(async () => {
    if (!isPerson) {
      loading = false;
      return;
    }
    try {
      preview = await previewPodLink(token);
    } catch (err) {
      if (err instanceof LobbyApiError && (err.status === 404 || err.status === 400)) dead = true;
      else error = podErrorLine(err);
    } finally {
      loading = false;
    }
  });

  async function join(): Promise<void> {
    if (busy) return;
    busy = true;
    error = "";
    try {
      const row = await joinPodByLink(token);
      navigate(podHash(row.id));
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 404) dead = true;
      else error = podErrorLine(err);
    } finally {
      busy = false;
    }
  }
</script>

<Page eyebrow="Pods" title="Join a pod" width="prose">
  {#if !isPerson}
    <div class="card note">
      <p>Sign in with Discord to join this pod.</p>
      <a class="btn primary" href="#/login">Sign in with Discord</a>
    </div>
  {:else if loading}
    <p class="card note" role="status">Loading…</p>
  {:else if dead}
    <div class="card note" role="alert">
      <p>{DEAD_LINK}</p>
      <a class="btn" href={PODS_HASH}>Your pods</a>
    </div>
  {:else if preview}
    <div class="card note">
      <h2>{preview.name}</h2>
      <p class="sub">
        Owned by {preview.owner_name} · {memberCountLine(preview.member_count)}
      </p>
      {#if preview.member && preview.pod_id}
        <p>You’re already in this pod.</p>
        <a class="btn primary" href={podHash(preview.pod_id)}>Open pod</a>
      {:else}
        {#if error}<p class="bad" role="alert">{error}</p>{/if}
        <div class="acts">
          <button type="button" class="btn primary" disabled={busy} onclick={() => void join()}
            >{L.joinPod}</button
          >
          <a class="btn ghost" href={PODS_HASH}>Not now</a>
        </div>
      {/if}
    </div>
  {:else if error}
    <p class="card note" role="alert">{error}</p>
  {/if}
</Page>

<style>
  h2 {
    margin: 0 0 4px;
    font-family: var(--font-display);
    font-size: 1.25rem;
  }
  .sub {
    color: var(--fg-muted);
    font-size: 0.88rem;
  }
  .note p {
    margin: 0 0 12px;
  }
  .acts {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .bad {
    color: var(--fg);
  }
</style>
