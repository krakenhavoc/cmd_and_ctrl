<script lang="ts">
  import { onMount } from "svelte";
  import { L } from "../../labels";
  import { scryfallImageURL } from "../../cardImage";
  import {
    fetchDeckPrintings,
    fetchCardPrintings,
    saveDeckPrintings,
    type CardPrinting,
    type DeckPrintings,
  } from "../../api";
  import type { MyDeckInfo } from "../../myDecks";

  const {
    deck,
    onclose,
    onsaved,
  }: { deck: MyDeckInfo; onclose: () => void; onsaved: (deck: MyDeckInfo) => void } = $props();
  let data = $state<DeckPrintings | null>(null);
  let active = $state("");
  let options = $state<CardPrinting[]>([]);
  let choices = $state<Record<number, CardPrinting>>({});
  let filter = $state("");
  let loading = $state(true);
  let saving = $state(false);
  let error = $state("");
  let failedImages = $state<string[]>([]);
  let generation = 0;
  const row = $derived(data?.cards.find((c) => String(c.entry) === active));
  const changed = $derived(Object.keys(choices).length);
  const selected = $derived(row ? (choices[row.entry]?.id ?? row.printing_id) : "");
  const visible = $derived(
    options.filter((p) =>
      `${p.name} ${p.flavor_name ?? ""} ${p.set} ${p.collector_number}`
        .toLowerCase()
        .includes(filter.trim().toLowerCase()),
    ),
  );

  async function loadOptions(): Promise<void> {
    const request = ++generation;
    options = [];
    filter = "";
    error = "";
    if (!row || row.printing_id === "00000000-0000-0000-0000-000000000000") {
      loading = false;
      return;
    }
    loading = true;
    try {
      const found = await fetchCardPrintings(row.printing_id);
      if (request === generation) options = found;
    } catch (err) {
      if (request === generation)
        error =
          err instanceof Error
            ? err.message
            : "Could not load printings. Select the card again to retry.";
    } finally {
      if (request === generation) loading = false;
    }
  }

  onMount(() => {
    let disposed = false;
    void (async () => {
      try {
        const result = await fetchDeckPrintings(deck.id);
        if (disposed) return;
        data = result;
        active = result.cards[0] ? String(result.cards[0].entry) : "";
        await loadOptions();
      } catch (err) {
        if (!disposed) {
          error =
            err instanceof Error
              ? err.message
              : "Could not load this deck. Reopen the picker to retry.";
          loading = false;
        }
      }
    })();
    return () => {
      disposed = true;
      generation++;
    };
  });

  function choose(printing: CardPrinting): void {
    if (!row) return;
    const next = { ...choices };
    if (printing.id === row.printing_id) delete next[row.entry];
    else next[row.entry] = printing;
    choices = next;
  }

  async function save(): Promise<void> {
    if (!data || saving || !changed) return;
    saving = true;
    error = "";
    try {
      const updated = await saveDeckPrintings(
        deck.id,
        data.revision,
        Object.entries(choices).map(([entry, p]) => ({ entry: Number(entry), printing_id: p.id })),
      );
      choices = {};
      onsaved(updated);
    } catch (err) {
      error = err instanceof Error ? err.message : "Could not save printings. Try again.";
    } finally {
      saving = false;
    }
  }

  function beforeUnload(event: BeforeUnloadEvent): void {
    if (changed) {
      event.preventDefault();
      event.returnValue = "";
    }
  }
</script>

<svelte:window onbeforeunload={beforeUnload} />

