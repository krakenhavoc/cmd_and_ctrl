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
  }: Props = $props();

  function handleHeaderClick(): void {
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
    sendAction("set_poison", { amount: Math.max(0, (seat.poison ?? 0) + delta) }, seat.id);
  }
  function changeEnergy(delta: number): void {
    sendAction("set_energy", { amount: Math.max(0, (seat.energy ?? 0) + delta) }, seat.id);
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
  class:eliminated={seat.eliminated}
  style:--seat-color={seatColor(seat.seat)}
  data-seat-id={seat.id}
  role={attackTargetable ? "button" : "group"}
  tabindex={attackTargetable ? 0 : undefined}
  onclick={handleHeaderClick}
  onkeydown={(e) => {
    if (attackTargetable && (e.key === "Enter" || e.key === " ")) {
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
    padding: 4px 10px;
    border-radius: 6px;
    background: color-mix(in srgb, var(--seat-color, #888) 25%, #111a2b);
    border: 1px solid color-mix(in srgb, var(--seat-color, #888) 35%, #2e3a55);
    color: #e0e6f5;
    font-size: 12px;
    height: 28px;
    box-sizing: border-box;
    position: relative;
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
    background: color-mix(in srgb, var(--seat-color, #5fb0ff) 35%, #111a2b);
  }
  .header.active {
    border-color: var(--seat-color, #5fb0ff);
  }
  .header.priority {
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--seat-color, #5fb0ff) 60%, transparent);
  }
  .header.targetable {
    cursor: pointer;
    box-shadow:
      0 0 0 2px #ff7a7a,
      0 0 12px rgba(255, 122, 122, 0.5);
  }
  .header.targetable:hover {
    background: color-mix(in srgb, #ff7a7a 25%, #111a2b);
  }
  .header.eliminated {
    opacity: 0.55;
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
       clipped on the top/bottom. A bigger portrait would need the
       header to grow — not worth it for the 4-seat grid layout
       where every px of vertical space is already tight. */
    width: 18px;
    height: 18px;
    border-radius: 50%;
    object-fit: cover;
    flex: 0 0 auto;
    box-shadow: 0 0 0 1px var(--seat-color, #888);
  }
  .name {
    font-weight: 600;
    flex: 0 1 auto;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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
    padding: 1px 4px;
    border-radius: 3px;
    background: rgba(0, 0, 0, 0.35);
    color: #c8c8c8;
    border: 1px solid transparent;
    cursor: default;
  }
  /* Buttons need their <button> defaults stripped. */
  button.marker {
    font-family: inherit;
    cursor: pointer;
  }
  .marker.active.monarch {
    color: #ffd07a;
    border-color: #ffd07a;
    background: rgba(255, 208, 122, 0.15);
  }
  .marker.active.initiative {
    color: #b08aff;
    border-color: #b08aff;
    background: rgba(176, 138, 255, 0.15);
  }
  .marker.poison {
    color: #7aff9a;
  }
  .marker.energy {
    color: #ffd07a;
  }
  .life-controls {
    margin-left: auto;
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }
  .life {
    font-weight: 700;
    font-size: 14px;
    min-width: 22px;
    text-align: center;
    margin-left: auto;
  }
  .life-controls .life {
    margin-left: 0;
  }
  .life-btn {
    width: 18px;
    height: 18px;
    padding: 0;
    border-radius: 3px;
    border: 1px solid #2e3a55;
    background: rgba(0, 0, 0, 0.35);
    color: #e0e6f5;
    font-size: 14px;
    line-height: 1;
    cursor: pointer;
    font-family: inherit;
  }
  .life-btn:hover {
    background: #1a2335;
    border-color: #5fb0ff;
  }
  .tag {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    padding: 1px 5px;
    border-radius: 3px;
    background: rgba(0, 0, 0, 0.35);
  }
  .tag.prio {
    color: #ffd07a;
  }
  .tag.act {
    color: #9ec7ff;
  }
  .tag.elim {
    color: #ff7a7a;
  }
  /* Inline poison / energy stepper, lives in the header marker row.
     Shape mirrors the life ± controls so the bar reads as one
     coherent strip of value-with-controls widgets. */
  .counter-inline {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 1px 4px;
    border-radius: 3px;
    background: rgba(0, 0, 0, 0.35);
    border: 1px solid transparent;
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
    color: #e0e6f5;
  }
  .counter-inline.poison .counter-val {
    color: #7aff9a;
  }
  .counter-inline.energy .counter-val {
    color: #ffd07a;
  }
  .counter-btn {
    width: 14px;
    height: 14px;
    padding: 0;
    border: 0;
    background: transparent;
    color: #6c7a99;
    font-size: 11px;
    line-height: 1;
    cursor: pointer;
    font-family: inherit;
    border-radius: 2px;
  }
  .counter-btn:hover {
    background: #1a2335;
    color: #e0e6f5;
  }
</style>
