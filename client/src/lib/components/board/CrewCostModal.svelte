<script lang="ts">
  // CrewCostModal — S27: pick the creatures you tap to crew a
  // Vehicle (CR 702.122a).
  //
  // Shaped like TapCostModal and deliberately not the same component,
  // because the arithmetic is the opposite way round. Convoke counts
  // PERMANENTS against a ceiling — tap at most N, tapping none is
  // fine. Crew counts POWER against a floor — tap any number, but the
  // total has to reach the crew number or the ability cannot be
  // activated at all. So this modal tracks a running power total, the
  // confirm button is disabled until it clears the bar, and there is
  // no "crew with nothing" escape hatch.
  //
  // Two things the option list gets right that are easy to get wrong:
  //
  //   - Summoning-sick creatures ARE offered. Tapping a creature to
  //     crew is not paying a {T} cost, so a creature cast this turn
  //     may crew (CR 702.122b). The server agrees; this list comes
  //     straight off its crew_options.
  //   - Crewing is a cost, not a target, so hexproof and protection
  //     never apply and the list is a plain roster of your own
  //     untapped creatures rather than the board-click targeting
  //     flow — the same shape as SacrificeCostModal.
  //
  // ADR 0111 PR 6: a sheet in the action dock; Crew and Cancel are the
  // dock's action bar (Enter / Escape through its one key handler).

  import type { ActivatedAbilityView, CardView } from "../../protocol";
  import { crewAvailablePower, crewPower, crewSatisfied } from "../../crew";
  import { cancelAction, confirmAction } from "../../dock";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    // The Vehicle being crewed; null closes the modal.
    card: CardView | null;
    // The crew ability, for its number and its label. Null for a
    // teamwork cost, which passes `threshold` instead.
    ability: ActivatedAbilityView | null;
    // #1703: teamwork (CR 702.194a) is crew's sentence on a spell —
    // the same floor-on-total-power picker. A teamwork prompt passes
    // the number here, and its own keyword and rule for the heading.
    threshold?: number;
    keyword?: string;
    rule?: string;
    // The creatures the server says could pay it right now.
    options: CardView[];
    onConfirm: (instanceIDs: string[]) => void;
    onCancel: () => void;
  }

  const {
    card,
    ability,
    threshold,
    keyword: keywordProp = "Crew",
    rule: ruleProp = "CR 702.122",
    options,
    onConfirm,
    onCancel,
  }: Props = $props();

  // #2695: a Mount's saddle ability (CR 702.171a) is this same picker —
  // crew's floor on total power — over OTHER creatures, so the server's
  // crew_options already leaves the Mount out. Only the words change.
  const saddle = $derived(ability?.saddle === true);
  const keyword = $derived(saddle ? "Saddle" : keywordProp);
  const rule = $derived(saddle ? "CR 702.171" : ruleProp);

  let chosen = $state<string[]>([]);

  // Reset when a different activation opens the prompt.
  let lastKey: string | null = null;
  $effect(() => {
    const key = card
      ? ability
        ? `${card.instance_id}:${ability.index}`
        : threshold !== undefined
          ? `${card.instance_id}:${keyword}`
          : null
      : null;
    if (key !== lastKey) {
      lastKey = key;
      chosen = [];
    }
  });

  const need = $derived(threshold ?? ability?.crew_cost ?? 0);
  const open = $derived(card !== null && (ability !== null || threshold !== undefined));

  // Power is read off the CardView, which already carries the
  // post-layer effective value the server will re-check against — so
  // the running total shown here and the total the server computes
  // agree without a second round trip. A card with no power (it
  // should not appear here) contributes nothing rather than NaN.
  const total = $derived(crewPower(options, chosen));
  const enough = $derived(crewSatisfied(options, chosen, need));

  // The best the whole board could do. When it falls short the modal
  // says so plainly instead of letting a player click around looking
  // for the creature that would finish the total.
  const available = $derived(crewAvailablePower(options));

  function toggle(id: string): void {
    chosen = chosen.includes(id) ? chosen.filter((c) => c !== id) : [...chosen, id];
  }

  function confirm(): void {
    if (!enough) return;
    onConfirm(chosen);
  }
</script>

{#if open && card}
  <DockSheet
    label={card.name}
    src={`${keyword} ${need} · ${rule}`}
    width={560}
    sheetKey={`crew:${card.instance_id}:${ability?.index ?? keyword}`}
    count={`${total} / ${need} power`}
    primary={confirmAction(keyword === "Crew" || saddle ? keyword : "Tap", confirm, {
      disabled: !enough,
    })}
    secondary={[cancelAction(onCancel)]}
  >
    <p class="prompt-hint">
      Tap any number of {saddle ? "other " : ""}untapped creatures you control with total power {need}
      or more.
    </p>
    {#if options.length === 0}
      <p class="prompt-hint error">You control no untapped creatures to tap.</p>
    {:else if available < need}
      <p class="prompt-hint error">
        Your untapped creatures total {available} power — not enough for {keyword.toLowerCase()}
        {need}.
      </p>
    {:else}
      <ul class="prompt-options">
        {#each options as c (c.instance_id)}
          <li>
            <button
              type="button"
              class="prompt-opt"
              class:on={chosen.includes(c.instance_id)}
              aria-pressed={chosen.includes(c.instance_id)}
              onclick={() => toggle(c.instance_id)}
            >
              <span class="prompt-radio" aria-hidden="true"></span>
              <span class="name">{c.name}</span>
              {#if c.power !== undefined && c.toughness !== undefined}
                <span class="note pt">{c.power}/{c.toughness}</span>
              {/if}
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </DockSheet>
{/if}

<style>
  .name {
    flex: 1 1 auto;
  }
  .pt {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--fg-muted);
  }
</style>
