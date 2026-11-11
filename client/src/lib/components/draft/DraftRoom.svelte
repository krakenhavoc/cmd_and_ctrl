<script lang="ts">
  // DraftRoom — one draft, for a browser that holds a seat in it (ADR
  // 0222 §5, #5438): the lobby, the picks and the finished pool.
  //
  // It owns the draft socket. Every `draft` frame is the whole view for
  // this seat and replaces the last one; the page keeps no state of its
  // own beyond what is chosen and what is in flight. The view holds this
  // seat's `pack`, `picks` and `pool` and, of every other seat, a name,
  // a status and a count (CR 905.1c).
  import { onMount, tick } from "svelte";

  import {
    cardCount,
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
    secondsLeft,
    SERVER_FULL_CODE,
    startDraft,
    takeOf,
    timerAnnouncement,
    viewAnnouncement,
    waitingLine,
    type DraftView,
  } from "../../draft";
  import { boosterName, PACKS_ARE_APPROXIMATIONS } from "../../draftSets";
  import { DraftSocket, type DraftError, type DraftSocketStatus } from "../../draftSocket";
  import { L } from "../../labels";
  import { navigate } from "../../router";
  import Icon from "../Icon.svelte";
  import Page from "../Page.svelte";
  import DraftPack from "./DraftPack.svelte";
  import DraftPool from "./DraftPool.svelte";
  import DraftSeats from "./DraftSeats.svelte";

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
  const seconds = $derived(
    view && view.state === "picking" ? secondsLeft(view, clock - frameAt) : null,
  );

  let pendingPick = $state<string | null>(null);
  let pickError = $state("");
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
    const said = viewAnnouncement(prev, next);
    if (said) announce = said;
    // The pick landed: the pack and its buttons are gone, so focus goes
    // to the line that says what happens next instead of to nowhere.
    if (prev?.pack && !next.pack && next.state === "picking") {
      void tick().then(() => waitingEl?.focus());
    }
  }

  function onStatus(next: DraftSocketStatus): void {
    status = next;
    if (next !== "connected") pendingPick = null;
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
    if (!view || view.state !== "picking" || !view.deadline) return;
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
    // In the lobby leaving only frees the seat. Once picking has begun
    // a bot takes it for the rest of the draft, so that asks first.
    if (view?.state === "picking") confirmLeave = true;
    else void leave();
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
      aria-label={L.draftTimer}
      data-seconds={seconds}
    >
      {#if isUrgent(seconds)}<Icon name="warn" size={14} />{/if}
      <span class="t-num">{clockText(seconds)}</span>
      <span class="t-word">{isUrgent(seconds) ? "left, pick now" : "left"}</span>
    </span>
  {/if}
{/snippet}

{#snippet meta()}
  {#if view}
    <div class="chips">
      <span class="chip">{view.set_name ?? view.set.toUpperCase()}</span>
      {#if boosterName(view.booster)}<span class="chip">{boosterName(view.booster)}</span>{/if}
      <span class="chip">{pickModeName(view.mode)}</span>
      <span class="chip">{view.seats} seats</span>
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
  {#if view && view.state !== "done" && !closed}
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
            A bot takes your seat for the rest of the draft, and you can’t come back to it.
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
                Each pick has {view.pick_seconds ?? 60} seconds.
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
            <p class="approx" data-testid="packs-approximate">{PACKS_ARE_APPROXIMATIONS}</p>
          </section>
        </div>
      {:else if view.state === "picking"}
        <section aria-labelledby="draft-pick-h">
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

        <div class="cols under">
          <section class="card" aria-labelledby="draft-picks-h">
            <h2 id="draft-picks-h">
              {L.draftYourPicks} <span class="n">{view.picks?.length ?? 0}</span>
            </h2>
            <DraftPool cards={view.picks ?? []} />
          </section>
          <section class="card side" aria-labelledby="draft-seats-h">
            <h2 id="draft-seats-h">Seats</h2>
            <DraftSeats {view} />
          </section>
        </div>
      {:else}
        <section class="card done" data-testid="draft-done">
          <h2 class="headline">{L.draftFinished}</h2>
          <p class="help">
            You drafted {cardCount((view.pool ?? view.picks ?? []).length)} from
            {view.set_name ?? view.set.toUpperCase()}.
          </p>
          <div class="next">
            <h3>{L.draftDeckbuildingNext}</h3>
            <p class="help">
              Building a deck from this pool and playing it is not ready yet. It is the next part of
              drafting to arrive. For now this is where a draft stops. The pool is not saved to your
              account: it is here while the draft stays open, and gone once it closes.
            </p>
          </div>
          {#if closed}
            <p class="help">
              This draft has closed. Your pool stays here until you leave the page.
            </p>
          {/if}
          <a class="link-arrow" href={DRAFT_NEW_HASH}>{L.draftNewDraft} →</a>
        </section>

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
  h2.headline {
    font-size: 24px;
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
  .waiting,
  .done {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  .done .headline {
    margin: 0;
  }
  .next {
    padding: 12px 14px;
    border-radius: var(--radius);
    background: var(--surface-sunken);
    border: 1px solid var(--border);
  }
  .under {
    margin-top: 22px;
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
