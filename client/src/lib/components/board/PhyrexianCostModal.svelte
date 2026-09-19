<script lang="ts">
  // PhyrexianCostModal — #916: how many of the cost's Phyrexian mana
  // symbols you are paying with 2 life each (CR 107.4c, and CR 107.4f
  // for the ten hybrid Phyrexian symbols).
  //
  // Which half of a {U/P} you pay is part of ANNOUNCING — CR 601.2b
  // for a cast, CR 602.2b for an activation — so this opens with the
  // other cost pickers, before anything is paid, and the answer rides
  // the one cast_spell / activate_ability message as `phyrexian_life`.
  // The engine refuses a claim larger than the cost prints and one
  // CR 119.4 forbids, so this picker enforces exactly those two
  // bounds and no third of its own.
  //
  // Shaped like XCostModal, and for the same reasons: it is one
  // number with a live auto-tap readout, so the player can see the
  // mana half shrink as they buy symbols with life rather than
  // discovering it after a rejected announcement. It serves both the
  // cast chain and the activation chain, told which by
  // `abilityIndex` — a second modal would be a second place for the
  // two to drift apart.
  //
  // It parses no mana strings: `symbols` comes from the server as
  // `phyrexian_symbols`, because the mana-cost syntax is the server's
  // to read (#787 added a whole symbol family to it).

  import { onDestroy } from "svelte";
  import { fetchAutoTapPreview, type AutoTapPreview } from "../../api";
  import {
    PhyrexianLifePerSymbol,
    clampPhyrexianLife,
    maxPhyrexianLife,
    phyrexianLifeCost,
  } from "../../phyrexianLife";
  import type { CardView } from "../../protocol";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    gameID: string;
    // The card being cast, or the permanent whose ability is being
    // activated. Null closes the modal.
    card: CardView | null;
    // How many Phyrexian symbols the cost being paid prints — the
    // server's `phyrexian_symbols`.
    symbols: number;
    // The announcer's life total, for the CR 119.4 cap.
    life: number;
    // The announced X, so the preview prices the same cost the
    // announcement will (an {X} ability is sized before this opens).
    xValue?: number;
    // Set when this belongs to an activated ability rather than a
    // cast: the preview then prices that ability's own mana
    // component instead of the card's printed cast cost.
    abilityIndex?: number;
    costLabel?: string;
    confirmVerb?: string;
    onConfirm: (n: number) => void;
    onCancel: () => void;
  }

  const {
    gameID,
    card,
    symbols,
    life,
    xValue = undefined,
    abilityIndex = undefined,
    costLabel = undefined,
    confirmVerb = "Cast",
    onConfirm,
    onCancel,
  }: Props = $props();

  // CR 119.4: a player may pay life only down to 0, so the ceiling is
  // the symbols the cost prints or what the life total can buy,
  // whichever is smaller. Recomputed from the live snapshot, so a
  // life loss while the modal is open shrinks it.
  const max = $derived(maxPhyrexianLife(symbols, life));

  let n = $state(0);
  let preview = $state<AutoTapPreview | null>(null);
  let loading = $state(false);

  // Reset when a different card — or a different ability on the same
  // card — opens the prompt. Default 0: paying the coloured half is
  // what nearly every announcement means to do.
  let lastKey: string | null = null;
  $effect(() => {
    const key = card ? `${card.instance_id}:${abilityIndex ?? "cast"}` : null;
    if (key !== lastKey) {
      lastKey = key;
      n = 0;
      preview = null;
    }
  });

  // Keep the claim inside the cap when the cap moves under an open
  // prompt — a life total can change while this is up.
  $effect(() => {
    const capped = clampPhyrexianLife(n, max);
    if (capped !== n) n = capped;
  });

  const lifeCost = $derived(phyrexianLifeCost(n));

  // Re-fetch the mana-half preview on every change; latest wins.
  let fetchSeq = 0;
  $effect(() => {
    const reqID = ++fetchSeq;
    const id = card?.instance_id;
    const claim = n;
    const x = xValue;
    const ability = abilityIndex;
    if (!id) return;
    loading = true;
    fetchAutoTapPreview(gameID, id, {
      xValue: x,
      abilityIndex: ability,
      phyrexianLife: claim,
    })
      .then((p) => {
        if (reqID === fetchSeq) preview = p;
      })
      .catch(() => {
        if (reqID === fetchSeq) preview = null;
      })
      .finally(() => {
        if (reqID === fetchSeq) loading = false;
      });
  });

  function step(delta: number): void {
    n = clampPhyrexianLife(n + delta, max);
  }

  function confirm(): void {
    if (!card) return;
    onConfirm(n);
  }

  function handleKey(e: KeyboardEvent): void {
    if (!card) return;
    if (e.key === "Enter") {
      e.preventDefault();
      confirm();
    } else if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
    } else if (e.key === "ArrowUp" || e.key === "ArrowRight") {
      e.preventDefault();
      step(1);
    } else if (e.key === "ArrowDown" || e.key === "ArrowLeft") {
      e.preventDefault();
      step(-1);
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
  <div
    class="prompt-backdrop"
    role="dialog"
    aria-modal="true"
    aria-labelledby="phyrexian-cost-title"
  >
    <div class="prompt-modal phyrexian-modal">
      <h2 id="phyrexian-cost-title">
        Phyrexian mana for {card.name}
        <span class="prompt-src" aria-hidden="true">{costLabel ?? card.mana_cost ?? ""}</span>
      </h2>
      <p class="prompt-hint">
        {symbols === 1
          ? "This cost prints one Phyrexian symbol"
          : `This cost prints ${symbols} Phyrexian symbols`}
        — each can be paid with its colour of mana, or with {PhyrexianLifePerSymbol} life. You have {life}
        life, so you can buy at most {max}.
      </p>
      <div class="pay-row">
        <span class="pay-label">Pay with life</span>
        <button
          type="button"
          class="ghost step"
          disabled={n <= 0}
          onclick={() => step(-1)}
          aria-label="one fewer symbol">−</button
        >
        <output class="count" aria-live="polite">{n} of {symbols}</output>
        <button
          type="button"
          class="ghost step"
          disabled={n >= max}
          onclick={() => step(1)}
          aria-label="one more symbol">+</button
        >
        <span class="price" class:spending={lifeCost > 0}>
          {lifeCost === 0 ? "no life" : `${lifeCost} life`}
        </span>
      </div>
      <p class="status" class:ok={preview?.ok === true} class:bad={preview?.ok === false}>
        {#if loading && !preview}
          checking the mana half…
        {:else if preview?.ok}
          the rest is affordable — auto-tap would use {preview.plan?.length ?? 0} source{(preview
            .plan?.length ?? 0) === 1
            ? ""
            : "s"}
        {:else if preview}
          the rest is not affordable
          {#if preview.missing && preview.missing.length > 0}
            — missing {preview.missing.join(" ")}
          {/if}
        {:else}
          &nbsp;
        {/if}
      </p>
      <div class="prompt-foot">
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" onclick={confirm}>
          {confirmVerb}{lifeCost > 0 ? ` for ${lifeCost} life` : ""}
          <span class="kbd">↵</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .phyrexian-modal {
    width: min(440px, calc(100vw - 32px));
  }
  .pay-row {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 14px;
  }
  .pay-label {
    font-weight: 600;
    color: var(--fg);
  }
  .step {
    min-width: 2rem;
    font-family: var(--font-mono);
    font-size: 16px;
    line-height: 1;
    padding: 4px 8px;
  }
  .count {
    font-family: var(--font-mono);
    font-size: 15px;
    min-width: 5.5rem;
    text-align: center;
  }
  .price {
    flex: 1;
    text-align: right;
    font-size: 12px;
    color: var(--fg-muted);
  }
  .price.spending {
    color: var(--danger);
  }
  .status {
    font-size: 12px;
    color: var(--fg-muted);
    line-height: 1.35;
    margin: 8px 0 0;
  }
  .status.ok {
    color: var(--mint);
  }
  .status.bad {
    color: var(--danger);
  }
  .primary .kbd {
    color: var(--accent-fg);
    border-color: rgba(28, 21, 3, 0.35);
    opacity: 0.8;
  }
</style>
