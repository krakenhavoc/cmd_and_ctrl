<script lang="ts">
  // XCostModal — S20 sub-PR 3: announce X for a spell with {X} in its
  // cost (Blaze, Exsanguinate, Stroke of Genius). Opens before the
  // targeting step / cast; the chosen X rides the cast_spell payload
  // as x_value and the S15 cost gate charges {X}·generic for it.
  //
  // Live validation: every change to X re-asks the server's read-only
  // auto-tap preview (pool + untapped sources) whether the full cost
  // is payable at that X, so the player sees "affordable / missing
  // {R}{C}" as they type rather than after a rejected cast. The
  // preview is advisory — the gameplay.strictMana setting decides
  // whether an unaffordable X is blocked or merely warned about at
  // cast time — so Confirm is allowed on a red preview too.

  import { onDestroy } from "svelte";
  import { fetchAutoTapPreview, type AutoTapPreview } from "../../api";
  import type { CardView } from "../../protocol";

  interface Props {
    gameID: string;
    card: CardView | null;
    // Highest X worth suggesting — the caller passes the size of the
    // viewer's mana pool + untapped sources as a hint. Not a limit.
    suggestedMax?: number;
    onConfirm: (x: number) => void;
    onCancel: () => void;
  }

  const { gameID, card, suggestedMax = 0, onConfirm, onCancel }: Props = $props();

  let x = $state(0);
  let preview = $state<AutoTapPreview | null>(null);
  let loading = $state(false);

  // Reset when a different card opens the prompt.
  let lastCardID: string | null = null;
  $effect(() => {
    const id = card?.instance_id ?? null;
    if (id !== lastCardID) {
      lastCardID = id;
      x = Math.max(0, suggestedMax);
      preview = null;
    }
  });

  // Re-fetch the preview on every X change; the latest request wins.
  let fetchSeq = 0;
  $effect(() => {
    const reqID = ++fetchSeq;
    const id = card?.instance_id;
    const value = x;
    if (!id) return;
    loading = true;
    fetchAutoTapPreview(gameID, id, { xValue: value })
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

  function clampX(raw: string): void {
    const n = Math.floor(Number(raw));
    x = Number.isFinite(n) && n >= 0 ? n : 0;
  }

  function confirm(): void {
    if (!card) return;
    onConfirm(x);
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
  <div class="backdrop" role="dialog" aria-modal="true" aria-labelledby="x-cost-title">
    <div class="modal">
      <h2 id="x-cost-title">Choose X for {card.name}</h2>
      <p class="cost">cost: <span class="mono">{card.mana_cost ?? "{X}"}</span></p>
      <label class="x-row">
        <span>X =</span>
        <input
          type="number"
          min="0"
          step="1"
          value={x}
          oninput={(e) => clampX((e.currentTarget as HTMLInputElement).value)}
          aria-label="X value"
        />
      </label>
      <p class="status" class:ok={preview?.ok === true} class:bad={preview?.ok === false}>
        {#if loading && !preview}
          checking…
        {:else if preview?.ok}
          affordable — auto-tap would use {preview.plan?.length ?? 0} source{(preview.plan
            ?.length ?? 0) === 1
            ? ""
            : "s"}
        {:else if preview}
          not affordable
          {#if preview.missing && preview.missing.length > 0}
            — missing {preview.missing.join(" ")}
          {/if}
        {:else}
          &nbsp;
        {/if}
      </p>
      <div class="actions">
        <button type="button" class="cancel" onclick={onCancel}>Cancel</button>
        <button type="button" class="confirm" onclick={confirm}>Cast with X = {x}</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.65);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1100;
  }
  .modal {
    background: #1c1f2a;
    color: #e6e8ee;
    border: 1px solid #3a4055;
    border-radius: 6px;
    padding: 1.25rem 1.5rem;
    min-width: 320px;
    max-width: 440px;
    box-shadow: 0 12px 28px rgba(0, 0, 0, 0.4);
  }
  h2 {
    margin: 0 0 0.5rem 0;
    font-size: 1.05rem;
    color: #f4ead5;
  }
  .cost {
    margin: 0 0 0.75rem 0;
    color: #aab2c8;
    font-size: 0.9rem;
  }
  .mono {
    font-family: "Menlo", "Monaco", monospace;
    color: #e6e8ee;
  }
  .x-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 1rem;
  }
  .x-row input {
    width: 5rem;
    font-size: 1.1rem;
    padding: 0.25rem 0.4rem;
    background: #0f1118;
    color: #e6e8ee;
    border: 1px solid #3a4055;
    border-radius: 4px;
  }
  .status {
    margin: 0.6rem 0 0.9rem 0;
    font-size: 0.85rem;
    color: #aab2c8;
    min-height: 1.2em;
  }
  .status.ok {
    color: #6fe3a4;
  }
  .status.bad {
    color: #ff9a9a;
  }
  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }
  .actions button {
    padding: 0.4rem 0.8rem;
    border-radius: 4px;
    border: 1px solid #3a4055;
    background: #262a38;
    color: #e6e8ee;
    cursor: pointer;
  }
  .actions .confirm {
    background: #2d5a3f;
    border-color: #3f7a55;
  }
</style>
