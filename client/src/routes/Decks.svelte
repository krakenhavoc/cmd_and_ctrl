<script lang="ts">
  // Decks — the one decks page (ADR 0112 §3). It fuses ADR 0095's
  // public deck check and ADR 0110 §6's saved library, from top to
  // bottom:
  //
  //   1. Check a deck, by link or pasted list (POST /deck-coverage).
  //   2. The report: "N of M cards play as printed", then the five
  //      buckets.
  //   3. Under it, two separate buttons: "Request missing cards" (POST
  //      /deck-requests) and "Save to my decks" (POST /me/decks) with a
  //      name field. A check writes nothing; saving is always the
  //      explicit button (owner answer 2), and neither button does the
  //      other's job (§3 item 6).
  //   4. Your decks: the library, with rename, delete, the report and a
  //      request button on every deck, pasted ones included (§3 item 5).
  //   5. Pre-built decks: read-only, never copied into the library
  //      (owner answer 4).
  //
  // Public. #/deck-check is a permanent alias of this route, because
  // the Discord bot's /c2-deck-check replies link there. Who sees which
  // control is decksAccess (lib/decksPage.ts, §3 item 3). A signed-out
  // visitor who follows "Sign in with Discord" comes back here, to the
  // same deck (§3 item 7, rememberAfterSignIn).

  import { onMount } from "svelte";
  import { route } from "../lib/router";
  import { LobbyApiError, session } from "../lib/session";
  import {
    deleteMyDeck,
    discordLoginHref,
    fetchMyDeckCoverage,
    fetchMyDecks,
    fetchPrebuiltDecks,
    renameMyDeck,
    saveMyDeck,
  } from "../lib/api";
  import { catalogSearchHash } from "../lib/roadmap";
  import Icon from "../lib/components/Icon.svelte";
  import SiteHeader from "../lib/components/SiteHeader.svelte";
  import {
    checkDeck,
    requestDeck,
    groupCardsByBucket,
    canRequestCards,
    deckRequestOutcomeMessage,
    shouldOfferPaste,
    BUCKET_ORDER,
    BUCKET_LABELS,
    BUCKET_BLURBS,
    DeckCheckError,
    PASTE_LIST_HINT,
    type CoverageReport,
    type CoverageBucket,
    type DeckCheckRequest,
    type DeckRequestResponse,
    type DeckCoverageViolation,
  } from "../lib/deckcheck";
  import {
    decksAccess,
    defaultSaveName,
    libraryDeckRequestable,
    rememberAfterSignIn,
    reportAsPrinted,
    reportDeckSize,
    reportNotFound,
    saveButtonLabel,
    savedMessage,
    takePendingDeckText,
  } from "../lib/decksPage";
  import {
    coverageDetail,
    coverageLine,
    deckSubtitle,
    sourceHost,
    type MyDeckInfo,
  } from "../lib/myDecks";
  import {
    deckSubtitle as prebuiltSubtitle,
    summariseCoverage,
    type PrebuiltDeck,
  } from "../lib/prebuiltDecks";

  type Tab = "link" | "paste";

  const access = $derived(decksAccess($session));
  const isSignedIn = $derived($session !== null);
  // The guest's own table, where "Link Discord" is (ADR 0051).
  const tableHref = $derived($session?.gameID ? `#/games/${$session.gameID}` : "");

  function message(err: unknown, fallback: string): string {
    return err instanceof Error && err.message ? err.message : fallback;
  }

  // The caller's library (section 4), declared first because the save
  // button reads it: a name already in it reads "Replace ‹name›".
  let decks = $state<MyDeckInfo[]>([]);
  let libLoading = $state(false);
  let libLoaded = $state(false);
  let libError = $state("");

  // --- 1. the check ---------------------------------------------------

  let tab = $state<Tab>("link");
  let urlInput = $state("");
  let textInput = $state("");

  let loading = $state(false);
  let errorMessage = $state("");
  let errorViolations = $state<DeckCoverageViolation[]>([]);
  // offerPaste: the error says to paste the list instead (a Moxfield link).
  let offerPaste = $state(false);
  let report = $state<CoverageReport | null>(null);
  // checked is what `report` was built from: the request, the save and
  // the sign-in return route all send the same.
  let checked = $state<DeckCheckRequest | null>(null);

  const groups = $derived(report ? groupCardsByBucket(report.cards) : null);
  const showRequest = $derived(canRequestCards(report) && access.request !== "hidden");
  const showSave = $derived(report !== null && access.save !== "hidden");

  const totalCards = $derived.by(() => {
    if (!report) return 0;
    return reportDeckSize(report);
  });

  function bucketPct(b: CoverageBucket): number {
    if (!report || totalCards === 0) return 0;
    return (report.copies[b] / totalCards) * 100;
  }

  function bucketCount(b: CoverageBucket): number {
    return report ? report.copies[b] : 0;
  }

  function violationLabel(v: DeckCoverageViolation): string {
    return v.card ? `${v.card}: ${v.message}` : v.message;
  }

  // The link the report names, canonical, or the list as pasted: what a
  // request or a save sends.
  function checkedDeck(): DeckCheckRequest | null {
    if (!report || !checked) return null;
    return report.source_url ? { url: report.source_url } : checked;
  }

  async function runCheck(): Promise<void> {
    const isLink = tab === "link";
    const value = (isLink ? urlInput : textInput).trim();
    if (!value) {
      errorMessage = isLink ? "Paste a deck link first." : "Paste a decklist first.";
      return;
    }
    loading = true;
    errorMessage = "";
    errorViolations = [];
    offerPaste = false;
    report = null;
    checked = null;
    requestResult = null;
    requestError = "";
    saveError = "";
    saveNotice = "";
    const req: DeckCheckRequest = isLink ? { url: value } : { text: value };
    try {
      const r = await checkDeck(req);
      report = r;
      checked = req;
      saveName = defaultSaveName(r);
    } catch (err) {
      if (err instanceof DeckCheckError) {
        errorMessage = err.message;
        offerPaste = shouldOfferPaste(err);
        // The hint's sentence says it all; the fetch violation under it
        // would only repeat the link.
        if (!offerPaste && err.violations && err.violations.length > 0) {
          errorViolations = err.violations;
        }
      } else {
        errorMessage = "Couldn't check that deck.";
      }
    } finally {
      loading = false;
    }
  }

  function submit(e: SubmitEvent): void {
    e.preventDefault();
    void runCheck();
  }

  function switchToPaste(): void {
    tab = "paste";
    offerPaste = false;
  }

  // Store this page, and the pasted list, so the Discord round trip
  // comes back to the same deck (§3 item 7).
  function beforeSignIn(): void {
    rememberAfterSignIn(checked);
  }

  // #/decks?url=<link> (and #/deck-check?url=, the bot's link) pre-fills
  // the link field and runs the check on load. Reread on every route
  // change, like Catalog.svelte's ?q=, so a second link lands on its
  // own deck.
  $effect(() => {
    const r = $route;
    if (r.name !== "decks" || !r.url) return;
    tab = "link";
    urlInput = r.url;
    void runCheck();
  });

  // --- 3. request and save the checked deck ----------------------------

  let requestBusy = $state(false);
  let requestError = $state("");
  let requestResult = $state<DeckRequestResponse | null>(null);

  async function fileRequest(): Promise<void> {
    const req = checkedDeck();
    if (!req) return;
    requestBusy = true;
    requestError = "";
    requestResult = null;
    try {
      requestResult = await requestDeck(req);
    } catch (err) {
      requestError = message(err, "Couldn't file that request.");
    } finally {
      requestBusy = false;
    }
  }

  let saveName = $state("");
  let saveBusy = $state(false);
  let saveError = $state("");
  let saveNotice = $state("");
  // The save hit the Moxfield block: paste the list, as the check would.
  let saveOfferPaste = $state(false);

  const saveLabel = $derived(saveButtonLabel(saveName, decks));

  async function saveChecked(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    const req = checkedDeck();
    if (!req || saveBusy) return;
    saveBusy = true;
    saveError = "";
    saveNotice = "";
    saveOfferPaste = false;
    try {
      const res = await saveMyDeck(req, saveName);
      decks = res.replaced
        ? decks.map((d) => (d.id === res.deck.id ? res.deck : d))
        : [res.deck, ...decks.filter((d) => d.id !== res.deck.id)];
      saveNotice = savedMessage(res);
    } catch (err) {
      saveError = message(err, "Couldn't save that deck.");
      if (err instanceof LobbyApiError) {
        const body = err.body as { hint?: string } | undefined;
        saveOfferPaste = body?.hint === PASTE_LIST_HINT;
      }
    } finally {
      saveBusy = false;
    }
  }

  // --- 4. your decks -------------------------------------------------

  let renaming = $state("");
  let renameValue = $state("");
  let renameError = $state("");
  let confirmDelete = $state("");
  let busy = $state("");

  let reportFor = $state("");
  let libReport = $state<CoverageReport | null>(null);
  let libReportError = $state("");
  const libGrouped = $derived(libReport ? groupCardsByBucket(libReport.cards) : null);

  // Per-deck "Request missing cards": which deck is asking, and each
  // deck's last answer.
  let requesting = $state("");
  let deckRequests = $state<Record<string, DeckRequestResponse>>({});
  let deckRequestErrors = $state<Record<string, string>>({});

  async function loadLibrary(): Promise<void> {
    libLoading = true;
    libError = "";
    try {
      decks = (await fetchMyDecks()).decks ?? [];
      libLoaded = true;
    } catch (err) {
      libError = message(err, "couldn't load your decks");
    } finally {
      libLoading = false;
    }
  }

  // The library loads once there is a person to load it for, which may
  // be after mount: the Discord round trip installs the session and then
  // routes here.
  $effect(() => {
    if (access.library === "allowed" && !libLoaded && !libLoading && !libError) {
      void loadLibrary();
    }
  });

  function startRename(d: MyDeckInfo): void {
    renaming = d.id;
    renameValue = d.name;
    renameError = "";
    confirmDelete = "";
  }

  async function saveRename(d: MyDeckInfo): Promise<void> {
    const name = renameValue.trim();
    if (!name || busy) return;
    if (name === d.name) {
      renaming = "";
      return;
    }
    busy = d.id;
    renameError = "";
    try {
      const updated = await renameMyDeck(d.id, name);
      decks = decks.map((x) => (x.id === d.id ? { ...x, name: updated.name } : x));
      renaming = "";
    } catch (err) {
      renameError = message(err, "couldn't rename that deck");
    } finally {
      busy = "";
    }
  }

  async function remove(d: MyDeckInfo): Promise<void> {
    if (busy) return;
    busy = d.id;
    libError = "";
    try {
      await deleteMyDeck(d.id);
      decks = decks.filter((x) => x.id !== d.id);
      if (reportFor === d.id) closeReport();
      confirmDelete = "";
    } catch (err) {
      libError = message(err, "couldn't delete that deck");
    } finally {
      busy = "";
    }
  }

  function closeReport(): void {
    reportFor = "";
    libReport = null;
    libReportError = "";
  }

  async function toggleReport(d: MyDeckInfo): Promise<void> {
    if (reportFor === d.id) {
      closeReport();
      return;
    }
    reportFor = d.id;
    libReport = null;
    libReportError = "";
    try {
      const r = await fetchMyDeckCoverage(d.id);
      if (reportFor === d.id) libReport = r;
    } catch (err) {
      if (reportFor === d.id) libReportError = message(err, "couldn't build the report");
    }
  }

  async function requestSaved(d: MyDeckInfo): Promise<void> {
    if (requesting) return;
    requesting = d.id;
    const results = { ...deckRequests };
    const errors = { ...deckRequestErrors };
    delete results[d.id];
    delete errors[d.id];
    deckRequests = results;
    deckRequestErrors = errors;
    try {
      const res = await requestDeck({ deck_id: d.id });
      deckRequests = { ...deckRequests, [d.id]: res };
    } catch (err) {
      deckRequestErrors = {
        ...deckRequestErrors,
        [d.id]: message(err, "Couldn't file that request."),
      };
    } finally {
      requesting = "";
    }
  }

  // --- 5. pre-built decks ------------------------------------------------

  // GET /decks is session-gated on the server, so the section shows for
  // any session and a signed-out visitor does not ask for it.
  let prebuilt = $state<PrebuiltDeck[]>([]);
  let prebuiltLoaded = $state(false);
  let prebuiltError = $state("");

  $effect(() => {
    if (!isSignedIn || prebuiltLoaded) return;
    prebuiltLoaded = true;
    void fetchPrebuiltDecks()
      .then((r) => {
        prebuilt = r.decks ?? [];
      })
      .catch((err: unknown) => {
        prebuiltError = message(err, "couldn't load the pre-built decks");
      });
  });

  // A pasted list carried across the Discord round trip (§3 item 7):
  // put it back in the paste tab and check it again, so the visitor is
  // looking at the same report they left.
  onMount(() => {
    const text = takePendingDeckText();
    if (!text) return;
    tab = "paste";
    textInput = text;
    void runCheck();
  });
