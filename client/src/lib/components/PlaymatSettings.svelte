<script lang="ts">
  // The Settings "Playmat" tab (ADR 0128): which playmats the board
  // draws (per device), and, for a signed-in account, your own: up to
  // three saved playmats, one of them on show (§11).
  //
  // Each slot is a card with its thumbnail and the actions that fit it:
  // Use / Stop using, Replace, Fit to best size (only when the image is
  // not the best shape) and Remove. Adding or replacing a playmat opens
  // an upload-or-link panel for that slot; when it is saved and the
  // image is not the best shape, the best-size prompt opens right after.
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
    PLAYMAT_SLOTS,
    activateMyPlaymat,
    fetchMyPlaymats,
    fitMyPlaymat,
    linkMyPlaymat,
    removeMyPlaymat,
    setMyPlaymatWash,
    uploadMyPlaymat,
    type MyPlaymats,
    type PlaymatSlot,
  } from "../api";
  import { PLAYMAT_MAX_BYTES, playmatSrc, type CropOrigin, type PlaymatsMode } from "../playmat";
  import PlaymatFitPrompt from "./PlaymatFitPrompt.svelte";

  const CHOICES: ReadonlyArray<{ value: PlaymatsMode; label: string }> = [
    { value: "all", label: "Everyone's playmats" },
    { value: "mine", label: "Only mine" },
    { value: "off", label: "Off" },
  ];

  // "loading" until the first GET answers; "hidden" when the account
  // cannot have a playmat; "ready" otherwise.
  let phase = $state<"loading" | "hidden" | "ready">("loading");
  let mine = $state<MyPlaymats>({ enabled: true });
  let busy = $state(false);
  let error = $state("");
  let linkText = $state("");
  let fileInput = $state<HTMLInputElement | null>(null);
  // Thumbnails that fail to load show no broken-image box, keyed by URL.
  let failed = $state<string[]>([]);

  // The slot whose upload-or-link panel is open, whether the full-slots
  // chooser is open, the slot whose best-size prompt is open, and the
  // slot waiting on a Remove confirmation.
  let editing = $state<number | null>(null);
  let choosing = $state(false);
  let fitting = $state<number | null>(null);
  let confirming = $state<number | null>(null);

  const maxSlots = $derived(mine.max_slots ?? PLAYMAT_SLOTS);
  const saved = $derived(mine.slots ?? []);
  const slotNumbers = $derived(Array.from({ length: maxSlots }, (_, i) => i + 1));
  const bySlot = (n: number): PlaymatSlot | undefined => saved.find((s) => s.slot === n);
  const firstFree = $derived(slotNumbers.find((n) => !bySlot(n)));
  const activeSlot = $derived(saved.find((s) => s.slot === mine.active));
  const fitSlot = $derived(fitting === null ? undefined : bySlot(fitting));

  const thumb = (s: PlaymatSlot): string | null =>
    failed.includes(s.url) ? null : playmatSrc(s.url);
  const preview = $derived(activeSlot ? thumb(activeSlot) : null);

  onMount(() => {
    if (!$session) {
      phase = "hidden";
      return;
    }
    void load();
  });

  async function load(): Promise<void> {
    try {
      const got = await fetchMyPlaymats();
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
        console.warn("could not read the playmats", e);
      }
    }
  }

  // run wraps one write: it blocks the other buttons, clears the last
  // error and shows the server's message when the write is refused. It
  // returns the account's new state, or null when the write failed.
  async function run(write: () => Promise<MyPlaymats>): Promise<MyPlaymats | null> {
    busy = true;
    error = "";
    try {
      const got = await write();
      mine = got;
      failed = [];
      return got;
    } catch (e) {
      error = e instanceof Error ? e.message : "something went wrong; try again";
      return null;
    } finally {
      busy = false;
    }
  }

  // afterSave closes the add panel and, when the image just saved is
  // not the best shape, opens the best-size prompt on it.
  function afterSave(got: MyPlaymats | null): void {
    if (!got) return;
    editing = null;
    const written =
      got.slot === undefined ? undefined : got.slots?.find((s) => s.slot === got.slot);
    fitting = written?.suggestion ? written.slot : null;
  }

  // Add a playmat: the first free slot, or, with all of them full, a
  // choice of which one to replace.
  function startAdd(): void {
    error = "";
    fitting = null;
    confirming = null;
    if (firstFree !== undefined) {
      editing = firstFree;
      choosing = false;
    } else {
      editing = null;
      choosing = true;
    }
  }

  function startEdit(n: number): void {
    error = "";
    fitting = null;
    confirming = null;
    choosing = false;
    linkText = "";
    editing = n;
  }

  function cancelEdit(): void {
    editing = null;
    choosing = false;
    error = "";
  }

  async function onFile(e: Event): Promise<void> {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = ""; // choosing the same file again must fire change again
    if (!file || editing === null) return;
    if (file.size > PLAYMAT_MAX_BYTES) {
      error = "That image is larger than 10 MB.";
      return;
    }
    const slot = editing;
    afterSave(await run(() => uploadMyPlaymat(slot, file)));
  }

  async function useLink(e: Event): Promise<void> {
    e.preventDefault();
    const url = linkText.trim();
    if (!url) {
      error = "Paste a link to an image first.";
      return;
    }
    if (editing === null) return;
    const slot = editing;
    const got = await run(() => linkMyPlaymat(slot, url));
    if (got) linkText = "";
    afterSave(got);
  }

  async function toggleUse(n: number): Promise<void> {
    await run(() => activateMyPlaymat(mine.active === n ? null : n));
  }

  async function remove(n: number): Promise<void> {
    confirming = null;
    if (fitting === n) fitting = null;
    await run(() => removeMyPlaymat(n));
  }

  function startFit(n: number): void {
    error = "";
    editing = null;
    choosing = false;
    confirming = null;
    fitting = n;
  }

  async function acceptFit(origin: CropOrigin): Promise<void> {
    if (fitting === null) return;
    const slot = fitting;
    const got = await run(() => fitMyPlaymat(slot, origin.x, origin.y));
    if (got) fitting = null;
  }

  // The owner-set wash (ADR 0128 §10). The slider moves the preview at
  // once and saves once it has rested, so a drag is one write, not one
  // per step. One wash per account, not per slot.
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
  <h4>Your playmats</h4>
  <p class="help">
    Keep up to {maxSlots} playmats and choose which one is on your table. Switching between them keeps
    every image; removing one deletes it.
  </p>

  <div class="slots" role="list" aria-label="Your playmats">
    {#each slotNumbers as n (n)}
      {@const s = bySlot(n)}
      {@const on = s !== undefined && mine.active === n}
      {@const src = s ? thumb(s) : null}
      <article class="slot" class:on role="listitem" aria-label={`Playmat ${n}`}>
        <div class="slot-head">
          <span class="slot-name">Playmat {n}</span>
          {#if on}<span class="badge">Using</span>{/if}
        </div>
        {#if s}
          <div class="thumb">
            {#if src}
              <img
                {src}
                alt={`Playmat ${n}`}
                loading="lazy"
                referrerpolicy="no-referrer"
                onerror={() => (failed = [...failed, s.url])}
              />
            {:else}
              <span class="thumb-none">Image unavailable</span>
            {/if}
          </div>
          <p class="size">
            <span>{s.width}&times;{s.height}</span>
            {#if s.suggestion}<span class="shape-note">not the best shape</span>{/if}
          </p>
          <div class="slot-actions">
            <button
              type="button"
              disabled={busy}
              aria-label={on ? `Stop using playmat ${n}` : `Use playmat ${n}`}
              onclick={() => toggleUse(n)}>{on ? "Stop using" : "Use"}</button
            >
            <button
              type="button"
              disabled={busy}
              aria-label={`Replace playmat ${n}`}
              onclick={() => startEdit(n)}>Replace</button
            >
            {#if s.suggestion}
              <button
                type="button"
                disabled={busy}
                aria-label={`Fit playmat ${n} to the best size`}
                onclick={() => startFit(n)}>Fit to best size</button
              >
            {/if}
            <button
              type="button"
              class="remove"
              disabled={busy}
              aria-label={`Remove playmat ${n}`}
              onclick={() => (confirming = n)}>Remove</button
            >
          </div>
          {#if confirming === n}
            <div class="confirm" role="alertdialog" aria-label={`Remove playmat ${n}?`}>
              <p>
                Remove playmat {n}?{on ? " It is the one on your table." : ""} Its image is deleted.
              </p>
              <div class="slot-actions">
                <button type="button" class="danger" disabled={busy} onclick={() => remove(n)}>
                  Yes, remove it
                </button>
                <button type="button" disabled={busy} onclick={() => (confirming = null)}>
                  Cancel
                </button>
              </div>
            </div>
          {/if}
        {:else}
          <div class="thumb empty">
            <span class="thumb-none">Empty</span>
          </div>
          <div class="slot-actions">
            <button
              type="button"
              disabled={busy}
              aria-label={`Add a playmat to slot ${n}`}
              onclick={() => startEdit(n)}>Add a playmat</button
            >
          </div>
        {/if}
      </article>
    {/each}
  </div>

  <div class="actions">
    <button type="button" class="primary" disabled={busy} onclick={startAdd}>Add a playmat</button>
    {#if mine.active !== undefined}
      <button type="button" disabled={busy} onclick={() => run(() => activateMyPlaymat(null))}>
        Show no playmat
      </button>
    {/if}
  </div>

  {#if choosing}
    <div class="panel" role="group" aria-label="Choose a slot to replace">
      <p class="said">All {maxSlots} slots are full. Which playmat should the new one replace?</p>
      <div class="slot-actions">
        {#each slotNumbers as n (n)}
          <button type="button" disabled={busy} onclick={() => startEdit(n)}>
            Replace playmat {n}
          </button>
        {/each}
        <button type="button" disabled={busy} onclick={cancelEdit}>Cancel</button>
      </div>
    </div>
  {/if}

  {#if editing !== null}
    {@const replacing = bySlot(editing) !== undefined}
    <div class="panel" role="group" aria-label={`Add a playmat to slot ${editing}`}>
      <p class="said">
        {replacing
          ? `Replace playmat ${editing}. The image it holds now is deleted.`
          : `Add a playmat to slot ${editing}.`}
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
        <button type="button" disabled={busy} onclick={() => fileInput?.click()}
          >Upload an image</button
        >
        <button type="button" disabled={busy} onclick={cancelEdit}>Cancel</button>
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
    </div>
  {/if}

  {#if fitSlot && fitSlot.suggestion}
    <PlaymatFitPrompt
      mat={fitSlot}
      idealWidth={mine.ideal_width ?? 2400}
      idealHeight={mine.ideal_height ?? 1400}
      {busy}
      onaccept={acceptFit}
      onkeep={() => (fitting = null)}
    />
  {/if}

  {#if error}
    <p class="error" role="alert">{error}</p>
  {/if}
  {#if busy}
    <p class="help" role="status">Working…</p>
  {/if}

  <h4>On your table</h4>
  {#if preview}
    <div class="preview" style:--playmat-wash={`${wash}%`}>
      <img src={preview} alt="Your playmat on the table" referrerpolicy="no-referrer" />
    </div>
  {:else}
    <p class="help none">
      {activeSlot ? "The playmat on your table cannot be shown." : "No playmat is on your table."}
    </p>
  {/if}

  <!-- ADR 0128 §10: how dark the playmat is under the cards, one value for
       whichever playmat you are using. The preview follows at once; the
       server hears once the slider rests. -->
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

  <p class="help">
    PNG, JPEG or WebP, up to 10 MB. The server keeps its own copy of the image, so a link is fetched
    once and never shown to other players, and your playmat does not stop working if the link does.
    Photo details such as the location are stripped. Everyone at your table sees the playmat you are
    using behind your battlefield.
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

  .slots {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(min(150px, 100%), 1fr));
    gap: 10px;
    margin: 0 0 12px;
  }
  .slot {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 10px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--surface-sunken);
    min-width: 0;
  }
  .slot.on {
    border-color: var(--accent-line);
    box-shadow: 0 0 0 1px var(--accent-line);
  }
  .slot-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
  .slot-name {
    font-size: 12px;
    font-weight: 600;
    color: var(--fg);
  }
  .badge {
    font-size: 10.5px;
    font-weight: 700;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    padding: 2px 7px;
    border-radius: 999px;
    background: var(--accent-soft);
    color: var(--accent-strong);
    border: 1px solid var(--accent-line);
  }
  .thumb {
    aspect-ratio: 12 / 7;
    border-radius: 8px;
    overflow: hidden;
    background: var(--surface);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .thumb img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .thumb.empty {
    border: 1px dashed var(--border-strong);
    background: transparent;
  }
  .thumb-none {
    font-size: 11.5px;
    color: var(--fg-dim);
  }
  .size {
    margin: 0;
    font-size: 11.5px;
    color: var(--fg-dim);
    font-variant-numeric: tabular-nums;
  }
  .size {
    display: flex;
    flex-wrap: wrap;
    gap: 2px 8px;
  }
  .shape-note {
    color: var(--accent-strong);
  }
  .slot-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .slot-actions button {
    padding: 0.35rem 0.65rem;
    font-size: 0.78rem;
  }
  .confirm {
    padding: 8px;
    border: 1px solid var(--danger-line);
    border-radius: 8px;
    background: var(--danger-soft);
  }
  .confirm p,
  .said {
    margin: 0 0 8px;
    font-size: 12.5px;
    font-weight: 600;
    color: var(--fg);
  }
  .panel {
    margin: 0 0 12px;
    padding: 12px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--surface-sunken);
  }
</style>
