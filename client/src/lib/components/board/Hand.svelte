<script lang="ts">
  // Hand renders the bottom-center hand strip for one player.
  //
  //   - Self: cards are face-up, click → play_card. Layout is a
  //     centered fan, each card rotated proportional to its offset
  //     from the centre, like the old Pixi version.
  //   - Opponents: cards are face-down. The server's FilterViewFor
  //     zeroes out the contents and the count is the only signal
  //     of how many they hold; we render `count` placeholder
  //     face-down cards in a tighter fan so the player count is
  //     visible at a glance without leaking content.

  import type { CardView, ZoneView } from "../../protocol";
  import Card from "./Card.svelte";

  interface Props {
    hand: ZoneView;
    isSelf: boolean;
    onPlayCard?: (card: CardView) => void;
  }

  const { hand, isSelf, onPlayCard }: Props = $props();

  // Synth a list of N face-down placeholder cards for opponent hands.
  // The server omits real CardView contents for opponent hands, so we
  // rely on `count` and never read scryfall_id / name from these.
  const placeholders = $derived.by((): CardView[] => {
    if (isSelf) return [];
    const n = hand.count;
    const out: CardView[] = [];
    for (let i = 0; i < n; i++) {
      out.push({
        instance_id: `opp-hand-${i}`,
        name: "",
        owner: "",
        controller: "",
      });
    }
    return out;
  });

  const cards = $derived(isSelf ? hand.cards : placeholders);

  // Per-card fan angle in degrees. Caps the total fan spread so very
  // large hands don't tip cards past sideways. Matches the Pixi math
  // from drawHandFan in the old table.ts.
  function fanAngle(i: number, n: number): number {
    if (n <= 1) return 0;
    const maxStepDeg = 7; // ~0.12 rad
    const maxTotalDeg = 60;
    const step = Math.min(maxStepDeg, maxTotalDeg / (n - 1));
    const totalDeg = step * (n - 1);
    return -totalDeg / 2 + i * step;
  }

  function fanLift(i: number, n: number): number {
    if (n <= 1) return 0;
    const center = (n - 1) / 2;
    const distFromCenter = Math.abs(i - center);
    return Math.round(distFromCenter * 3);
  }

  function handleCardClick(card: CardView): void {
    if (!isSelf) return;
    onPlayCard?.(card);
  }
</script>

<div class="hand" class:opponent={!isSelf} aria-label={isSelf ? "your hand" : "opponent hand"}>
  {#each cards as c, i (c.instance_id)}
    <div
      class="hand-slot"
      style:transform="rotate({fanAngle(i, cards.length)}deg) translateY({fanLift(
        i,
        cards.length,
      )}px)"
    >
      <Card card={c} faceDown={!isSelf} onClick={isSelf ? () => handleCardClick(c) : undefined} />
    </div>
  {/each}
  {#if !isSelf && hand.count === 0}
    <span class="empty">empty</span>
  {/if}
</div>

<style>
  .hand {
    display: flex;
    flex-direction: row;
    align-items: flex-end;
    justify-content: center;
    gap: -24px;
    height: 100%;
    min-height: 0;
    padding: 4px;
    overflow: visible;
  }
  .hand-slot {
    margin-left: -24px;
    transform-origin: bottom center;
    transition: transform 80ms ease;
  }
  .hand-slot:first-child {
    margin-left: 0;
  }
  .hand.opponent .hand-slot {
    margin-left: -42px;
  }
  .hand.opponent .hand-slot:first-child {
    margin-left: 0;
  }
  /* Opponent hand sizes are inherited from the parent
     .panel.opponent via --card-w/--card-h, so no explicit override
     here. The fan offset stays smaller (see .hand.opponent .hand-slot
     above) because face-down stacks read better tightly packed than
     the self hand's wider fan. */
  .empty {
    font-size: 11px;
    color: #6c7a99;
    align-self: center;
  }
</style>
