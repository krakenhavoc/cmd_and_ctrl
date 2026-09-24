<script lang="ts">
  // ExileStrip — the castable-from-exile strip beside the viewer's hand
  // (#1389). Every exiled card the viewer holds a permission over sits
  // here like a second hand: same card size, the same peek-and-lift,
  // the same hover zoom (Card writes the shared hover store), and a
  // click hands the card to the Board's one cast chain exactly as the
  // zone browser's exile button does (#874) — X, modes, targets, costs.
  //
  // What goes in, in what state, and what the corner badge says are
  // all decided in exileStrip.ts from server answers; this file only
  // lays them out.
  //
  //   - castable now: lit, clickable.
  //   - waiting on its window (a plotted card outside its main phase,
  //     a sorcery in an end step): greyed like an uncastable hand card.
  //   - waiting on a later turn (warp, plot, foretell on the turn they
  //     were set up): dimmed, with a "next turn" caption.
  //
  // The badge is a price tag on the top-right corner, where the card
  // prints its own cost, and it only appears when the price is NOT the
  // printed cost: 0 for a plotted card, 2 for airbend, the foretell
  // cost. Hovering it lists the printed cost and any other price.
  //
  // On a narrow panel the strip collapses to a count chip that opens
  // the cards in a popover, so a phone keeps its hand width.

  import type { CardView, GameView } from "../../protocol";
  import Card from "./Card.svelte";
  import { dealIn, dealOut } from "../../animations";
  import { handOverlap } from "../../handFan";
  import { canCastFromHand, type Legality } from "../../timing";
  import { exileCostBadge, exileStripEntries, type ExileStripEntry } from "../../exileStrip";
  import type { CastSourceZone } from "../../targeting";

  interface Props {
    view: GameView;
    viewerID: string | null;
    onCastCard?: (card: CardView, fromZone: CastSourceZone, face?: number) => void;
  }

  const { view, viewerID, onCastCard }: Props = $props();

  const entries = $derived(exileStripEntries(view, viewerID));
  const readyCount = $derived(entries.filter((e) => e.state === "now").length);
  // Two cards sit side by side; from three on they overlap like the
  // hand, tightening with handOverlap so a long strip cannot outgrow
  // the panel (#956's rule).
  const overlap = $derived(entries.length <= 2 ? -0.05 : handOverlap(entries.length, 0.4));

  // Collapsed-mode popover. Irrelevant on a wide panel, where the CSS
  // shows the cards inline whatever this says.
  let open = $state(false);
  $effect(() => {
    if (entries.length === 0) open = false;
  });

  function legalityFor(e: ExileStripEntry): Legality {
    if (e.state === "later") return { legal: false, reason: `Castable from exile ${e.hint}` };
    if (e.state === "waiting") {
      return { legal: false, reason: e.card.cant_cast || "Not castable from exile right now" };
    }
    // The same verdict the hand reads: the server's own move list,
    // which covers exile casts and knows about mana.
    return canCastFromHand(e.card, view, viewerID);
  }

  function cast(e: ExileStripEntry): void {
    if (!onCastCard) return;
    open = false;
    onCastCard(e.card, "exile", e.face);
  }

  function symbolClass(s: string): string {
    return /^[WUBRGC]$/.test(s) ? `sym-${s}` : "sym-generic";
  }
</script>

