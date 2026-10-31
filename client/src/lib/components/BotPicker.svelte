<script lang="ts">
  import { onDestroy, untrack } from "svelte";
  import { addBotSeat, type BotOptions } from "../api";
  import { lobbyErrorMessage } from "../serverFull";
  import {
    defaultBotTier,
    ownKeyTierChoices,
    tierForOwnKey,
    tierLabel,
    tierNote,
    unavailableNotes,
  } from "../botTier";
  import { botDeckTile, FILTER_FROM, sameName, type PickerCommit } from "../deckPicker";
  import type { PrebuiltDeck } from "../prebuiltDecks";
  import { L } from "../labels";
  import DeckGrid from "./DeckGrid.svelte";
  import ShopLinks from "./ShopLinks.svelte";

  // BotPicker — Add bot, in the picker's place under the seats (ADR 0212
  // owner decision 8), titled with the seat the bot will take: a tier as
  // a segmented control (a tier the server cannot run is disabled, and
  // the line under it says why), the own-key switch and model field when
  // GET /bot/options offers them (ADR 0182), and the Pre-built grid,
  // since the bot decks are the pre-built decks. The two <select>s it
  // replaces are gone.
  //
  // Like DeckPicker, it sends nothing until its caller's primary says so:
  // it reports what adding would do (`onchange`) and the lobby card's
  // footer primary calls `commit()`.

  interface Props {
    gameID: string;
    /** The seat the bot takes, 0-based: the first open one, as the server seats it. */
    seat: number;
    options: BotOptions;
    /** GET /decks, for each bot deck's coverage and its list. */
    prebuilt: PrebuiltDeck[];
    onadded: () => void;
    oncancel: () => void;
    onchange?: (commit: PickerCommit | null) => void;
  }

  const { gameID, seat, options, prebuilt, onadded, oncancel, onchange }: Props = $props();

  const uid = $props.id();

  // The picker opens on the default tier (heuristic, never random) and
  // the first deck. A server with no tier of its own to offer starts on
  // the player's key, when there is one with games left today.
  const startOnKey = untrack(
    () => defaultBotTier(options.tiers) === "" && options.own_key?.games_left === true,
  );
  let ownKey = $state(startOnKey);
  let tier = $state(
    untrack(() => (startOnKey ? tierForOwnKey("") : defaultBotTier(options.tiers))),
  );
  let deck = $state(untrack(() => options.decks[0]?.id ?? ""));
  let model = $state(untrack(() => options.own_key?.model ?? ""));
  let query = $state("");
  let busy = $state(false);
  let error = $state("");

  const choices = $derived(ownKey ? ownKeyTierChoices(options.tiers) : options.tiers);
  const offNotes = $derived(ownKey ? [] : unavailableNotes(options.tiers));
  const note = $derived(tierNote(options.tiers, tier));
  const tiles = $derived(
    options.decks.map((d) =>
      botDeckTile(
        d,
        prebuilt.find((p) => p.id === d.id),
      ),
    ),
  );
  const current = $derived(options.decks.find((d) => d.id === deck));
  const cards = $derived(prebuilt.find((p) => p.id === deck)?.cards ?? []);

  function toggleOwnKey(on: boolean): void {
    ownKey = on;
    tier = on ? tierForOwnKey(tier) : defaultBotTier(options.tiers);
  }

  const commitState = $derived.by((): PickerCommit => {
    const tierName = choices.find((t) => t.tier === tier)?.label ?? "";
    return {
      label: busy ? "Adding…" : L.addBot,
      disabled: busy || !tier || !deck,
      title: !tier
        ? "no bot tier can be seated on this server"
        : !deck
          ? "pick the bot's deck"
          : `seat a ${tierName} bot with ${current?.name ?? "this deck"} in seat ${seat + 1}`,
      busy,
    };
  });

  $effect(() => {
    const c = commitState;
    untrack(() => onchange?.(c));
  });
  onDestroy(() => onchange?.(null));

  /** commit adds the bot; the lobby card's footer primary calls it. */
  export async function commit(): Promise<void> {
    if (commitState.disabled) return;
    busy = true;
    error = "";
    try {
      await addBotSeat(
        gameID,
        ownKey ? { tier, deck, own_key: true, model: model.trim() || undefined } : { tier, deck },
      );
      onadded();
    } catch (err) {
      error = lobbyErrorMessage(err, "could not add the bot");
    } finally {
      busy = false;
    }
  }
</script>

