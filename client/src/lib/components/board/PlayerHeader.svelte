<script lang="ts">
  // PlayerHeader is the small bar above each PlayerPanel that shows
  // who the player is, their current life, an active/priority dot,
  // and the eliminated state. During combat it doubles as the
  // attack-target affordance: when the viewer is in attack mode with
  // a selected attacker, opponent headers light up and become
  // clickable to commit the attack.

  import { seatColor } from "../../colors";
  import type { PlayerView } from "../../protocol";
  import { floatUp, fadeOut } from "../../animations";

  interface Props {
    seat: PlayerView;
    isSelf: boolean;
    isActive: boolean;
    hasPriority: boolean;
    attackTargetable: boolean;
    onDeclareAttack?: (targetPlayerID: string) => void;
  }

  const { seat, isSelf, isActive, hasPriority, attackTargetable, onDeclareAttack }: Props =
    $props();

  function handleClick(): void {
    if (!attackTargetable) return;
    onDeclareAttack?.(seat.id);
  }

  // Damage / heal popup driven by the seat's life_history. The server
  // appends a LifeChangeView every change; we snapshot the count on
  // first render so reconnect / page-refresh doesn't pop a flurry of
  // historical changes, then react to anything that arrives after.
  // Only the most recent change is shown — if two changes arrive in
  // quick succession, the second overwrites the first (the user reads
  // the net effect from the life total either way).
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
  onclick={handleClick}
  onkeydown={(e) => {
    if (attackTargetable && (e.key === "Enter" || e.key === " ")) {
      e.preventDefault();
      handleClick();
    }
  }}
  aria-label={attackTargetable ? `attack ${seat.name}` : `${seat.name}, ${seat.life} life`}
>
  <span class="seat-dot" aria-hidden="true"></span>
  <span class="name">{seat.name}</span>
  <span class="life">{seat.life}</span>
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
    /* Anchor for the floating dmg-popup so it positions relative to
       the header rather than the document. */
    position: relative;
  }
  .dmg-popup {
    position: absolute;
    right: 12px;
    /* Anchored just below the header (top: 100%) instead of above it.
       The parent .panel uses overflow: hidden — needed to clip the
       rotated opponent content — so a popup positioned ABOVE the
       header is invisible. The floatUp / fadeOut tick callbacks both
       translate negatively, so the popup still reads as "floats
       upward off the player" while staying inside the clip box. */
    top: 100%;
    margin-top: 4px;
    font-size: 22px;
    font-weight: 800;
    pointer-events: none;
    z-index: 20;
    text-shadow:
      0 1px 0 rgba(0, 0, 0, 0.8),
      0 0 8px rgba(0, 0, 0, 0.55);
    /* Inline transforms are owned by GSAP via the floatUp / fadeOut
       transitions; no static transform here. */
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
  .life {
    margin-left: auto;
    font-weight: 700;
    font-size: 14px;
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
</style>
