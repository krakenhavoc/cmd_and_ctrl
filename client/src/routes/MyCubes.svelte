<script lang="ts">
  // MyCubes — a signed-in person's saved cubes (ADR 0222 §8, #5441), at
  // #/cubes.
  //
  // A cube is a named list of cards a draft can deal its packs from in
  // place of a set's. The page adds one (a pasted list or a deck link),
  // lists them with the server's check of each, names the cards that
  // block one so its owner can drop them, and deletes one. Everything
  // reads and writes /me/cubes, which answers for the caller's own cubes
  // only. A visitor has none: a cube needs an account to belong to.
  import { onMount } from "svelte";

  import Icon from "../lib/components/Icon.svelte";
  import Page from "../lib/components/Page.svelte";
  import {
    createCube,
    cubeBlockerCount,
    cubeCards,
    cubeNeed,
    cubeStatus,
    CUBE_LIMITS,
    deleteCube,
    dropCards,
    fetchMyCube,
    fetchMyCubes,
    updateCube,
    type Cube,
    type CubeLimits,
  } from "../lib/cubes";
  import { DRAFT_NEW_HASH, draftNewCubeHash, MIN_SEATS } from "../lib/draft";
  import { L } from "../lib/labels";
  import { signedInUserID } from "../lib/myGames";
  import { session } from "../lib/session";

  const signedIn = $derived(signedInUserID($session) !== null);

  let cubes = $state<Cube[]>([]);
  let limits = $state<CubeLimits>(CUBE_LIMITS);
  let loading = $state(true);
  let failed = $state("");

  // The new-cube form.
  let name = $state("");
  let from = $state<"text" | "url">("text");
  let text = $state("");
  let url = $state("");
  let singleton = $state(true);
  let saving = $state(false);
  let saveError = $state("");

  // One cube open at a time: its check in full, with its list.
  let open = $state<Cube | null>(null);
  let busyID = $state("");
  let rowError = $state<{ id: string; message: string } | null>(null);
  let editing = $state(false);
  let editText = $state("");
  let confirming = $state("");
  // What the page last did, for a screen reader.
  let said = $state("");

  const full = $derived(cubes.length >= limits.max_cubes);
  const smallest = $derived(cubeNeed(MIN_SEATS, limits.pack_size, limits.packs_per_seat));
  const largest = $derived(cubeNeed(8, limits.pack_size, limits.packs_per_seat));

  const message = (err: unknown, fallback: string): string =>
    err instanceof Error && err.message ? err.message : fallback;

  async function load(signal?: AbortSignal): Promise<void> {
    try {
      const r = await fetchMyCubes(signal);
      cubes = r.cubes;
      limits = r;
      failed = "";
    } catch (err) {
      if (signal?.aborted) return;
      failed = message(err, "Couldn’t load your cubes. Try again in a moment.");
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    if (!signedIn) {
      loading = false;
      return;
    }
    const ctrl = new AbortController();
    void load(ctrl.signal);
    return () => ctrl.abort();
  });

  // put replaces a cube in the list with a fresh copy, kept in name order.
  function put(c: Cube): void {
    const { list: _list, ...row } = c;
    cubes = [...cubes.filter((x) => x.id !== c.id), row].sort((a, b) =>
      a.name.localeCompare(b.name),
    );
  }

  async function save(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    if (saving) return;
    saving = true;
    saveError = "";
    try {
      const made = await createCube(
        from === "url" ? { name, url, singleton } : { name, text, singleton },
      );
      put(made);
      open = made;
      editing = false;
      name = text = url = "";
      singleton = true;
      said = `Saved ${made.name}: ${cubeStatus(made.check).toLowerCase()}.`;
    } catch (err) {
      saveError = message(err, "Couldn’t save the cube. Try again in a moment.");
    } finally {
      saving = false;
    }
  }

  // act runs one change to a cube, with the row's own busy and error.
  async function act(id: string, what: () => Promise<void>): Promise<void> {
    if (busyID) return;
    busyID = id;
    rowError = null;
    try {
      await what();
    } catch (err) {
      rowError = { id, message: message(err, "That didn’t work. Try again in a moment.") };
    } finally {
      busyID = "";
    }
  }

  const check = (id: string): Promise<void> =>
    act(id, async () => {
      const c = await fetchMyCube(id);
      put(c);
      open = c;
      editing = false;
      said = `${c.name}: ${cubeStatus(c.check).toLowerCase()}.`;
    });

  const saveList = (c: Cube, list: string, done: string): Promise<void> =>
    act(c.id, async () => {
      const next = await updateCube(c.id, { text: list });
      put(next);
      open = next;
      editing = false;
      said = `${done} ${next.name}: ${cubeStatus(next.check).toLowerCase()}.`;
    });

  const setSingleton = (c: Cube, on: boolean): Promise<void> =>
    act(c.id, async () => {
      const next = await updateCube(c.id, { singleton: on });
      put(next);
      open = next;
    });

  const remove = (c: Cube): Promise<void> =>
    act(c.id, async () => {
      await deleteCube(c.id);
      cubes = cubes.filter((x) => x.id !== c.id);
      if (open?.id === c.id) open = null;
      confirming = "";
      said = `Deleted ${c.name}.`;
    });

  function startEdit(c: Cube): void {
    editText = c.list ?? "";
    editing = true;
  }
