<script lang="ts">
  // BattlefieldRow renders a horizontal strip of battlefield cards
  // for one type bucket (CREATURES at the top, LANDS on the middle
  // left). The row owns layout (flex with wrap) and per-card visual
  // state derivation; the parent supplies the click handler so combat
  // / tap routing stays centralised.
  //
  // Cards are sorted by battle_x so a future row-reorder UX has a
  // stable handle. battle_y is intentionally ignored: typed rows make
  // a 2-D placement meaningless, and the server already defaults
  // both axes to 0 for cards that have never been positioned.

  import type { CardView } from "../../protocol";
  import Card from "./Card.svelte";
  import { etbPulse } from "../../animations";

  interface Props {
    label: string;
    cards: CardView[];
    viewerID: string | null;
    selectedCombatCardID?: string | null;
    onCardClick?: (card: CardView, ev: MouseEvent) => void;
    // onActivateManaAbility — fires `activate_mana_ability` for a
    // battlefield permanent the viewer controls. Routed down to
    // Card.svelte so the right-click / context menu can hit it.
    // Undefined suppresses the menu entirely (opponent panels).
    onActivateManaAbility?: (card: CardView, abilityIndex: number) => void;
    // #1438: "Tap (no mana)" in the same menu — a left-click on a
    // mana source now taps it FOR mana, so the raw tap moved here.
    onRawTap?: (card: CardView) => void;
    // S21 sub-PR 2: CR 602 activated abilities, same menu.
    onActivateAbility?: (card: CardView, abilityIndex: number) => void;
    // S31: why the CR 307.1 sorcery-speed window is shut, or "" when
    // it is open. Pass-through to Card → ManaAbilityMenu, which greys
    // `sorcery_speed` abilities with it.
    sorcerySpeedBlocked?: string;
    // compact — one card size down (PlayerPanel's --card-w-sm). Used
    // for the middle band: non-creature permanents and lands.
    compact?: boolean;
    // strip — piles instead of a wrapping grid. Lands: untapped
    // copies of the same name stack into one pile with a count
    // badge (four Forests take the width of one), tapped lands sort
    // to the right of every pile so the untapped count reads at a
    // glance, and hovering a pile spreads it so each copy can be
    // clicked on its own.
    strip?: boolean;
    // S24: the Equipment and Auras attached to each host, keyed by
    // the host's instance ID. Derived by PlayerPanel from the
    // one-directional `attached_to` on the wire. They are drawn
    // inside the host's own listitem, offset behind it — the same
    // negative-margin overlap the land strip uses — so a sword reads
    // as being ON the creature rather than as a separate permanent.
    attachmentsByHost?: Record<string, CardView[]>;
    // S24: the name of the player each Curse-style permanent enchants,
    // keyed by the permanent's instance ID. A card attached to a
    // PLAYER has no host to be drawn behind, so the relation would be
    // invisible without this. Derived by PlayerPanel, which is the
    // component that can see the seat list.
    curseTargets?: Record<string, string>;
  }

  const {
    label,
    cards,
    selectedCombatCardID = null,
    onCardClick,
    onActivateManaAbility,
    onRawTap,
    onActivateAbility,
    sorcerySpeedBlocked = "",
    compact = false,
    strip = false,
    attachmentsByHost = {},
    curseTargets = {},
  }: Props = $props();

  const sorted = $derived.by(() => {
    const byX = [...cards].sort((a, b) => (a.battle_x ?? 0) - (b.battle_x ?? 0));
    if (!strip) return byX;
    // Stable partition: untapped first, tapped after, order kept within each.
    return [...byX.filter((c) => !c.tapped), ...byX.filter((c) => c.tapped)];
  });

  // Piles (strip only). One pile per NAME among the untapped lands,
  // in first-appearance order, then one single-card pile per tapped
  // land. A card that carries an attachment (an enchanted land) or is
  // a commander keeps a pile of its own: stacking it under identical
  // basics would hide the very thing that makes it different. The
  // non-strip rows render `sorted` as one pile each so the markup
  // has a single shape.
  interface Pile {
    key: string;
    tapped: boolean;
    cards: CardView[];
  }
  const piles = $derived.by((): Pile[] => {
    if (!strip) return sorted.map((c) => ({ key: c.instance_id, tapped: !!c.tapped, cards: [c] }));
    const out: Pile[] = [];
    const byName = new Map<string, Pile>();
    for (const c of sorted) {
      const solo =
        !!c.tapped || !!c.is_commander || (attachmentsByHost[c.instance_id] ?? []).length > 0;
      if (solo) {
        out.push({ key: c.instance_id, tapped: !!c.tapped, cards: [c] });
        continue;
      }
      const name = c.name || c.instance_id;
      const hit = byName.get(name);
      if (hit) {
        hit.cards.push(c);
        continue;
      }
      const p: Pile = { key: `pile:${name}`, tapped: false, cards: [c] };
      byName.set(name, p);
      out.push(p);
    }
    return out;
  });
