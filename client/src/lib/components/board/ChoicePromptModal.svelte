<script lang="ts">
  // ChoicePromptModal opens when the wire's pending_choices queue
  // has an entry addressed to the viewer.
  //
  // ADR 0111 Delivery PR 5: the small kinds are not a modal any more.
  // The yes/no family, pay_unless without picks, coin_call,
  // loop_shortcut, mana_pick, choose_color and a short option_pick /
  // entry_controller are answered INLINE in the action dock
  // (lib/choiceDock.ts): this component still owns the prompt, its
  // state, its answer and its refusal, and opens a dock request for it
  // instead of a backdrop.
  //
  // ADR 0111 Delivery PR 6: and no kind is a modal any more. Every
  // other kind (the scry family, the card grids and search, the order
  // kinds, mode_pick, the type and name pickers, pay_unless with card
  // or tap picks, a long option_pick, damage assignment) is a SHEET that
  // grows up out of the dock (DockSheet), with its confirm in the dock's
  // action bar. This component still owns the state, the answer and the
  // refusal of every kind. Generic counterpart to
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
    DivideShieldEntryView,
    DivideShieldView,
    GameView,
    PendingChoiceView,
    PickOptionView,
    ReplacementOptionView,
  } from "../../protocol";
  import Card from "./Card.svelte";
  import { isBoardAnsweredChoice } from "../../boardAnsweredChoice";
  import ModalLayer from "../ModalLayer.svelte";
  import DockRequest from "./DockRequest.svelte";
  import DockSheet from "./DockSheet.svelte";
  import { confirmAction, type DockAction } from "../../dock";
  import { onDestroy } from "svelte";
  import { choiceRequest, inlineRefusal, isInlineChoice } from "../../choiceDock";
  import {
    rejectionForPrompt,
    type ChoiceRejection,
    type ChoiceSubmission,
    type ServerErrorLike,
  } from "../../choiceRejection";
  import { colorButtons, colorPromptAnswerable } from "../../manaPick";
  import { colorPickOptions } from "../../manaSource";
  import ManaSymbolPicker from "./ManaSymbolPicker.svelte";
  import { payUnlessAnswer, waterbendLimit } from "../../waterbend";
  import {
    canPayCards,
    payCardsOptions,
    payCardsReady,
    payCardsVerb,
    togglePayCard,
  } from "../../payCards";
  import { damageSourceCaption } from "../../damageSource";
  import { freeCastRequest, mayCastKeywordsThatOpenACast } from "../../freeCastRequest";

  interface Props {
    snap: GameView;
    viewerID: string | null;
    sendAction: (type: ActionType, params?: unknown, player?: string) => void;
    // GameClient.lastError. The board's toast for it sits under this
    // modal's backdrop, so a refusal of this prompt's answer is shown
    // here instead (#624). Optional so a caller with no error feed
    // still mounts the modal.
    lastError?: ServerErrorLike | null;
    // ADR 0111 PR 5: the action dock is on screen (Game.svelte's
    // dockShown). An inline kind is drawn there, so without a dock (the
    // dev replay scrubber's past frame) it is not drawn at all.
    docked?: boolean;
  }

  const { snap, viewerID, sendAction, lastError = null, docked = true }: Props = $props();

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
      // #1623: the legend rule and choose_protector are board-answered
      // too (Board.svelte drives them through the targeting banner);
      // the one shared list says so, for both sides.
      if (isBoardAnsweredChoice(c.kind)) continue;
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
  // ADR 0111 PR 5: answered in the action dock, not in this modal.
  const inline = $derived(isInlineChoice(active));

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
      inlineRefusal.set(null);
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
    if (hit) {
      rejection = hit;
      // ADR 0111 PR 5 / PR 6: an inline prompt or a sheet shows the
      // refusal in the dock, beside the board, so the strip's toast
      // stands down for it.
      if (docked) inlineRefusal.set(lastError);
    }
  });
  onDestroy(() => inlineRefusal.set(null));

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

  // ADR 0098 — the same entry pause with a card that is SPENT. Mox
  // Diamond's "you may discard a land card instead" (floor zero: showing
  // nothing puts the Mox into its owner's graveyard, so the decline
  // button says so) and Heart of Yavimaya's "sacrifice a Forest instead"
  // (floor = ceiling: not a "may").
  const isEntryDiscard = $derived(active?.kind === "entry_discard_from_hand");
  const isEntrySacrifice = $derived(active?.kind === "entry_sacrifice");

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
  // ADR 0107 §6, CR 609.7a: "a source of your choice". One card, from
  // several zones, so each candidate is captioned with whose it is and
  // where it is (damageSource.ts).
  const isChooseSource = $derived(active?.kind === "choose_source");
  // The kinds that share the bounded card-set grid.
  const isCardSetPick = $derived(
    isChooseCards ||
      isUntapChoice ||
      isEntryReveal ||
      isEntryDiscard ||
      isEntrySacrifice ||
      isRevealPick ||
      isPermanentPick ||
      isChooseSource,
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
  // nothing here sorts. The words are lib/choiceDock.ts's now (PR 5).

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

  // S17 sub-PR 6 optional-replacement branch — "may"
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

  // #1311: a pay-unless whose payment is a WATERBEND cost ("Ward—
  // Waterbend {4}", The Unagi of Kyoshi Island) ships tap_cost: the
  // chooser may tap untapped artifacts and creatures they control,
  // each paying {1} of the generic (CR 701.67a), and pays the rest
  // with mana. The picks ride the "Pay" answer as tap_ids. Optional —
  // tapping none pays it all with mana, as an ordinary ward would.
  let payTaps = $state<string[]>([]);
  let payTapsFor: string | null = null;
  $effect(() => {
    const id = active?.id ?? null;
    if (id !== payTapsFor) {
      payTapsFor = id;
      payTaps = [];
      payCardPicks = [];
    }
  });
  // ADR 0108 §5: a pay-unless whose payment discards or sacrifices
  // ("Echo—Discard a card", "Cumulative upkeep—Sacrifice a land")
  // ships pay_cards; the "Pay" answer names exactly `count` of them.
  let payCardPicks = $state<string[]>([]);
  const payCards = $derived(isPayUnless ? (active?.pay_cards ?? null) : null);
  const payCardOptions = $derived.by((): CardView[] => {
    if (!payCards) return [];
    return payCardsOptions(
      payCards,
      snap.battlefield?.cards,
      snap.seats?.find((s) => s.id === viewerID),
    );
  });
  const payBlocked = $derived(payCards !== null && !payCardsReady(payCards, payCardPicks));
  const payTapCost = $derived(isPayUnless ? (active?.tap_cost ?? null) : null);
  const payTapLimit = $derived(payTapCost ? waterbendLimit(payTapCost, undefined) : 0);
  const payTapOptions = $derived.by((): CardView[] => {
    if (!payTapCost) return [];
    const ids = new Set(payTapCost.options?.cards ?? []);
    return snap.battlefield.cards.filter((c) => ids.has(c.instance_id));
  });
  function togglePayTap(id: string): void {
    if (payTaps.includes(id)) {
      payTaps = payTaps.filter((c) => c !== id);
    } else if (payTaps.length < payTapLimit) {
      payTaps = [...payTaps, id];
    }
  }

  // S28 cascade branch — "you may cast it without paying its mana
  // cost". Same {choice_id, apply} payload; the server routes to
  // ResolveMayCast by kind. "Yes" stamps a free-cast permission on
  // the exiled card, which then casts out of exile like any other
  // impulse grant; "No" puts it on the bottom of the library with
  // the rest of the cards cascade turned over.
  //
  // ADR 0099 §7: the same prompt serves discover, suspend and madness,
  // each of which does something different on "no". The server names
  // the rule (may_cast_keyword) and the branches; mayCastCopy words
  // them. The offered card is in exile face up, found by may_cast_card.
  const isMayCast = $derived(active?.kind === "may_cast");
  const mayCastCard = $derived(
    active?.may_cast_card
      ? snap.exile?.cards?.find((c) => c.instance_id === active.may_cast_card)
      : active?.options?.[0],
  );

  // Shockland entry branch — "as this land enters, you may pay 2
  // life. If you don't, it enters tapped." Same {choice_id, apply}
  // payload; the server routes by kind. The permanent is still in
  // hand while this is open, which is the point: the answer decides
  // how it ENTERS, so there is no tapped land on the board to look
  // at yet and the prompt has to say the card's name itself.
  const isEntryPayLife = $derived(active?.kind === "entry_pay_life");

  // ADR 0109 §10 riot — "a +1/+1 counter, or haste?", asked before the
  // permanent enters. The same {choice_id, apply} payload: apply takes
  // the counter. Answered from the keyboard with C and H rather than Y
  // and N, because neither answer is a "no".
  const isEntryRiot = $derived(active?.kind === "entry_riot");

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

  // ADR 0102 entry_controller — "enters under the control of an
  // opponent of your choice" (CR 614.12a). The same seat buttons and
  // the same {option_index} answer as option_pick; what differs is the
  // sentence. The permanent is not on the battlefield yet, and nothing
  // can happen until an opponent is named.
  const isEntryController = $derived(active?.kind === "entry_controller");
  const entryControllerHint = $derived(
    active?.control_purpose === "benefit"
      ? "Whoever you choose will control it and get what it does."
      : "Whoever you choose will control it and live with what it does.",
  );

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
  // ADR 0097 (#1749): every bullet in printed order, the ones this
  // object's ability has already chosen marked used. The used ones
  // ride beside the offer (`mode_used_*`), so they are merged in here
  // by index and rendered disabled — never sent.
  const modeRows = $derived.by(() => {
    const rows: { idx: number; label: string; used: boolean }[] = modeLabels.map((label, i) => ({
      idx: modeIndexes[i] ?? i,
      label,
      used: false,
    }));
    const usedLabels = active?.mode_used_options ?? [];
    (active?.mode_used_indexes ?? []).forEach((idx, i) => {
      rows.push({ idx, label: usedLabels[i] ?? "", used: true });
    });
    return rows.sort((a, b) => a.idx - b.idx);
  });
  const modeUsedNote = $derived(
    active?.mode_not_chosen === "this_turn" ? "already chosen this turn" : "already chosen",
  );
  // The chosen bullets IN THE ORDER CHOSEN — CR 608.2c resolves them
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

  // #804 loop_shortcut — CR 732. The loop breaker has fired and this
  // viewer controls the ability that is repeating, so they get the
  // question paper asks: how many more times? A number, not a yes/no,
  // because that is what CR 732 lets a player propose — and 0 is a
  // real answer ("stop here"), which leaves the table paused exactly
  // where the breaker put it, banner and all.
  const isLoopShortcut = $derived(active?.kind === "loop_shortcut");
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

  // #1298: the put_in_library top lane's refinements. `topCount` is
  // EXACTLY how many cards stay on top (Cream of the Crop: one), and
  // Done waits for it; `topDepth` is where that lane lands (Temporal
  // Cleansing: second from the top), which only changes its name.
  const topCount = $derived(isPutInLibrary ? (active?.top_count ?? 0) : 0);
  const topDepth = $derived(isPutInLibrary ? (active?.top_depth ?? 0) : 0);
  const topLaneLabel = $derived(topDepth > 1 ? `${ordinal(topDepth)} from the top` : "On top");

  // Whose library the cards go back to, for the hint's wording: Jace's
  // +2 and Portent order ANOTHER player's library, so "your next draw"
  // would be wrong there.
  const ownLibrary = $derived(
    (active?.options ?? []).every((c) => !c.owner || !viewerID || c.owner === viewerID),
  );
  const nextDraw = $derived(ownLibrary ? "your next draw" : "that player's next draw");

  function ordinal(n: number): string {
    const words = ["", "", "Second", "Third", "Fourth", "Fifth", "Sixth", "Seventh"];
    return words[n] ?? `#${n}`;
  }

  let scryTop = $state<string[]>([]);
  let scryBottom = $state<string[]>([]);

  // An exact top count holds Done until the top lane has that many.
  const canSubmitLookAtTop = $derived(topCount === 0 || scryTop.length === topCount);

  // Seed the default whenever a scry / surveil prompt opens:
  // everything stays on top, in the order the server listed it (which
  // is current library order) — or, with an exact top count, the first
  // that many stay and the rest start on the bottom, so the seed is
  // already an answer the server accepts.
  let lastScryID: string | null = null;
  $effect(() => {
    if (!isLookAtTop || !active) {
      lastScryID = null;
      return;
    }
    if (active.id === lastScryID) return;
    lastScryID = active.id;
    const all = (active.options ?? []).map((c) => c.instance_id);
    const keep = isBottomOnly ? 0 : topCount > 0 ? topCount : all.length;
    scryTop = all.slice(0, keep);
    scryBottom = all.slice(keep);
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
    if (!canSubmitLookAtTop) return;
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
  // which reads as a bug. Warn the chooser and relabel "Yes". (The
  // warning and a doubled trigger's note are lib/choiceDock.ts's now.)

  // Y / N answer the yes-no prompts (optional replacement, may-
  // trigger, pay-unless) from the keyboard; each dock button shows its
  // key in a cap (the sheet's footer, for pay_unless with picks).
  // Ignored while typing in a field.
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
    if (isEntryRiot) {
      if (e.key === "c" || e.key === "C") {
        e.preventDefault();
        answerOptional(true);
      } else if (e.key === "h" || e.key === "H") {
        e.preventDefault();
        answerOptional(false);
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
    // #1311: a waterbend pay-unless names its taps beside the apply.
    if (isPayUnless) {
      // ADR 0108 §5: a discard or sacrifice payment is not a "Pay"
      // until the picks add up to the count.
      if (apply && payBlocked) return;
      answer(payUnlessAnswer(active, apply, payTaps, payCardPicks));
      return;
    }
    // ADR 0099 §7: "Cast it free" hands the card to Board's cast chain
    // once the snapshot carrying the grant arrives.
    if (
      apply &&
      isMayCast &&
      active.may_cast_card &&
      mayCastKeywordsThatOpenACast.has(active.may_cast_keyword ?? "")
    ) {
      freeCastRequest.set(active.may_cast_card);
    }
    answer({ apply });
  }

  // ADR 0111 PR 5: the open prompt as a dock request, when it is one of
  // the inline kinds. The dock draws it; this component keeps the
  // answer and its refusal.
  const inlineRequest = $derived(
    active && inline
      ? choiceRequest(
          active,
          {
            sourceName: triggerSourceName(active.source),
            mayCastCardName: mayCastCard?.name || undefined,
            loopIterations,
            loopAnswerable,
            body: isManaPick
              ? manaBody
              : isColorChoice
                ? colorBody
                : isLoopShortcut
                  ? loopBody
                  : undefined,
            rejection: rejection?.message ?? null,
          },
          {
            onAnswer: answerOptional,
            onCoin: answerCoin,
            onOption: answerOptionPick,
            onLoop: submitLoopShortcut,
          },
        )
      : null,
  );

  // sourceCardName resolves the source-card display name for a
  // trigger prompt. Walks battlefield + every seated player's
  // graveyard + the shared exile zone — LTB triggers prompt after
  // the source has already moved off the battlefield, so the
  // lookup has to span destination zones. Falls back to a generic
  // string when the card isn't visible to the viewer (redacted
  // CardView entries arrive with empty names).
  // ADR 0098: the name of the card whose entry an entry prompt is
  // pausing. It is not on the battlefield yet — it is on the stack (a
  // spell resolving), in a hand, a library, a graveyard or exile — so
  // triggerSourceName's search does not reach it. Falls back to "it".
  function enteringCardName(sourceID?: string): string {
    if (!sourceID) return "it";
    const seek = (cards: CardView[] | undefined) =>
      cards?.find((c) => c.instance_id === sourceID && c.name)?.name;
    const zones: (CardView[] | undefined)[] = [
      snap.stack?.cards,
      snap.battlefield?.cards,
      snap.exile?.cards,
    ];
    for (const p of snap.seats ?? []) {
      zones.push(p.hand?.cards, p.graveyard?.cards);
    }
    for (const z of zones) {
      const name = seek(z);
      if (name) return name;
    }
    return "it";
  }

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

  // ADR 0108 §7 divide_shield — CR 615.7: a charged shield ("prevent
  // the next 3 damage") meets more damage at once than it can cover,
  // and the protected player chooses which of it the shield prevents.
  // One number per damage event, 0..its amount, adding up to the charge.
  const isDivideShield = $derived(active?.kind === "divide_shield");
  const divideFrame = $derived<DivideShieldView | null>(active?.divide_shield ?? null);
  let divideShares = $state<Record<string, number>>({});

  // Reset on a new prompt, the damage-assignment branch's last-id
  // pattern.
  let lastDivideChoiceID: string | null = null;
  $effect(() => {
    const nextID = active?.id ?? null;
    if (nextID === lastDivideChoiceID) return;
    lastDivideChoiceID = nextID;
    if (!divideFrame) return;
    const next: Record<string, number> = {};
    for (const e of divideFrame.entries) next[e.id] = 0;
    divideShares = next;
  });

  const dividedTotal = $derived(
    divideFrame ? divideFrame.entries.reduce((acc, e) => acc + (divideShares[e.id] ?? 0), 0) : 0,
  );
  const canSubmitDivide = $derived(divideFrame !== null && dividedTotal === divideFrame.charge);

  function setDivideShare(id: string, raw: string, max: number): void {
    const n = Math.min(max, Math.max(0, Math.floor(Number(raw) || 0)));
    divideShares = { ...divideShares, [id]: n };
  }

  function divideTargetName(e: DivideShieldEntryView): string {
    if (e.target_id === viewerID) return "you";
    return e.target_name || e.target_id.slice(0, 8);
  }

  function submitDivideShield(): void {
    if (!active || !viewerID || !divideFrame || !canSubmitDivide) return;
    const distribution: Record<string, number> = {};
    for (const e of divideFrame.entries) distribution[e.id] = divideShares[e.id] ?? 0;
    answer({ distribution });
  }

  // ADR 0111 PR 6: every kind that is not inline is a sheet in the
  // action dock. This is its dialog name (the modal's heading, without
  // its aria-hidden source tag), the tag, the running count, and the
  // action bar's buttons. The body is the template's.
  //
  // Keys (ADR 0111 §1): Enter presses a confirm that commits what the
  // player picked (Done, Take, Choose, Deal damage, an order). It never
  // presses a decline — an empty pick where nothing is a legal answer
  // ("Fail to find", "Enter as itself", "Reveal nothing", "Don't
  // discard …") — nor a payment ("Pay {2}", which keeps Y / N, as the
  // inline pay_unless does). A pending choice has no Escape: the game is
  // waiting on an answer, not on a way out.
  interface SheetSpec {
    label: string;
    src?: string;
    count?: string;
    width: number;
    primary: DockAction | null;
    secondary: DockAction[];
  }
  const sheet = $derived.by((): SheetSpec | null => {
    const c = active;
    if (!c || inline) return null;
    const s = (n: number | undefined) => (n === 1 ? "" : "s");
    if (isLookAtTop) {
      return {
        label:
          c.reason ||
          (isPutInLibrary
            ? "Put them in your library"
            : isSurveil
              ? "Surveil"
              : isReorderOnly
                ? "Look at the top"
                : "Scry"),
        src: isPutInLibrary
          ? "CR 401.4"
          : isReorderOnly
            ? undefined
            : isSurveil
              ? "CR 701.25"
              : "CR 701.22",
        count: isBottomOnly
          ? `${scryBottom.length} on the bottom`
          : isReorderOnly
            ? `${scryTop.length} ${isPutInLibrary ? "on top" : "back on top"}`
            : `${scryTop.length}${topCount > 0 ? ` of ${topCount}` : ""} ${
                topDepth > 1 ? topLaneLabel.toLowerCase() : "on top"
              } · ${scryBottom.length} ${isSurveil ? "in the graveyard" : "on the bottom"}`,
        width: 720,
        primary: confirmAction("Done", submitScry, {
          disabled: !canSubmitLookAtTop,
          title: canSubmitLookAtTop ? undefined : `Keep exactly ${topCount} on top`,
        }),
        secondary: [],
      };
    }
    if (isCreatureTypePick) {
      // A click on a type is the answer (and Enter in the filter takes
      // the one best match): no bar.
      return {
        label: c.reason || "Choose a creature type",
        src: "as this enters · CR 614.12",
        width: 560,
        primary: null,
        secondary: [],
      };
    }
    if (isCardNamePick) {
      const typed = nameFilter.trim();
      return {
        label: c.reason || "Choose a card name",
        src: "as this enters · CR 614.12",
        width: 560,
        primary: confirmAction(`Name “${typed || "…"}”`, () => pickCardName(nameFilter), {
          disabled: typed === "",
        }),
        secondary: [],
      };
    }
    if (isOptionPick) {
      return {
        label: c.reason || "Choose one",
        src: "choose one · CR 608.2",
        width: 560,
        primary: null,
        secondary: [],
      };
    }
    if (isEntryController) {
      return {
        label: c.reason || "Choose an opponent",
        src: "choose an opponent · CR 614.12a",
        width: 560,
        primary: null,
        secondary: [],
      };
    }
    if (isModePick) {
      return {
        label: triggerSourceName(c.source),
        src: c.reason || "choose one",
        count: `${modePicks.length} / ${modeMax > 0 ? modeMax : "any"}`,
        width: 560,
        primary: confirmAction("Choose", answerModes, { disabled: !canConfirmModes }),
        secondary: [],
      };
    }
    if (isPayUnless) {
      // Y / N, as the inline pay_unless: a payment is not one stray
      // Enter away.
      return {
        label: c.reason || `${triggerSourceName(c.source)} — pay ${c.pay_cost ?? ""}?`,
        src: "pay unless",
        width: 560,
        primary: {
          id: "yes",
          label: `Pay ${c.pay_cost ?? ""}`.trim(),
          disabled: payBlocked,
          keyShortcuts: "Y",
          cap: "Y",
          onPress: () => answerOptional(true),
        },
        secondary: [
          {
            id: "no",
            label: "Don't pay",
            keyShortcuts: "N",
            cap: "N",
            onPress: () => answerOptional(false),
          },
        ],
      };
    }
    if (isDamageAssignment && damageFrame) {
      return {
        label: c.reason || "Assign combat damage",
        src: "CR 510.1c",
        count: `${assignedTotal} / ${damageFrame.attacker_power} assigned`,
        width: 560,
        primary: confirmAction("Deal damage", submitDamageAssignment, {
          disabled: !canSubmitAssignment,
        }),
        secondary: [],
      };
    }
    if (isDivideShield && divideFrame) {
      return {
        label: c.reason || `Divide ${divideFrame.label ?? "the shield"}`,
        src: "CR 615.7",
        count: `${dividedTotal} / ${divideFrame.charge} prevented`,
        width: 560,
        primary: confirmAction("Prevent", submitDivideShield, {
          disabled: !canSubmitDivide,
          title: canSubmitDivide ? undefined : `Prevent exactly ${divideFrame.charge}`,
        }),
        secondary: [],
      };
    }
    if (isReplacementOrder || isTriggerOrder) {
      return {
        label: c.reason || (isTriggerOrder ? "Order your triggers" : "Order replacement effects"),
        src: isTriggerOrder ? "CR 603.3b" : "CR 616",
        count: `${ordered.length} / ${replacementOptions.length} ordered`,
        width: 560,
        primary: confirmAction(
          isTriggerOrder ? "Resolve in this order" : "Apply in this order",
          submitReplacementOrder,
          { disabled: ordered.length !== replacementOptions.length },
        ),
        secondary: [],
      };
    }
    // The card grid.
    const none = selected.size === 0;
    const [label, src] = isSacrifice
      ? [c.reason || "Sacrifice a permanent", "sacrifice"]
      : isSearch
        ? [c.reason || "Search your library", "search · CR 701.23"]
        : isCopyTarget
          ? [c.reason || "Enter as a copy of…", "copy · CR 707"]
          : isUntapChoice
            ? [c.reason || "Untap step — choose which permanents untap", "untap · CR 502.3"]
            : isEntryReveal
              ? [c.reason || "Reveal a card from your hand?", "reveal · CR 614"]
              : isEntryDiscard
                ? [c.reason || "Discard a card so it enters?", "discard · CR 614"]
                : isEntrySacrifice
                  ? [c.reason || "Sacrifice so it enters", "sacrifice · CR 614"]
                  : isChooseCards
                    ? [c.reason || "Choose cards", "choose"]
                    : isRevealPick
                      ? [c.reason || "Choose from the revealed cards", "reveal · CR 701.20"]
                      : isPermanentPick
                        ? [c.reason || "Choose permanents", "choose · CR 608.2"]
                        : isChooseSource
                          ? [c.reason || "Choose a source of damage", "source · CR 609.7a"]
                          : [
                              `${c.reason || "Choose"} — pick ${c.count} card${s(c.count)}`,
                              isSelfSource ? "discard" : "reveal",
                            ];
    const verb = isSacrifice
      ? "Sacrifice"
      : isSearch
        ? none
          ? "Fail to find"
          : "Take"
        : isCopyTarget
          ? none
            ? "Enter as itself"
            : "Enter as a copy"
          : isUntapChoice
            ? "Untap"
            : isEntryReveal
              ? none
                ? "Reveal nothing"
                : "Reveal"
              : isEntryDiscard
                ? none
                  ? `Don't discard — put ${enteringCardName(c.source)} into its owner's graveyard`
                  : "Discard"
                : isEntrySacrifice
                  ? "Sacrifice"
                  : isChooseSource
                    ? "Choose this source"
                    : isChooseCards || isRevealPick || isPermanentPick
                      ? "Choose"
                      : "Confirm";
    const clearable = isSearch || isCopyTarget || (isCardSetPick && pickMin === 0);
    return {
      label,
      src,
      count: `${selected.size} / ${pickMax} selected`,
      width: 720,
      // An empty pick where empty is legal is a decline: no Enter.
      primary: confirmAction(verb, submit, {
        disabled: !canSubmit,
        enter: !(none && pickMin === 0),
      }),
      secondary: clearable
        ? [
            {
              id: "clear",
              label: "Clear",
              disabled: none,
              onPress: () => (selected = new Set()),
            },
          ]
        : [],
    };
  });
</script>

<!-- ADR 0111 PR 5: the inline kinds' bodies, drawn inside the dock's
     prompt area. -->
{#snippet manaBody()}
  <!-- #1438: the same symbol picker a click on a multi-ability source
       opens. No onCancel: the source is already tapped when the server
       asks. Keys 1-9 pick. -->
  <div class="dock-mana">
    <ManaSymbolPicker
      options={colorPickOptions(buttons, "add")}
      onPick={(o) => o.color && pickColor(o.color)}
      label="mana colors"
    />
  </div>
{/snippet}
{#snippet colorBody()}
  <div class="dock-mana">
    <ManaSymbolPicker
      options={colorPickOptions(buttons, "choose")}
      onPick={(o) => o.color && pickColor(o.color)}
      label="colors"
    />
  </div>
{/snippet}
{#snippet loopBody()}
  <label class="loop-iterations">
    <span>More times</span>
    <input
      type="number"
      min="0"
      max={loopMax}
      step="1"
      bind:value={loopIterations}
      aria-label="How many more times to resolve it"
      onkeydown={(e) => {
        // The number typed is the answer: Enter in the field sends it.
        // (The dock's own Enter stands down while a field has focus.)
        if (e.key !== "Enter" || !loopAnswerable) return;
        e.preventDefault();
        submitLoopShortcut(loopIterations);
      }}
    />
  </label>
{/snippet}

{#if open && active && inline}
  {#if docked && inlineRequest}
    <!-- Not a modal: it blurs and blocks nothing. The layer is for the
         keys it owns (Y / N, H / T / S, 1-9), so the global shortcuts
         stand down while it is open, as they did (ADR 0111 §2). -->
    <ModalLayer />
    <DockRequest request={inlineRequest} />
  {/if}
{:else if open && active && docked && sheet}
  <!-- ADR 0111 PR 6: every other kind is a sheet that grows up out of
       the action dock. Its confirm and any secondary are the dock's
       action bar (sheet.primary / sheet.secondary); its body is here. -->
  <DockSheet
    rank="choice"
    label={sheet.label}
    src={sheet.src}
    count={sheet.count}
    width={sheet.width}
    sheetKey={active.id}
    primary={sheet.primary}
    secondary={sheet.secondary}
    refusal={rejection
      ? { tag: "Not accepted", text: rejection.message, actions: [], tone: "danger" }
      : null}
  >
    {#if isLookAtTop}
      <p class="prompt-hint">
        {#if isBottomOnly}
          Put them on the bottom of the library in any order — the last one listed is the very
          bottom card.
        {:else if isReorderOnly && topDepth > 1}
          Put them in any order, the first {topLaneLabel.toLowerCase()}.
        {:else if isReorderOnly}
          Put them {isPutInLibrary ? "on top" : "back"} in any order — the topmost is {nextDraw}.
        {:else if topCount > 0}
          Keep exactly {topCount} on top — the topmost is {nextDraw} — and put the rest on the bottom
          in any order.
        {:else if scryTop.length + scryBottom.length === 1 && topDepth > 1}
          Put it {topLaneLabel.toLowerCase()}, or on the bottom of {ownLibrary
            ? "your library"
            : "its owner's library"}.
        {:else if scryTop.length + scryBottom.length === 1}
          Keep it on top, or put it {isSurveil
            ? "into your graveyard"
            : ownLibrary
              ? "on the bottom of your library"
              : "on the bottom of its owner's library"}.
        {:else}
          Keep any of these on top — the topmost is {nextDraw} — and put the rest {isSurveil
            ? "into your graveyard"
            : "on the bottom"}.
        {/if}
        {isPutInLibrary ? "Nobody else learns the order." : "Only you can see them."}
      </p>
      {#if !isBottomOnly}
        <div class="scry-lane">
          <h3 class="lane-label">
            {topLaneLabel} ({scryTop.length}{topCount > 0 ? ` of ${topCount}` : ""})
          </h3>
          {#if scryTop.length === 0}
            <p class="lane-empty">
              {ownLibrary ? "Nothing — your next draw comes from under these." : "Nothing."}
            </p>
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
                      aria-label={`keep ${scryCardName(id)} on top`}
                      >{topDepth > 1 ? topLaneLabel : "Keep on top"}</button
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
    {:else if isCreatureTypePick}
      <p class="prompt-hint">
        The choice is locked in for as long as this permanent stays on the battlefield.
      </p>
      <input
        class="type-filter"
        type="text"
        data-sheet-focus
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
      <p class="prompt-hint">
        Any card name is legal — type one. The suggestions are the cards everyone can currently see.
      </p>
      <input
        class="type-filter"
        type="text"
        data-sheet-focus
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
    {:else if isOptionPick}
      <p class="prompt-hint">
        Someone else's spell or ability is asking you. Every option listed is one you can take, and
        the game waits until you pick one.
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
    {:else if isEntryController}
      <p class="prompt-hint">
        It hasn't entered yet: it enters under the control of the opponent you choose.
        {entryControllerHint}
      </p>
      <ul class="pick-options">
        {#each pickOptions as opt, i (i)}
          <li>
            <button type="button" class="pick-option" onclick={() => answerOptionPick(i)}>
              <span class="pick-label">{opt.label}</span>
            </button>
          </li>
        {/each}
      </ul>
    {:else if isModePick}
      <p class="prompt-hint">
        The ability is not on the stack until you answer — its mode is chosen as it goes there (CR
        603.3c), and any targets it asks for come after.
        {#if modeRepeatable}
          You may choose the same mode more than once.
        {/if}
      </p>
      <ul class="prompt-options" role={modeSingle ? "radiogroup" : "group"}>
        {#each modeRows as row (row.idx)}
          {@const idx = row.idx}
          {@const times = modeTimes(idx)}
          <li>
            <button
              type="button"
              class="prompt-opt"
              class:on={times > 0}
              role={modeSingle ? "radio" : "checkbox"}
              aria-checked={times > 0}
              aria-disabled={row.used}
              disabled={row.used}
              onclick={() => {
                if (!row.used) toggleMode(idx);
              }}
            >
              <span class="prompt-radio" aria-hidden="true"></span>
              <span class="mode-label">{row.label}</span>
              {#if row.used}
                <span class="note">{modeUsedNote}</span>
              {/if}
              {#if modeRepeatable && times > 0}
                <span class="mode-times">&times;{times}</span>
              {/if}
            </button>
          </li>
        {/each}
      </ul>
    {:else if isPayUnless}
      {#if payCards}
        <p class="prompt-hint">
          Pay — {active.pay_cost ?? "the cost"} — or don't and let {triggerSourceName(
            active.source,
          )} do its thing.
        </p>
        {#if !canPayCards(payCards)}
          <p class="prompt-hint">You don't have enough to pay this.</p>
        {:else}
          <p class="prompt-hint sub">{payCardsVerb(payCards)}</p>
          <ul class="prompt-options">
            {#each payCardOptions as c (c.instance_id)}
              <li>
                <button
                  type="button"
                  class="prompt-opt"
                  class:on={payCardPicks.includes(c.instance_id)}
                  disabled={payCardPicks.length >= payCards.count &&
                    !payCardPicks.includes(c.instance_id)}
                  aria-pressed={payCardPicks.includes(c.instance_id)}
                  onclick={() =>
                    (payCardPicks = togglePayCard(payCards, payCardPicks, c.instance_id))}
                >
                  <span class="prompt-radio" aria-hidden="true"></span>
                  <span class="name">{c.name}</span>
                </button>
              </li>
            {/each}
          </ul>
          <p class="prompt-hint sub">{payCardPicks.length} / {payCards.count} chosen</p>
        {/if}
      {:else}
        <!-- A waterbend tap list without card picks (#1311). -->
        <p class="prompt-hint">
          Pay {active.pay_cost ?? "the cost"} from your pool (untapped sources auto-tap if it's short),
          or don't and let {triggerSourceName(active.source)} do its thing.
        </p>
      {/if}
      {#if payTapCost}
        <p class="prompt-hint sub">
          {payTapCost.label ?? "Waterbend"}: tap your untapped artifacts and creatures to help —
          each pays for {"{1}"}. Mana pays the rest.
        </p>
        {#if payTapOptions.length === 0}
          <p class="prompt-hint">You control nothing untapped that can help.</p>
        {:else}
          <ul class="prompt-options">
            {#each payTapOptions as c (c.instance_id)}
              <li>
                <button
                  type="button"
                  class="prompt-opt"
                  class:on={payTaps.includes(c.instance_id)}
                  disabled={payTaps.length >= payTapLimit && !payTaps.includes(c.instance_id)}
                  aria-pressed={payTaps.includes(c.instance_id)}
                  onclick={() => togglePayTap(c.instance_id)}
                >
                  <span class="prompt-radio" aria-hidden="true"></span>
                  <span class="name">{c.name}</span>
                </button>
              </li>
            {/each}
          </ul>
          <p class="prompt-hint sub">{payTaps.length} / {payTapLimit} tapped</p>
        {/if}
      {/if}
    {:else if isDamageAssignment && damageFrame}
      {#if damageFrame.blocker_divides}
        <!-- #1706, CR 510.1d: a blocker dividing its damage among the
               attackers it blocks — any split that adds up. -->
        <p class="prompt-hint">
          <strong>{attackerName(damageFrame.attacker_card_id)}</strong>
          is blocking {damageFrame.blocker_card_ids.length} creatures. Divide its
          {damageFrame.attacker_power} damage among them however you like.
        </p>
      {:else}
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
      {/if}
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
    {:else if isDivideShield && divideFrame}
      <!-- ADR 0108 §7, CR 615.7: which of this damage the shield
           prevents. Each row is one damage event; the shares add up to
           the shield's charge. -->
      <p class="prompt-hint">
        {divideFrame.label ?? "The shield"} can prevent {divideFrame.charge} of this damage, all dealt
        at the same time. Choose how much of each to prevent.
      </p>
      <ul class="assign-list">
        {#each divideFrame.entries as e (e.id)}
          <li class="assign-row">
            <span class="assign-name">
              {e.source_name || "A source"} → {divideTargetName(e)}: {e.amount}{e.combat
                ? " combat"
                : ""} damage
            </span>
            <label class="assign-input">
              <span class="sr-only">damage to prevent of {e.amount} to {divideTargetName(e)}</span>
              <input
                type="number"
                min="0"
                max={e.amount}
                value={divideShares[e.id] ?? 0}
                oninput={(ev) =>
                  setDivideShare(e.id, (ev.currentTarget as HTMLInputElement).value, e.amount)}
              />
            </label>
          </li>
        {/each}
      </ul>
    {:else if isReplacementOrder || isTriggerOrder}
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
    {:else}
      <p class="prompt-hint">
        {#if isSacrifice}
          Choose {active.count === 1 ? "a permanent" : `${active.count} permanents`} you control to sacrifice.
          This isn't optional — {triggerSourceName(active.source)} is making you.
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
          untapped. Revealing costs nothing — the card stays in your hand — but everyone gets to see it,
          and you may show nothing instead.
        {:else if isEntryDiscard}
          Discard one of these and {enteringCardName(active.source)} enters the battlefield. Discard nothing
          and it goes to its owner's graveyard instead — it never enters.
        {:else if isEntrySacrifice}
          Sacrifice {pickMax === 1 ? "one of these" : `${pickMax} of these`} and {enteringCardName(
            active.source,
          )} enters the battlefield. This isn't optional.
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
        {:else if isChooseSource}
          Pick the source whose next damage this prevents. It isn't targeted, so anything listed can
          be chosen — a permanent, a spell on the stack, or a card that something on the stack still
          refers to. If the card names a kind of source, that is checked again when the damage would
          be dealt.
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
            {#if isChooseSource}
              <span class="source-caption">{damageSourceCaption(snap, c, viewerID)}</span>
            {/if}
          </button>
        {/each}
      </div>
    {/if}
  </DockSheet>
{/if}

<svelte:window onkeydown={handleKey} />

<style>
  /* ADR 0111 PR 5: the mana symbols in the dock's prompt area. Five
     colours fit one row of the 300-380px dock; a sixth wraps. */
  .dock-mana :global(.mana-picker) {
    gap: 5px;
  }
  .dock-mana :global(.mana-option) {
    flex: 1 1 0;
    min-width: 50px;
    max-width: 72px;
    min-height: 64px;
    padding: 8px 2px 6px;
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
  /* ADR 0107 §6: whose a candidate source is, and where. */
  .source-caption {
    display: block;
    margin-top: 4px;
    font-size: 11px;
    line-height: 1.3;
    color: var(--fg-muted);
    text-align: center;
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
  /* #804 CR 732 shortcut. One number, sitting in the button row with
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
