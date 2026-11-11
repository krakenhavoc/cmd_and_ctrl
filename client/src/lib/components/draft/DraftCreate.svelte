<script lang="ts">
  // DraftCreate — the create form, at #/draft (ADR 0222 §5, #5438).
  //
  // A set (the draftable ones only, newest first, from the public GET
  // /draft/sets), a pick mode, a number of seats and an optional name.
  // POST /draft answers with a seat token, kept under the draft's id,
  // and the page moves to the draft.
  //
  // Public: a visitor with no sign-in gets a solo draft, every other
  // seat a bot. A signed-in person gets an invite link.
  import { onMount } from "svelte";

  import {
    createDraft,
    DEFAULT_SEATS,
    defaultPickMode,
    draftErrorMessage,
    draftHash,
    DraftApiError,
    MAX_DRAFT_NAME,
    rememberDraftSeat,
    seatChoices,
    SET_NOT_DRAFTABLE_CODE,
    storedDraftIDs,
    type DraftMode,
  } from "../../draft";
  import {
    boosterName,
    DRAFT_SETS_HASH,
    fetchDraftSets,
    PACKS_ARE_APPROXIMATIONS,
    readySets,
    releaseYear,
    type DraftSet,
  } from "../../draftSets";
  import { L } from "../../labels";
  import { signedInUserID } from "../../myGames";
  import { navigate } from "../../router";
  import { session } from "../../session";
  import Icon from "../Icon.svelte";
  import Page from "../Page.svelte";

  interface Props {
    /** ?set=<code>: the set the sets page's link chose. */
    set?: string;
  }
  const { set: wanted }: Props = $props();

  let sets = $state<DraftSet[] | null>(null);
  let loadError = $state("");
  let loading = $state(true);

  // svelte-ignore state_referenced_locally
  let setCode = $state(wanted?.toLowerCase() ?? "");
  let seats = $state(DEFAULT_SEATS);
  // The mode follows the seat count's default until the person chooses.
  let modeChosen = $state<DraftMode | null>(null);
  const mode = $derived(modeChosen ?? defaultPickMode(seats));
  let name = $state("");
  // The length of the draft's matches (#5440): one game, or best of three.
  let bestOf = $state<1 | 3>(1);
  let busy = $state(false);
  let error = $state("");
  let notDraftable = $state(false);

  const ready = $derived(readySets(sets ?? []));
  const signedIn = $derived(signedInUserID($session) !== null);
  // The set the link asked for, when it cannot be drafted (or is unknown).
  const wantedMissing = $derived(
    Boolean(wanted) && sets !== null && !ready.some((s) => s.code === wanted?.toLowerCase()),
  );
  // Drafts this browser still holds a seat in: the way back to one.
  const held = storedDraftIDs();

  onMount(() => {
    const ctrl = new AbortController();
    fetchDraftSets(ctrl.signal)
      .then((r) => {
        sets = r.sets;
        const open = readySets(r.sets);
        if (!open.some((s) => s.code === setCode)) setCode = open[0]?.code ?? "";
      })
      .catch((err: unknown) => {
        if (ctrl.signal.aborted) return;
        loadError = err instanceof Error ? err.message : "couldn’t load the sets";
      })
      .finally(() => {
        loading = false;
      });
    return () => ctrl.abort();
  });

  function setLabel(s: DraftSet): string {
    const bits = [s.code.toUpperCase(), releaseYear(s.released), boosterName(s.booster)];
    return `${s.name} (${bits.filter(Boolean).join(", ")})`;
  }

  async function submit(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    if (busy || !setCode) return;
    busy = true;
    error = "";
    notDraftable = false;
    try {
      const made = await createDraft({ set: setCode, mode, seats, name, best_of: bestOf });
      rememberDraftSeat(made.draft.id, made);
      navigate(draftHash(made.draft.id));
    } catch (err) {
      notDraftable = err instanceof DraftApiError && err.code === SET_NOT_DRAFTABLE_CODE;
      error = draftErrorMessage(err, "Couldn’t create the draft. Try again in a moment.");
    } finally {
      busy = false;
    }
  }
