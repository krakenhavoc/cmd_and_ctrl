<script lang="ts">
  // DeckCheck — the public deck coverage checker (#1631, ADR 0095 §5):
  // how much of a decklist the rules engine automates, and a button to
  // ask for the rest. Public like the roadmap: the report carries
  // names, oracle IDs and caveat sentences, never art or oracle text
  // (App.svelte lists "deckCheck" as a public route; the server mounts
  // POST /deck-coverage with no auth.Middleware).
  //
  // Two ways in: paste a link or a decklist here, or follow the
  // Discord bot's `/c2-deck-check` reply, which links to
  // #/deck-check?url=<link> for the full report the ephemeral message
  // only summarises.
  //
  // "Request these cards" needs a session with a Discord identity
  // (POST /deck-requests, ADR 0095 §3) — a signed-out or guest visitor
  // sees a sign-in prompt instead of a button that would 403. The
  // Discord-sign-in link has no return-target support today (neither
  // discordLoginHref nor the server's /auth/discord/start callback
  // round-trip carries one — see AGENTS.md's login flow, unchanged by
  // this page), so following it lands back on #/login, same as every
  // other "sign in first" link on the site (Home.svelte, SiteHeader).

  import { route } from "../lib/router";
  import { session } from "../lib/session";
  import { signedInUserID } from "../lib/myGames";
  import { discordLoginHref } from "../lib/api";
  import { catalogSearchHash } from "../lib/roadmap";
  import Icon from "../lib/components/Icon.svelte";
  import SiteHeader from "../lib/components/SiteHeader.svelte";
  import {
    checkDeck,
    requestDeck,
    groupCardsByBucket,
    canRequestCards,
    deckRequestOutcomeMessage,
    BUCKET_ORDER,
    BUCKET_LABELS,
    BUCKET_BLURBS,
    DeckCheckError,
    type CoverageReport,
    type CoverageBucket,
    type DeckRequestResponse,
    type DeckCoverageViolation,
  } from "../lib/deckcheck";

  type Tab = "link" | "paste";

  let tab = $state<Tab>("link");
  let urlInput = $state("");
  let textInput = $state("");

  let loading = $state(false);
  let errorMessage = $state("");
  let errorViolations = $state<DeckCoverageViolation[]>([]);
  let report = $state<CoverageReport | null>(null);

  let requestBusy = $state(false);
  let requestError = $state("");
  let requestResult = $state<DeckRequestResponse | null>(null);

  const isSignedIn = $derived($session !== null);
  const hasDiscordIdentity = $derived(signedInUserID($session) !== null);

  const groups = $derived(report ? groupCardsByBucket(report.cards) : null);
  const showRequestButton = $derived(canRequestCards(report));

  const totalCards = $derived.by(() => {
    if (!report) return 0;
    const c = report.counts;
    return c.manual + c.unreviewed + c.caveats + c.automated + c.no_effect;
  });

  function bucketPct(b: CoverageBucket): number {
    if (!report || totalCards === 0) return 0;
    return (report.counts[b] / totalCards) * 100;
  }

  function bucketCount(b: CoverageBucket): number {
    return report ? report.counts[b] : 0;
  }

  function violationLabel(v: DeckCoverageViolation): string {
    return v.card ? `${v.card}: ${v.message}` : v.message;
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
    report = null;
    requestResult = null;
    requestError = "";
    try {
      report = await checkDeck(isLink ? { url: value } : { text: value });
    } catch (err) {
      if (err instanceof DeckCheckError) {
        errorMessage = err.message;
        if (err.violations && err.violations.length > 0) errorViolations = err.violations;
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

  async function fileRequest(): Promise<void> {
    if (!report?.source_url) return;
    requestBusy = true;
    requestError = "";
    requestResult = null;
    try {
      requestResult = await requestDeck(report.source_url);
    } catch (err) {
      requestError = err instanceof Error ? err.message : "Couldn't file that request.";
    } finally {
      requestBusy = false;
    }
  }

  // #/deck-check?url=<link> pre-fills the link field and runs the
  // check on load — how the bot's reply opens the full report. Reread
  // on every route change, like Catalog.svelte's ?q=, so a second link
  // lands on its own deck.
  $effect(() => {
    const r = $route;
    if (r.name !== "deckCheck" || !r.url) return;
    tab = "link";
    urlInput = r.url;
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

<SiteHeader />

<section class="entry">
  <div class="stack">
    <div class="head">
      <p class="eyebrow">How much of your deck plays itself</p>
      <h1>Deck check</h1>
      <p class="lede">
        Paste a Moxfield or Archidekt link, or a plain decklist, and see how much of it the rules
        engine automates before you sit down. Anything missing can be requested straight from the
        report.
      </p>
    </div>

    <form class="input-card" onsubmit={submit}>
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
          placeholder="https://moxfield.com/decks/AbC123"
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
      {#if errorViolations.length > 0}
        <ul class="violations">
          {#each errorViolations as v (v.code + (v.card ?? "") + v.message)}
            <li>{violationLabel(v)}</li>
          {/each}
        </ul>
      {/if}
    {/if}

    {#if report && groups}
      <section class="report" aria-live="polite">
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
        </div>

        <ul class="legend">
          {#each BUCKET_ORDER as b (b)}
            <li class="lg-item {b}">
              <span class="dot {b}" aria-hidden="true"></span>
              <span class="lg-label">{BUCKET_LABELS[b]}</span>
              <span class="lg-n">{bucketCount(b)}</span>
            </li>
          {/each}
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

        {#if showRequestButton}
          <div class="request">
            {#if hasDiscordIdentity}
              <button type="button" class="primary" onclick={fileRequest} disabled={requestBusy}>
                {requestBusy ? "Requesting…" : "Request these cards"}
              </button>
              {#if requestResult}
                <p class="request-outcome" class:warn={requestResult.status === "rate_limited"}>
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
            {:else}
              <p class="signin-prompt">
                <a class="ghost-link" href={discordLoginHref()}>
                  <Icon name="chevronRight" size={12} /> Sign in with Discord to request these cards
                </a>
              </p>
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
    color: var(--gold-strong);
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

  /* --- input card -------------------------------------------------- */
  .input-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px;
    border-radius: var(--radius-lg);
    background: var(--surface);
    border: 1px solid var(--border);
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
    border-color: var(--gold);
    color: var(--gold-strong);
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
    color: var(--gold-strong);
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
    background: var(--gold);
  }
  .seg.automated,
  .dot.automated {
    background: var(--mint);
  }
  .seg.no_effect,
  .dot.no_effect {
    background: var(--border-strong);
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

  /* --- request button -------------------------------------------------- */
  .request {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding-top: 4px;
    border-top: 1px solid var(--border);
  }
  .request-outcome {
    margin: 0;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    color: var(--mint);
    font-size: 12.5px;
  }
  .request-outcome.warn {
    color: var(--gold-strong);
  }
  .request-outcome a {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: var(--gold-strong);
    text-decoration: none;
    font-family: var(--font-mono);
  }
  .request-outcome a:hover {
    text-decoration: underline;
  }
  .signin-prompt {
    margin: 0;
  }
  .ghost-link {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: var(--gold-strong);
    text-decoration: none;
    font-size: 13px;
    font-weight: 600;
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
    text-decoration-color: color-mix(in srgb, var(--gold) 50%, transparent);
    text-underline-offset: 2px;
  }
  a.cname:hover {
    color: var(--gold-strong);
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
    border-left: 2px solid color-mix(in srgb, var(--gold) 50%, transparent);
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
