<script lang="ts">
  // PlayerHeader is the bar above each PlayerPanel. It shows who the
  // player is, their life, marker badges (monarch / initiative / poison
  // / energy), and various live state (active turn, priority, eliminated,
  // attack-targetable in combat). The viewer's own header gets +/- life
  // controls and inline poison/energy steppers; opponents read-only.
  //
  // Marker badges are click-targets: the monarch / initiative crowns are
  // claim/release toggles for the viewer (the viewer can crown themselves
  // or — when they already hold it — release the marker). Admin sessions
  // can target any seat, but the viewer-can-only-flip-self path covers
  // the casual "I dealt combat damage so I'm the monarch now" flow
  // without needing an opponent-targeting picker.

  import type { ActionPayload, PlayerView } from "../../protocol";
  import { seatColor } from "../../colors";
  import { floatUp, fadeOut } from "../../animations";
  import { play } from "../../sounds";
  import { avatarURL } from "../../api";
  import { targeting, isTargetingPlayer } from "../../targeting";
  import ManaPoolPips from "./ManaPoolPips.svelte";

  type ActionSender = (type: string, params?: ActionPayload["params"], player?: string) => void;

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
  }: Props = $props();

  // S14: when a cast-targeting prompt is active and the mode accepts
  // a player, light up this header and route clicks to the picker
  // callback instead of the combat path. Using the store directly so
  // every seat reacts to the same targeting state without needing
  // the parent to re-pass a prop per seat.
  const targetableByCast = $derived.by(() => {
    const t = $targeting;
    if (!t) return false;
    if (!isTargetingPlayer(t.mode)) return false;
    return !seat.eliminated;
  });

  function handleHeaderClick(): void {
    if (targetableByCast) {
      onTargetPlayer?.(seat.id);
      return;
    }
    if (!attackTargetable) return;
    onDeclareAttack?.(seat.id);
  }

  // Discord identity (S12.5). avatar resolves to /avatars/{id}/{hash}.png
  // when both fields are present; null falls back to the seat-color
  // dot. displayLabel prefers the OAuth-provided global_name over
  // the bare lobby name so a friend who signed in as "Alice" doesn't
  // see "Alice.1234" on their seat.
  const avatar = $derived(avatarURL(seat.discord_id, seat.discord_avatar_hash));
  const displayLabel = $derived(seat.display_name ?? seat.name);
  // failedAvatarURL latches a single URL that errored, so a transient
  // server failure for one specific (id, hash) doesn't degrade
  // to the dot permanently. When the user changes their avatar
  // (new hash → new URL), or the server starts succeeding again,
  // a re-render with a different avatar URL clears the latch and
  // lets the <img> retry. Stored as a string instead of a boolean
  // for exactly that reason — boolean would stick across URL
  // changes since avatarFailed has no input to reset against.
  let failedAvatarURL = $state<string | null>(null);
  const avatarFailed = $derived(avatar !== null && failedAvatarURL === avatar);

  // Life / poison / energy adjusters. Self only — these all go through
  // the player-scoped guard server-side, so a non-admin viewer trying
  // to mutate an opponent would get a clean error.
  function changeLife(delta: number): void {
    sendAction("change_life", { delta }, seat.id);
  }
  function changePoison(delta: number): void {
    // S13.2: route through add_player_counter so the SBA loop fires
    // (poison ≥ 10 is a game-loss). The server keeps Player.Poison
    // in sync with the unified Counters["poison"] map.
    sendAction("add_player_counter", { name: "poison", delta }, seat.id);
  }
  function changeEnergy(delta: number): void {
    sendAction("add_player_counter", { name: "energy", delta }, seat.id);
  }
  function toggleMonarch(): void {
    // Empty player clears; passing this seat's id sets it.
    sendAction("set_monarch", undefined, isMonarch ? "" : seat.id);
  }
  function toggleInitiative(): void {
    sendAction("set_initiative", undefined, isInitiative ? "" : seat.id);
  }

  // Damage / heal popup driven by the seat's life_history. Same pattern
  // as the prior implementation: snapshot the count on first render to
  // skip historical deltas, then react only to entries that arrive after.
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
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
  class="header"
  class:self={isSelf}
  class:active={isActive}
  class:priority={hasPriority}
  class:targetable={attackTargetable}
  class:cast-targetable={targetableByCast}
  class:eliminated={seat.eliminated}
  style:--seat-color={seatColor(seat.seat)}
  data-seat-id={seat.id}
  role={attackTargetable || targetableByCast ? "button" : "group"}
  tabindex={attackTargetable || targetableByCast ? 0 : undefined}
  onclick={handleHeaderClick}
  onkeydown={(e) => {
    if ((attackTargetable || targetableByCast) && (e.key === "Enter" || e.key === " ")) {
      e.preventDefault();
      handleHeaderClick();
    }
  }}
  aria-label={attackTargetable ? `attack ${displayLabel}` : `${displayLabel}, ${seat.life} life`}
