<script lang="ts">
  // DraftRoom — one draft, for a browser that holds a seat in it (ADR
  // 0222 §5, #5438): the lobby, the picks, and what follows them
  // (#5440): the deck built from the pool, and the rounds it is played
  // in.
  //
  // It owns the draft socket. Every `draft` frame is the whole view for
  // this seat and replaces the last one; the page keeps no state of its
  // own beyond what is chosen and what is in flight. The view holds this
  // seat's `pack`, `picks` and `pool` and, of every other seat, a name,
  // a status and a count (CR 905.1c).
  import { onMount, tick } from "svelte";

  import {
    DRAFT_NEW_HASH,
    draftErrorMessage,
    draftInviteURL,
    draftSocketURL,
    forgetDraftSeat,
    isDraftGone,
    isHost,
    isUrgent,
    clockText,
    leaveDraft,
    passTarget,
    pickModeName,
    previewDraft,
    progressLine,
    rememberDraftMatch,
    seatDraftMatch,
    secondsLeft,
    SERVER_FULL_CODE,
    startDraft,
    storedDraftIDs,
    takeOf,
    timerAnnouncement,
    viewAnnouncement,
    waitingLine,
    type DraftView,
  } from "../../draft";
  import { buildSecondsLeft } from "../../draftBuild";
  import {
    forgetSplit,
    loadSplit,
    moveTo,
    pruneSide,
    pruneSplits,
    saveSplit,
    type Zone,
  } from "../../draftSplit";
  import { matchLengthName, roundsAnnouncement } from "../../draftRounds";
  import { boosterName, PACKS_ARE_APPROXIMATIONS } from "../../draftSets";
  import {
    DRAFT_DECK_REFUSED,
    DraftSocket,
    type DraftError,
    type DraftSocketStatus,
  } from "../../draftSocket";
  import { L } from "../../labels";
  import { navigate } from "../../router";
  import { signedInUserID } from "../../myGames";
  import { session, setSession } from "../../session";
  import Icon from "../Icon.svelte";
  import Page from "../Page.svelte";
  import DraftBuild from "./DraftBuild.svelte";
  import DraftPack from "./DraftPack.svelte";
  import DraftRounds from "./DraftRounds.svelte";
  import DraftPool from "./DraftPool.svelte";
  import DraftSeats from "./DraftSeats.svelte";
  import DraftSplit from "./DraftSplit.svelte";
  import DraftStats from "./DraftStats.svelte";

  interface Props {
    id: string;
    /** The seat token this browser holds for the draft. */
    token: string;
    /** The server says the draft, or this seat, is no longer there. */
    ongone: () => void;
  }
  const { id, token, ongone }: Props = $props();

  let view = $state<DraftView | null>(null);
  let status = $state<DraftSocketStatus>("connecting");
  // The draft closed under us (close 1000, "draft closed").
  const closed = $derived(status === "closed");

  // The pick timer counts from the server's clock: `deadline - now` when
  // the frame was built, less the time passed here since it arrived.
  let frameAt = $state(0);
  let clock = $state(0);
  // The build timer (#5440) counts the same way, from `build`.
  const building = $derived(view?.state === "done" && view.phase === "building");
  const seconds = $derived(
    !view
      ? null
      : view.state === "picking"
        ? secondsLeft(view, clock - frameAt)
        : building
          ? buildSecondsLeft(view.build, clock - frameAt)
          : null,
  );

  let pendingPick = $state<string | null>(null);
  let pickError = $state("");
  // A deck on its way to the server, and the server's sentence for one
  // it refused (#5440).
  let pendingDeck = $state<string | null>(null);
  let deckError = $state("");
  // A finished draft is kept for a signed-in person (ADR 0222 §9).
  const kept = $derived(signedInUserID($session) !== null);
  let serverNote = $state("");
  let announce = $state("");

  let busy = $state(false);
  let actionError = $state("");
  let confirmLeave = $state(false);
  let copied = $state(false);
  let copyFailed = $state(false);

  let inviteEl = $state<HTMLInputElement | null>(null);
  let waitingEl = $state<HTMLElement | null>(null);

  let sock: DraftSocket | null = null;

  // The picks split into Main and the Sideboard (#5699) is this browser's
  // own state, kept under the draft's id so a reload or a reconnect
  // finds it again. It holds the Sideboard; every other pick is in Main.
  // svelte-ignore state_referenced_locally
  let side = $state<string[]>(loadSplit(id));
  const picks = $derived(view?.picks ?? []);
  const sideIDs = $derived(pruneSide(picks, side));

  function setSide(next: string[]): void {
    side = next;
    saveSplit(id, next);
  }
  function moveCard(card: string, to: Zone): void {
    setSide(moveTo(sideIDs, card, to));
  }
  // The splits of drafts this browser no longer holds a seat in go.
  onMount(() => pruneSplits([...storedDraftIDs(), id]));
  // A draft that is over has no use for its split.
  $effect(() => {
    if (view?.state === "done" && view.phase && view.phase !== "building") forgetSplit(id);
  });

  const take = $derived(view ? takeOf(view) : 0);
  const host = $derived(view ? isHost(view) : false);
  const hostName = $derived(view?.players.find((p) => p.host)?.name ?? "the host");
  const inviteLink = $derived(
    view && view.state === "lobby" && !view.solo && view.invite
      ? draftInviteURL(view.id, view.invite)
      : "",
  );
  const target = $derived(view ? passTarget(view) : undefined);
  const live = $derived(status === "connected");

  function onView(next: DraftView): void {
    const prev = view;
    view = next;
    frameAt = performance.now();
    clock = frameAt;
    serverNote = "";
    // A frame that leaves the same pack in hand is someone else's news
    // (a seat picked, a socket came or went): a pick of ours that is on
    // its way, or was just refused, still stands.
    const samePack =
      prev !== null &&
      next.state === "picking" &&
      (next.pack?.length ?? 0) > 0 &&
      prev.round === next.round &&
      prev.pick === next.pick;
    if (!samePack) {
      pendingPick = null;
      pickError = "";
    }
    // A deck that landed, or a build that ended, is no longer on its way.
    if (pendingDeck !== null && (next.phase !== "building" || next.build?.submitted)) {
      pendingDeck = null;
      deckError = "";
    }
    // After the picks the news is the build's and the rounds'.
    const said = next.phase ? roundsAnnouncement(prev, next) : viewAnnouncement(prev, next);
    if (said) announce = said;
    // The pick landed: the pack and its buttons are gone, so focus goes
    // to the line that says what happens next instead of to nowhere.
    if (prev?.pack && !next.pack && next.state === "picking") {
      void tick().then(() => waitingEl?.focus());
    }
  }

  function onStatus(next: DraftSocketStatus): void {
    status = next;
    if (next !== "connected") {
      pendingPick = null;
      pendingDeck = null;
    }
    if (next === "gone") {
      forgetDraftSeat(id);
      ongone();
    } else if (next === "closed") {
      forgetDraftSeat(id);
    }
  }

  function onError(err: DraftError): void {
    if (err.code === SERVER_FULL_CODE) {
      serverNote =
        err.message || "The server is full right now. Please try again in a few minutes.";
      return;
    }
    if (err.code === DRAFT_DECK_REFUSED || (pendingDeck !== null && err.id === pendingDeck)) {
      pendingDeck = null;
      deckError = err.message || "That deck was refused. Check it and submit again.";
      return;
    }
    pendingPick = null;
    pickError = err.message || "That pick was refused. Choose again.";
  }

  onMount(() => {
    const s = new DraftSocket({
      url: draftSocketURL(id, token),
      onView,
      onStatus,
      onError,
      // A browser cannot read a refused upgrade, so a failed dial asks
      // over HTTP whether the draft and the seat are still there.
      check: async () => {
        try {
          await previewDraft(id, { token });
          return "here";
        } catch (err) {
          return isDraftGone(err) ? "gone" : "unknown";
        }
      },
    });
    sock = s;
    s.connect();
    return () => s.disconnect();
  });

  // The countdown's tick, only while a timer is running.
  $effect(() => {
    const ticking =
      view?.state === "picking" ? !!view.deadline : building && !!view?.build?.deadline;
    if (!ticking) return;
    const t = setInterval(() => (clock = performance.now()), 250);
    return () => clearInterval(t);
  });

  // The timer speaks twice, at 30 and 10 seconds, and only to someone
  // who still has a card to take.
  let lastSeconds: number | null = null;
  $effect(() => {
    const s = seconds;
    const said = take > 0 ? timerAnnouncement(lastSeconds, s) : "";
    lastSeconds = s;
    if (said) announce = said;
  });

  function pick(cards: string[]): void {
    pickError = "";
    const sent = sock?.pick(cards) ?? null;
    if (!sent) {
      pickError = "Not connected. Pick again once the connection is back.";
      return;
    }
    pendingPick = sent;
  }

  function submitDeck(main: string[], basics: Record<string, number>): void {
    deckError = "";
    const sent = sock?.deck(main, basics) ?? null;
    if (!sent) {
      deckError = "Not connected. Submit again once the connection is back.";
      return;
    }
    pendingDeck = sent;
  }

  // Sits the viewer at their match's table and goes there. The server
  // sets the table seat's session as the cookie the game page reads;
  // the draft seat token stays in this browser, so the way back is this
  // page.
  async function playMatch(): Promise<void> {
    if (busy) return;
    busy = true;
    actionError = "";
    try {
      const seat = await seatDraftMatch(id, token);
      if (!seat.gameID) throw new Error("no table");
      rememberDraftMatch(seat.gameID, id);
      setSession(seat);
      sock?.disconnect();
      navigate(`#/games/${seat.gameID}`);
    } catch (err) {
      actionError = draftErrorMessage(
        err,
        "Couldn’t take your seat at the match. Try again in a moment.",
      );
      busy = false;
    }
  }

  async function start(): Promise<void> {
    if (busy) return;
    busy = true;
    actionError = "";
    try {
      const next = await startDraft(id, token);
      // The socket's frame may have landed first; `seq` only rises.
      if (!view || next.seq >= view.seq) onView(next);
    } catch (err) {
      actionError = draftErrorMessage(err, "Couldn’t start the draft. Try again in a moment.");
    } finally {
      busy = false;
    }
  }

  async function leave(): Promise<void> {
    if (busy) return;
    busy = true;
    actionError = "";
    try {
      await leaveDraft(id, token);
    } catch (err) {
      // A seat that is already gone has left: carry on out.
      if (!isDraftGone(err)) {
        actionError = draftErrorMessage(err, "Couldn’t leave the draft. Try again in a moment.");
        busy = false;
        return;
      }
    }
    sock?.disconnect();
    forgetDraftSeat(id);
    navigate(DRAFT_NEW_HASH);
  }

  function askLeave(): void {
    // In the lobby leaving only frees the seat, and a finished draft
    // has nothing left to lose. In between a bot takes the seat for the
    // rest of the draft, so that asks first.
    if (view?.state === "picking" || (view?.state === "done" && view.phase !== "finished")) {
      confirmLeave = true;
    } else void leave();
  }

  async function copyInvite(): Promise<void> {
    copyFailed = false;
    try {
      await navigator.clipboard.writeText(inviteLink);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch {
      // No clipboard here: select the link so it can be copied by hand.
      copyFailed = true;
      inviteEl?.focus();
      inviteEl?.select();
    }
  }
</script>

{#snippet timer()}
  {#if seconds !== null}
    <span
      class="timer"
      class:urgent={isUrgent(seconds)}
      role="timer"
      aria-label={building ? L.draftBuildTimer : L.draftTimer}
      data-seconds={seconds}
    >
      {#if isUrgent(seconds)}<Icon name="warn" size={14} />{/if}
      <span class="t-num">{clockText(seconds)}</span>
      <span class="t-word"
        >{!isUrgent(seconds) ? "left" : building ? "left, submit now" : "left, pick now"}</span
      >
    </span>
  {/if}
{/snippet}

{#snippet meta()}
  {#if view}
    <div class="chips">
      <span class="chip">{view.set_name ?? view.set.toUpperCase()}</span>
      {#if boosterName(view.booster)}<span class="chip">{boosterName(view.booster)}</span>{/if}
      {#if view.cube}<span class="chip">cube</span>{/if}
      <span class="chip">{pickModeName(view.mode)}</span>
      <span class="chip">{view.seats} seats</span>
      {#if view.state === "done"}<span class="chip">{matchLengthName(view.best_of)}</span>{/if}
      {#if view.solo}<span class="chip">solo, against bots</span>{/if}
      {#if !closed}
        {#if live}
          <span class="tag tag-connected">live</span>
        {:else}
          <span class="tag tag-connecting">reconnecting</span>
        {/if}
      {/if}
    </div>
  {/if}
{/snippet}

{#snippet actions()}
  {#if view && !closed}
    <button type="button" class="btn ghost" disabled={busy} onclick={askLeave}>
      {L.draftLeave}
    </button>
  {/if}
{/snippet}

<Page
  eyebrow="Draft"
  title={closed && view?.state !== "done" ? L.draftGone : (view?.name ?? "Draft")}
  {meta}
  {actions}
  width="wide"
>
  <!-- Said once per change, never per second (DESIGN.md §8). -->
  <p class="sr-only" aria-live="polite" data-testid="draft-announce">{announce}</p>

  <div class="stack">
    {#if serverNote}
      <p class="notice warn" role="status">
        <Icon name="warn" size={14} />
        <span>{serverNote}</span>
      </p>
    {/if}

    {#if closed && view?.state !== "done"}
      <div class="card">
        <p class="help">
          This draft was closed: nobody was in it for a while, or everyone left. Nothing from it is
          kept.
        </p>
        <a class="link-arrow" href={DRAFT_NEW_HASH}>{L.draftNewDraft} →</a>
      </div>
    {:else if !view}
      <p class="help" role="status">
        {status === "reconnecting" ? "Connection lost. Reconnecting…" : "Connecting to the draft…"}
      </p>
    {:else}
      {#if !live && !closed}
        <p class="notice warn" role="status" data-testid="draft-reconnecting">
          <Icon name="warn" size={14} />
          <span>
            Connection lost. Reconnecting… Your seat and your picks are kept.
            {#if view.state === "picking"}
              While you are away, a card is picked for you when the timer runs out.
            {/if}
          </span>
        </p>
      {/if}

      {#if confirmLeave}
        <div class="card ask" role="group" aria-label="leave this draft?">
          <p class="help strong">Leave this draft?</p>
          <p class="help">
            {#if view.state === "done"}
              A bot takes your seat and plays on with your deck, and you can’t come back to it. If
              you are in the middle of a match, you concede it.
            {:else}
              A bot takes your seat for the rest of the draft, and you can’t come back to it.
            {/if}
          </p>
          <div class="frow">
            <button type="button" class="btn" disabled={busy} onclick={() => void leave()}>
              {L.draftLeaveConfirm}
            </button>
            <button
              type="button"
              class="btn ghost"
              disabled={busy}
              onclick={() => (confirmLeave = false)}
            >
              {L.draftStay}
            </button>
          </div>
        </div>
      {/if}

      {#if actionError}
        <p class="notice err" role="alert">
          <Icon name="x" size={14} />
          <span>{actionError}</span>
        </p>
      {/if}

      {#if view.state === "lobby"}
        <div class="cols">
          <section class="card" aria-labelledby="draft-seats-h">
            <h2 id="draft-seats-h">
              Seats <span class="n">{view.players.length}/{view.seats}</span>
            </h2>
            <DraftSeats {view} />
            <p class="help foot">
              Every seat nobody takes is filled by a bot when the draft starts.
            </p>
          </section>

          <section class="card side" aria-labelledby="draft-start-h">
            <h2 id="draft-start-h">{host ? "Ready when you are" : "Waiting to start"}</h2>
            {#if inviteLink}
              <div class="invite">
                <p class="help">Send this link to the people you want at the draft.</p>
                <div class="invite-row">
                  <input
                    type="text"
                    readonly
                    value={inviteLink}
                    aria-label={L.draftInviteLink}
                    bind:this={inviteEl}
                    onfocus={(e) => e.currentTarget.select()}
                  />
                  <button type="button" class="btn" onclick={() => void copyInvite()}>
                    <Icon name="link" size={14} />
                    {L.draftCopyInvite}
                  </button>
                </div>
                <p class="help" role="status">
                  {#if copied}Copied.{:else if copyFailed}Couldn’t reach the clipboard. The link is
                    selected: copy it from the field.{/if}
                </p>
              </div>
            {:else if view.solo}
              <p class="help">
                This is a solo draft: every other seat is a bot. Sign in to draft with people.
              </p>
            {/if}

            {#if host}
              <p class="help">
                {view.mode === "pick_two"
                  ? "Three packs each, two cards from every pack."
                  : "Three packs each, one card from every pack."}
                Each pick has {view.pick_seconds ?? 60} seconds. Then everyone builds a 40-card deck and
                plays three rounds, {matchLengthName(view.best_of).toLowerCase()}.
              </p>
              <div class="frow">
                <button
                  type="button"
                  class="btn primary"
                  disabled={busy || !live}
                  onclick={() => void start()}
                >
                  {busy ? "Starting…" : L.draftStart}
                </button>
              </div>
            {:else}
              <p class="help">{hostName} starts the draft. This page moves on by itself.</p>
            {/if}
            {#if view.cube}
              <p class="approx" data-testid="packs-cube">
                The packs are cut from {view.set_name ?? "the host’s cube"}, a list of cards the
                host saved: it is shuffled and dealt, with no rarity slots, and no card is dealt
                twice.
              </p>
            {:else}
              <p class="approx" data-testid="packs-approximate">{PACKS_ARE_APPROXIMATIONS}</p>
            {/if}
          </section>
        </div>
      {:else if view.state === "picking"}
        <!-- The pack, and the picks beside it when there is room (#5699). -->
        <div class="pickgrid">
          <section class="packarea" aria-labelledby="draft-pick-h">
            <h2 class="progress" id="draft-pick-h" data-testid="draft-progress">
              {progressLine(view)}{#if target}{" "}<span class="to">to {target.name}</span>{/if}
            </h2>

            {#if pickError}
              <p class="notice err" role="alert">
                <Icon name="x" size={14} />
                <span>{pickError}</span>
              </p>
            {/if}

            {#if view.pack && view.pack.length > 0 && take > 0}
              <DraftPack
                pack={view.pack}
                {take}
                busy={pendingPick !== null || !live}
                blocked={!live ? "Reconnecting…" : pendingPick !== null ? "Sending your pick…" : ""}
                onpick={pick}
                {timer}
              />
            {:else}
              <div class="card waiting" data-testid="draft-waiting">
                <h3 tabindex="-1" bind:this={waitingEl}>{L.draftWaiting}</h3>
                <p class="help">{waitingLine(view)}</p>
                {@render timer()}
              </div>
            {/if}
          </section>

          <aside class="picksarea" aria-labelledby="draft-picks-h">
            <h2 id="draft-picks-h">
              {L.draftYourPicks} <span class="n">{picks.length}</span>
            </h2>
            <DraftSplit
              cards={picks}
              side={sideIDs}
              onmove={moveCard}
              ids={{ main: "draft-picks-main", side: "draft-picks-side" }}
              emptyMain="No picks yet."
              emptySide="Drag a card here, or click it, to set it aside."
            />
          </aside>

          <div class="statsarea">
            <DraftStats main={picks.filter((c) => !sideIDs.includes(c.id))} pool={picks} />
          </div>

          <section class="card side seatsarea" aria-labelledby="draft-seats-h">
            <h2 id="draft-seats-h">Seats</h2>
            <DraftSeats {view} />
          </section>
        </div>
      {:else if building}
        <DraftBuild
          {view}
          {live}
          pending={pendingDeck !== null}
          error={deckError}
          onsubmit={submitDeck}
          side={sideIDs}
          onside={setSide}
          {timer}
        />
      {:else}
        <DraftRounds
          {view}
          {live}
          {busy}
          error=""
          {closed}
          {kept}
          onplay={() => void playMatch()}
        />

        <section aria-labelledby="draft-pool-h">
          <h2 id="draft-pool-h">
            {L.draftYourPool} <span class="n">{(view.pool ?? view.picks ?? []).length}</span>
          </h2>
          <DraftPool cards={view.pool ?? view.picks ?? []} images />
        </section>
      {/if}
    {/if}
  </div>
</Page>

<style>
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
  .stack {
    display: flex;
    flex-direction: column;
    gap: 22px;
    min-width: 0;
  }
  .cols {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(280px, 380px);
    gap: 22px;
    align-items: start;
  }
  @media (max-width: 819px) {
    .cols {
      grid-template-columns: minmax(0, 1fr);
    }
  }
  h2 {
    margin: 0 0 12px;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    letter-spacing: -0.005em;
    text-transform: none;
    color: var(--fg);
  }
  h2 .n {
    margin-left: 6px;
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  /* The pick's place in the draft: mono, as counts are. */
  h2.progress {
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 700;
    letter-spacing: 0.04em;
    font-variant-numeric: tabular-nums;
    color: var(--fg);
  }
  .progress .to {
    color: var(--fg-muted);
  }
  h3 {
    margin: 0 0 6px;
    font-family: var(--font-display);
    font-size: 17px;
    font-weight: 700;
    color: var(--fg);
  }
  h3:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 4px;
    border-radius: var(--radius-sm);
  }
  .help {
    margin: 0;
    font-size: 0.875rem;
    line-height: 1.6;
    color: var(--fg-muted);
  }
  .help.strong {
    font-weight: 700;
    color: var(--fg);
  }
  .help.foot {
    margin-top: 12px;
  }
  .side {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .side h2 {
    margin-bottom: 0;
  }
  .approx {
    margin: 0;
    padding-top: 14px;
    border-top: 1px solid var(--border);
    font-size: 0.75rem;
    line-height: 1.6;
    color: var(--fg-muted);
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px 8px;
    margin-top: 12px;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    padding: 3px 10px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.04em;
    color: var(--fg-muted);
  }
  .frow {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }
  .invite {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .invite-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .invite-row input {
    flex: 1 1 18ch;
    min-width: 0;
    margin: 0;
    font-family: var(--font-mono);
    font-size: 12px;
  }
  .ask,
  .waiting {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  /* The pick screen (#5699): the pack, with the picks to its right on a
     wide screen and below it on a narrow one. Named areas keep the
     reading order: pack, picks, stats, seats. */
  .pickgrid {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(280px, 380px);
    grid-template-areas:
      "pack picks"
      "stats picks"
      "seats picks";
    grid-template-rows: auto auto minmax(0, 1fr);
    gap: 22px;
    align-items: start;
  }
  .packarea {
    grid-area: pack;
    min-width: 0;
  }
  .statsarea {
    grid-area: stats;
    min-width: 0;
  }
  .seatsarea {
    grid-area: seats;
    min-width: 0;
  }
  .picksarea {
    grid-area: picks;
    min-width: 0;
    /* Stays beside the pack while it scrolls, and scrolls on its own when
       the picks outgrow the window. It sits below the sticky site
       header (z 30, about 56px tall). */
    position: sticky;
    top: 72px;
    max-height: calc(100vh - 88px);
    overflow-y: auto;
  }
  .picksarea h2 {
    margin: 0 0 12px;
  }
  @media (max-width: 819px) {
    .pickgrid {
      grid-template-columns: minmax(0, 1fr);
      grid-template-areas:
        "pack"
        "picks"
        "stats"
        "seats";
      grid-template-rows: none;
    }
    .picksarea {
      position: static;
      max-height: none;
      overflow-y: visible;
    }
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
  section > .notice {
    margin-bottom: 12px;
  }
  .notice.err {
    background: color-mix(in srgb, var(--danger) 8%, transparent);
    border: 1px solid color-mix(in srgb, var(--danger) 35%, transparent);
  }
  .notice.warn {
    background: var(--surface-sunken);
    border: 1px solid var(--border-strong);
  }

  /* The pick timer. Urgent in the last ten seconds: the danger ink, a
     heavier weight, a warning glyph and the words "pick now", so it is
     never colour alone. No motion. */
  .timer {
    display: inline-flex;
    align-items: baseline;
    gap: 6px;
    padding: 4px 10px;
    border-radius: var(--radius);
    border: 1px solid var(--border-strong);
    background: var(--surface-sunken);
    color: var(--fg);
  }
  .timer :global(svg) {
    align-self: center;
  }
  .t-num {
    font-family: var(--font-mono);
    font-size: 1.125rem;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
  }
  .t-word {
    font-size: 0.75rem;
    color: var(--fg-muted);
  }
  .timer.urgent {
    border-color: var(--danger-line);
    background: var(--danger-soft);
    color: var(--danger);
  }
  .timer.urgent .t-word {
    font-weight: 700;
    color: var(--fg);
  }
</style>