</script>

<div class="row" class:compact class:strip data-zone={label}>
  <span class="row-label" aria-hidden="true">
    {label}
    {#if cards.length > 0}<span class="row-count">{cards.length}</span>{/if}
  </span>
  <div class="row-cards" role="list" aria-label={label}>
    {#each piles as p (p.key)}
      <div
        class="pile"
        class:tapped={p.tapped}
        class:multi={p.cards.length > 1}
        title={p.cards.length > 1 ? `${p.cards.length} × ${p.cards[0].name}` : undefined}
      >
        {#each p.cards as c, i (c.instance_id)}
          <div role="listitem" class:tapped={!!c.tapped} style:--i={i} use:etbPulse>
            <div
              class="host-stack"
              class:has-attachments={(attachmentsByHost[c.instance_id] ?? []).length > 0}
            >
              {#each attachmentsByHost[c.instance_id] ?? [] as a (a.instance_id)}
                <div class="attachment">
                  <Card
                    card={a}
                    enchantedPlayer={curseTargets[a.instance_id]}
                    onClick={onCardClick}
                    onActivateManaAbility={onActivateManaAbility
                      ? (idx) => onActivateManaAbility(a, idx)
                      : undefined}
                    onRawTap={onRawTap ? () => onRawTap(a) : undefined}
                    onActivateAbility={onActivateAbility
                      ? (idx) => onActivateAbility(a, idx)
                      : undefined}
                    {sorcerySpeedBlocked}
                  />
                </div>
              {/each}
              <Card
                card={c}
                enchantedPlayer={curseTargets[c.instance_id]}
                selected={selectedCombatCardID === c.instance_id}
                attacking={!!c.attacking_target}
                blocking={!!c.blocking_target}
                onClick={onCardClick}
                onActivateManaAbility={onActivateManaAbility
                  ? (idx) => onActivateManaAbility(c, idx)
                  : undefined}
                onRawTap={onRawTap ? () => onRawTap(c) : undefined}
                onActivateAbility={onActivateAbility
                  ? (idx) => onActivateAbility(c, idx)
                  : undefined}
                {sorcerySpeedBlocked}
              />
            </div>
          </div>
        {/each}
        {#if p.cards.length > 1}
          <span class="pile-count" aria-hidden="true">{p.cards.length}</span>
        {/if}
      </div>
    {/each}
  </div>
</div>

<style>
  .row {
    position: relative;
    padding: 18px 6px 6px;
    min-height: 0;
    min-width: 0;
    overflow: auto;
  }
  .row.compact {
    --card-w: var(--card-w-sm, 88px);
    --card-h: var(--card-h-sm, 123px);
  }
  .row-label {
    position: absolute;
    top: 4px;
    left: 6px;
    display: inline-flex;
    align-items: baseline;
    gap: 6px;
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--fg-dim);
    pointer-events: none;
    font-weight: 600;
    white-space: nowrap;
  }
  .row-count {
    letter-spacing: 0;
    font-weight: 500;
  }
  /* A host and everything attached to it. The attachments are laid
     out first and overlapped leftwards behind the host, which is
     drawn last so it sits on top; the wrapper claims only the host's
     width, so a creature carrying two swords does not reflow the row
     out from under its neighbours. */
  .host-stack {
    display: flex;
    flex-direction: row;
    align-items: flex-start;
  }
  .host-stack.has-attachments {
    margin-left: calc(var(--card-w, 88px) * 0.34);
  }
  .host-stack .attachment {
    flex: 0 0 auto;
    margin-right: calc(var(--card-w, 88px) * -0.72);
    transform: translateY(8px);
    filter: brightness(0.88);
  }
  .host-stack .attachment:hover {
    z-index: 7;
    filter: none;
  }
  /* An attachment's failed-art pip (#33) goes in its bottom-left
     corner, tapped or not. The pip cannot paint over the tile after
     it — each tile is its own stacking context — so it has to sit
     where that tile leaves the attachment uncovered, whatever is
     tapped. Upright next to an upright tile, that is the left 28% of
     the attachment. A tapped host turns to cover a band across the
     middle of its height, leaving the bottom of the attachment (the
     8px drop helps) — Card's 22px, the old spot, went under it. A
     tapped attachment turns its bottom edge to the left, clear of an
     upright host, and its bottom-left corner to the top-left, left of a
     turned host. 2px up keeps it clear of the band at the smallest
     size, and the left edge keeps it inside the sliver. It covers the
     start of an AUTO badge or keyword row there, which on an Aura or
     Equipment is rare. boardArtPip.test.ts checks every combination. */
  .host-stack .attachment :global(.card) {
    --art-error-top: calc(100% - 16px);
    --art-error-left: 0px;
  }
  /* Rows grow from the middle of the table: cards centre in the row
     rather than piling up in the top-left corner, so a two-creature
     board reads as a board and not as a corner. */
  .row-cards {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: center;
    gap: 8px;
    align-content: flex-start;
    align-items: flex-start;
    height: 100%;
  }
  /* Outside the strip a pile is one card and nothing more; the
     wrapper exists so the markup has one shape. */
  .pile {
    position: relative;
    display: flex;
    flex-direction: row;
    align-items: flex-start;
    flex: 0 0 auto;
  }
  /* Land strip: no wrap. Between piles each overlaps the previous so
     the name band stays readable; tapped piles (rotated by
     Card.svelte) are sorted to the end and given room for their
     rotated width. The margin rules below are read by
     boardArtPip.test.ts, which resolves their cascade by hand — keep
     each on its own `.row.strip .pile…` selector. */
  .row.strip {
    /* Room for the pile count badges, which sit 6px above the tiles,
       under the row label. */
    padding-top: 26px;
  }
  .row.strip .row-cards {
    flex-wrap: nowrap;
    justify-content: flex-start;
    gap: 0;
    padding-right: calc(var(--card-h, 123px) * 0.2);
  }
  .row.strip .pile {
    margin-left: calc(var(--card-w, 88px) * -0.6);
  }
  .row.strip .pile:first-child {
    margin-left: 0;
  }
  .row.strip .pile.tapped {
    margin-left: calc(var(--card-w, 88px) * -0.35);
  }
  .row.strip .pile:not(.tapped) + .pile.tapped {
    margin-left: calc(var(--card-h, 123px) * 0.2);
  }
  /* Every land tapped: the first is tapped too. Without this rule the
     .tapped overlap above, as specific as :first-child and later,
     won, and pushed the first land a third of its turned width out of
     the row's clip, its failed-art pip (#33) with it. The turned tile
     overhangs its box by (h - w) / 2 on each side; that is its room. */
  .row.strip .pile.tapped:first-child {
    margin-left: calc((var(--card-h, 123px) - var(--card-w, 88px)) / 2);
  }
  .row.strip .pile:hover {
    z-index: 6;
  }
  /* Inside a pile every copy after the first sits 4px right and 3px
     up of the one before, so the pile reads as a stack and takes the
     width of one card plus a sliver per copy. The top copy is the
     last in the DOM and paints over the rest, which is where the
     count badge and the pip live. */
  .row.strip .pile > [role="listitem"] {
    position: relative;
    top: calc(var(--i, 0) * -3px);
    transition: margin-left 140ms var(--ease);
  }
  .row.strip .pile > [role="listitem"] + [role="listitem"] {
    margin-left: calc(var(--card-w, 88px) * -1 + 4px);
  }
  /* Hovering a pile spreads it to the between-pile overlap so each
     copy can be told apart, hovered and clicked (a specific Forest
     for a fetch, say) — the same 40% of each card the strip already
     leaves visible between names. */
  .row.strip .pile.multi:hover > [role="listitem"] + [role="listitem"] {
    margin-left: calc(var(--card-w, 88px) * -0.6);
  }
  .pile-count {
    position: absolute;
    top: -6px;
    left: -6px;
    z-index: 8;
    min-width: 18px;
    height: 18px;
    padding: 0 5px;
    box-sizing: border-box;
    border-radius: 9px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    color: var(--gold);
    background: var(--surface, #0b0a09);
    border: 1px solid rgba(217, 180, 92, 0.6);
    pointer-events: none;
  }
</style>
