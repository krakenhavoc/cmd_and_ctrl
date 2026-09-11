<script lang="ts">
  // PlayerIdentity is the circular, identity-first rebuild of PlayerHeader.
  // It replaces the horizontal pill in the panel with an avatar disc as
  // the visual anchor: name above, life over the circle, markers +
  // ±-controls arranged around it. Same behaviors as the old header —
  // click-to-attack, click-to-cast-target, life/poison/energy steppers,
  // monarch/initiative toggles, damage floaters, mana pool pips.
  //
  // The avatar wrapper carries data-seat-id so CombatArrows can anchor
  // attack arrows on it (see CombatArrows.svelte:161).

  import type { ActionPayload, ActionType, PlayerView } from "../../protocol";
  import { seatColor } from "../../colors";
  import { floatUp, fadeOut } from "../../animations";
  import { play } from "../../sounds";
  import { avatarURL } from "../../api";
  import { scryfallImageURL } from "../../cardImage";
  import { targeting, isLegalPlayerTarget, isPicked } from "../../targeting";
  import ManaPoolPips from "./ManaPoolPips.svelte";
  import Icon from "../Icon.svelte";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seat: PlayerView;
    isSelf: boolean;
    isActive: boolean;
    hasPriority: boolean;
    attackTargetable: boolean;
    isMonarch: boolean;
    isInitiative: boolean;
    sendAction: ActionSender;
    onDeclareAttack?: (targetPlayerID: string) => void;
    onTargetPlayer?: (targetPlayerID: string) => void;
    // Scryfall id of the seat's commander (command zone or
    // battlefield), for the art-crop avatar when the seat has no
    // Discord avatar. Precedence: Discord → commander art → seat disc.
    commanderScryfallID?: string | null;
  }

  const {
    seat,
    isSelf,
    isActive,
    hasPriority,
    attackTargetable,
    isMonarch,
    isInitiative,
    sendAction,
    onDeclareAttack,
    onTargetPlayer,
    commanderScryfallID = null,
  }: Props = $props();

  const targetableByCast = $derived.by(() => {
    const t = $targeting;
    if (!t) return false;
    if (!isLegalPlayerTarget(t, seat.id)) return false;
    return !seat.eliminated;
  });
  // S20 sub-PR 5: already in a multi-target pick list.
  const pickedByCast = $derived.by(() => {
    const t = $targeting;
    return t !== null && isPicked(t, seat.id);
  });

  function handleAvatarClick(): void {
    if (targetableByCast) {
      onTargetPlayer?.(seat.id);
      return;
    }
    if (!attackTargetable) return;
    onDeclareAttack?.(seat.id);
  }

  const discordAvatar = $derived(avatarURL(seat.discord_id, seat.discord_avatar_hash));
  // The commander's art crop is always the FRONT face's — a
  // double-faced commander is identified by the side it is cast as,
  // and the seat header is an identity badge, not a board state.
  const commanderArt = $derived(
    commanderScryfallID ? scryfallImageURL(commanderScryfallID, "art_crop") : null,
  );
  // Discord avatar first, the commander's art crop when there is
  // none, the seat-colour disc when neither loads.
  let failedAvatarURL = $state<string | null>(null);
  const avatar = $derived.by(() => {
    if (discordAvatar && failedAvatarURL !== discordAvatar) return discordAvatar;
    if (commanderArt && failedAvatarURL !== commanderArt) return commanderArt;
    return null;
  });
  const displayLabel = $derived(seat.display_name ?? seat.name);

  function changeLife(delta: number): void {
    sendAction("change_life", { delta }, seat.id);
  }
  function changePoison(delta: number): void {
    sendAction("add_player_counter", { name: "poison", delta }, seat.id);
  }
  function changeEnergy(delta: number): void {
    sendAction("add_player_counter", { name: "energy", delta }, seat.id);
  }
  function toggleMonarch(): void {
    sendAction("set_monarch", undefined, isMonarch ? "" : seat.id);
  }
  function toggleInitiative(): void {
    sendAction("set_initiative", undefined, isInitiative ? "" : seat.id);
  }

  const POPUP_HOLD_MS = 1100;
  let popup = $state<{ id: string; delta: number } | null>(null);
  let popupTimer: ReturnType<typeof setTimeout> | null = null;
  let baselineCount = -1;
  $effect(() => {
    const count = seat.life_history?.length ?? 0;
    if (baselineCount === -1) {
      baselineCount = count;
      return;
    }
    if (count <= baselineCount) {
      baselineCount = count;
      return;
    }
    const last = seat.life_history[count - 1];
    if (!last || last.delta === 0) {
      baselineCount = count;
      return;
    }
    popup = { id: last.at, delta: last.delta };
    play(last.delta < 0 ? "damage" : "heal");
    if (popupTimer) clearTimeout(popupTimer);
    popupTimer = setTimeout(() => {
      popup = null;
      popupTimer = null;
    }, POPUP_HOLD_MS);
    baselineCount = count;
  });

  const interactive = $derived(attackTargetable || targetableByCast);
