<script lang="ts">
  // ChoicePromptModal opens when the wire's pending_choices queue
  // has an entry addressed to the viewer. Generic counterpart to
  // DiscardPromptModal: that one handles the S13.4 cleanup-specific
  // map where chooser == owner; this one handles the S14+ queue
  // where chooser can differ from the pool owner (Thoughtseize:
  // caster picks from target's revealed hand).
  //
  // The options[] slice on each entry is pre-filtered by the
  // server's per-viewer KnownBy projection — cards the viewer
  // is legally allowed to see arrive face-up; the rest arrive
  // redacted (backs). So this modal just renders options[] as-is.

  import type {
    CardView,
    GameView,
    PendingChoiceView,
    ReplacementOptionView,
  } from "../../protocol";
  import Card from "./Card.svelte";

  interface Props {
    snap: GameView;
    viewerID: string | null;
    sendAction: (type: string, params?: unknown, player?: string) => void;
  }

  const { snap, viewerID, sendAction }: Props = $props();

  // First choice addressed to the viewer. Queue ordering: front of
  // list is "what the chooser sees next." One modal at a time; when
  // they resolve this one, the next pops automatically on the
  // snapshot after the server drains the entry.
  const active = $derived.by((): PendingChoiceView | null => {
    if (!viewerID || !snap.pending_choices) return null;
    for (const c of snap.pending_choices) {
      if (c.chooser === viewerID) return c;
    }
    return null;
  });

  const open = $derived(active !== null);

  // Source player (whose hand the picks come from). Used for the
  // modal header copy.
  const fromName = $derived.by(() => {
    if (!active) return "";
    const seat = snap.seats.find((s) => s.id === active.from_player);
    return seat?.name ?? "opponent";
  });

  const isSelfSource = $derived(active && viewerID && active.from_player === viewerID);

  let selected = $state<Set<string>>(new Set());
  // S17 replacement_order: array of effect IDs in the order the
  // chooser has picked. Click a row to append; click again to
  // remove (and subsequent positions compact down). Submit when
  // the array covers every candidate.
  let ordered = $state<string[]>([]);

  // Reset selection whenever the modal opens fresh (active changes
  // from null → non-null, or the choice ID changes).
  let lastChoiceID: string | null = null;
  $effect(() => {
    const nextID = active?.id ?? null;
    if (nextID !== lastChoiceID) {
      selected = new Set();
      ordered = [];
      lastChoiceID = nextID;
    }
  });

  function toggle(id: string): void {
    if (!active) return;
    const next = new Set(selected);
    if (next.has(id)) next.delete(id);
    else if (next.size < active.count) next.add(id);
    selected = next;
  }

  function submit(): void {
    if (!active || !viewerID) return;
    if (selected.size !== active.count) return;
    sendAction(
      "resolve_choice",
      { choice_id: active.id, card_ids: Array.from(selected) },
      viewerID,
    );
  }

  // Options come redacted for non-knowers; filter to cards the
  // viewer can identify so we don't render a grid of anonymous
  // backs the viewer can't meaningfully pick from. Unknown
  // options here would mean the server misrouted — the chooser
  // should always be a knower of the revealed cards.
  const optionCards = $derived<CardView[]>(active?.options ?? []);

  // S15 mana_pick branch — a color-pick choice from Arcane Signet /
  // Birds of Paradise. `active.color_options` is the server-filtered
  // legal button list. Submits via resolve_choice with `{choice_id,
  // color}` (card_ids absent).
  const isManaPick = $derived(active?.kind === "mana_pick");
  const colorOptions = $derived<string[]>(active?.color_options ?? []);

  const COLOR_META: Record<string, { label: string; fill: string }> = {
    W: { label: "White", fill: "#f4ead5" },
    U: { label: "Blue", fill: "#aad4ff" },
    B: { label: "Black", fill: "#2b2b3d" },
    R: { label: "Red", fill: "#ff9a85" },
    G: { label: "Green", fill: "#92c493" },
    C: { label: "Colorless", fill: "#c6cfdd" },
  };

  function pickColor(color: string): void {
    if (!active || !viewerID) return;
    sendAction("resolve_choice", { choice_id: active.id, color }, viewerID);
  }

  // S17 replacement_order branch — CR 616 affected-player-chooses-
  // order prompt. Click a row to append it to the `ordered` array;
  // click a row already in the array to remove it (later rows
  // compact down). Submit with { choice_id, order: [...] }.
  const isReplacementOrder = $derived(active?.kind === "replacement_order");
  const replacementOptions = $derived<ReplacementOptionView[]>(
    active?.replacement_options ?? [],
  );

  function toggleReplacement(id: string): void {
    const idx = ordered.indexOf(id);
    if (idx >= 0) {
      ordered = [...ordered.slice(0, idx), ...ordered.slice(idx + 1)];
    } else {
      ordered = [...ordered, id];
    }
  }

  function submitReplacementOrder(): void {
    if (!active || !viewerID) return;
    if (ordered.length !== replacementOptions.length) return;
    sendAction(
      "resolve_choice",
      { choice_id: active.id, order: ordered },
      viewerID,
    );
  }

  function positionFor(id: string): number {
    return ordered.indexOf(id) + 1; // 1-indexed; 0 = unselected
  }

  function sourceCardName(opt: ReplacementOptionView): string {
    if (!opt.source_card_id || !snap.battlefield) return "";
    for (const c of snap.battlefield.cards) {
      if (c.instance_id === opt.source_card_id) return c.name ?? "";
    }
    return "";
  }