<section class="bot-picker" aria-labelledby={`${uid}-title`}>
  <div class="phead">
    <h4 class="ptitle" id={`${uid}-title`}>{L.addBotTo(seat + 1)}</h4>
    <button type="button" class="btn ghost sm" disabled={busy} onclick={oncancel}>Cancel</button>
  </div>

  <div class="tiers">
    <div class="seg" role="radiogroup" aria-label={L.botTier}>
      {#each choices as t (t.tier)}
        <label
          class="opt"
          class:on={tier === t.tier}
          class:off={!t.available}
          title={t.available ? t.description : t.reason || "not available on this server"}
        >
          <input
            type="radio"
            name={`${uid}-tier`}
            value={t.tier}
            bind:group={tier}
            disabled={busy || !t.available}
          />
          {tierLabel(t)}
        </label>
      {/each}
    </div>
    {#if options.own_key}
      <label class="own-key">
        <input
          type="checkbox"
          checked={ownKey}
          disabled={busy || (!options.own_key.games_left && !ownKey)}
          onchange={(e) => toggleOwnKey(e.currentTarget.checked)}
        />
        <span>Run on my model key ({options.own_key.provider_label})</span>
      </label>
    {/if}
  </div>
  {#if note}
    <p class="note" class:warn={note.warning}>{note.text}</p>
  {/if}
  {#each offNotes as n (n)}
    <p class="note">{n}</p>
  {/each}
  {#if options.own_key && !options.own_key.games_left}
    <p class="note warn">
      Your model key has played all the games you allowed it today (Settings › AI opponents).
    </p>
  {/if}
  {#if ownKey && options.own_key}
    <label class="model">
      <span>Model</span>
      <input
        type="text"
        autocomplete="off"
        spellcheck="false"
        maxlength="128"
        placeholder={options.own_key.model}
        bind:value={model}
        disabled={busy}
      />
    </label>
  {/if}

  {#if options.decks.length > FILTER_FROM}
    <input
      class="filter"
      type="search"
      placeholder="Filter the pre-built decks"
      aria-label="filter the bot's decks"
      bind:value={query}
    />
  {/if}
  <DeckGrid {tiles} name="bot-deck" label={L.prebuiltDecks} bind:selected={deck} {query} />

  {#if current}
    <div class="detail">
      <p class="dline">
        <b>{current.name}</b
        >{#if current.commander && !sameName(current.name, current.commander)}{" · "}{current.commander}{/if}{#if current.description}{" · "}{current.description}{/if}
      </p>
    </div>
    {#if cards.length > 0}
      <ShopLinks list={cards} label="Play this deck in paper" />
    {/if}
  {/if}

  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <p class="note">
    A bot takes a real seat, arrives with its deck already loaded, and starts playing when the game
    starts.
  </p>
</section>

<style>
  .bot-picker {
    display: flex;
    flex-direction: column;
    gap: 10px;
    min-width: 0;
  }
  .phead {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }
  .ptitle {
    margin: 0;
    font: 700 15px/1.3 var(--font-display);
    letter-spacing: -0.01em;
    color: var(--fg);
    text-transform: none;
  }
  .tiers {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px 16px;
  }
  /* The tier as a segmented control: real radios, so the arrow keys move
     it, drawn as one row of segments. */
  .seg {
    display: inline-flex;
    flex-wrap: wrap;
    overflow: hidden;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
  }
  .opt {
    position: relative;
    display: inline-flex;
    align-items: center;
    height: 30px;
    padding: 0 11px;
    color: var(--fg-muted);
    font: 600 12.5px/1 var(--font-ui);
    white-space: nowrap;
    cursor: pointer;
  }
  .opt + .opt {
    border-left: 1px solid var(--border);
  }
  .opt:hover:not(.off) {
    color: var(--fg);
  }
  .opt.on {
    background: var(--surface-raised);
    color: var(--fg);
  }
  .opt.off {
    opacity: 0.45;
    cursor: not-allowed;
  }
  /* The radio covers its segment, unseen, so a press lands on it. */
  .opt input {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    margin: 0;
    opacity: 0;
    cursor: inherit;
  }
  .opt:has(input:focus-visible) {
    outline: 2px solid var(--accent-strong);
    outline-offset: -2px;
  }
  .own-key {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 12.5px;
    color: var(--fg);
  }
  .own-key input {
    margin: 0;
  }
  .model {
    display: flex;
    flex-direction: column;
    gap: 4px;
    max-width: 320px;
    font-size: 11.5px;
    color: var(--fg-muted);
  }
  .model input {
    margin: 0;
    font-size: 12.5px;
  }
  .note {
    margin: 0;
    font-size: 12px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .note.warn {
    color: var(--danger);
  }
  .filter {
    align-self: flex-end;
    width: 160px;
    height: 28px;
    margin: 0;
    padding: 0 10px;
    box-sizing: border-box;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    background: var(--surface-sunken);
    color: var(--fg);
    font: inherit;
    font-size: 12.5px;
  }
  .detail {
    padding-top: 10px;
    border-top: 1px solid var(--border);
  }
  .dline {
    margin: 0;
    font-size: 12.5px;
    line-height: 1.45;
    color: var(--fg-muted);
  }
  .dline b {
    color: var(--fg);
  }
  .error {
    margin: 0;
    font-size: 12px;
    color: var(--danger);
  }
</style>
