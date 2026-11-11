<script lang="ts">
  // DraftJoin — the invite page, at #/draft/<id>?invite=<code> for a
  // browser that holds no seat in that draft (ADR 0222 §5, #5438).
  //
  // GET /draft/{id}?invite= shows what the link leads to before anything
  // is asked. A signed-in person sits under their Discord name; anyone
  // else types one. POST /draft/{id}/join answers with the seat token,
  // which the route keeps and connects with.
  import { onMount } from "svelte";

  import {
    DRAFT_NEW_HASH,
    DraftApiError,
    draftErrorMessage,
    isDraftGone,
    joinDraft,
    MAX_GUEST_NAME,
    pickModeName,
    previewDraft,
    type DraftPreview,
    type DraftSession,
  } from "../../draft";
  import { loadGuestName, rememberGuestName } from "../../guestName";
  import { L } from "../../labels";
  import { signedInUserID } from "../../myGames";
  import { session } from "../../session";
  import Icon from "../Icon.svelte";
  import Page from "../Page.svelte";

  interface Props {
    id: string;
    invite: string;
    onjoined: (s: DraftSession) => void;
  }
  const { id, invite, onjoined }: Props = $props();

  let preview = $state<DraftPreview | null>(null);
  let loading = $state(true);
  let gone = $state(false);
  let loadError = $state("");
  let name = $state(loadGuestName());
  let busy = $state(false);
  let error = $state("");

  const signedIn = $derived(signedInUserID($session) !== null);
  const started = $derived(preview !== null && preview.state !== "lobby");
  const full = $derived(preview !== null && preview.taken >= preview.seats);

  onMount(() => {
    const ctrl = new AbortController();
    previewDraft(id, { invite }, ctrl.signal)
      .then((p) => {
        preview = p;
      })
      .catch((err: unknown) => {
        if (ctrl.signal.aborted) return;
        if (isDraftGone(err)) gone = true;
        else loadError = draftErrorMessage(err, "Couldn’t load the draft.");
      })
      .finally(() => {
        loading = false;
      });
    return () => ctrl.abort();
  });

  async function submit(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    if (busy) return;
    busy = true;
    error = "";
    try {
      const s = await joinDraft(id, invite, signedIn ? undefined : name);
      if (!signedIn) rememberGuestName(name);
      onjoined(s);
    } catch (err) {
      if (err instanceof DraftApiError && err.status === 404) gone = true;
      else error = draftErrorMessage(err, "Couldn’t join the draft. Try again in a moment.");
    } finally {
      busy = false;
    }
  }
</script>

<Page eyebrow="Draft" title={gone ? L.draftGone : (preview?.name ?? "Join a draft")} width="prose">
  <div class="stack">
    {#if loading}
      <p class="help" aria-live="polite">Loading the draft…</p>
    {:else if gone}
      <div class="card">
        <p class="help">
          That invite doesn’t lead to a draft any more. The draft may have closed, or the link may
          be incomplete. Ask the host for a fresh link.
        </p>
        <a class="link-arrow" href={DRAFT_NEW_HASH}>{L.draftNewDraft} →</a>
      </div>
    {:else}
      {#if loadError}
        <p class="notice err" role="alert">
          <Icon name="x" size={14} />
          <span>{loadError}</span>
        </p>
      {/if}
      {#if preview}
        <dl class="card facts" data-testid="draft-preview">
          <div>
            <dt>Set</dt>
            <dd>{preview.set_name ?? preview.set.toUpperCase()}</dd>
          </div>
          <div>
            <dt>Picks</dt>
            <dd>{pickModeName(preview.mode)}</dd>
          </div>
          <div>
            <dt>Seats</dt>
            <dd>{preview.taken} of {preview.seats} taken</dd>
          </div>
          {#if preview.host}
            <div>
              <dt>Host</dt>
              <dd>{preview.host}</dd>
            </div>
          {/if}
        </dl>
      {/if}

      <form class="card form" onsubmit={submit}>
        {#if started}
          <p class="help">
            This draft has started, so no new seat can be taken. If you already hold a seat in it,
            joining puts you back in it.
          </p>
        {:else if full}
          <p class="help">Every seat is taken. If one of them is yours, joining puts you back.</p>
        {:else}
          <p class="help">Take a seat. Seats nobody takes are filled by bots at the start.</p>
        {/if}

        {#if signedIn}
          <p class="help">
            You sit as <strong>{$session?.principal.name ?? "your Discord name"}</strong>.
          </p>
        {:else}
          <div class="field">
            <label for="draft-guest-name">{L.draftGuestName}</label>
            <input
              id="draft-guest-name"
              type="text"
              bind:value={name}
              maxlength={MAX_GUEST_NAME}
              autocomplete="nickname"
              placeholder="Player"
              disabled={busy}
            />
          </div>
        {/if}

        {#if error}
          <p class="notice err" role="alert">
            <Icon name="x" size={14} />
            <span>{error}</span>
          </p>
        {/if}

        <div class="frow">
          <button type="submit" class="btn primary" disabled={busy}>
            {busy ? "Joining…" : L.draftJoin}
          </button>
        </div>
      </form>
    {/if}
  </div>
</Page>

<style>
  .stack {
    display: flex;
    flex-direction: column;
    gap: 22px;
    min-width: 0;
  }
  .form {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .facts {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: 14px 24px;
    margin: 0;
  }
  dt {
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  dd {
    margin: 4px 0 0;
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  label {
    font-size: 0.8125rem;
    font-weight: 700;
    color: var(--fg);
  }
  input[type="text"] {
    width: 100%;
    max-width: 32ch;
    margin: 0;
    box-sizing: border-box;
  }
  .help {
    margin: 0;
    font-size: 0.875rem;
    line-height: 1.6;
    color: var(--fg-muted);
  }
  .help strong {
    color: var(--fg);
  }
  .card .link-arrow {
    margin-top: 12px;
  }
  .frow {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }
  .notice {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 0;
    padding: 10px 14px;
    border-radius: var(--radius);
    font-size: 0.8125rem;
    line-height: 1.5;
    color: var(--fg);
  }
  .notice.err {
    background: color-mix(in srgb, var(--danger) 8%, transparent);
    border: 1px solid color-mix(in srgb, var(--danger) 35%, transparent);
  }
</style>
