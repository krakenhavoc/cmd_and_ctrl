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
  import { fanAngle, fanLift, handOverlap } from "../../handFan";
  import { play } from "../../sounds";
  import { settings } from "../../settings";
  import { canCastFromHand, type Legality } from "../../timing";

  interface Props {
    hand: ZoneView;
    isSelf: boolean;
    onPlayCard?: (card: CardView) => void;
    // #660: a card in hand can have activated abilities that function
    // THERE — cycling, typecycling (CR 702.29a/e). They ride
    // `hand_abilities` on the wire and open the same popover a
    // permanent's abilities do. Self hands only: an opponent's hand
    // cards carry no abilities on the wire in the first place.
    onActivateAbility?: (card: CardView, abilityIndex: number) => void;
    // Why the CR 307.1 sorcery-speed window is shut, or "" when it is
    // open. Passed through to the popover exactly as the battlefield
    // rows pass it; no cycling ability is sorcery-speed today, but a
    // hand ability that is would grey correctly.
    sorcerySpeedBlocked?: string;
    // S13.3 — when present, each hand card is checked for legality
    // and rendered greyed-out with a reason tooltip if illegal.
    // Self hands only; opponent hands always render face-down with
    // no per-card legality (we don't know what they are).
    snap?: GameView | null;
    viewerID?: string | null;
  }

  const {
    hand,
    isSelf,
    onPlayCard,
    onActivateAbility,
    sorcerySpeedBlocked = "",
    snap = null,
    viewerID = null,
  }: Props = $props();

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

  // #956 — the resting overlap for this layout, tightened for large
  // hands so the fan cannot outgrow its panel and get clipped. The
  // three resting values are the ones the CSS used to hard-code per
  // layout; handOverlap only ever tightens them.
  const overlapBase = $derived(layout === "stacked" ? 0.85 : isSelf ? 0.5 : 0.62);
  const overlap = $derived(handOverlap(cards.length, overlapBase));

  function handleCardClick(card: CardView): void {
    if (!isSelf) return;
    onPlayCard?.(card);
  }

  // legalityFor computes the cast-from-hand legality for a single
  // card. Cheap enough to call on every render: since S31 the
  // verdict is a scan of the server's own `legal_moves` list rather
  // than a derivation. Returns a "spectator" Legality for opponent
  // hands so the .timing-disabled class stays off (we never grey
  // opponent hands — face-down already says "not yours", and the
  // wire ships no move list for anyone but the viewer anyway).
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
  style:--hand-overlap={overlap}
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
          priority={isSelf}
          onActivateAbility={isSelf && (c.hand_abilities?.length ?? 0) > 0
            ? (idx) => onActivateAbility?.(c, idx)
            : undefined}
          {sorcerySpeedBlocked}
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
    align-items: flex-start;
    justify-content: center;
    min-height: 0;
    padding: 4px;
    /* Default "peek" state: clip to the top 62% of a card so the hand
       takes up well under its full vertical footprint while exposing
       the name, mana cost, art, and type line and hiding P/T +
       flavour / rules text. Hover lifts the whole fan up and over the
       board (see .hand:hover) to reveal full cards without pushing
       layout. PlayerPanel's .hand-zone reserves the same 62%. */
    max-height: calc(var(--card-h, 168px) * 0.62);
    /* #956 — never wider than the zone that holds it. The fan's own
       width is bounded by --hand-overlap below; this is the backstop
       for the case where it is not. */
    max-width: 100%;
    overflow: hidden;
    position: relative;
    z-index: 1;
    transition:
      max-height 220ms var(--ease),
      transform 220ms var(--ease);
  }
  /* Self hand expands on hover: overflow goes visible, the whole strip
     translates upward by the hidden 38% so full cards poke over the
     battlefield and the fan's bottom edge stays on the panel edge,
     and z-index jumps so nothing on the board occludes the revealed
     cards. */
  .hand:not(.opponent):hover {
    max-height: none;
    overflow: visible;
    transform: translateY(calc(var(--card-h, 168px) * -0.38));
    z-index: 20;
  }
  /* Opponent hands stay compact — they're face-down anyway and the
     peek/reveal interaction would feel wrong on someone else's hand. */
  .hand.opponent {
    max-height: calc(var(--card-h, 168px) * 0.55);
    overflow: hidden;
  }
  /* #956 — the overlap is handOverlap()'s answer, published by the
     markup as --hand-overlap, rather than three hard-coded per-layout
     values. A constant let the fan's width grow linearly with the
     card count (1 + (n-1)/2 card-widths for the self hand: 4 wide at
     seven cards, 7.5 at fourteen) until .panel's overflow: hidden cut
     it off — and --card-h's 168px floor means the cards do not shrink
     to fit. The resting values are unchanged: 0.5 self, 0.62 for an
     opponent's face-down fan, 0.85 stacked; only large hands tighten.
     The fallback matches the self hand's resting value.

     The overlap is sized relative to the card width (--card-w, which
     cascades from the card-size setting) rather than the container,
     because a % margin-left in flexbox resolves against the flex
     container, pushing cards off-screen. */
  .hand-slot {
    margin-left: calc(var(--card-w, 80px) * -1 * var(--hand-overlap, 0.5));
    transform-origin: bottom center;
    transition: transform 80ms ease;
  }
  .hand-slot:first-child {
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
     here. Their resting fan is tighter than the self hand's (0.62 vs
     0.5, see overlapBase in the script) because face-down stacks read
     better tightly packed. */
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
