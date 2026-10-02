<script lang="ts">
  // VotingPanel is the council's-dilemma modal that opens for every
  // player when view.vote is set. Each player sees the topic, the
  // options, the live tally, and a button to cast / re-cast their
  // ballot. Initiator (or any seated player) can close the vote.
  //
  // ADR 0111 PR 7: the "start a vote" launcher that sat in the board's
  // top-left corner is "Call a vote…" in the action dock's ⋯ menu now
  // (GameMenu.svelte). A spectator never had a vote to call: the
  // launcher needed a seat.

  import type { ActionPayload, ActionType, GameView } from "../../protocol";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    view: GameView;
    viewerID: string | null;
    sendAction: ActionSender;
    // ADR 0111 PR 5: the action dock is on screen. An open vote is then
    // the dock's (lib/choiceDock.ts voteRequest), so this panel draws
    // nothing. Without a dock (a spectator) the open vote is drawn here
    // as before.
    docked?: boolean;
  }

  const { view, viewerID, sendAction, docked = false }: Props = $props();

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

  function nameOf(playerID: string): string {
    return view.seats.find((s) => s.id === playerID)?.name ?? "?";
  }
</script>

{#if vote && !docked}
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
    background: linear-gradient(180deg, rgba(19, 26, 44, 0.92) 0%, rgba(8, 12, 24, 0.92) 100%);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    border: 1px solid rgba(176, 138, 255, 0.6);
    border-radius: var(--radius-lg);
    box-shadow:
      0 14px 36px rgba(0, 0, 0, 0.65),
      0 0 20px rgba(176, 138, 255, 0.2),
      inset 0 1px 0 rgba(255, 255, 255, 0.06);
    padding: 14px 18px;
    min-width: 280px;
    max-width: 420px;
    color: var(--fg);
  }
  .head {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 10px;
  }
  .badge {
    background: linear-gradient(180deg, #c09dff 0%, #9c78f0 100%);
    color: #1a0d33;
    font-size: 9px;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    padding: 3px 8px;
    border-radius: 999px;
    box-shadow:
      0 2px 6px rgba(0, 0, 0, 0.4),
      inset 0 1px 0 rgba(255, 255, 255, 0.3);
  }
  .topic {
    font-size: 14px;
    flex: 1;
    font-weight: 600;
    letter-spacing: -0.01em;
  }
  .initiator {
    font-size: 10px;
    color: var(--fg-dim);
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
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: var(--radius);
    padding: 8px 12px;
    color: inherit;
    font: inherit;
    cursor: pointer;
    text-align: left;
    box-shadow: none;
    transition:
      background 140ms var(--ease),
      border-color 140ms var(--ease);
  }
  .option:hover {
    background: rgba(176, 138, 255, 0.1);
    border-color: rgba(176, 138, 255, 0.6);
  }
  .option.mine {
    background: rgba(176, 138, 255, 0.22);
    border-color: #b08aff;
    color: #e1d3ff;
    box-shadow: 0 0 14px rgba(176, 138, 255, 0.25);
  }
  .option-text {
    flex: 1;
  }
  .option-tally {
    font-weight: 800;
    font-variant-numeric: tabular-nums;
    color: #b08aff;
    min-width: 18px;
    text-align: right;
  }
  .foot {
    display: flex;
    justify-content: flex-end;
    margin-top: 14px;
  }
  .end {
    background: transparent;
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: var(--fg-muted);
    padding: 5px 14px;
    border-radius: 999px;
    cursor: pointer;
    font: inherit;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    font-weight: 700;
    box-shadow: none;
    transition:
      background 120ms var(--ease),
      border-color 120ms var(--ease),
      color 120ms var(--ease);
  }
  .end:hover {
    background: rgba(255, 122, 122, 0.15);
    color: var(--danger);
    border-color: rgba(255, 122, 122, 0.6);
  }
</style>
