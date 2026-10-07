<script lang="ts">
  // S15 sub-PR 5 auto-tap-and-cast preview modal.
  //
  // Opens when the viewer chooses "Auto-tap & cast" from the
  // insufficient-mana request in the action dock (or any future right-click → cast
  // affordance). On mount it fetches `/games/:id/auto-tap-preview`
  // for the chosen card and renders the proposed plan: the ordered
  // list of sources the server would spend, each with what paying
  // with it costs — a tap, a sacrifice, or (#1285) a card exiled out
  // of the hand, which is named rather than silently missing from a
  // battlefield lookup. Enter confirms
  // (re-fires cast_spell with auto_tap: true); ESC cancels.
  //
  // ADR 0111 PR 6: a sheet in the action dock, not a modal. Cast is the
  // bar's primary (Enter) and Cancel its secondary (Escape), through the
  // dock's one key handler; the document listener this had is gone.
  //
  // Lock-tap UI: each row in the proposed plan has a "lock" toggle.
  // Clicking it adds the source to the excluded set and re-fetches
  // — letting the player reserve a land for a later cast and watch
  // the planner re-route around it. When the planner can no longer
  // satisfy the cost, the modal shows the missing-symbols list and
  // disables the confirm button.

  import { fetchAutoTapPreview, type AutoTapPreview } from "../../api";
  import { energyBadge, paymentVerb, planRows, planSummary } from "../../autoTapPlan";
  import ManaCost from "./ManaCost.svelte";
  import type { AutoTapCastParams } from "../../castPreview";
  import type { CardView, GameView } from "../../protocol";
  import { cancelAction, confirmAction } from "../../dock";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    gameID: string;
    snap: GameView;
    cardID: string | null;
    // #696: the announce-time choices of the cast being retried —
    // source zone, alternative cost, optional costs, convoke taps,
    // face. The confirm button replays the original cast payload, so
    // the preview has to price THAT cast: without these it priced the
    // card's printed cost from the hand and disabled the button on
    // every flashback, escape, overload and granted exile cast whose
    // real price was affordable.
    castParams?: AutoTapCastParams;
    onConfirm: (lockedSources: string[]) => void;
    onCancel: () => void;
  }

  const { gameID, snap, cardID, castParams = {}, onConfirm, onCancel }: Props = $props();

  let preview = $state<AutoTapPreview | null>(null);
  let loading = $state(false);
  let fetchError = $state<string | null>(null);
  let lockedSources = $state<string[]>([]);

  // Reset locked set whenever a new card opens the modal. Without
  // this, an excluded source from a prior cast would carry over
  // and silently sabotage the next plan.
  let lastCardID: string | null = null;
  $effect(() => {
    if (cardID !== lastCardID) {
      lastCardID = cardID;
      lockedSources = [];
      preview = null;
      fetchError = null;
    }
  });

  // Refetch whenever the card or the locked set changes. The server
  // is read-only on this endpoint so the round-trip is cheap; a
  // small debounce isn't necessary because lock toggles are user-
  // driven (one click → one fetch).
  //
  // fetchSeq is a monotonically increasing request id (deliberately a
  // plain let, not $state — it must not re-trigger this effect). Each
  // run claims the next id; a response only lands if it's still the
  // latest. Guarding by cardID alone isn't enough: two quick lock
  // toggles for the same card leave two fetches in flight and the
  // last to RESOLVE would win, possibly showing a plan computed for
  // an outdated lockedSources set. Bumping before the early return
  // also invalidates in-flight responses when the modal closes.
  let fetchSeq = 0;
  $effect(() => {
    const reqID = ++fetchSeq;
    if (!cardID) return;
    const card = cardID;
    const locked = lockedSources.slice();
    const cast = castParams;
    loading = true;
    fetchError = null;
    fetchAutoTapPreview(gameID, card, { excluded: locked, cast })
      .then((p) => {
        if (reqID === fetchSeq) {
          preview = p;
        }
      })
      .catch((err: unknown) => {
        if (reqID === fetchSeq) {
          fetchError = err instanceof Error ? err.message : "preview failed";
        }
      })
      .finally(() => {
        if (reqID === fetchSeq) loading = false;
      });
  });

  // A lookup from instance_id → CardView over the battlefield AND the
  // hands the viewer can see. #1285: a planned source need not be a
  // permanent — a Spirit Guide is spent out of the hand (#1228) — and
  // a locked hand source has to keep its name in the "locked" list
  // after it drops out of the plan. The server names every planned
  // source itself (`sources`); this is the fallback for a locked one
  // and for an older server.
  const cardsByID = $derived.by(() => {
    const m = new Map<string, CardView>();
    for (const c of snap.battlefield?.cards ?? []) {
      m.set(c.instance_id, c);
    }
    for (const seat of snap.seats ?? []) {
      for (const c of seat.hand?.cards ?? []) {
        if (c.name) m.set(c.instance_id, c);
      }
    }
    return m;
  });

  const rows = $derived(planRows(preview, (id) => cardsByID.get(id)?.name));

  function lockedCards(): { id: string; name: string }[] {
    return lockedSources.map((id) => {
      const card = cardsByID.get(id);
      return { id, name: card?.name ?? id.slice(0, 8) };
    });
  }

  function toggleLock(id: string): void {
    if (lockedSources.includes(id)) {
      lockedSources = lockedSources.filter((s) => s !== id);
    } else {
      lockedSources = [...lockedSources, id];
    }
  }

  function confirm(): void {
    if (!preview?.ok || loading) return;
    onConfirm(lockedSources.slice());
  }
