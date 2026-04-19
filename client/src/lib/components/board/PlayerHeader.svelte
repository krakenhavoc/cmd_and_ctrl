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
  aria-label={attackTargetable ? `attack ${seat.name}` : `${seat.name}, ${seat.life} life`}
>
  <span class="seat-dot" aria-hidden="true"></span>
  <span class="name">{seat.name}</span>

  <!-- Marker badges. Monarch and initiative are claim toggles for self
       (clear when already-held, set otherwise). Poison / energy show
       only when non-zero (kept terse on opponent headers; self has the
       inline steppers below). -->
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
    {:else}
      {#if isMonarch}
        <span class="marker monarch active" title="monarch" aria-label="monarch">👑</span>
      {/if}
      {#if isInitiative}
        <span class="marker initiative active" title="initiative" aria-label="initiative">⚔</span>
      {/if}
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

{#if isSelf}
  <!-- Inline second-row steppers for poison / energy. Hidden by default
       in the header proper to keep the bar uncluttered; here they sit
       just below it so the viewer can fiddle without a modal. -->
  <div class="counters-row" aria-label="poison and energy controls">
    <span class="counter-stepper">
      <span class="counter-label">poison</span>
      <button
        type="button"
        title="-1 poison"
        aria-label="lose 1 poison"
        onclick={() => changePoison(-1)}>−</button
      >
      <span class="counter-value">{seat.poison ?? 0}</span>
      <button
        type="button"
        title="+1 poison"
        aria-label="gain 1 poison"
        onclick={() => changePoison(1)}>+</button
      >
    </span>
    <span class="counter-stepper">
      <span class="counter-label">energy</span>
      <button
        type="button"
        title="-1 energy"
        aria-label="lose 1 energy"
        onclick={() => changeEnergy(-1)}>−</button
      >
      <span class="counter-value">{seat.energy ?? 0}</span>
      <button
        type="button"
        title="+1 energy"
        aria-label="gain 1 energy"
        onclick={() => changeEnergy(1)}>+</button
      >
    </span>
  </div>
{/if}

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
  .counters-row {
    display: flex;
    gap: 12px;
    align-items: center;
    margin-top: 2px;
    padding: 0 10px;
    font-size: 10px;
    color: #6c7a99;
  }
  .counter-stepper {
    display: inline-flex;
    align-items: center;
    gap: 3px;
  }
  .counter-label {
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: #6c7a99;
  }
  .counter-stepper button {
    width: 16px;
    height: 16px;
    padding: 0;
    border: 1px solid #2e3a55;
    border-radius: 3px;
    background: transparent;
    color: #c8c8c8;
    font-size: 11px;
    cursor: pointer;
    font-family: inherit;
    line-height: 1;
  }
  .counter-stepper button:hover {
    background: #1a2335;
    color: #e0e6f5;
  }
  .counter-value {
    min-width: 18px;
    text-align: center;
    color: #e0e6f5;
    font-weight: 600;
  }
</style>