>
  {#if avatar && !avatarFailed}
    <img
      class="avatar"
      src={avatar}
      alt=""
      aria-hidden="true"
      onerror={() => (failedAvatarURL = avatar)}
    />
  {:else}
    <span class="seat-dot" aria-hidden="true"></span>
  {/if}
  <span class="name">{displayLabel}</span>

  <!-- Marker badges. Monarch and initiative are claim toggles for self
       (clear when already-held, set otherwise). Poison / energy show
       inline ± steppers for self so the controls live next to the
       value; opponents see read-only badges and only when non-zero
       (the bar gets crowded fast in 4-player). -->
  <span class="markers">
    {#if isSelf}
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
        }}>👑</button
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
        }}>⚔</button
      >
      <span class="counter-inline poison" title="poison counters">
        <span class="counter-icon" aria-hidden="true">🟢</span>
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
        <span class="counter-icon" aria-hidden="true">⚡</span>
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
        <span class="marker monarch active" title="monarch" aria-label="monarch">👑</span>
      {/if}
      {#if isInitiative}
        <span class="marker initiative active" title="initiative" aria-label="initiative">⚔</span>
      {/if}
      {#if (seat.poison ?? 0) > 0}
        <span class="marker poison" title={`${seat.poison} poison`} aria-label="poison">
          🟢{seat.poison}
        </span>
      {/if}
      {#if (seat.energy ?? 0) > 0}
        <span class="marker energy" title={`${seat.energy} energy`} aria-label="energy">
          ⚡{seat.energy}
        </span>
      {/if}
      <!-- S13.2: surface non-zero player counters that aren't
           already shown via the legacy poison / energy markers.
           experience, rad, and any homebrew names land here. -->
      {#if seat.counters}
        {#each Object.entries(seat.counters) as [name, count] (name)}
          {#if count > 0 && name !== "poison" && name !== "energy"}
            <span class="marker counter" title={`${count} ${name}`} aria-label={name}>
              {name === "experience" ? "⭐" : name === "rad" ? "☢" : "•"}{count}
            </span>
          {/if}
        {/each}
      {/if}
    {/if}
  </span>

  <!-- Life with +/- controls on self. Click life value also pulses; the
       buttons stop event propagation so they don't trip the
       attackTargetable header click. -->
  {#if isSelf}
    <span class="life-controls">
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
    <span class="life">{seat.life}</span>
  {/if}

  {#if seat.eliminated}
    <span class="tag elim">eliminated</span>
  {:else if hasPriority}
    <span class="tag prio">priority</span>
  {:else if isActive}
    <span class="tag act">active</span>
  {/if}

  <ManaPoolPips pool={seat.mana_pool} />

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

<style>
  .header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 12px;
    border-radius: 999px;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.06) 0%, rgba(255, 255, 255, 0) 60%),
      color-mix(in srgb, var(--seat-color, #888) 22%, #0f1829);
    border: 1px solid color-mix(in srgb, var(--seat-color, #888) 38%, #2e3a55);
    color: var(--fg);
    font-size: 12px;
    height: 28px;
    box-sizing: border-box;
    position: relative;
    box-shadow:
      0 2px 8px rgba(0, 0, 0, 0.3),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
    transition:
      box-shadow 160ms var(--ease),
      background 160ms var(--ease),
      border-color 160ms var(--ease);
  }
  .dmg-popup {
    position: absolute;
    right: 12px;
    top: 100%;
    margin-top: 4px;
    font-size: 22px;
    font-weight: 800;
    pointer-events: none;
    z-index: 20;
    text-shadow:
      0 1px 0 rgba(0, 0, 0, 0.8),
      0 0 8px rgba(0, 0, 0, 0.55);
  }
  .dmg-popup.loss {
    color: #ff7a7a;
  }
  .dmg-popup.gain {
    color: #7aff9a;
  }
  .header.self {
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0) 60%),
      color-mix(in srgb, var(--seat-color, #5fb0ff) 32%, #0f1829);
  }
  .header.active {
    border-color: var(--seat-color, #5fb0ff);
    box-shadow:
      0 0 0 1px var(--seat-color, #5fb0ff),
      0 0 16px color-mix(in srgb, var(--seat-color, #5fb0ff) 40%, transparent),
      inset 0 1px 0 rgba(255, 255, 255, 0.08);
  }
  .header.priority {
    box-shadow:
      0 0 0 2px color-mix(in srgb, var(--seat-color, #5fb0ff) 70%, transparent),
      0 0 18px color-mix(in srgb, var(--seat-color, #5fb0ff) 55%, transparent),
      inset 0 1px 0 rgba(255, 255, 255, 0.08);
  }
  .header.targetable {
    cursor: pointer;
    box-shadow:
      0 0 0 2px var(--danger),
      0 0 18px rgba(255, 122, 122, 0.55),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
  }
  .header.targetable:hover {
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0) 60%),
      color-mix(in srgb, var(--danger) 28%, #0f1829);
  }
  .header.cast-targetable {
    cursor: pointer;
    box-shadow:
      0 0 0 2px var(--gold),
      0 0 18px rgba(255, 208, 122, 0.55),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
  }
  .header.cast-targetable:hover {
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0) 60%),
      color-mix(in srgb, var(--gold) 28%, #0f1829);
  }
  .header.eliminated {
    opacity: 0.5;
    filter: grayscale(0.6);
  }
  .seat-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--seat-color, #888);
    flex: 0 0 auto;
  }
  .avatar {
    /* Header is 28px tall with 4px padding = 20px content. Avatar
       is 18px + 1px ring so it sits cleanly inside without being
       clipped on the top/bottom. */
    width: 20px;
    height: 20px;
    border-radius: 50%;
    object-fit: cover;
    flex: 0 0 auto;
    box-shadow:
      0 0 0 1px rgba(0, 0, 0, 0.4),
      0 0 0 2px var(--seat-color, #888);
  }
  .name {
    font-weight: 600;
    letter-spacing: 0.01em;
    flex: 0 1 auto;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--fg);
    text-shadow: 0 1px 0 rgba(0, 0, 0, 0.4);
  }
  .markers {
    display: flex;
    gap: 4px;
    align-items: center;
    flex: 0 0 auto;
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
  .life-controls {
    margin-left: auto;
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 4px;
    background: rgba(0, 0, 0, 0.35);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 999px;
  }
  .life {
    font-weight: 800;
    font-size: 15px;
    min-width: 24px;
    text-align: center;
    margin-left: auto;
    font-variant-numeric: tabular-nums;
    color: var(--fg);
    text-shadow: 0 1px 0 rgba(0, 0, 0, 0.5);
    letter-spacing: -0.02em;
  }
  .life-controls .life {
    margin-left: 0;
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
  .tag {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    padding: 2px 7px;
    border-radius: 999px;
    background: rgba(0, 0, 0, 0.35);
    border: 1px solid rgba(255, 255, 255, 0.06);
    font-weight: 700;
  }
  .tag.prio {
    color: var(--gold);
    border-color: rgba(255, 208, 122, 0.4);
    box-shadow: 0 0 8px rgba(255, 208, 122, 0.3);
  }
  .tag.act {
    color: #9ec7ff;
    border-color: rgba(158, 199, 255, 0.4);
  }
  .tag.elim {
    color: var(--danger);
    border-color: rgba(255, 122, 122, 0.4);
  }
  /* Inline poison / energy stepper, lives in the header marker row.
     Shape mirrors the life ± controls so the bar reads as one
     coherent strip of value-with-controls widgets. */
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
</style>
