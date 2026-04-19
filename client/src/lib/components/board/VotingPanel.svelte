<script lang="ts">
  // VotingPanel is the council's-dilemma modal that opens for every
  // player when view.vote is set. Each player sees the topic, the
  // options, the live tally, and a button to cast / re-cast their
  // ballot. Initiator (or any seated player) can close the vote.
  //
  // When no vote is open, the panel renders a small "start a vote"
  // launcher in the corner — a freeform topic + comma-separated
  // options field. Kept inline with the modal to keep the politics
  // surface in one place.

  import type { ActionPayload, GameView } from "../../protocol";

  type ActionSender = (type: string, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    view: GameView;
    viewerID: string | null;
    sendAction: ActionSender;
  }

  const { view, viewerID, sendAction }: Props = $props();

  const vote = $derived(view.vote ?? null);

  // Tally derived from ballots. Map index → count for live display.
  const tally = $derived.by(() => {
    if (!vote) return [];
    const counts = new Array(vote.options.length).fill(0);
    for (const opt of Object.values(vote.ballots ?? {})) {
      if (opt >= 0 && opt < counts.length) counts[opt]++;
    }
    return counts;
  });

  const myBallot = $derived(viewerID && vote ? (vote.ballots?.[viewerID] ?? null) : null);

  function castBallot(optionIndex: number): void {
    if (!viewerID) return;
    sendAction("cast_vote", { option: optionIndex }, viewerID);
  }

  function endVote(): void {
    sendAction("end_vote");
  }

  // Launcher state — only relevant when no vote is open.
  let launcherOpen = $state(false);
  let topic = $state("");
  let optionsText = $state("yes, no");

  function startVote(): void {
    if (!viewerID) return;
    const options = optionsText
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    if (options.length < 2 || !topic.trim()) return;
    sendAction("start_vote", { topic: topic.trim(), options }, viewerID);
    topic = "";
    optionsText = "yes, no";
    launcherOpen = false;
  }

  function nameOf(playerID: string): string {
    return view.seats.find((s) => s.id === playerID)?.name ?? "?";
  }
</script>

{#if vote}
  <div class="vote-modal" role="dialog" aria-modal="true" aria-label="open vote">
    <header class="head">
      <span class="badge" aria-hidden="true">vote</span>
      <strong class="topic">{vote.topic || "(no topic)"}</strong>
      <span class="initiator">called by {nameOf(vote.initiator)}</span>
    </header>
    <ol class="options">
      {#each vote.options as option, i (i)}
        {@const count = tally[i] ?? 0}
        {@const isMine = myBallot === i}
        <li>
          <button
            type="button"
            class="option"
            class:mine={isMine}
            onclick={() => castBallot(i)}
            aria-pressed={isMine}
          >
            <span class="option-text">{option}</span>
            <span class="option-tally">{count}</span>
          </button>
        </li>
      {/each}
    </ol>
    <footer class="foot">
      <button type="button" class="end" onclick={endVote}>end vote</button>
    </footer>
  </div>
{:else}
  <div class="vote-launcher">
    {#if launcherOpen}
      <form
        class="launcher-form"
        onsubmit={(e) => {
          e.preventDefault();
          startVote();
        }}
      >
        <input
          type="text"
          bind:value={topic}
          placeholder="topic (e.g. 'monarchy?')"
          aria-label="vote topic"
        />
        <input
          type="text"
          bind:value={optionsText}
          placeholder="options, comma-separated"
          aria-label="vote options"
        />
        <button type="submit">start</button>
        <button type="button" onclick={() => (launcherOpen = false)}>cancel</button>
      </form>
    {:else}
      <button
        type="button"
        class="launcher-btn"
        onclick={() => (launcherOpen = true)}
        title="call a vote"
        aria-label="call a vote"
      >
        vote ▾
      </button>
    {/if}
  </div>
{/if}

<style>
  /* Open-vote modal — top-centre, large enough to read across the
     table, but not full-width so the board stays partly visible
     behind it. Mirrors the look of other floating overlays. */
  .vote-modal {
    position: absolute;
    top: 12px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 50;
    background: #0d1424;
    border: 1px solid #b08aff;
    border-radius: 10px;
    box-shadow: 0 12px 30px rgba(0, 0, 0, 0.7);
    padding: 12px 16px;
    min-width: 280px;
    max-width: 420px;
    color: #e0e6f5;
  }
  .head {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 10px;
  }
  .badge {
    background: #b08aff;
    color: #0d1424;
    font-size: 9px;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    padding: 2px 6px;
    border-radius: 3px;
  }
  .topic {
    font-size: 14px;
    flex: 1;
  }
  .initiator {
    font-size: 10px;
    color: #6c7a99;
    text-transform: lowercase;
  }
  .options {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .option {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
    background: transparent;
    border: 1px solid #2e3a55;
    border-radius: 5px;
    padding: 6px 10px;
    color: inherit;
    font: inherit;
    cursor: pointer;
    text-align: left;
  }
  .option:hover {
    background: #1a2335;
    border-color: #b08aff;
  }
  .option.mine {
    background: rgba(176, 138, 255, 0.18);
    border-color: #b08aff;
    color: #d8c8ff;
  }
  .option-text {
    flex: 1;
  }
  .option-tally {
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    color: #b08aff;
    min-width: 18px;
    text-align: right;
  }
  .foot {
    display: flex;
    justify-content: flex-end;
    margin-top: 12px;
  }
  .end {
    background: transparent;
    border: 1px solid #4a5270;
    color: #c8c8c8;
    padding: 4px 12px;
    border-radius: 4px;
    cursor: pointer;
    font: inherit;
    font-size: 11px;
  }
  .end:hover {
    background: #1a2335;
    color: #ff7a7a;
    border-color: #ff7a7a;
  }

  /* Launcher when no vote is open — small button, tucked next to the
     other overlays. Top-left under the StackOverlay corner. */
  .vote-launcher {
    position: absolute;
    top: 12px;
    left: 12px;
    z-index: 30;
  }
  .launcher-btn {
    background: #0d1424;
    border: 1px solid #4a5270;
    color: #b08aff;
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    padding: 3px 8px;
    border-radius: 4px;
    cursor: pointer;
    font-family: inherit;
  }
  .launcher-btn:hover {
    background: #1a2335;
    border-color: #b08aff;
  }
  .launcher-form {
    display: flex;
    gap: 6px;
    align-items: center;
    background: #0d1424;
    border: 1px solid #b08aff;
    border-radius: 6px;
    padding: 6px 8px;
  }
  .launcher-form input {
    background: #1a2335;
    border: 1px solid #2e3a55;
    color: #e0e6f5;
    border-radius: 4px;
    padding: 3px 6px;
    font: inherit;
    font-size: 11px;
    width: 140px;
  }
  .launcher-form input:focus {
    outline: 1px solid #b08aff;
  }
  .launcher-form button {
    background: transparent;
    border: 1px solid #4a5270;
    color: #c8c8c8;
    padding: 3px 8px;
    border-radius: 3px;
    cursor: pointer;
    font: inherit;
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }
  .launcher-form button[type="submit"] {
    color: #b08aff;
    border-color: #b08aff;
  }
</style>
