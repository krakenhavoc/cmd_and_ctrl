<script lang="ts">
  import { onDestroy, onMount, untrack } from "svelte";
  import {
    fetchMyDecks,
    fetchPrebuiltDecks,
    installPrebuiltDeck,
    seatLibraryDeck,
    uploadDeck,
    type UploadDeckResponse,
  } from "../api";
  import { LobbyApiError, currentSession, session, type ApiViolation } from "../session";
  import { loadLastDeck, preselectFor, rememberGuestLastDeck, type LastDeck } from "../lastDeck";
  import { signedInUserID } from "../myGames";
  import { isSignedIn, sourceHost, type MyDeckInfo } from "../myDecks";
  import { imperfectCardLine, summariseCoverage, type PrebuiltDeck } from "../prebuiltDecks";
  import { sortDecks } from "../deckGrid";
  import {
    FILTER_FROM,
    firstTab,
    libraryDetail,
    libraryTile,
    pickerTabs,
    prebuiltTile,
    sameName,
    tilePreview,
    type PickerCommit,
    type PickerState,
    type PickerTab,
  } from "../deckPicker";
  import { L } from "../labels";
  import DeckGrid from "./DeckGrid.svelte";
  import SideboardList from "./decks/SideboardList.svelte";
  import Icon from "./Icon.svelte";
  import ShopLinks from "./ShopLinks.svelte";

  // DeckPicker — the one deck picker (ADR 0212 §3): a grid of commander
  // art behind three tabs, My decks, Pre-built and Paste a list. It
  // replaces the three stacked pickers (YourDecksPicker, the pre-built
  // list and the paste box of DeckUploadForm), and both places that seat
  // a deck mount it: the lobby card, under the seats with a notch at your
  // seat, and the table route's deck dialog (Game.svelte).
  //
  //   - My decks is shown to a signed-in person with saved decks. A
  //     guest, or an empty library, gets Pre-built and Paste a list and
  //     opens on Pre-built (owner decision 7). Otherwise it opens on My
  //     decks, unless the last deck seated was a pre-built one, and that
  //     deck is preselected (lastDeck.ts, ADR 0110 §5 item 5).
  //   - Selecting a tile sends nothing. The picker tells its caller what
  //     is selected (`onchange`), so the lobby can preview it in your
  //     seat and make its footer's primary "Sit down with this deck";
  //     the caller seats it with `commit()`. With actions="inline" (the
  //     deck dialog) the picker draws that button itself.
  //   - The detail line under the grid says what a tile cannot: the
  //     coverage sentence, the source, a pre-built deck's summary and
  //     caveats, "Play this deck in paper" (ADR 0179 §7), and a saved
  //     deck's Sideboard (ADR 0191).
  //   - Paste a list keeps the textarea, the link handling and the
  //     Commander note; violations, warnings, the unimplemented list and
  //     the 200-deck library note render under it, as before.
  //
  // Every route it posts to (POST /games/{id}/decks, …/library) is
  // refused by the server once the game has started, so the caller only
  // mounts it while the table is in the lobby.

  interface Props {
    gameID: string;
    playerID: string;
    /**
     * After a deck is seated, with the server's answer, and whether the
     * picker has more to say about it (its warnings, or a pasted list's
     * unimplemented cards and library note) that a caller closing it
     * would hide.
     */
    onSuccess?: (res: UploadDeckResponse, more: boolean) => void;
    /** Offer "Play this deck in paper" on the pre-built decks (ADR 0179 §7); the lobby only. */
    shopLinks?: boolean;
    /** "inline" draws the commit button here; "external" leaves it to the caller's primary. */
    actions?: "inline" | "external";
    /** What is selected and what committing it says, whenever either changes. */
    onchange?: (state: PickerState) => void;
  }

  const {
    gameID,
    playerID,
    onSuccess,
    shopLinks = false,
    actions = "inline",
    onchange,
  }: Props = $props();

  const uid = $props.id();
  const signedIn = $derived(isSignedIn($session?.principal.user_id));

  // --- the lists, read once on mount -------------------------------
  //
  // The tabs wait for all three reads, so the picker never opens on one
  // tab and then jumps to another when the last deck arrives.
  let libDecks = $state<MyDeckInfo[]>([]);
  let preDecks = $state<PrebuiltDeck[]>([]);
  let lastDeck = $state<LastDeck | null>(null);
  let libLoaded = $state(false);
  let preLoaded = $state(false);
  let lastLoaded = $state(false);
  const loaded = $derived(libLoaded && preLoaded && lastLoaded);

  onMount(() => {
    void loadLastDeck(currentSession())
      .then((d) => (lastDeck = d))
      .finally(() => (lastLoaded = true));
    if (signedIn) {
      void fetchMyDecks()
        .then((res) => (libDecks = res.decks ?? []))
        // Session expired between render and fetch: no My decks tab.
        .catch(() => (libDecks = []))
        .finally(() => (libLoaded = true));
    } else {
      libLoaded = true;
    }
    void fetchPrebuiltDecks()
      .then((res) => (preDecks = res.decks ?? []))
      // A server without the route: Paste a list still works.
      .catch(() => (preDecks = []))
      .finally(() => (preLoaded = true));
  });

  // --- tabs ----------------------------------------------------------
  const tabs = $derived(pickerTabs(libDecks.length, preDecks.length));
  let chosenTab = $state<PickerTab | null>(null);
  const tab = $derived(
    chosenTab && tabs.includes(chosenTab) ? chosenTab : firstTab(tabs, lastDeck),
  );
  const TAB_LABEL: Record<PickerTab, string> = {
    library: L.myDecksTab,
    prebuilt: L.prebuiltTab,
    paste: L.pasteListTab,
  };

  function pickTab(t: PickerTab): void {
    if (t === tab) return;
    chosenTab = t;
    clearFeedback();
  }

  // The tabs pattern: the arrow keys, Home and End move between tabs.
  function onTabKey(e: KeyboardEvent): void {
    const i = tabs.indexOf(tab);
    const to =
      e.key === "ArrowRight"
        ? tabs[(i + 1) % tabs.length]
        : e.key === "ArrowLeft"
          ? tabs[(i - 1 + tabs.length) % tabs.length]
          : e.key === "Home"
            ? tabs[0]
            : e.key === "End"
              ? tabs[tabs.length - 1]
              : null;
    if (!to) return;
    e.preventDefault();
    pickTab(to);
    document.getElementById(`${uid}-tab-${to}`)?.focus();
  }

  // --- the two grids --------------------------------------------------
  let sort = $state<"recent" | "name">("recent");
  let libQuery = $state("");
  let preQuery = $state("");
  let libSel = $state("");
  let preSel = $state("");
  // Set once the player picks a tile, so a late preselect never moves it.
  let libTouched = false;
  let preTouched = false;

  const libTiles = $derived(sortDecks(libDecks, sort).map(libraryTile));
  const preTiles = $derived(preDecks.map(prebuiltTile));

  // Preselect the last deck once the list and the last deck are in,
  // else keep what is selected, else the first (the newest saved deck).
  $effect(() => {
    const ids = libDecks.map((d) => d.id);
    const pick = preselectFor("library", ids, lastDeck);
    untrack(() => {
      if (libTouched) return;
      libSel = pick ?? (ids.includes(libSel) ? libSel : (ids[0] ?? ""));
    });
  });
  $effect(() => {
    const ids = preDecks.map((d) => d.id);
    const pick = preselectFor("prebuilt", ids, lastDeck);
    untrack(() => {
      if (preTouched) return;
      preSel = pick ?? (ids.includes(preSel) ? preSel : (ids[0] ?? ""));
    });
  });

  const libCurrent = $derived(libDecks.find((d) => d.id === libSel));
  const preCurrent = $derived(preDecks.find((d) => d.id === preSel));
  const currentTile = $derived(
    tab === "library"
      ? libTiles.find((t) => t.id === libSel)
      : tab === "prebuilt"
        ? preTiles.find((t) => t.id === preSel)
        : undefined,
  );

  // ADR 0191 (amended 2026-10-19, #3833): the selected saved deck's
  // sideboard, read only. It closes when another deck is selected, so it
  // is never the last deck's list under this deck's name.
  let sideboardOpen = $state(false);
  let sideboardFor = $state("");
  $effect(() => {
    if (libSel !== sideboardFor) sideboardOpen = false;
  });

  // --- Paste a list --------------------------------------------------
  let source = $state("");
  const link = $derived(looksLikeURL(source));

  function looksLikeURL(s: string): boolean {
    const t = s.trim();
    return t.startsWith("http://") || t.startsWith("https://");
  }

  function linkHost(s: string): string {
    try {
      return new URL(s.trim()).hostname.replace(/^www\./, "");
    } catch {
      return s.trim();
    }
  }

  // --- what committing says, for the caller's primary ----------------
  let busy = $state(false);
  const commitState = $derived.by((): PickerCommit | null => {
    if (!loaded) return null;
    if (tab === "paste") {
      const empty = source.trim() === "";
      return {
        label: busy
          ? link
            ? `Fetching from ${linkHost(source)}…`
            : "Uploading…"
          : link
            ? L.importDeck
            : L.uploadDeck,
        disabled: busy || empty,
        empty,
        title: empty
          ? "paste a decklist or a deck link first"
          : link
            ? "import the deck at this link and sit down with it"
            : "check this decklist and sit down with it",
        busy,
      };
    }
    const t = currentTile;
    if (!t) return null;
    return {
      label: busy ? "Seating…" : L.sitDownWithThisDeck,
      disabled: busy,
      title: `sit down with ${t.name}`,
      busy,
    };
  });
  const preview = $derived(currentTile ? tilePreview(currentTile) : null);

  $effect(() => {
    const state: PickerState = { commit: commitState, preview };
    untrack(() => onchange?.(state));
  });
  onDestroy(() => onchange?.({ commit: null, preview: null }));

  // --- seating it ----------------------------------------------------
  let errorMessage = $state("");
  let violations = $state<ApiViolation[]>([]);
  let warnings = $state<ApiViolation[]>([]);
  let successMessage = $state("");
  // Cards in the seated deck whose printed rules the engine does not
  // carry out. Rendered collapsed: the count is the part that sets
  // expectations; the list is for whoever wants the names.
  let unimplemented = $state<string[]>([]);
  // The deck was seated but not saved to the library: the 200-deck cap.
  let libraryNote = $state("");

  function clearFeedback(): void {
    errorMessage = "";
    violations = [];
    warnings = [];
    successMessage = "";
    unimplemented = [];
    libraryNote = "";
  }

  function violationLabel(v: ApiViolation): string {
    return v.card ? `${v.card}: ${v.message}` : v.message;
  }

  function pastedMessage(res: UploadDeckResponse): string {
    const cmd = `${res.commanders.length === 1 ? "commander" : "commanders"}: ${res.commanders.join(" / ")}`;
    const companion = res.companion ? `, companion: ${res.companion}` : "";
    const side = res.sideboard ? `, sideboard: ${res.sideboard} outside the game` : "";
    return `Uploaded ${res.deck_name || "deck"}: ${res.card_count} cards, ${cmd}${companion}${side}`;
  }

  /**
   * commit seats what the open tab holds: the selected saved or pre-built
   * deck, or the pasted list. The lobby's footer primary calls it.
   */
  export async function commit(): Promise<void> {
    const state = commitState;
    if (busy || !state || state.disabled) return;
    const which = tab;
    const name = currentTile?.name ?? "";
    busy = true;
    clearFeedback();
    try {
      let res: UploadDeckResponse;
      if (which === "paste") {
        res = await uploadDeck(gameID, playerID, source);
        successMessage = pastedMessage(res);
        source = "";
      } else if (which === "library") {
        res = await seatLibraryDeck(gameID, libSel);
        const side = res.sideboard ? `, with ${res.sideboard} outside the game` : "";
        successMessage = `Playing ${res.deck_name || name}${side}`;
      } else {
        res = await installPrebuiltDeck(gameID, playerID, preSel);
        successMessage = `Playing ${res.deck_name || name}`;
        // A signed-in person's last deck is recorded by the server; a
        // guest's lives in this browser (ADR 0110 §5 item 5).
        if (signedInUserID($session) === null) rememberGuestLastDeck(res.deck_id || preSel);
      }
      if (res.warnings?.length) warnings = res.warnings;
      if (res.unimplemented?.length) unimplemented = res.unimplemented;
      if (res.library_note) libraryNote = res.library_note;
      const more =
        warnings.length > 0 ||
        (which === "paste" && (unimplemented.length > 0 || libraryNote !== ""));
      onSuccess?.(res, more);
    } catch (err) {
      if (err instanceof LobbyApiError) {
        errorMessage = err.message;
        if (err.violations?.length) violations = err.violations;
        if (err.warnings?.length) warnings = err.warnings;
      } else {
        errorMessage =
          which === "paste"
            ? "upload failed"
            : which === "library"
              ? "could not seat that deck"
              : "could not install that deck";
      }
    } finally {
      busy = false;
    }
  }