</script>

<div
  class="identity"
  class:self={isSelf}
  class:active={isActive}
  class:priority={hasPriority}
  class:targetable={attackTargetable}
  class:cast-targetable={targetableByCast}
  class:cast-picked={pickedByCast}
  class:eliminated={seat.eliminated}
  style:--seat-color={seatColor(seat.seat)}
>
  <span class="name" title={displayLabel}>{displayLabel}</span>

  <div class="core-row">
    <!-- Left: mana pool floats alongside the avatar instead of stacking
         below it. Empty when no mana is pooled, so the column
         collapses and the avatar stays visually centred. -->
    <div class="side left" aria-hidden={!seat.mana_pool || seat.mana_pool.length === 0}>
      <ManaPoolPips pool={seat.mana_pool} />
    </div>

    <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
    <div
      class="avatar-wrap"
      data-seat-id={seat.id}
      role={interactive ? "button" : "group"}
      tabindex={interactive ? 0 : undefined}
      onclick={handleAvatarClick}
      onkeydown={(e) => {
        if (interactive && (e.key === "Enter" || e.key === " ")) {
          e.preventDefault();
          handleAvatarClick();
        }
      }}
      aria-label={attackTargetable
        ? `attack ${displayLabel}`
        : `${displayLabel}, ${seat.life} life`}
    >
      {#if avatar}
        {#key avatar}
          <img
            class="avatar"
            src={avatar}
            alt=""
            aria-hidden="true"
            onerror={() => (failedAvatarURL = avatar)}
          />
        {/key}
      {:else}
        <span class="avatar seat-dot-fallback" aria-hidden="true"></span>
      {/if}

      <!-- Life overlay sits on the bottom arc of the avatar circle. On
           self we expose ± chips flanking the number; on opponents it
           reads as a static chip. -->
      {#if isSelf}
        <span class="life-chip">
          <button
            type="button"
            class="life-btn dec"
            title="-1 life"
            aria-label="lose 1 life"
            onclick={(e) => {
              e.stopPropagation();
              changeLife(-1);
            }}>−</button
          >
          <span class="life">{seat.life}</span>
          <button
            type="button"
            class="life-btn inc"
            title="+1 life"
            aria-label="gain 1 life"
            onclick={(e) => {
              e.stopPropagation();
              changeLife(1);
            }}>+</button
          >
        </span>
      {:else}
        <span class="life-chip readonly">
          <span class="life">{seat.life}</span>
        </span>
      {/if}

      {#if popup}
        {#key popup.id}
          <span
            class="dmg-popup"
            class:loss={popup.delta < 0}
            class:gain={popup.delta > 0}
            in:floatUp
            out:fadeOut
            aria-live="polite"
          >
            {popup.delta > 0 ? "+" : ""}{popup.delta}
          </span>
        {/key}
      {/if}
    </div>

    <!-- Right: monarch/initiative toggles + poison/energy steppers +
         any ad-hoc player counters. Stacked vertically so they don't
         push the panel taller; each row is the same height as a
         single chip. -->
    <div class="side right">
      {#if isSelf}
        <div class="crown-row">
          <button
            type="button"
            class="marker monarch"
            class:active={isMonarch}
            title={isMonarch ? "release the monarch" : "claim the monarch"}
            aria-label="toggle monarch"
            aria-pressed={isMonarch}
            onclick={(e) => {
              e.stopPropagation();
              toggleMonarch();
            }}><Icon name="crown" size={13} /></button
          >
          <button
            type="button"
            class="marker initiative"
            class:active={isInitiative}
            title={isInitiative ? "release the initiative" : "claim the initiative"}
            aria-label="toggle initiative"
            aria-pressed={isInitiative}
            onclick={(e) => {
              e.stopPropagation();
              toggleInitiative();
            }}><Icon name="sword" size={13} /></button
          >
        </div>
        <span class="counter-inline poison" title="poison counters">
          <span class="counter-icon" aria-hidden="true"><Icon name="drop" size={11} /></span>
          <button
            type="button"
            class="counter-btn"
            aria-label="lose 1 poison"
            onclick={(e) => {
              e.stopPropagation();
              changePoison(-1);
            }}>−</button
          >
          <span class="counter-val">{seat.poison ?? 0}</span>
          <button
            type="button"
            class="counter-btn"
            aria-label="gain 1 poison"
            onclick={(e) => {
              e.stopPropagation();
              changePoison(1);
            }}>+</button
          >
        </span>
        <span class="counter-inline energy" title="energy counters">
          <span class="counter-icon" aria-hidden="true"><Icon name="bolt" size={11} /></span>
          <button
            type="button"
            class="counter-btn"
            aria-label="lose 1 energy"
            onclick={(e) => {
              e.stopPropagation();
              changeEnergy(-1);
            }}>−</button
          >
          <span class="counter-val">{seat.energy ?? 0}</span>
          <button
            type="button"
            class="counter-btn"
            aria-label="gain 1 energy"
            onclick={(e) => {
              e.stopPropagation();
              changeEnergy(1);
            }}>+</button
          >
        </span>
      {:else}
        {#if isMonarch}
          <span class="marker monarch active" title="monarch" aria-label="monarch"
            ><Icon name="crown" size={12} /></span
          >
        {/if}
        {#if isInitiative}
          <span class="marker initiative active" title="initiative" aria-label="initiative"
            ><Icon name="sword" size={12} /></span
          >
        {/if}
        {#if (seat.poison ?? 0) > 0}
          <span class="marker poison" title={`${seat.poison} poison`} aria-label="poison">
            <Icon name="drop" size={11} />{seat.poison}
          </span>
        {/if}
        {#if (seat.energy ?? 0) > 0}
          <span class="marker energy" title={`${seat.energy} energy`} aria-label="energy">
            <Icon name="bolt" size={11} />{seat.energy}
          </span>
        {/if}
        {#if seat.counters}
          {#each Object.entries(seat.counters) as [name, count] (name)}
            {#if count > 0 && name !== "poison" && name !== "energy"}
              <span class="marker counter" title={`${count} ${name}`} aria-label={name}>
                {#if name === "experience"}<Icon
                    name="star"
                    size={11}
                  />{:else if name === "rad"}<Icon name="rad" size={11} />{:else}<Icon
                    name="dot"
                    size={11}
                  />{/if}{count}
              </span>
            {/if}
          {/each}
        {/if}
      {/if}
    </div>
  </div>

  {#if seat.eliminated}
    <span class="tag elim">eliminated</span>
  {/if}
</div>

<style>
  .identity {
    --avatar-size: calc(var(--avatar-size-base, 88px) * var(--card-scale, 1));
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    color: var(--fg);
    font-size: 12px;
    position: relative;
    min-width: 0;
    max-width: 100%;
  }
  .name {
    font-weight: 600;
    letter-spacing: 0.01em;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--fg);
    text-shadow: 0 1px 0 rgba(0, 0, 0, 0.4);
  }
  .avatar-wrap {
    position: relative;
    width: var(--avatar-size);
    height: var(--avatar-size);
    border-radius: 50%;
    display: grid;
    place-items: center;
    padding: 3px;
    background: var(--surface-raised);
    border: 2.5px solid color-mix(in srgb, var(--seat-color, #888) 55%, var(--surface));
    box-shadow: 0 8px 20px rgba(0, 0, 0, 0.5);
    transition:
      box-shadow 160ms var(--ease),
      border-color 160ms var(--ease);
    box-sizing: border-box;
  }
  .avatar {
    width: 100%;
    height: 100%;
    border-radius: 50%;
    object-fit: cover;
    display: block;
  }
  .seat-dot-fallback {
    background: var(--seat-color, #888);
  }

  /* State rings translate the old pill shadows into circle-shaped ones.
     Active = whose turn it is → seat-coloured ring.
     Priority = holds priority right now → gold ring (overrides active). */
  .identity.active .avatar-wrap {
    border-color: var(--seat-color, #5fb0ff);
    box-shadow:
      0 0 0 1px var(--seat-color, #5fb0ff),
      0 0 22px color-mix(in srgb, var(--seat-color, #5fb0ff) 45%, transparent),
      inset 0 1px 0 rgba(255, 255, 255, 0.08);
  }
  .identity.priority .avatar-wrap {
    border-color: var(--gold);
    box-shadow:
      0 0 0 2px var(--gold),
      0 0 26px rgba(255, 208, 122, 0.65),
      inset 0 1px 0 rgba(255, 255, 255, 0.08);
  }
  .identity.targetable .avatar-wrap {
    cursor: pointer;
    border-color: var(--danger);
    box-shadow:
      0 0 0 2px var(--danger),
      0 0 22px rgba(255, 122, 122, 0.55);
  }
  .identity.targetable .avatar-wrap:hover {
    box-shadow:
      0 0 0 3px var(--danger),
      0 0 28px rgba(255, 122, 122, 0.75);
  }
  .identity.cast-targetable .avatar-wrap {
    cursor: pointer;
    border-color: var(--gold);
    box-shadow:
      0 0 0 2px var(--gold),
      0 0 22px rgba(255, 208, 122, 0.55);
  }
  .identity.cast-targetable .avatar-wrap:hover {
    box-shadow:
      0 0 0 3px var(--gold),
      0 0 28px rgba(255, 208, 122, 0.75);
  }
  .identity.cast-picked .avatar-wrap {
    border-color: #ffe69a;
    box-shadow:
      0 0 0 3px #ffe69a,
      0 0 28px rgba(255, 230, 154, 0.8);
  }
  .identity.eliminated {
    opacity: 0.5;
    filter: grayscale(0.6);
  }

  /* Life chip overlays the bottom third of the avatar circle. Ellipse
     backdrop sits flush with the circle's lower arc so the number reads
     as part of the identity disc rather than a floating badge. */
  .life-chip {
    position: absolute;
    left: 50%;
    bottom: -6px;
    transform: translateX(-50%);
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 2px 6px;
    border-radius: 999px;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0) 60%),
      rgba(6, 10, 22, 0.92);
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-shadow: 0 3px 10px rgba(0, 0, 0, 0.5);
  }
  .life-chip.readonly {
    padding: 2px 10px;
  }
  .life {
    font-weight: 800;
    font-size: 15px;
    min-width: 22px;
    text-align: center;
    font-variant-numeric: tabular-nums;
    color: var(--fg);
    letter-spacing: -0.02em;
    text-shadow: 0 1px 0 rgba(0, 0, 0, 0.5);
  }
  .life-btn {
    width: 18px;
    height: 18px;
    padding: 0;
    border-radius: 50%;
    border: 1px solid rgba(255, 255, 255, 0.1);
    background: rgba(0, 0, 0, 0.3);
    color: var(--fg);
    font-size: 13px;
    line-height: 1;
    cursor: pointer;
    font-family: inherit;
    box-shadow: none;
    transition:
      border-color 120ms var(--ease),
      background 120ms var(--ease),
      color 120ms var(--ease);
  }
  .life-btn.dec:hover {
    background: rgba(255, 122, 122, 0.18);
    border-color: rgba(255, 122, 122, 0.5);
    color: var(--danger);
  }
  .life-btn.inc:hover {
    background: rgba(122, 255, 154, 0.18);
    border-color: rgba(122, 255, 154, 0.5);
    color: var(--mint);
  }

  /* Damage popup floats up over the circle's top arc; heal/gain
     intentionally also floats up (same pattern as the old header). */
  .dmg-popup {
    position: absolute;
    left: 50%;
    top: -14px;
    transform: translateX(-50%);
    font-size: 24px;
    font-weight: 800;
    pointer-events: none;
    z-index: 20;
    text-shadow:
      0 1px 0 rgba(0, 0, 0, 0.85),
      0 0 8px rgba(0, 0, 0, 0.6);
  }
  .dmg-popup.loss {
    color: #ff7a7a;
  }
  .dmg-popup.gain {
    color: #7aff9a;
  }

  /* Horizontal arrangement: mana (left column) — avatar (center) —
     counters (right column). The side columns stack vertically so a
     full marker set (crown + sword + poison + energy) is roughly the
     same height as the avatar, keeping the panel compact. */
  .core-row {
    /* In the rail everything stacks: avatar, then the floating mana
       pool, then the marker chips as a wrapping row. */
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    max-width: 100%;
  }
  .side {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: center;
    gap: 4px;
    flex: 0 0 auto;
    min-width: 0;
    max-width: 100%;
  }
  .side.left {
    order: 1;
  }
  .side.left[aria-hidden="true"] {
    display: none;
  }
  .side.right {
    order: 2;
    margin-top: 6px;
  }
  /* Crown + sword sit on a single row so monarch/initiative read as
     peer toggles rather than a tall stack. */
  .crown-row {
    display: flex;
    gap: 4px;
  }
  .marker {
    font-size: 12px;
    line-height: 1;
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 2px 6px;
    border-radius: 999px;
    background: rgba(0, 0, 0, 0.3);
    color: #c8c8c8;
    border: 1px solid rgba(255, 255, 255, 0.05);
    cursor: default;
    backdrop-filter: blur(4px);
  }
  button.marker {
    font-family: inherit;
    cursor: pointer;
    box-shadow: none;
    transition:
      transform 120ms var(--ease),
      border-color 120ms var(--ease),
      background 120ms var(--ease);
  }
  button.marker:hover {
    border-color: rgba(255, 255, 255, 0.2);
    background: rgba(0, 0, 0, 0.55);
  }
  button.marker:active {
    transform: scale(0.95);
  }
  .marker.active.monarch {
    color: var(--gold);
    border-color: rgba(255, 208, 122, 0.7);
    background: var(--gold-soft);
    box-shadow: 0 0 10px rgba(255, 208, 122, 0.35);
  }
  .marker.active.initiative {
    color: #b08aff;
    border-color: rgba(176, 138, 255, 0.7);
    background: rgba(176, 138, 255, 0.18);
    box-shadow: 0 0 10px rgba(176, 138, 255, 0.35);
  }
  .marker.poison {
    color: var(--mint);
  }
  .marker.energy {
    color: var(--gold);
  }
  .marker.counter {
    color: #b8c8e8;
  }
  .counter-inline {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 2px 6px;
    border-radius: 999px;
    background: rgba(0, 0, 0, 0.3);
    border: 1px solid rgba(255, 255, 255, 0.06);
  }
  .counter-icon {
    font-size: 11px;
    line-height: 1;
    margin-right: 2px;
  }
  .counter-val {
    min-width: 14px;
    text-align: center;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    color: var(--fg);
  }
  .counter-inline.poison .counter-val {
    color: var(--mint);
  }
  .counter-inline.energy .counter-val {
    color: var(--gold);
  }
  .counter-btn {
    width: 14px;
    height: 14px;
    padding: 0;
    border: 0;
    background: transparent;
    color: var(--fg-dim);
    font-size: 11px;
    line-height: 1;
    cursor: pointer;
    font-family: inherit;
    border-radius: 50%;
    box-shadow: none;
    transition:
      background 100ms var(--ease),
      color 100ms var(--ease);
  }
  .counter-btn:hover {
    background: rgba(255, 255, 255, 0.08);
    color: var(--fg);
  }

  .tag {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    padding: 2px 7px;
    border-radius: 999px;
    background: rgba(0, 0, 0, 0.35);
    border: 1px solid rgba(255, 255, 255, 0.06);
    font-weight: 700;
    margin-top: 6px;
  }
  .tag.elim {
    color: var(--danger);
    border-color: rgba(255, 122, 122, 0.4);
  }
</style>
