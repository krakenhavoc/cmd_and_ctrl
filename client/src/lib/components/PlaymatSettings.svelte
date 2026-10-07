<script lang="ts">
  // The Settings "Playmat" tab (ADR 0128): which playmats the board
  // draws (per device), and, for a signed-in account, your own: a
  // preview, an upload, a link, and Remove.
  //
  // The account half is hidden whenever the server says you cannot have
  // one: a 403 (a guest, the admin token, a server with no database) or
  // enabled:false (no data directory). The per-device choice is always
  // shown, because it is about this screen and a guest has screens too.

  import { onDestroy, onMount } from "svelte";
  import { settings, updateSettings } from "../settings";
  import { session, LobbyApiError } from "../session";
  import {
    PLAYMAT_DEFAULT_WASH,
    PLAYMAT_MAX_WASH,
    PLAYMAT_MIN_WASH,
    fetchMyPlaymat,
    linkMyPlaymat,
    removeMyPlaymat,
    setMyPlaymatWash,
    uploadMyPlaymat,
    type MyPlaymat,
  } from "../api";
  import { PLAYMAT_MAX_BYTES, playmatSrc, type PlaymatsMode } from "../playmat";

  const CHOICES: ReadonlyArray<{ value: PlaymatsMode; label: string }> = [
    { value: "all", label: "Everyone's playmats" },
    { value: "mine", label: "Only mine" },
    { value: "off", label: "Off" },
  ];

  // "loading" until the first GET answers; "hidden" when the account
  // cannot have a playmat; "ready" otherwise.
  let phase = $state<"loading" | "hidden" | "ready">("loading");
  let mine = $state<MyPlaymat>({ enabled: true });
  let busy = $state(false);
  let error = $state("");
  let linkText = $state("");
  let fileInput = $state<HTMLInputElement | null>(null);
  // A preview that fails to load shows no broken-image box.
  let previewFailed = $state<string | null>(null);

  const preview = $derived(mine.url && previewFailed !== mine.url ? playmatSrc(mine.url) : null);

  onMount(() => {
    if (!$session) {
      phase = "hidden";
      return;
    }
    void load();
  });

  async function load(): Promise<void> {
    try {
      const got = await fetchMyPlaymat();
      if (!got.enabled) {
        phase = "hidden";
        return;
      }
      mine = got;
      phase = "ready";
    } catch (e) {
      // 403: not a signed-in person. Anything else: leave the section
      // out rather than show a form that cannot work.
      phase = "hidden";
      if (!(e instanceof LobbyApiError && e.status === 403)) {
        console.warn("could not read the playmat", e);
      }
    }
  }

  // run wraps one write: it blocks the other buttons, clears the last
  // error and shows the server's message when the write is refused.
  async function run(write: () => Promise<MyPlaymat>): Promise<boolean> {
    busy = true;
    error = "";
    try {
      mine = await write();
      previewFailed = null;
      return true;
    } catch (e) {
      error = e instanceof Error ? e.message : "something went wrong; try again";
      return false;
    } finally {
      busy = false;
    }
  }

  async function onFile(e: Event): Promise<void> {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = ""; // choosing the same file again must fire change again
    if (!file) return;
    if (file.size > PLAYMAT_MAX_BYTES) {
      error = "That image is larger than 10 MB.";
      return;
    }
    await run(() => uploadMyPlaymat(file));
  }

  async function useLink(e: Event): Promise<void> {
    e.preventDefault();
    const url = linkText.trim();
    if (!url) {
      error = "Paste a link to an image first.";
      return;
    }
    if (await run(() => linkMyPlaymat(url))) linkText = "";
  }

  async function remove(): Promise<void> {
    await run(() => removeMyPlaymat());
  }

  // The owner-set wash (ADR 0128 amendment). The slider moves the
  // preview at once and saves once it has rested, so a drag is one
  // write, not one per step.
  let wash = $state(PLAYMAT_DEFAULT_WASH);
  let washTimer: ReturnType<typeof setTimeout> | null = null;
  $effect(() => {
    // Every answer from the server carries the account's wash.
    if (mine.wash) wash = mine.wash;
  });
  function onWash(value: number): void {
    wash = value;
    if (washTimer) clearTimeout(washTimer);
    washTimer = setTimeout(() => {
      washTimer = null;
      void setMyPlaymatWash(wash).then(
        () => (error = ""),
        (e: unknown) => (error = e instanceof Error ? e.message : "could not save the darkness"),
      );
    }, 400);
  }
  onDestroy(() => {
    if (washTimer) clearTimeout(washTimer);
  });
</script>

<h3>Playmat</h3>