</script>

{#if cardID}
  <DockSheet
    label="Auto-tap & cast"
    src={preview ? preview.cost || "no cost" : undefined}
    width={460}
    sheetKey={`autotap:${cardID}`}
    primary={confirmAction("Cast", confirm, { disabled: !preview?.ok || loading })}
    secondary={[cancelAction(onCancel)]}
  >
    {#if loading && !preview}
      <p class="prompt-hint">planning…</p>
    {:else if fetchError}
      <p class="prompt-hint error">preview failed: {fetchError}</p>
    {:else if preview}
      {#if preview.ok}
        <p class="prompt-hint">
          {planSummary(rows)} Lock a source to keep it for a later cast.
        </p>
        <ul class="prompt-options plan">
          {#each rows as row (row.id)}
            <li class="prompt-opt src-row">
              <span class="card-name">{row.name}</span>
              <span class="pay-verb" class:gone={row.gone} data-payment={row.payment}
                >{paymentVerb(row.payment)}</span
              >
              {#if row.energy > 0}
                <!-- ADR 0129 §5: the energy this source's ability pays. -->
                <span class="pay-energy" data-energy={row.energy}>
                  Pay <ManaCost
                    cost={energyBadge(row)}
                    size={12}
                    label={`pay ${row.energy} energy`}
                  />
                </span>
              {/if}
              <button
                type="button"
                class="ghost lock-btn"
                onclick={() => toggleLock(row.id)}
                title="reserve this source for a later cast"
              >
                lock
              </button>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="prompt-hint error">
          Insufficient mana
          {#if preview.missing && preview.missing.length > 0}
            — missing {preview.missing.join(" ")}
          {/if}
        </p>
      {/if}
      {#if lockedSources.length > 0}
        <p class="locked-label">locked sources</p>
        <ul class="prompt-options locked">
          {#each lockedCards() as row (row.id)}
            <li class="prompt-opt src-row on">
              <span class="card-name">{row.name}</span>
              <button
                type="button"
                class="ghost lock-btn"
                onclick={() => toggleLock(row.id)}
                title="release this source back to the auto-tapper"
              >
                unlock
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    {/if}
  </DockSheet>
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
    flex: 1;
  }
  /* #1285: what paying with this source costs. A sacrifice or an
     exile spends the source for good, so it reads louder than a tap. */
  .pay-verb {
    margin-right: 8px;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .pay-verb.gone {
    color: var(--danger);
    font-weight: 700;
  }
  /* ADR 0129 §5: "Pay {E}" against an energy-tier source. */
  .pay-energy {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    margin-right: 8px;
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    color: var(--fg-dim);
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
