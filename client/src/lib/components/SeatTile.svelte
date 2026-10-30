<script lang="ts">
  // SeatTile — one seat of a table in the lobby, drawn as a character
  // select (ADR 0212 §1): the seat's colour along the top edge and as a
  // faint wash, a "Seat N" badge, an art slot, then who sits there, what
  // they play and whether they are ready. The art slot draws the
  // commander's crop where the page has one: a bot's pre-built deck (ADR
  // 0212 phase B1), and your own seat while you browse the picker, art and
  // all, marked Preview and "Not seated yet" (§1), since nothing is sent
  // until you sit down. Without a crop, or until a person's seat carries
  // its commanders (B2), it is the seat's wash.
  //
  //   - A taken seat: the avatar, the name, the host / you tags (a bot's
  //     dashed chip with its tier; an agent's AgentChip), the deck, then a
  //     status row of identity pips and power on the left and Ready, a
  //     check on a mint disc and the word, on the right. Ready is --mint,
  //     never --ready (which means "a legal action now" on the board).
  //   - Your seat is a button that opens the deck picker, dashed in your
  //     colour while it has no deck.
  //   - An open seat is dashed, with Invite and Add bot where you may.
  //
  // The × (remove a seat, leave your own) stays on the tile with its
  // confirm, which the page owns.
  //
  // While the table's ready-check moment plays (ADR 0212 §2, `moment`),
  // a ready tile's check pops in seat order and the tile whose deck just
  // landed gets one brief ring. The page sets `moment` only under full
  // motion; a still moment draws nothing here.
  import { avatarURL, type SeatInfo } from "../api";
  import { seatColor } from "../colors";
  import type { DeckArt } from "../deckGrid";
  import { L } from "../labels";
  import { readyDelayMs } from "../readyMoment";
  import { seatName, type SeatTileText } from "../tablesPage";
  import AgentChip from "./AgentChip.svelte";
  import CommanderArt from "./decks/CommanderArt.svelte";
  import DeckPowerBadge from "./DeckPowerBadge.svelte";
  import Icon from "./Icon.svelte";
  import IdentityPips from "./IdentityPips.svelte";

  interface Props {
    /** The absolute seat number, 0-based; the badge counts from 1. */
    seat: number;
    /** Who sits here, or null for an open seat. */
    p: SeatInfo | null;
    /** What a taken seat's tile says (tablesPage.seatTileText). */
    text?: SeatTileText;
    /** The commanders' art crops, partners side by side; the seat's wash without. */
    art?: DeckArt[];
    /** Your seat shows the deck selected in the picker, not seated yet. */
    preview?: boolean;
    /** The seat this session holds. */
    you?: boolean;
    /** A bot seat's chip: its tier, and "(own key)". */
    botChip?: string;
    /** The ready-check moment is playing: `landed` is the seat whose deck came in last. */
    moment?: { landed: boolean };
    /** Your seat: the tile opens the deck picker under the seats. */
    pick?: { expanded: boolean; controls: string; onclick: () => void };
    /** The ×, when this person may remove the seat or leave it. */
    remove?: { label: string; title: string; disabled: boolean; onclick: () => void };
    /** An open seat's Invite, which opens the table's Invite menu. */
    oninvite?: (opener: HTMLElement) => void;
    inviteExpanded?: boolean;
    /** An open seat's Add bot. */
    addBot?: { disabled: boolean; onclick: () => void };
  }

  const {
    seat,
    p,
    text,
    art = [],
    preview = false,
    you = false,
    botChip = "",
    moment,
    pick,
    remove,
    oninvite,
    inviteExpanded = false,
    addBot,
  }: Props = $props();

  const ready = $derived(text?.ready === true);
  const avatar = $derived(p ? avatarURL(p.discord_id, p.discord_avatar_hash) : null);
  const power = $derived(p && !p.is_bot && p.deck_uploaded ? (p.deck_power ?? 0) : 0);
  const initial = $derived(p ? seatName(p).trim().slice(0, 1).toUpperCase() || "?" : "");

  // Crops that failed to load; when every one has, the slot is the wash.
  let failed = $state<string[]>([]);
  const crops = $derived(art.filter((a) => !failed.includes(a.scryfallId)));
</script>