</script>

<!-- A card name: a catalogue link for a signed-in viewer, plain text
     otherwise — same rule and same reason as Roadmap.svelte's
     `cardName` snippet (the catalogue needs a session). -->
{#snippet cardName(name: string)}
  {#if isSignedIn}
    <a class="cname" href={catalogSearchHash(name)}>{name}</a>
  {:else}
    <span class="cname">{name}</span>
  {/if}
{/snippet}

<!-- What a visitor who may not do something sees in its place (§3 item
     3): a Discord sign-in that returns here, or, for a guest seat, the
     way to an account from their table. -->
{#snippet gate(action: "signIn" | "linkDiscord", what: string)}
  {#if action === "signIn"}
    <a class="ghost-link" href={discordLoginHref()} onclick={beforeSignIn}>
      <Icon name="chevronRight" size={12} /> Sign in with Discord to {what}
    </a>
  {:else}
    <span class="gate-note">
      Link Discord from your table's menu to {what}.
      {#if tableHref}<a class="ghost-link inline" href={tableHref}>Back to your table</a>{/if}
    </span>
  {/if}
{/snippet}

<SiteHeader />

<section class="entry">
  <div class="stack">
    <div class="head">
      <p class="eyebrow">Check, request, keep</p>
      <h1>Decks</h1>
      <p class="lede">
        Paste an Archidekt link or a plain decklist to see how much of it plays as printed. Request
        what's missing, and save the decks you play. On Moxfield, open the deck → Export → Copy
        plain text, and paste that.
      </p>
    </div>

    <!-- 1. Check a deck -->
    <form class="input-card" onsubmit={submit} aria-labelledby="check-h">
      <h2 id="check-h" class="sec-title">Check a deck</h2>
      <div class="tabs" role="tablist" aria-label="how to check a deck">
        <button
          type="button"
          role="tab"
          class="tab"
          class:on={tab === "link"}
          aria-selected={tab === "link"}
          onclick={() => (tab = "link")}
        >
          Link
        </button>
        <button
          type="button"
          role="tab"
          class="tab"
          class:on={tab === "paste"}
          aria-selected={tab === "paste"}
          onclick={() => (tab = "paste")}
        >
          Paste a list
        </button>
      </div>

      {#if tab === "link"}
        <input
          type="text"
          class="url-field"
          placeholder="https://archidekt.com/decks/123456"
          aria-label="deck link"
          bind:value={urlInput}
        />
      {:else}
        <textarea
          rows="7"
          class="paste-field"
          placeholder={"1 Atraxa, Praetors' Voice *CMDR*\n1 Sol Ring\n1 Doubling Season\n..."}
          aria-label="decklist"
          bind:value={textInput}
        ></textarea>
      {/if}

      <div class="row-actions">
        <button type="submit" class="primary lg" disabled={loading}>
          {loading ? "Checking…" : "Check this deck"}
        </button>
      </div>
    </form>

    {#if errorMessage}
      <p class="notice err" role="alert">
        <Icon name="x" size={14} />
        <span>{errorMessage}</span>
      </p>
      {#if offerPaste}
        <div class="row-actions start">
          <button type="button" class="ghost" onclick={switchToPaste}>
            Paste the list instead
          </button>
        </div>
      {/if}
      {#if errorViolations.length > 0}
        <ul class="violations">
          {#each errorViolations as v (v.code + (v.card ?? "") + v.message)}
            <li>{violationLabel(v)}</li>
          {/each}
        </ul>
      {/if}
    {/if}

    <!-- 2. The report, and 3. its actions -->
    {#if report && groups}
      <section class="report" aria-live="polite" aria-label="deck report">
        <div class="deck-id">
          <h2 class="deck-name">{report.deck_name || "This deck"}</h2>
          {#if report.commanders.length > 0}
            <p class="commanders">
              {report.commanders.length === 1 ? "Commander" : "Commanders"}: {report.commanders.join(
                ", ",
              )}
            </p>
          {/if}
          {#if report.source_url}
            <a
              class="source-link"
              href={report.source_url}
              target="_blank"
              rel="noreferrer noopener"
            >
              {report.source_url}
              <Icon name="link" size={11} />
            </a>
          {/if}
        </div>

        {#if reportAsPrinted(report)}
          <p class="as-printed">{reportAsPrinted(report)}</p>
        {/if}

        <div class="bar" role="img" aria-label="coverage by bucket">
          {#each BUCKET_ORDER as b (b)}
            {#if bucketCount(b) > 0}
              <span
                class="seg {b}"
                style="width: {bucketPct(b)}%"
                title="{BUCKET_LABELS[b]}: {bucketCount(b)}"
              ></span>
            {/if}
          {/each}
          {#if report.unknown_copies > 0}
            <span
              class="seg unknown"
              style="width: {(report.unknown_copies / totalCards) * 100}%"
              title="Not found: {report.unknown_copies}"
            ></span>
          {/if}
        </div>

        <ul class="legend">
          {#each BUCKET_ORDER as b (b)}
            <li class="lg-item {b}">
              <span class="dot {b}" aria-hidden="true"></span>
              <span class="lg-label">{BUCKET_LABELS[b]}</span>
              <span class="lg-n">{bucketCount(b)}</span>
            </li>
          {/each}
          {#if reportNotFound(report)}
            <li class="lg-item unknown">
              <span class="dot unknown" aria-hidden="true"></span>
              <span class="lg-label">Not found</span>
              <span class="lg-n">{report.unknown_copies}</span>
            </li>
          {/if}
        </ul>

        {#if report.violations.length > 0}
          <p class="notice info">
            <Icon name="flag" size={13} />
            <span
              >{report.violations.length === 1 ? "One thing to know" : "A couple of things to know"}
              — this only affects deck legality, not the report below.</span
            >
          </p>
          <ul class="violations info">
            {#each report.violations as v (v.code + (v.card ?? "") + v.message)}
              <li>{violationLabel(v)}</li>
            {/each}
          </ul>
        {/if}

        {#if showRequest || showSave}
          <div class="actions-card">
            {#if showRequest}
              <div class="action">
                {#if access.request === "allowed"}
                  <button
                    type="button"
                    class="primary"
                    onclick={fileRequest}
                    disabled={requestBusy}
                  >
                    {requestBusy ? "Requesting…" : "Request missing cards"}
                  </button>
                  {#if requestResult}
                    <p
                      class="outcome"
                      class:warn={requestResult.status === "rate_limited"}
                      role="status"
                    >
                      {deckRequestOutcomeMessage(requestResult)}
                      {#if requestResult.issue_url}
                        <a href={requestResult.issue_url} target="_blank" rel="noreferrer noopener">
                          #{requestResult.issue_number}
                          <Icon name="link" size={11} />
                        </a>
                      {/if}
                    </p>
                  {/if}
                  {#if requestError}
                    <p class="notice err" role="alert">
                      <Icon name="x" size={14} />
                      <span>{requestError}</span>
                    </p>
                  {/if}
                {:else if access.request !== "hidden"}
                  {@render gate(access.request, "request these cards")}
                {/if}
              </div>
            {/if}

            {#if showSave}
              <div class="action">
                {#if access.save === "allowed"}
                  <form class="save" onsubmit={saveChecked}>
                    <input
                      type="text"
                      bind:value={saveName}
                      maxlength="100"
                      placeholder="Deck name"
                      aria-label="name to save this deck as"
                    />
                    <button type="submit" disabled={saveBusy}>
                      {saveBusy ? "Saving…" : saveLabel}
                    </button>
                  </form>
                  {#if saveNotice}
                    <p class="outcome" role="status">{saveNotice}</p>
                  {/if}
                  {#if saveError}
                    <p class="notice err" role="alert">
                      <Icon name="x" size={14} />
                      <span>{saveError}</span>
                    </p>
                    {#if saveOfferPaste}
                      <div class="row-actions start">
                        <button type="button" class="ghost" onclick={switchToPaste}>
                          Paste the list instead
                        </button>
                      </div>
                    {/if}
                  {/if}
                {:else if access.save !== "hidden"}
                  {@render gate(access.save, "save this deck")}
                {/if}
              </div>
            {/if}
          </div>
        {/if}

        {#each BUCKET_ORDER as b (b)}
          {#if groups[b].length > 0}
            <section class="bucket-sec {b}" aria-labelledby="bk-{b}">
              <div class="sec-head">
                <h3 id="bk-{b}">
                  <span class="dot {b}" aria-hidden="true"></span>
                  {BUCKET_LABELS[b]}
                  <span class="sec-n">{groups[b].length}</span>
                </h3>
                <p class="sec-note">{BUCKET_BLURBS[b]}</p>
              </div>
              <ul class="cards">
                {#each groups[b] as c (c.oracle_id)}
                  <li class="card-row">
                    <span class="card-name">
                      {@render cardName(c.name)}
                      {#if c.count > 1}<span class="count">×{c.count}</span>{/if}
                    </span>
                    {#if c.caveats && c.caveats.length > 0}
                      <ul class="caveats">
                        {#each c.caveats as cav (cav)}
                          <li>{cav}</li>
                        {/each}
                      </ul>
                    {/if}
                  </li>
                {/each}
              </ul>
            </section>
          {/if}
        {/each}

        {#if report.unknown.length > 0}
          <section class="bucket-sec unknown">
            <div class="sec-head">
              <h3>
                Not resolved
                <span class="sec-n">{report.unknown.length}</span>
              </h3>
              <p class="sec-note">
                The card index couldn't match these names — check the spelling, or they may not be a
                real Magic card.
              </p>
            </div>
            <ul class="unknown-list">
              {#each report.unknown as name (name)}
                <li>{name}</li>
              {/each}
            </ul>
          </section>
        {/if}
      </section>
    {/if}

    <!-- 4. Your decks -->
    {#if access.library !== "hidden"}
      <section class="panel" aria-labelledby="yours-h">
        <h2 id="yours-h" class="sec-title">Your decks</h2>
        {#if access.library === "signIn"}
          <p class="help">
            Saved decks are kept for your Discord account.
            <a class="ghost-link inline" href={discordLoginHref()} onclick={beforeSignIn}
              >Sign in with Discord to see them.</a
            >
          </p>
        {:else if access.library === "linkDiscord"}
          <p class="help">
            Saved decks are kept for a Discord account. Link Discord from your table's menu to keep
            decks here.
          </p>
        {:else if libLoading && decks.length === 0}
          <p class="help" aria-live="polite">Loading your decks…</p>
        {:else if decks.length === 0 && !libError}
          <p class="help">
            No saved decks yet. Decks you save here, or import at a table while signed in, are kept.
          </p>
        {:else}
          <ul class="decks" aria-label="your decks">
            {#each decks as d (d.id)}
              <li class="deck">
                <div class="line">
                  {#if renaming === d.id}
                    <form
                      class="rename"
                      onsubmit={(e) => {
                        e.preventDefault();
                        void saveRename(d);
                      }}
                    >
                      <input
                        type="text"
                        bind:value={renameValue}
                        maxlength="100"
                        aria-label={`new name for ${d.name}`}
                      />
                      <button
                        type="submit"
                        class="primary"
                        disabled={busy !== "" || !renameValue.trim()}
                      >
                        Save
                      </button>
                      <button type="button" onclick={() => (renaming = "")}>Cancel</button>
                    </form>
                  {:else}
                    <span class="name">{d.name}</span>
                  {/if}
                </div>
                {#if renaming === d.id && renameError}
                  <p class="notice err" role="alert">{renameError}</p>
                {/if}
                <div class="meta">{deckSubtitle(d)}</div>
                {#if coverageLine(d)}
                  <div class="cov">{coverageLine(d)}</div>
                  {#if coverageDetail(d)}
                    <div class="cov-detail">{coverageDetail(d)}</div>
                  {/if}
                {/if}
                {#if d.source_url}
                  <div class="meta">
                    From <a href={d.source_url} target="_blank" rel="noopener noreferrer"
                      >{sourceHost(d.source_url)}</a
                    >
                  </div>
                {/if}

                <div class="row-btns">
                  {#if d.coverage}
                    <button
                      type="button"
                      onclick={() => toggleReport(d)}
                      aria-expanded={reportFor === d.id}
                    >
                      {reportFor === d.id ? "Hide report" : "Coverage report"}
                    </button>
                  {/if}
                  {#if libraryDeckRequestable(d)}
                    <button
                      type="button"
                      onclick={() => requestSaved(d)}
                      disabled={requesting !== ""}
                      aria-label={`request missing cards for ${d.name}`}
                    >
                      {requesting === d.id ? "Requesting…" : "Request missing cards"}
                    </button>
                  {/if}
                  <button type="button" onclick={() => startRename(d)} disabled={busy !== ""}>
                    Rename
                  </button>
                  {#if confirmDelete === d.id}
                    <span class="confirm" role="alert">
                      Delete {d.name}?
                      <button
                        type="button"
                        class="danger"
                        disabled={busy !== ""}
                        onclick={() => remove(d)}
                      >
                        {busy === d.id ? "…" : "Delete"}
                      </button>
                      <button type="button" onclick={() => (confirmDelete = "")}>Keep</button>
                    </span>
                  {:else}
                    <button
                      type="button"
                      onclick={() => {
                        confirmDelete = d.id;
                        renaming = "";
                      }}
                      disabled={busy !== ""}
                      aria-label={`delete ${d.name}`}
                    >
                      Delete
                    </button>
                  {/if}
                </div>

                {#if deckRequests[d.id]}
                  {@const res = deckRequests[d.id]}
                  <p class="outcome" class:warn={res.status === "rate_limited"} role="status">
                    {deckRequestOutcomeMessage(res)}
                    {#if res.issue_url}
                      <a href={res.issue_url} target="_blank" rel="noreferrer noopener">
                        #{res.issue_number}
                        <Icon name="link" size={11} />
                      </a>
                    {/if}
                  </p>
                {/if}
                {#if deckRequestErrors[d.id]}
                  <p class="notice err" role="alert">
                    <Icon name="x" size={14} />
                    <span>{deckRequestErrors[d.id]}</span>
                  </p>
                {/if}

                {#if reportFor === d.id}
                  <div class="lib-report">
                    {#if libReportError}
                      <p class="notice err" role="alert">{libReportError}</p>
                    {:else if !libGrouped || !libReport}
                      <p class="help">Building the report…</p>
                    {:else}
                      {#each BUCKET_ORDER as b (b)}
                        {#if libGrouped[b].length > 0}
                          <details open={b === "manual" || b === "unreviewed"}>
                            <summary>{BUCKET_LABELS[b]} ({libGrouped[b].length})</summary>
                            <ul>
                              {#each libGrouped[b] as c (c.oracle_id || c.name)}
                                <li>
                                  {c.name}
                                  {#if c.caveats && c.caveats.length > 0}
                                    <span class="caveat">— {c.caveats.join(" ")}</span>
                                  {/if}
                                </li>
                              {/each}
                            </ul>
                          </details>
                        {/if}
                      {/each}
                      {#if libReport.unknown.length > 0}
                        <p class="help">
                          Not found in the card index: {libReport.unknown.join(", ")}
                        </p>
                      {/if}
                    {/if}
                  </div>
                {/if}
              </li>
            {/each}
          </ul>
        {/if}
        {#if libError}
          <p class="notice err" role="alert">
            <Icon name="x" size={14} />
            <span>{libError}</span>
          </p>
        {/if}
      </section>
    {/if}

    <!-- 5. Pre-built decks (owner answer 4): read-only, never copied. -->
    {#if isSignedIn && (prebuilt.length > 0 || prebuiltError)}
      <section class="panel" aria-labelledby="prebuilt-h">
        <h2 id="prebuilt-h" class="sec-title">Pre-built decks</h2>
        <p class="help">
          Ready-made decks you can pick at any table. They stay here, and are never copied into your
          decks.
        </p>
        {#if prebuiltError}
          <p class="notice err" role="alert">
            <Icon name="x" size={14} />
            <span>{prebuiltError}</span>
          </p>
        {:else}
          <ul class="decks prebuilt" aria-label="pre-built decks">
            {#each prebuilt as p (p.id)}
              {@const cov = summariseCoverage(p.coverage)}
              <li class="deck">
                <span class="name">{p.name}</span>
                {#if prebuiltSubtitle(p)}
                  <div class="meta">{prebuiltSubtitle(p)}</div>
                {/if}
                <div class="cov tone-{cov.tone}">{cov.headline}</div>
                {#if cov.detail}
                  <div class="cov-detail">{cov.detail}</div>
                {/if}
              </li>
            {/each}
          </ul>
        {/if}
      </section>
    {/if}
  </div>
</section>

<style>
  .entry {
    position: relative;
    min-height: calc(100vh - 3rem - 68px);
    display: flex;
    justify-content: center;
  }
  .stack {
    width: min(920px, 100%);
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 20px;
    margin: clamp(8px, 2vh, 24px) 0 48px;
  }

  .eyebrow {
    margin: 0 0 6px;
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--accent-strong);
  }
  h1 {
    margin: 0;
    font-family: var(--font-display);
    font-weight: 800;
    letter-spacing: -0.01em;
  }
  .lede {
    margin: 10px 0 0;
    max-width: 64ch;
    color: var(--fg-muted);
    font-size: 14px;
    line-height: 1.6;
  }
  .sec-title {
    margin: 0;
    font-family: var(--font-display);
    font-size: 16px;
    font-weight: 800;
    letter-spacing: -0.01em;
    color: var(--fg);
  }

  /* --- input card -------------------------------------------------- */
  .input-card,
  .panel {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px;
    border-radius: var(--radius-lg);
    background: var(--surface);
    border: 1px solid var(--border);
    min-width: 0;
  }
  .tabs {
    display: inline-flex;
    gap: 4px;
    align-self: flex-start;
  }
  .tab {
    height: 28px;
    padding: 0 12px;
    border-radius: 999px;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    color: var(--fg-muted);
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    cursor: pointer;
  }
  .tab:hover {
    background: var(--surface-hover);
    color: var(--fg);
  }
  .tab.on {
    background: var(--accent-soft);
    border-color: var(--accent);
    color: var(--accent-strong);
  }
  .url-field {
    width: 100%;
    box-sizing: border-box;
    font-family: var(--font-mono);
    font-size: 13px;
  }
  .paste-field {
    width: 100%;
    box-sizing: border-box;
    min-height: 128px;
    font-family: var(--font-mono);
    font-size: 12px;
    line-height: 1.5;
    resize: vertical;
  }
  .row-actions {
    display: flex;
    justify-content: flex-end;
  }
  .row-actions.start {
    justify-content: flex-start;
  }
  .lg {
    height: 38px;
    padding: 0 18px;
    font-size: 13px;
  }

  /* --- notices ------------------------------------------------------ */
  .notice {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    margin: 0;
    padding: 10px 12px;
    border-radius: var(--radius);
    background: var(--surface);
    border: 1px solid var(--border-strong);
    font-size: 13px;
    line-height: 1.5;
  }
  .notice.err {
    color: var(--danger);
  }
  .notice.info {
    color: var(--fg-muted);
  }
  .violations {
    list-style: none;
    margin: -4px 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .violations li {
    padding: 6px 10px;
    border-radius: var(--radius);
    background: color-mix(in srgb, var(--danger) 8%, transparent);
    border: 1px solid color-mix(in srgb, var(--danger) 30%, transparent);
    font-size: 12px;
    color: var(--fg-muted);
  }
  .violations.info li {
    background: var(--surface);
    border-color: var(--border);
  }

  /* --- report shell --------------------------------------------------- */
  .report {
    display: flex;
    flex-direction: column;
    gap: 16px;
    min-width: 0;
  }
  .deck-id {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }
  .deck-name {
    margin: 0;
    font-family: var(--font-display);
    font-size: 20px;
    font-weight: 800;
    letter-spacing: -0.01em;
    overflow-wrap: anywhere;
  }
  .commanders {
    margin: 0;
    color: var(--fg-muted);
    font-size: 13px;
  }
  .source-link {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    margin-top: 2px;
    color: var(--fg-dim);
    font-family: var(--font-mono);
    font-size: 11px;
    text-decoration: none;
    overflow-wrap: anywhere;
  }
  .source-link:hover {
    color: var(--accent-strong);
  }
  .as-printed {
    margin: 0;
    font-size: 15px;
    font-weight: 700;
    color: var(--mint);
  }

  /* --- proportion bar + legend --------------------------------------- */
  .bar {
    display: flex;
    height: 12px;
    width: 100%;
    border-radius: 999px;
    overflow: hidden;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
  }
  .seg {
    height: 100%;
  }
  .seg.manual,
  .dot.manual {
    background: var(--rose);
  }
  .seg.unreviewed,
  .dot.unreviewed {
    background: var(--fg-dim);
  }
  .seg.caveats,
  .dot.caveats {
    background: var(--accent);
  }
  .seg.automated,
  .dot.automated {
    background: var(--mint);
  }
  .seg.no_effect,
  .dot.no_effect {
    background: var(--border-strong);
  }
  .seg.unknown,
  .dot.unknown {
    background: repeating-linear-gradient(
      45deg,
      var(--rose),
      var(--rose) 2px,
      transparent 2px,
      transparent 4px
    );
  }
  .legend {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 8px;
  }
  .lg-item {
    display: flex;
    align-items: center;
    gap: 7px;
    padding: 8px 10px;
    border-radius: var(--radius);
    background: var(--surface);
    border: 1px solid var(--border);
    font-size: 12px;
  }
  .dot {
    flex: 0 0 auto;
    width: 9px;
    height: 9px;
    border-radius: 50%;
  }
  .lg-label {
    flex: 1;
    min-width: 0;
    color: var(--fg-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .lg-n {
    font-family: var(--font-mono);
    font-weight: 700;
    color: var(--fg);
  }

  /* --- request + save ------------------------------------------------- */
  .actions-card {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    gap: 12px 24px;
    padding: 14px 0 4px;
    border-top: 1px solid var(--border);
  }
  .action {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
    flex: 1 1 260px;
  }
  .action > button {
    align-self: flex-start;
  }
  .save {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .save input {
    flex: 1 1 180px;
    min-width: 0;
  }
  .outcome {
    margin: 0;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    color: var(--mint);
    font-size: 12.5px;
  }
  .outcome.warn {
    color: var(--accent-strong);
  }
  .outcome a {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: var(--accent-strong);
    text-decoration: none;
    font-family: var(--font-mono);
  }
  .outcome a:hover {
    text-decoration: underline;
  }
  .gate-note {
    font-size: 13px;
    color: var(--fg-muted);
    line-height: 1.5;
  }
  .ghost-link {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: var(--accent-strong);
    text-decoration: none;
    font-size: 13px;
    font-weight: 600;
  }
  .ghost-link.inline {
    display: inline;
  }
  .ghost-link:hover {
    text-decoration: underline;
  }

  /* --- bucket sections --------------------------------------------- */
  .sec-head {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 4px 14px;
    margin-bottom: 10px;
  }
  .sec-head h3 {
    margin: 0;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-family: var(--font-display);
    font-size: 15px;
    font-weight: 700;
    color: var(--fg);
  }
  .sec-note {
    margin: 0;
    color: var(--fg-dim);
    font-size: 12px;
  }
  .sec-n {
    font-family: var(--font-mono);
    font-size: 11.5px;
    font-weight: 600;
    color: var(--fg-dim);
  }
  .cards {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 6px 14px;
  }
  .card-row {
    padding: 6px 0;
    border-bottom: 1px solid var(--border);
    min-width: 0;
  }
  .card-name {
    display: flex;
    align-items: baseline;
    gap: 6px;
    font-size: 13px;
  }
  .cname {
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  a.cname {
    text-decoration: underline;
    text-decoration-color: color-mix(in srgb, var(--accent) 50%, transparent);
    text-underline-offset: 2px;
  }
  a.cname:hover {
    color: var(--accent-strong);
  }
  .count {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--fg-dim);
  }
  .caveats {
    list-style: none;
    margin: 4px 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }
  .caveats li {
    font-size: 11.5px;
    line-height: 1.45;
    color: var(--fg-muted);
    border-left: 2px solid color-mix(in srgb, var(--accent) 50%, transparent);
    padding-left: 8px;
  }
  .unknown-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .unknown-list li {
    padding: 3px 9px;
    border-radius: 999px;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    font-size: 12px;
    color: var(--fg-muted);
    overflow-wrap: anywhere;
  }

  /* --- the library and the pre-built list --------------------------- */
  .help {
    margin: 0;
    font-size: 12.5px;
    color: var(--fg-muted);
    line-height: 1.5;
  }
  .decks {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
  }
  .deck {
    padding: 14px 0;
    border-top: 1px solid var(--border);
    min-width: 0;
  }
  .deck:first-child {
    border-top: none;
    padding-top: 0;
  }
  .deck:last-child {
    padding-bottom: 0;
  }
  .name {
    font-weight: 700;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .meta {
    margin-top: 3px;
    font-size: 12px;
    color: var(--fg-muted);
  }
  .cov {
    margin-top: 4px;
    font-size: 12.5px;
    font-weight: 600;
    color: var(--mint);
  }
  .cov.tone-caveats {
    color: var(--accent-strong);
  }
  .cov.tone-gap {
    color: var(--danger);
  }
  .cov-detail {
    margin-top: 2px;
    font-size: 11.5px;
    color: var(--fg-dim);
  }
  .row-btns {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    margin-top: 10px;
  }
  .deck .outcome,
  .deck .notice {
    margin-top: 8px;
  }
  .rename {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    width: 100%;
  }
  .rename input {
    flex: 1 1 200px;
    min-width: 0;
  }
  .confirm {
    display: inline-flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    font-size: 12.5px;
    color: var(--danger);
  }
  .danger {
    color: var(--danger);
    border-color: var(--danger);
  }
  .lib-report {
    margin-top: 12px;
    padding: 10px 12px;
    border-radius: 8px;
    border: 1px solid var(--border);
    background: var(--surface-sunken);
    font-size: 12.5px;
  }
  .lib-report details + details {
    margin-top: 8px;
  }
  .lib-report summary {
    cursor: pointer;
    font-weight: 600;
    color: var(--fg);
  }
  .lib-report ul {
    margin: 6px 0 0;
    padding-left: 18px;
    color: var(--fg-muted);
  }
  .caveat {
    color: var(--fg-dim);
  }

  @media (max-width: 480px) {
    .legend {
      grid-template-columns: minmax(0, 1fr);
    }
    .cards {
      grid-template-columns: minmax(0, 1fr);
    }
    .row-actions {
      justify-content: stretch;
    }
    .row-actions .lg {
      width: 100%;
    }
  }
</style>
