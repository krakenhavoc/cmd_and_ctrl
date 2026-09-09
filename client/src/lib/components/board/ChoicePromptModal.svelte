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
    ActionType,
    CardView,
    DamageAssignmentView,
    GameView,
    PendingChoiceView,
    ReplacementOptionView,
  } from "../../protocol";
  import Card from "./Card.svelte";

  interface Props {
    snap: GameView;
    viewerID: string | null;
    sendAction: (type: ActionType, params?: unknown, player?: string) => void;
  }

  const { snap, viewerID, sendAction }: Props = $props();

  // First choice addressed to the viewer. Queue ordering: front of
  // list is "what the chooser sees next." One modal at a time; when
  // they resolve this one, the next pops automatically on the
  // snapshot after the server drains the entry.
  const active = $derived.by((): PendingChoiceView | null => {
    if (!viewerID || !snap.pending_choices) return null;
    for (const c of snap.pending_choices) {
      // S20 sub-PR 2: pick_target is answered by clicking the board
      // (Board.svelte drives the targeting store), not by a modal.
      if (c.kind === "pick_target") continue;
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
  // S19 sub-PR 8 trigger_order — CR 603.3b: the same reorder list,
  // fed from trigger_options. Submitted order = resolution order
  // (top of the list resolves first); the server stacks in reverse.
  const isTriggerOrder = $derived(active?.kind === "trigger_order");
  const replacementOptions = $derived<ReplacementOptionView[]>(
    active?.kind === "trigger_order"
      ? (active?.trigger_options ?? [])
      : (active?.replacement_options ?? []),
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
    sendAction("resolve_choice", { choice_id: active.id, order: ordered }, viewerID);
  }

  function positionFor(id: string): number {
    return ordered.indexOf(id) + 1; // 1-indexed; 0 = unselected
  }

  function sourceCardName(opt: ReplacementOptionView): string {
    if (!opt.source_card_id || !snap.battlefield) return "";
    for (const c of snap.battlefield.cards) {
      if (c.instance_id === opt.source_card_id) return c.name ?? "";
    }
    // S19 dies-triggers: the source has already left for a
    // graveyard / exile by the time an ordering prompt shows.
    const elsewhere = triggerSourceName(opt.source_card_id);
    return elsewhere === "Triggered ability" ? "" : elsewhere;
  }

  // S17 sub-PR 6 optional-replacement branch — CR 614.10 "may"
  // prompt. Used by CR 903.9 commander-zone replacement today:
  // commander's owner picks yes (route to command zone) or no
  // (let the event proceed to graveyard/exile/hand/library).
  const isOptionalReplacement = $derived(active?.kind === "optional_replacement");

  // S19 sub-PR 2 trigger-prompt branch — CR 603.4 "you may" prompt
  // for an optional triggered ability. Same {choice_id, apply}
  // payload as optional-replacement; the server routes to
  // ResolveTriggerPrompt vs ResolveOptionalReplacement by inspecting
  // the choice's kind.
  const isTriggerPrompt = $derived(active?.kind === "trigger_prompt");

  // S19 sub-PR 6 pay-unless branch — CR 118.12 "unless that player
  // pays {N}" (Rhystic Study, Smothering Tithe, Esper Sentinel).
  // The chooser is the player being taxed, not the card's
  // controller. Same {choice_id, apply} payload; the server routes
  // to ResolvePayUnless by kind. "Pay" spends from the pool and
  // auto-taps untapped sources if the pool is short; a "Pay" the
  // player can't cover degrades to a decline server-side.
  const isPayUnless = $derived(active?.kind === "pay_unless");

  // S19 follow-up: the server flags optional triggers whose effect
  // has no legal target (Reclamation Sage with no opponent artifact,
  // Eternal Witness with an empty graveyard). Until the S20 target
  // picker lands, the auto-targeter silently no-ops in that case —
  // which reads as a bug. Warn the chooser and relabel "Yes".
  const noLegalTarget = $derived(active?.no_legal_target === true);

  // Y / N answer the yes-no prompts (optional replacement, may-
  // trigger, pay-unless) from the keyboard; the footer shows the
  // hint. Ignored while typing in a field.
  const isYesNo = $derived(isOptionalReplacement || isTriggerPrompt || isPayUnless);
  function handleKey(e: KeyboardEvent): void {
    if (!open || !isYesNo) return;
    const t = e.target as HTMLElement | null;
    if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable)) return;
    if (e.key === "y" || e.key === "Y") {
      e.preventDefault();
      answerOptional(true);
    } else if (e.key === "n" || e.key === "N") {
      e.preventDefault();
      answerOptional(false);
    }
  }

  function answerOptional(apply: boolean): void {
    if (!active || !viewerID) return;
    sendAction("resolve_choice", { choice_id: active.id, apply }, viewerID);
  }

  // sourceCardName resolves the source-card display name for a
  // trigger prompt. Walks battlefield + every seated player's
  // graveyard + the shared exile zone — LTB triggers prompt after
  // the source has already moved off the battlefield, so the
  // lookup has to span destination zones. Falls back to a generic
  // string when the card isn't visible to the viewer (redacted
  // CardView entries arrive with empty names).
  function triggerSourceName(sourceID?: string): string {
    if (!sourceID) return "Triggered ability";
    const seek = (cards: CardView[] | undefined) => {
      if (!cards) return undefined;
      for (const c of cards) {
        if (c.instance_id === sourceID && c.name) return c.name;
      }
      return undefined;
    };
    const fromBF = seek(snap.battlefield?.cards);
    if (fromBF) return fromBF;
    const fromExile = seek(snap.exile?.cards);
    if (fromExile) return fromExile;
    for (const p of snap.seats ?? []) {
      const fromGY = seek(p.graveyard?.cards);
      if (fromGY) return fromGY;
    }
    return "Triggered ability";
  }

  // S18 damage_assignment branch — CR 510.1c multi-blocker combat
  // damage prompt. The attacker's controller reorders the blockers
  // (drag via up/down buttons since drag-and-drop UX lives outside
  // this sprint) and assigns damage per blocker. Server validates
  // at-least-lethal prefix + total = attacker_power, and trample
  // overflow if AllowTrample.
  const isDamageAssignment = $derived(active?.kind === "damage_assignment");
  const damageFrame = $derived<DamageAssignmentView | null>(active?.damage_assignment ?? null);

  // Ordered blocker IDs (reorderable). Damage amount per blocker,
  // keyed by blocker ID. Trample-to-player bucket.
  let blockerOrder = $state<string[]>([]);
  let damageAmounts = $state<Record<string, number>>({});
  let trampleToPlayer = $state(0);

  // Reset assignment state when the prompt's identity changes — the
  // same untracked last-id pattern as the selected/ordered reset
  // above. Keying on content equality with blockerOrder here would
  // make the user's own ▲/▼ reorder re-trigger the effect and revert
  // their order (and zero their amounts) on the first click.
  let lastDamageChoiceID: string | null = null;
  $effect(() => {
    const nextID = active?.id ?? null;
    if (nextID === lastDamageChoiceID) return;
    lastDamageChoiceID = nextID;
    if (!damageFrame) return;
    blockerOrder = [...damageFrame.blocker_card_ids];
    const next: Record<string, number> = {};
    for (const id of damageFrame.blocker_card_ids) next[id] = 0;
    damageAmounts = next;
    trampleToPlayer = 0;
  });

  const assignedTotal = $derived(
    blockerOrder.reduce((acc, id) => acc + (damageAmounts[id] ?? 0), 0) + trampleToPlayer,
  );

  const canSubmitAssignment = $derived(
    damageFrame !== null && assignedTotal === damageFrame.attacker_power,
  );

  function moveBlocker(id: string, delta: -1 | 1): void {
    const idx = blockerOrder.indexOf(id);
    if (idx < 0) return;
    const target = idx + delta;
    if (target < 0 || target >= blockerOrder.length) return;
    const next = [...blockerOrder];
    [next[idx], next[target]] = [next[target], next[idx]];
    blockerOrder = next;
  }

  function setDamageAmount(id: string, raw: string): void {
    const n = Math.max(0, Math.floor(Number(raw) || 0));
    damageAmounts = { ...damageAmounts, [id]: n };
  }

  function setTrampleAmount(raw: string): void {
    const n = Math.max(0, Math.floor(Number(raw) || 0));
    trampleToPlayer = n;
  }

  function blockerName(id: string): string {
    if (!snap.battlefield) return id.slice(0, 8);
    const card = snap.battlefield.cards.find((c) => c.instance_id === id);
    return card?.name ?? id.slice(0, 8);
  }

  function attackerName(id: string): string {
    if (!snap.battlefield) return id.slice(0, 8);
    const card = snap.battlefield.cards.find((c) => c.instance_id === id);
    return card?.name ?? id.slice(0, 8);
  }

  function submitDamageAssignment(): void {
    if (!active || !viewerID || !damageFrame) return;
    if (!canSubmitAssignment) return;
    const assignments = blockerOrder.map((id) => ({
      blocker_id: id,
      amount: damageAmounts[id] ?? 0,
    }));
    sendAction(
      "resolve_choice",
      {
        choice_id: active.id,
        assignments,
        trample_to_player: trampleToPlayer,
      },
      viewerID,
    );
  }
</script>

{#if open && active}
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="choice-title">
    <div class="prompt-modal">
      {#if isManaPick}
        <h2 id="choice-title">
          {active.reason || "Pick a color"}
          <span class="prompt-src" aria-hidden="true">mana ability</span>
        </h2>
        <p class="prompt-hint">Choose a color to add to your mana pool.</p>
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
      {:else if isOptionalReplacement}
        <h2 id="choice-title">
          {active.reason || "Apply replacement?"}
          <span class="prompt-src" aria-hidden="true">optional replacement · CR 614.10</span>
        </h2>
        <p class="prompt-hint">
          You (the affected player) decide whether this substitution applies.
        </p>
        <div class="prompt-foot">
          <span class="prompt-count"><span class="kbd">Y</span> / <span class="kbd">N</span></span>
          <button type="button" onclick={() => answerOptional(false)}>No</button>
          <button type="button" class="primary" onclick={() => answerOptional(true)}>Yes</button>
        </div>
      {:else if isTriggerPrompt}
        <h2 id="choice-title">
          {active.reason || `${triggerSourceName(active.source)} triggered`}
          <span class="prompt-src" aria-hidden="true">may trigger · CR 603.4</span>
        </h2>
        {#if noLegalTarget}
          <p class="prompt-hint warn">
            No legal target — “Yes” passes without effect (picker lands in S20).
          </p>
        {:else}
          <p class="prompt-hint">Fire the ability, or let it pass without effect.</p>
        {/if}
        <div class="prompt-foot">
          <span class="prompt-count"><span class="kbd">Y</span> / <span class="kbd">N</span></span>
          <button type="button" onclick={() => answerOptional(false)}>No</button>
          <button type="button" class="primary" onclick={() => answerOptional(true)}>Yes</button>
        </div>
      {:else if isPayUnless}
        <h2 id="choice-title">
          {active.reason || `${triggerSourceName(active.source)} — pay ${active.pay_cost ?? ""}?`}
          <span class="prompt-src" aria-hidden="true">pay unless</span>
        </h2>
        <p class="prompt-hint">
          Pay {active.pay_cost ?? "the cost"} from your pool (untapped sources auto-tap if it's short),
          or don't and let {triggerSourceName(active.source)} do its thing.
        </p>
        <div class="prompt-foot">
          <span class="prompt-count"><span class="kbd">Y</span> / <span class="kbd">N</span></span>
          <button type="button" onclick={() => answerOptional(false)}>Don't pay</button>
          <button type="button" class="primary" onclick={() => answerOptional(true)}>
            Pay {active.pay_cost ?? ""}
          </button>
        </div>
      {:else if isDamageAssignment && damageFrame}
        <h2 id="choice-title">
          {active.reason || "Assign combat damage"}
          <span class="prompt-src" aria-hidden="true">CR 510.1c</span>
        </h2>
        <p class="prompt-hint">
          <strong>{attackerName(damageFrame.attacker_card_id)}</strong>
          is blocked by {damageFrame.blocker_card_ids.length} creatures. Order them and divide
          {damageFrame.attacker_power} damage — earlier blockers must be dealt at-least-lethal before
          the next gets any.
          {#if damageFrame.allow_trample}
            Trample lets leftover damage spill to the defending player.
          {/if}
          {#if damageFrame.has_deathtouch}
            Deathtouch makes 1 damage lethal.
          {/if}
        </p>
        <ul class="assign-list">
          {#each blockerOrder as id, i (id)}
            <li class="assign-row">
              <div class="assign-order">
                <button
                  type="button"
                  class="reorder-btn"
                  disabled={i === 0}
                  onclick={() => moveBlocker(id, -1)}
                  aria-label={`move ${blockerName(id)} up`}
                >
                  ▲
                </button>
                <span class="assign-pos">{i + 1}</span>
                <button
                  type="button"
                  class="reorder-btn"
                  disabled={i === blockerOrder.length - 1}
                  onclick={() => moveBlocker(id, 1)}
                  aria-label={`move ${blockerName(id)} down`}
                >
                  ▼
                </button>
              </div>
              <span class="assign-name">{blockerName(id)}</span>
              <label class="assign-input">
                <span class="sr-only">damage to {blockerName(id)}</span>
                <input
                  type="number"
                  min="0"
                  max={damageFrame.attacker_power}
                  value={damageAmounts[id] ?? 0}
                  oninput={(e) => setDamageAmount(id, (e.currentTarget as HTMLInputElement).value)}
                />
              </label>
            </li>
          {/each}
          {#if damageFrame.allow_trample}
            <li class="assign-row trample">
              <div class="assign-order"><span class="assign-pos">→</span></div>
              <span class="assign-name">Defending player (trample)</span>
              <label class="assign-input">
                <span class="sr-only">trample damage to defending player</span>
                <input
                  type="number"
                  min="0"
                  max={damageFrame.attacker_power}
                  value={trampleToPlayer}
                  oninput={(e) => setTrampleAmount((e.currentTarget as HTMLInputElement).value)}
                />
              </label>
            </li>
          {/if}
        </ul>
        <div class="prompt-foot">
          <span class="prompt-count">{assignedTotal} / {damageFrame.attacker_power} assigned</span>
          <button
            type="button"
            class="primary"
            disabled={!canSubmitAssignment}
            onclick={submitDamageAssignment}
          >
            Deal damage
          </button>
        </div>
      {:else if isReplacementOrder || isTriggerOrder}
        <h2 id="choice-title">
          {active.reason || (isTriggerOrder ? "Order your triggers" : "Order replacement effects")}
          <span class="prompt-src" aria-hidden="true"
            >{isTriggerOrder ? "CR 603.3b" : "CR 616"}</span
          >
        </h2>
        <p class="prompt-hint">
          {#if isTriggerOrder}
            Two or more of your abilities triggered at once. Click them in the order they should
            resolve — the first you pick resolves first.
          {:else}
            Click each effect in the order it should apply. Different orders can produce different
            results — you choose as the affected player.
          {/if}
        </p>
        <ul class="prompt-options">
          {#each replacementOptions as opt (opt.id)}
            {@const pos = positionFor(opt.id)}
            {@const src = sourceCardName(opt)}
            <li>
              <button
                type="button"
                class="prompt-opt"
                class:on={pos > 0}
                onclick={() => toggleReplacement(opt.id)}
                aria-pressed={pos > 0}
                aria-label={`${pos > 0 ? "deselect" : "select"} ${opt.label || "effect"}`}
              >
                <span class="prompt-num">{pos > 0 ? pos : "·"}</span>
                <span class="order-label">
                  <strong>{opt.label || "Replacement effect"}</strong>
                  {#if src}<span class="order-src">{src}</span>{/if}
                </span>
              </button>
            </li>
          {/each}
        </ul>
        <div class="prompt-foot">
          <span class="prompt-count">{ordered.length} / {replacementOptions.length} ordered</span>
          <button
            type="button"
            class="primary"
            disabled={ordered.length !== replacementOptions.length}
            onclick={submitReplacementOrder}
          >
            {isTriggerOrder ? "Resolve in this order" : "Apply in this order"}
          </button>
        </div>
      {:else}
        <h2 id="choice-title">
          {active.reason || "Choose"} — pick {active.count} card{active.count === 1 ? "" : "s"}
          <span class="prompt-src" aria-hidden="true">{isSelfSource ? "discard" : "reveal"}</span>
        </h2>
        <p class="prompt-hint">
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
        <div class="prompt-foot">
          <span class="prompt-count">{selected.size} / {active.count} selected</span>
          <button
            type="button"
            class="primary"
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

<svelte:window onkeydown={handleKey} />

<style>
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
    border-color: var(--border-strong);
    background: transparent;
    transform: translateY(-2px);
  }
  .card-pick.selected {
    border-color: var(--gold);
    box-shadow: 0 0 16px rgba(217, 180, 92, 0.35);
  }
  .card-pick:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .color-row {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
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
    border: 1px solid rgba(0, 0, 0, 0.35);
    border-radius: 10px;
    font-weight: 800;
    cursor: pointer;
    min-width: 88px;
    box-shadow: var(--shadow-sm);
    transition:
      transform 120ms var(--ease),
      box-shadow 120ms var(--ease),
      filter 120ms var(--ease);
  }
  .color-pick:hover,
  .color-pick:focus-visible {
    background: var(--fill);
    border-color: rgba(0, 0, 0, 0.35);
    transform: translateY(-2px);
    box-shadow: var(--shadow);
    filter: brightness(1.05);
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
  .order-label {
    display: flex;
    flex-direction: column;
    gap: 2px;
    line-height: 1.3;
  }
  .order-label strong {
    font-weight: 600;
    font-size: 13px;
  }
  .order-src {
    color: var(--fg-muted);
    font-size: 11.5px;
  }
  .assign-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: min(440px, calc(100vw - 80px));
  }
  .assign-row {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: 12px;
    padding: 8px 12px;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    border-radius: 10px;
  }
  .assign-row.trample {
    border-color: rgba(255, 107, 107, 0.35);
  }
  .assign-order {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .assign-pos {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    border-radius: 6px;
    background: var(--surface-raised);
    color: var(--fg-dim);
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 700;
  }
  .reorder-btn {
    padding: 2px 7px;
    font-size: 10px;
    border-radius: 6px;
  }
  .assign-name {
    font-weight: 600;
    font-size: 13px;
  }
  .assign-input input {
    width: 70px;
    padding: 5px 10px;
    text-align: right;
    font-family: var(--font-mono);
    font-size: 13px;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
</style>
