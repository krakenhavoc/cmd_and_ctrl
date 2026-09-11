<script lang="ts">
  import { onMount } from "svelte";
  import { GameClient } from "../lib/ws";
  import { navigate } from "../lib/router";
  import { session } from "../lib/session";
  import { seatColor } from "../lib/colors";
  import DeckUploadForm from "../lib/components/DeckUploadForm.svelte";
  import BugReportModal from "../lib/components/BugReportModal.svelte";
  import { fetchBugReportConfig } from "../lib/api";
  import Board from "../lib/components/board/Board.svelte";
  import DiscardPromptModal from "../lib/components/board/DiscardPromptModal.svelte";
  import ChoicePromptModal from "../lib/components/board/ChoicePromptModal.svelte";
  import AutoTapPreviewModal from "../lib/components/board/AutoTapPreviewModal.svelte";
  import TargetingBanner from "../lib/components/board/TargetingBanner.svelte";
  import Icon from "../lib/components/Icon.svelte";
  import { cancel as cancelTargeting, confirm as confirmTargeting } from "../lib/targeting";
  import type { ActionType, PlayerView } from "../lib/protocol";
  import type { StepID } from "../lib/turn";
  import { armAudioOnFirstGesture, isMuted, play, toggleMuted } from "../lib/sounds";
  import { openSettings, settings } from "../lib/settings";
  import { hasAnyLegalResponse } from "../lib/priority";
  import {
    attackAllLabel,
    attackAllParams,
    blockedSummary,
    planAttackAll,
    seatLabel,
  } from "../lib/attackAll";
  import { stackEmpty } from "../lib/timing";
  import { consumeManualStop, manualStops } from "../lib/priorityStops";

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
  const { status, snapshot, lastSeq, lastError, log } = client;

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

  // Report-a-bug affordance (ADR 0017). Probed once per mount: the
  // server says whether it can file GitHub issues; when it can't
  // (CMDCTRL_GITHUB_TOKEN unset) the button never renders — same
  // hide-don't-503 posture as the Discord sign-in button.
  let bugReportAvailable = $state(false);
  // Attachment support is probed alongside the feature flag: a server
  // with a GitHub token but no data dir files text reports fine and
  // can't host screenshots, so the modal hides its picker rather than
  // offering an upload that 503s.
  let bugReportAttachments = $state(false);
  let bugReportOpen = $state(false);
  onMount(() => {
    void fetchBugReportConfig().then((cfg) => {
      bugReportAvailable = cfg.enabled;
      bugReportAttachments = cfg.attachments;
    });
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
  // S13.6: autopass mode is a session-scoped toggle ("get me
  // through this turn" / "I'm tapped out, don't ask me"). Stays on
  // until the viewer clicks the button again — not a one-shot.
  // When on, overrides every other gate: settings.autoPassPriority,
  // stepStops grid, smartAutoPass predicate, manual pins. The
  // effect still requires the viewer to actually hold priority (so
  // we don't spam the server with "you do not hold priority"
  // rejections on opponents' turns — the toggle is ambient intent,
  // not a manual pass loop).
  let autopassEnabled = $state(false);
  let lastAutoPassedSeq = $state(-1);
  $effect(() => {
    if (!viewerHasPriority) return;
    if (mulligansOpen || gameEnded || viewerEliminated) return;
    // Never auto-pass priority while the viewer has an open choice to
    // make (e.g. an optional trigger's yes/no prompt). The choice
    // persists server-side regardless, but auto-passing here would
    // race the modal and let priority slip away before the player
    // answers. The chooser must resolve their pending choice first.
    if (view?.pending_choices?.some((c) => c.chooser === viewerID)) return;
    const step = view?.turn?.step;
    if (!step) return;

    const autopass = autopassEnabled;

    // S13.6 safety belt (gameplay.autopassPersistThroughTurns): when
    // the flag is off (default), autopass auto-clears the first time
    // the cursor enters the viewer's own precombat_main — so a
    // forgotten toggle doesn't silently skip your turn. Users who
    // know they want autopass to outlive their own main phase flip
    // the danger setting on and accept the trade. We clear BEFORE
    // firing pass_priority so the toggle going off means the cursor
    // holds for the viewer's turn.
    if (
      autopass &&
      step === "precombat_main" &&
      viewerIsActive &&
      !$settings.gameplay.autopassPersistThroughTurns
    ) {
      autopassEnabled = false;
      return;
    }

    // Conventional (non-autopass) path: honour every gate.
    if (!autopass) {
      if (!$settings.gameplay.autoPassPriority) return;
      // Anything on the stack stops: a spell card OR an ability
      // item (S19 triggers have no card on Game.Stack, only a
      // stack_items entry). Autopass off means Arena-style "full
      // control" — every trigger is a window the viewer gets to
      // answer. The autopass toggle above is the way through
      // routine upkeep triggers.
      if (!stackEmpty(view)) return;
      // Manual one-time stops override everything below. Click a
      // phase icon in PhaseDisplay to pin; the pin clears on step
      // transition. "Fake a game action" — viewer gets the cursor
      // even when the engine has nothing to offer (want to think /
      // bluff / respond off-catalog). Read via the $manualStops
      // subscription (not the non-reactive hasManualStop helper) so
      // unpinning while holding priority re-runs this effect and
      // resumes auto-pass immediately rather than on the next
      // snapshot.
      if ($manualStops.has(step as StepID)) return;
      // Stop here if the viewer has opted to stop on this step. The
      // map omits no-priority steps (Untap / Cleanup); for those,
      // the viewer can never hold priority anyway. `smartAutoPass`
      // adds an escape hatch: if the stop lands on the viewer but
      // the legality engine reports no legal response, pass anyway.
      // Predicate errs conservative (false-positive-stop > false-
      // negative-skip per ADR 0009 §3).
      if ($settings.gameplay.stepStops[step] === true) {
        const smart = $settings.gameplay.smartAutoPass;
        const canRespond = smart ? hasAnyLegalResponse(view, viewerID, $lastSeq) : true;
        if (canRespond) return;
      }
    }

    // Dedupe by snapshot seq so we don't fire twice on the same
    // priority window if the effect re-runs for an unrelated reason
    // before the next snapshot lands.
    const seq = $lastSeq;
    if (seq === lastAutoPassedSeq) return;
    lastAutoPassedSeq = seq;
    client.sendAction("pass_priority");
  });

  function toggleAutopass(): void {
    autopassEnabled = !autopassEnabled;
  }

  // S13.6: consume manual one-time stops. When the snapshot step
  // advances away from a pinned step, clear the pin — the whole
  // point of "one-time" is that the next cycle of that step isn't
  // held unless re-pinned. No-op for any step that wasn't pinned.
  let lastStep = $state<string | null>(null);
  $effect(() => {
    const step = view?.turn?.step ?? null;
    if (step === lastStep) return;
    if (lastStep) consumeManualStop(lastStep);
    lastStep = step;
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
  // lastCastByCardID stashes the most recent cast_spell payload per
  // instance_id so retry dispatchers (castAnyway after insufficient_mana,
  // confirmAutoTap) can replay the ORIGINAL payload with added flags.
  // Fixes a bug where re-firing a cast with force_cast=true dropped the
  // user's already-picked target[s] / modes / x_value / distribution,
  // leaving the spell to resolve as a silent no-op.
  const lastCastByCardID = new Map<string, Record<string, unknown>>();

  const sendAction = (type: ActionType, params?: unknown, player?: string): void => {
    if (type === "cast_spell") {
      const strict = $settings.gameplay.strictMana;
      const incoming = (params ?? {}) as Record<string, unknown>;
      if (incoming.strict === undefined) {
        params = { ...incoming, strict };
      }
      // Stash by instance_id so retry paths can replay targets etc.
      const stash = params as Record<string, unknown>;
      const instanceID = stash.instance_id;
      if (typeof instanceID === "string") {
        lastCastByCardID.set(instanceID, { ...stash });
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
    // Replay the original cast payload so targets / modes / X /
    // distribution survive the retry. Fallback to a bare payload
    // if the stash is missing (shouldn't happen — cast_spell
    // always stashes before the error round-trip).
    const prev = lastCastByCardID.get(cardID) ?? { instance_id: cardID };
    sendAction("cast_spell", { ...prev, strict: true, force_cast: true }, viewerID ?? undefined);
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
    // Same replay-original-payload pattern as castAnyway — the
    // auto-tap retry path also dropped targets until this fix.
    const prev = lastCastByCardID.get(cardID) ?? { instance_id: cardID };
    sendAction(
      "cast_spell",
      { ...prev, strict: true, auto_tap: true, locked_sources: lockedSources },
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
  // Spectators never have a viewerID, regardless of what the session
  // happens to carry. The server's /spectate handler doesn't set
  // player_id, but stale localStorage state, legacy player sessions
  // demoted to spectator, and future code paths that surface a
  // spectator-scoped id all get collapsed here to null so every
  // downstream viewer-action gate (`{#if viewerID}`, `?? undefined`
  // casts, targeting logic) uniformly denies spectators. Cheaper and
  // safer than threading `!isSpectator` through every call site.
  const viewerID = $derived(sess?.principal.role === "spectator" ? null : (sess?.playerID ?? null));
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
  // ---- Attack with all (#318) ----
  // Declaring a wide board one creature at a time is the loudest
  // ergonomics complaint from live play. The cluster below sits in
  // the attention strip for the whole declare-attackers step and
  // offers one button per attackable opponent.
  //
  // "Attack all" is ambiguous at a Commander table, so this never
  // guesses: it does NOT spread creatures across opponents. Every
  // eligible creature goes at ONE named seat, and the seat's name is
  // on the button the player presses. Splitting an attack is a
  // strategic choice with no defensible default, so it stays on the
  // two-click flow and the per-card context menu.
  //
  // The whole set goes out as a single `declare_attackers` action.
  // Looping the per-creature verb would push one undo entry and one
  // broadcast per creature — a twelve-creature alpha strike would
  // need twelve undo presses against a per-turn budget of one. One
  // action means one snapshot and one exact inverse.
  const attackPlan = $derived(planAttackAll(view, viewerID));
  const attackAllReady = $derived(
    canDeclareAttackers && attackPlan.eligible.length > 0 && attackPlan.defenders.length > 0,
  );
  const attackBlockedHint = $derived(blockedSummary(attackPlan.blocked));

  function attackAllAt(defenderSeatID: string): void {
    const params = attackAllParams(attackPlan, defenderSeatID);
    if (!params) return;
    combatSelection = null;
    client.sendAction("declare_attackers", undefined, params);
    play("attack");
  }

  // The inverse of a wide declaration is undo, not a bulk "unattack":
  // nothing in the engine records which creatures the declaration
  // tapped, so clearing declarations afterwards would strand them
  // tapped and not attacking — strictly worse than never having
  // clicked. Because the bulk declare is one room.Apply, a single
  // undo restores tap state and declarations together. That is why
  // this button is here rather than only in the ⋯ menu.
  const canUndoDeclaration = $derived(
    canDeclareAttackers &&
      attackPlan.declared.length > 0 &&
      (isAdmin || (viewerSeat?.undos_remaining ?? 0) > 0),
  );
  function undoDeclaration(): void {
    client.sendAction("undo");
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

  // Concede is irreversible — it asks first, in a styled popover
  // anchored to the command bar rather than window.confirm.
  let concedeConfirm = $state(false);
  function requestConcede(): void {
    if (!viewerID || viewerEliminated || gameEnded) return;
    menuOpen = false;
    concedeConfirm = true;
  }
  function confirmConcede(): void {
    concedeConfirm = false;
    if (!viewerID || viewerEliminated || gameEnded) return;
    client.sendAction("concede", viewerID);
  }

  // The ⋯ menu in the command bar holds the sandbox utilities
  // (draw / untap / shuffle / mulligan / undo) and table actions so
  // the bar itself stays status + three icon buttons + pass turn.
  let menuOpen = $state(false);
  function closeMenu(): void {
    menuOpen = false;
  }
  function viaMenu(fn: () => void): void {
    menuOpen = false;
    fn();
  }

  function fmtTime(d: Date): string {
    if (Number.isNaN(d.getTime())) return "";
    const hh = String(d.getHours()).padStart(2, "0");
    const mm = String(d.getMinutes()).padStart(2, "0");
    return `${hh}:${mm}`;
  }
</script>

<section>
  <header class="bar">
    <button class="ghost bar-nav" onclick={back}><Icon name="chevronLeft" size={14} /> Lobby</button
    >
    <span class="bar-sep" aria-hidden="true"></span>
    <span class="wordmark" aria-hidden="true"><i></i>CMD &amp; CTRL</span>
    <h1 class="crumb"><b>/</b><span class="mono">{gameID.slice(0, 8)}</span></h1>
    <span class="bar-spacer"></span>
    {#if isSpectator}
      <span
        class="tag tag-spectator"
        title="read-only — your action frames are rejected by the server">spectating</span
      >
    {/if}
    <span class={`status status-${$status}`} title={`seq ${$lastSeq}`}>
      <i class="dot" aria-hidden="true"></i>{$status}
      <span class="seq">· seq {$lastSeq}</span>
    </span>
    {#if view && viewerID}
      <button
        class="bar-btn"
        onclick={passTurn}
        disabled={!viewerIsActive}
        title={viewerIsActive
          ? "skip the rest of your turn"
          : `${activePlayer?.name ?? "another seat"} is the active player`}
      >
        Pass turn
      </button>
    {/if}
    <div class="bar-icons">
      <button
        type="button"
        class="ibtn"
        onclick={onToggleMute}
        aria-pressed={muted}
        aria-label={muted ? "unmute sound effects" : "mute sound effects"}
        title={muted ? "sounds muted — click to unmute" : "sounds on — click to mute"}
      >
        {#if muted}<Icon name="volumeOff" size={17} />{:else}<Icon name="volume" size={17} />{/if}
      </button>
      <button
        class="ibtn"
        title="settings (press , from anywhere)"
        aria-label="open settings"
        onclick={openSettings}><Icon name="gear" size={17} /></button
      >
      {#if view && viewerID}
        <div class="more">
          <button
            class="ibtn"
            class:on={menuOpen}
            aria-haspopup="menu"
            aria-expanded={menuOpen}
            aria-label="more actions"
            title="sandbox actions and more"
            onclick={() => (menuOpen = !menuOpen)}><Icon name="more" size={17} /></button
          >
          {#if menuOpen}
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div class="menu-backdrop" onclick={closeMenu}></div>
            <div class="menu" role="menu" aria-label="game actions">
              <div class="menu-h">Sandbox</div>
              <button class="mi" role="menuitem" onclick={() => viaMenu(draw)}>
                <Icon name="draw" size={15} /> Draw a card
              </button>
              <button class="mi" role="menuitem" onclick={() => viaMenu(untapAll)}>
                <Icon name="untap" size={15} /> Untap all
              </button>
              <button class="mi" role="menuitem" onclick={() => viaMenu(shuffle)}>
                <Icon name="shuffle" size={15} /> Shuffle library
              </button>
              <div class="mi mi-row">
                <Icon name="hand" size={15} />
                <span>Mulligan to</span>
                <input
                  type="number"
                  min="0"
                  max="20"
                  bind:value={mulliganTo}
                  aria-label="mulligan hand size"
                />
                <button class="sm" onclick={() => viaMenu(mulligan)}>Go</button>
              </div>
              <button
                class="mi"
                role="menuitem"
                onclick={() => viaMenu(() => client.sendAction("undo"))}
                disabled={!isAdmin && (viewerSeat?.undos_remaining ?? 0) <= 0}
                title={isAdmin
                  ? "rewind the most recent action (admin — bypasses caller / budget gates)"
                  : (viewerSeat?.undos_remaining ?? 0) <= 0
                    ? "no undos remaining this turn (refreshes on your next untap)"
                    : `undo your most recent action — ${viewerSeat?.undos_remaining ?? 0} left this turn`}
              >
                <Icon name="undo" size={15} /> Undo
                {#if !isAdmin && viewerSeat}
                  <span class="mi-r">{viewerSeat.undos_remaining ?? 0} left</span>
                {/if}
              </button>
              <label
                class="mi mi-row"
                title="per-player undo budget refreshed each turn (any seat may change)"
              >
                <span class="mi-indent">Undo limit</span>
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
                class="mi"
                role="menuitem"
                onclick={() => viaMenu(() => (showLifeHistory = true))}
              >
                <Icon name="drop" size={15} /> Life history
              </button>
              <div class="sep"></div>
              <div class="menu-h">Table</div>
              {#if bugReportAvailable}
                <button
                  class="mi"
                  role="menuitem"
                  onclick={() => viaMenu(() => (bugReportOpen = true))}
                >
                  <Icon name="bug" size={15} /> Report a bug
                </button>
              {/if}
              <button class="mi" role="menuitem" onclick={() => viaMenu(back)}>
                <Icon name="chevronLeft" size={15} /> Back to lobby
              </button>
              <div class="sep"></div>
              <button
                class="mi danger"
                role="menuitem"
                onclick={requestConcede}
                disabled={viewerEliminated || gameEnded}
                title={viewerEliminated
                  ? "you are already eliminated"
                  : gameEnded
                    ? "the game has ended"
                    : "concede the game (irreversible)"}
              >
                <Icon name="flag" size={15} /> Concede…
              </button>
            </div>
          {/if}
        </div>
      {/if}
    </div>
    {#if concedeConfirm}
      <div class="confirm" role="dialog" aria-modal="true" aria-label="concede the game?">
        <div class="confirm-title">Concede the game?</div>
        <p class="confirm-body">
          You'll be eliminated and keep watching as a spectator. This can't be undone.
        </p>
        <div class="confirm-actions">
          <button onclick={() => (concedeConfirm = false)}>Keep playing</button>
          <button class="danger" onclick={confirmConcede}
            ><Icon name="flag" size={14} /> Concede</button
          >
        </div>
      </div>
    {/if}
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
            <Icon name="x" size={14} />
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
  </header>

  {#if viewerNeedsDeck && !deckImportDismissed && viewerID}
    <div
      class="prompt-backdrop deck-import-modal-backdrop"
      role="dialog"
      aria-modal="true"
      aria-labelledby="deck-import-title"
    >
      <div class="prompt-modal deck-import-modal">
        <h2 id="deck-import-title">
          Import your deck
          <span class="prompt-src" aria-hidden="true">before the game starts</span>
        </h2>
        <p class="prompt-hint">
          Paste a Moxfield or Archidekt deck URL, a Moxfield JSON export, or a plain-text decklist.
          The server validates against Commander rules (100-card singleton, color identity, format
          legality).
        </p>
        <DeckUploadForm
          {gameID}
          playerID={viewerID}
          onSuccess={() => (deckImportDismissed = true)}
        />
        <p class="prompt-hint import-hint">
          You can also manage decks (and other seats) from the
          <button class="linkish" onclick={back}>lobby</button>.
        </p>
      </div>
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
        {autopassEnabled}
        onPassPriority={passPriority}
        onToggleAutopass={toggleAutopass}
      >
        <!-- Everything that asks for the viewer's attention shares the
             board's strip (under the stack card): targeting prompt,
             combat hint, opening-hand roll-call, toasts, game end.
             Nothing here pushes the table around. -->
        {#snippet attention()}
          <TargetingBanner />

          <!-- #318: the attack-with-all cluster. Present for the whole
               declare-attackers step so the count stays live as
               creatures are declared one by one; it disappears the
               moment nothing is left that could attack. -->
          {#if canDeclareAttackers && !mulligansOpen && !gameEnded && (attackAllReady || canUndoDeclaration)}
            <div class="att attack-all" aria-label="declare attackers">
              <span class="att-label danger">
                <Icon name="sword" size={12} />
                attack
              </span>
              <span class="att-text">
                {#if attackAllReady}
                  <strong>{attackPlan.eligible.length}</strong>
                  ready to attack
                  {#if attackPlan.declared.length > 0}
                    <span class="muted">· {attackPlan.declared.length} already declared</span>
                  {/if}
                  {#if attackBlockedHint}
                    <span class="muted">· can't: {attackBlockedHint}</span>
                  {/if}
                {:else}
                  <strong>{attackPlan.declared.length}</strong>
                  declared
                  {#if attackBlockedHint}
                    <span class="muted">· {attackBlockedHint} can't attack</span>
                  {/if}
                {/if}
              </span>
              {#if attackAllReady}
                {#if attackPlan.defenders.length === 1}
                  <button
                    type="button"
                    class="primary att-btn"
                    title={attackAllLabel(attackPlan, attackPlan.defenders[0])}
                    onclick={() => attackAllAt(attackPlan.defenders[0].id)}
                  >
                    {attackAllLabel(attackPlan, attackPlan.defenders[0])}
                  </button>
                {:else}
                  <!-- Multi-opponent: one button per seat rather than a
                       bare "attack all", so the control always says who
                       gets hit. Nothing here spreads an attack. -->
                  <span class="att-text all-at">Attack all →</span>
                  {#each attackPlan.defenders as opp (opp.id)}
                    <button
                      type="button"
                      class="att-btn opp-btn"
                      title={attackAllLabel(attackPlan, opp)}
                      onclick={() => attackAllAt(opp.id)}
                    >
                      <span class="seat-dot" style="background:{seatColor(opp.seat)}"></span>
                      {seatLabel(opp)}
                    </button>
                  {/each}
                {/if}
              {/if}
              {#if canUndoDeclaration}
                <button
                  type="button"
                  class="ghost att-btn"
                  title="take back your last declaration — restores tap state too"
                  onclick={undoDeclaration}
                >
                  <Icon name="undo" size={12} /> Undo
                </button>
              {/if}
            </div>
          {/if}

          {#if combatSelection && !mulligansOpen}
            <div class="att combat-hint" role="status" aria-live="polite">
              <span class="att-label danger">
                <Icon name="sword" size={12} />
                {combatSelection.kind === "attacker" ? "attack" : "block"}
              </span>
              <span class="att-text">
                {#if combatSelection.kind === "attacker"}
                  Attacking with <strong>{cardLabel(combatSelection.cardID).name}</strong> — click an
                  opponent's seat to commit, or the creature again to cancel.
                {:else}
                  Blocking with <strong>{cardLabel(combatSelection.cardID).name}</strong> — click an incoming
                  attacker to commit, or the creature again to cancel.
                {/if}
              </span>
              <button type="button" class="ghost att-btn" onclick={() => (combatSelection = null)}>
                Cancel <kbd>Esc</kbd>
              </button>
            </div>
          {/if}

          {#if mulligansOpen && !gameEnded}
            <div class="att mulligan-banner" aria-label="opening hand decisions">
              <span class="att-label">Opening hands</span>
              {#each seats as seat (seat.id)}
                <span
                  class="mull"
                  class:waiting={!seat.hand_kept && !seat.eliminated}
                  class:kept={seat.hand_kept && !seat.eliminated}
                  style="--seat-color: {seatColor(seat.seat)}"
                >
                  <span class="seat-dot" style="background:{seatColor(seat.seat)}"></span>
                  <b>{seat.name}</b>
                  {#if seat.eliminated}
                    <span class="muted">eliminated</span>
                  {:else if seat.hand_kept}
                    kept <Icon name="check" size={11} />
                  {:else}
                    deciding…
                  {/if}
                  {#if (seat.mulligans_taken ?? 0) > 0}
                    <span class="muted mull-count">×{seat.mulligans_taken}</span>
                  {/if}
                </span>
              {/each}
            </div>
          {/if}

          {#if manaOverride}
            <div class="att toast mana-override" role="alert" aria-live="polite">
              <span class="att-label gold">mana</span>
              <span class="att-text">
                <strong>Insufficient mana</strong>
                {#if manaOverride.missing.length > 0}
                  <span class="muted">· missing {manaOverride.missing.join(" ")}</span>
                {/if}
              </span>
              <button type="button" class="primary att-btn" onclick={openAutoTap}>
                Auto-tap & cast
              </button>
              <button type="button" class="att-btn" onclick={castAnyway}>Cast anyway</button>
              <button
                type="button"
                class="ghost att-close"
                onclick={dismissManaOverride}
                aria-label="dismiss"
              >
                <Icon name="x" size={12} />
              </button>
            </div>
          {:else if $lastError}
            <div class="att toast error" role="alert" aria-live="polite">
              <span class="att-label danger">rejected</span>
              <span class="att-text">
                {$lastError.message}
                <span class="muted mono">({$lastError.code})</span>
              </span>
              <button
                type="button"
                class="ghost att-close"
                onclick={() => lastError.set(null)}
                aria-label="dismiss"
              >
                <Icon name="x" size={12} />
              </button>
            </div>
          {/if}

          {#if gameEnded}
            <div class="att game-end" role="alert">
              <span class="att-label gold"><Icon name="crown" size={12} /> game over</span>
              <span class="att-text">
                {#if winner}
                  <span class="seat-dot" style="background:{seatColor(winner.seat)}"></span>
                  <strong>{winner.name}</strong> wins the game.
                {:else}
                  Game ended — no survivors.
                {/if}
              </span>
              <button type="button" class="primary att-btn" onclick={back}> Back to lobby </button>
            </div>
          {:else if viewerEliminated}
            <div class="att eliminated" role="status">
              <span class="att-label danger">eliminated</span>
              <span class="att-text">You have been eliminated. Spectating.</span>
            </div>
          {/if}
        {/snippet}
      </Board>
      {#if viewerNeedsToDecide}
        <div class="mulligan-scrim"></div>
        <div class="mulligan-dialog" role="dialog" aria-label="keep or mulligan your hand">
          <header>
            <h2>Your opening hand</h2>
            {#if (viewerSeat?.mulligans_taken ?? 0) > 0}
              <p class="muted">
                Mulligans taken: {viewerSeat?.mulligans_taken}. You'll redraw 7 cards (simplified
                London — no bottom-N penalty yet).
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
                      src={`/cards/${card.scryfall_id}/image?size=normal`}
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
            <button onclick={mulliganDecide}>Mulligan</button>
            <button class="primary" onclick={keepHand}>Keep hand</button>
          </div>
        </div>
      {/if}
      <DiscardPromptModal snap={view} {viewerID} {sendAction} />
      <ChoicePromptModal snap={view} {viewerID} {sendAction} />
      <AutoTapPreviewModal
        {gameID}
        snap={view}
        cardID={autoTapCardID}
        onConfirm={confirmAutoTap}
        onCancel={cancelAutoTap}
      />
    {/if}
  </div>

  {#if !view}
    <p class="muted centered">waiting for snapshot…</p>
  {/if}

  {#if bugReportOpen}
    <BugReportModal
      {gameID}
      {view}
      seq={$lastSeq}
      connection={$status}
      wsLog={$log}
      attachments={bugReportAttachments}
      onclose={() => (bugReportOpen = false)}
    />
  {/if}
</section>

<svelte:window
  onkeydown={(ev: KeyboardEvent) => {
    if (ev.key === "Escape") {
      cancelTargeting();
      menuOpen = false;
      concedeConfirm = false;
    }
    // S20 sub-PR 5: Enter confirms a multi-target pick list (no-op
    // for single-target prompts and when fewer than min are picked).
    if (ev.key === "Enter" && !(ev.target instanceof HTMLInputElement)) confirmTargeting();
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
    padding: 0.45rem 0.6rem 0.6rem;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    overflow: hidden;
  }
  /* Command bar: lobby · wordmark · game · status · pass turn · icons.
     Sandbox utilities and table actions live behind the ⋯ menu. */
  .bar {
    display: flex;
    align-items: center;
    gap: 12px;
    height: 44px;
    margin: -0.45rem -0.6rem 0;
    padding: 0 14px;
    background: var(--bg-1);
    border-bottom: 1px solid var(--border);
    position: relative;
    z-index: 40;
  }
  .bar-nav {
    margin-left: -6px;
  }
  .bar-sep {
    width: 1px;
    height: 18px;
    background: var(--border-strong);
  }
  .wordmark {
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 14px;
    letter-spacing: 0.18em;
    color: var(--fg);
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }
  .wordmark i {
    display: inline-block;
    width: 14px;
    height: 14px;
    border: 2px solid var(--gold);
    transform: rotate(45deg);
    border-radius: 3px;
    box-sizing: border-box;
  }
  .bar :global(h1.crumb) {
    margin: 0;
    font-family: var(--font-ui);
    font-size: 13px;
    font-weight: 500;
    letter-spacing: 0;
    color: var(--fg-muted);
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }
  .crumb b {
    color: var(--fg-dim);
    font-weight: 400;
  }
  .mono {
    font-family: var(--font-mono);
  }
  .bar-spacer {
    flex: 1;
  }
  .status {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--fg-muted);
  }
  .status .seq {
    color: var(--fg-dim);
  }
  .status .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--fg-dim);
  }
  .status-connected .dot {
    background: var(--mint);
    box-shadow: 0 0 8px var(--mint);
  }
  .status-connecting .dot,
  .status-reconnecting .dot {
    background: var(--gold);
    box-shadow: 0 0 8px var(--gold);
  }
  .status-reconnecting .dot {
    animation: status-pulse 1.2s var(--ease) infinite;
  }
  @keyframes status-pulse {
    50% {
      opacity: 0.45;
    }
  }
  .status-disconnected .dot {
    background: var(--danger);
    box-shadow: 0 0 8px var(--danger);
  }
  .tag-spectator {
    background: rgba(176, 138, 255, 0.15);
    color: var(--magenta);
    border: 1px solid rgba(176, 138, 255, 0.5);
    text-transform: uppercase;
    letter-spacing: 0.1em;
    font-size: 0.7em;
    font-weight: 700;
    padding: 0.15rem 0.5rem;
    border-radius: 999px;
    font-family: var(--font-mono);
  }
  .bar-btn {
    height: 30px;
    padding: 0 12px;
    font-size: 12.5px;
  }
  .bar-icons {
    display: flex;
    gap: 2px;
    align-items: center;
  }
  .ibtn {
    width: 32px;
    height: 32px;
    padding: 0;
    border-radius: 8px;
    border: 1px solid transparent;
    background: transparent;
    color: var(--fg-muted);
  }
  .ibtn:hover,
  .ibtn.on {
    color: var(--fg);
    background: rgba(255, 255, 255, 0.06);
    border-color: var(--border);
  }
  .more {
    position: relative;
  }
  .menu-backdrop {
    position: fixed;
    inset: 0;
    z-index: 50;
  }
  .menu {
    position: absolute;
    right: 0;
    top: calc(100% + 6px);
    width: 284px;
    z-index: 60;
    background: var(--surface);
    border: 1px solid var(--border-strong);
    border-radius: 12px;
    box-shadow: var(--shadow-lg);
    padding: 6px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .menu-h {
    font-family: var(--font-mono);
    font-size: 9.5px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
    padding: 8px 10px 4px;
  }
  .mi {
    display: flex;
    align-items: center;
    gap: 10px;
    height: 32px;
    padding: 0 10px;
    border-radius: 8px;
    border: 1px solid transparent;
    background: transparent;
    color: var(--fg);
    font-size: 12.5px;
    font-weight: 500;
    width: 100%;
    justify-content: flex-start;
    text-align: left;
    box-shadow: none;
    box-sizing: border-box;
  }
  .mi:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.06);
    border-color: transparent;
  }
  .mi:disabled {
    opacity: 0.45;
  }
  .mi.danger {
    color: var(--danger);
  }
  .mi-r {
    margin-left: auto;
    font-family: var(--font-mono);
    font-size: 10.5px;
    color: var(--fg-muted);
    background: var(--surface-raised);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 2px 6px;
  }
  .mi-row {
    cursor: default;
  }
  .mi-row input {
    width: 44px;
    padding: 3px 6px;
    margin: 0 0 0 auto;
    font-size: 12px;
    font-family: var(--font-mono);
    border-radius: 6px;
  }
  .mi-row .sm {
    height: 24px;
    padding: 0 8px;
    font-size: 11.5px;
    border-radius: 6px;
  }
  .mi-indent {
    margin-left: 25px;
    color: var(--fg-muted);
  }
  .sep {
    height: 1px;
    background: var(--border);
    margin: 4px 6px;
  }
  .confirm {
    position: absolute;
    right: 60px;
    top: calc(100% + 6px);
    width: 300px;
    z-index: 60;
    background: var(--surface);
    border: 1px solid var(--border-strong);
    border-radius: 12px;
    box-shadow: var(--shadow-lg);
    padding: 14px 14px 12px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .confirm-title {
    font-family: var(--font-display);
    font-size: 15px;
    font-weight: 700;
    color: var(--fg);
  }
  .confirm-body {
    margin: 0;
    font-size: 12.5px;
    color: var(--fg-muted);
    line-height: 1.45;
  }
  .confirm-actions {
    display: flex;
    justify-content: flex-end;
    gap: 6px;
    margin-top: 4px;
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

  /* ---- Attention strip rows (rendered inside Board's .strip) ----
     One flat card per live prompt: mono label on the left, text in
     the middle, actions on the right. Same shell as StackOverlay and
     TargetingBanner so the column reads as one instrument. */
  .att {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 10px 9px 14px;
    background: color-mix(in srgb, var(--surface) 94%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.35;
    box-sizing: border-box;
    animation: att-in 200ms var(--ease);
  }
  @keyframes att-in {
    from {
      opacity: 0;
      transform: translateY(-6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  .att-label {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    font-weight: 700;
    color: var(--fg-dim);
    display: inline-flex;
    align-items: center;
    gap: 6px;
    flex: 0 0 auto;
    white-space: nowrap;
  }
  .att-label.gold {
    color: var(--gold-strong);
  }
  .att-label.danger {
    color: var(--danger);
  }
  .att-text {
    flex: 1;
    min-width: 0;
  }
  .att-text strong {
    color: var(--fg);
    font-weight: 700;
  }
  .att-text .mono {
    font-family: var(--font-mono);
    font-size: 11px;
  }
  .att-btn {
    flex: 0 0 auto;
    height: 28px;
    padding: 0 10px;
    font-size: 11.5px;
    border-radius: 7px;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .att-btn kbd {
    font-family: var(--font-mono);
    font-size: 9.5px;
    color: var(--fg-dim);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 0 4px;
    line-height: 16px;
  }
  .att-close {
    flex: 0 0 auto;
    width: 26px;
    height: 26px;
    padding: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 7px;
  }
  .att .seat-dot {
    display: inline-block;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    vertical-align: middle;
    margin-right: 4px;
  }
  .combat-hint,
  .mana-override,
  .game-end {
    border-color: rgba(217, 180, 92, 0.45);
  }

  /* #318 attack-with-all cluster. Wraps rather than overflowing —
     a four-player table puts three opponent buttons in the strip and
     the strip is only ~512px wide. */
  .attack-all {
    flex-wrap: wrap;
    border-color: rgba(255, 122, 122, 0.4);
  }
  .attack-all .all-at {
    flex: 0 0 auto;
    color: var(--fg-dim);
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
  }
  .attack-all .opp-btn {
    border-color: rgba(255, 122, 122, 0.35);
  }
  .toast.error,
  .eliminated {
    border-color: rgba(255, 107, 107, 0.4);
  }

  /* Opening-hand roll-call: one pill per seat. */
  .mulligan-banner {
    flex-wrap: wrap;
    gap: 8px;
  }
  .mull {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-size: 12px;
    color: var(--fg-muted);
    padding: 3px 8px;
    border-radius: 999px;
    border: 1px solid var(--border);
  }
  .mull b {
    color: var(--fg);
    font-weight: 600;
  }
  .mull.kept {
    color: var(--mint);
  }
  .mull.waiting {
    border-color: rgba(217, 180, 92, 0.5);
    color: var(--gold-strong);
  }
  .mull-count {
    font-family: var(--font-mono);
    font-size: 10px;
  }

  /* The viewer's own keep-or-mulligan decision: the table dims and
     the dialog docks low with large cards, so the hand is the only
     thing in focus. */
  .mulligan-scrim {
    /* Sits under Board's attention strip (z 40) so the opening-hand
       roll-call stays readable while the table behind it dims. */
    position: absolute;
    inset: 0;
    z-index: 38;
    background: rgba(11, 10, 9, 0.6);
    backdrop-filter: blur(2px);
    -webkit-backdrop-filter: blur(2px);
  }
  .mulligan-dialog {
    position: absolute;
    left: 50%;
    bottom: 24px;
    transform: translateX(-50%);
    z-index: 71;
    width: min(880px, calc(100% - 48px));
    padding: 16px 18px 14px;
    background: var(--surface);
    border: 1px solid rgba(217, 180, 92, 0.4);
    border-radius: var(--radius-xl);
    color: var(--fg);
    box-shadow: var(--shadow-lg);
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .mulligan-dialog header {
    display: flex;
    align-items: baseline;
    gap: 12px;
    flex-wrap: wrap;
  }
  .mulligan-dialog header h2 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 18px;
    font-weight: 700;
  }
  .mulligan-dialog header p {
    margin: 0;
    font-size: 12.5px;
  }
  .mulligan-cards {
    display: flex;
    gap: 8px;
    overflow-x: auto;
    padding: 2px 0;
  }
  .mulligan-card {
    flex: 0 0 auto;
    width: 112px;
    height: 157px;
    border-radius: 7px;
    overflow: hidden;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    box-shadow: var(--shadow-sm);
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
    padding: 6px;
    text-align: center;
    font-size: 11px;
    color: var(--fg-muted);
    box-sizing: border-box;
  }
  .mulligan-actions {
    display: flex;
    justify-content: flex-end;
    gap: 6px;
  }

  .life-history-popover {
    position: absolute;
    top: calc(100% + 6px);
    right: 14px;
    z-index: 60;
    width: min(320px, calc(100vw - 2rem));
    max-height: 360px;
    background: var(--surface);
    border: 1px solid var(--border-strong);
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
    color: var(--fg-muted);
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
    color: var(--fg);
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
    color: var(--mint);
  }
  .life-delta.loss {
    color: var(--danger);
  }
  .life-newtotal {
    color: var(--fg-muted);
    font-variant-numeric: tabular-nums;
  }
  .life-time {
    font-size: 0.85em;
  }

  /* Pre-game deck import modal (S08.5 wave 1) — the shared prompt
     shell, sitting under the settings / bug-report modals. */
  .deck-import-modal-backdrop {
    z-index: 50;
    padding: 1rem;
  }
  .deck-import-modal {
    width: min(640px, 100%);
  }
  .import-hint {
    margin-top: 4px;
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
