<script lang="ts">
  // ChoicePromptModal opens when the wire's pending_choices queue
  // has an entry addressed to the viewer. Generic counterpart to
  // DiscardPromptModal: that one handles the S13.4 cleanup-specific
  // map where chooser == owner; this one handles the S14+ queue
  // where chooser can differ from the pool owner (Thoughtseize:
  // caster picks from target's revealed hand).
  //
  // The options[] slice on each entry is pre-filtered by the
  // server's per-viewer KnownBy projection — cards the viewer
  // is legally allowed to see arrive face-up; the rest arrive
  // redacted (backs). So this modal just renders options[] as-is.

  import type {
    ActionType,
    CardView,
    DamageAssignmentView,
    GameView,
    PendingChoiceView,
    PickOptionView,
    ReplacementOptionView,
  } from "../../protocol";
  import Card from "./Card.svelte";
  import ModalLayer from "../ModalLayer.svelte";
  import {
    rejectionForPrompt,
    type ChoiceRejection,
    type ChoiceSubmission,
    type ServerErrorLike,
  } from "../../choiceRejection";
  import { doubledTriggerLabel } from "../../triggerDoubling";
  import { colorButtons, colorPromptAnswerable, colorPromptCopy } from "../../manaPick";

  interface Props {
    snap: GameView;
    viewerID: string | null;
    sendAction: (type: ActionType, params?: unknown, player?: string) => void;
    // GameClient.lastError. The board's toast for it sits under this
    // modal's backdrop, so a refusal of this prompt's answer is shown
    // here instead (#624). Optional so a caller with no error feed
    // still mounts the modal.
    lastError?: ServerErrorLike | null;
  }

  const { snap, viewerID, sendAction, lastError = null }: Props = $props();

  // First choice addressed to the viewer. Queue ordering: front of
  // list is "what the chooser sees next." One modal at a time; when
  // they resolve this one, the next pops automatically on the
  // snapshot after the server drains the entry.
  const active = $derived.by((): PendingChoiceView | null => {
    if (!viewerID || !snap.pending_choices) return null;
    for (const c of snap.pending_choices) {
      // S20 sub-PR 2: pick_target is answered by clicking the board
      // (Board.svelte drives the targeting store), not by a modal.
      // #1196: the CR 115.7 retarget prompt is answered the same way
      // — click the new target on the board — so it is not a modal
      // either.
      if (c.kind === "pick_target" || c.kind === "retarget") continue;
      // #844, CR 903.4f: a colour prompt with no colours on offer is
      // not a choice anybody can answer, and an empty picker modal
      // would block the board. The server stopped queueing one when a
      // "commander's color identity" source has no identity; this is
      // the floor under that.
      if (!colorPromptAnswerable(c)) continue;
      if (c.chooser === viewerID) return c;
    }
    return null;
  });

  const open = $derived(active !== null);

  // Source player (whose hand the picks come from). Used for the
  // modal header copy.
  const fromName = $derived.by(() => {
    if (!active) return "";
    const seat = snap.seats.find((s) => s.id === active.from_player);
    return seat?.name ?? "opponent";
  });

  const isSelfSource = $derived(active && viewerID && active.from_player === viewerID);

  let selected = $state<Set<string>>(new Set());
  // S17 replacement_order: array of effect IDs in the order the
  // chooser has picked. Click a row to append; click again to
  // remove (and subsequent positions compact down). Submit when
  // the array covers every candidate.
  let ordered = $state<string[]>([]);

  // #624: the last answer this modal sent, and the server's refusal of
  // it if one came back. See the effect below the reset.
  let submission: ChoiceSubmission | null = null;
  let rejection = $state<ChoiceRejection | null>(null);

  // Reset selection whenever the modal opens fresh (active changes
  // from null → non-null, or the choice ID changes).
  let lastChoiceID: string | null = null;
  $effect(() => {
    const nextID = active?.id ?? null;
    if (nextID !== lastChoiceID) {
      // #826: an untap prompt whose ceiling is "all of them" is a board
      // of nothing but "you may choose NOT to untap" permanents, and
      // the default there is to untap — so the player's click should be
      // the deselection. Where a cap binds there is no such default and
      // the picker starts empty rather than pre-filled with a set the
      // player would have to undo.
      const options = active?.options ?? [];
      const preselect =
        active?.kind === "untap_choice" &&
        options.length > 0 &&
        (active.choose_max ?? 0) >= options.length;
      selected = preselect ? new Set(options.map((c) => c.instance_id)) : new Set();
      ordered = [];
      rejection = null;
      submission = null;
      lastChoiceID = nextID;
    }
  });

  // #624: a refused answer leaves the prompt open for another try, and
  // the reason has to be readable while it is open. `submission` is a
  // plain variable, like lastChoiceID, so sending an answer does not
  // re-run the effect below; a new error frame or a new prompt does.
  //
  // The shown rejection is latched rather than derived: lastError
  // clears itself after a few seconds, and a player reading the card's
  // text again should not lose the reason halfway. It clears when they
  // send another answer or the prompt changes.
  $effect(() => {
    const hit = rejectionForPrompt(submission, active?.id ?? null, lastError);
    if (hit) rejection = hit;
  });

  // answer sends a resolve_choice for the open prompt. Every kind's
  // submit goes through here so a refusal of any of them is shown.
  // A new mode_pick prompt must not arrive holding the last one's
  // picks — two Gala Greeters triggers in a turn are two questions.
  let lastModeID: string | null = null;
  $effect(() => {
    const id = isModePick ? (active?.id ?? null) : null;
    if (id !== lastModeID) {
      lastModeID = id;
      modePicks = [];
    }
  });

  function answer(params: Record<string, unknown>): void {
    if (!active || !viewerID) return;
    submission = { choiceID: active.id, sentAt: Date.now() };
    rejection = null;
    sendAction("resolve_choice", { choice_id: active.id, ...params }, viewerID);
  }

  // S22 search_library — "search your library for ..." (CR 701.23).
  // Shares the card grid and the {choice_id, card_ids} payload with
  // discard / sacrifice; what differs is the floor. Every other
  // card-grid kind demands exactly `count` picks, but a search may
  // always find FEWER than it looked for, including none at all
  // (CR 701.23b "you may fail to find"). So search_max is a ceiling
  // and zero is a legal answer.
  const isSearch = $derived(active?.kind === "search_library");

  // S16.5 copy_target — "you may have this creature enter as a copy
  // of ...". Shares the card grid and the {choice_id, card_ids}
  // payload; like search, its floor is zero, because every printed
  // card in the class says "you may" and declining is a real answer
  // (the permanent enters as its own printed self instead).
  //
  // Nothing on the board changes while this is open: the permanent
  // is still on the stack, and the answer decides what it enters AS.
  const isCopyTarget = $derived(active?.kind === "copy_target");

  // #74 choose_cards — the chained-choice card-set pick: "choose N of
  // these cards", with the card deciding what being chosen means
  // (Sylvan Library: the two you then pay 4 life each to keep).
  // Shares the card grid and the {choice_id, card_ids} payload; what
  // it does NOT share is a fixed count, because its floor and ceiling
  // are sent separately and can differ.
  const isChooseCards = $derived(active?.kind === "choose_cards");

  // #826 untap_choice — CR 502.3's "the active player determines which
  // permanents they control will untap", asked when a cap ("players
  // can't untap more than one land") or an opt-out ("you may choose
  // not to untap this") makes it a real decision. Same payload, same
  // bounds and the same picker as choose_cards; what differs is the
  // sentence, and that the candidates are public permanents rather
  // than somebody's hand.
  const isUntapChoice = $derived(active?.kind === "untap_choice");

  // #1198 entry_reveal_from_hand — CR 614.1c's "as this land enters,
  // you may reveal an Island or Swamp card from your hand. If you
  // don't, it enters tapped." Same payload, same bounds and the same
  // picker; what differs is that the land is NOT on the battlefield
  // yet (the answer decides how it enters) and that revealing costs
  // nothing — the card stays in hand, so the floor of zero is a real
  // bluff rather than a formality.
  const isEntryReveal = $derived(active?.kind === "entry_reveal_from_hand");

  // #1214 — the three resolution-time picks (CR 608.2). Same
  // {choice_id, card_ids} payload and the same choose_min / choose_max
  // bounds as choose_cards, so they render through the same grid;
  // what differs is the sentence, because the three questions are not
  // the same question.
  //
  //   reveal_pick       an opponent picks from a set you revealed —
  //                     the cards are not theirs, and the reveal is
  //                     what lets them look at all.
  //   their_permanents  a pick over somebody ELSE's battlefield.
  //                     `from_player` names whose.
  //   own_permanents    "choose N of your own permanents", untargeted
  //                     and made on resolution.
  const isRevealPick = $derived(active?.kind === "reveal_pick");
  const isTheirPermanents = $derived(active?.kind === "their_permanents");
  const isOwnPermanents = $derived(active?.kind === "own_permanents");
  const isPermanentPick = $derived(isTheirPermanents || isOwnPermanents);
  // The kinds that share the bounded card-set grid.
  const isCardSetPick = $derived(
    isChooseCards || isUntapChoice || isEntryReveal || isRevealPick || isPermanentPick,
  );

  // How many cards this prompt accepts, and how few it will settle
  // for. Search and copy are the two that move the floor off the
  // ceiling.
  const pickMax = $derived(
    isSearch
      ? (active?.search_max ?? 1)
      : isCardSetPick
        ? (active?.choose_max ?? active?.count ?? 0)
        : (active?.count ?? 0),
  );
  const pickMin = $derived(
    isSearch || isCopyTarget ? 0 : isCardSetPick ? (active?.choose_min ?? 0) : (active?.count ?? 0),
  );
  const canSubmit = $derived(selected.size >= pickMin && selected.size <= pickMax);

  function toggle(id: string): void {
    if (!active) return;
    const next = new Set(selected);
    if (next.has(id)) next.delete(id);
    else if (next.size < pickMax) next.add(id);
    selected = next;
  }

  function submit(): void {
    if (!active || !viewerID) return;
    if (!canSubmit) return;
    answer({ card_ids: Array.from(selected) });
  }

  // Options come redacted for non-knowers; filter to cards the
  // viewer can identify so we don't render a grid of anonymous
  // backs the viewer can't meaningfully pick from. Unknown
  // options here would mean the server misrouted — the chooser
  // should always be a knower of the revealed cards.
  const optionCards = $derived<CardView[]>(active?.options ?? []);

  // S15 mana_pick branch — a color-pick choice from Arcane Signet /
  // Birds of Paradise. `active.color_options` is the server's legal
  // button list, already ordered commander identity first; it renders
  // in that order. Submits via resolve_choice with `{choice_id,
  // color}` (card_ids absent).
  const isManaPick = $derived(active?.kind === "mana_pick");
  // #742: "N mana of any one color" (Gilded Lotus) is one pick that
  // adds several tokens of the picked colour; the amount can differ
  // per colour (Nyx Lotus's devotion). A colour missing from the map
  // adds one. colorButtons keeps the server's order.
  const colorAmounts = $derived<Record<string, number>>(active?.color_amounts ?? {});
  const hasColorAmounts = $derived(Object.keys(colorAmounts).length > 0);
  const buttons = $derived(colorButtons(active?.color_options, colorAmounts));

  // #742 choose_color branch — "choose a color" (CR 105.4), either as
  // a permanent enters (Coldsteel Heart; the answer is remembered) or
  // while a spell resolves (Wash Out). Same buttons and the same
  // `{choice_id, color}` answer as a mana pick; the server routes the
  // two by kind.
  const isColorChoice = $derived(active?.kind === "choose_color");
  // #986: the card declares what it will DO with the colour
  // (`color_purpose`), and the picker says it back. Five identical
  // buttons cannot tell a player whether they are naming the colour
  // their Coldsteel Heart will produce or the colour their Wash Out is
  // about to bounce, and those are opposite answers. The button ORDER
  // is the server's — it ranks the options by the same purpose — so
  // nothing here sorts.
  const colorCopy = $derived(colorPromptCopy(active?.color_purpose));

  function pickColor(color: string): void {
    if (!active || !viewerID) return;
    answer({ color });
  }

  // S26 choose_creature_type branch — "as this permanent enters,
  // choose a creature type" (CR 614.12). The legal set is the whole
  // CR 205.3m vocabulary, so this is a filter box over a scrolling
  // list rather than a button row like the colours: nobody scans 345
  // buttons, and everybody already knows the word they want.
  const isCreatureTypePick = $derived(active?.kind === "choose_creature_type");
  const typeOptions = $derived<string[]>(active?.type_options ?? []);
  let typeFilter = $state("");
  const filteredTypes = $derived.by(() => {
    const q = typeFilter.trim().toLowerCase();
    if (!q) return typeOptions;
    // Prefix matches first — typing "el" should offer Elf before
    // Shapeshifter, even though both contain the letters.
    const starts = typeOptions.filter((t) => t.toLowerCase().startsWith(q));
    const contains = typeOptions.filter(
      (t) => !t.toLowerCase().startsWith(q) && t.toLowerCase().includes(q),
    );
    return [...starts, ...contains];
  });

  function pickCreatureType(creatureType: string): void {
    if (!active || !viewerID) return;
    typeFilter = "";
    answer({ creature_type: creatureType });
  }

  // Enter submits the single best match, so a player who knows their
  // deck can type "sliv" and hit return without reaching for the
  // mouse. Deliberately requires an unambiguous top hit rather than
  // guessing among several: naming the wrong tribe is unrecoverable,
  // since the choice is made once as the permanent enters.
  function onTypeFilterKey(e: KeyboardEvent): void {
    if (e.key !== "Enter" || filteredTypes.length === 0) return;
    e.preventDefault();
    pickCreatureType(filteredTypes[0]);
  }

  // #1210 choose_card_name branch — "as this permanent enters, choose
  // a card name" (CR 614.12): Pithing Needle, Phyrexian Revoker,
  // Sorcerous Spyglass.
  //
  // The one prompt in the engine with NO legal set. CR 201.2 lets a
  // player name any card name at all, so `name_options` is a
  // SUGGESTION list — the names visible in public zones — and the
  // text box is the real answer. That is why this is not a second
  // copy of the creature-type picker despite looking like one: there,
  // typing filters a closed vocabulary and Enter takes the top match;
  // here, Enter submits WHAT WAS TYPED, because a name the list does
  // not have is an ordinary answer and guessing over the player would
  // be the bug.
  const isCardNamePick = $derived(active?.kind === "choose_card_name");
  const nameOptions = $derived<string[]>(active?.name_options ?? []);
  let nameFilter = $state("");
  const filteredNames = $derived.by(() => {
    const q = nameFilter.trim().toLowerCase();
    if (!q) return nameOptions;
    const starts = nameOptions.filter((n) => n.toLowerCase().startsWith(q));
    const contains = nameOptions.filter(
      (n) => !n.toLowerCase().startsWith(q) && n.toLowerCase().includes(q),
    );
    return [...starts, ...contains];
  });

  function pickCardName(name: string): void {
    const trimmed = name.trim();
    if (!active || !viewerID || !trimmed) return;
    nameFilter = "";
    answer({ card_name: trimmed });
  }

  function onNameFilterKey(e: KeyboardEvent): void {
    if (e.key !== "Enter") return;
    e.preventDefault();
    pickCardName(nameFilter);
  }

  // S17 replacement_order branch — CR 616 affected-player-chooses-
  // order prompt. Click a row to append it to the `ordered` array;
  // click a row already in the array to remove it (later rows
  // compact down). Submit with { choice_id, order: [...] }.
  const isReplacementOrder = $derived(active?.kind === "replacement_order");
  // S19 sub-PR 8 trigger_order — CR 603.3b: the same reorder list,
  // fed from trigger_options. Submitted order = resolution order
  // (top of the list resolves first); the server stacks in reverse.
  const isTriggerOrder = $derived(active?.kind === "trigger_order");
  const replacementOptions = $derived<ReplacementOptionView[]>(
    active?.kind === "trigger_order"
      ? (active?.trigger_options ?? [])
      : (active?.replacement_options ?? []),
  );

  function toggleReplacement(id: string): void {
    const idx = ordered.indexOf(id);
    if (idx >= 0) {
      ordered = [...ordered.slice(0, idx), ...ordered.slice(idx + 1)];
    } else {
      ordered = [...ordered, id];
    }
  }

  function submitReplacementOrder(): void {
    if (!active || !viewerID) return;
    if (ordered.length !== replacementOptions.length) return;
    answer({ order: ordered });
  }

  function positionFor(id: string): number {
    return ordered.indexOf(id) + 1; // 1-indexed; 0 = unselected
  }

  function sourceCardName(opt: ReplacementOptionView): string {
    if (!opt.source_card_id || !snap.battlefield) return "";
    for (const c of snap.battlefield.cards) {
      if (c.instance_id === opt.source_card_id) return c.name ?? "";
    }
    // S19 dies-triggers: the source has already left for a
    // graveyard / exile by the time an ordering prompt shows.
    const elsewhere = triggerSourceName(opt.source_card_id);
    return elsewhere === "Triggered ability" ? "" : elsewhere;
  }

  // S17 sub-PR 6 optional-replacement branch — CR 614.10 "may"
  // prompt. Used by CR 903.9 commander-zone replacement today:
  // commander's owner picks yes (route to command zone) or no
  // (let the event proceed to graveyard/exile/hand/library).
  const isOptionalReplacement = $derived(active?.kind === "optional_replacement");

  // S19 sub-PR 2 trigger-prompt branch — CR 603.5 "you may" prompt
  // for an optional triggered ability. Same {choice_id, apply}
  // payload as optional-replacement; the server routes to
  // ResolveTriggerPrompt vs ResolveOptionalReplacement by inspecting
  // the choice's kind.
  const isTriggerPrompt = $derived(active?.kind === "trigger_prompt");

  // S19 sub-PR 6 pay-unless branch — CR 118.12 "unless that player
  // pays {N}" (Rhystic Study, Smothering Tithe, Esper Sentinel).
  // The chooser is the player being taxed, not the card's
  // controller. Same {choice_id, apply} payload; the server routes
  // to ResolvePayUnless by kind. "Pay" spends from the pool and
  // auto-taps untapped sources if the pool is short; a "Pay" the
  // player can't cover degrades to a decline server-side.
  const isPayUnless = $derived(active?.kind === "pay_unless");

  // S28 cascade branch — "you may cast it without paying its mana
  // cost". Same {choice_id, apply} payload; the server routes to
  // ResolveMayCast by kind. "Yes" stamps a free-cast permission on
  // the exiled card, which then casts out of exile like any other
  // impulse grant; "No" puts it on the bottom of the library with
  // the rest of the cards cascade turned over.
  const isMayCast = $derived(active?.kind === "may_cast");
  const mayCastCard = $derived(active?.options?.[0]);

  // Shockland entry branch — "as this land enters, you may pay 2
  // life. If you don't, it enters tapped." Same {choice_id, apply}
  // payload; the server routes by kind. The permanent is still in
  // hand while this is open, which is the point: the answer decides
  // how it ENTERS, so there is no tapped land on the board to look
  // at yet and the prompt has to say the card's name itself.
  const isEntryPayLife = $derived(active?.kind === "entry_pay_life");

  // #74 confirm — the chained-choice two-way prompt, "do A, or do B."
  // Same {choice_id, apply} payload as the other yes/no kinds; the
  // server routes to ResolveConfirm by kind, and the card supplies
  // the two button labels because "Pay 4 life" / "Put it on top" is
  // the actual question and "Yes" / "No" is not.
  //
  // This is also the prompt that arrives AFTER another prompt —
  // Ponder's "you may shuffle", Sylvan Library's per-card payment —
  // so it is routinely the second modal a player sees without having
  // done anything in between. Nothing extra is needed for that: the
  // queue drains in order and the modal reopens on the next frame.
  const isConfirm = $derived(active?.kind === "confirm");
  const confirmAccept = $derived(active?.accept_label || "Yes");
  const confirmDecline = $derived(active?.decline_label || "No");

  // #568 option_pick — "choose one of the following", CR 608.2. The
  // prompt an OPPONENT is asked while somebody else's spell resolves:
  // Torment of Hailfire's three-way question, and the pile a Fact or
  // Fiction chooser takes.
  //
  // A button per branch, answered with the INDEX. Not the card grid
  // below: the answer is which consequence, not which cards, and an
  // option's cards are context rather than the thing being picked.
  // An option whose cards this seat may not see arrives with its
  // label and no cards, which is the redaction pass working and not a
  // missing render — so the button is still live.
  const isOptionPick = $derived(active?.kind === "option_pick");
  const pickOptions = $derived<PickOptionView[]>(active?.pick_options ?? []);
  function answerOptionPick(index: number): void {
    if (!active || !viewerID) return;
    answer({ option_index: index });
  }

  // #764 mode_pick — CR 603.3c. A modal TRIGGERED ability's bullet,
  // chosen as the ability is put on the stack. A spell and an
  // activated ability need no prompt (the player who announces is
  // the player who chooses); a trigger has nobody to ask, because
  // the engine is what puts it on the stack.
  //
  // Only the bullets that can actually be taken are on the wire —
  // one whose clause has no legal target was dropped server-side —
  // so `mode_indexes` says which ModeSpec index each label is, and
  // that is what the answer sends back.
  const isModePick = $derived(active?.kind === "mode_pick");
  const modeLabels = $derived(active?.mode_options ?? []);
  const modeIndexes = $derived(active?.mode_indexes ?? []);
  const modeMin = $derived(active?.mode_min ?? 1);
  const modeMax = $derived(active?.mode_max ?? 1);
  const modeRepeatable = $derived(active?.mode_repeatable ?? false);
  // The chosen bullets IN THE ORDER CHOSEN — CR 700.2c resolves them
  // in that order, and CR 700.2d lets the same one appear twice.
  let modePicks = $state<number[]>([]);
  const modeSingle = $derived(modeMax === 1 && !modeRepeatable);
  const canConfirmModes = $derived(
    modePicks.length >= modeMin && (modeMax <= 0 || modePicks.length <= modeMax),
  );
  function toggleMode(idx: number): void {
    if (modeSingle) {
      modePicks = [idx];
      return;
    }
    if (!modeRepeatable && modePicks.includes(idx)) {
      modePicks = modePicks.filter((x) => x !== idx);
      return;
    }
    if (modeMax > 0 && modePicks.length >= modeMax) return;
    modePicks = [...modePicks, idx];
  }
  function modeTimes(idx: number): number {
    return modePicks.filter((x) => x === idx).length;
  }
  function answerModes(): void {
    if (!canConfirmModes) return;
    answer({ modes: modePicks });
  }

  // #804 loop_shortcut — CR 726. The loop breaker has fired and this
  // viewer controls the ability that is repeating, so they get the
  // question paper asks: how many more times? A number, not a yes/no,
  // because that is what CR 726 lets a player propose — and 0 is a
  // real answer ("stop here"), which leaves the table paused exactly
  // where the breaker put it, banner and all.
  const isLoopShortcut = $derived(active?.kind === "loop_shortcut");
  const loopCount = $derived(active?.loop_count ?? 0);
  const loopMax = $derived(active?.loop_max_iterations ?? 1000);
  let loopIterations = $state(10);
  // Re-seed the field whenever a shortcut prompt opens, so a second
  // ask does not arrive holding the first ask's number.
  let lastLoopID: string | null = null;
  $effect(() => {
    if (!isLoopShortcut || !active) {
      lastLoopID = null;
      return;
    }
    if (active.id === lastLoopID) return;
    lastLoopID = active.id;
    loopIterations = 10;
  });
  const loopAnswerable = $derived(
    Number.isFinite(loopIterations) && loopIterations >= 0 && loopIterations <= loopMax,
  );

  function submitLoopShortcut(iterations: number): void {
    if (!active || !viewerID) return;
    if (iterations < 0 || iterations > loopMax) return;
    answer({ iterations });
  }

  // #744 coin_call — one heads/tails answer covers the number of coins
  // in this instruction. A stop button is shown only for effects such
  // as Fiery Gambit that explicitly allow ending a winning chain.
  const isCoinCall = $derived(active?.kind === "coin_call");
  const coinCount = $derived(active?.coins ?? 1);
  const coinWins = $derived(active?.wins ?? 0);
  const coinAllowStop = $derived(active?.allow_stop === true);

  function answerCoin(call: "heads" | "tails" | "stop"): void {
    answer({ call });
  }

  // S21 sacrifice_choice branch — "each player sacrifices a creature
  // of their choice" (Grave Pact, Fleshbag Marauder). Reuses the
  // generic card grid and its {choice_id, card_ids} payload; only the
  // copy differs, because the default grid describes a discard from a
  // hand and this is a sacrifice from the battlefield.
  //
  // There is no cancel. A sacrifice cost of this kind isn't optional,
  // and the server has already filtered the options to permanents the
  // chooser controls — a player with none was never prompted.
  const isSacrifice = $derived(active?.kind === "sacrifice_choice");

  // S21 scry branch — CR 701.22. Every looked-at card goes somewhere:
  // back on top (in an order the player controls) or to the bottom.
  // Default is "keep everything, in the order shown", so the common
  // case — bottom the one bad card, or accept the top — is one click
  // or none.
  //
  // Answered with {bottom, top_order}; top_order is TOP-FIRST, so its
  // first entry is the next card drawn.
  const isScry = $derived(active?.kind === "scry");

  // S22 surveil branch — CR 701.25. Structurally identical to scry:
  // same prompt, same two lanes, same ordering control. The only
  // difference is where the cards that leave the top go, so this
  // shares every line of the scry branch and swaps the destination
  // in the copy and in the submitted payload key.
  //
  // The distinction is worth the copy: a card put on the bottom of a
  // library is gone, and a card put in a graveyard is a resource. A
  // dialog that said "to bottom" on a surveil would be offering the
  // wrong move.
  const isSurveil = $derived(active?.kind === "surveil");

  // S22 "look at the top N cards of your library, then put them back
  // in any order" — Ponder, Sensei's Divining Top. The family's third
  // member and the one with NO away lane: every card goes back on
  // top, so the away column and its buttons are hidden entirely and
  // the answer is a pure reorder.
  // #996 / ADR 0088: "put these cards on top of / on the bottom of
  // the library in any order". The family's fourth member, and the
  // one whose lanes come off the prompt: `placement` says whether the
  // answer is a reorder on top (Brainstorm's put-back), a reorder of a
  // pile going UNDER the library ("the rest on the bottom in any
  // order"), or scry's two lanes without the keyword (Aetherspouts'
  // "top or bottom").
  const isPutInLibrary = $derived(active?.kind === "put_in_library");
  const placement = $derived(isPutInLibrary ? (active?.placement ?? "top") : null);

  const isReorderOnly = $derived(
    active?.kind === "look_at_top" || (isPutInLibrary && placement === "top"),
  );

  // Everything goes under the library: the top lane is hidden and the
  // bottom lane is the whole answer.
  const isBottomOnly = $derived(isPutInLibrary && placement === "bottom");

  // The shared branch. Everything below keys off this; the flags
  // above only pick the wording, the lanes and the payload.
  const isLookAtTop = $derived(isScry || isSurveil || isReorderOnly || isPutInLibrary);

  // The lane the cards leaving the top go to, as the card prints it.
  // Unused when there is no away lane.
  const awayLabel = $derived(isSurveil ? "graveyard" : "bottom");

  let scryTop = $state<string[]>([]);
  let scryBottom = $state<string[]>([]);

  // Seed the default whenever a scry / surveil prompt opens:
  // everything stays on top, in the order the server listed it (which
  // is current library order).
  let lastScryID: string | null = null;
  $effect(() => {
    if (!isLookAtTop || !active) {
      lastScryID = null;
      return;
    }
    if (active.id === lastScryID) return;
    lastScryID = active.id;
    const all = (active.options ?? []).map((c) => c.instance_id);
    scryTop = isBottomOnly ? [] : all;
    scryBottom = isBottomOnly ? all : [];
  });

  function scryToBottom(id: string): void {
    scryTop = scryTop.filter((x) => x !== id);
    if (!scryBottom.includes(id)) scryBottom = [...scryBottom, id];
  }

  function scryToTop(id: string): void {
    scryBottom = scryBottom.filter((x) => x !== id);
    if (!scryTop.includes(id)) scryTop = [...scryTop, id];
  }

  // Move a kept card one place closer to the top. The only ordering
  // control needed: scry N is 1 or 2 on every printed card, so "swap
  // these two" is the whole requirement, and this generalises to 3.
  function scryMoveUp(id: string): void {
    const i = scryTop.indexOf(id);
    if (i <= 0) return;
    const next = [...scryTop];
    [next[i - 1], next[i]] = [next[i], next[i - 1]];
    scryTop = next;
  }

  // The same control on the bottom lane. CR 701.22a puts a scry's
  // bottom cards there "in any order" too, and ADR 0088's "the rest on
  // the bottom in any order" is nothing BUT this lane. The list is
  // top-first: the last entry is the bottom card of the library.
  function scryBottomMoveUp(id: string): void {
    const i = scryBottom.indexOf(id);
    if (i <= 0) return;
    const next = [...scryBottom];
    [next[i - 1], next[i]] = [next[i], next[i - 1]];
    scryBottom = next;
  }

  function scryCardName(id: string): string {
    const c = (active?.options ?? []).find((o) => o.instance_id === id);
    return c?.name || "card";
  }

  // Each member of the family answers on its own destination key.
  // The server routes on the prompt's kind rather than on the keys,
  // so a wrong key is a rejection rather than a silent miscarriage —
  // but send the right one anyway: it is what the resolver reads.
  function submitScry(): void {
    if (!active || !viewerID) return;
    const total = (active.options ?? []).length;
    if (scryTop.length + scryBottom.length !== total) return;
    // put_in_library sends both lanes whatever its placement; the one
    // the placement does not open is empty, which the server accepts.
    const away = isPutInLibrary
      ? { bottom: scryBottom }
      : isReorderOnly
        ? {}
        : isSurveil
          ? { graveyard: scryBottom }
          : { bottom: scryBottom };
    answer({ ...away, top_order: scryTop });
  }

  // S19 follow-up: the server flags optional triggers whose effect
  // has no legal target (Reclamation Sage with no opponent artifact,
  // Eternal Witness with an empty graveyard). Until the S20 target
  // picker lands, the auto-targeter silently no-ops in that case —
  // which reads as a bug. Warn the chooser and relabel "Yes".
  const noLegalTarget = $derived(active?.no_legal_target === true);
  const doubledLabel = $derived(doubledTriggerLabel(active?.doubled_by, active?.doubled_by_name));

  // Y / N answer the yes-no prompts (optional replacement, may-
  // trigger, pay-unless) from the keyboard; the footer shows the
  // hint. Ignored while typing in a field.
  const isYesNo = $derived(
    isOptionalReplacement ||
      isTriggerPrompt ||
      isPayUnless ||
      isEntryPayLife ||
      isMayCast ||
      isConfirm,
  );
  function handleKey(e: KeyboardEvent): void {
    if (!open) return;
    const t = e.target as HTMLElement | null;
    if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable)) return;
    if (isCoinCall) {
      if (e.key === "h" || e.key === "H") {
        e.preventDefault();
        answerCoin("heads");
      } else if (e.key === "t" || e.key === "T") {
        e.preventDefault();
        answerCoin("tails");
      } else if (coinAllowStop && (e.key === "s" || e.key === "S")) {
        e.preventDefault();
        answerCoin("stop");
      }
      return;
    }
    if (!isYesNo) return;
    if (e.key === "y" || e.key === "Y") {
      e.preventDefault();
      answerOptional(true);
    } else if (e.key === "n" || e.key === "N") {
      e.preventDefault();
      answerOptional(false);
    }
  }

  function answerOptional(apply: boolean): void {
    if (!active || !viewerID) return;
    answer({ apply });
  }

  // sourceCardName resolves the source-card display name for a
  // trigger prompt. Walks battlefield + every seated player's
  // graveyard + the shared exile zone — LTB triggers prompt after
  // the source has already moved off the battlefield, so the
  // lookup has to span destination zones. Falls back to a generic
  // string when the card isn't visible to the viewer (redacted
  // CardView entries arrive with empty names).
  function triggerSourceName(sourceID?: string): string {
    if (!sourceID) return "Triggered ability";
    const seek = (cards: CardView[] | undefined) => {
      if (!cards) return undefined;
      for (const c of cards) {
        if (c.instance_id === sourceID && c.name) return c.name;
      }
      return undefined;
    };
    const fromBF = seek(snap.battlefield?.cards);
    if (fromBF) return fromBF;
    const fromExile = seek(snap.exile?.cards);
    if (fromExile) return fromExile;
    for (const p of snap.seats ?? []) {
      const fromGY = seek(p.graveyard?.cards);
      if (fromGY) return fromGY;
    }
    return "Triggered ability";
  }

  // S18 damage_assignment branch — CR 510.1c multi-blocker combat
  // damage prompt. The attacker's controller reorders the blockers
  // (drag via up/down buttons since drag-and-drop UX lives outside
  // this sprint) and assigns damage per blocker. Server validates
  // at-least-lethal prefix + total = attacker_power, and trample
  // overflow if AllowTrample.
  const isDamageAssignment = $derived(active?.kind === "damage_assignment");
  const damageFrame = $derived<DamageAssignmentView | null>(active?.damage_assignment ?? null);

  // Ordered blocker IDs (reorderable). Damage amount per blocker,
  // keyed by blocker ID. Trample-to-player bucket.
  let blockerOrder = $state<string[]>([]);
  let damageAmounts = $state<Record<string, number>>({});
  let trampleToPlayer = $state(0);

  // Reset assignment state when the prompt's identity changes — the
  // same untracked last-id pattern as the selected/ordered reset
  // above. Keying on content equality with blockerOrder here would
  // make the user's own ▲/▼ reorder re-trigger the effect and revert
  // their order (and zero their amounts) on the first click.
  let lastDamageChoiceID: string | null = null;
  $effect(() => {
    const nextID = active?.id ?? null;
    if (nextID === lastDamageChoiceID) return;
    lastDamageChoiceID = nextID;
    if (!damageFrame) return;
    blockerOrder = [...damageFrame.blocker_card_ids];
    const next: Record<string, number> = {};
    for (const id of damageFrame.blocker_card_ids) next[id] = 0;
    damageAmounts = next;
    trampleToPlayer = 0;
  });

  const assignedTotal = $derived(
    blockerOrder.reduce((acc, id) => acc + (damageAmounts[id] ?? 0), 0) + trampleToPlayer,
  );

  const canSubmitAssignment = $derived(
    damageFrame !== null && assignedTotal === damageFrame.attacker_power,
  );

  function moveBlocker(id: string, delta: -1 | 1): void {
    const idx = blockerOrder.indexOf(id);
    if (idx < 0) return;
    const target = idx + delta;
    if (target < 0 || target >= blockerOrder.length) return;
    const next = [...blockerOrder];
    [next[idx], next[target]] = [next[target], next[idx]];
    blockerOrder = next;
  }

  function setDamageAmount(id: string, raw: string): void {
    const n = Math.max(0, Math.floor(Number(raw) || 0));
    damageAmounts = { ...damageAmounts, [id]: n };
  }

  function setTrampleAmount(raw: string): void {
    const n = Math.max(0, Math.floor(Number(raw) || 0));
    trampleToPlayer = n;
  }

  function blockerName(id: string): string {
    if (!snap.battlefield) return id.slice(0, 8);
    const card = snap.battlefield.cards.find((c) => c.instance_id === id);
    return card?.name ?? id.slice(0, 8);
  }

  function attackerName(id: string): string {
    if (!snap.battlefield) return id.slice(0, 8);
    const card = snap.battlefield.cards.find((c) => c.instance_id === id);
    return card?.name ?? id.slice(0, 8);
  }

  function submitDamageAssignment(): void {
    if (!active || !viewerID || !damageFrame) return;
    if (!canSubmitAssignment) return;
    const assignments = blockerOrder.map((id) => ({
      blocker_id: id,
      amount: damageAmounts[id] ?? 0,
    }));
    answer({ assignments, trample_to_player: trampleToPlayer });
  }
