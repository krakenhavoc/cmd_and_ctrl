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
  //
  // ADR 0111 PR 6: a sheet in the action dock; the confirm and Cancel
  // are the dock's action bar (Enter / Escape through its one key
  // handler).

  import { fetchAutoTapPreview, type AutoTapPreview } from "../../api";
  import type { AutoTapCastParams } from "../../castPreview";
  import { chooseDelveForMe, delveLimit, delveOptionIDs } from "../../delve";
  import type { CardView } from "../../protocol";
  import { cancelAction, confirmAction } from "../../dock";
  import DockSheet from "./DockSheet.svelte";

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
</script>

{#if card}
  <DockSheet
    label={card.name}
    src="delve · CR 702.66"
    width={560}
    sheetKey={`delve:${card.instance_id}`}
    count={`${chosen.length} / ${limit} exiled`}
    primary={confirmAction(chosen.length === 0 ? "Exile nothing" : "Exile", confirm)}
    secondary={[
      ...(limit > 0 ? [{ id: "choose-for-me", label: "Choose for me", onPress: chooseForMe }] : []),
      cancelAction(onCancel),
    ]}
  >
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
  </DockSheet>
{/if}

<style>
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
</style>
