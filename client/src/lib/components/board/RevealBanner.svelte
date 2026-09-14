<script lang="ts">
  // RevealBanner renders the S22 broadcast reveal window (CR 701.16)
  // into the board's attention strip.
  //
  // A reveal is the one thing on GameView that every seat receives
  // identically, and it is the only way an opponent ever learns what
  // came off the top of somebody's library — the zone itself is
  // wholesale-hidden on their wire, and the cards usually go straight
  // back into a hidden zone afterwards. So this surface is not
  // decoration: without it the frame is invisible and Dark Confidant,
  // Hermit Druid and every tutor's "reveal it" are silent again.
  //
  // It is deliberately NOT a modal. The scry family gets a modal
  // because it asks the controller a question and blocks on the
  // answer; a reveal asks nobody anything, arrives at three seats who
  // did not act, and must never take the board away from a player who
  // is mid-decision. Same strip as the bot feed and the toasts, same
  // aria-live="polite".
  //
  // Cards are drawn from printed identity alone. There are no instance
  // IDs on a reveal entry, on purpose — see protocol/reveal_frame.go —
  // so this cannot and does not link a banner to a board object.
  import { untrack } from "svelte";
  import { scryfallImageURL } from "../../cardImage";
  import { seatColor } from "../../colors";
  import type { GameView, RevealView } from "../../protocol";
  import {
    dismissReveal,
    emptyRevealState,
    hiddenRevealCount,
    primeRevealState,
    revealHeadline,
    trackReveals,
    type RevealState,
  } from "../../reveals";
  import Icon from "../Icon.svelte";

  interface Props {
    snap: GameView | null;
  }
  const { snap }: Props = $props();

  let cueState = $state<RevealState>(emptyRevealState());
  let primed = false;

  // The fold reads cueState to produce the next one, so the read is
  // untracked: tracking it would make this effect depend on its own
  // write and loop. The dependency that matters is `snap`, read first
  // and on every pass.
  $effect(() => {
    const reveals = snap?.reveals;
    untrack(() => {
      if (!primed) {
        // The first frame this client has seen. Whatever is already in
        // the window happened before we were looking, so swallow it
        // rather than announcing old news as if it were live.
        primed = true;
        cueState = primeRevealState(reveals);
        return;
      }
      cueState = trackReveals(cueState, reveals, Date.now());
    });
  });

  // Cues age out on a timer of their own. A table can sit on one frame
  // for a long time while somebody thinks, and a banner that only
  // expires when the next snapshot arrives would still be up.
  $effect(() => {
    const t = setInterval(() => {
      untrack(() => {
        if (cueState.cues.length === 0) return;
        cueState = trackReveals(cueState, undefined, Date.now());
      });
    }, 500);
    return () => clearInterval(t);
  });

  function nameOfSeat(seat: number): string {
    const p = snap?.seats?.[seat];
    return p ? p.name : "";
  }

  function dismiss(r: RevealView) {
    cueState = dismissReveal(cueState, r.seq);
  }
</script>

{#each cueState.cues as cue (cue.reveal.seq)}
  {@const r = cue.reveal}
  {@const more = hiddenRevealCount(r)}
  <div class="att reveal" role="status" aria-live="polite">
    <span class="att-label gold">
      <Icon name="spark" size={12} />
      revealed
    </span>
    <span class="att-text">
      {#if r.seat >= 0}
        <span class="seat-dot" style="background:{seatColor(r.seat)}"></span>
      {/if}
      <strong>{revealHeadline(r, nameOfSeat(r.seat))}</strong>
      {#if r.reason}
        <span class="muted">· {r.reason}</span>
      {/if}
    </span>

    <span class="reveal-cards" aria-label="revealed cards">
      {#each r.cards as c, i (i)}
        <span class="reveal-card" title={c.name}>
          {#if c.scryfall_id}
            <img src={scryfallImageURL(c.scryfall_id, "small")} alt={c.name} loading="lazy" />
          {:else}
            <span class="reveal-fallback">{c.name}</span>
          {/if}
          <span class="sr-only">{c.name}{c.mana_cost ? ` ${c.mana_cost}` : ""}</span>
        </span>
      {/each}
      {#if more > 0}
        <span class="reveal-more" title="{more} more cards were revealed">+{more}</span>
      {/if}
    </span>

    <button type="button" class="ghost att-close" onclick={() => dismiss(r)} aria-label="dismiss">
      <Icon name="x" size={12} />
    </button>
  </div>
{/each}

<style>
  /* The strip's .att shell comes from Game.svelte; only the card row
     is this component's business. */
  .reveal-cards {
    display: flex;
    align-items: center;
    gap: 4px;
    flex-wrap: wrap;
    min-width: 0;
  }
  .reveal-card {
    position: relative;
    display: block;
    width: 34px;
    height: 48px;
    border-radius: 3px;
    overflow: hidden;
    background: rgba(255, 255, 255, 0.05);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.45);
  }
  .reveal-card img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .reveal-fallback {
    display: block;
    padding: 3px;
    font-size: 7px;
    line-height: 1.15;
    overflow: hidden;
    color: var(--text-dim, #b9b3a7);
  }
  .reveal-more {
    font-size: 11px;
    font-variant-numeric: tabular-nums;
    padding: 2px 6px;
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-dim, #b9b3a7);
  }
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
</style>
