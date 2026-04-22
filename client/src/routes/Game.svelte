<script lang="ts">
  import { onMount } from "svelte";
  import { GameClient } from "../lib/ws";
  import { navigate } from "../lib/router";
  import { session } from "../lib/session";
  import { seatColor } from "../lib/colors";
  import DeckUploadForm from "../lib/components/DeckUploadForm.svelte";
  import Board from "../lib/components/board/Board.svelte";
  import DiscardPromptModal from "../lib/components/board/DiscardPromptModal.svelte";
  import ChoicePromptModal from "../lib/components/board/ChoicePromptModal.svelte";
  import AutoTapPreviewModal from "../lib/components/board/AutoTapPreviewModal.svelte";
  import TargetingBanner from "../lib/components/board/TargetingBanner.svelte";
  import { cancel as cancelTargeting } from "../lib/targeting";
  import type { PlayerView } from "../lib/protocol";
  import { armAudioOnFirstGesture, isMuted, play, toggleMuted } from "../lib/sounds";
  import { openSettings, settings } from "../lib/settings";
  import { grantsPriority } from "../lib/turn";

  interface Props {
    gameID: string;
  }
  const { gameID }: Props = $props();

  // Build the WS URL from the stored session + route. Query string
  // carries the session token (browsers can't send Authorization on
  // WS upgrade) and identifies which seat to render. If the session
  // has a player_id bound to this game, we pin it; admin sessions
  // may omit ?player= and fall through to the spectator view.
  const baseURL = (location.protocol === "https:" ? "wss://" : "ws://") + location.host + "/ws";
  const sess = $derived($session);
  const wsURL = $derived.by(() => {
    const params = new URLSearchParams();
    params.set("game", gameID);
    if (sess?.token) params.set("token", sess.token);
    if (sess?.playerID && sess.gameID === gameID) params.set("player", sess.playerID);
    return `${baseURL}?${params.toString()}`;
  });

  // One GameClient per component instance. Start with an empty URL
  // — the $effect below installs the real URL on first run and
  // reconnects whenever wsURL changes. This preserves the store
  // subscribers across reactive reruns (vs. replacing the client
  // instance, which would strand subscriptions on the old object).
  const client = new GameClient("");
  // chat store stays populated server-side but isn't rendered: the
  // chat UI was removed in S08.5 wave 1 in favour of out-of-band
  // (Discord) coordination. Re-add `chat` to this destructure when
  // a chat panel returns.
  const { status, snapshot, lastSeq, lastError } = client;

  $effect(() => {
    client.disconnect();
    client.setURL(wsURL);
    client.connect();
    return () => client.disconnect();
  });

  // Autoplay-policy unlock. The first pointerdown / keydown anywhere in
  // the window preloads + decodes every sound variant; sounds.ts no-ops
  // after the first call so reruns (new wsURL) are harmless.
  $effect(() => {
    armAudioOnFirstGesture();
  });

  // settings.gameplay.confirmExit also covers the browser-level
  // close-tab / hard-refresh case via beforeunload. Returning a
  // string from the handler triggers the browser's native confirm
  // dialog (text is browser-controlled — the string we return is
  // ignored in modern browsers but the non-empty return value still
  // arms the prompt). Only installed while the gate is on AND the
  // game is live, so an idle lobby tab doesn't ask "really leave?".
  onMount(() => {
    const handler = (e: BeforeUnloadEvent) => {
      if (
        $settings.gameplay.confirmExit &&
        view?.state === "active" &&
        !viewerEliminated &&
        !gameEnded
      ) {
        e.preventDefault();
        return "Leave this game?";
      }
      return undefined;
    };
    window.addEventListener("beforeunload", handler);
    return () => window.removeEventListener("beforeunload", handler);
  });

  // settings.gameplay.autoPassPriority + settings.gameplay.stepStops
  // (S13): when the viewer holds priority on an empty stack, auto-
  // pass unless the current step is one they opted to stop on. The
  // pre-S13 behaviour was "auto-pass through opponents' turns only";
  // S13 generalises that to "auto-pass through every step the user
  // hasn't pinned." Active-turn stops default-on for the main phases
  // and combat declarations, so the active player still gets stopped
  // for their plays even with autoPassPriority enabled.
  let lastAutoPassedSeq = $state(-1);
  $effect(() => {
    if (!$settings.gameplay.autoPassPriority) return;
    if (!viewerHasPriority) return;
    if ((view?.stack?.cards?.length ?? 0) > 0) return;
    if (mulligansOpen || gameEnded || viewerEliminated) return;
    const step = view?.turn?.step;
    if (!step) return;
    // Stop here if the viewer has opted to stop on this step. The
    // map omits no-priority steps (Untap / Cleanup); for those, the
    // viewer can never hold priority anyway.
    if ($settings.gameplay.stepStops[step] === true) return;
    // Dedupe by snapshot seq so we don't fire twice on the same
    // priority window if the effect re-runs for an unrelated reason
    // before the next snapshot lands.
    const seq = $lastSeq;
    if (seq === lastAutoPassedSeq) return;
    lastAutoPassedSeq = seq;
    client.sendAction("pass_priority");
  });

  // sendAction is a thin shim over GameClient.sendAction that the
  // Board passes to its child components for interactive mutations.
  // Bound at module scope so session changes (logout + re-login) pick
  // up the new viewer ID via the closure on `client`.
  //
  // S15: cast_spell actions get the viewer's gameplay.strictMana
  // preference auto-stamped onto the params unless the caller has
  // already set `strict` (e.g. the override toast injects
  // force_cast=true and we want to skip the auto-stamp on that
  // path). Other action types pass through unchanged.
  const sendAction = (type: string, params?: unknown, player?: string): void => {
    if (type === "cast_spell") {
      const strict = $settings.gameplay.strictMana;
      const incoming = (params ?? {}) as Record<string, unknown>;
      if (incoming.strict === undefined) {
        params = { ...incoming, strict };
      }
    }
    client.sendAction(type, player, params);
  };

  // S15: insufficient-mana override toast. Subscribes to the
  // GameClient's lastError store; when an `insufficient_mana` frame
  // lands, we capture the missing list + card_id so the override
  // banner can surface "Cast anyway" — clicking re-fires the cast
  // with `force_cast: true` (which the server treats as a permissive
  // proceed without touching the pool). Cleared when the next
  // snapshot or non-mana error arrives. lastError is already
  // destructured at the top of this script from the GameClient.
  let manaOverride = $state<{ cardID: string; missing: string[] } | null>(null);
  $effect(() => {
    const err = $lastError;
    if (!err) {
      manaOverride = null;
      return;
    }
    if (err.code !== "insufficient_mana" || !err.cardID) {
      manaOverride = null;
      return;
    }
    manaOverride = { cardID: err.cardID, missing: err.missing ?? [] };
  });
  function castAnyway(): void {
    if (!manaOverride) return;
    const cardID = manaOverride.cardID;
    manaOverride = null;
    sendAction(
      "cast_spell",
      { instance_id: cardID, strict: true, force_cast: true },
      viewerID ?? undefined,
    );
  }
  function dismissManaOverride(): void {
    manaOverride = null;
    client.lastError.set(null);
  }

  // S15 sub-PR 5 auto-tap-and-cast modal driver. autoTapCardID
  // doubles as the open / closed state — when null, the modal is
  // closed; when set, AutoTapPreviewModal mounts and fetches the
  // preview for that card. The "Auto-tap & cast" button on the
  // insufficient-mana toast is the canonical entry point; the
  // dismiss button (and ESC inside the modal) closes it.
  let autoTapCardID = $state<string | null>(null);
  function openAutoTap(): void {
    if (!manaOverride) return;
    autoTapCardID = manaOverride.cardID;
    manaOverride = null;
  }
  function confirmAutoTap(lockedSources: string[]): void {
    if (!autoTapCardID) return;
    const cardID = autoTapCardID;
    autoTapCardID = null;
    sendAction(
      "cast_spell",
      {
        instance_id: cardID,
        strict: true,
        auto_tap: true,
        locked_sources: lockedSources,
      },
      viewerID ?? undefined,
    );
  }
  function cancelAutoTap(): void {
    autoTapCardID = null;
  }

  function back(): void {
    // settings.gameplay.confirmExit: guard the manual back-to-lobby
    // button so a misclick doesn't drop a live game. The browser
    // close-tab case is handled separately via beforeunload below.
    // We only prompt while a game is in flight — during the lobby
    // phase (state === "lobby") there's nothing to lose.
    if (
      $settings.gameplay.confirmExit &&
      view?.state === "active" &&
      !viewerEliminated &&
      !gameEnded
    ) {
      if (!confirm("Leave this game? Your seat stays active — you can rejoin from the lobby.")) {
        return;
      }
    }
    navigate("#/lobby");
  }

  // ---- Turn / priority / quick actions ----

  const view = $derived($snapshot);
  const seats = $derived<PlayerView[]>(view?.seats ?? []);
  const turn = $derived(view?.turn);
  const activeSeat = $derived(turn?.active_seat ?? 0);
  const prioritySeat = $derived(turn?.priority_holder ?? 0);
  const activePlayer = $derived(seats[activeSeat]);
  const priorityPlayer = $derived(seats[prioritySeat]);
  const viewerID = $derived(sess?.playerID ?? null);
  const viewerSeat = $derived(seats.find((s) => s.id === viewerID) ?? null);
  const viewerHasPriority = $derived(viewerID !== null && priorityPlayer?.id === viewerID);
  const viewerIsActive = $derived(viewerID !== null && activePlayer?.id === viewerID);
  const viewerEliminated = $derived(viewerSeat?.eliminated === true);

  // Game-end state. The server transitions State to "ended" once
  // exactly one non-eliminated seat remains; the survivor is the
  // implicit winner.
  const gameEnded = $derived(view?.state === "ended");
  const survivors = $derived(seats.filter((s) => !s.eliminated));
  const winner = $derived(gameEnded && survivors.length === 1 ? survivors[0] : null);

  // Mulligan window: open between Start and the moment everyone has
  // KeptHand. The dialog blocks the viewer's normal toolbar until
  // they commit. The viewer can still see the table, chat, etc.
  const mulligansOpen = $derived(view?.mulligans_open === true);
  const viewerNeedsToDecide = $derived(
    mulligansOpen && !!viewerSeat && !viewerSeat.eliminated && !viewerSeat.hand_kept,
  );

  // Pre-game deck-import modal (S08.5 wave 1). A player who accepted
  // an invite lands directly in Game.svelte without ever visiting
  // Lobby, so they have no affordance to upload a deck unless we
  // surface one here. Fires when the game is still in lobby state,
  // the viewer is seated, and they haven't yet imported a real deck.
  // The check uses the deck_imported flag (server-stamped) rather
  // than a card-count test because Lobby.Join hands out a 1-card
  // placeholder commander at seat-creation time — so library / command
  // counts are non-zero even before import. `deckImportDismissed`
  // lets the modal hide after a successful upload (SetDeck currently
  // doesn't trigger a snapshot rebroadcast, so the local flag is
  // what dismisses for the uploader; the next snapshot is
  // authoritative for everyone else).
  let deckImportDismissed = $state(false);
  const viewerNeedsDeck = $derived.by(() => {
    if (!view || view.state !== "lobby") return false;
    if (!viewerSeat) return false;
    return !viewerSeat.deck_imported;
  });

  // Step-transition sound cues. Snapshot-driven, so we track the last
  // seen step and only fire on a real change; the initial snapshot (or
  // a reconnect rebuild) sets the baseline silently.
  let prevStep: string | null = null;
  $effect(() => {
    if (!turn) return;
    const step = turn.step;
    if (prevStep !== null && step !== prevStep) {
      if (step === "untap") {
        play("turn_change");
        play("untap_all");
      } else if (step === "combat_damage") {
        play("combat_resolve");
      }
    }
    prevStep = step;
  });

  // Win / loss cue on the state→ended transition. Spectators (no
  // viewerID) hear neither — the outcome isn't theirs.
  let prevEnded = false;
  $effect(() => {
    const ended = gameEnded;
    if (ended && !prevEnded && viewerID) {
      play(winner?.id === viewerID ? "win" : "loss");
    }
    prevEnded = ended;
  });

  // Mute toggle backing state. Seeded from localStorage via isMuted();
  // the click handler flips both the shared store and this local copy
  // so the button label re-renders immediately.
  let muted = $state(isMuted());
  function onToggleMute(): void {
    toggleMuted();
    muted = isMuted();
  }

  function passPriority(): void {
    client.sendAction("pass_priority");
  }

  function passTurn(): void {
    client.sendAction("pass_turn");
  }

  // advanceStep is the sandbox shortcut that bumps the step cursor
  // forward by one without requiring both players to pass priority.
  // Used by the "next step" toolbar button and the "done" button in
  // the combat panel — solo testing and casual play don't need to
  // simulate the priority hand-off rigorously.
  function advanceStep(): void {
    client.sendAction("advance_step");
  }

  // "Pass until end of turn": send pass_priority repeatedly, waiting
  // for each snapshot to settle, until the cursor reaches the cleanup
  // step or the active seat changes. Capped at 24 iterations as a
  // safety belt against an unexpected state machine loop.
  let passingToEnd = $state(false);
  async function passToEnd(): Promise<void> {
    if (passingToEnd || !turn) return;
    passingToEnd = true;
    const startSeat = activeSeat;
    const startSeq = $lastSeq;
    let lastSeenSeq = startSeq;
    try {
      for (let i = 0; i < 24; i++) {
        const v = $snapshot;
        if (!v) break;
        if (v.turn.active_seat !== startSeat) break;
        if (v.turn.step === "cleanup") break;
        client.sendAction("pass_priority");
        // Wait for the snapshot store to tick at least once. Polling
        // the lastSeq store is simpler than wiring a one-shot
        // subscription and matches the other reactive paths in the
        // route.
        const before = lastSeenSeq;
        const deadline = Date.now() + 1500;
        while (Date.now() < deadline) {
          await new Promise((r) => setTimeout(r, 25));
          if ($lastSeq > before) {
            lastSeenSeq = $lastSeq;
            break;
          }
        }
        if (lastSeenSeq === before) break;
      }
    } finally {
      passingToEnd = false;
    }
  }

  // "Pass to my next stop" (S13): same loop shape as passToEnd, but
  // the exit condition is "stop at the next step the viewer has
  // pinned in settings.gameplay.stepStops, or whenever the active
  // seat / stack changes." Used by the "→ next stop" button so the
  // viewer can skip past steps they don't want to babysit (their own
  // upkeep, opponent post-combat main, etc.) without holding pass.
  let passingToNextStop = $state(false);
  async function passToNextStop(): Promise<void> {
    if (passingToNextStop || !turn) return;
    passingToNextStop = true;
    const startSeat = activeSeat;
    const startStack = view?.stack?.cards?.length ?? 0;
    let lastSeenSeq = $lastSeq;
    try {
      for (let i = 0; i < 24; i++) {
        const v = $snapshot;
        if (!v) break;
        if (v.turn.active_seat !== startSeat) break;
        if ((v.stack?.cards?.length ?? 0) !== startStack) break;
        // Stop on a configured stop (only meaningful when priority
        // landed on the viewer; other priority-holders' loops are
        // their own concern). The `i > 0` guard skips the very first
        // iteration so pressing the button on a configured stop
        // doesn't immediately exit without doing anything.
        if (i > 0 && $settings.gameplay.stepStops[v.turn.step] === true) break;
        // Server rejects pass_priority during no-priority steps —
        // those auto-advance via the step-entry hook, so we can
        // simply wait for the next snapshot rather than poke.
        if (grantsPriority(v.turn.step)) {
          client.sendAction("pass_priority");
        }
        const before = lastSeenSeq;
        const deadline = Date.now() + 1500;
        while (Date.now() < deadline) {
          await new Promise((r) => setTimeout(r, 25));
          if ($lastSeq > before) {
            lastSeenSeq = $lastSeq;
            break;
          }
        }
        if (lastSeenSeq === before) break;
      }
    } finally {
      passingToNextStop = false;
    }
  }

  function draw(): void {
    if (!viewerID) return;
    client.sendAction("draw_card", viewerID);
  }
  function untapAll(): void {
    if (!viewerID) return;
    client.sendAction("untap_all", viewerID);
  }
  function shuffle(): void {
    if (!viewerID) return;
    client.sendAction("shuffle_library", viewerID);
    play("shuffle");
  }
  let mulliganTo = $state(7);
  let showLifeHistory = $state(false);
  function mulligan(): void {
    if (!viewerID) return;
    const n = Math.max(0, Math.min(20, Math.floor(mulliganTo)));
    client.sendAction("mulligan", viewerID, { hand_size: n });
    play("shuffle");
  }
  function changeLife(delta: number): void {
    if (!viewerID) return;
    client.sendAction("change_life", viewerID, { delta });
  }

  // ---- Combat ----
  // Two-click flow: click an attacker (or blocker) row to "select"
  // it, click a target row to commit. Selection state is local to
  // the viewer's tab; nothing on the wire until the second click
  // dispatches the action.

  type CombatSelection =
    | { kind: "attacker"; cardID: string }
    | { kind: "blocker"; cardID: string }
    | null;
  let combatSelection = $state<CombatSelection>(null);

  // Creatures on the battlefield currently declared as attacking the
  // viewer — used to gate the block-mode flag below so the canvas
  // doesn't enter block mode when there's nothing to block.
  const incomingAttackers = $derived.by(() => {
    if (!viewerID || !view) return [];
    return view.battlefield.cards.filter((c) => c.attacking_target === viewerID);
  });

  // Step-based gating mirrors the server's MTG-rules check. Attackers
  // can only be declared during declare_attackers and only by the
  // active player; blockers only during declare_blockers and only
  // when the viewer is being attacked.
  const canDeclareAttackers = $derived(
    !!turn && turn.step === "declare_attackers" && viewerIsActive,
  );
  const canDeclareBlockers = $derived(
    !!turn && turn.step === "declare_blockers" && incomingAttackers.length > 0,
  );

  // combatMode tells the Board how to interpret battlefield / seat
  // clicks. Derived directly from the step gates so the panel UX
  // tracks the priority/step state without separate state.
  const combatMode = $derived<"idle" | "attack" | "block">(
    canDeclareAttackers ? "attack" : canDeclareBlockers ? "block" : "idle",
  );

  const isAdmin = $derived(sess?.principal.role === "admin");
  // Spectator sessions (S11) are read-only — the server rejects every
  // action frame with bad_request, so the toolbar / mulligan / deck-
  // import / quick-action surfaces all hide here too. Bound by role,
  // not by playerID-being-Nil, so admin spectators (no ?player=)
  // still keep their moderator affordances.
  const isSpectator = $derived(sess?.principal.role === "spectator");
  const selectedCombatCardID = $derived(combatSelection?.cardID ?? null);
  function handleSelectCombatCard(cardID: string): void {
    if (combatMode === "attack") selectAttacker(cardID);
    else if (combatMode === "block") selectBlocker(cardID);
  }

  function selectAttacker(cardID: string): void {
    combatSelection =
      combatSelection?.kind === "attacker" && combatSelection.cardID === cardID
        ? null
        : { kind: "attacker", cardID };
  }
  function selectBlocker(cardID: string): void {
    combatSelection =
      combatSelection?.kind === "blocker" && combatSelection.cardID === cardID
        ? null
        : { kind: "blocker", cardID };
  }
  function declareAttackTarget(targetPlayerID: string): void {
    if (!viewerID || combatSelection?.kind !== "attacker") return;
    client.sendAction("declare_attacker", undefined, {
      attacker: combatSelection.cardID,
      target: targetPlayerID,
    });
    combatSelection = null;
    play("attack");
  }
  function declareBlockTarget(attackerCardID: string): void {
    if (!viewerID || combatSelection?.kind !== "blocker") return;
    client.sendAction("declare_blocker", undefined, {
      blocker: combatSelection.cardID,
      attacker: attackerCardID,
    });
    combatSelection = null;
    play("block");
  }
  // Look up a card's name + controller-name by instance ID. Used to
  // label the combat-hint banner during selection.
  function cardLabel(cardID: string): { name: string; controller: string } {
    if (!view) return { name: "?", controller: "?" };
    const c = view.battlefield.cards.find((x) => x.instance_id === cardID);
    if (!c) return { name: "?", controller: "?" };
    const ctrl = seats.find((s) => s.id === c.controller);
    return { name: c.name, controller: ctrl?.name ?? "?" };
  }

  function keepHand(): void {
    if (!viewerID) return;
    client.sendAction("keep_hand", viewerID);
  }

  function mulliganDecide(): void {
    if (!viewerID) return;
    // Simplified London — redraw to OpeningHandSize (7) every time.
    // No card-to-bottom penalty; that lands with rules enforcement.
    client.sendAction("mulligan", viewerID, { hand_size: 7 });
    play("shuffle");
  }

  function concede(): void {
    if (!viewerID || viewerEliminated || gameEnded) return;
    // Concede is irreversible — confirm to guard against misclicks.
    // window.confirm is acceptable for a hobby-scale sandbox; a proper
    // styled modal can land alongside the S09 polish pass if needed.
    const ok = window.confirm("Concede the game? This cannot be undone.");
    if (!ok) return;
    client.sendAction("concede", viewerID);
  }

  function fmtTime(d: Date): string {
    if (Number.isNaN(d.getTime())) return "";
    const hh = String(d.getHours()).padStart(2, "0");
    const mm = String(d.getMinutes()).padStart(2, "0");
    return `${hh}:${mm}`;
  }