</script>

{#if open && active}
  <div class="backdrop" role="dialog" aria-modal="true" aria-labelledby="choice-title">
    <div class="modal">
      {#if isManaPick}
        <h2 id="choice-title">{active.reason || "Pick a color"}</h2>
        <p class="hint">Choose a color to add to your mana pool.</p>
        <div class="color-row">
          {#each colorOptions as color (color)}
            {@const meta = COLOR_META[color] ?? { label: color, fill: "#ccc" }}
            <button
              type="button"
              class="color-pick"
              style:--fill={meta.fill}
              title={meta.label}
              aria-label={`add ${meta.label} mana`}
              onclick={() => pickColor(color)}
            >
              <span class="color-letter">{color}</span>
              <span class="color-name">{meta.label}</span>
            </button>
          {/each}
        </div>
      {:else if isReplacementOrder}
        <h2 id="choice-title">{active.reason || "Order replacement effects"}</h2>
        <p class="hint">
          Click each effect in the order it should apply. Different orders can
          produce different results — you choose as the affected player
          (CR 616).
        </p>
        <ul class="order-list">
          {#each replacementOptions as opt (opt.id)}
            {@const pos = positionFor(opt.id)}
            {@const src = sourceCardName(opt)}
            <li>
              <button
                type="button"
                class="order-row"
                class:selected={pos > 0}
                onclick={() => toggleReplacement(opt.id)}
                aria-pressed={pos > 0}
                aria-label={`${pos > 0 ? "deselect" : "select"} ${opt.label || "effect"}`}
              >
                <span class="order-pos">{pos > 0 ? pos : "·"}</span>
                <span class="order-label">
                  <strong>{opt.label || "Replacement effect"}</strong>
                  {#if src}<span class="order-src">{src}</span>{/if}
                </span>
              </button>
            </li>
          {/each}
        </ul>
        <div class="footer">
          <span class="counter">{ordered.length} / {replacementOptions.length} ordered</span>
          <button
            type="button"
            class="submit"
            disabled={ordered.length !== replacementOptions.length}
            onclick={submitReplacementOrder}
          >
            Apply in this order
          </button>
        </div>
      {:else}
        <h2 id="choice-title">
          {active.reason || "Choose"} — pick {active.count} card{active.count === 1 ? "" : "s"}
        </h2>
        <p class="hint">
          {#if isSelfSource}
            Pick {active.count} card{active.count === 1 ? "" : "s"} from your hand to discard.
          {:else}
            Pick {active.count} card{active.count === 1 ? "" : "s"} from
            <strong>{fromName}</strong>'s revealed hand.
            <strong>{fromName}</strong> will discard your pick{active.count === 1 ? "" : "s"}.
          {/if}
        </p>
        <div class="card-grid">
          {#each optionCards as c (c.instance_id)}
            <button
              type="button"
              class="card-pick"
              class:selected={selected.has(c.instance_id)}
              disabled={!selected.has(c.instance_id) && selected.size >= active.count}
              onclick={() => toggle(c.instance_id)}
              aria-pressed={selected.has(c.instance_id)}
              aria-label={`select ${c.name || "card"}`}
            >
              <Card card={c} />
            </button>
          {/each}
        </div>
        <div class="footer">
          <span class="counter">{selected.size} / {active.count} selected</span>
          <button
            type="button"
            class="submit"
            disabled={selected.size !== active.count}
            onclick={submit}
          >
            Confirm
          </button>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgba(4, 8, 16, 0.7);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
    animation: fade-in 160ms var(--ease);
  }
  @keyframes fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  .modal {
    background: linear-gradient(180deg, var(--surface) 0%, var(--bg-2) 100%);
    border: 1px solid rgba(122, 167, 255, 0.22);
    border-radius: var(--radius-xl);
    padding: 22px 26px;
    max-width: 760px;
    max-height: 86vh;
    overflow: auto;
    box-shadow:
      0 30px 80px rgba(0, 0, 0, 0.7),
      0 0 0 1px rgba(0, 0, 0, 0.4),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
    animation: modal-in 220ms var(--ease);
  }
  @keyframes modal-in {
    from {
      opacity: 0;
      transform: translateY(12px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
  h2 {
    margin: 0 0 6px;
    font-size: 18px;
    letter-spacing: -0.01em;
    color: var(--gold);
    text-transform: none;
    font-weight: 700;
  }
  .hint {
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.4;
    margin: 0 0 14px;
  }
  .hint strong {
    color: var(--fg);
    font-weight: 600;
  }
  .card-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(96px, 1fr));
    gap: 8px;
  }
  .card-pick {
    background: transparent;
    border: 2px solid transparent;
    border-radius: var(--radius);
    padding: 3px;
    cursor: pointer;
    box-shadow: none;
    transition:
      border-color 120ms var(--ease),
      transform 120ms var(--ease),
      box-shadow 120ms var(--ease);
  }
  .card-pick:hover:not(:disabled) {
    border-color: var(--accent);
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.4);
    transform: translateY(-2px);
  }
  .card-pick.selected {
    border-color: var(--gold);
    box-shadow:
      0 0 18px rgba(255, 208, 122, 0.55),
      0 6px 18px rgba(0, 0, 0, 0.4);
  }
  .card-pick:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 16px;
    padding-top: 14px;
    border-top: 1px solid rgba(255, 255, 255, 0.06);
    gap: 12px;
  }
  .counter {
    font-size: 12px;
    color: var(--fg-muted);
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
  }
  .submit {
    padding: 8px 22px;
    border-radius: 999px;
    background: linear-gradient(180deg, #ffe59a 0%, #e6b85f 100%);
    color: #231806;
    border: 1px solid rgba(255, 230, 160, 0.6);
    font-weight: 800;
    letter-spacing: 0.02em;
    cursor: pointer;
    box-shadow:
      0 6px 18px rgba(255, 208, 122, 0.25),
      inset 0 1px 0 rgba(255, 255, 255, 0.4);
  }
  .submit:hover:not(:disabled) {
    filter: brightness(1.04);
  }
  .submit:disabled {
    opacity: 0.4;
    cursor: not-allowed;
    box-shadow: none;
  }
  .color-row {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 4px;
  }
  .color-pick {
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    padding: 14px 18px;
    background: var(--fill);
    color: #0a0e1a;
    border: 2px solid rgba(0, 0, 0, 0.4);
    border-radius: 10px;
    font-weight: 800;
    cursor: pointer;
    min-width: 88px;
    box-shadow:
      0 6px 16px rgba(0, 0, 0, 0.4),
      inset 0 1px 0 rgba(255, 255, 255, 0.25);
    transition:
      transform 120ms var(--ease),
      box-shadow 120ms var(--ease),
      filter 120ms var(--ease);
  }
  .color-pick:hover,
  .color-pick:focus-visible {
    transform: translateY(-2px);
    box-shadow:
      0 12px 24px rgba(0, 0, 0, 0.55),
      inset 0 1px 0 rgba(255, 255, 255, 0.25);
    filter: brightness(1.05);
    outline: none;
  }
  .color-letter {
    font-size: 22px;
    line-height: 1;
  }
  .color-name {
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    opacity: 0.8;
  }
  .order-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 420px;
  }
  .order-row {
    display: flex;
    align-items: center;
    gap: 14px;
    width: 100%;
    padding: 12px 14px;
    background: rgba(255, 255, 255, 0.04);
    border: 2px solid rgba(255, 255, 255, 0.08);
    border-radius: 10px;
    cursor: pointer;
    color: var(--fg);
    text-align: left;
    font: inherit;
    transition:
      border-color 120ms var(--ease),
      background 120ms var(--ease),
      transform 120ms var(--ease);
  }
  .order-row:hover {
    border-color: var(--accent);
    background: rgba(122, 167, 255, 0.08);
  }
  .order-row.selected {
    border-color: var(--gold);
    background: rgba(255, 208, 122, 0.08);
    box-shadow: 0 0 0 1px rgba(255, 208, 122, 0.25);
  }
  .order-pos {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: rgba(0, 0, 0, 0.35);
    color: var(--gold);
    font-weight: 800;
    font-variant-numeric: tabular-nums;
    font-size: 15px;
    flex-shrink: 0;
  }
  .order-row.selected .order-pos {
    background: linear-gradient(180deg, #ffe59a 0%, #e6b85f 100%);
    color: #231806;
  }
  .order-label {
    display: flex;
    flex-direction: column;
    gap: 2px;
    line-height: 1.3;
  }
  .order-label strong {
    font-weight: 700;
    font-size: 14px;
  }
  .order-src {
    color: var(--fg-muted);
    font-size: 12px;
  }
</style>
