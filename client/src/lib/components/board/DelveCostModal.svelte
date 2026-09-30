<script lang="ts">
  // DelveCostModal — ADR 0100: pick the cards in your graveyard that
  // delve exiles to help pay for a spell (CR 702.66a). Each one pays for
  // {1} of the generic mana; none can pay a coloured symbol.
  //
  // Opens after the X prompt and after the convoke picker, because both
  // change how much generic mana is left to delve. The cap is not
  // worked out here: the modal asks the server's auto-tap preview for
  // `delve_budget`, priced for the announcement as it stands, so an X,
  // a kicker or a convoke tap all move it the way the server will.
  //
  // Optional in both directions, like TapCostModal: exiling nothing is
  // always a legal cast, and Confirm is always enabled. "Choose for me"
  // fills the budget from the front of the server's payment order
  // (lands first, then cards with no graveyard cast surface) and never
  // confirms for the player — ADR 0100 owner decision 2.

  import { onDestroy } from "svelte";
  import { fetchAutoTapPreview, type AutoTapPreview } from "../../api";
  import type { AutoTapCastParams } from "../../castPreview";
  import { chooseDelveForMe, delveLimit, delveOptionIDs } from "../../delve";
  import type { CardView } from "../../protocol";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    gameID: string;
    // The spell being cast; null closes the modal.
    card: CardView | null;
    // The caster's graveyard cards that may be exiled, in the server's
    // payment order.
    options: CardView[];
    // The announcement so far, so the preview prices the right cast.
    xValue?: number;
    phyrexianLife?: number;
    castParams?: AutoTapCastParams;
    onConfirm: (instanceIDs: string[]) => void;
    onCancel: () => void;
  }

  const {
    gameID,
    card,
    options,
    xValue = 0,
    phyrexianLife = 0,
    castParams = {},
    onConfirm,
    onCancel,
  }: Props = $props();

  let chosen = $state<string[]>([]);
  let preview = $state<AutoTapPreview | null>(null);

  // Reset when a different cast opens the prompt.
  let lastCardID: string | null = null;
  $effect(() => {
    const id = card?.instance_id ?? null;
    if (id !== lastCardID) {
      lastCardID = id;
      chosen = [];
      preview = null;
    }
  });

  // The budget, from the server. The latest request wins.
  let fetchSeq = 0;
  $effect(() => {
    const reqID = ++fetchSeq;
    const id = card?.instance_id;
    if (!id) return;
    fetchAutoTapPreview(gameID, id, { xValue, phyrexianLife, cast: castParams })
      .then((p) => {
        if (reqID === fetchSeq) preview = p;
      })
      .catch(() => {
        if (reqID === fetchSeq) preview = null;
      });
  });

  const limit = $derived(delveLimit(preview, card));
  const full = $derived(chosen.length >= limit);

  function toggle(id: string): void {
    if (chosen.includes(id)) {
      chosen = chosen.filter((c) => c !== id);
      return;
    }
    if (full) return;
    chosen = [...chosen, id];
  }

  function chooseForMe(): void {
    chosen = chooseDelveForMe(delveOptionIDs(card), limit);
  }

  function confirm(): void {
    onConfirm(chosen.slice(0, limit));
  }

  function handleKey(e: KeyboardEvent): void {
    if (!card) return;
    if (e.key === "Enter") {
      e.preventDefault();
      confirm();
    } else if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
    }
  }
  $effect(() => {
    if (!card) return;
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  });
  onDestroy(() => document.removeEventListener("keydown", handleKey));
</script>

{#if card}
  <ModalLayer />
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="delve-cost-title">
    <div class="prompt-modal dv-modal">
      <h2 id="delve-cost-title">
        {card.name}
        <span class="prompt-src" aria-hidden="true">delve · CR 702.66</span>
      </h2>
      <p class="prompt-hint">Delve</p>
      <p class="prompt-hint sub">
        Each card you exile from your graveyard pays for {"{1}"} of the generic mana.
      </p>
      {#if limit === 0}
        <p class="prompt-hint error">
          There is no generic mana left to delve — exile nothing and continue.
        </p>
      {:else}
        <ul class="prompt-options">
          {#each options as c (c.instance_id)}
            <li>
              <button
                type="button"
                class="prompt-opt"
                class:on={chosen.includes(c.instance_id)}
                disabled={full && !chosen.includes(c.instance_id)}
                aria-pressed={chosen.includes(c.instance_id)}
                onclick={() => toggle(c.instance_id)}
              >
                <span class="prompt-radio" aria-hidden="true"></span>
                <span class="name">{c.name}</span>
                {#if c.type_line}
                  <span class="note">{c.type_line}</span>
                {/if}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
      <div class="prompt-foot">
        <span class="prompt-count">{chosen.length} / {limit} exiled</span>
        {#if limit > 0}
          <button type="button" class="ghost" onclick={chooseForMe}>Choose for me</button>
        {/if}
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" onclick={confirm}>
          {chosen.length === 0 ? "Exile nothing" : "Exile"}
          <span class="kbd">↵</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .dv-modal {
    width: min(440px, calc(100vw - 32px));
  }
  .name {
    flex: 1 1 auto;
  }
  .note {
    font-size: 12px;
    color: var(--fg-muted);
  }
  .sub {
    font-size: 12px;
    color: var(--fg-muted);
  }
  .primary .kbd {
    color: var(--accent-fg);
    border-color: rgba(28, 21, 3, 0.35);
    opacity: 0.8;
  }
</style>