{#if entries.length > 0}
  <div class="exile-strip" class:open aria-label="castable from exile">
    <button
      type="button"
      class="strip-toggle"
      aria-expanded={open}
      aria-label={`${entries.length} exiled ${entries.length === 1 ? "card" : "cards"} you may cast, ${readyCount} ready`}
      onclick={() => (open = !open)}
    >
      <span class="toggle-label">exile</span>
      <span class="toggle-count" class:ready={readyCount > 0}>{entries.length}</span>
    </button>
    <div class="strip-body" style:--strip-overlap={overlap}>
      <span class="strip-tag" aria-hidden="true">from exile</span>
      <div class="strip-cards">
        {#each entries as e (e.card.instance_id)}
          {@const leg = legalityFor(e)}
          {@const badge = exileCostBadge(e.card)}
          <div
            class="strip-slot"
            class:blocked={!leg.legal}
            class:later={e.state === "later"}
            class:ready={leg.legal}
            title={leg.legal ? undefined : leg.reason}
          >
            <div class="deal-wrap" in:dealIn out:dealOut>
              <Card
                card={e.card}
                showManaCost={badge === null}
                onClick={leg.legal && onCastCard ? () => cast(e) : undefined}
              />
              {#if badge}
                <span class="cost-tag" title={badge.title} aria-label={badge.label}>
                  {#each badge.symbols as s, i (i)}
                    <span class="sym {symbolClass(s)}">{s}</span>
                  {/each}
                  {#if badge.life}
                    <span class="life">+{badge.life}♥</span>
                  {/if}
                </span>
              {/if}
              {#if e.hint}
                <span class="hint">{e.hint}</span>
              {:else if e.verb === "play" && leg.legal}
                <span class="hint verb">play</span>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    </div>
  </div>
{/if}

<style>
  .exile-strip {
    position: relative;
    flex: 0 1 auto;
    min-width: 0;
    align-self: flex-end;
    display: flex;
    align-items: flex-end;
    /* Same resting footprint as the self hand beside it (PlayerPanel's
       .hand-zone): 62% of a card, so the two rows line up. */
    height: calc(var(--card-h, 168px) * 0.62);
  }

  /* ---- inline (wide panel) ------------------------------------ */

  .strip-toggle {
    display: none;
  }
  .strip-body {
    position: relative;
    display: flex;
    align-items: flex-end;
    height: 100%;
    /* A gold hairline to the left: this is a second hand, and it
       should read as one — the same cards, set apart. */
    padding: 0 6px 0 12px;
    border-left: 1px dashed rgba(217, 180, 92, 0.45);
  }
  .strip-tag {
    position: absolute;
    left: 3px;
    bottom: 4px;
    writing-mode: vertical-rl;
    transform: rotate(180deg);
    font-family: var(--font-mono);
    font-size: 8.5px;
    font-weight: 700;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--gold);
    opacity: 0.75;
    pointer-events: none;
  }
  .strip-cards {
    display: flex;
    flex-direction: row;
    align-items: flex-start;
    /* Room above the cards for the price tag, which sits on the
       corner rather than inside the art. */
    padding: 7px 7px 0 4px;
    max-height: 100%;
    max-width: 100%;
    overflow: hidden;
    position: relative;
    z-index: 1;
    transition:
      max-height 220ms var(--ease),
      transform 220ms var(--ease);
  }
  /* The hand's lift, exactly: overflow goes visible and the row rises
     by the hidden 38% so whole cards show over the board. */
  .strip-cards:hover {
    max-height: none;
    overflow: visible;
    transform: translateY(calc(var(--card-h, 168px) * -0.38));
    z-index: 20;
  }
  .strip-slot {
    position: relative;
    margin-left: calc(var(--card-w, 80px) * -1 * var(--strip-overlap, 0.42));
    transition: transform 120ms var(--ease);
  }
  .strip-slot:first-child {
    margin-left: 0;
  }
  .strip-slot:last-child .cost-tag {
    right: -5px;
  }
  .strip-slot.ready:hover {
    transform: translateY(-6px);
    z-index: 2;
  }
  .deal-wrap {
    position: relative;
  }
  /* A castable card glows faintly gold — the one thing on it that says
     "this is yours to play" before you read the badge. */
  .strip-slot.ready .deal-wrap :global(.card) {
    box-shadow:
      0 0 0 1px rgba(217, 180, 92, 0.55),
      0 0 12px rgba(217, 180, 92, 0.25);
  }
  /* Dimmed with a filter rather than opacity: the cards overlap, and a
     translucent card lets the one beneath it — its badges included —
     show through, which reads as two cards smeared together. */
  /* On the card alone, so the caption and the tag stay legible. */
  .strip-slot.blocked .deal-wrap :global(.card) {
    filter: grayscale(0.5) brightness(0.6);
  }
  .strip-slot.later .deal-wrap :global(.card) {
    filter: grayscale(0.85) brightness(0.45);
  }

  /* ---- the price tag ------------------------------------------ */

  .cost-tag {
    position: absolute;
    top: -6px;
    /* The top-right of the part of the card you can SEE: the next card
       in the row covers this one's right-hand share, and a tag on the
       true corner would sit on top of its neighbour and read as the
       neighbour's price. The last card has no neighbour. */
    right: calc(var(--card-w, 80px) * max(var(--strip-overlap, 0), 0) - 5px);
    z-index: 6;
    display: inline-flex;
    align-items: center;
    gap: 1px;
    padding: 2px 3px;
    border-radius: 999px;
    /* The table's one accent (app.css): the tag is the thing on the
       card to read, and gold is how this board says so. */
    background: #1c1503;
    border: 1.5px solid var(--gold-strong);
    box-shadow:
      0 2px 8px rgba(0, 0, 0, 0.55),
      0 0 0 1px rgba(0, 0, 0, 0.4);
    cursor: help;
  }
  .sym {
    display: inline-grid;
    place-items: center;
    min-width: 15px;
    height: 15px;
    padding: 0 2px;
    box-sizing: border-box;
    border-radius: 999px;
    font-family: ui-monospace, Menlo, monospace;
    font-size: 10px;
    font-weight: 800;
    line-height: 1;
    color: #111;
    background: #cfd6e2;
  }
  .sym-W {
    background: #f4ead5;
  }
  .sym-U {
    background: #aad4ff;
  }
  .sym-B {
    background: #7a7390;
    color: #f4f0ff;
  }
  .sym-R {
    background: #ff9a85;
  }
  .sym-G {
    background: #92c493;
  }
  .sym-C {
    background: #c6cfdd;
  }
  .life {
    margin-left: 2px;
    font-size: 9px;
    font-weight: 700;
    color: #ffb3b3;
  }

  /* ---- captions ----------------------------------------------- */

  .hint {
    position: absolute;
    left: 50%;
    /* Inside the visible 62% peek, so it reads at rest. */
    top: calc(var(--card-h, 168px) * 0.62 - 28px);
    transform: translateX(-50%);
    z-index: 5;
    padding: 2px 7px;
    border-radius: 999px;
    white-space: nowrap;
    font-family: var(--font-mono);
    font-size: 8.5px;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: #b9d8ff;
    background: rgba(12, 22, 44, 0.92);
    border: 1px solid rgba(145, 195, 255, 0.55);
    pointer-events: none;
  }
  .hint.verb {
    color: var(--gold-strong);
    background: rgba(40, 30, 6, 0.92);
    border-color: rgba(217, 180, 92, 0.6);
  }

  /* ---- collapsed (narrow panel) -------------------------------
     .panel is a size container (PlayerPanel.svelte), so this asks
     how wide the viewer's own panel is, not the window. Below 640px
     the strip gives its width back to the hand and becomes a chip;
     the cards open in a popover above it that scrolls sideways inside
     itself, never the page. */
  @container (max-width: 640px) {
    .exile-strip {
      /* Static, so the popover below positions against .panel (which
         is position: relative) and can span the panel's width rather
         than hang off a chip near its right edge. */
      position: static;
      height: auto;
      align-self: flex-end;
    }
    .strip-toggle {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      height: 30px;
      padding: 0 8px 0 10px;
      border-radius: 999px;
      border: 1px solid rgba(217, 180, 92, 0.55);
      background: var(--surface-raised);
      color: var(--gold-strong);
      font-family: var(--font-mono);
      font-size: 10px;
      font-weight: 700;
      letter-spacing: 0.12em;
      text-transform: uppercase;
    }
    .toggle-count {
      display: inline-grid;
      place-items: center;
      min-width: 18px;
      height: 18px;
      border-radius: 999px;
      background: var(--surface-sunken, rgba(0, 0, 0, 0.35));
      color: var(--fg-muted);
      letter-spacing: 0;
    }
    .toggle-count.ready {
      background: var(--gold);
      color: var(--accent-fg, #1c1503);
    }
    .strip-body {
      display: none;
      position: absolute;
      left: 8px;
      right: 8px;
      bottom: calc(var(--card-h, 168px) * 0.62 + 16px);
      z-index: 40;
      height: auto;
      padding: 10px 10px 8px;
      border: 1px solid var(--border-strong);
      border-radius: 12px;
      background: var(--surface);
      box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
      /* Sideways scroll INSIDE the popover, never the page. */
      overflow-x: auto;
      overflow-y: hidden;
    }
    /* A phone-sized card: the panel's --card-h is sized for the board
       and would leave the popover room for one card. Set on the row
       rather than the popover so the popover's own offset above still
       reads the panel's value. */
    .strip-cards {
      --card-h: 154px;
      --card-w: 110px;
    }
    .exile-strip.open .strip-body {
      display: block;
    }
    .strip-tag {
      display: none;
    }
    .strip-cards,
    .strip-cards:hover {
      width: max-content;
      max-height: none;
      overflow: visible;
      transform: none;
    }
    .strip-slot {
      margin-left: 6px;
    }
    /* Side by side here, so the true corner is the visible one. */
    .cost-tag {
      right: -5px;
    }
    .hint {
      top: auto;
      bottom: 6px;
    }
  }
</style>
