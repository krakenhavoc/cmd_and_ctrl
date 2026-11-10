<script lang="ts">
  // PlanarZone — the face-up plane at a Planechase table (ADR 0216 §10).
  //
  // A plane is printed sideways, so the face-up card is drawn landscape:
  // the ordinary Card (hover zoom, focus, long-press and the art credit
  // all come with it) turned a quarter, inside a box the size of a
  // command-zone card laid on its side. Under it: the plane's name, a
  // hairline in the planar controller's seat colour, and the communal
  // planar deck's count as a mono tag. The text of the plane is one
  // gesture away in the hover zoom, read from Scryfall, never written by
  // C2.
  //
  // Where it sits is Board's call (DESIGN.md §5): never over the dock,
  // the stack pile, the hover zoom or an avatar. This file only draws.
  //
  // A planeswalk (a new card in face_up) slides the old plane out and
  // the new one in. With any reduced-motion signal the swap is instant;
  // the log line and the new name carry the news either way.

  import type { CardView, GameView } from "../../protocol";
  import Card from "./Card.svelte";
  import { L } from "../../labels";
  import { seatColor } from "../../colors";
  import { settings } from "../../settings";

  interface Props {
    view: GameView;
  }

  const { view }: Props = $props();

  const pc = $derived(view.planechase ?? null);
  const faceUp = $derived<CardView[]>(pc?.face_up ?? []);
  const controller = $derived(
    pc?.planar_controller ? view.seats.find((s) => s.id === pc.planar_controller) : undefined,
  );
  const hairline = $derived(controller ? seatColor(controller.seat) : "var(--border-strong)");
  const motion = $derived($settings.animations.enabled && !$settings.accessibility.reduceMotion);
  const deckWord = $derived(pc?.deck_count === 1 ? "plane" : "planes");
</script>

{#if pc}
  <section
    class="planar-zone"
    class:motion
    role="group"
    aria-label={L.planarZone}
    style:--planar-line={hairline}
    data-testid="planar-zone"
  >
    {#each faceUp as card (card.instance_id)}
      <div class="plane" title={card.name}>
        <div class="landscape">
          <div class="turn">
            <Card {card} />
          </div>
        </div>
        <span class="plane-name">{card.name}</span>
      </div>
    {:else}
      <!-- CR 901.5: the first plane turns up once every hand is kept. -->
      <div class="plane waiting">
        <div class="landscape empty" aria-hidden="true"></div>
        <span class="plane-name muted">The first plane turns up after the mulligans</span>
      </div>
    {/each}
    <span class="tag deck" role="status" aria-label={L.planarDeck(pc.deck_count)}
      >{pc.deck_count} {deckWord}</span
    >
    {#if controller}
      <span class="controller">
        <span class="dot" style:background={hairline} aria-hidden="true"></span>
        {controller.name}
      </span>
    {/if}
  </section>
{/if}

<style>
  /* A command-zone card on its side: --planar-h tall, 88:63 wide. */
  .planar-zone {
    --planar-h: 76px;
    --planar-w: calc(var(--planar-h) * 88 / 63);
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
    width: var(--planar-w);
    padding: 6px;
    border: 1px solid var(--border);
    border-top: 2px solid var(--planar-line);
    border-radius: var(--radius);
    background: color-mix(in srgb, var(--surface) 92%, transparent);
    box-shadow: var(--shadow-sm);
    pointer-events: auto;
  }
  .plane {
    display: flex;
    flex-direction: column;
    gap: 3px;
    width: 100%;
  }
  .landscape {
    position: relative;
    width: var(--planar-w);
    height: var(--planar-h);
  }
  .landscape.empty {
    border: 1px dashed var(--border-strong);
    border-radius: 6px;
  }
  /* The card is drawn portrait (the scan is a turned plane, like a
     split card's) and turned a quarter clockwise so it reads. */
  .turn {
    --card-w: var(--planar-h);
    --card-h: var(--planar-w);
    position: absolute;
    left: 50%;
    top: 50%;
    width: var(--planar-h);
    height: var(--planar-w);
    transform: translate(-50%, -50%) rotate(90deg);
  }
  .motion .plane:not(.waiting) {
    animation: plane-in 260ms var(--ease) both;
  }
  @keyframes plane-in {
    from {
      opacity: 0;
      translate: -12px 0;
    }
    to {
      opacity: 1;
      translate: 0 0;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .motion .plane:not(.waiting) {
      animation: none;
    }
  }
  .plane-name {
    max-width: 100%;
    overflow: hidden;
    color: var(--fg);
    font-size: 0.75rem;
    font-weight: 600;
    line-height: 1.2;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .plane-name.muted {
    color: var(--fg-muted);
    font-weight: 400;
    white-space: normal;
  }
  .tag.deck {
    font-family: var(--font-mono);
    font-size: 10px;
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  .controller {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    max-width: 100%;
    overflow: hidden;
    color: var(--fg-muted);
    font-size: 0.7rem;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .dot {
    width: 6px;
    height: 6px;
    flex: 0 0 auto;
    border-radius: 50%;
  }
</style>