</script>

<section>
  <header>
    <button onclick={back}>← lobby</button>
    <h1>game {gameID.slice(0, 8)}</h1>
    <span class={`tag tag-${$status}`}>{$status}</span>
    {#if isSpectator}
      <span
        class="tag tag-spectator"
        title="read-only — your action frames are rejected by the server">spectating</span
      >
    {/if}
    <span class="muted">seq {$lastSeq}</span>
    <button
      class="gear"
      title="settings (press , from anywhere)"
      aria-label="open settings"
      onclick={openSettings}>⚙</button
    >
  </header>

  {#if viewerNeedsDeck && !deckImportDismissed && viewerID}
    <div
      class="deck-import-modal-backdrop"
      role="dialog"
      aria-modal="true"
      aria-labelledby="deck-import-title"
    >
      <div class="deck-import-modal">
        <h2 id="deck-import-title">import your deck</h2>
        <p class="muted">
          The game hasn't started yet. Paste a Moxfield or Archidekt deck URL, a Moxfield JSON
          export, or a plain-text decklist. The server validates against Commander rules (100-card
          singleton, color identity, format legality).
        </p>
        <DeckUploadForm
          {gameID}
          playerID={viewerID}
          onSuccess={() => (deckImportDismissed = true)}
        />
        <p class="muted import-hint">
          You can also manage decks (and other seats) from the
          <button class="linkish" onclick={back}>lobby</button>.
        </p>
      </div>
    </div>
  {/if}

  {#if mulligansOpen && !gameEnded}
    <div class="mulligan-banner" aria-label="opening hand decisions">
      <strong>Opening hand:</strong>
      {#each seats as seat (seat.id)}
        <span
          class="mulligan-seat"
          class:waiting={!seat.hand_kept && !seat.eliminated}
          style="--seat-color: {seatColor(seat.seat)}"
        >
          <span class="seat-dot" style="background:{seatColor(seat.seat)}"></span>
          {seat.name}
          {#if seat.eliminated}
            <span class="muted">eliminated</span>
          {:else if seat.hand_kept}
            <span class="kept">kept ✓</span>
          {:else}
            <span class="deciding">deciding…</span>
          {/if}
          {#if (seat.mulligans_taken ?? 0) > 0}
            <span class="muted mull-count">×{seat.mulligans_taken}</span>
          {/if}
        </span>
      {/each}
    </div>
  {/if}

  {#if viewerNeedsToDecide}
    <div class="mulligan-dialog" role="dialog" aria-label="keep or mulligan your hand">
      <header>
        <h2>Your opening hand</h2>
        {#if (viewerSeat?.mulligans_taken ?? 0) > 0}
          <p class="muted">
            Mulligans taken: {viewerSeat?.mulligans_taken}. You'll redraw 7 cards (simplified London
            — no bottom-N penalty yet).
          </p>
        {:else}
          <p class="muted">Hand size: {viewerSeat?.hand.count ?? 0}. Keep or mulligan?</p>
        {/if}
      </header>
      {#if (viewerSeat?.hand.cards.length ?? 0) > 0}
        <div class="mulligan-cards" role="list" aria-label="your opening hand">
          {#each viewerSeat?.hand.cards ?? [] as card (card.instance_id)}
            <div class="mulligan-card" role="listitem" title={card.name}>
              {#if card.scryfall_id}
                <img
                  src={`/cards/${card.scryfall_id}/image?size=small`}
                  alt={card.name}
                  loading="lazy"
                />
              {:else}
                <span class="mulligan-card-fallback">{card.name}</span>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
      <div class="mulligan-actions">
        <button class="primary" onclick={keepHand}>Keep hand</button>
        <button onclick={mulliganDecide}>Mulligan</button>
      </div>
    </div>
  {/if}

  {#if manaOverride}
    <div class="error-toast mana-override" role="alert" aria-live="polite">
      <strong>insufficient mana</strong>
      {#if manaOverride.missing.length > 0}
        <span class="muted">missing {manaOverride.missing.join(" ")}</span>
      {/if}
      <button type="button" class="override-btn" onclick={openAutoTap}>Auto-tap & cast</button>
      <button type="button" class="override-btn" onclick={castAnyway}>Cast anyway</button>
      <button
        type="button"
        class="error-toast-close"
        onclick={dismissManaOverride}
        aria-label="dismiss"
      >
        ×
      </button>
    </div>
  {:else if $lastError}
    <div class="error-toast" role="alert" aria-live="polite">
      <strong>server rejected action:</strong>
      {$lastError.message}
      <span class="muted">({$lastError.code})</span>
      <button
        type="button"
        class="error-toast-close"
        onclick={() => lastError.set(null)}
        aria-label="dismiss"
      >
        ×
      </button>
    </div>
  {/if}

  {#if gameEnded}
    <div class="game-end-banner" role="alert">
      {#if winner}
        <strong>
          <span class="seat-dot" style="background:{seatColor(winner.seat)}"></span>
          {winner.name}
        </strong>
        wins the game.
      {:else}
        Game ended — no survivors.
      {/if}
    </div>
  {:else if viewerEliminated}
    <div class="eliminated-banner" role="status">You have been eliminated. Spectating.</div>
  {/if}

  <!-- S16.5: the global turn-bar moved into the self PlayerPanel's
       grid-phases slot as PhaseDisplay. Single source of truth for
       turn/step/priority UI now lives with the player it concerns. -->

  {#if view && viewerID}
    <div class="toolbar" aria-label="quick actions">
      <div class="toolbar-group">
        <button onclick={draw}>draw</button>
        <button onclick={untapAll}>untap all</button>
        <button onclick={shuffle}>shuffle</button>
        <span class="mulligan-group">
          <button onclick={mulligan}>mulligan</button>
          <input
            type="number"
            min="0"
            max="20"
            bind:value={mulliganTo}
            aria-label="mulligan hand size"
          />
        </span>
        <button
          type="button"
          class="mute-toggle"
          onclick={onToggleMute}
          aria-pressed={muted}
          aria-label={muted ? "unmute sound effects" : "mute sound effects"}
          title={muted ? "sounds muted — click to unmute" : "sounds on — click to mute"}
        >
          {muted ? "🔇" : "🔊"}
        </button>
      </div>
      <div class="toolbar-group life-group">
        <button
          class="life-label"
          type="button"
          onclick={() => (showLifeHistory = !showLifeHistory)}
          aria-expanded={showLifeHistory}
          aria-controls="life-history-popover"
          title="click to toggle life-change history"
        >
          life {viewerSeat?.life ?? "—"}
        </button>
        <button onclick={() => changeLife(-5)}>−5</button>
        <button onclick={() => changeLife(-1)}>−1</button>
        <button onclick={() => changeLife(1)}>+1</button>
        <button onclick={() => changeLife(5)}>+5</button>

        {#if showLifeHistory}
          <div class="life-history-popover" id="life-history-popover" role="dialog">
            <header class="life-history-header">
              <span>life history — all seats</span>
              <button
                type="button"
                class="life-history-close"
                onclick={() => (showLifeHistory = false)}
                aria-label="close life history"
              >
                ×
              </button>
            </header>
            <div class="life-history-body">
              {#each seats as seat (seat.id)}
                <section class="life-history-seat">
                  <h4 class="life-history-seat-name">
                    <span class="seat-dot" style="background:{seatColor(seat.seat)}"></span>
                    {seat.name}
                    <span class="muted">· now {seat.life}</span>
                  </h4>
                  {#if (seat.life_history?.length ?? 0) === 0}
                    <p class="muted life-history-empty">no changes yet</p>
                  {:else}
                    <ol class="life-history-entries">
                      {#each seat.life_history.slice().reverse() as entry, idx (idx)}
                        <li>
                          <span
                            class="life-delta"
                            class:gain={entry.delta > 0}
                            class:loss={entry.delta < 0}
                          >
                            {entry.delta > 0 ? "+" : ""}{entry.delta}
                          </span>
                          <span class="life-newtotal">→ {entry.new_total}</span>
                          <span class="life-time muted">{fmtTime(new Date(entry.at))}</span>
                        </li>
                      {/each}
                    </ol>
                  {/if}
                </section>
              {/each}
            </div>
          </div>
        {/if}
      </div>
      <div class="toolbar-group priority-controls">
        <button
          onclick={advanceStep}
          disabled={!viewerIsActive}
          class:advance-step={viewerIsActive}
          title={viewerIsActive
            ? "advance the step cursor by one (sandbox shortcut — bypasses opponent priority pass)"
            : `${activePlayer?.name ?? "another seat"} is the active player`}
        >
          next step
        </button>
        <button
          onclick={passPriority}
          disabled={!viewerHasPriority}
          class:viewer-priority={viewerHasPriority}
          title={viewerHasPriority
            ? "pass priority — rotates to next seat"
            : `${priorityPlayer?.name ?? "another seat"} holds priority`}
        >
          pass priority
        </button>
        <button
          onclick={passToNextStop}
          disabled={passingToNextStop || !viewerHasPriority}
          title={viewerHasPriority
            ? "auto-pass priority until the next step you've pinned in Settings"
            : `${priorityPlayer?.name ?? "another seat"} holds priority`}
        >
          {passingToNextStop ? "passing…" : "→ next stop"}
        </button>
        <button
          onclick={passToEnd}
          disabled={passingToEnd || !viewerHasPriority}
          title={viewerHasPriority
            ? "pass priority repeatedly until end of this turn"
            : `${priorityPlayer?.name ?? "another seat"} holds priority`}
        >
          {passingToEnd ? "passing…" : "pass until end of turn"}
        </button>
        <button
          onclick={passTurn}
          disabled={!viewerIsActive}
          title={viewerIsActive
            ? "skip the rest of your turn"
            : `${activePlayer?.name ?? "another seat"} is the active player`}
        >
          pass turn
        </button>
        <button
          onclick={() => client.sendAction("undo")}
          disabled={!isAdmin && (viewerSeat?.undos_remaining ?? 0) <= 0}
          title={isAdmin
            ? "rewind the most recent action (admin — bypasses caller / budget gates)"
            : (viewerSeat?.undos_remaining ?? 0) <= 0
              ? "no undos remaining this turn (refreshes on your next untap)"
              : `undo your most recent action — ${viewerSeat?.undos_remaining ?? 0} left this turn`}
        >
          undo {!isAdmin && viewerSeat ? `(${viewerSeat.undos_remaining ?? 0})` : ""}
        </button>
        <label
          class="undo-limit"
          title="per-player undo budget refreshed each turn (any seat may change)"
        >
          limit
          <input
            type="number"
            min="0"
            max="20"
            value={view?.undo_limit ?? 1}
            onchange={(e) => {
              const next = Number((e.currentTarget as HTMLInputElement).value);
              if (Number.isFinite(next) && next >= 0) {
                client.sendAction("set_undo_limit", undefined, { limit: next });
              }
            }}
          />
        </label>
        <button
          onclick={concede}
          disabled={viewerEliminated || gameEnded}
          class="concede"
          title={viewerEliminated
            ? "you are already eliminated"
            : gameEnded
              ? "the game has ended"
              : "concede the game (irreversible)"}
        >
          concede
        </button>
      </div>
    </div>
  {/if}

  {#if combatSelection && !mulligansOpen}
    <div class="combat-hint" role="status" aria-live="polite">
      {#if combatSelection.kind === "attacker"}
        Attacking with <strong>{cardLabel(combatSelection.cardID).name}</strong> — click an opponent's
        seat to commit, or click the creature again to cancel.
      {:else}
        Blocking with <strong>{cardLabel(combatSelection.cardID).name}</strong> — click an incoming attacker
        to commit, or click the creature again to cancel.
      {/if}
      <button class="combat-cancel" onclick={() => (combatSelection = null)}>cancel</button>
    </div>
  {/if}

  <!--
    Chat UI removed in S08.5 wave 1 — players coordinate on Discord
    during play. The wire protocol (KindChat, ChatPayload),
    GameClient.chat store, GameClient.sendChat method, and the
    server-side handleChat/broadcastChat path all stay live so a
    future sprint can re-mount a panel here without touching the
    transport layer.
  -->
  <div class="play-area">
    {#if view}
      <Board
        {view}
        {viewerID}
        {isAdmin}
        {sendAction}
        {combatMode}
        {selectedCombatCardID}
        onSelectCombatCard={handleSelectCombatCard}
        onDeclareAttack={declareAttackTarget}
        onDeclareBlock={declareBlockTarget}
      />
      <DiscardPromptModal snap={view} {viewerID} {sendAction} />
      <ChoicePromptModal snap={view} {viewerID} {sendAction} />
      <AutoTapPreviewModal
        {gameID}
        snap={view}
        cardID={autoTapCardID}
        onConfirm={confirmAutoTap}
        onCancel={cancelAutoTap}
      />
      <TargetingBanner />
    {/if}
  </div>

  {#if !view}
    <p class="muted centered">waiting for snapshot…</p>
  {/if}
</section>

<svelte:window
  onkeydown={(ev: KeyboardEvent) => {
    if (ev.key === "Escape") cancelTargeting();
  }}
/>

<style>
  /* The Game route needs the whole viewport, not the 800px column
     #app imposes on the lobby / login screens. Pinning the section
     with position: fixed + inset: 0 sidesteps #app entirely — no
     global override, no CSS load-order fight, and it's scoped to
     this component so the cap is restored automatically on navigate
     away. overflow: hidden keeps the outer scrollbar off when
     child layout is tight. */
  section {
    position: fixed;
    inset: 0;
    padding: 0.45rem 0.6rem;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    overflow: hidden;
  }
  header {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    margin: 0;
  }
  header :global(h1) {
    font-size: 1rem;
    font-weight: 600;
    letter-spacing: -0.01em;
    margin: 0;
  }
  .gear {
    margin-left: auto;
    background: none;
    border: 1px solid transparent;
    color: var(--fg-muted);
    font-size: 1.1rem;
    cursor: pointer;
    padding: 0.25rem 0.5rem;
    border-radius: var(--radius);
    line-height: 1;
    box-shadow: none;
    transition:
      background 120ms var(--ease),
      color 120ms var(--ease),
      border-color 120ms var(--ease);
  }
  .gear:hover {
    color: var(--fg);
    background: rgba(255, 255, 255, 0.06);
    border-color: rgba(255, 255, 255, 0.1);
  }
  .play-area {
    /* Single-column layout since S08.5 removed the chat sidebar.
       The 1fr row + min-height: 0 combo lets the Board sit at the
       full available height without collapsing to fit-content. */
    display: grid;
    grid-template-columns: 1fr;
    grid-template-rows: 1fr;
    align-items: stretch;
    flex: 1;
    min-height: 0;
    min-width: 0;
    /* position: relative anchors the Board's absolutely-positioned
       overlays (HoverZoomOverlay, StackOverlay) to this container
       rather than to the viewport. */
    position: relative;
  }
  .muted {
    color: var(--fg-dim);
  }
  .centered {
    text-align: center;
    margin-top: 1rem;
  }
  .tag {
    padding: 0.15rem 0.5rem;
    border-radius: 999px;
    font-size: 0.72em;
    font-weight: 600;
    letter-spacing: 0.04em;
    border: 1px solid transparent;
  }
  .tag-connected {
    background: rgba(122, 255, 154, 0.12);
    color: var(--mint);
    border-color: rgba(122, 255, 154, 0.35);
  }
  .tag-connecting {
    background: rgba(255, 208, 122, 0.12);
    color: var(--gold);
    border-color: rgba(255, 208, 122, 0.35);
  }
  .tag-disconnected {
    background: rgba(255, 122, 122, 0.12);
    color: var(--danger);
    border-color: rgba(255, 122, 122, 0.35);
  }
  .tag-spectator {
    background: rgba(176, 138, 255, 0.15);
    color: #b08aff;
    border: 1px solid rgba(176, 138, 255, 0.5);
    text-transform: uppercase;
    letter-spacing: 0.1em;
    font-size: 0.7em;
    font-weight: 700;
  }

  /* Mulligan window */
  .mulligan-banner {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem 0.75rem;
    align-items: center;
    padding: 0.55rem 0.9rem;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.03) 0%, rgba(0, 0, 0, 0.15) 100%),
      var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--fg-muted);
    font-size: 0.9em;
    box-shadow: var(--shadow-sm);
  }
  .mulligan-seat {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
  }
  .mulligan-seat.waiting {
    border-bottom: 2px dotted var(--seat-color);
    padding-bottom: 0.05rem;
  }
  .kept {
    color: #b3e5b3;
    font-weight: 600;
  }
  .deciding {
    color: #ffd07a;
    font-style: italic;
  }
  .mull-count {
    font-size: 0.85em;
  }
  .mulligan-dialog {
    padding: 0.9rem 1.1rem;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.03) 0%, rgba(0, 0, 0, 0.2) 100%), var(--surface);
    border: 1px solid rgba(179, 229, 179, 0.55);
    border-radius: var(--radius-lg);
    color: var(--fg);
    box-shadow:
      0 10px 28px rgba(0, 0, 0, 0.45),
      0 0 20px rgba(179, 229, 179, 0.14),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
  }
  .mulligan-dialog header h2 {
    margin: 0 0 0.25rem 0;
    font-size: 1.05em;
  }
  .mulligan-dialog header p {
    margin: 0 0 0.6rem 0;
    font-size: 0.9em;
  }
  .mulligan-actions {
    display: flex;
    gap: 0.5rem;
  }
  .mulligan-actions button {
    padding: 0.4rem 1rem;
    font-size: 0.95em;
  }
  .mulligan-actions button.primary {
    background: linear-gradient(180deg, #c3f0c3 0%, #8acc8a 100%);
    color: #0c1426;
    font-weight: 700;
    border: 1px solid rgba(138, 204, 138, 0.7);
    box-shadow:
      0 6px 14px rgba(138, 204, 138, 0.25),
      inset 0 1px 0 rgba(255, 255, 255, 0.35);
  }
  .mulligan-actions button.primary:hover {
    filter: brightness(1.04);
  }
  .mulligan-cards {
    display: flex;
    gap: 0.4rem;
    overflow-x: auto;
    padding: 0.4rem 0;
    margin-bottom: 0.6rem;
  }
  .mulligan-card {
    flex: 0 0 auto;
    width: 96px;
    height: 134px;
    border-radius: var(--radius);
    overflow: hidden;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.35);
  }
  .mulligan-card img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .mulligan-card-fallback {
    display: flex;
    width: 100%;
    height: 100%;
    align-items: center;
    justify-content: center;
    padding: 0.3rem;
    text-align: center;
    font-size: 0.75em;
    color: #cfd6ee;
  }

  /* Combat: just the active selection hint. The actual combat UI
     is canvas-only — click your creature, click an opponent's seat. */
  .combat-hint {
    padding: 0.55rem 0.9rem;
    background:
      linear-gradient(180deg, rgba(80, 60, 0, 0.55) 0%, rgba(40, 28, 0, 0.55) 100%), var(--surface);
    border: 1px solid rgba(200, 168, 106, 0.7);
    border-radius: var(--radius);
    color: var(--gold);
    font-size: 0.9em;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    box-shadow:
      0 4px 12px rgba(0, 0, 0, 0.3),
      0 0 14px rgba(255, 208, 122, 0.12),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
  }
  .combat-hint strong {
    color: #fff;
  }
  .combat-cancel {
    margin-left: auto;
    padding: 0.15rem 0.6rem;
    font-size: 0.85em;
    background: #2a3550;
    color: #cfd6ee;
    border: 1px solid #3a4570;
  }

  /* Server-error toast */
  .error-toast {
    padding: 0.55rem 0.9rem;
    background: linear-gradient(180deg, rgba(90, 30, 30, 0.9) 0%, rgba(60, 18, 18, 0.9) 100%);
    border: 1px solid rgba(255, 122, 122, 0.5);
    border-radius: var(--radius);
    color: #ffd0d0;
    font-size: 0.9em;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    box-shadow:
      0 6px 18px rgba(0, 0, 0, 0.35),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
  }
  .error-toast strong {
    color: #fff;
  }
  .error-toast-close {
    margin-left: auto;
    background: transparent;
    border: none;
    color: #ffd0d0;
    font-size: 1.2em;
    line-height: 1;
    cursor: pointer;
    padding: 0 0.25rem;
  }
  /* S15 strict-mode override toast — same shell as the error
     toast but with the gold "Cast anyway" affordance. */
  .error-toast.mana-override {
    background: linear-gradient(180deg, rgba(78, 56, 18, 0.95) 0%, rgba(54, 36, 8, 0.95) 100%);
    border-color: rgba(200, 168, 106, 0.65);
    color: #f4ead5;
  }
  .override-btn {
    margin-left: 0.5rem;
    padding: 0.3rem 0.85rem;
    border-radius: 999px;
    background: linear-gradient(180deg, #ffe59a 0%, #e6b85f 100%);
    color: #231806;
    border: 1px solid rgba(255, 230, 160, 0.6);
    font-weight: 800;
    letter-spacing: 0.02em;
    cursor: pointer;
    font-size: 0.85em;
  }
  .override-btn:hover {
    filter: brightness(1.05);
  }

  /* End-of-game banners */
  .game-end-banner {
    padding: 0.75rem 1rem;
    background: linear-gradient(90deg, rgba(42, 90, 42, 0.65) 0%, rgba(26, 37, 64, 0.65) 100%);
    border: 1px solid rgba(122, 255, 154, 0.45);
    border-radius: var(--radius);
    color: #e0ffe0;
    font-size: 1.05em;
    text-align: center;
    box-shadow:
      0 8px 22px rgba(0, 0, 0, 0.35),
      0 0 18px rgba(122, 255, 154, 0.15),
      inset 0 1px 0 rgba(255, 255, 255, 0.08);
    font-weight: 600;
  }
  .game-end-banner strong {
    color: #fff;
    margin-right: 0.3rem;
  }
  .eliminated-banner {
    padding: 0.55rem 0.9rem;
    background: linear-gradient(180deg, rgba(58, 26, 26, 0.85) 0%, rgba(42, 18, 18, 0.85) 100%);
    border: 1px solid rgba(255, 122, 122, 0.4);
    border-radius: var(--radius);
    color: #ffd0d0;
    font-size: 0.95em;
    text-align: center;
    box-shadow:
      0 4px 14px rgba(0, 0, 0, 0.3),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
  }
  .toolbar .undo-limit {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 10px;
    color: var(--fg-dim);
    text-transform: lowercase;
  }
  .toolbar .undo-limit input {
    width: 36px;
    padding: 2px 6px;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    color: var(--fg);
    border-radius: var(--radius-sm);
    font: inherit;
    font-size: 11px;
    margin: 0;
  }
  .toolbar button.concede {
    margin-left: 0.5rem;
    background: linear-gradient(180deg, rgba(90, 30, 30, 0.9) 0%, rgba(60, 18, 18, 0.9) 100%);
    color: #ffd0d0;
    border: 1px solid rgba(255, 122, 122, 0.4);
    box-shadow: none;
  }
  .toolbar button.concede:hover:not(:disabled) {
    background: linear-gradient(180deg, rgba(110, 40, 40, 0.95) 0%, rgba(80, 24, 24, 0.95) 100%);
    border-color: rgba(255, 122, 122, 0.65);
  }
  .toolbar button.concede:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  /* Toolbar */
  .toolbar {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
    align-items: center;
    padding: 0.4rem 0.6rem;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.04) 0%, rgba(0, 0, 0, 0.15) 100%),
      var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--fg-muted);
    font-size: 0.8em;
    box-shadow: var(--shadow-sm);
  }
  .toolbar button {
    padding: 0.3rem 0.7rem;
    font-size: 0.9em;
    background: linear-gradient(180deg, var(--bg-3) 0%, var(--bg-2) 100%);
    color: var(--fg);
    border: 1px solid var(--border);
    font-weight: 500;
    box-shadow: none;
    transition:
      background 120ms var(--ease),
      border-color 120ms var(--ease),
      transform 120ms var(--ease);
  }
  .toolbar button:hover:not(:disabled) {
    background: linear-gradient(180deg, var(--bg-2) 0%, var(--bg-3) 100%);
    border-color: var(--border-strong);
  }
  .toolbar button:disabled {
    background: rgba(0, 0, 0, 0.25);
    border-color: var(--border);
    color: var(--fg-dim);
  }
  .toolbar-group {
    display: flex;
    gap: 0.35rem;
    align-items: center;
  }
  .mulligan-group input {
    width: 3em;
    padding: 0.1rem 0.3rem;
  }
  .mute-toggle {
    padding: 0.15rem 0.45rem;
    font-size: 0.95em;
    line-height: 1;
  }
  .mute-toggle[aria-pressed="true"] {
    background: #3a1a1a;
    color: #ffd0d0;
    border: 1px solid #6a3a3a;
  }
  .life-label {
    color: #e0e8ff;
    margin-right: 0.25rem;
    background: transparent;
    border: 1px solid transparent;
    padding: 0.15rem 0.4rem;
    cursor: pointer;
    font: inherit;
  }
  .life-label:hover {
    border-color: #2a3550;
    border-radius: 3px;
  }
  .life-group {
    position: relative;
  }
  .life-history-popover {
    position: absolute;
    top: calc(100% + 0.5rem);
    left: 0;
    z-index: 5;
    width: min(320px, calc(100vw - 2rem));
    max-height: 360px;
    background: linear-gradient(180deg, rgba(19, 26, 44, 0.96) 0%, rgba(8, 12, 24, 0.96) 100%);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(122, 167, 255, 0.22);
    border-radius: var(--radius-lg);
    color: var(--fg-muted);
    font-size: 0.85em;
    display: flex;
    flex-direction: column;
    box-shadow:
      0 12px 32px rgba(0, 0, 0, 0.55),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
  }
  .life-history-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 0.75rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.06);
    text-transform: uppercase;
    letter-spacing: 0.12em;
    font-size: 0.75em;
    color: var(--fg-dim);
    font-weight: 700;
  }
  .life-history-close {
    background: transparent;
    border: none;
    color: #888;
    font-size: 1.2em;
    line-height: 1;
    cursor: pointer;
    padding: 0 0.25rem;
  }
  .life-history-body {
    overflow-y: auto;
    padding: 0.4rem 0.6rem;
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }
  .life-history-seat-name {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin: 0 0 0.25rem 0;
    font-size: 0.95em;
    font-weight: 600;
    color: #e0e8ff;
  }
  .life-history-empty {
    margin: 0 0 0 1.1rem;
    font-size: 0.85em;
  }
  .life-history-entries {
    list-style: none;
    margin: 0;
    padding: 0 0 0 1.1rem;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }
  .life-history-entries li {
    display: grid;
    grid-template-columns: 3rem 4rem 1fr;
    align-items: baseline;
    column-gap: 0.4rem;
  }
  .life-delta {
    font-variant-numeric: tabular-nums;
    font-weight: 600;
    text-align: right;
  }
  .life-delta.gain {
    color: #b3e5b3;
  }
  .life-delta.loss {
    color: #ffadad;
  }
  .life-newtotal {
    color: #cfd6ee;
    font-variant-numeric: tabular-nums;
  }
  .life-time {
    font-size: 0.85em;
  }
  .priority-controls .viewer-priority {
    background: linear-gradient(180deg, #b3e5b3 0%, #7fc87f 100%);
    color: #0a1a0a;
    font-weight: 700;
    border-color: rgba(127, 200, 127, 0.6);
    box-shadow:
      0 4px 12px rgba(127, 200, 127, 0.3),
      inset 0 1px 0 rgba(255, 255, 255, 0.3);
  }
  .priority-controls .viewer-priority:hover:not(:disabled) {
    background: linear-gradient(180deg, #c4f0c4 0%, #8fd88f 100%);
    border-color: rgba(127, 200, 127, 0.85);
  }
  .priority-controls .advance-step {
    background: linear-gradient(180deg, #9bbfff 0%, #5a88ee 100%);
    color: #0a1024;
    font-weight: 700;
    border-color: rgba(90, 136, 238, 0.65);
    box-shadow:
      0 4px 12px rgba(90, 136, 238, 0.3),
      inset 0 1px 0 rgba(255, 255, 255, 0.3);
  }
  .priority-controls .advance-step:hover:not(:disabled) {
    background: linear-gradient(180deg, #b4d0ff 0%, #7098f0 100%);
    border-color: rgba(90, 136, 238, 0.9);
  }

  /* Pre-game deck import modal (S08.5 wave 1) — overlays the table
     for a freshly-joined player who hasn't uploaded a deck yet. The
     light card-on-dark-backdrop palette borrows from the lobby's
     deck-upload panel rather than the table's chrome (#1a2540) so
     the form's #fee / #ffb violation banners stay readable. */
  .deck-import-modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(4, 8, 16, 0.7);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 50;
    padding: 1rem;
  }
  .deck-import-modal {
    background: linear-gradient(180deg, var(--surface) 0%, var(--bg-2) 100%);
    color: var(--fg);
    border: 1px solid rgba(122, 167, 255, 0.22);
    border-radius: var(--radius-xl);
    padding: 1.4rem 1.6rem;
    max-width: 640px;
    width: 100%;
    max-height: calc(100vh - 2rem);
    overflow-y: auto;
    box-shadow:
      0 30px 80px rgba(0, 0, 0, 0.7),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
  }
  .deck-import-modal h2 {
    margin: 0 0 0.5rem 0;
    font-size: 1.15em;
  }
  .import-hint {
    margin-top: 1rem;
    font-size: 0.85em;
  }
  .linkish {
    background: none;
    border: none;
    color: var(--accent-strong);
    text-decoration: underline;
    cursor: pointer;
    padding: 0;
    font: inherit;
    box-shadow: none;
  }
  .linkish:hover {
    color: var(--accent);
    background: transparent;
    box-shadow: none;
  }
</style>