<li
  class="seat tile"
  class:you
  class:open={p === null}
  class:bot={p?.is_bot === true}
  class:empty={you && p !== null && !ready && !preview}
  class:has-x={remove !== undefined}
  class:moment={moment !== undefined}
  class:landed={moment?.landed === true}
  style:--seat={seatColor(seat)}
  style:--wave="{readyDelayMs(seat)}ms"
  aria-label={L.seatTile(seat + 1, p ? seatName(p) : "open")}
>
  {#if p && pick}
    <button
      type="button"
      class="face pick"
      aria-label={ready ? "change your deck" : "choose a deck for your seat"}
      aria-expanded={pick.expanded}
      aria-controls={pick.controls}
      title={ready ? "change the deck in your seat" : "choose a deck for your seat"}
      onclick={pick.onclick}
    >
      {@render slot()}
      {@render info()}
    </button>
  {:else}
    <div class="face">
      {@render slot()}
      {@render info()}
    </div>
  {/if}
  {#if remove}
    <button
      type="button"
      class="seat-x"
      title={remove.title}
      aria-label={remove.label}
      disabled={remove.disabled}
      onclick={remove.onclick}
    >
      <Icon name="x" size={12} />
    </button>
  {/if}
</li>

<!-- Spans only: on your seat these sit inside a button. -->
{#snippet slot()}
  <span class="art" class:crop={crops.length > 0}>
    {#if crops.length > 0}
      <span class="crops">
        {#each crops as a (a.scryfallId)}
          <CommanderArt
            scryfallId={a.scryfallId}
            name={a.name}
            onfail={() => (failed = [...failed, a.scryfallId])}
          />
        {/each}
      </span>
    {:else if p === null}
      <span class="ghost-av" aria-hidden="true"></span>
    {:else if you && !ready && !preview}
      <span class="bigplus" aria-hidden="true"><Icon name="plus" size={20} /></span>
    {/if}
    <span class="seatno"><i aria-hidden="true"></i>Seat {seat + 1}</span>
    {#if preview}<span class="preview-mark">Preview</span>{/if}
  </span>
{/snippet}

{#snippet info()}
  <span class="sinfo">
    {#if p}
      <span class="who">
        <span class="av" class:bot={p.is_bot}>
          {#if p.is_bot}
            <Icon name="robot" size={15} />
          {:else if avatar}
            <img src={avatar} alt="" />
          {:else}
            {initial}
          {/if}
        </span>
        <span class="name">{seatName(p)}</span>
        {#if p.is_bot && botChip}<span class="stag botchip">{botChip}</span>{/if}
        {#if p.is_host}<span
            class="stag"
            title="Table host: may manage the table alongside the admin">host</span
          >{/if}
        {#if you}<span class="stag">you</span>{/if}
      </span>
      {#if p.is_agent && !p.is_bot}
        <span class="agent"><AgentChip seat={p} /></span>
      {/if}
      <span class="what" class:muted={!ready && !preview}>{text?.what ?? ""}</span>
      {#if text?.second}<span class="second">{text.second}</span>{/if}
      <span class="srow">
        <span class="ident"><IdentityPips colors={text?.colors} /><DeckPowerBadge {power} /></span>
        {#if preview}
          <span class="pending">Not seated yet</span>
        {:else if ready}
          <span class="ready"
            ><span class="ck" aria-hidden="true"
              ><Icon name="check" size={12} strokeWidth={2.6} /></span
            >Ready</span
          >
        {:else}
          <span class="pending" class:lead={you}
            >{you ? "Choose one to sit down" : "Choosing a deck…"}</span
          >
        {/if}
      </span>
    {:else}
      <span class="who"><span class="name dim">Open seat</span></span>
      {#if oninvite || addBot}
        <span class="openacts">
          {#if oninvite}
            <button
              type="button"
              class="btn sm xs"
              aria-haspopup="menu"
              aria-expanded={inviteExpanded}
              onclick={(e) => oninvite(e.currentTarget)}
            >
              <Icon name="link" size={13} />
              {L.inviteButton}
            </button>
          {/if}
          {#if addBot}
            <button
              type="button"
              class="btn ghost sm xs"
              disabled={addBot.disabled}
              onclick={addBot.onclick}
            >
              <Icon name="robot" size={13} />
              {L.addBot}
            </button>
          {/if}
        </span>
      {:else}
        <span class="second">Waiting for a player</span>
      {/if}
    {/if}
  </span>
{/snippet}

<style>
  .tile {
    --seat: var(--fg-dim);
    position: relative;
    min-width: 0;
    display: flex;
    flex-direction: column;
    border-radius: var(--radius-lg);
    background: color-mix(in srgb, var(--seat) 9%, var(--surface-sunken));
    box-shadow:
      inset 0 3px 0 var(--seat),
      inset 0 0 0 1px color-mix(in srgb, var(--seat) 30%, transparent);
    overflow: hidden;
  }
  /* Your seat with no deck yet: the tile is the call to action. */
  .tile.empty {
    background: color-mix(in srgb, var(--seat) 6%, var(--surface-sunken));
    box-shadow: inset 0 3px 0 var(--seat);
    outline: 1.5px dashed color-mix(in srgb, var(--seat) 55%, transparent);
    outline-offset: -1.5px;
  }
  .tile.open {
    background: transparent;
    box-shadow: none;
    outline: 1.5px dashed color-mix(in srgb, var(--seat) 32%, var(--border-strong));
    outline-offset: -1.5px;
  }
  .face {
    display: flex;
    flex-direction: column;
    flex: 1 1 auto;
    min-width: 0;
  }
  /* Undo the site's button (app.css): the face is the tile. */
  button.face {
    align-items: stretch;
    justify-content: flex-start;
    gap: 0;
    width: 100%;
    margin: 0;
    padding: 0;
    border: 0;
    border-radius: inherit;
    background: transparent;
    color: inherit;
    font: inherit;
    letter-spacing: normal;
    line-height: inherit;
    text-align: left;
    cursor: pointer;
  }
  button.face:hover:not(:disabled) {
    background: var(--overlay-faint);
    border-color: transparent;
  }
  button.face:active {
    transform: none;
  }
  button.face:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: -2px;
  }

  /* The art slot: the seat's wash until a commander's crop fills it. */
  .art {
    position: relative;
    display: grid;
    place-items: center;
    aspect-ratio: 16 / 9;
    background:
      radial-gradient(
        110% 90% at 85% 0%,
        color-mix(in srgb, var(--seat) 24%, transparent),
        transparent 72%
      ),
      color-mix(in srgb, var(--seat) 7%, transparent);
  }
  .tile.open .art,
  .tile.empty .art {
    background: color-mix(in srgb, var(--seat) 6%, transparent);
  }
  /* A commander's crop (partners side by side), faded into the tile at
     its foot, as the Decks grid does. */
  .crops {
    position: absolute;
    inset: 0;
    display: flex;
    gap: 2px;
    overflow: hidden;
  }
  .art.crop::after {
    content: "";
    position: absolute;
    inset: 0;
    background: linear-gradient(
      to bottom,
      transparent 52%,
      color-mix(in srgb, var(--seat) 9%, var(--surface-sunken))
    );
    pointer-events: none;
  }
  /* The × sits in the top-right corner, so the artist's credit moves
     under it. */
  .tile.has-x .crops :global(.art-credit) {
    top: 34px;
  }
  .preview-mark {
    position: absolute;
    left: 9px;
    bottom: 9px;
    z-index: 1;
    padding: 3px 7px;
    border-radius: 999px;
    border: 1px solid var(--accent-line);
    background: color-mix(in srgb, var(--bg) 66%, transparent);
    color: var(--accent-strong);
    font: 700 9.5px/1 var(--font-mono);
    letter-spacing: 0.1em;
    text-transform: uppercase;
  }
  .seatno {
    z-index: 1;
    position: absolute;
    top: 9px;
    left: 9px;
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 3px 7px;
    border-radius: 999px;
    background: color-mix(in srgb, var(--bg) 66%, transparent);
    color: var(--fg);
    font: 700 9.5px/1 var(--font-mono);
    letter-spacing: 0.1em;
    text-transform: uppercase;
  }
  .seatno i {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--seat);
  }
  .ghost-av {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    border: 1.5px dashed color-mix(in srgb, var(--seat) 45%, transparent);
  }
  .bigplus {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    border: 1.5px dashed color-mix(in srgb, var(--seat) 70%, transparent);
    color: var(--seat);
  }

  .sinfo {
    display: flex;
    flex-direction: column;
    gap: 5px;
    min-width: 0;
    padding: 8px 12px 12px;
  }
  .who {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }
  .av {
    display: grid;
    place-items: center;
    flex: none;
    width: 26px;
    height: 26px;
    box-sizing: border-box;
    border-radius: 50%;
    border: 2px solid var(--seat);
    background: var(--seat);
    /* Dark ink on the bright seat colour, which no skin changes. */
    color: color-mix(in srgb, var(--seat) 15%, black);
    font: 700 11.5px/1 var(--font-ui);
    overflow: hidden;
  }
  .av img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .av.bot {
    background: transparent;
    border-style: dashed;
    color: var(--seat);
  }
  .name {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 14px;
    font-weight: 700;
    color: var(--fg);
  }
  .name.dim {
    color: var(--fg-muted);
    font-weight: 600;
  }
  .stag {
    flex: none;
    padding: 3px 6px;
    border-radius: 999px;
    background: color-mix(in srgb, var(--seat) 22%, transparent);
    color: var(--fg);
    font: 700 9px/1 var(--font-mono);
    letter-spacing: 0.1em;
    text-transform: uppercase;
  }
  .stag.botchip {
    background: transparent;
    border: 1px dashed color-mix(in srgb, var(--seat) 70%, transparent);
    padding: 2px 6px;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 0 1 auto;
  }
  .agent {
    display: flex;
    min-width: 0;
  }
  .what {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    font-size: 13px;
    font-weight: 600;
    line-height: 1.3;
    color: var(--fg);
  }
  .what.muted {
    color: var(--fg-muted);
  }
  .second {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 12px;
    color: var(--fg-muted);
  }
  .srow {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 6px 8px;
    margin-top: auto;
    padding-top: 2px;
  }
  .ident {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .ready {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    margin-left: auto;
    color: var(--fg);
    font: 700 10.5px/1 var(--font-mono);
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }
  /* Mint, success: always with the word, since seat 4's green is the
     same hue. */
  .ck {
    display: grid;
    place-items: center;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--mint);
    color: color-mix(in srgb, var(--mint) 15%, black);
  }
  .pending {
    margin-left: auto;
    font-size: 12px;
    color: var(--fg-muted);
  }
  /* Your own seat's call to action reads first, on the left. */
  .pending.lead {
    order: -1;
    margin-left: 0;
  }
  .openacts {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .btn.xs {
    height: 26px;
    padding: 0 9px;
    gap: 5px;
    font-size: 12px;
  }

  /* The ready-check moment (ADR 0212 §2): the check pops in seat order and
     the tile whose deck landed last gets one brief ring in mint, after its
     own check. The ring is outside the tile (a tile's own box-shadow is not
     clipped by its overflow); the keyframes repeat the resting shadow so
     the tile keeps its edge. The last frame is the resting look. */
  .tile.moment .ck {
    animation: ck-land 480ms var(--ease) both;
    animation-delay: var(--wave, 0ms);
  }
  .tile.landed {
    animation: tile-ring 1500ms var(--ease) both;
    animation-delay: var(--wave, 0ms);
  }
  @keyframes ck-land {
    from {
      transform: scale(0.4);
      opacity: 0.2;
    }
    60% {
      transform: scale(1.18);
      opacity: 1;
    }
    to {
      transform: scale(1);
      opacity: 1;
    }
  }
  @keyframes tile-ring {
    0% {
      box-shadow:
        inset 0 3px 0 var(--seat),
        inset 0 0 0 1px color-mix(in srgb, var(--seat) 30%, transparent),
        0 0 0 0 color-mix(in srgb, var(--mint) 0%, transparent);
    }
    25% {
      box-shadow:
        inset 0 3px 0 var(--seat),
        inset 0 0 0 1px color-mix(in srgb, var(--seat) 30%, transparent),
        0 0 0 3px color-mix(in srgb, var(--mint) 60%, transparent);
    }
    100% {
      box-shadow:
        inset 0 3px 0 var(--seat),
        inset 0 0 0 1px color-mix(in srgb, var(--seat) 30%, transparent),
        0 0 0 6px color-mix(in srgb, var(--mint) 0%, transparent);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .tile.moment .ck,
    .tile.landed {
      animation: none;
    }
  }

  .seat-x {
    position: absolute;
    top: 7px;
    right: 7px;
    display: grid;
    place-items: center;
    width: 22px;
    height: 22px;
    padding: 0;
    border-radius: var(--radius);
    border: 1px solid var(--border-strong);
    background: color-mix(in srgb, var(--bg) 66%, transparent);
    color: var(--fg-muted);
    cursor: pointer;
  }
  .seat-x:hover:not(:disabled) {
    color: var(--fg);
    border-color: var(--accent);
  }
  .seat-x:disabled {
    opacity: 0.5;
    cursor: default;
  }
</style>
