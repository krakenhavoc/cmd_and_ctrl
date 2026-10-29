<script lang="ts">
  // PlayingAsBanner — the attention strip's line for a turn one player
  // plays for another (ADR 0167 §9, Emrakul, the Promised End).
  //
  //   - The controller: "Playing as Bob" with View my board, or
  //     "Viewing your board, Bob's turn is yours" with Play as Bob. The
  //     board never switches by itself after the turn begins; the jump
  //     to the other seat's decision is HandoffChip's, by the dock
  //     (#4434).
  //   - The controlled player: "Alice controls you this turn", and
  //     "Alice is choosing for you" while she answers a prompt of his.
  //   - Everyone else, spectators included, and a control still waiting
  //     for its turn: the text only.
  //   - #4154 (Master Warcraft): "Choosing Bob's attackers" for the
  //     player choosing them, "Alice is choosing your blockers" for the
  //     player chosen for, and, tagged "combat", who chooses this turn
  //     and the active player's "Alice chose your attackers".
  //
  // What it says is lib/playingAs.ts playingAsBanner's; this draws it.
  import type { PlayerView } from "../../protocol";
  import type { PlayingAsBannerModel } from "../../playingAs";
  import { seatColor } from "../../colors";
  import Icon from "../Icon.svelte";

  interface Props {
    line: PlayingAsBannerModel;
    seats: PlayerView[];
    // The toggle. Absent: no button.
    onToggle?: () => void;
  }
  const { line, seats, onToggle }: Props = $props();

  const seat = $derived(seats.find((s) => s.id === line.seatID) ?? null);
  const color = $derived(seat ? seatColor(seat.seat) : "var(--accent-strong)");
  const tag = $derived(
    line.kind === "playing"
      ? "playing as"
      : line.kind === "viewing"
        ? "your board"
        : line.kind === "controlled"
          ? "controlled"
          : line.kind === "scheduled"
            ? "next turn"
            : line.kind === "choosing"
              ? "combat"
              : "control",
  );
</script>

<div
  class="playing-as"
  class:playing={line.kind === "playing"}
  class:controlled={line.kind === "controlled"}
  role="status"
  aria-live="polite"
  aria-label="player control"
  style="--seat-color: {color}"
  data-kind={line.kind}
>
  <span class="tag"><Icon name="crown" size={11} /> {tag}</span>
  <span class="text">
    {#if seat}<span class="seat-dot" style="background:{color}"></span>{/if}
    <strong>{line.text}</strong>
    {#if line.source}<span class="src">· {line.source}</span>{/if}
    {#if line.note}<span class="note">{line.note}</span>{/if}
  </span>
  {#if onToggle && line.toggle}
    <button type="button" class="btn" onclick={onToggle}>{line.toggle.label}</button>
  {/if}
</div>

<style>
  /* The strip's card shell (Game.svelte's .att), edged in the colour of
     the seat the line is about. */
  .playing-as {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    padding: 9px 10px 9px 14px;
    background: color-mix(in srgb, var(--surface) 94%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid color-mix(in srgb, var(--seat-color) 55%, var(--border-strong));
    border-left: 4px solid var(--seat-color);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.35;
    box-sizing: border-box;
  }
  .tag {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    font-weight: 700;
    color: var(--accent-strong);
    white-space: nowrap;
  }
  .text {
    display: inline-flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    flex: 1 1 auto;
    min-width: 0;
  }
  .text strong {
    color: var(--fg);
    font-weight: 600;
  }
  .src {
    color: var(--fg-muted);
  }
  .note {
    flex-basis: 100%;
    color: var(--accent-strong);
    font-size: 12px;
  }
  .seat-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }
  .btn {
    font-size: 12px;
    padding: 4px 10px;
    white-space: nowrap;
  }
</style>
