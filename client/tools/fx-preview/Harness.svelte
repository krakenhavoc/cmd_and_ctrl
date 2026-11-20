<script lang="ts">
  // Harness — the fx preview page (#4868): a button per scenario, the
  // settings a cue answers to, and the stage it plays on.
  //
  // Playing a scenario remounts the effect layers on its `before` table
  // (their first frame primes, so nothing is cued), waits two animation
  // frames so they measure it, then lands its `after` table: the board
  // first, then the view with the scenario's log, as Game.svelte delivers
  // a frame. The hash says what to play on load: `#counters`, or
  // `#counters@300` to pause every animation 300 ms after the frame lands.
  //
  // This page is a dev tool, not product UI: its own text is not in
  // lib/labels.ts, and it never ships (vite.config.ts here, README.md).

  import { onMount, tick } from "svelte";
  import { settings, updateSettings } from "../../src/lib/settings";
  import type { EffectsMode } from "../../src/lib/tableEffects";
  import type { GameView } from "../../src/lib/protocol";
  import Stage from "./Stage.svelte";
  import { pauseClock, resumeClock } from "./clock";
  import { SCENARIOS, type Scenario } from "./scenarios";
  import { stamp, toView, type FxCard } from "./table";

  const GAP_MS = 2800;

  let cards: readonly FxCard[] = $state.raw(SCENARIOS[0].before);
  let view: GameView = $state.raw(toView(SCENARIOS[0].before, []));
  let run = $state(0);
  let current: string | null = $state(null);
  let paused = $state(false);
  let playingAll = $state(false);
  // Bumped by every play, so an older one that is still waiting stops.
  let token = 0;
  let freezeTimer: ReturnType<typeof setTimeout> | undefined;
  let allTimer: ReturnType<typeof setTimeout> | undefined;

  const groups = [...new Set(SCENARIOS.map((s) => s.group))];
  const speeds = [0.5, 1, 1.5, 2] as const;
  const modes: { value: EffectsMode; text: string }[] = [
    { value: "full", text: "Full" },
    { value: "short", text: "Short" },
    { value: "off", text: "Off" },
  ];

  const frames = (n: number) =>
    new Promise<void>((done) => {
      const step = (left: number) =>
        left === 0 ? done() : requestAnimationFrame(() => step(left - 1));
      step(n);
    });

  // Pause holds the animations and the timers that start and clear a
  // cue's beats (clock.ts), so a frozen cue stays on the table.
  function setPaused(p: boolean): void {
    paused = p;
    if (p) pauseClock();
    else resumeClock();
  }

  async function play(s: Scenario, freezeAt?: number): Promise<void> {
    const mine = ++token;
    clearTimeout(freezeTimer);
    if (paused) setPaused(false);
    current = s.id;
    history.replaceState(null, "", `#${s.id}`);
    // A fresh mount on the table as it stood: its first frame primes.
    run += 1;
    cards = s.before;
    view = toView(s.before, []);
    await tick();
    // Let the layers measure every tile and pile where it stands.
    await frames(2);
    if (mine !== token) return;
    // The frame lands: the board shows its end state, then the log arrives.
    cards = s.after;
    await tick();
    view = toView(s.after, stamp(s.log), s.out);
    if (freezeAt !== undefined) freezeTimer = setTimeout(() => setPaused(true), freezeAt);
  }

  function playAll(from = 0): void {
    clearTimeout(allTimer);
    if (from === 0 && playingAll) {
      playingAll = false;
      return;
    }
    if (from >= SCENARIOS.length) {
      playingAll = false;
      return;
    }
    playingAll = true;
    void play(SCENARIOS[from]);
    allTimer = setTimeout(() => playAll(from + 1), GAP_MS);
  }

  function choose(s: Scenario): void {
    clearTimeout(allTimer);
    playingAll = false;
    void play(s);
  }

  // The reduced-motion setting also cuts app.css's own transitions, as
  // App.svelte does with this attribute.
  $effect(() => {
    document.documentElement.dataset.reduceMotion = $settings.accessibility.reduceMotion
      ? "1"
      : "0";
  });

  // A press anywhere skips the cues playing, as at the table: the layers
  // listen on window, in the capture phase. Pause must not skip what it
  // pauses, so this listener, added before any of theirs, keeps the press
  // on [data-keeps-cues] from them. Its click still follows.
  function keepCues(e: PointerEvent): void {
    if (e.target instanceof Element && e.target.closest("[data-keeps-cues]")) {
      e.stopImmediatePropagation();
    }
  }

  onMount(() => {
    window.addEventListener("pointerdown", keepCues, true);
    const [id, at] = location.hash.slice(1).split("@");
    const s = SCENARIOS.find((x) => x.id === id);
    if (s) void play(s, at ? Number(at) : undefined);
    return () => {
      window.removeEventListener("pointerdown", keepCues, true);
      clearTimeout(freezeTimer);
      clearTimeout(allTimer);
    };
  });