</script>

<Page
  eyebrow="Draft"
  title={L.myCubes}
  lede="A cube is your own list of cards to draft: it is shuffled and dealt into packs in place of a set’s. Only you can see your cubes."
  width="wide"
>
  <div class="stack">
    <p class="sr-only" role="status" aria-live="polite">{said}</p>

    {#if !signedIn}
      <div class="card empty" data-testid="cubes-signed-out">
        <p class="help strong">Sign in to keep a cube.</p>
        <p class="help">
          A cube is saved to an account, so it needs a sign-in. You can draft a set without one.
        </p>
        <a class="link-arrow" href="#/">Sign in on the front page</a>
        <a class="link-arrow" href={DRAFT_NEW_HASH}>{L.draftNewTitle}</a>
      </div>
    {:else if loading}
      <p class="help" role="status">Loading your cubes…</p>
    {:else if failed}
      <div class="card empty">
        <p class="help" role="alert">{failed}</p>
        <button type="button" class="btn sm" onclick={() => void load()}>Try again</button>
      </div>
    {:else}
      {#if cubes.length === 0}
        <div class="card empty">
          <p class="help strong">No cubes yet.</p>
          <p class="help">Add one below, then start a draft from it.</p>
        </div>
      {:else}
        <ul class="list" data-testid="my-cubes">
          {#each cubes as c (c.id)}
            {@const isOpen = open?.id === c.id}
            {@const busy = busyID === c.id}
            <li class="card cube" class:ready={c.check.ready}>
              <div class="head">
                <h2 class="title">{c.name}</h2>
                <span class="tag" class:ok={c.check.ready && c.check.max_seats > 0}>
                  {cubeStatus(c.check)}
                </span>
              </div>
              <p class="what">
                {cubeCards(c.check.cards)} · {c.singleton ? "singleton" : "any number of copies"}
              </p>

              <div class="row">
                {#if c.check.ready && c.check.max_seats > 0}
                  <a class="btn sm primary" href={draftNewCubeHash(c.id)}>{L.cubeDraft}</a>
                {/if}
                <button
                  type="button"
                  class="btn sm"
                  aria-expanded={isOpen}
                  disabled={busy}
                  onclick={() => void check(c.id)}
                >
                  {busy && !isOpen ? "Checking…" : L.cubeCheck}
                </button>
                {#if confirming === c.id}
                  <span class="ask">Delete {c.name}? This can’t be undone.</span>
                  <button
                    type="button"
                    class="btn sm danger"
                    disabled={busy}
                    onclick={() => void remove(c)}
                  >
                    {L.cubeDeleteConfirm}
                  </button>
                  <button type="button" class="btn sm ghost" onclick={() => (confirming = "")}>
                    Keep it
                  </button>
                {:else}
                  <button
                    type="button"
                    class="btn sm ghost"
                    disabled={busy}
                    onclick={() => (confirming = c.id)}
                  >
                    {L.cubeDelete}
                  </button>
                {/if}
              </div>

              {#if rowError?.id === c.id}
                <p class="notice err" role="alert">
                  <Icon name="x" size={14} />
                  <span>{rowError.message}</span>
                </p>
              {/if}

              {#if isOpen && open}
                {@const k = open.check}
                {@const droppable = [...k.unknown, ...k.not_full]}
                <div class="check" data-testid="cube-check">
                  {#if k.ready && k.max_seats > 0}
                    <p class="help">
                      Every card is ready. This cube can deal a draft of up to {k.max_seats} seats with
                      packs of {limits.pack_size}.
                    </p>
                  {:else if k.ready}
                    <p class="help">
                      Every card is ready, but {cubeCards(k.cards)} is too few to draft. The smallest
                      draft, {MIN_SEATS} seats with packs of {limits.pack_size}, needs {smallest}.
                      Smaller packs need fewer.
                    </p>
                  {:else if cubeBlockerCount(k) > 0}
                    <p class="help">
                      This cube can’t be drafted until the cards below are dropped or fixed.
                    </p>
                  {/if}
                  {#if k.too_large}
                    <p class="help">
                      {cubeCards(k.cards)} is over the {limits.max_cards} a cube can hold.
                    </p>
                  {/if}

                  {#if k.not_full.length > 0}
                    <section aria-labelledby="cube-notfull-{c.id}">
                      <h3 id="cube-notfull-{c.id}">
                        Not fully built yet <span class="n">{k.not_full.length}</span>
                      </h3>
                      <p class="help">
                        A draft never hands out a card that doesn’t work yet. These stop blocking
                        the cube by themselves once they are built.
                      </p>
                      <ul class="names">
                        {#each k.not_full as n (n)}<li>{n}</li>{/each}
                      </ul>
                    </section>
                  {/if}
                  {#if k.unknown.length > 0}
                    <section aria-labelledby="cube-unknown-{c.id}">
                      <h3 id="cube-unknown-{c.id}">
                        Not recognised <span class="n">{k.unknown.length}</span>
                      </h3>
                      <p class="help">No card has this name. Check the spelling.</p>
                      <ul class="names">
                        {#each k.unknown as n (n)}<li>{n}</li>{/each}
                      </ul>
                    </section>
                  {/if}
                  {#if k.repeated.length > 0}
                    <section aria-labelledby="cube-repeated-{c.id}">
                      <h3 id="cube-repeated-{c.id}">
                        More than one copy <span class="n">{k.repeated.length}</span>
                      </h3>
                      <p class="help">
                        A singleton cube holds one copy of each card, basic lands aside. Cut the
                        extra copies, or turn singleton off.
                      </p>
                      <ul class="names">
                        {#each k.repeated as n (n)}<li>{n}</li>{/each}
                      </ul>
                    </section>
                  {/if}

                  <label class="tick">
                    <input
                      type="checkbox"
                      checked={open.singleton}
                      disabled={busy}
                      onchange={(e) => void setSingleton(c, e.currentTarget.checked)}
                    />
                    <span>{L.cubeSingleton}</span>
                  </label>

                  {#if editing}
                    <div class="field">
                      <label for="cube-edit-{c.id}">{L.cubeListField}</label>
                      <textarea
                        id="cube-edit-{c.id}"
                        rows="12"
                        spellcheck="false"
                        bind:value={editText}
                        disabled={busy}
                      ></textarea>
                    </div>
                    <div class="row">
                      <button
                        type="button"
                        class="btn sm primary"
                        disabled={busy || editText.trim() === ""}
                        onclick={() => open && void saveList(open, editText, "Saved.")}
                      >
                        {busy ? "Saving…" : L.cubeSaveList}
                      </button>
                      <button type="button" class="btn sm ghost" onclick={() => (editing = false)}>
                        Cancel
                      </button>
                    </div>
                  {:else}
                    <div class="row">
                      {#if droppable.length > 0}
                        <button
                          type="button"
                          class="btn sm"
                          disabled={busy}
                          onclick={() =>
                            open &&
                            void saveList(
                              open,
                              dropCards(open.list ?? "", droppable),
                              `Dropped ${cubeCards(droppable.length)}.`,
                            )}
                        >
                          {L.cubeDropBlockers}
                        </button>
                      {/if}
                      <button
                        type="button"
                        class="btn sm"
                        disabled={busy}
                        onclick={() => open && startEdit(open)}
                      >
                        {L.cubeEditList}
                      </button>
                    </div>
                  {/if}
                </div>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}

      <form class="card form" onsubmit={save} aria-labelledby="cube-new-h">
        <h2 id="cube-new-h" class="title">{L.cubeNew}</h2>
        {#if full}
          <p class="help" data-testid="cubes-full">
            You have {limits.max_cubes} cubes, the most you can keep. Delete one to add another.
          </p>
        {:else}
          <div class="field">
            <label for="cube-name">
              {L.cubeNameField}
              {#if from === "url"}<span class="opt">optional with a link</span>{/if}
            </label>
            <input
              id="cube-name"
              type="text"
              bind:value={name}
              maxlength={limits.max_name_length}
              autocomplete="off"
              placeholder="Pauper cube"
              disabled={saving}
            />
          </div>

          <fieldset class="field" disabled={saving}>
            <legend>Where the list comes from</legend>
            <div class="row">
              <label class="tick">
                <input type="radio" name="cube-from" value="text" bind:group={from} />
                <span>Paste a list</span>
              </label>
              <label class="tick">
                <input type="radio" name="cube-from" value="url" bind:group={from} />
                <span>A deck link</span>
              </label>
            </div>
          </fieldset>

          {#if from === "text"}
            <div class="field">
              <label for="cube-list">{L.cubeListField}</label>
              <textarea
                id="cube-list"
                rows="10"
                spellcheck="false"
                bind:value={text}
                placeholder={"Lightning Bolt\nCounterspell\n2 Llanowar Elves\n…"}
                aria-describedby="cube-list-hint"
                disabled={saving}
              ></textarea>
              <p class="hint" id="cube-list-hint">
                One card a line, with a count in front if you want more than one copy. A deck list
                works too: every card in it joins the cube, whatever section it was in.
              </p>
            </div>
          {:else}
            <div class="field">
              <label for="cube-link">{L.cubeLinkField}</label>
              <input
                id="cube-link"
                type="text"
                inputmode="url"
                bind:value={url}
                placeholder="https://archidekt.com/decks/123456"
                aria-describedby="cube-link-hint"
                disabled={saving}
              />
              <p class="hint" id="cube-link-hint">
                A link to a list on a deck site the deck checker reads. If a site refuses the link,
                paste the list instead.
              </p>
            </div>
          {/if}

          <label class="tick">
            <input type="checkbox" bind:checked={singleton} disabled={saving} />
            <span>{L.cubeSingleton}</span>
          </label>

          <p class="hint" data-testid="cube-sizes">
            A draft deals {limits.packs_per_seat} packs a seat: {smallest} cards for {MIN_SEATS} seats
            and {largest} for 8, with packs of {limits.pack_size}. A cube holds at most {limits.max_cards}
            cards, and every card has to be fully built before the cube can be drafted.
          </p>

          {#if saveError}
            <p class="notice err" role="alert">
              <Icon name="x" size={14} />
              <span>{saveError}</span>
            </p>
          {/if}

          <div class="row">
            <button
              type="submit"
              class="btn"
              class:primary={cubes.length === 0}
              disabled={saving || (from === "text" ? text.trim() === "" : url.trim() === "")}
            >
              {saving ? "Saving…" : L.cubeSave}
            </button>
          </div>
        {/if}
      </form>

      <a class="link-arrow" href={DRAFT_NEW_HASH}>{L.draftNewTitle}</a>
    {/if}
  </div>
</Page>

<style>
  .stack {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 22px;
    min-width: 0;
  }
  .stack > .link-arrow {
    align-self: flex-start;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    margin: -1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }
  .title {
    margin: 0;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    /* a .card's h2 is a mono eyebrow by default; these are headings */
    letter-spacing: -0.005em;
    text-transform: none;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  h3 {
    margin: 0 0 4px;
    font-size: 0.875rem;
    font-weight: 700;
    color: var(--fg);
  }
  .n {
    margin-left: 6px;
    font-family: var(--font-mono);
    font-size: 12px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .help,
  .hint,
  .what {
    margin: 0;
    font-size: 0.8125rem;
    line-height: 1.6;
    color: var(--fg-muted);
  }
  .help.strong {
    font-weight: 700;
    color: var(--fg);
  }
  .empty,
  .form,
  .cube,
  .check {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  .form {
    gap: 18px;
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px 12px;
  }
  .tag.ok {
    border-color: var(--accent-line);
    color: var(--accent-strong);
  }
  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px 10px;
  }
  .ask {
    font-size: 0.8125rem;
    color: var(--fg);
  }
  .check {
    align-self: stretch;
    gap: 14px;
    margin-top: 6px;
    padding-top: 14px;
    border-top: 1px solid var(--border);
  }
  .check section {
    align-self: stretch;
  }
  .names {
    margin: 8px 0 0;
    padding: 0;
    list-style: none;
    columns: 3 22ch;
    column-gap: 24px;
    font-size: 0.8125rem;
    line-height: 1.7;
    color: var(--fg);
  }
  .names li {
    break-inside: avoid;
    overflow-wrap: anywhere;
  }
  .field {
    display: flex;
    flex-direction: column;
    align-self: stretch;
    gap: 6px;
    min-width: 0;
    margin: 0;
    padding: 0;
    border: 0;
  }
  .field > label,
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
  input[type="text"],
  textarea {
    width: 100%;
    max-width: 64ch;
    margin: 0;
    box-sizing: border-box;
  }
  textarea {
    font-family: var(--font-mono);
    font-size: 13px;
    line-height: 1.5;
    resize: vertical;
  }
  .tick {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 0.875rem;
    font-weight: 400;
    color: var(--fg);
    cursor: pointer;
  }
  .tick input {
    flex: none;
    margin: 0;
    accent-color: var(--accent);
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
