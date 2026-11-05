<script lang="ts">
  // ShareStatsDialog — the share link (ADR 0217 §6): make a link to a
  // read-only copy of your stats, copy it, or revoke it. Only the link's
  // hash is kept on the server, so a link is shown once, right after it
  // is made; making a new one rotates it and the old one stops working.
  // The dialog says in one sentence what the link reveals.
  import { onMount } from "svelte";
  import ModalLayer from "../ModalLayer.svelte";
  import { L } from "../../labels";
  import { createShareLink, revokeShareLink, shareURL, SHARE_REVEALS } from "../../advancedStats";
  import { shareLinkChanged, statsPrefs } from "../../statsPrefs";

  interface Props {
    onclose: () => void;
  }
  const { onclose }: Props = $props();

  let url = $state("");
  let busy = $state(false);
  let error = $state("");
  let status = $state("");
  let box: HTMLElement | undefined = $state();

  const madeAt = $derived($statsPrefs.share_link_at);

  onMount(() => {
    box?.querySelector<HTMLButtonElement>("button.primary")?.focus();
  });

  function when(ms: number): string {
    return new Date(ms).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }

  async function make(): Promise<void> {
    busy = true;
    error = "";
    status = "";
    try {
      const link = await createShareLink();
      url = shareURL(location.origin, link.token);
      shareLinkChanged(link.created_at);
      status = "Link made. Copy it now: it is shown only once.";
    } catch (e) {
      error = e instanceof Error ? e.message : "could not make a link; try again";
    } finally {
      busy = false;
    }
  }

  async function copy(): Promise<void> {
    error = "";
    try {
      await navigator.clipboard.writeText(url);
      status = "Link copied.";
    } catch {
      error = "Could not copy. Select the link and copy it yourself.";
    }
  }

  async function revoke(): Promise<void> {
    busy = true;
    error = "";
    status = "";
    try {
      await revokeShareLink();
      url = "";
      shareLinkChanged(null);
      status = "Link revoked. It no longer works.";
    } catch (e) {
      error = e instanceof Error ? e.message : "could not revoke the link; try again";
    } finally {
      busy = false;
    }
  }

  function onKey(e: KeyboardEvent): void {
    if (e.key === "Escape") {
      e.preventDefault();
      onclose();
    }
  }
</script>

<svelte:window onkeydown={onKey} />

<ModalLayer />

<div
  class="prompt-backdrop"
  role="presentation"
  onclick={(e) => e.target === e.currentTarget && onclose()}
>
  <div
    class="prompt-modal share"
    role="dialog"
    aria-modal="true"
    aria-label={L.shareStatsDialog}
    data-testid="share-stats-dialog"
    bind:this={box}
  >
    <h2 class="prompt-title">Share your stats</h2>
    <p class="reveals">{SHARE_REVEALS}</p>
    <p class="prompt-hint">
      No sign-in is needed to open it, so it works when posted in Discord. You can revoke it at any
      time.
    </p>

    {#if url}
      <label class="link-field">
        <span>Your link</span>
        <input
          type="text"
          readonly
          value={url}
          onfocus={(e) => e.currentTarget.select()}
          data-testid="share-link-url"
        />
      </label>
    {:else if madeAt}
      <p class="prompt-hint" data-testid="share-link-active">
        You made a link on {when(madeAt)}, and it still works. It was shown only once: make a new
        link to copy one again, and the old one stops working.
      </p>
    {:else}
      <p class="prompt-hint">You have no link yet.</p>
    {/if}

    <p class="status" role="status" aria-live="polite">{status}</p>
    {#if error}<p class="prompt-hint error" role="alert">{error}</p>{/if}

    <div class="prompt-foot">
      <button type="button" class="ghost" onclick={onclose}>Close</button>
      {#if url || madeAt}
        <button type="button" class="danger" disabled={busy} onclick={() => void revoke()}
          >{L.revokeShareLink}</button
        >
      {/if}
      {#if url}
        <button type="button" class="primary" onclick={() => void copy()}>{L.copyShareLink}</button>
      {:else if madeAt}
        <button type="button" class="primary" disabled={busy} onclick={() => void make()}
          >{L.makeNewShareLink}</button
        >
      {:else}
        <button type="button" class="primary" disabled={busy} onclick={() => void make()}
          >{L.makeShareLink}</button
        >
      {/if}
    </div>
  </div>
</div>

<style>
  .share {
    width: min(560px, calc(100vw - 32px));
  }
  .reveals {
    margin: 0;
    font-size: 0.92rem;
    line-height: 1.5;
  }
  .link-field {
    display: grid;
    gap: 4px;
    font-size: 0.82rem;
    color: var(--fg-muted);
  }
  .link-field input {
    font: 500 0.85rem/1.2 var(--font-mono);
    padding: 9px 10px;
    border-radius: var(--radius);
    border: 1px solid var(--border-strong);
    background: var(--surface-sunken);
    color: var(--fg);
    min-width: 0;
  }
  .link-field input:focus {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-soft);
  }
  .status {
    margin: 0;
    min-height: 1.2em;
    font-size: 0.85rem;
    color: var(--fg-muted);
  }
  .prompt-foot {
    flex-wrap: wrap;
  }
</style>