</script>

{#if open && active}
  <ModalLayer />
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="choice-title">
    <div class="prompt-modal">
      {#if isLookAtTop}
        <h2 id="choice-title">
          {active.reason ||
            (isPutInLibrary
              ? "Put them in your library"
              : isSurveil
                ? "Surveil"
                : isReorderOnly
                  ? "Look at the top"
                  : "Scry")}
          {#if isPutInLibrary}
            <span class="prompt-src" aria-hidden="true">CR 401.4</span>
          {:else if !isReorderOnly}
            <span class="prompt-src" aria-hidden="true"
              >{isSurveil ? "CR 701.25" : "CR 701.22"}</span
            >
          {/if}
        </h2>
        <p class="prompt-hint">
          {#if isBottomOnly}
            Put them on the bottom of the library in any order — the last one listed is the very
            bottom card.
          {:else if isReorderOnly}
            Put them {isPutInLibrary ? "on top" : "back"} in any order — the topmost is your next draw.
          {:else if scryTop.length + scryBottom.length === 1}
            Keep it on top, or put it {isSurveil
              ? "into your graveyard"
              : "on the bottom of your library"}.
          {:else}
            Keep any of these on top — the topmost is your next draw — and put the rest {isSurveil
              ? "into your graveyard"
              : "on the bottom"}.
          {/if}
          Only you can see them.
        </p>
        {#if !isBottomOnly}
          <div class="scry-lane">
            <h3 class="lane-label">On top ({scryTop.length})</h3>
            {#if scryTop.length === 0}
              <p class="lane-empty">Nothing — your next draw comes from under these.</p>
            {:else}
              <ol class="scry-list">
                {#each scryTop as id, i (id)}
                  <li>
                    <span class="prompt-num">{i + 1}</span>
                    <span class="scry-name">{scryCardName(id)}</span>
                    <button
                      type="button"
                      class="lane-btn"
                      disabled={i === 0}
                      title="move closer to the top"
                      aria-label={`move ${scryCardName(id)} up`}
                      onclick={() => scryMoveUp(id)}>↑</button
                    >
                    {#if !isReorderOnly}
                      <button
                        type="button"
                        class="lane-btn"
                        onclick={() => scryToBottom(id)}
                        aria-label={`put ${scryCardName(id)} ${
                          isSurveil ? "into your graveyard" : "on the bottom"
                        }`}>To {awayLabel}</button
                      >
                    {/if}
                  </li>
                {/each}
              </ol>
            {/if}
          </div>
        {/if}
        {#if !isReorderOnly}
          <div class="scry-lane">
            <h3 class="lane-label">
              {isSurveil ? "Into your graveyard" : "On the bottom"} ({scryBottom.length})
            </h3>
            {#if scryBottom.length === 0}
              <p class="lane-empty">None.</p>
            {:else}
              <ol class="scry-list">
                {#each scryBottom as id, i (id)}
                  <li>
                    {#if !isSurveil}
                      <span class="prompt-num">{i + 1}</span>
                    {/if}
                    <span class="scry-name">{scryCardName(id)}</span>
                    {#if !isSurveil}
                      <button
                        type="button"
                        class="lane-btn"
                        disabled={i === 0}
                        title="move closer to the top of the pile"
                        aria-label={`move ${scryCardName(id)} up in the bottom pile`}
                        onclick={() => scryBottomMoveUp(id)}>↑</button
                      >
                    {/if}
                    {#if !isBottomOnly}
                      <button
                        type="button"
                        class="lane-btn"
                        onclick={() => scryToTop(id)}
                        aria-label={`keep ${scryCardName(id)} on top`}>Keep on top</button
                      >
                    {/if}
                  </li>
                {/each}
              </ol>
            {/if}
          </div>
        {/if}
        <div class="card-grid">
          {#each optionCards as c (c.instance_id)}
            <div class="card-pick" class:bottomed={scryBottom.includes(c.instance_id)}>
              <Card card={c} />
            </div>
          {/each}
        </div>
        <div class="prompt-foot">
          <span class="prompt-count">
            {#if isBottomOnly}
              {scryBottom.length} on the bottom
            {:else if isReorderOnly}
              {scryTop.length} {isPutInLibrary ? "on top" : "back on top"}
            {:else}
              {scryTop.length} on top · {scryBottom.length}
              {isSurveil ? "in the graveyard" : "on the bottom"}
            {/if}
          </span>
          <button type="button" class="primary" onclick={submitScry}>Done</button>
        </div>
      {:else if isManaPick}
        <h2 id="choice-title">
          {active.reason || "Pick a color"}
          <span class="prompt-src" aria-hidden="true">mana ability</span>
        </h2>
        <p class="prompt-hint">
          {#if hasColorAmounts}
            Choose one color. All of this mana is added in that color.
          {:else}
            Choose a color to add to your mana pool.
          {/if}
        </p>
        <div class="color-row">
          {#each buttons as b (b.color)}
            <button
              type="button"
              class="color-pick"
              style:--fill={b.fill}
              title={b.label}
              aria-label={b.amount > 1 ? `add ${b.amount} ${b.label} mana` : `add ${b.label} mana`}
              onclick={() => pickColor(b.color)}
            >
              <span class="color-letter">{b.amount > 1 ? `${b.amount}×${b.color}` : b.color}</span>
              <span class="color-name">{b.label}</span>
            </button>
          {/each}
        </div>
      {:else if isColorChoice}
        <h2 id="choice-title">
          {active.reason || colorCopy.title}
          <span class="prompt-src" aria-hidden="true">choose a color · CR 105.4</span>
        </h2>
        <!-- The hint is the card's declared purpose put into words
             (#780 / #986). A prompt with no purpose falls back to the
             neutral line, because without one the view genuinely does
             not know whether the colour is remembered on a permanent
             or used once as a spell resolves. -->
        <p class="prompt-hint">{colorCopy.hint}</p>
        <div class="color-row">
          {#each buttons as b (b.color)}
            <button
              type="button"
              class="color-pick"
              style:--fill={b.fill}
              title={b.label}
              aria-label={`choose ${b.label}`}
              onclick={() => pickColor(b.color)}
            >
              <span class="color-letter">{b.color}</span>
              <span class="color-name">{b.label}</span>
            </button>
          {/each}
        </div>
      {:else if isCoinCall}
        <h2 id="choice-title">
          {active.reason || "Call the flip"}
          <span class="prompt-src" aria-hidden="true">coin flip</span>
        </h2>
        <p class="prompt-hint">
          Call heads or tails for {coinCount}
          {coinCount === 1 ? "coin" : "coins"}.
          {#if coinWins > 0}
            You have won {coinWins} {coinWins === 1 ? "flip" : "flips"} so far.
          {/if}
        </p>
        <div class="prompt-foot">
          <span class="prompt-count">
            <span class="kbd">H</span> heads · <span class="kbd">T</span> tails
            {#if coinAllowStop}
              · <span class="kbd">S</span> stop{/if}
          </span>
          <button type="button" onclick={() => answerCoin("heads")}>Heads</button>
          <button type="button" class="primary" onclick={() => answerCoin("tails")}>Tails</button>
          {#if coinAllowStop}
            <button type="button" class="ghost" onclick={() => answerCoin("stop")}>Stop</button>
          {/if}
        </div>
      {:else if isLoopShortcut}
        <h2 id="choice-title">
          {active.reason || "This ability keeps resolving"}
          <span class="prompt-src" aria-hidden="true">shortcut · CR 726</span>
        </h2>
        <p class="prompt-hint">
          It has resolved {loopCount}
          {loopCount === 1 ? "time" : "times"} this turn with nobody doing anything in between. Say how
          many more times it should resolve and the table will run them without stopping; stop here leaves
          auto-pass paused so you can step through by hand.
        </p>
        <div class="prompt-foot">
          <label class="loop-iterations">
            <span>More times</span>
            <input
              type="number"
              min="0"
              max={loopMax}
              step="1"
              bind:value={loopIterations}
              aria-label="How many more times to resolve it"
            />
          </label>
          <button type="button" class="ghost" onclick={() => submitLoopShortcut(0)}
            >Stop here</button
          >
          <button
            type="button"
            class="primary"
            disabled={!loopAnswerable}
            onclick={() => submitLoopShortcut(loopIterations)}
          >
            Resolve {loopIterations} more
          </button>
        </div>
      {:else if isCreatureTypePick}
        <h2 id="choice-title">
          {active.reason || "Choose a creature type"}
          <span class="prompt-src" aria-hidden="true">as this enters · CR 614.12</span>
        </h2>
        <p class="prompt-hint">
          The choice is locked in for as long as this permanent stays on the battlefield.
        </p>
        <!-- svelte-ignore a11y_autofocus -->
        <input
          class="type-filter"
          type="text"
          autofocus
          placeholder="Filter creature types…"
          aria-label="Filter creature types"
          bind:value={typeFilter}
          onkeydown={onTypeFilterKey}
        />
        <div class="type-list">
          {#each filteredTypes as t (t)}
            <button type="button" class="type-pick" onclick={() => pickCreatureType(t)}>{t}</button>
          {:else}
            <p class="prompt-hint warn">No creature type matches “{typeFilter}”.</p>
          {/each}
        </div>
      {:else if isCardNamePick}
        <h2 id="choice-title">
          {active.reason || "Choose a card name"}
          <span class="prompt-src" aria-hidden="true">as this enters · CR 614.12</span>
        </h2>
        <p class="prompt-hint">
          Any card name is legal — type one. The suggestions are the cards everyone can currently
          see.
        </p>
        <!-- svelte-ignore a11y_autofocus -->
        <input
          class="type-filter"
          type="text"
          autofocus
          placeholder="Name a card…"
          aria-label="Name a card"
          bind:value={nameFilter}
          onkeydown={onNameFilterKey}
        />
        <div class="type-list">
          {#each filteredNames as n (n)}
            <button type="button" class="type-pick" onclick={() => pickCardName(n)}>{n}</button>
          {:else}
            <p class="prompt-hint">No visible card matches — press Enter to name it anyway.</p>
          {/each}
        </div>
        <div class="prompt-foot">
          <button
            type="button"
            class="primary"
            disabled={nameFilter.trim() === ""}
            onclick={() => pickCardName(nameFilter)}
          >
            Name “{nameFilter.trim() || "…"}”
          </button>
        </div>
      {:else if isOptionalReplacement}
        <h2 id="choice-title">
          {active.reason || "Apply replacement?"}
          <span class="prompt-src" aria-hidden="true">optional replacement · CR 614.10</span>
        </h2>
        <p class="prompt-hint">
          You (the affected player) decide whether this substitution applies.
        </p>
        <div class="prompt-foot">
          <span class="prompt-count"><span class="kbd">Y</span> / <span class="kbd">N</span></span>
          <button type="button" onclick={() => answerOptional(false)}>No</button>
          <button type="button" class="primary" onclick={() => answerOptional(true)}>Yes</button>
        </div>
      {:else if isTriggerPrompt}
        <h2 id="choice-title">
          {active.reason || `${triggerSourceName(active.source)} triggered`}
          <span class="prompt-src" aria-hidden="true">may trigger · CR 603.5</span>
          {#if doubledLabel}
            <span class="prompt-src">{doubledLabel}</span>
          {/if}
        </h2>
        {#if noLegalTarget}
          <p class="prompt-hint warn">
            No legal target — “Yes” passes without effect (picker lands in S20).
          </p>
        {:else}
          <p class="prompt-hint">Fire the ability, or let it pass without effect.</p>
        {/if}
        <div class="prompt-foot">
          <span class="prompt-count"><span class="kbd">Y</span> / <span class="kbd">N</span></span>
          <button type="button" onclick={() => answerOptional(false)}>No</button>
          <button type="button" class="primary" onclick={() => answerOptional(true)}>Yes</button>
        </div>
      {:else if isMayCast}
        <h2 id="choice-title">
          {active.reason || "Cast it without paying its mana cost?"}
          <span class="prompt-src" aria-hidden="true">cascade · CR 702.85</span>
        </h2>
        <p class="prompt-hint">
          {#if mayCastCard}
            <strong>{mayCastCard.name}</strong> is exiled face up.
          {/if}
          Say yes and it stays in exile, castable for nothing until end of turn. Say no and it goes to
          the bottom of your library with everything else cascade turned over.
        </p>
        <div class="prompt-foot">
          <span class="prompt-count"><span class="kbd">Y</span> / <span class="kbd">N</span></span>
          <button type="button" onclick={() => answerOptional(false)}>To the bottom</button>
          <button type="button" class="primary" onclick={() => answerOptional(true)}>
            Cast it free
          </button>
        </div>
      {:else if isEntryPayLife}
        <h2 id="choice-title">
          {active.reason || `Pay ${active.pay_cost ?? ""} as it enters?`}
          <span class="prompt-src" aria-hidden="true">as this enters · CR 614</span>
        </h2>
        <p class="prompt-hint">
          Pay {active.pay_cost ?? "the life"} and it enters untapped; don't, and it enters tapped. Nothing
          has entered yet — this choice is part of the entry, not a trigger.
        </p>
        <div class="prompt-foot">
          <span class="prompt-count"><span class="kbd">Y</span> / <span class="kbd">N</span></span>
          <button type="button" onclick={() => answerOptional(false)}>Enter tapped</button>
          <button type="button" class="primary" onclick={() => answerOptional(true)}>
            Pay {active.pay_cost ?? ""}
          </button>
        </div>
      {:else if isOptionPick}
        <h2 id="choice-title">
          {active.reason || "Choose one"}
          <span class="prompt-src" aria-hidden="true">choose one · CR 608.2</span>
        </h2>
        <p class="prompt-hint">
          Someone else's spell or ability is asking you. Every option listed is one you can take,
          and the game waits until you pick one.
        </p>
        <ul class="pick-options">
          {#each pickOptions as opt, i (i)}
            <li>
              <button type="button" class="pick-option" onclick={() => answerOptionPick(i)}>
                <span class="pick-label">{opt.label}</span>
                {#if opt.cards && opt.cards.length > 0}
                  <span class="pick-cards">
                    {#each opt.cards as c (c.instance_id)}
                      <Card card={c} />
                    {/each}
                  </span>
                {/if}
              </button>
            </li>
          {/each}
        </ul>
      {:else if isModePick}
        <h2 id="choice-title">
          {triggerSourceName(active.source)}
          <span class="prompt-src" aria-hidden="true">{active.reason || "choose one"}</span>
        </h2>
        <p class="prompt-hint">
          The ability is not on the stack until you answer — its mode is chosen as it goes there (CR
          603.3c), and any targets it asks for come after.
          {#if modeRepeatable}
            You may choose the same mode more than once.
          {/if}
        </p>
        <ul class="prompt-options" role={modeSingle ? "radiogroup" : "group"}>
          {#each modeLabels as label, i (i)}
            {@const idx = modeIndexes[i] ?? i}
            {@const times = modeTimes(idx)}
            <li>
              <button
                type="button"
                class="prompt-opt"
                class:on={times > 0}
                role={modeSingle ? "radio" : "checkbox"}
                aria-checked={times > 0}
                onclick={() => toggleMode(idx)}
              >
                <span class="prompt-radio" aria-hidden="true"></span>
                <span class="mode-label">{label}</span>
                {#if modeRepeatable && times > 0}
                  <span class="mode-times">&times;{times}</span>
                {/if}
              </button>
            </li>
          {/each}
        </ul>
        <div class="prompt-foot">
          <span class="prompt-count">{modePicks.length} / {modeMax > 0 ? modeMax : "any"}</span>
          <button type="button" class="primary" disabled={!canConfirmModes} onclick={answerModes}>
            Choose
          </button>
        </div>
      {:else if isConfirm}
        <h2 id="choice-title">
          {active.reason || "Choose one"}
          <span class="prompt-src" aria-hidden="true">choose one</span>
        </h2>
        <p class="prompt-hint">
          Both answers are legal — this is a choice between two things the card does, not a
          yes-or-no. There may be another question after it.
        </p>
        <div class="prompt-foot">
          <span class="prompt-count"><span class="kbd">Y</span> / <span class="kbd">N</span></span>
          <button type="button" onclick={() => answerOptional(false)}>{confirmDecline}</button>
          <button type="button" class="primary" onclick={() => answerOptional(true)}>
            {confirmAccept}
          </button>
        </div>
      {:else if isPayUnless}
        <h2 id="choice-title">
          {active.reason || `${triggerSourceName(active.source)} — pay ${active.pay_cost ?? ""}?`}
          <span class="prompt-src" aria-hidden="true">pay unless</span>
        </h2>
        <p class="prompt-hint">
          Pay {active.pay_cost ?? "the cost"} from your pool (untapped sources auto-tap if it's short),
          or don't and let {triggerSourceName(active.source)} do its thing.
        </p>
        <div class="prompt-foot">
          <span class="prompt-count"><span class="kbd">Y</span> / <span class="kbd">N</span></span>
          <button type="button" onclick={() => answerOptional(false)}>Don't pay</button>
          <button type="button" class="primary" onclick={() => answerOptional(true)}>
            Pay {active.pay_cost ?? ""}
          </button>
        </div>
      {:else if isDamageAssignment && damageFrame}
        <h2 id="choice-title">
          {active.reason || "Assign combat damage"}
          <span class="prompt-src" aria-hidden="true">CR 510.1c</span>
        </h2>
        <p class="prompt-hint">
          <strong>{attackerName(damageFrame.attacker_card_id)}</strong>
          is blocked by {damageFrame.blocker_card_ids.length} creatures. Order them and divide
          {damageFrame.attacker_power} damage — earlier blockers must be dealt at-least-lethal before
          the next gets any.
          {#if damageFrame.allow_trample}
            Trample lets leftover damage spill to the defending player.
          {/if}
          {#if damageFrame.has_deathtouch}
            Deathtouch makes 1 damage lethal.
          {/if}
        </p>
        <ul class="assign-list">
          {#each blockerOrder as id, i (id)}
            <li class="assign-row">
              <div class="assign-order">
                <button
                  type="button"
                  class="reorder-btn"
                  disabled={i === 0}
                  onclick={() => moveBlocker(id, -1)}
                  aria-label={`move ${blockerName(id)} up`}
                >
                  ▲
                </button>
                <span class="assign-pos">{i + 1}</span>
                <button
                  type="button"
                  class="reorder-btn"
                  disabled={i === blockerOrder.length - 1}
                  onclick={() => moveBlocker(id, 1)}
                  aria-label={`move ${blockerName(id)} down`}
                >
                  ▼
                </button>
              </div>
              <span class="assign-name">{blockerName(id)}</span>
              <label class="assign-input">
                <span class="sr-only">damage to {blockerName(id)}</span>
                <input
                  type="number"
                  min="0"
                  max={damageFrame.attacker_power}
                  value={damageAmounts[id] ?? 0}
                  oninput={(e) => setDamageAmount(id, (e.currentTarget as HTMLInputElement).value)}
                />
              </label>
            </li>
          {/each}
          {#if damageFrame.allow_trample}
            <li class="assign-row trample">
              <div class="assign-order"><span class="assign-pos">→</span></div>
              <span class="assign-name">Defending player (trample)</span>
              <label class="assign-input">
                <span class="sr-only">trample damage to defending player</span>
                <input
                  type="number"
                  min="0"
                  max={damageFrame.attacker_power}
                  value={trampleToPlayer}
                  oninput={(e) => setTrampleAmount((e.currentTarget as HTMLInputElement).value)}
                />
              </label>
            </li>
          {/if}
        </ul>
        <div class="prompt-foot">
          <span class="prompt-count">{assignedTotal} / {damageFrame.attacker_power} assigned</span>
          <button
            type="button"
            class="primary"
            disabled={!canSubmitAssignment}
            onclick={submitDamageAssignment}
          >
            Deal damage
          </button>
        </div>
      {:else if isReplacementOrder || isTriggerOrder}
        <h2 id="choice-title">
          {active.reason || (isTriggerOrder ? "Order your triggers" : "Order replacement effects")}
          <span class="prompt-src" aria-hidden="true"
            >{isTriggerOrder ? "CR 603.3b" : "CR 616"}</span
          >
        </h2>
        <p class="prompt-hint">
          {#if isTriggerOrder}
            Two or more of your abilities triggered at once. Click them in the order they should
            resolve — the first you pick resolves first.
          {:else}
            Click each effect in the order it should apply. Different orders can produce different
            results — you choose as the affected player.
          {/if}
        </p>
        <ul class="prompt-options">
          {#each replacementOptions as opt (opt.id)}
            {@const pos = positionFor(opt.id)}
            {@const src = sourceCardName(opt)}
            <li>
              <button
                type="button"
                class="prompt-opt"
                class:on={pos > 0}
                onclick={() => toggleReplacement(opt.id)}
                aria-pressed={pos > 0}
                aria-label={`${pos > 0 ? "deselect" : "select"} ${opt.label || "effect"}`}
              >
                <span class="prompt-num">{pos > 0 ? pos : "·"}</span>
                <span class="order-label">
                  <strong>{opt.label || "Replacement effect"}</strong>
                  {#if src}<span class="order-src">{src}</span>{/if}
                </span>
              </button>
            </li>
          {/each}
        </ul>
        <div class="prompt-foot">
          <span class="prompt-count">{ordered.length} / {replacementOptions.length} ordered</span>
          <button
            type="button"
            class="primary"
            disabled={ordered.length !== replacementOptions.length}
            onclick={submitReplacementOrder}
          >
            {isTriggerOrder ? "Resolve in this order" : "Apply in this order"}
          </button>
        </div>
      {:else}
        <h2 id="choice-title">
          {#if isSacrifice}
            {active.reason || "Sacrifice a permanent"}
            <span class="prompt-src" aria-hidden="true">sacrifice</span>
          {:else if isSearch}
            {active.reason || "Search your library"}
            <span class="prompt-src" aria-hidden="true">search · CR 701.23</span>
          {:else if isCopyTarget}
            {active.reason || "Enter as a copy of…"}
            <span class="prompt-src" aria-hidden="true">copy · CR 707</span>
          {:else if isUntapChoice}
            {active.reason || "Untap step — choose which permanents untap"}
            <span class="prompt-src" aria-hidden="true">untap · CR 502.3</span>
          {:else if isEntryReveal}
            {active.reason || "Reveal a card from your hand?"}
            <span class="prompt-src" aria-hidden="true">reveal · CR 614</span>
          {:else if isChooseCards}
            {active.reason || "Choose cards"}
            <span class="prompt-src" aria-hidden="true">choose</span>
          {:else if isRevealPick}
            {active.reason || "Choose from the revealed cards"}
            <span class="prompt-src" aria-hidden="true">reveal · CR 701.20</span>
          {:else if isPermanentPick}
            {active.reason || "Choose permanents"}
            <span class="prompt-src" aria-hidden="true">choose · CR 608.2</span>
          {:else}
            {active.reason || "Choose"} — pick {active.count} card{active.count === 1 ? "" : "s"}
            <span class="prompt-src" aria-hidden="true">{isSelfSource ? "discard" : "reveal"}</span>
          {/if}
        </h2>
        <p class="prompt-hint">
          {#if isSacrifice}
            Choose {active.count === 1 ? "a permanent" : `${active.count} permanents`} you control to
            sacrifice. This isn't optional — {triggerSourceName(active.source)} is making you.
          {:else if isSearch}
            {#if pickMax === 1}
              Take one of these, or none.
            {:else}
              Take up to {pickMax} of these, or none.
            {/if}
            Only you can see them, and your library is shuffled either way.
          {:else if isCopyTarget}
            Pick what it enters as a copy of — it copies the printed card, so counters, damage and
            other effects don't come across. Or copy nothing and let it enter as itself.
          {:else if isUntapChoice}
            {#if pickMin === 0}
              These don't have to untap. Leave any of them tapped, or untap them all.
            {:else if pickMin === pickMax}
              Only {pickMax} of these can untap this turn.
            {:else}
              Between {pickMin} and {pickMax} of these can untap this turn; the rest stay tapped.
            {/if}
            Nothing else on your board is affected — everything that could untap without a decision already
            has.
          {:else if isEntryReveal}
            Show {pickMax === 1 ? "one of these" : `up to ${pickMax} of these`} to the table and it enters
            untapped. Revealing costs nothing — the card stays in your hand — but everyone gets to see
            it, and you may show nothing instead.
          {:else if isChooseCards}
            {#if pickMin === pickMax}
              Pick {pickMax} of these.
            {:else if pickMin === 0}
              Pick up to {pickMax} of these, or none.
            {:else}
              Pick between {pickMin} and {pickMax} of these.
            {/if}
            What happens to them is the card's business, and it will tell you next — choosing them costs
            nothing on its own.
          {:else if isRevealPick}
            {#if pickMin === pickMax}
              Pick {pickMax} of these.
            {:else if pickMin === 0}
              Pick any of these, or none.
            {:else}
              Pick between {pickMin} and {pickMax} of these.
            {/if}
            <strong>{fromName}</strong> revealed them, so the whole table can see them — and what happens
            to the ones you leave is the card's business.
          {:else if isTheirPermanents}
            {#if pickMin === pickMax}
              Pick {pickMax} of <strong>{fromName}</strong>'s permanents.
            {:else if pickMin === 0}
              Pick any of <strong>{fromName}</strong>'s permanents, or none.
            {:else}
              Pick between {pickMin} and {pickMax} of <strong>{fromName}</strong>'s permanents.
            {/if}
            Nothing here is targeted, so hexproof and shroud don't protect anything from being chosen.
          {:else if isOwnPermanents}
            {#if pickMin === pickMax}
              Pick {pickMax} of your permanents.
            {:else if pickMin === 0}
              Pick any number of your permanents, or none.
            {:else}
              Pick between {pickMin} and {pickMax} of your permanents.
            {/if}
            Nothing here is targeted — the choice is being made now, as the card resolves.
          {:else if isSelfSource}
            Pick {active.count} card{active.count === 1 ? "" : "s"} from your hand to discard.
          {:else}
            Pick {active.count} card{active.count === 1 ? "" : "s"} from
            <strong>{fromName}</strong>'s revealed hand.
            <strong>{fromName}</strong> will discard your pick{active.count === 1 ? "" : "s"}.
          {/if}
        </p>
        <div class="card-grid">
          {#each optionCards as c (c.instance_id)}
            <button
              type="button"
              class="card-pick"
              class:selected={selected.has(c.instance_id)}
              disabled={!selected.has(c.instance_id) && selected.size >= pickMax}
              onclick={() => toggle(c.instance_id)}
              aria-pressed={selected.has(c.instance_id)}
              aria-label={`select ${c.name || "card"}`}
            >
              <Card card={c} />
            </button>
          {/each}
        </div>
        <div class="prompt-foot">
          <span class="prompt-count">{selected.size} / {pickMax} selected</span>
          {#if isSearch || isCopyTarget || (isCardSetPick && pickMin === 0)}
            <button
              type="button"
              onclick={() => (selected = new Set())}
              disabled={selected.size === 0}
            >
              Clear
            </button>
          {/if}
          <button type="button" class="primary" disabled={!canSubmit} onclick={submit}>
            {#if isSacrifice}
              Sacrifice
            {:else if isSearch}
              {selected.size === 0 ? "Fail to find" : "Take"}
            {:else if isCopyTarget}
              {selected.size === 0 ? "Enter as itself" : "Enter as a copy"}
            {:else if isUntapChoice}
              Untap
            {:else if isEntryReveal}
              {selected.size === 0 ? "Reveal nothing" : "Reveal"}
            {:else if isChooseCards || isRevealPick || isPermanentPick}
              Choose
            {:else}
              Confirm
            {/if}
          </button>
        </div>
      {/if}
      {#if rejection}
        <p class="prompt-hint error prompt-rejection" role="alert">
          <span class="rejection-label">Not accepted</span>
          {rejection.message}
        </p>
      {/if}
    </div>
  </div>
{/if}

<svelte:window onkeydown={handleKey} />

<style>
  /* #624: the server's reason for refusing this prompt's answer. Below
     the footer, next to the button that was just pressed. */
  .prompt-rejection {
    margin: 0;
    padding: 8px 10px;
    border-radius: 8px;
    border: 1px solid color-mix(in srgb, var(--danger) 45%, transparent);
    background: color-mix(in srgb, var(--danger) 10%, transparent);
  }
  .rejection-label {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    font-weight: 700;
    margin-right: 6px;
  }
  .scry-lane {
    margin: 2px 0;
  }
  .lane-label {
    margin: 0 0 6px;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--fg-dim);
    font-weight: 600;
  }
  .lane-empty {
    margin: 0;
    font-size: 12px;
    color: var(--fg-dim);
  }
  .scry-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .scry-list li {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 10px;
    border-radius: 8px;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    font-size: 13px;
  }
  .scry-name {
    flex: 1;
    font-weight: 600;
    color: var(--fg);
  }
  .lane-btn {
    height: 22px;
    padding: 0 8px;
    border-radius: 6px;
    font-size: 10.5px;
  }
  .card-pick.bottomed {
    opacity: 0.45;
  }
  .pick-options {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .pick-option {
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
    text-align: left;
  }
  .pick-label {
    font-weight: 600;
  }
  .pick-cards {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(88px, 1fr));
    gap: 6px;
    width: 100%;
  }
  .card-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(96px, 1fr));
    gap: 8px;
  }
  .card-pick {
    background: transparent;
    border: 2px solid transparent;
    border-radius: var(--radius);
    padding: 3px;
    cursor: pointer;
    box-shadow: none;
    transition:
      border-color 120ms var(--ease),
      transform 120ms var(--ease),
      box-shadow 120ms var(--ease);
  }
  .card-pick:hover:not(:disabled) {
    border-color: var(--border-strong);
    background: transparent;
    transform: translateY(-2px);
  }
  .card-pick.selected {
    border-color: var(--gold);
    box-shadow: 0 0 16px rgba(217, 180, 92, 0.35);
  }
  .card-pick:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .color-row {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }
  .color-pick {
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    padding: 14px 18px;
    background: var(--fill);
    color: #0a0e1a;
    border: 1px solid rgba(0, 0, 0, 0.35);
    border-radius: 10px;
    font-weight: 800;
    cursor: pointer;
    min-width: 88px;
    box-shadow: var(--shadow-sm);
    transition:
      transform 120ms var(--ease),
      box-shadow 120ms var(--ease),
      filter 120ms var(--ease);
  }
  .color-pick:hover,
  .color-pick:focus-visible {
    background: var(--fill);
    border-color: rgba(0, 0, 0, 0.35);
    transform: translateY(-2px);
    box-shadow: var(--shadow);
    filter: brightness(1.05);
  }
  .color-letter {
    font-size: 22px;
    line-height: 1;
  }
  .color-name {
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    opacity: 0.8;
  }
  /* S26 creature-type picker. The list is the whole CR 205.3m
     vocabulary, so it scrolls inside a fixed box rather than growing
     the modal past the viewport, and the filter box takes focus on
     open because typing is how anyone finds a word in 345. */
  .type-filter {
    width: 100%;
    padding: 10px 12px;
    border-radius: 8px;
    border: 1px solid var(--line, rgba(255, 255, 255, 0.18));
    background: rgba(0, 0, 0, 0.22);
    color: inherit;
    font-size: 14px;
  }
  .type-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    max-height: 46vh;
    overflow-y: auto;
    padding: 4px 2px;
  }
  /* #804 CR 726 shortcut. One number, sitting in the button row with
     the two answers it feeds, because the question is "how many" and
     everything else about the prompt is already said above it. */
  .loop-iterations {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-right: auto;
    font-size: 13px;
    opacity: 0.85;
  }
  .loop-iterations input {
    width: 88px;
    padding: 6px 8px;
    border-radius: 8px;
    border: 1px solid var(--line, rgba(255, 255, 255, 0.18));
    background: rgba(0, 0, 0, 0.22);
    color: inherit;
    font-size: 14px;
  }
  .type-pick {
    padding: 6px 12px;
    border-radius: 999px;
    border: 1px solid var(--line, rgba(255, 255, 255, 0.18));
    background: rgba(255, 255, 255, 0.06);
    color: inherit;
    font-size: 13px;
    cursor: pointer;
  }
  .type-pick:hover,
  .type-pick:focus-visible {
    background: rgba(255, 255, 255, 0.16);
  }
  .order-label {
    display: flex;
    flex-direction: column;
    gap: 2px;
    line-height: 1.3;
  }
  .order-label strong {
    font-weight: 600;
    font-size: 13px;
  }
  .order-src {
    color: var(--fg-muted);
    font-size: 11.5px;
  }
  .assign-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: min(440px, calc(100vw - 80px));
  }
  .assign-row {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: 12px;
    padding: 8px 12px;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    border-radius: 10px;
  }
  .assign-row.trample {
    border-color: rgba(255, 107, 107, 0.35);
  }
  .assign-order {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .assign-pos {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    border-radius: 6px;
    background: var(--surface-raised);
    color: var(--fg-dim);
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 700;
  }
  .reorder-btn {
    padding: 2px 7px;
    font-size: 10px;
    border-radius: 6px;
  }
  .assign-name {
    font-weight: 600;
    font-size: 13px;
  }
  .assign-input input {
    width: 70px;
    padding: 5px 10px;
    text-align: right;
    font-family: var(--font-mono);
    font-size: 13px;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
</style>
