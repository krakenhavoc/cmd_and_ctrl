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
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="x-cost-title">
    <div class="prompt-modal x-modal">
      <h2 id="x-cost-title">
        Choose X for {card.name}
        <span class="prompt-src" aria-hidden="true">{card.mana_cost ?? "{X}"}</span>
      </h2>
      <p class="prompt-hint">
        Pick a value for X. The check below reads your untapped sources and says whether auto-tap
        can pay for it.
      </p>
      <label class="x-row">
        <span class="x-label">X =</span>
        <input
          type="number"
          min="0"
          step="1"
          value={x}
          oninput={(e) => clampX((e.currentTarget as HTMLInputElement).value)}
          aria-label="X value"
        />
        <span class="status" class:ok={preview?.ok === true} class:bad={preview?.ok === false}>
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
        </span>
      </label>
      <div class="prompt-foot">
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" onclick={confirm}>
          Cast with X = {x} <span class="kbd">↵</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .x-modal {
    width: min(440px, calc(100vw - 32px));
  }
  .x-row {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 14px;
  }
  .x-label {
    font-family: var(--font-mono);
    font-weight: 700;
    color: var(--fg);
  }
  .x-row input {
    width: 5rem;
    font-family: var(--font-mono);
    font-size: 15px;
    padding: 6px 10px;
    text-align: center;
  }
  .status {
    flex: 1;
    min-width: 0;
    font-size: 12px;
    color: var(--fg-muted);
    line-height: 1.35;
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
