<script lang="ts">
  // The Settings "AI opponents" tab (ADR 0182): add, test, replace and
  // remove your own model API key, so the assisted and strong bots you
  // add run on your account with the provider. Settings shows the tab
  // only when the server says the feature is available (a signed-in
  // person, on a server that can store the key encrypted).
  //
  // The key field is write-only: it starts empty, is cleared after every
  // save, and the server never sends the key back.

  import {
    CUSTOM_PROVIDER,
    describeKey,
    formFor,
    keyRequired,
    modelKeyState,
    removeModelKey,
    saveModelKey,
    testModelKey,
    type ModelKeyForm,
  } from "../modelKey";

  let form = $state<ModelKeyForm>(formFor($modelKeyState));
  let busy = $state(false);
  let error = $state("");
  let status = $state("");
  let statusOk = $state(false);
  let confirmRemove = $state(false);

  const saved = $derived($modelKeyState.key);
  const providers = $derived($modelKeyState.providers);
  const limits = $derived($modelKeyState.limits);
  const provider = $derived(providers.find((p) => p.id === form.provider));
  const needsKey = $derived(keyRequired($modelKeyState, form));

  function when(iso: string | undefined): string {
    if (!iso) return "never";
    const d = new Date(iso);
    return Number.isNaN(d.getTime()) ? "—" : d.toLocaleDateString();
  }

  function message(e: unknown, fallback: string): string {
    return e instanceof Error && e.message ? e.message : fallback;
  }

  async function save(): Promise<void> {
    busy = true;
    error = "";
    status = "";
    try {
      const got = await saveModelKey(form);
      form = formFor(got);
      status = "Saved.";
      statusOk = true;
    } catch (e) {
      error = message(e, "could not save; try again");
    } finally {
      // Whatever happened, the typed key does not stay in the page.
      form.key = "";
      busy = false;
    }
  }

  async function test(): Promise<void> {
    busy = true;
    error = "";
    status = "";
    try {
      const res = await testModelKey();
      status = res.message;
      statusOk = res.ok;
    } catch (e) {
      error = message(e, "could not test the key; try again");
    } finally {
      busy = false;
    }
  }

  async function remove(): Promise<void> {
    busy = true;
    error = "";
    status = "";
    try {
      const got = await removeModelKey();
      form = formFor(got);
      confirmRemove = false;
      status = "Removed.";
      statusOk = true;
    } catch (e) {
      error = message(e, "could not remove the key; try again");
    } finally {
      busy = false;
    }
  }
</script>

<h3>AI opponents</h3>
<p class="help">
  Run the <strong>assisted</strong> and <strong>strong</strong> bots you add on your own model API key.
  Your seats then call the provider you choose, on your account and under its terms. The key is stored
  encrypted, is used only for bots you add, and CMD &amp; CTRL never shows it again or shares it with
  the other players. When you add a bot, tick “Run on my model key”.
</p>