</script>

<Page
  eyebrow="Draft"
  title={L.draftNewTitle}
  lede="Open packs of one set, take a card, pass the rest on. Three packs each, and the cards you take are your pool: build a 40-card deck from it and play three rounds."
  width="prose"
>
  <div class="stack">
    {#if held.length > 0}
      <p class="held">
        <span>You have a draft open in this browser.</span>
        <a class="link-arrow" href={draftHash(held[held.length - 1])}>{L.draftReturn} →</a>
      </p>
    {/if}

    {#if signedIn}
      <p class="held">
        <span>The drafts you have finished are kept for you.</span>
        <a class="link-arrow" href="#/drafts">{L.myDrafts} →</a>
      </p>
    {/if}

    {#if loading}
      <p class="help" aria-live="polite">Loading the sets…</p>
    {:else if loadError}
      <p class="notice err" role="alert">
        <Icon name="x" size={14} />
        <span>Couldn’t load the sets ({loadError}). Reload the page to try again.</span>
      </p>
    {:else if ready.length === 0}
      <div class="card">
        <p class="help strong">No set can be drafted yet.</p>
        <p class="help">
          A set opens for drafting once every card its packs can contain is fully built.
        </p>
        <a class="link-arrow" href={DRAFT_SETS_HASH}>{L.draftSeeSets} →</a>
      </div>
    {:else}
      <form class="card form" onsubmit={submit}>
        {#if wantedMissing}
          <p class="notice warn" role="status">
            <Icon name="warn" size={14} />
            <span>
              That set can’t be drafted yet, so another one is chosen below.
              <a href={DRAFT_SETS_HASH}>{L.draftSeeSets}</a>.
            </span>
          </p>
        {/if}

        <div class="field">
          <label for="draft-set">{L.draftSetField}</label>
          <select id="draft-set" bind:value={setCode} disabled={busy}>
            {#each ready as s (s.code)}
              <option value={s.code}>{setLabel(s)}</option>
            {/each}
          </select>
          <p class="hint">
            Only sets that are ready to draft are listed.
            <a href={DRAFT_SETS_HASH}>{L.draftSeeSets}</a>.
          </p>
        </div>

        <div class="field">
          <label for="draft-seats">{L.draftSeatsField}</label>
          <select id="draft-seats" bind:value={seats} disabled={busy}>
            {#each seatChoices() as n (n)}
              <option value={n}>{n}</option>
            {/each}
          </select>
          <p class="hint" data-testid="bots-note">
            {#if signedIn}
              You get a link to invite people. Every seat nobody takes is filled by a bot when you
              start.
            {:else}
              You are not signed in, so this is a solo draft: every other seat is a bot. Sign in on
              the <a href="#/">front page</a> to invite people.
            {/if}
          </p>
        </div>

        <fieldset class="field modes" disabled={busy}>
          <legend>{L.draftPickMode}</legend>
          <label class="mode" class:on={mode === "pick_one"}>
            <input
              type="radio"
              name="draft-mode"
              value="pick_one"
              checked={mode === "pick_one"}
              onchange={() => (modeChosen = "pick_one")}
            />
            <span class="mode-text">
              <span class="mode-name">{L.draftPickOne}</span>
              <span class="mode-what">Take one card from each pack, then pass it on.</span>
            </span>
          </label>
          <label class="mode" class:on={mode === "pick_two"}>
            <input
              type="radio"
              name="draft-mode"
              value="pick_two"
              checked={mode === "pick_two"}
              onchange={() => (modeChosen = "pick_two")}
            />
            <span class="mode-text">
              <span class="mode-name">{L.draftPickTwo}</span>
              <span class="mode-what"
                >Take two cards from each pack: a quicker draft, made for four seats.</span
              >
            </span>
          </label>
        </fieldset>

        <fieldset class="field modes" disabled={busy}>
          <legend>{L.draftMatchLength}</legend>
          <label class="mode" class:on={bestOf === 1}>
            <input
              type="radio"
              name="draft-best-of"
              value="1"
              checked={bestOf === 1}
              onchange={() => (bestOf = 1)}
            />
            <span class="mode-text">
              <span class="mode-name">{L.matchBestOfOne}</span>
              <span class="mode-what"
                >After the draft, three rounds of one game each with the deck you build.</span
              >
            </span>
          </label>
          <label class="mode" class:on={bestOf === 3}>
            <input
              type="radio"
              name="draft-best-of"
              value="3"
              checked={bestOf === 3}
              onchange={() => (bestOf = 3)}
            />
            <span class="mode-text">
              <span class="mode-name">{L.matchBestOfThree}</span>
              <span class="mode-what"
                >Each round is first to two games, with sideboarding from the rest of your pool
                between them.</span
              >
            </span>
          </label>
        </fieldset>

        <div class="field">
          <label for="draft-name">{L.draftNameField} <span class="opt">optional</span></label>
          <input
            id="draft-name"
            type="text"
            bind:value={name}
            maxlength={MAX_DRAFT_NAME}
            autocomplete="off"
            placeholder="Friday draft"
            disabled={busy}
          />
        </div>

        {#if error}
          <p class="notice err" role="alert">
            <Icon name="x" size={14} />
            <span>
              {error}
              {#if notDraftable}<a href={DRAFT_SETS_HASH}>{L.draftSeeSets}</a>.{/if}
            </span>
          </p>
        {/if}

        <div class="frow">
          <button type="submit" class="btn primary" disabled={busy || !setCode}>
            {busy ? "Creating…" : L.draftCreate}
          </button>
        </div>
      </form>
    {/if}

    <p class="approx" data-testid="packs-approximate">{PACKS_ARE_APPROXIMATIONS}</p>
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
    gap: 20px;
  }
  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
    margin: 0;
    padding: 0;
    border: 0;
  }
  label,
  legend {
    padding: 0;
    font-size: 0.8125rem;
    font-weight: 700;
    color: var(--fg);
  }
  legend {
    margin-bottom: 6px;
  }
  .opt {
    margin-left: 6px;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  select,
  input[type="text"] {
    width: 100%;
    max-width: 44ch;
    margin: 0;
    box-sizing: border-box;
  }
  #draft-seats {
    max-width: 10ch;
  }
  .hint,
  .help {
    margin: 0;
    font-size: 0.8125rem;
    line-height: 1.6;
    color: var(--fg-muted);
  }
  .help.strong {
    font-weight: 700;
    color: var(--fg);
  }
  .card .help + .help {
    margin-top: 6px;
  }
  .card .link-arrow {
    margin-top: 12px;
  }
  .hint a,
  .notice a {
    color: var(--accent-strong);
  }
  .modes {
    gap: 8px;
  }
  .mode {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 10px 12px;
    border-radius: var(--radius);
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    font-weight: 400;
    cursor: pointer;
  }
  .mode.on {
    border-color: var(--accent-line);
    background: var(--accent-soft);
  }
  .mode:has(input:focus-visible) {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }
  .mode input {
    flex: none;
    margin: 3px 0 0;
    accent-color: var(--accent);
  }
  .mode-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }
  .mode-name {
    font-weight: 700;
    color: var(--fg);
  }
  .mode-what {
    font-size: 0.8125rem;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .frow {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }
  .held {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px 14px;
    margin: 0;
    padding: 10px 14px;
    border-radius: var(--radius);
    background: var(--surface);
    border: 1px solid var(--border-strong);
    font-size: 0.875rem;
    color: var(--fg);
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
  .notice.warn {
    background: var(--surface-sunken);
    border: 1px solid var(--border-strong);
  }
  .approx {
    margin: 0;
    font-size: 0.8125rem;
    line-height: 1.6;
    color: var(--fg-muted);
  }
</style>
