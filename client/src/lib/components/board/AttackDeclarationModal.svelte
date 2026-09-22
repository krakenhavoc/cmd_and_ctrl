<script lang="ts">
  // AttackDeclarationModal — #1162, ADR 0080 §7's deferred client
  // half.
  //
  // Opens when "attack with all" is refused for want of the CR 508.1a
  // attack tax (Propaganda, Ghostly Prison, Sphere of Safety), or when
  // the player chooses "pick attackers…" beside a taxed target
  // proactively. Either way, the problem is the same: the seat can pay
  // for SOME of a wide swing but not all of it, and the bulk verb is
  // all-or-nothing (ADR 0080 §3) — nothing about which attacks to drop
  // is the engine's to decide. This is the subset picker over the
  // eligible set that lets the player make that choice once, instead
  // of working out a smaller number by hand and declaring one attacker
  // at a time.
  //
  // The SECOND half is the lock-a-land toggle: declare_attacker and
  // declare_attackers both accept `locked_sources` (ADR 0080 §4) but
  // nothing sent it before this. The candidate rows are the viewer's
  // own currently-untapped mana-producing permanents — the same
  // "usable right now" read seatSummary.ts's opponent-panel mana count
  // already computes — locking one reserves it for a later cast.
  //
  // What this is NOT: a tap-plan preview like AutoTapPreviewModal's.
  // That modal fetches `/games/:id/auto-tap-preview`, which is
  // card-shaped (`?card=<uuid>`, priced through g.PriceCast) and has
  // no declaration-shaped sibling — extending it to preview a set of
  // attackers is server work ADR 0080 §7 explicitly deferred, and this
  // PR does not add it (see the PR body). So there is no fetched
  // `plan` here and no "which permanents WILL tap" readout — only the
  // price (server-priced, read off `attack_targets[].tax`, never
  // derived) and a manual lock affordance, same interaction shape as
  // that modal's per-row lock toggle, sourced locally instead of from
  // a preview response.

  import { onDestroy } from "svelte";
  import type { CardView, GameView } from "../../protocol";
  import { attackTaxLabelForCount, attackTaxOn, planAttackAll } from "../../attackAll";
  import { usableManaAbilities } from "../../seatSummary";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    view: GameView;
    viewerID: string | null;
    // Doubles as open/closed, like AutoTapPreviewModal's `cardID` —
    // null means the modal is unmounted.
    defenderSeatID: string | null;
    onConfirm: (attackerIDs: string[], lockedSources: string[]) => void;
    onCancel: () => void;
  }

  const { view, viewerID, defenderSeatID, onConfirm, onCancel }: Props = $props();

  const plan = $derived(planAttackAll(view, viewerID));
  const defender = $derived(plan.defenders.find((s) => s.id === defenderSeatID) ?? null);
  const each = $derived(defenderSeatID ? attackTaxOn(view, defenderSeatID) : "");

  let selected = $state<string[]>([]);
  let lockedSources = $state<string[]>([]);

  // Seed the selection with every eligible attacker when the modal
  // OPENS for a defender — not on every later snapshot, which would
  // silently discard a player's in-progress deselection. Mirrors
  // ModePickerModal's lastCardID reset.
  let lastDefenderSeatID: string | null = null;
  $effect(() => {
    if (defenderSeatID !== lastDefenderSeatID) {
      lastDefenderSeatID = defenderSeatID;
      selected = plan.eligible.map((c) => c.instance_id);
      lockedSources = [];
    }
  });

  // A selected ID can go stale between the modal opening and the
  // confirm click — a creature dies to a response, a snapshot lands
  // mid-pick. Filtering against the live plan on every read (rather
  // than only at confirm time) keeps the checked count and the price
  // label honest with what will actually be declared.
  const liveSelected = $derived.by(() => {
    const eligibleIDs = new Set(plan.eligible.map((c) => c.instance_id));
    return selected.filter((id) => eligibleIDs.has(id));
  });

  const lockCandidates = $derived.by((): CardView[] => {
    if (!viewerID) return [];
    return (view.battlefield?.cards ?? []).filter(
      (c) => c.controller === viewerID && usableManaAbilities(c).length > 0,
    );
  });

  function toggleAttacker(id: string): void {
    selected = selected.includes(id) ? selected.filter((x) => x !== id) : [...selected, id];
  }

  function toggleLock(id: string): void {
    lockedSources = lockedSources.includes(id)
      ? lockedSources.filter((x) => x !== id)
      : [...lockedSources, id];
  }

  function selectAll(): void {
    selected = plan.eligible.map((c) => c.instance_id);
  }
  function selectNone(): void {
    selected = [];
  }

  function confirm(): void {
    if (liveSelected.length === 0) return;
    onConfirm(liveSelected.slice(), lockedSources.slice());
  }

  function handleKey(e: KeyboardEvent): void {
    if (!defenderSeatID) return;
    if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
    }
  }
  $effect(() => {
    if (!defenderSeatID) return;
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  });
  onDestroy(() => document.removeEventListener("keydown", handleKey));
