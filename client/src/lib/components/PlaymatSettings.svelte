<script lang="ts">
  // PlaymatSettings is ADR 0128's control: a signed-in person's playmat,
  // the image behind their part of the board that everyone at the table
  // sees. Upload one, set how strongly it is darkened under the cards
  // (the wash), or remove it. A change reaches every table the person
  // sits at as soon as the server has it.
  //
  // A guest has no account to keep one on, so they are told to sign in.
  import { onDestroy, onMount } from "svelte";
  import {
    PLAYMAT_DEFAULT_WASH,
    PLAYMAT_MAX_BYTES,
    PLAYMAT_MAX_WASH,
    PLAYMAT_MIN_WASH,
    deletePlaymat,
    fetchMyPlaymat,
    playmatURL,
    setPlaymatWash,
    uploadPlaymat,
    type MyPlaymat,
  } from "../api";
  import { session } from "../session";
  import { signedInUserID } from "../myGames";

  const signedIn = $derived(signedInUserID($session) !== null);

  let current = $state<MyPlaymat>({});
  let wash = $state(PLAYMAT_DEFAULT_WASH);
  let busy = $state(false);
  let error = $state("");
  let loaded = $state(false);
  let fileInput = $state<HTMLInputElement | null>(null);

  const imageURL = $derived(playmatURL(current.path));

  onMount(async () => {
    if (!signedIn) return;
    try {
      current = await fetchMyPlaymat();
      wash = current.wash ?? PLAYMAT_DEFAULT_WASH;
    } catch {
      // Unreadable here (a server with no database answers 403): the
      // control still offers an upload, which says why if it fails.
    } finally {
      loaded = true;
    }
  });

  function message(err: unknown): string {
    return err instanceof Error && err.message ? err.message : "something went wrong; try again";
  }

  async function onPick(e: Event & { currentTarget: HTMLInputElement }) {
    const file = e.currentTarget.files?.[0];
    e.currentTarget.value = "";
    if (!file) return;
    error = "";
    if (file.size > PLAYMAT_MAX_BYTES) {
      error =
        "That image is larger than 4 MB. A JPEG or WebP of the same picture is usually much smaller.";
      return;
    }
    busy = true;
    try {
      current = await uploadPlaymat(file, wash);
      wash = current.wash ?? wash;
    } catch (err) {
      error = message(err);
    } finally {
      busy = false;
    }
  }

  // The slider previews at once and saves once it rests, so dragging it
  // is one write rather than one per step.
  let washTimer: ReturnType<typeof setTimeout> | null = null;
  function onWash(value: number) {
    wash = value;
    if (!current.path) return;
    if (washTimer) clearTimeout(washTimer);
    washTimer = setTimeout(async () => {
      washTimer = null;
      try {
        current = await setPlaymatWash(wash);
        error = "";
      } catch (err) {
        error = message(err);
      }
    }, 400);
  }
  onDestroy(() => {
    if (washTimer) clearTimeout(washTimer);
  });

  async function onRemove() {
    error = "";
    busy = true;
    try {
      await deletePlaymat();
      current = {};
    } catch (err) {
      error = message(err);
    } finally {
      busy = false;
    }
  }
</script>

<fieldset class="playmat">
  <legend>Playmat</legend>
  {#if !signedIn}
    <p class="help">
      Sign in with Discord to set a playmat: an image behind your part of the board that everyone at
      your table sees.
    </p>
  {:else}
    <p class="help">
      An image behind your part of the board. Everyone at your table sees it, under a dark wash so
      cards stay readable.
    </p>
    <div
      class="preview"
      class:empty={!imageURL}
      style:--playmat={imageURL ? `url("${imageURL}")` : undefined}
      style:--playmat-wash={`${wash}%`}
      aria-hidden="true"
    >
      {#if !imageURL}<span>{loaded ? "No playmat" : "…"}</span>{/if}
      <span class="sample-card"></span>
      <span class="sample-card"></span>
    </div>
    <div class="row">
      <input
        bind:this={fileInput}
        class="file"
        type="file"
        accept="image/png,image/jpeg,image/webp,image/gif"
        onchange={onPick}
        disabled={busy}
        aria-label="Playmat image"
      />
      <button type="button" onclick={() => fileInput?.click()} disabled={busy}>
        {current.path ? "Replace image" : "Upload image"}
      </button>
      {#if current.path}
        <button type="button" class="remove" onclick={onRemove} disabled={busy}>Remove</button>
      {/if}
    </div>
    <label class="wash">
      Darken
      <input
        type="range"
        min={PLAYMAT_MIN_WASH}
        max={PLAYMAT_MAX_WASH}
        step="5"
        value={wash}
        oninput={(e) => onWash(Number(e.currentTarget.value))}
        aria-label="Playmat darkness"
      />
      <span class="value">{wash}%</span>
    </label>
    <p class="help">
      PNG, JPEG, WebP or GIF, up to 4 MB. A wide landscape image fits a board best.
    </p>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
  {/if}
</fieldset>

<style>
  .playmat {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .preview {
    position: relative;
    height: 96px;
    border-radius: 10px;
    border: 1px solid var(--border);
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    background:
      linear-gradient(
        color-mix(in srgb, var(--bg) var(--playmat-wash, 60%), transparent),
        color-mix(in srgb, var(--bg) var(--playmat-wash, 60%), transparent)
      ),
      var(--playmat) center / cover no-repeat,
      var(--surface);
  }
  .preview.empty {
    background: var(--surface);
    color: var(--fg-dim);
    font-size: 12px;
  }
  .preview.empty .sample-card {
    display: none;
  }
  /* Two card-sized blanks, so the darkness is judged against cards. */
  .sample-card {
    width: 40px;
    height: 56px;
    border-radius: 4px;
    background: var(--surface-raised);
    border: 1px solid var(--border);
  }
  .row {
    display: flex;
    gap: 8px;
    align-items: center;
  }
  .file {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
    pointer-events: none;
  }
  .wash {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .wash input {
    flex: 1;
  }
  .value {
    font-variant-numeric: tabular-nums;
    min-width: 3ch;
  }
  .error {
    color: var(--danger, #e66);
    margin: 0;
  }
</style>