<section class="picker" aria-label={L.printingPicker}>
  <h4>{L.choosePrintings}</h4>
  <p class="note">
    Choose the art for each card. Saved choices apply the next time you use this deck.
  </p>
  {#if data && data.cards.length > 0}
    <label class="field">
      <span>{L.printingCard}</span>
      <select
        aria-label={L.printingCard}
        bind:value={active}
        onchange={loadOptions}
        disabled={saving}
      >
        {#each data.cards as card (card.entry)}
          <option value={String(card.entry)}
            >{card.board} · {card.count} × {card.name}{choices[card.entry]
              ? " · changed"
              : ""}</option
          >
        {/each}
      </select>
    </label>
    {#if row}
      {#if row.count > 1}<p class="note">
          This choice applies to all {row.count} copies in this row.
        </p>{/if}
      <p class="current">
        Current: {row.set
          ? `${row.set.toUpperCase()} #${row.collector_number}`
          : "not found in the card index"}
      </p>
    {/if}
    {#if options.length > 0}
      <label class="field">
        <span>{L.filterPrintings}</span>
        <input
          type="search"
          name="printing-filter"
          aria-label={L.filterPrintings}
          bind:value={filter}
          placeholder="Set, number or alternate name…"
          autocomplete="off"
        />
      </label>
      <div class="versions" aria-busy={loading}>
        {#each visible as p (p.id)}
          <button
            type="button"
            class="version"
            class:selected={selected === p.id}
            aria-pressed={selected === p.id}
            disabled={saving}
            onclick={() => choose(p)}
          >
            {#if !failedImages.includes(p.id)}
              <img
                src={scryfallImageURL(p.id, "small")}
                alt=""
                width="146"
                height="204"
                loading="lazy"
                onerror={() => (failedImages = [...failedImages, p.id])}
              />
            {:else}
              <span class="no-image">Preview unavailable</span>
            {/if}
            <span class="version-name">{p.flavor_name || p.name}</span>
            <span class="set">{p.set.toUpperCase()} · #{p.collector_number}</span>
            <span class="selection">{selected === p.id ? "Selected" : "Choose this printing"}</span>
          </button>
        {/each}
      </div>
      {#if visible.length === 0}<p class="note">
          No printings match. Try a different set or name.
        </p>{/if}
    {:else if !loading && !error}
      <p class="note">No selectable printings were found for this card.</p>
    {/if}
  {:else if data && !loading}
    <p class="note">This deck has no cards to customize.</p>
  {/if}
  <p class="status" role="status">
    {loading
      ? "Loading printings…"
      : changed
        ? `${changed} card row${changed === 1 ? "" : "s"} changed`
        : ""}
  </p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <div class="actions">
    <button type="button" class="btn sm primary" onclick={save} disabled={saving || !changed}
      >{saving ? "Saving…" : L.savePrintings}</button
    >
    <button type="button" class="btn sm ghost" onclick={onclose} disabled={saving}
      >{L.cancelPrintings}</button
    >
  </div>
</section>

<style>
  .picker {
    border-top: 1px solid var(--border-strong);
    padding-top: 14px;
    min-width: 0;
  }
  h4 {
    margin: 0 0 6px;
    color: var(--fg);
    font-size: 15px;
  }
  .note,
  .current,
  .status {
    color: var(--fg-muted);
    font-size: 12px;
    line-height: 1.5;
    margin: 6px 0 12px;
  }
  .field {
    display: grid;
    gap: 5px;
    margin: 12px 0;
    color: var(--fg-muted);
    font-size: 12px;
  }
  input,
  select {
    width: 100%;
    min-width: 0;
    box-sizing: border-box;
    background: var(--surface-sunken);
    color: var(--fg);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-sm);
    padding: 8px;
    font: inherit;
  }
  .versions {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(136px, 1fr));
    gap: 8px;
    max-height: 440px;
    overflow-y: auto;
    overscroll-behavior: contain;
    padding: 3px;
  }
  .version {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 5px;
    min-width: 0;
    padding: 7px;
    background: var(--surface-sunken);
    color: var(--fg);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-sm);
    cursor: pointer;
    touch-action: manipulation;
    content-visibility: auto;
    contain-intrinsic-size: auto 245px;
  }
  .version:hover {
    background: var(--surface-hover);
  }
  .version.selected {
    border-color: var(--accent);
    background: var(--accent-soft);
  }
  .version:focus-visible,
  input:focus-visible,
  select:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .version img {
    width: 100%;
    height: auto;
    aspect-ratio: 146 / 204;
    object-fit: contain;
  }
  .version-name {
    font-size: 11px;
    overflow-wrap: anywhere;
  }
  .set,
  .selection {
    font-size: 10px;
    color: var(--fg-muted);
    font-variant-numeric: tabular-nums;
  }
  .selected .selection {
    color: var(--accent);
  }
  .no-image {
    aspect-ratio: 146 / 204;
    display: grid;
    place-items: center;
    font-size: 12px;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 12px;
  }
  .error {
    color: var(--danger);
    font-size: 12px;
    overflow-wrap: anywhere;
  }
</style>