<fieldset class="choice">
  <legend>Show on the table</legend>
  {#each CHOICES as c (c.value)}
    <label>
      <input
        type="radio"
        name="playmats-mode"
        value={c.value}
        checked={$settings.display.playmats === c.value}
        onchange={() => updateSettings("display", "playmats", c.value)}
      />
      {c.label}
    </label>
  {/each}
</fieldset>
<p class="help">
  A playmat is drawn behind its owner's battlefield, as on a paper table. Turn them off if someone
  else's art is distracting, or on a slow device. This choice is for this device only.
</p>

{#if phase === "ready"}
  <h4>Your playmat</h4>
  {#if preview}
    <div class="preview" style:--playmat-wash={`${wash}%`}>
      <img
        src={preview}
        alt="Your playmat"
        referrerpolicy="no-referrer"
        onerror={() => (previewFailed = mine.url ?? null)}
      />
    </div>
  {:else}
    <p class="help none">You have no playmat.</p>
  {/if}

  <!-- ADR 0128 amendment: how dark the playmat is under the cards. The
       preview follows at once; the server hears once the slider rests. -->
  <label class="wash">
    Darken
    <input
      type="range"
      min={PLAYMAT_MIN_WASH}
      max={PLAYMAT_MAX_WASH}
      step="1"
      value={wash}
      oninput={(e) => onWash(Number(e.currentTarget.value))}
      aria-label="Playmat darkness"
    />
    <span class="wash-value">{wash}%</span>
  </label>
  <p class="help">
    How strongly your playmat is darkened under the cards. Everyone at your table sees it at this
    strength; darker keeps busy art from getting in the way of the cards.
  </p>

  <div class="actions">
    <input
      bind:this={fileInput}
      class="file"
      type="file"
      accept="image/png,image/jpeg,image/webp"
      onchange={onFile}
      aria-label="Playmat image file"
    />
    <button type="button" disabled={busy} onclick={() => fileInput?.click()}>Upload an image</button
    >
    {#if mine.url}
      <button type="button" class="remove" disabled={busy} onclick={remove}>Remove</button>
    {/if}
  </div>

  <form class="link" onsubmit={useLink}>
    <label for="playmat-link">Or paste a link to an image</label>
    <div class="link-row">
      <input
        id="playmat-link"
        type="url"
        inputmode="url"
        placeholder="https://example.com/mat.jpg"
        autocomplete="off"
        bind:value={linkText}
        disabled={busy}
      />
      <button type="submit" disabled={busy || linkText.trim() === ""}>Use this image</button>
    </div>
  </form>

  {#if error}
    <p class="error" role="alert">{error}</p>
  {/if}
  {#if busy}
    <p class="help" role="status">Working…</p>
  {/if}

  <p class="help">
    PNG, JPEG or WebP, up to 10 MB. The server keeps its own copy of the image, so a link is fetched
    once and never shown to other players, and your playmat does not stop working if the link does.
    Photo details such as the location are stripped. Everyone at your table sees it behind your
    battlefield.
  </p>
{:else if phase === "hidden" && !$session}
  <p class="help">Sign in with Discord to give yourself a playmat.</p>
{/if}

<style>
  /* The Settings panel's own row, heading and help styles are scoped to
     it, so this tab restates them with the same tokens. */
  h3 {
    margin: 0 0 6px;
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
    font-weight: 600;
  }
  label {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 0;
    cursor: pointer;
    font-size: 13px;
    font-weight: 600;
    color: var(--fg);
    border-bottom: 1px solid var(--border);
  }
  .choice input[type="radio"] {
    accent-color: var(--accent);
  }
  .help {
    color: var(--fg-dim);
    font-size: 11.5px;
    line-height: 1.45;
    margin: 4px 0 8px;
  }
  .link label {
    border-bottom: 0;
    padding: 0;
    cursor: default;
  }
  input[type="url"] {
    padding: 7px 10px;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border);
    background: var(--surface-sunken);
    color: var(--fg);
    font: inherit;
    font-size: 13px;
  }
  .choice {
    border: 0;
    padding: 0;
    margin: 0 0 8px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .choice legend {
    padding: 0;
    margin-bottom: 6px;
    font-weight: 600;
  }
  h4 {
    margin: 18px 0 8px;
    font-size: 0.95em;
  }
  .preview {
    margin: 0 0 10px;
    max-width: 320px;
    border: 1px solid var(--border);
    border-radius: 10px;
    overflow: hidden;
    background: var(--surface-sunken);
  }
  .preview img {
    display: block;
    width: 100%;
    max-height: 180px;
    object-fit: cover;
  }
  /* The board's own scrim (PlayerPanel's .playmat::after), so the
     preview shows the darkness the table will see. */
  .preview {
    position: relative;
  }
  .preview::after {
    content: "";
    position: absolute;
    inset: 0;
    background: color-mix(in srgb, var(--bg) var(--playmat-wash, 58%), transparent);
    pointer-events: none;
  }
  .wash {
    border-bottom: 0;
  }
  .wash input[type="range"] {
    flex: 1;
    accent-color: var(--accent);
  }
  .wash-value {
    min-width: 3.5ch;
    font-variant-numeric: tabular-nums;
    color: var(--fg-dim);
  }
  .none {
    margin: 0 0 10px;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;
  }
  .file {
    display: none;
  }
  .link {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-bottom: 8px;
  }
  .link-row {
    display: flex;
    gap: 8px;
  }
  .link-row input {
    flex: 1 1 auto;
    min-width: 0;
  }
  .error {
    color: var(--danger);
    margin: 8px 0;
  }
</style>