</script>

<div class="harness">
  <aside>
    <h1>Table effects preview</h1>
    <p class="note">
      Each button lands one moment on a fake table, through the real cue layers. A click or Escape
      skips what is playing, as at the table.
    </p>

    <div class="row">
      <button type="button" class="btn" onclick={() => playAll()}>
        {playingAll ? "Stop" : "Play all"}
      </button>
      <button type="button" class="btn" data-keeps-cues onclick={() => setPaused(!paused)}>
        {paused ? "Resume" : "Pause"}
      </button>
    </div>

    <fieldset>
      <legend>Effects</legend>
      {#each modes as m (m.value)}
        <label>
          <input
            type="radio"
            name="effects"
            value={m.value}
            checked={$settings.animations.effects === m.value}
            onchange={() => updateSettings("animations", "effects", m.value)}
          />
          {m.text}
        </label>
      {/each}
    </fieldset>
    <fieldset>
      <legend>Speed</legend>
      {#each speeds as sp (sp)}
        <label>
          <input
            type="radio"
            name="speed"
            value={sp}
            checked={$settings.animations.speed === sp}
            onchange={() => updateSettings("animations", "speed", sp)}
          />
          {sp}×
        </label>
      {/each}
    </fieldset>
    <fieldset>
      <legend>Motion and sound</legend>
      <label>
        <input
          type="checkbox"
          checked={$settings.animations.enabled}
          onchange={(e) => updateSettings("animations", "enabled", e.currentTarget.checked)}
        />
        Animations on
      </label>
      <label>
        <input
          type="checkbox"
          checked={$settings.accessibility.reduceMotion}
          onchange={(e) => updateSettings("accessibility", "reduceMotion", e.currentTarget.checked)}
        />
        Reduce motion
      </label>
      <label>
        <input
          type="checkbox"
          checked={!$settings.audio.muted}
          onchange={(e) => updateSettings("audio", "muted", !e.currentTarget.checked)}
        />
        Sound
      </label>
    </fieldset>

    {#each groups as g (g)}
      <h2>{g}</h2>
      <ul>
        {#each SCENARIOS.filter((s) => s.group === g) as s (s.id)}
          <li>
            <button
              type="button"
              class="scenario"
              class:current={current === s.id}
              aria-pressed={current === s.id}
              onclick={() => choose(s)}
            >
              <span>{s.title}</span>
              <span class="issue">#{s.issue}</span>
            </button>
          </li>
        {/each}
      </ul>
    {/each}
  </aside>

  <main>
    <Stage {cards} {view} {run} />
  </main>
</div>

<style>
  .harness {
    display: grid;
    grid-template-columns: 280px 1fr;
    gap: 16px;
    height: 100vh;
    padding: 16px;
    box-sizing: border-box;
    background: var(--bg);
    color: var(--fg);
    font-family: var(--font-ui);
  }
  aside {
    overflow-y: auto;
    display: grid;
    align-content: start;
    gap: 12px;
  }
  h1 {
    margin: 0;
    font: 600 20px var(--font-display);
  }
  h2 {
    margin: 8px 0 0;
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--fg-muted);
  }
  .note {
    margin: 0;
    font-size: 13px;
    color: var(--fg-muted);
  }
  .row {
    display: flex;
    gap: 8px;
  }
  fieldset {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 12px;
    margin: 0;
    padding: 8px 10px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    font-size: 13px;
  }
  legend {
    padding: 0 4px;
    color: var(--fg-muted);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 4px;
  }
  .scenario {
    display: flex;
    justify-content: space-between;
    width: 100%;
    padding: 6px 10px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--surface);
    color: var(--fg);
    font: inherit;
    font-size: 13px;
    text-align: left;
    cursor: pointer;
    transition: background-color 120ms var(--ease);
  }
  .scenario:hover {
    background: var(--surface-hover);
  }
  .scenario.current {
    border-color: var(--accent-line);
    background: var(--accent-soft);
  }
  .issue {
    color: var(--fg-dim);
    font-family: var(--font-mono);
  }
  main {
    min-width: 0;
    min-height: 0;
  }
</style>