</script>

<div class="deck-picker">
  {#if !loaded}
    <p class="note loading" role="status">Loading decks…</p>
  {:else}
    <div class="pbar">
      <div class="tabs" role="tablist" tabindex="-1" onkeydown={onTabKey}>
        {#each tabs as t (t)}
          <button
            type="button"
            class="tab"
            role="tab"
            id={`${uid}-tab-${t}`}
            aria-selected={tab === t}
            aria-controls={`${uid}-panel`}
            tabindex={tab === t ? 0 : -1}
            onclick={() => pickTab(t)}
          >
            {TAB_LABEL[t]}
            {#if t === "library"}<span class="n">{libDecks.length}</span>{/if}
            {#if t === "prebuilt"}<span class="n">{preDecks.length}</span>{/if}
          </button>
        {/each}
      </div>
      <div class="ptools">
        {#if tab === "library"}
          {#if libDecks.length > FILTER_FROM}
            <input
              class="filter"
              type="search"
              placeholder="Filter your decks"
              aria-label="filter your decks"
              bind:value={libQuery}
            />
          {/if}
          <div class="seg" role="group" aria-label="sort your decks">
            <button
              type="button"
              aria-pressed={sort === "recent"}
              class:on={sort === "recent"}
              onclick={() => (sort = "recent")}>Recent</button
            >
            <button
              type="button"
              aria-pressed={sort === "name"}
              class:on={sort === "name"}
              onclick={() => (sort = "name")}>Name</button
            >
          </div>
        {:else if tab === "prebuilt" && preDecks.length > FILTER_FROM}
          <input
            class="filter"
            type="search"
            placeholder="Filter the pre-built decks"
            aria-label="filter the pre-built decks"
            bind:value={preQuery}
          />
        {/if}
      </div>
    </div>

    <div
      class="panel"
      role="tabpanel"
      id={`${uid}-panel`}
      aria-labelledby={`${uid}-tab-${tab}`}
      tabindex="-1"
    >
      {#if tab === "library"}
        <DeckGrid
          tiles={libTiles}
          name="your-deck"
          label={L.yourDecks}
          bind:selected={libSel}
          query={libQuery}
          onpick={() => (libTouched = true)}
        />
        {#if libCurrent}
          <div class="detail">
            <p class="dline dtext">
              <b>{libCurrent.name}</b
              >{#each libraryDetail(libCurrent) as part, i (i)}{" · "}{part}{/each}{#if libCurrent.source_url}{" · "}from
                <a href={libCurrent.source_url} target="_blank" rel="noopener noreferrer"
                  >{sourceHost(libCurrent.source_url)}</a
                >{/if}
            </p>
            <button
              type="button"
              class="btn ghost sm"
              data-testid="your-deck-sideboard"
              aria-expanded={sideboardOpen}
              onclick={() => {
                sideboardFor = libSel;
                sideboardOpen = !sideboardOpen;
              }}><Icon name="list" size={13} /> Sideboard</button
            >
          </div>
          {#if sideboardOpen}
            {#key libCurrent.id}
              <SideboardList deck={libCurrent} onclose={() => (sideboardOpen = false)} />
            {/key}
          {/if}
        {/if}
      {:else if tab === "prebuilt"}
        <DeckGrid
          tiles={preTiles}
          name="prebuilt-deck"
          label={L.prebuiltDecks}
          bind:selected={preSel}
          query={preQuery}
          onpick={() => (preTouched = true)}
        />
        {#if preCurrent}
          {@const cov = summariseCoverage(preCurrent.coverage)}
          {@const imperfect = preCurrent.coverage.imperfect ?? []}
          <div class="detail">
            <div class="dtext">
              <p class="dline">
                <b>{preCurrent.name}</b
                >{#if preCurrent.commander && !sameName(preCurrent.name, preCurrent.commander)}{" · "}{preCurrent.commander}{/if}{" · "}{cov.headline}{#if cov.tone !== "full" && cov.detail}{" · "}{cov.detail}{/if}
              </p>
              {#if preCurrent.summary}<p class="summary">{preCurrent.summary}</p>{/if}
            </div>
          </div>
          {#if imperfect.length > 0}
            <!-- Same register as the unimplemented list below, and for
                 the same reason: this deck is fine to seat. Collapsed,
                 since the count sets the expectation and the names are
                 for whoever wants to check a card. -->
            <details class="caveats">
              <summary>
                Which {imperfect.length}
                {imperfect.length === 1 ? "card" : "cards"}, and what changes?
              </summary>
              <ul>
                {#each imperfect as card (card.name)}
                  <li><b>{card.name}</b> <span>{imperfectCardLine(card)}</span></li>
                {/each}
              </ul>
            </details>
          {/if}
          {#if shopLinks && preCurrent.cards && preCurrent.cards.length > 0}
            <div class="paper">
              <ShopLinks list={preCurrent.cards} label="Play this deck in paper" />
            </div>
          {/if}
        {/if}
      {:else}
        <p class="note">
          Every deck is validated against Commander rules — 100-card singleton, color identity,
          format legality — whichever way it arrives.
        </p>
        <p class="note">
          An Archidekt deck link, or a plain-text decklist. On Moxfield, open the deck, then Export
          and Copy plain text, and paste it here.
        </p>
        {#if !signedIn}
          <p class="note keep">Sign in with Discord to keep the decks you paste.</p>
        {/if}
        <textarea
          rows="7"
          placeholder={"https://archidekt.com/decks/123456\n\n— or paste the list —\n\nCommander:\n1 Atraxa, Praetors' Voice\n\nMainboard:\n1 Sol Ring\n..."}
          bind:value={source}
          aria-label={L.decklistSource}
        ></textarea>
      {/if}
    </div>

    {#if actions === "inline" && commitState}
      <div class="row-actions">
        <button
          type="button"
          class="primary"
          disabled={commitState.disabled}
          title={commitState.title}
          onclick={() => void commit()}>{commitState.label}</button
        >
      </div>
    {/if}

    {#if successMessage}
      <p class="success" role="status"><Icon name="check" size={13} /> {successMessage}</p>
    {/if}
    {#if errorMessage}
      <pre class="deck-error">{errorMessage}</pre>
    {/if}
    {#if libraryNote}
      <p class="library-note" role="status">
        {libraryNote} <a href="#/decks">Open your decks</a>
      </p>
    {/if}
    {#if violations.length || warnings.length}
      <ul class="viol">
        {#each violations as v (v.code + (v.card ?? "") + v.message)}
          <li class="vi"><span class="code">{v.code}</span>{violationLabel(v)}</li>
        {/each}
        {#each warnings as v (v.code + (v.card ?? "") + v.message)}
          <li class="vi warn"><span class="code">{v.code}</span>{violationLabel(v)}</li>
        {/each}
      </ul>
    {/if}
    {#if unimplemented.length}
      <!-- Not a violation and not a warning: the deck is legal and
           seated. This is the one honest sentence about what the engine
           will and won't do for it, said before anyone has cast
           anything. -->
      <details class="unimpl">
        <summary>
          {unimplemented.length} of these cards {unimplemented.length === 1 ? "has" : "have"} rules the
          engine doesn't implement yet — {unimplemented.length === 1 ? "it" : "they"} behave as manual
          sandbox cards
        </summary>
        <ul>
          {#each unimplemented as name (name)}
            <li>{name}</li>
          {/each}
        </ul>
      </details>
    {/if}
  {/if}
</div>

<style>
  .deck-picker {
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-width: 0;
  }
  .note {
    margin: 0;
    font-size: 12.5px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .note.loading {
    padding: 12px 0 0;
  }
  .note.keep {
    color: var(--fg);
  }

  /* The tab bar runs edge to edge of whatever holds the picker: the
     holder sets --picker-pad to its own side padding. */
  .pbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 6px 12px;
    margin: 0 calc(-1 * var(--picker-pad, 0px));
    padding: 0 var(--picker-pad, 0px);
    border-bottom: 1px solid var(--border);
  }
  .tabs {
    display: flex;
    gap: 2px;
  }
  .tabs:focus {
    outline: none;
  }
  .tab {
    position: relative;
    display: inline-flex;
    align-items: center;
    gap: 7px;
    height: auto;
    padding: 12px 10px 11px;
    border: 0;
    border-radius: 0;
    background: none;
    box-shadow: none;
    color: var(--fg-muted);
    font: 600 13px/1 var(--font-ui);
    cursor: pointer;
  }
  .tab:hover {
    color: var(--fg);
    background: none;
  }
  .tab:active,
  .seg button:active {
    transform: none;
  }
  .tab[aria-selected="true"] {
    color: var(--fg);
  }
  .tab[aria-selected="true"]::after {
    content: "";
    position: absolute;
    left: 8px;
    right: 8px;
    bottom: -1px;
    height: 2px;
    border-radius: 2px;
    background: var(--accent);
  }
  .tab:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: -2px;
  }
  .n {
    font: 600 10.5px/1 var(--font-mono);
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .ptools {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 0;
  }
  .filter {
    width: 160px;
    height: 28px;
    margin: 0;
    padding: 0 10px;
    box-sizing: border-box;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    background: var(--surface-sunken);
    color: var(--fg);
    font: inherit;
    font-size: 12.5px;
  }
  .filter:focus {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-soft);
  }
  .filter::placeholder {
    color: var(--fg-dim);
  }
  .seg {
    display: inline-flex;
    overflow: hidden;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
  }
  .seg button {
    height: 26px;
    padding: 0 9px;
    border: 0;
    border-radius: 0;
    background: transparent;
    color: var(--fg-muted);
    font: 600 11.5px/1 var(--font-ui);
    cursor: pointer;
  }
  .seg button:hover {
    color: var(--fg);
  }
  .seg button.on {
    background: var(--surface-raised);
    color: var(--fg);
  }

  .panel {
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-width: 0;
  }
  .panel:focus {
    outline: none;
  }

  /* The detail line: the selected deck's longer story. */
  .detail {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 8px 14px;
    padding-top: 12px;
    border-top: 1px solid var(--border);
  }
  .dtext {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
    flex: 1 1 20rem;
  }
  .dline {
    min-width: 0;
    margin: 0;
    font-size: 12.5px;
    line-height: 1.45;
    color: var(--fg-muted);
  }
  .dline b {
    color: var(--fg);
    font-weight: 700;
  }
  .dline a {
    color: var(--fg-muted);
  }
  .dline a:hover {
    color: var(--accent-strong);
  }
  .summary {
    margin: 0;
    font-size: 12px;
    line-height: 1.45;
    color: var(--fg-muted);
  }
  .detail .btn {
    gap: 6px;
  }
  .caveats {
    font-size: 12px;
    color: var(--fg-muted);
  }
  .caveats summary {
    cursor: pointer;
  }
  .caveats ul {
    list-style: none;
    margin: 8px 0 0;
    padding: 0;
    max-height: 200px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .caveats li {
    display: flex;
    flex-direction: column;
    gap: 1px;
    line-height: 1.4;
  }
  .caveats b {
    color: var(--fg);
    font-weight: 600;
  }

  textarea {
    width: 100%;
    min-height: 128px;
    margin: 0;
    padding: 10px 12px;
    box-sizing: border-box;
    resize: vertical;
    font-family: var(--font-mono);
    font-size: 12px;
    line-height: 1.5;
  }

  .row-actions {
    display: flex;
    justify-content: flex-end;
  }
  /* A short fixed label, but never let it widen the button past the
     dialog. */
  .row-actions button {
    max-width: min(100%, 22rem);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* --- what seating said --- */
  .success {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin: 0;
    color: var(--mint);
    font-size: 12.5px;
  }
  .library-note {
    margin: 0;
    padding: 8px 10px;
    border-radius: var(--radius);
    border: 1px solid var(--border-strong);
    font-size: 12.5px;
    color: var(--fg);
  }
  .deck-error {
    white-space: pre-wrap;
    margin: 0;
    padding: 8px 10px;
    border-radius: var(--radius);
    background: color-mix(in srgb, var(--danger) 8%, transparent);
    border: 1px solid color-mix(in srgb, var(--danger) 30%, transparent);
    color: var(--fg);
    font-family: var(--font-ui);
    font-size: 12.5px;
  }
  .viol {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .vi {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    border-radius: var(--radius);
    background: color-mix(in srgb, var(--danger) 8%, transparent);
    border: 1px solid color-mix(in srgb, var(--danger) 30%, transparent);
    font-size: 12.5px;
    color: var(--fg);
  }
  .vi .code {
    flex: 0 0 auto;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.08em;
    color: var(--danger);
  }
  .vi.warn {
    background: var(--accent-soft);
    border-color: color-mix(in srgb, var(--accent) 40%, transparent);
  }
  .vi.warn .code {
    color: var(--accent-strong);
  }
  /* Muted on purpose. Red is for a deck that was rejected and the accent
     for one that lost something; this deck is fine, and the note is
     information rather than a problem. */
  .unimpl {
    padding: 8px 10px;
    border-radius: var(--radius);
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    font-size: 12.5px;
    color: var(--fg-muted);
  }
  .unimpl summary {
    cursor: pointer;
    line-height: 1.45;
  }
  .unimpl ul {
    list-style: none;
    margin: 8px 0 0;
    padding: 0;
    max-height: 180px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 2px;
    font-family: var(--font-mono);
    font-size: 11.5px;
    color: var(--fg-muted);
  }
</style>
