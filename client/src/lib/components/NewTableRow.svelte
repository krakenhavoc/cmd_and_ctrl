<script lang="ts">
  // The lobby's create-a-table row (#2630): the name box, a dice that
  // fills it with a suggestion, and the create button.
  //
  // A blank name is allowed. The server names the table, so the button
  // is only ever busy-disabled. The dice asks GET /games/name-suggestion
  // for a name the person can keep, reroll or edit.
  import { fetchNameSuggestion } from "../api";
  import { L } from "../labels";

  interface Props {
    name: string;
    busy?: boolean;
    /** Injected so a test needs no network. */
    suggest?: () => Promise<string>;
  }

  let { name = $bindable(), busy = false, suggest = fetchNameSuggestion }: Props = $props();

  let rolling = $state(false);

  async function roll(): Promise<void> {
    if (rolling) return;
    rolling = true;
    try {
      name = await suggest();
    } catch {
      // The box keeps what it had: a failed suggestion is not worth an
      // error, and a blank name still creates a table.
    } finally {
      rolling = false;
    }
  }
</script>

<div class="name-row">
  <input
    type="text"
    placeholder="game name (blank = surprise me)"
    aria-label="game name"
    maxlength="80"
    bind:value={name}
  />
  <button
    type="button"
    class="dice"
    aria-label={L.suggestTableName}
    title="Suggest a name"
    disabled={busy || rolling}
    onclick={roll}
  >
    🎲
  </button>
  <button type="submit" class="primary" disabled={busy}>create</button>
</div>

<style>
  .name-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .name-row input[type="text"] {
    flex: 1;
    min-width: 0;
    margin: 0;
    height: 36px;
    padding: 0 12px;
    box-sizing: border-box;
    font-size: 13px;
  }
  .name-row button {
    height: 36px;
    flex: 0 0 auto;
  }
  .dice {
    padding: 0 10px;
  }
</style>
