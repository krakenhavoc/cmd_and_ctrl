<script lang="ts">
  // CoachCard — the tutorial's one card, in the six states the CMD CTRL
  // Tutorial canvas draws (ADR 0076 §2.3, #1079):
  //
  //   opening    step 1: Skip tutorial / Start. No scrim yet.
  //   action     the default. No primary button: the game itself is the
  //              button, so the card only reports what it waits for.
  //   hint       the same after ~20s idle, with the hint added.
  //   watch      the bot's turn. No spotlight and no ask.
  //   recovered  the predicate saw the common wrong action; the step
  //              stays live and the card says what happened.
  //   done       step 11: Replay / Finish.
  //
  // Presentational only: TutorialCoach.svelte owns the step machine and
  // the placement. Every state has a way out (Skip step, Skip tutorial
  // or Finish), and none of them asks twice.
  //
  // On a phone (≤599px) the card is a strip above the dock bar and can
  // fold down to its one-line header, like the dock's sheets. On a
  // desktop it never folds; the minimise control is not drawn.

  import Icon from "../Icon.svelte";
  import type { CoachState } from "../../tutorial";

  interface Props {
    state: CoachState;
    /** This step's number and the tutorial's length, for "5 / 11". */
    n: number;
    total: number;
    title: string;
    body: string;
    /** Shown in the hint and recovered states. */
    hint?: string;
    /** The status line; defaults by state. */
    status?: string;
    /** Phone strip folded to one line. */
    minimised?: boolean;
    onStart?: () => void;
    onSkipTutorial?: () => void;
    onSkipStep?: () => void;
    onReplay?: () => void;
    onFinish?: () => void;
    onToggleMinimise?: () => void;
  }

  const {
    state,
    n,
    total,
    title,
    body,
    hint = "",
    status,
    minimised = false,
    onStart,
    onSkipTutorial,
    onSkipStep,
    onReplay,
    onFinish,
    onToggleMinimise,
  }: Props = $props();

  const showHint = $derived((state === "hint" || state === "recovered") && hint !== "");
  const statusLine = $derived(
    status ?? (state === "watch" ? "Bot is thinking" : "Waiting for you"),
  );
  const segments = $derived(Array.from({ length: total }, (_, i) => i + 1));
</script>

<aside class="coach" class:minimised data-state={state} aria-label="tutorial coach">
  <div class="coach-head">
    <span class="coach-count">{n} / {total}</span>
    <span class="coach-eyebrow">Tutorial</span>
    {#if minimised}
      <span class="coach-min-title">{title}</span>
    {/if}
    <button
      type="button"
      class="coach-fold"
      aria-label={minimised ? `restore: ${title}` : "minimise"}
      aria-expanded={!minimised}
      onclick={() => onToggleMinimise?.()}
    >
      <Icon name={minimised ? "chevron-up" : "chevron-down"} size={14} />
    </button>
  </div>
  {#if !minimised}
    <div class="coach-copy" aria-live="polite">
      <h2 class="coach-title">{title}</h2>
      <p class="coach-body">{body}</p>
      {#if showHint}
        <p class="coach-hint"><span class="coach-hint-tag">Hint</span>{hint}</p>
      {/if}
    </div>
    <div class="coach-progress" aria-hidden="true">
      {#each segments as s (s)}
        <span class="seg" class:past={s < n} class:now={s === n}></span>
      {/each}
    </div>
    <div class="coach-bar">
      {#if state === "opening"}
        <button type="button" class="coach-quiet" onclick={() => onSkipTutorial?.()}
          >Skip tutorial</button
        >
        <span class="grow"></span>
        <button type="button" class="coach-primary" onclick={() => onStart?.()}>Start</button>
      {:else if state === "done"}
        <button type="button" class="coach-quiet" onclick={() => onReplay?.()}>Replay</button>
        <span class="grow"></span>
        <button type="button" class="coach-primary" onclick={() => onFinish?.()}>Finish</button>
      {:else}
        <button type="button" class="coach-quiet" onclick={() => onSkipStep?.()}>Skip step</button>
        <span class="grow"></span>
        <span class="coach-status" class:watch={state === "watch"}>
          <span class="pip" aria-hidden="true"></span>{statusLine}
        </span>
      {/if}
    </div>
  {/if}
</aside>

<style>
  .coach {
    box-sizing: border-box;
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 9px;
    padding: 14px 16px 12px;
    border-radius: 14px;
    background: var(--surface);
    border: 1px solid var(--border-strong);
    box-shadow: var(--shadow-lg);
    color: var(--fg);
  }
  .coach-head {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
  }
  .coach-count {
    flex: none;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.1em;
    color: var(--gold);
    background: var(--gold-soft);
    border: 1px solid rgba(217, 180, 92, 0.4);
    border-radius: 999px;
    padding: 2px 8px;
  }
  .coach-eyebrow {
    flex: none;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .coach-min-title {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 13px;
    font-weight: 600;
  }
  /* The fold control is the phone strip's only; a desktop card never
     folds (see the media query at the end). */
  .coach-fold {
    display: none;
    margin-left: auto;
    flex: none;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 28px;
    padding: 0;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--fg-muted);
    cursor: pointer;
  }
  .coach-copy {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .coach-title {
    margin: 0;
    font-family: var(--font-display);
    font-size: 17px;
    line-height: 1.2;
    font-weight: 700;
    letter-spacing: -0.01em;
    text-transform: none;
    color: var(--fg);
  }
  .coach-body {
    margin: 0;
    font-size: 13px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .coach-hint {
    margin: 2px 0 0;
    padding: 8px 10px;
    border-radius: 8px;
    background: var(--gold-soft);
    border: 1px solid rgba(217, 180, 92, 0.3);
    font-size: 12.5px;
    line-height: 1.45;
    color: var(--fg);
  }
  .coach-hint-tag {
    margin-right: 8px;
    font-family: var(--font-mono);
    font-size: 9.5px;
    font-weight: 600;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--gold);
  }
  .coach-progress {
    display: flex;
    gap: 4px;
  }
  .seg {
    flex: 1 1 0;
    height: 3px;
    border-radius: 2px;
    background: rgba(255, 255, 255, 0.1);
  }
  .seg.past {
    background: var(--gold);
  }
  .seg.now {
    background: rgba(217, 180, 92, 0.45);
  }
  .coach-bar {
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: 32px;
  }
  .grow {
    flex: 1 1 auto;
  }
  .coach-quiet {
    background: none;
    border: 0;
    padding: 6px 2px;
    font: 500 12px var(--font-ui);
    color: var(--fg-muted);
    cursor: pointer;
  }
  .coach-quiet:hover {
    color: var(--fg);
  }
  .coach-primary {
    min-width: 92px;
    height: 32px;
    padding: 0 16px;
    border-radius: var(--radius);
    border: 0;
    background: var(--gold);
    color: var(--accent-fg);
    font: 600 13px var(--font-ui);
    cursor: pointer;
  }
  .coach-primary:hover {
    background: var(--gold-strong);
  }
  .coach-status {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    font-size: 12px;
    font-weight: 500;
    color: var(--gold);
  }
  .coach-status.watch {
    color: var(--magenta);
  }
  .pip {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: currentColor;
  }
  @media (max-width: 599px) {
    .coach {
      gap: 7px;
      padding: 10px 12px;
    }
    .coach.minimised {
      padding: 6px 12px;
    }
    .coach-fold {
      display: inline-flex;
    }
    .coach-title {
      font-size: 15px;
    }
    .coach-body {
      font-size: 12.5px;
    }
    .coach-quiet,
    .coach-primary {
      min-height: 44px;
    }
  }
</style>
