<script lang="ts">
  // Hand renders the bottom-center hand strip for one player.
  //
  //   - Self: cards are face-up, click → cast_spell. Layout is a
  //     centered fan, each card rotated proportional to its offset
  //     from the centre, like the old Pixi version.
  //   - Opponents: cards are face-down. The server's FilterViewFor
  //     zeroes out the contents and the count is the only signal
  //     of how many they hold; we render `count` placeholder
  //     face-down cards in a tighter fan so the player count is
  //     visible at a glance without leaking content.

  import type { Action } from "svelte/action";
  import type { CardView, GameView, ZoneView } from "../../protocol";
  import Card from "./Card.svelte";
  import { dealIn, dealOut } from "../../animations";
  import { play } from "../../sounds";
  import { settings } from "../../settings";
  import { canCastFromHand, type Legality } from "../../timing";

  interface Props {
    hand: ZoneView;
    isSelf: boolean;
    onPlayCard?: (card: CardView) => void;
    // S13.3 — when present, each hand card is checked for legality
    // and rendered greyed-out with a reason tooltip if illegal.
    // Self hands only; opponent hands always render face-down with
    // no per-card legality (we don't know what they are).
    snap?: GameView | null;
    viewerID?: string | null;
  }

  const { hand, isSelf, onPlayCard, snap = null, viewerID = null }: Props = $props();

  // Opponent hand composition:
  //   - hand.cards contains any cards the server has revealed to the
  //     viewer (Thoughtseize-style reveals, sticky per S13.5). Those
  //     render face-up.
  //   - The remaining (hand.count - revealed) cards render as face-
  //     down placeholders synthesized here.
  //   - settings.display.showOpponentHandCount collapses the full
  //     fan to a single face-down stand-in when the viewer prefers a
  //     compact opponent hand; revealed cards still render in full
  //     so a Thoughtseize peek isn't obscured by the setting.
  const cards = $derived.by((): CardView[] => {
    if (isSelf) return hand.cards;
    const revealed = hand.cards;
    const hidden = Math.max(0, hand.count - revealed.length);
    const showCount = $settings.display.showOpponentHandCount;
    // Cap the back-count when compact-opponent-hand mode is on, but
    // only if we also have no revealed cards; mixing a revealed card
    // with a single back stand-in reads cleaner than hiding the
    // revealed card entirely.
    const hiddenToShow = showCount ? hidden : Math.min(hidden, revealed.length > 0 ? hidden : 1);
    const out: CardView[] = [...revealed];
    for (let i = 0; i < hiddenToShow; i++) {
      out.push({
        instance_id: `opp-hand-${i}`,
        name: "",
        owner: "",
        controller: "",
      });
    }
    return out;
  });
  const layout = $derived($settings.display.handLayout);

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

  // legalityFor computes the cast-from-hand legality for a single
  // card. Cheap enough to call on every render; the predicate just
  // walks a few snapshot fields. Returns a "spectator" Legality for
  // opponent hands so the .timing-disabled class stays off (we
  // never grey opponent hands — face-down already says "not yours").
  function legalityFor(c: CardView): Legality {
    if (!isSelf) return { legal: true };
    return canCastFromHand(c, snap, viewerID);
  }

  // dealIn / dealOut are Svelte transitions (no mount/destroy callback
  // surface), so this action piggy-backs on the same element's
  // lifecycle: mount = "card entered hand" (draw), destroy = "card
  // left hand" (play). Mass events (opening-hand 7-card deal,
  // mulligan 7-card replacement) collapse to a single sound via the
  // 40ms same-name rate limit in sounds.ts — exactly what we want.
  const handLifecycle: Action<HTMLElement> = () => {
    play("draw");
    return {
      destroy() {
        play("play");
      },
    };
  };
</script>

<div
  class="hand"
  class:opponent={!isSelf}
  class:stacked={layout === "stacked"}
  aria-label={isSelf ? "your hand" : "opponent hand"}
>
  {#each cards as c, i (c.instance_id)}
    {@const leg = legalityFor(c)}
    <div
      class="hand-slot"
      class:timing-disabled={isSelf && !leg.legal}
      title={isSelf && !leg.legal ? leg.reason : undefined}
      style:transform={layout === "stacked"
        ? "none"
        : `rotate(${fanAngle(i, cards.length)}deg) translateY(${fanLift(i, cards.length)}px)`}
    >
      <!-- Inner wrapper carries the deal-in / deal-out transforms so
           they don't fight the .hand-slot's fan-layout transform. -->
      <div class="deal-wrap" in:dealIn out:dealOut use:handLifecycle>
        <Card
          card={c}
          faceDown={!isSelf && c.known_by_you !== true}
          showManaCost={isSelf}
          onClick={isSelf && leg.legal ? () => handleCardClick(c) : undefined}
        />
      </div>
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
  /* Stacked layout: drop the fan rotation and tightly overlap the
     cards into a deck-like pile. The overlap is sized relative to
     the card width (--card-w, which cascades from the card-size
     setting) rather than the container, because a % margin-left in
     flexbox resolves against the flex container, pushing cards off-
     screen. 85% overlap leaves a 15% sliver of each trailing card
     visible — enough to see how many are in hand without spreading
     them across the strip. */
  .hand.stacked .hand-slot {
    margin-left: calc(var(--card-w, 80px) * -0.85);
  }
  .hand.stacked .hand-slot:first-child {
    margin-left: 0;
  }
  /* Stacked layout needs left alignment — justify-content: center
     with near-zero total width collapses every slot onto the same
     point, which hides every card but the last. flex-start anchors
     the stack to the left edge so the strip grows visibly. */
  .hand.stacked {
    justify-content: flex-start;
    padding-left: 12px;
  }
  /* Opponent hand sizes are inherited from the parent
     .panel.opponent via --card-w/--card-h, so no explicit override
     here. The fan offset stays smaller (see .hand.opponent .hand-slot
     above) because face-down stacks read better tightly packed than
     the self hand's wider fan. */
  .empty {
    font-size: 11px;
    color: var(--fg-dim);
    align-self: center;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    font-weight: 600;
  }
  /* S13.3 — illegal-cast greying. Hover/zoom still works; only the
     click affordance is muted via the conditional onClick on Card. */
  .hand-slot.timing-disabled {
    opacity: 0.5;
    filter: grayscale(0.5) brightness(0.85);
  }
</style>
