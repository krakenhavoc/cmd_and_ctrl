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
  //
  // It also serves a CR 602 activated ability whose cost carries {X}
  // (Helm of Obedience, Treasure Vault, Soothsaying). Same question,
  // same moment in the announcement, same answer shape — so it is
  // the same picker, told which ability to price via `abilityIndex`
  // and which floor to respect via `minX`. A second modal would be a
  // second place for the two to drift apart.

  import { onDestroy } from "svelte";
  import { fetchAutoTapPreview, type AutoTapPreview } from "../../api";
  import type { AutoTapCastParams } from "../../castPreview";
  import { xPickerCostNotes } from "../../costNotes";
  import type { CardView } from "../../protocol";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    gameID: string;
    card: CardView | null;
    // Highest X worth suggesting — the caller passes the size of the
    // viewer's mana pool + untapped sources as a hint. Not a limit.
    suggestedMax?: number;
    // Set when the X belongs to an activated ability rather than a
    // cast: `abilityIndex` prices that ability's own cost in the
    // preview, `costLabel` is what the heading shows, and `minX` is
    // the printed floor ("X can't be 0" passes 1). The server
    // rejects an announcement below the floor, so the input does too.
    abilityIndex?: number;
    costLabel?: string;
    minX?: number;
    // #696: the rest of the announcement the preview prices against —
    // source zone, alternative cost, optional costs, face. All of them
    // are chosen before X, so the affordable / missing readout can be
    // about the cast the player is actually making: an overloaded
    // Cyclonic Rift, a flashed-back Deep Analysis. Ignored on the
    // ability branch, which prices the ability's own cost.
    castParams?: AutoTapCastParams;
    confirmVerb?: string;
    onConfirm: (x: number) => void;
    onCancel: () => void;
  }

  const {
    gameID,
    card,
    suggestedMax = 0,
    abilityIndex = undefined,
    costLabel = undefined,
    minX = 0,
    castParams = {},
    confirmVerb = "Cast",
    onConfirm,
    onCancel,
  }: Props = $props();

  const floor = $derived(Math.max(0, minX));
  // #746: a spell whose price depends on its targets (Fireball) is
  // priced here at one target, because targets come after X. Quote
  // the printed clause rather than show a surcharge nobody knows yet.
  const costNotes = $derived(xPickerCostNotes(card, abilityIndex));

  let x = $state(0);
  let preview = $state<AutoTapPreview | null>(null);
  let loading = $state(false);

  // Reset when a different card — or a different ability on the same
  // card — opens the prompt.
  let lastKey: string | null = null;
  $effect(() => {
    const key = card ? `${card.instance_id}:${abilityIndex ?? "cast"}` : null;
    if (key !== lastKey) {
      lastKey = key;
      x = Math.max(floor, suggestedMax);
      preview = null;
    }
  });

  // Re-fetch the preview on every X change; the latest request wins.
  let fetchSeq = 0;
  $effect(() => {
    const reqID = ++fetchSeq;
    const id = card?.instance_id;
    const value = x;
    const ability = abilityIndex;
    const cast = castParams;
    if (!id) return;
    loading = true;
    fetchAutoTapPreview(gameID, id, { xValue: value, abilityIndex: ability, cast })
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
    x = Number.isFinite(n) && n >= floor ? n : floor;
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
  <ModalLayer />
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="x-cost-title">
    <div class="prompt-modal x-modal">
      <h2 id="x-cost-title">
        Choose X for {card.name}
        <span class="prompt-src" aria-hidden="true">{costLabel ?? card.mana_cost ?? "{X}"}</span>
      </h2>
      {#if floor > 0}
        <!-- "X can't be 0" is a printed floor, not advice: the
             server refuses an announcement under it outright. -->
        <p class="prompt-hint">
          Pick a value for X — this ability's X can't be less than {floor}. The check below reads
          your untapped sources and says whether auto-tap can pay for it.
        </p>
      {:else if card.additional_cost?.demands_x}
        <!-- S23: Toxic Deluge's X is paid in LIFE, not mana, so the
             auto-tap line below is about the flat printed cost and
             says nothing about whether the X itself is affordable.
             Name the real price instead of letting a green
             "affordable" imply it covers both. -->
        <p class="prompt-hint">
          {card.additional_cost.label ?? "Pay X life"} as an additional cost. You'll pay {x} life.
        </p>
      {:else}
        <p class="prompt-hint">
          Pick a value for X. The check below reads your untapped sources and says whether auto-tap
          can pay for it.
        </p>
      {/if}
      <label class="x-row">
        <span class="x-label">X =</span>
        <input
          type="number"
          min={floor}
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
      {#if costNotes.length > 0}
        <p class="prompt-hint cost-note">
          Checked at one target.
          {#each costNotes as note (note)}
            <span class="cost-clause">{note}</span>
          {/each}
        </p>
      {/if}
      <div class="prompt-foot">
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" onclick={confirm}>
          {confirmVerb} with X = {x} <span class="kbd">↵</span>
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
  .cost-note {
    margin-top: 8px;
    font-size: 12px;
  }
  .cost-clause {
    display: block;
    font-style: italic;
  }
  .primary .kbd {
    color: var(--accent-fg);
    border-color: rgba(28, 21, 3, 0.35);
    opacity: 0.8;
  }
</style>