{#if saved}
  <div class="saved-key" role="status">
    <div class="summary">{describeKey(saved)}</div>
    <div class="meta">
      Added {when(saved.created_at)} · last used {when(saved.last_used_at)} · {saved.games_today} of
      {saved.games_per_day} games today
    </div>
    {#if saved.last_error}
      <p class="refused">
        {saved.last_error}
        {#if saved.last_error_at}<span class="meta">({when(saved.last_error_at)})</span>{/if}
      </p>
    {/if}
  </div>
{/if}

<form
  class="key-form"
  onsubmit={(e) => {
    e.preventDefault();
    void save();
  }}
>
  <label class="row">
    <span>Provider</span>
    <select bind:value={form.provider} disabled={busy}>
      {#each providers as p (p.id)}
        <option value={p.id}>{p.label}</option>
      {/each}
    </select>
  </label>

  {#if form.provider === CUSTOM_PROVIDER}
    <label class="row">
      <span>API address</span>
      <input
        type="url"
        inputmode="url"
        autocomplete="off"
        spellcheck="false"
        placeholder="https://api.example.com/v1"
        bind:value={form.base_url}
        disabled={busy}
      />
    </label>
    <p class="help small">
      An OpenAI-compatible endpoint on the public internet, over https. Addresses on a private or
      local network are refused.
    </p>
  {/if}

  <label class="row">
    <span>Model</span>
    <input
      type="text"
      autocomplete="off"
      spellcheck="false"
      placeholder={provider?.model_hint ? `e.g. ${provider.model_hint}` : "model id"}
      bind:value={form.model}
      disabled={busy}
    />
  </label>

  <label class="row">
    <span>API key</span>
    <input
      type="password"
      autocomplete="off"
      spellcheck="false"
      placeholder={needsKey
        ? "paste your key"
        : `saved (ending ${saved?.last4 ?? ""}); paste to replace`}
      bind:value={form.key}
      disabled={busy}
    />
  </label>

  <div class="limits">
    <label class="row">
      <span>Games a day</span>
      <input
        type="number"
        min={limits.min_games_per_day}
        max={limits.max_games_per_day}
        bind:value={form.games_per_day}
        disabled={busy}
      />
    </label>
    <label class="row">
      <span>Calls per game</span>
      <input
        type="number"
        min={limits.min_calls_per_game}
        max={limits.max_calls_per_game}
        bind:value={form.calls_per_game}
        disabled={busy}
      />
    </label>
  </div>
  <p class="help small">
    Your limits for each bot on your key: how many games a day it may play, and how many model calls
    it may make in one game. Past either, or if the provider refuses the key, the bot keeps playing
    on the built-in heuristic and says so in chat.
  </p>

  <div class="actions">
    <button
      class="primary"
      type="submit"
      disabled={busy || form.model.trim() === "" || (needsKey && form.key.trim() === "")}
    >
      {saved ? "Save changes" : "Add key"}
    </button>
    {#if saved}
      <button type="button" disabled={busy} onclick={test}>Test key</button>
      {#if confirmRemove}
        <button type="button" class="danger" disabled={busy} onclick={remove}
          >Remove for good</button
        >
        <button type="button" class="ghost" disabled={busy} onclick={() => (confirmRemove = false)}>
          Keep it
        </button>
      {:else}
        <button type="button" class="ghost" disabled={busy} onclick={() => (confirmRemove = true)}>
          Remove key
        </button>
      {/if}
    {/if}
  </div>
</form>

{#if status}
  <p class="status" class:ok={statusOk} class:bad={!statusOk} role="status">{status}</p>
{/if}
{#if error}
  <p class="error" role="alert">{error}</p>
{/if}

<style>
  .help {
    color: var(--fg-dim);
    font-size: 11.5px;
    line-height: 1.45;
    margin: 4px 0 8px;
  }
  .help.small {
    margin-top: -2px;
  }
  .saved-key {
    border: 1px solid var(--border, var(--line));
    border-radius: 6px;
    padding: 0.5rem 0.7rem;
    margin: 0.4rem 0 0.8rem;
  }
  .summary {
    font-weight: 600;
    overflow-wrap: anywhere;
  }
  .meta {
    color: var(--fg-dim);
    font-size: 11.5px;
  }
  .key-form {
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
    max-width: 32rem;
  }
  .row {
    display: grid;
    grid-template-columns: 8rem minmax(0, 1fr);
    align-items: center;
    gap: 0.6rem;
  }
  .row input,
  .row select {
    min-width: 0;
    width: 100%;
  }
  .limits {
    display: flex;
    flex-wrap: wrap;
    gap: 0.45rem 1rem;
  }
  .limits .row {
    flex: 1 1 12rem;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-top: 0.3rem;
  }
  .refused {
    padding: 0.4rem 0.6rem;
    margin: 0.4rem 0 0;
    border-left: 3px solid var(--danger-line, var(--danger));
    background: var(--danger-soft);
    font-size: 12px;
  }
  .status.ok {
    color: var(--ok, var(--fg));
  }
  .status.bad,
  .error {
    color: var(--danger);
  }
  /* Phone width: the label goes above its field so the field keeps its width. */
  @media (max-width: 520px) {
    .row {
      grid-template-columns: minmax(0, 1fr);
      gap: 0.2rem;
    }
  }
</style>