</script>

{#if defenderSeatID && defender}
  <ModalLayer />
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="attack-pick-title">
    <div class="prompt-modal">
      <h2 id="attack-pick-title">
        Choose attackers
        <span class="prompt-src" aria-hidden="true">{defender.display_name || defender.name}</span>
      </h2>
      <p class="prompt-hint">
        {#if each}
          Attacking {defender.display_name || defender.name} costs {each} per creature. Pick which ones
          to send.
          {#if attackTaxLabelForCount(each, liveSelected.length)}
            <strong>{attackTaxLabelForCount(each, liveSelected.length)}</strong> for the
            {liveSelected.length} checked below.
          {/if}
        {:else}
          Pick which creatures to send at {defender.display_name || defender.name}.
        {/if}
      </p>
      <ul class="prompt-options" role="group" aria-label="attackers">
        {#each plan.eligible as c (c.instance_id)}
          {@const on = selected.includes(c.instance_id)}
          <li>
            <button
              type="button"
              class="prompt-opt"
              class:on
              role="checkbox"
              aria-checked={on}
              onclick={() => toggleAttacker(c.instance_id)}
            >
              <span class="prompt-radio" aria-hidden="true"></span>
              <span class="label">{c.name || "unknown creature"}</span>
            </button>
          </li>
        {/each}
      </ul>
      <div class="prompt-foot">
        <span class="prompt-count">{liveSelected.length} / {plan.eligible.length} attacking</span>
        <button type="button" class="ghost" onclick={selectAll}>All</button>
        <button type="button" class="ghost" onclick={selectNone}>None</button>
      </div>
      {#if lockCandidates.length > 0}
        <p class="prompt-hint locked-label">
          lock a land — reserved sources the auto-tapper won't reach for
        </p>
        <ul class="prompt-options plan">
          {#each lockCandidates as c (c.instance_id)}
            {@const locked = lockedSources.includes(c.instance_id)}
            <li class="prompt-opt src-row" class:on={locked}>
              <span class="card-name">{c.name || "unknown permanent"}</span>
              <button
                type="button"
                class="ghost lock-btn"
                onclick={() => toggleLock(c.instance_id)}
                title={locked
                  ? "release this source back to the auto-tapper"
                  : "reserve this source for a later cast"}
              >
                {locked ? "unlock" : "lock"}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
      <div class="prompt-foot">
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button
          type="button"
          class="primary"
          onclick={confirm}
          disabled={liveSelected.length === 0}
        >
          Attack with {liveSelected.length}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .src-row {
    justify-content: space-between;
    padding: 6px 6px 6px 12px;
    cursor: default;
  }
  .card-name {
    font-size: 13px;
    font-weight: 500;
  }
  .lock-btn {
    height: 26px;
    padding: 0 10px;
    font-size: 11px;
    font-family: var(--font-mono);
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }
  .locked-label {
    margin: 2px 0 -6px;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
    font-weight: 700;
  }
</style>
