<script lang="ts">
  // TableEffects — the generic effect cues (ADR 0209 §1 and §3, #4396).
  //
  // When a permanent is destroyed, exiled, sacrificed or bounced, a spell
  // is countered or fizzles (#4450), a planeswalker arrives or goes, a
  // token is created, or a board wipe clears three or more permanents at
  // once, this layer draws one cue over the table and plays its sound, so
  // a player sees what
  // happened before they read the log. The board already shows the frame's end state: this is
  // presentation over it, and nothing waits for it (ADR 0209 §1: no input,
  // priority pass or bot seat ever does).
  //
  // A permanent that has left has no tile, so what moves is an ART-ONLY
  // COPY drawn here, from the last box measured while it was on the
  // battlefield (a box cached on a board of another size is never used).
  // The copy is mounted the moment the frame arrives, so the card never
  // blinks out; a board wipe's per-card cues hold the copy still until
  // the sweep reaches them.
  //
  // A created token (#4441) is the exception: it is on the board, so its
  // pop-in moves its own tile (effectCueMotion.ts tokenPop, presentation
  // only, cancelled with the cue) and this layer adds the glow over it.
  // Counters put on a permanent (#4442) are on the board too: their pulse
  // and tick are drawn over its power/toughness badge, and the tile is
  // left alone.
  //
  // A commander (#4443) cast from the command zone gets golden light from
  // where it sat to the spell on the stack (CastFlight flies the card);
  // one put back into the command zone drifts home as a copy with a soft
  // trail, from wherever it was, while its card there waits hidden until
  // the copy lands (effectCueMotion.ts commandArrive, cancelled with the
  // cue).
  //
  // An Equipment attached or an Aura enchanting a permanent (#4444) is on
  // the board too, tucked under its host: its own tile slides there from
  // where it was (effectCueMotion.ts attachSlide, cancelled with the cue),
  // and this layer adds the clasp's spark (Equipment) or a soft ring of
  // light (Aura) over the host.
  //
  // Cards milled or discarded into a graveyard (#4445) have left a library
  // or a hand the board draws as a pile or a fan, so what moves is a copy
  // again: each card slides from the owner's library pile (turning face up
  // on the way) or is tossed from their hand onto their graveyard pile, a
  // stagger apart, and five or more go as one sweep.
  //
  // A card returned from a graveyard (#4446) is the mirror image: a copy
  // rises out of its owner's graveyard pile in a dark-green glow and goes
  // to where it went, its tile on the battlefield (which waits hidden until
  // the copy lands, effectCueMotion.ts riseArrive, cancelled with the cue)
  // or its owner's hand.
  //
  // A library searched (#4447) riffles: a few card backs flick up off that
  // library's pile and settle, over a cool light. They are backs only:
  // nothing about what the library holds, or what the search found, is
  // drawn, and a search that finds nothing looks the same.
  //
  // A double-faced permanent that transforms (#4448) flips on its own
  // vertical axis from the face it was to the face it is, as a copy over
  // its tile, which waits hidden until the copy settles (effectCueMotion.ts
  // flipHold, cancelled with the cue). The game becoming day or night
  // washes the table in the sky's tint, and the werewolves it turns over
  // flip a beat later.
  //
  // Every rule (which entry is which cue, what a board wipe is, the
  // timing and its cap, the three motion gates) is in lib/tableEffects.ts,
  // pure and unit-tested. This file measures, draws and plays.
  //
  // Reduced motion, all three ways (DESIGN.md §4): accessibility.reduceMotion,
  // animations.enabled (and animations.effects Off) and the OS signal, here
  // and in the @media block below. A still cue draws no motion: a badge
  // says what happened and stays for a fixed reading time; the sound plays.
  //
  // Skip: a click anywhere or Escape clears every cue playing. It is this
  // player's alone (nothing is sent), and the layer takes no pointer
  // events, so the click also reaches the board under it.
  //
  // Signatures (ADR 0209 §4, #4397): a named card's scene (fx/signatures.ts,
  // drawn from fx/scenes/) replaces the generic cue for the batch it names,
  // once its scene has loaded; until then, and whenever the cues are still,
  // the generic cues play. The cards its batch took keep their own cue,
  // lighter, on the signature's clock. Scenes load the first time a card
  // that has one appears in any zone this player can see. A scene with a
  // tint (Toxic Deluge, #4416) marks the creatures its batch left standing
  // until the turn number changes (fx/tints.ts; FxTint.svelte draws it).
  //
  // Decked out (#4879): a player who loses by drawing from an empty library
  // (cause empty_draw, CR 704.5b) gets the library's ghost, a signature-
  // length cue: their library slot pulses red, cold wisps rise from it, a
  // translucent blue skull rises out of the pile and swells to fill their
  // mat (clipped to it, never over another seat), sways and dissolves
  // upward, and as it fills the mat their hand's cards are flung outward,
  // spinning, and vanish (their faces for the player, backs for everyone
  // else). It counts on the busy clock and a skip clears it. Still and
  // reduced motion draw none of it: the grey seat and its "Decked" stone
  // are Tombstone.svelte's, and the sound is EliminationFx.svelte's.
  //
  // z 57 on DESIGN.md §4's ladder: above the action dock (55), below the
  // tutorial's coach (58). aria-hidden: the log and the board carry the news.

  import { onDestroy, untrack } from "svelte";
  import { get } from "svelte/store";
  import type { CardView, GameView } from "../../protocol";
  import { settings } from "../../settings";
  import {
    cardAnchors,
    findAnchor,
    findCardAnchor,
    findSeatAnchor,
    hasSize,
  } from "../../boardAnchor";
  import { stackItemSelector } from "../../castFlight";
  import { pileSelector } from "../../stackLinger";
  import { boardExpandLayout } from "../../boardExpand";
  import { cardImageURL } from "../../cardImage";
  import { showsCardBack } from "../../cardBack";
  import { isCreature, isPlaneswalker } from "../../cardTypes";
  import { crumble } from "../../animations";
  import { cancelAnimations, settled } from "../../motion";
  import { deckedMotion, type Box } from "../../deckedOutMotion";
  import { play } from "../../sounds";
  import { emptyBeatTracker, track } from "../../combatBeats";
  import {
    combatMotion,
    crumbleShards,
    strikesScheduled,
    usableTile,
    type CachedTile,
  } from "../../combatStrikes";
  import {
    BusyClock,
    EFFECT_CUE_CAP,
    SACRIFICE_KNIFE_SOUND,
    counterBurst,
    crackPlan,
    cueSound,
    cueTiming,
    deckedFlings,
    deckedSeats,
    deckedTiming,
    deckedWisps,
    effectLate,
    effectMotion,
    effectsScheduled,
    fizzleMotes,
    headlines,
    planEffects,
    sacrificeBeats,
    shatterPlan,
    soundBatch,
    spiralPath,
    swirlArms,
    voidMotes,
    windCuts,
    wipeTiming,
    type DeckedFling,
    type DeckedWisp,
    pileShown,
    PILE_SWEEP_MIN,
    RIFFLE_CARDS,
    type CounterTone,
    type CrackKind,
    type EffectCue,
    type EffectKind,
    type EffectMotion,
    type WipeTone,
  } from "../../tableEffects";
  import {
    attachSlide,
    commandArrive,
    cueMotion,
    flipHold,
    pileLandAt,
    riseArrive,
    tokenPop,
    type PilePath,
  } from "../../effectCueMotion";
  import { L } from "../../labels";
  import {
    SCENES,
    planSignatures,
    sceneForCard,
    signatureCueTiming,
    signatureScenes,
    signatureTiming,
    withSignatures,
    type SceneID,
    type SignatureCue,
    type SignatureLoader,
  } from "../../fx/signatures";
  import type { SceneCard, SceneSeat } from "../../fx/scenes/kit";
  import { syncTableTints, tintSurvivors } from "../../fx/tints";

  interface Props {
    view: GameView;
    boardEl: HTMLElement | null;
    // Changing this primes the next frame, as on a first frame: nothing
    // already on the wire is cued (Game.svelte's beatsPrimeKey: a
    // reconnect, a replay toggle).
    beatsPrimeKey?: string;
    // Where signature scenes come from (tests pass their own).
    signatures?: SignatureLoader;
  }

  const { view, boardEl, beatsPrimeKey = "", signatures = signatureScenes }: Props = $props();

  type FxKind = EffectKind | "wipe" | "signature";

  interface Fx {
    key: string;
    kind: FxKind;
    motion: EffectMotion;
    // Inside a board wipe: lighter.
    light: boolean;
    cardID: string;
    // The card's centre, in board pixels, its size at rest and its rotation.
    x: number;
    y: number;
    width: number;
    height: number;
    rot: number;
    src: string | null;
    // When the motion starts (after the frame) and how long it runs, in ms.
    delay: number;
    ms: number;
    // Where a bounced card flies to, from its own centre.
    fly: { x: number; y: number };
    tone: WipeTone;
    board: { w: number; h: number };
    seed: string;
    count: number;
    badge: string;
    // An exile that was a destruction (it crumbles a little as it goes),
    // and an exile drawn into a board wipe's vortex (its spiral path).
    fromDestroy?: boolean;
    pull?: { x: number; y: number }[] | null;
    // Counters put on a permanent (#4442): which kind, so which colour
    // and which way the tick goes.
    counterTone?: CounterTone;
    // A commander cast or going home (#4443): the size of what it reaches
    // (the stack item, its command-zone card), the copy's scale there,
    // and whether it lands on the card itself (else on the owner's seat).
    dest?: { w: number; h: number };
    commandScale?: number;
    landsOnCard?: boolean;
    // An attachment sliding onto its host (#4444): `fly` is where its
    // tile starts, in the tile's own px; the cue is centred on the host,
    // and `seam` is where the two meet, from that centre.
    seam?: { x: number; y: number };
    // Cards milled or discarded (#4445): each copy's art (null: its back),
    // in order, and whether they go as one sweep.
    slides?: (string | null)[];
    sweep?: boolean;
    // Cards returned from a graveyard (#4446): where each went, each one's
    // own way from its graveyard pile, whether it turns face down on the
    // way (into a hand this viewer cannot see), and the battlefield tile
    // that waits hidden for it (null: none).
    reanimateTo?: "battlefield" | "hand";
    paths?: PilePath[];
    flips?: boolean[];
    tileIDs?: (string | null)[];
    // A library searched (#4447): whose library pile riffles, a seat ID.
    library?: string;
    // A table-wide drain (#4450): each line, source to player, in board px.
    lines?: { x1: number; y1: number; x2: number; y2: number }[];
    // A permanent that transforms (#4448): the face it turns from (`src`
    // is the face it turns to). The game becoming day or night: which.
    from?: string;
    sky?: "day" | "night";
    // A permanent sacrificed for its own ability (#4427): what it was,
    // for the colour of the dust off its crack.
    crack?: CrackKind;
    // A signature: its scene, the cards its batch took, where it lands,
    // and the factor from the scene's ms to real ms.
    scene?: SceneID;
    sceneCards?: SceneCard[];
    landing?: SceneCard | null;
    seats?: SceneSeat[];
    k?: number;
  }

  let fx = $state<Fx[]>([]);

  // A decked-out cue (#4879). Boxes are in board px.
  interface Decked {
    key: string;
    seat: string;
    ms: number;
    // The player's mat (their seat panel), the library pile within it (in
    // the mat's own px) and their hand area.
    mat: Box;
    lib: Box;
    hand: Box;
    skull: number;
    diag: number;
    wisps: DeckedWisp[];
    cards: { src: string; w: number; h: number; f: DeckedFling }[];
  }
  let decks = $state<Decked[]>([]);
  // Each seat's hand as of the frame before: the frame that decks a player
  // has already taken their cards out of the game (CR 800.4a).
  const lastHands = new Map<string, { count: number; cards: CardView[] }>();
  let serial = 0;
  let destroyed = false;

  // Each battlefield tile's last measured box, and the image it showed, by
  // instance ID, and each stack item's last box. Not reactive.
  const tiles = new Map<string, CachedTile>();
  const images = new Map<string, string>();
  const lastViews = new Map<string, CardView>();
  const stackBoxes = new Map<string, { x: number; y: number; w: number; h: number }>();
  const timers = new Set<ReturnType<typeof setTimeout>>();

  let tracker = emptyBeatTracker();
  let reprimeNext = false;
  // How many cues started lately: a busy table plays them shorter (#4416).
  const busy = new BusyClock();

  // Every card this player can see, by instance ID, as of this frame and
  // the one before (a spell that resolved into a hidden zone, a permanent
  // that left): what a signature knows a card by.
  type Seen = Pick<CardView, "name" | "is_token" | "token_copy">;
  let seenNow = new Map<string, Seen>();
  let seenBefore = new Map<string, Seen>();

  function visibleCards(v: GameView): CardView[] {
    const out: CardView[] = [
      ...(v.battlefield?.cards ?? []),
      ...(v.stack?.cards ?? []),
      ...(v.exile?.cards ?? []),
    ];
    for (const s of v.seats ?? []) {
      out.push(
        ...(s.hand?.cards ?? []),
        ...(s.graveyard?.cards ?? []),
        ...(s.command?.cards ?? []),
      );
    }
    return out;
  }

  function sceneOfCard(id: string): SceneID | null {
    const c = seenNow.get(id) ?? seenBefore.get(id);
    return c ? sceneForCard(c) : null;
  }

  function later(fn: () => void, ms: number): void {
    const t = setTimeout(() => {
      timers.delete(t);
      if (!destroyed) fn();
    }, ms);
    timers.add(t);
  }

  function boardSize(): { rect: DOMRect; size: { w: number; h: number } } | null {
    if (!boardEl) return null;
    const rect = boardEl.getBoundingClientRect();
    return { rect, size: { w: rect.width, h: rect.height } };
  }

  function tileOf(el: HTMLElement, b: DOMRect, size: { w: number; h: number }): CachedTile {
    const r = el.getBoundingClientRect();
    const tapped = el.dataset.tapped === "true";
    return {
      box: {
        x: r.left + r.width / 2 - b.left,
        y: r.top + r.height / 2 - b.top,
        w: r.width,
        h: r.height,
      },
      width: tapped ? r.height : r.width,
      height: tapped ? r.width : r.height,
      rot: tapped ? 90 : 0,
      board: size,
    };
  }

  function onBattlefieldNow(): Set<string> {
    return new Set(view.battlefield.cards.map((c) => c.instance_id));
  }

  // Measure every live battlefield tile and stack item, and drop what
  // nothing needs. Runs after every frame, so the box of a permanent that
  // leaves in the next frame is the one it had just before.
  function measure(): void {
    measureBoxes(false);
  }

  // `onlyNew` measures only what has no box yet (#4410).
  function measureBoxes(onlyNew: boolean): void {
    const b = boardSize();
    if (!b || !boardEl) return;
    const live = onBattlefieldNow();
    for (const [id, el] of cardAnchors(boardEl)) {
      if (!live.has(id) || (onlyNew && tiles.has(id)) || !hasSize(el)) continue;
      tiles.set(id, tileOf(el, b.rect, b.size));
      const src = el.querySelector<HTMLImageElement>("img")?.getAttribute("src");
      if (src) images.set(id, src);
    }
    for (const el of boardEl.querySelectorAll<HTMLElement>("[data-stack-item-id]")) {
      const id = el.dataset.stackItemId;
      if (!id || (onlyNew && stackBoxes.has(id)) || !hasSize(el)) continue;
      const r = el.getBoundingClientRect();
      stackBoxes.set(id, {
        x: r.left + r.width / 2 - b.rect.left,
        y: r.top + r.height / 2 - b.rect.top,
        w: r.width,
        h: r.height,
      });
    }
    measureCommand(b.rect);
    prune();
  }

  function prune(): void {
    if (fx.length > 0 || timers.size > 0) return;
    const live = onBattlefieldNow();
    for (const id of [...tiles.keys()]) if (!live.has(id)) tiles.delete(id);
    for (const id of [...images.keys()]) if (!live.has(id)) images.delete(id);
    for (const id of [...lastViews.keys()]) if (!live.has(id)) lastViews.delete(id);
    if (stackBoxes.size > 64 && boardEl) {
      const now = new Set(
        [...boardEl.querySelectorAll<HTMLElement>("[data-stack-item-id]")].map(
          (e) => e.dataset.stackItemId,
        ),
      );
      for (const id of [...stackBoxes.keys()]) if (!now.has(id)) stackBoxes.delete(id);
    }
  }

  // A card's box now: its live tile if it is on the battlefield, else the
  // cached one, if that was measured on a board of this size.
  function cardTile(
    id: string,
    b: { rect: DOMRect; size: { w: number; h: number } },
    liveOnly = false,
  ): CachedTile | null {
    if (boardEl && onBattlefieldNow().has(id)) {
      const el = findCardAnchor(boardEl, id, { accept: hasSize });
      if (el) return tileOf(el, b.rect, b.size);
    }
    if (liveOnly) return null;
    return usableTile(tiles.get(id), b.size);
  }

  function imageFor(id: string): string | null {
    const cached = images.get(id);
    if (cached) return cached;
    const card = lastViews.get(id) ?? view.battlefield.cards.find((c) => c.instance_id === id);
    if (!card) return null;
    return showsCardBack(card) ? "/card-back-small.jpg" : cardImageURL(card, "small");
  }

  function isPW(id: string): boolean {
    const card = lastViews.get(id) ?? view.battlefield.cards.find((c) => c.instance_id === id);
    return card ? isPlaneswalker(card) : false;
  }

  function osReducesMotion(): boolean {
    return (
      typeof window !== "undefined" &&
      typeof window.matchMedia === "function" &&
      window.matchMedia("(prefers-reduced-motion: reduce)").matches
    );
  }

  const BADGE: Record<EffectKind, string | null> = {
    destroy: L.effectBadgeDestroyed,
    dies: L.effectBadgeDied,
    exile: L.effectBadgeExiled,
    sacrifice: L.effectBadgeSacrificed,
    bounce: L.effectBadgeBounced,
    // StackLinger's "Countered" badge already says it, in every motion mode.
    counter: null,
    // And its "Fizzled" badge says this one (#4450).
    fizzle: null,
    // A drain and the crown (#4450): the life totals and the marker on the
    // panel say it, and the log names it.
    drain_table: null,
    monarch: null,
    initiative: null,
    pw_enter: L.effectBadgePwEnter,
    pw_leave: L.effectBadgePwLeave,
    // The token itself is the news: it is on the board, and the log says
    // who made it. Still, it draws nothing, and its sound plays.
    token_create: null,
    // So are the counters: the badge already shows the new numbers.
    counters: null,
    // And a commander (#4443): the spell is on the stack, or the card is
    // back in its command zone, where the board shows it.
    commander_cast: null,
    commander_return: null,
    // And an attachment (#4444): the board draws it under its host.
    attach_equip: null,
    attach_aura: null,
    // And a mill or a discard (#4445): the cards are in the graveyard,
    // where the board shows them, and the log names them.
    mill: null,
    discard: null,
    // And a card returned from a graveyard (#4446): it is where it went,
    // where the board shows it, and the log names it.
    reanimate: null,
    // And a library searched (#4447): the log says who searched whose
    // library, and the board shows where whatever was found went.
    library_search: null,
    // And a permanent that transforms, or the game becoming day or night
    // (#4448): the board shows the new face, the dock shows day or night,
    // and the log says both.
    transform: null,
    day_night: null,
    // And a permanent sacrificed for its own ability (#4427): it plays
    // several times a game, the log says it, and the mana is in the pool.
    crack: null,
  };

  function add(f: Omit<Fx, "key">): void {
    const key = `fx-${++serial}`;
    fx = [...fx, { ...f, key }];
    // The copy is gone when its motion is (plus a beat); a still badge
    // stays for its reading time.
    later(() => remove(key), f.delay + f.ms + 80);
  }

  function remove(key: string): void {
    fx = fx.filter((f) => f.key !== key);
    if (fx.length === 0) prune();
  }

  // Skip (and a priming frame): every cue playing is gone now. The state
  // under them is untouched, and the next frame's cues play as usual.
  function clearAll(): void {
    for (const t of timers) clearTimeout(t);
    timers.clear();
    fx = [];
    decks = [];
    busy.reset();
  }

  const blank = {
    cardID: "",
    x: 0,
    y: 0,
    width: 0,
    height: 0,
    rot: 0,
    src: null,
    fly: { x: 0, y: 0 },
    tone: "dark" as WipeTone,
    seed: "",
    count: 0,
    badge: "",
  };

  function drawCue(
    cue: EffectCue,
    motion: EffectMotion,
    t: { delayMs: number; ms: number },
    b: { rect: DOMRect; size: { w: number; h: number } },
    vortex = false,
  ): void {
    const base = {
      ...blank,
      kind: cue.kind as FxKind,
      motion,
      light: cue.weight === "light",
      delay: t.delayMs,
      ms: t.ms,
      board: b.size,
      seed: cue.id,
      cardID: cue.cardID ?? "",
      fromDestroy: cue.fromDestroy === true,
    };
    if (cue.kind === "counter" || cue.kind === "fizzle") {
      // The stack linger's badge says "Countered" (or "Fizzled", #4450);
      // only the motion is ours.
      if (motion === "still") return;
      const box = cue.stackItemID ? stackBoxes.get(cue.stackItemID) : undefined;
      if (!box) return;
      add({ ...base, x: box.x, y: box.y, width: box.w, height: box.h });
      return;
    }
    // A mill names its cards only when the server says which (#4445).
    if (cue.kind === "mill" || cue.kind === "discard") {
      placePile(cue, base, motion, b);
      return;
    }
    if (cue.kind === "reanimate") {
      placeReanimate(cue, base, motion, b);
      return;
    }
    if (cue.kind === "library_search") {
      placeRiffle(cue, base, motion, b);
      return;
    }
    if (cue.kind === "drain_table") {
      placeDrain(cue, base, motion, b);
      return;
    }
    if (cue.kind === "monarch" || cue.kind === "initiative") {
      placeGlint(cue, base, motion, b);
      return;
    }
    if (cue.kind === "day_night") {
      placeSky(cue, base, motion, b);
      return;
    }
    if (cue.kind === "transform") {
      placeTransform(cue, base, motion, b);
      return;
    }
    const cardID = cue.cardID;
    if (!cardID) return;
    if (cue.kind === "token_create") {
      placeToken(cardID, base, motion, b);
      return;
    }
    if (cue.kind === "counters") {
      placeCounters(cue, base, motion, b);
      return;
    }
    if (cue.kind === "commander_cast" || cue.kind === "commander_return") {
      placeCommander(cue, base, motion, b);
      return;
    }
    if (cue.kind === "attach_equip" || cue.kind === "attach_aura") {
      placeAttach(cue, base, motion, b);
      return;
    }
    // #4427: a permanent sacrificed for its own ability snaps where it
    // stood, drawn from its last tile. Still, it draws nothing (no badge:
    // it plays several times a game, and the log says it); its sound plays.
    if (cue.kind === "crack") {
      if (motion === "still") return;
      place(cue, { ...base, crack: cue.crackOf ?? "other" }, motion, b);
      return;
    }
    // A planeswalker that just entered is still in its entrance pop (the
    // tile is scaled down and fading in), so its box is read once that
    // has settled, not when the frame arrives. A leaver is drawn at once
    // from its cache, so the card never blinks out.
    if (cue.kind === "pw_enter") {
      const tile = boardEl ? findCardAnchor(boardEl, cardID, { accept: hasSize }) : null;
      if (!tile) return;
      const arrive = () =>
        void settled(tile).then(() => {
          if (destroyed) return;
          const now = boardSize();
          if (now) place(cue, base, motion, now);
        });
      // One that rose out of a graveyard (#4446) follows its rise, whose
      // cue mounts after this frame's cues are placed and holds the tile
      // hidden until its copy lands: it waits for that too.
      if (cue.follows) later(arrive, 0);
      else arrive();
      return;
    }
    place(cue, base, motion, b, vortex);
  }

  // place measures the card behind a cue and mounts it, when it can be found.
  function place(
    cue: EffectCue,
    base: Omit<Fx, "key">,
    motion: EffectMotion,
    b: { rect: DOMRect; size: { w: number; h: number } },
    vortex = false,
  ): void {
    const cardID = cue.cardID ?? "";
    const where = cardTile(cardID, b, cue.kind === "pw_enter");
    if (!where) return;
    const f = {
      ...base,
      x: where.box.x,
      y: where.box.y,
      width: where.width,
      height: where.height,
      rot: where.rot,
      src: cue.kind === "pw_enter" ? null : imageFor(cardID),
      badge: BADGE[cue.kind] ?? "",
    };
    if (cue.kind === "bounce" && motion !== "still") {
      // To its owner's seat; straight up when that seat cannot be found.
      const owner = lastViews.get(cardID)?.owner;
      const el = owner && boardEl ? findSeatAnchor(boardEl, owner, { accept: hasSize }) : null;
      if (el) {
        const r = el.getBoundingClientRect();
        f.fly = {
          x: r.left + r.width / 2 - b.rect.left - where.box.x,
          y: r.top + r.height / 2 - b.rect.top - where.box.y,
        };
      } else {
        f.fly = { x: 0, y: -Math.max(40, where.height * 0.6) };
      }
    }
    if (vortex && motion !== "still") {
      f.pull = spiralPath({ x: where.box.x, y: where.box.y }, { x: b.size.w / 2, y: b.size.h / 2 });
    }
    add(f);
  }

  // A created token's glow, over its live tile (#4441), mounted at once so
  // the pop starts with the frame. The tile is mid-way through its own
  // entrance pop, which scales its box about its centre, so the centre is
  // read from the box and the size from the layout, which no transform
  // changes. A token already gone, or a still cue, draws nothing.
  function placeToken(
    cardID: string,
    base: Omit<Fx, "key">,
    motion: EffectMotion,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): void {
    if (motion === "still" || !boardEl || !onBattlefieldNow().has(cardID)) return;
    const el = findCardAnchor(boardEl, cardID, { accept: hasSize });
    if (!el) return;
    const r = el.getBoundingClientRect();
    const tapped = el.dataset.tapped === "true";
    add({
      ...base,
      x: r.left + r.width / 2 - b.rect.left,
      y: r.top + r.height / 2 - b.rect.top,
      width: el.offsetWidth || (tapped ? r.height : r.width),
      height: el.offsetHeight || (tapped ? r.width : r.height),
      rot: tapped ? 90 : 0,
    });
  }

  // Counters put on a permanent (#4442): the pulse rings what shows the
  // count. +1/+1 and -1/-1 counters ring the power/toughness badge;
  // another kind rings its own pip (CounterPips). A tile with neither
  // (a token group's member, a vehicle not yet a creature) rings its own
  // corner. The badges stay upright on a tapped tile, so no turn. A
  // permanent already gone, or a still cue, draws nothing.
  function placeCounters(
    cue: EffectCue,
    base: Omit<Fx, "key">,
    motion: EffectMotion,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): void {
    const cardID = cue.cardID ?? "";
    if (motion === "still" || !boardEl || !onBattlefieldNow().has(cardID)) return;
    const tile = findCardAnchor(boardEl, cardID, { accept: hasSize });
    if (!tile) return;
    const tone = cue.counterTone ?? "other";
    const label = cue.counterLabel ?? "";
    const pip = [...tile.querySelectorAll<HTMLElement>("[data-counter]")].find(
      (el) => el.dataset.counter === label && hasSize(el),
    );
    const pt = tile.querySelector<HTMLElement>(".badge.pt");
    const el = (tone === "other" ? pip : pt && hasSize(pt) ? pt : pip) ?? null;
    const r = (el ?? tile).getBoundingClientRect();
    // A tile's corner: a box the size of a P/T badge, inside its bottom right.
    const w = Math.min(36, r.width);
    const box = el
      ? r
      : { left: r.left + r.width - w - 4, top: r.top + r.height - 22, width: w, height: 18 };
    add({
      ...base,
      x: box.left + box.width / 2 - b.rect.left,
      y: box.top + box.height / 2 - b.rect.top,
      width: box.width,
      height: box.height,
      counterTone: tone,
    });
  }

  // ---- A commander cast from the command zone, or going home (#4443) ----

  type CmdBox = { x: number; y: number; w: number; h: number };

  // Each commander's last box in its command zone, and whose command zone
  // it was, so a cast's light starts where it sat once the card has left.
  // Rebuilt on every measure. Not reactive.
  const commandBoxes = new Map<string, CmdBox & { owner: string }>();

  // The other motion layers' copies are never where a commander is.
  const CMD_OVERLAYS =
    "[data-table-effects], [data-cast-flight-layer], [data-stack-linger-layer], [data-field-motion-layer], [data-stack-arrival-layer]";
  const cmdLive = (el: HTMLElement) => hasSize(el) && !el.closest(CMD_OVERLAYS);

  function cmdBox(el: HTMLElement, b: DOMRect): CmdBox {
    const r = el.getBoundingClientRect();
    return {
      x: r.left + r.width / 2 - b.left,
      y: r.top + r.height / 2 - b.top,
      w: r.width,
      h: r.height,
    };
  }

  function measureCommand(b: DOMRect): void {
    commandBoxes.clear();
    if (!boardEl) return;
    for (const s of view.seats ?? []) {
      for (const c of s.command?.cards ?? []) {
        const el = findCardAnchor(boardEl, c.instance_id, { accept: cmdLive });
        if (el) commandBoxes.set(c.instance_id, { ...cmdBox(el, b), owner: s.id });
      }
    }
  }

  // The commander's card in a command zone now, and that zone's seat.
  function inCommand(id: string): { card: CardView; seat: string } | null {
    for (const s of view.seats ?? []) {
      const card = s.command?.cards.find((c) => c.instance_id === id);
      if (card) return { card, seat: s.id };
    }
    return null;
  }

  function isCommanderNow(id: string): boolean {
    return inCommand(id)?.card.is_commander === true;
  }

  function commanderOwner(id: string): string {
    return (
      inCommand(id)?.seat ??
      commandBoxes.get(id)?.owner ??
      lastViews.get(id)?.owner ??
      view.stack?.cards.find((c) => c.instance_id === id)?.owner ??
      ""
    );
  }

  function liveBox(selector: string, b: DOMRect): CmdBox | null {
    if (!boardEl) return null;
    const el = findAnchor(boardEl, selector, { accept: cmdLive });
    return el ? cmdBox(el, b) : null;
  }

  function seatBox(seat: string, b: DOMRect): CmdBox | null {
    if (!boardEl || !seat) return null;
    const el = findSeatAnchor(boardEl, seat, { accept: cmdLive });
    return el ? cmdBox(el, b) : null;
  }

  // A cast rises from the commander's last command-zone box (else its
  // owner's seat) to its spell on the stack (else straight up); a return
  // drifts from where it was to its card in the command zone (else its
  // owner's seat). One end missing is made up from the other; with
  // neither, or a still cue, nothing is drawn and the sound plays.
  function placeCommander(
    cue: EffectCue,
    base: Omit<Fx, "key">,
    motion: EffectMotion,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): void {
    if (motion === "still" || !boardEl) return;
    const id = cue.cardID ?? "";
    const owner = commanderOwner(id);
    if (cue.kind === "commander_cast") {
      let from: CmdBox | null = commandBoxes.get(id) ?? seatBox(owner, b.rect);
      let to = liveBox(stackItemSelector(id), b.rect);
      if (!from && !to) return;
      if (!from && to) from = { x: to.x, y: to.y + to.h, w: to.w * 0.6, h: to.h * 0.6 };
      if (!from) return;
      if (!to) {
        to = {
          x: from.x,
          y: from.y - Math.max(120, from.h * 1.5),
          w: from.w * 1.4,
          h: from.h * 1.4,
        };
      }
      add({
        ...base,
        x: from.x,
        y: from.y,
        width: from.w,
        height: from.h,
        fly: { x: to.x - from.x, y: to.y - from.y },
        dest: { w: to.w, h: to.h },
      });
      return;
    }

    // commander_return: where it lands first, so a pile's start can take
    // the card's size.
    const cardEl = findCardAnchor(boardEl, id, { accept: cmdLive });
    const onCard = cardEl !== null && inCommand(id) !== null;
    let to: CmdBox | null = onCard && cardEl ? cmdBox(cardEl, b.rect) : seatBox(owner, b.rect);
    let start: { box: CmdBox; width: number; height: number; rot: number } | null = null;
    const zone = cue.commanderFrom ?? "";
    if (zone === "battlefield") {
      const t = cardTile(id, b);
      if (t) start = { box: t.box, width: t.width, height: t.height, rot: t.rot };
    } else if (zone === "stack") {
      const s = stackBoxes.get(id);
      if (s) start = { box: s, width: s.w, height: s.h, rot: 0 };
    } else {
      const pile =
        zone === "hand"
          ? liveBox(`[data-hand-owner="${owner.replace(/["\\]/g, "\\$&")}"]`, b.rect)
          : zone === "graveyard" || zone === "exile" || zone === "library"
            ? liveBox(pileSelector(zone, owner), b.rect)
            : null;
      const p = pile ?? seatBox(owner, b.rect);
      // A pile or a seat is no card's size: the copy starts at its card's.
      const w = to && onCard ? to.w : 60;
      const h = to && onCard ? to.h : 84;
      if (p) start = { box: p, width: w, height: h, rot: 0 };
    }
    if (!start && !to) return;
    if (!start && to) {
      start = { box: { ...to, y: to.y - 90 }, width: to.w, height: to.h, rot: 0 };
    }
    if (!start) return;
    if (!to) to = { ...start.box, y: start.box.y - 80, w: start.width, h: start.height };
    const scale = onCard ? to.w / Math.max(1, start.width) : 0.4;
    const card = inCommand(id)?.card;
    add({
      ...base,
      x: start.box.x,
      y: start.box.y,
      width: start.width,
      height: start.height,
      rot: start.rot,
      src: (card ? cardImageURL(card, "small") : null) ?? imageFor(id),
      fly: { x: to.x - start.box.x, y: to.y - start.box.y },
      dest: { w: to.w, h: to.h },
      commandScale: Math.min(3, Math.max(0.2, scale)),
      landsOnCard: onCard,
    });
  }

  // The commander's own card in the command zone, which waits hidden for
  // its copy to land (#4443).
  function commandCardOf(f: Fx): HTMLElement | null {
    if (!f.landsOnCard || !boardEl) return null;
    return findCardAnchor(boardEl, f.cardID, { accept: cmdLive });
  }

  // A beam or a trail from a cue's start to its end: its length, and its
  // turn from straight up, in degrees.
  function reachOf(f: Fx): { len: number; turn: number } {
    return {
      len: Math.hypot(f.fly.x, f.fly.y),
      turn: (Math.atan2(f.fly.x, -f.fly.y) * 180) / Math.PI,
    };
  }
  const CMD_MOTES = [0, 1, 2, 3, 4, 5, 6];

  // ---- An Equipment is attached, or an Aura enchants a permanent (#4444) ----

  // How far a new attachment slides in from, as shares of its own size:
  // from above its rest and out to the side it peeks from.
  const ATTACH_IN = { x: -0.3, y: -0.45 };

  // The attachment's own tile slides onto its host from its last box on
  // the battlefield (an Equipment moving from where it stood, or from
  // another creature), or, when it is new there (an Aura that resolved or
  // entered attached, a token created attached), from just above and to
  // the side of its rest. The cue is centred on the host, where its spark
  // or ring of light is drawn; a host the board draws no tile for (a
  // token group's member) lends the attachment's own box. An attachment
  // already gone, or a still cue, draws nothing, and the sound plays.
  function placeAttach(
    cue: EffectCue,
    base: Omit<Fx, "key">,
    motion: EffectMotion,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): void {
    const id = cue.cardID ?? "";
    if (motion === "still" || !boardEl || !onBattlefieldNow().has(id)) return;
    const el = findCardAnchor(boardEl, id, { accept: cmdLive });
    if (!el) return;
    const now = tileOf(el, b.rect, b.size);
    const w = el.offsetWidth || now.width;
    const h = el.offsetHeight || now.height;
    // Board px to the tile's own px (a row may draw its cards scaled).
    const k = el.offsetWidth > 0 ? now.width / el.offsetWidth : 1;
    const before = usableTile(tiles.get(id), b.size);
    const moved =
      before !== null && Math.hypot(before.box.x - now.box.x, before.box.y - now.box.y) > 4;
    const fly = moved
      ? { x: (before.box.x - now.box.x) / k, y: (before.box.y - now.box.y) / k }
      : { x: ATTACH_IN.x * w, y: ATTACH_IN.y * h };
    const hostEl = cue.hostID ? findCardAnchor(boardEl, cue.hostID, { accept: cmdLive }) : null;
    const r = (hostEl ?? el).getBoundingClientRect();
    add({
      ...base,
      x: r.left + r.width / 2 - b.rect.left,
      y: r.top + r.height / 2 - b.rect.top,
      width: r.width,
      height: r.height,
      fly,
      // The host's left edge, a little below its middle: where the
      // attachment, drawn out to its left and a little lower, meets it.
      seam: hostEl ? { x: -r.width / 2 + 2, y: r.height * 0.08 } : { x: 0, y: 0 },
    });
  }
  const ATT_SPARKS = [0, 1, 2, 3, 4, 5];

  // ---- Cards milled or discarded into a graveyard (#4445) ----

  // A card copy's size when no pile on the board gives one.
  const PILE_CARD = { w: 46, h: 64 };

  // Whose graveyard a card is in now, as a seat ID.
  function graveyardOwner(id: string): string | null {
    for (const s of view.seats ?? []) {
      if (s.graveyard?.cards.some((c) => c.instance_id === id)) return s.id;
    }
    return null;
  }

  function graveyardCard(id: string): CardView | null {
    for (const s of view.seats ?? []) {
      const c = s.graveyard?.cards.find((g) => g.instance_id === id);
      if (c) return c;
    }
    return null;
  }

  // A pile's top card, where a copy lands or leaves from, else the pile.
  function pileBox(zone: "graveyard" | "library", owner: string, b: DOMRect): CmdBox | null {
    const sel = pileSelector(zone, owner);
    return liveBox(`${sel} .thumb`, b) ?? liveBox(sel, b);
  }

  // A discarded card's own tile in its owner's hand, while the hand lets
  // it go (the viewer's own hand; an opponent's shows only backs), never
  // the graveyard pile's thumbnail of it; else the hand.
  function inHand(id: string | undefined, owner: string, b: DOMRect): CmdBox | null {
    if (!boardEl) return null;
    const hand = `[data-hand-owner="${owner.replace(/["\\]/g, "\\$&")}"]`;
    const el = id
      ? findCardAnchor(boardEl, id, {
          accept: (e) =>
            cmdLive(e) && e.closest(hand) !== null && e.closest("[data-pile]") === null,
        })
      : null;
    return el ? cmdBox(el, b) : liveBox(hand, b);
  }

  // The cards slide from the owner's library pile (a mill) or hand (a
  // discard) to their graveyard pile, at the size of the graveyard's top
  // card. With no graveyard pile on the board they go to the owner's
  // seat, and with no start they drop onto it from just above; with
  // neither end, or a still cue, nothing is drawn and the sound plays.
  function placePile(
    cue: EffectCue,
    base: Omit<Fx, "key">,
    motion: EffectMotion,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): void {
    if (motion === "still" || !boardEl) return;
    const ids = cue.cardIDs ?? [];
    const owner =
      (ids.length > 0 ? graveyardOwner(ids[0]) : null) ?? view.seats?.[cue.seat ?? -1]?.id ?? "";
    if (!owner) return;
    const to = pileBox("graveyard", owner, b.rect) ?? seatBox(owner, b.rect);
    if (!to) return;
    const from = (cue.kind === "mill"
      ? pileBox("library", owner, b.rect)
      : inHand(ids[0], owner, b.rect)) ?? { ...to, y: to.y - 90 };
    const count = Math.max(1, cue.count ?? ids.length);
    const sweep = count >= PILE_SWEEP_MIN;
    const shown = pileShown(count);
    // The pile's top card is card-sized; a seat is not.
    const card = to.h > to.w * 1.2 ? { w: to.w, h: to.h } : PILE_CARD;
    const slides: (string | null)[] = [];
    for (let i = 0; i < shown; i++) {
      const c = ids[i] ? graveyardCard(ids[i]) : null;
      slides.push(c && !showsCardBack(c) ? cardImageURL(c, "small") : null);
    }
    add({
      ...base,
      cardID: ids[0] ?? "",
      x: from.x,
      y: from.y,
      width: card.w,
      height: card.h,
      fly: { x: to.x - from.x, y: to.y - from.y },
      count,
      slides,
      sweep,
    });
  }

  // ---- A card returned from a graveyard (#4446) ----

  // Every graveyard card this player saw, by instance ID, with whose
  // graveyard it was in, as of this frame and the one before: a card that
  // left for a hand is not in it any more, and an opponent's hand shows
  // only backs, so its owner and its face are read from the frame before.
  type InGrave = { owner: string; card: CardView };
  let graveNow = new Map<string, InGrave>();
  let graveBefore = new Map<string, InGrave>();

  function graveyardCards(v: GameView): Map<string, InGrave> {
    const out = new Map<string, InGrave>();
    for (const s of v.seats ?? []) {
      for (const c of s.graveyard?.cards ?? []) out.set(c.instance_id, { owner: s.id, card: c });
    }
    return out;
  }

  // A card in a hand this player sees (their own), and whose hand it is.
  function handCard(id: string): { card: CardView; seat: string } | null {
    for (const s of view.seats ?? []) {
      const card = s.hand?.cards.find((c) => c.instance_id === id);
      if (card) return { card, seat: s.id };
    }
    return null;
  }

  // A permanent's own tile on the battlefield, never a pile's thumbnail.
  const fieldTile = (el: HTMLElement) => cmdLive(el) && el.closest("[data-pile]") === null;

  // Each card rises out of its owner's graveyard pile (else their seat)
  // and goes to where it went: its tile on the battlefield, at the tile's
  // size (else its controller's seat), or its card in the viewer's own
  // hand (else the owner's hand, else their seat), turning face down on
  // the way into a hand this player cannot see. A card with neither end,
  // or a still cue, draws nothing; the sound plays.
  function placeReanimate(
    cue: EffectCue,
    base: Omit<Fx, "key">,
    motion: EffectMotion,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): void {
    if (motion === "still" || !boardEl) return;
    const ids = cue.cardIDs ?? [];
    const count = Math.max(1, cue.count ?? ids.length);
    const sweep = count >= PILE_SWEEP_MIN;
    const toHand = cue.reanimateTo === "hand";
    const slides: (string | null)[] = [];
    const flips: boolean[] = [];
    const paths: PilePath[] = [];
    const tileIDs: (string | null)[] = [];
    let origin: CmdBox | null = null;
    let card: { w: number; h: number } | null = null;
    for (const id of ids.slice(0, pileShown(count))) {
      const onField = toHand ? null : view.battlefield.cards.find((c) => c.instance_id === id);
      const held = toHand ? handCard(id) : null;
      const before = graveBefore.get(id);
      const owner =
        onField?.owner ?? before?.owner ?? held?.seat ?? view.seats?.[cue.seat ?? -1]?.id ?? "";
      if (!owner) continue;
      const pile = pileBox("graveyard", owner, b.rect);
      // The pile's top card is card-sized; a seat is not.
      card ??= pile && pile.h > pile.w * 1.2 ? { w: pile.w, h: pile.h } : PILE_CARD;
      let to: CmdBox | null = null;
      let scale = 1;
      let tile: string | null = null;
      if (onField) {
        const el = findCardAnchor(boardEl, id, { accept: fieldTile });
        if (el) {
          // Mid-way through its entrance pop: the centre from the box, the
          // size from the layout, which no transform changes.
          const r = el.getBoundingClientRect();
          to = {
            x: r.left + r.width / 2 - b.rect.left,
            y: r.top + r.height / 2 - b.rect.top,
            w: el.offsetWidth || r.width,
            h: el.offsetHeight || r.height,
          };
          scale = to.w / Math.max(1, card.w);
          // An Aura returned attached keeps its own slide onto its host
          // in sight (the attach cue); any other tile waits for its copy.
          if (!onField.attached_to) tile = id;
        }
        to ??= seatBox(onField.controller || owner, b.rect);
      } else if (toHand) {
        to = inHand(id, owner, b.rect);
        if (to && held) scale = to.w / Math.max(1, card.w);
        to ??= seatBox(owner, b.rect);
      }
      let from: CmdBox | null = pile ?? seatBox(owner, b.rect);
      if (!from && !to) continue;
      if (!from && to) from = { ...to, y: to.y + 90 };
      if (!from) continue;
      if (!to) to = { ...from, y: from.y - 90 };
      origin ??= from;
      paths.push({
        x: from.x - origin.x,
        y: from.y - origin.y,
        dx: to.x - from.x,
        dy: to.y - from.y,
        s: Math.min(3, Math.max(0.3, scale)),
      });
      // Its face: as the board shows it now, or, on its way into a hand
      // this player cannot see, as the graveyard showed it, turning over.
      const face = onField ?? held?.card ?? before?.card ?? null;
      slides.push(face && !showsCardBack(face) ? cardImageURL(face, "small") : null);
      flips.push(toHand && !held && face !== null);
      tileIDs.push(tile);
    }
    if (!origin || !card || paths.length === 0) return;
    add({
      ...base,
      cardID: ids[0] ?? "",
      x: origin.x,
      y: origin.y,
      width: card.w,
      height: card.h,
      fly: { x: paths[0].dx, y: paths[0].dy },
      count,
      slides,
      sweep,
      reanimateTo: cue.reanimateTo,
      paths,
      flips,
      tileIDs,
    });
  }
  // ---- A library searched (#4447) ----

  // The riffled card backs, one each.
  const RIFFLE = Array.from({ length: RIFFLE_CARDS }, (_, i) => i);

  // The searched library's pile riffles, at the size of its top card (a
  // back); with no library pile on the board, its owner's seat does, at a
  // card's size. With neither, or a still cue, nothing is drawn and the
  // sound plays.
  function placeRiffle(
    cue: EffectCue,
    base: Omit<Fx, "key">,
    motion: EffectMotion,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): void {
    if (motion === "still" || !boardEl) return;
    const owner = view.seats?.[cue.librarySeat ?? -1]?.id ?? "";
    if (!owner) return;
    const at = pileBox("library", owner, b.rect) ?? seatBox(owner, b.rect);
    if (!at) return;
    // The pile's top card is card-sized; a seat is not.
    const card = at.h > at.w * 1.2 ? { w: at.w, h: at.h } : PILE_CARD;
    add({
      ...base,
      x: at.x,
      y: at.y,
      width: card.w,
      height: card.h,
      count: RIFFLE_CARDS,
      library: owner,
    });
  }

  // ---- A table-wide drain, the crown and the initiative (#4450) ----

  // One thin dark line from the source (its tile on the battlefield, else
  // the middle of the table) to each player's panel that lost life. A
  // player with no panel on the board gets none; with no line at all, or
  // a still cue, nothing is drawn and the sound plays.
  function placeDrain(
    cue: EffectCue,
    base: Omit<Fx, "key">,
    motion: EffectMotion,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): void {
    if (motion === "still" || !boardEl) return;
    const id = cue.cardID ?? "";
    const el = id ? findCardAnchor(boardEl, id, { accept: fieldTile }) : null;
    const from = el ? cmdBox(el, b.rect) : null;
    const sx = from ? from.x : b.size.w / 2;
    const sy = from ? from.y : b.size.h / 2;
    const lines = [];
    for (const i of cue.drainSeats ?? []) {
      const to = seatBox(view.seats?.[i]?.id ?? "", b.rect);
      if (to) lines.push({ x1: sx, y1: sy, x2: to.x, y2: to.y });
    }
    if (lines.length === 0) return;
    add({ ...base, x: sx, y: sy, lines });
  }

  // A soft ring and four sparks over the new holder's panel.
  const GLINT_SPARKS = [
    { dx: -26, dy: -22 },
    { dx: 28, dy: -16 },
    { dx: -20, dy: 24 },
    { dx: 24, dy: 20 },
  ];
  function placeGlint(
    cue: EffectCue,
    base: Omit<Fx, "key">,
    motion: EffectMotion,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): void {
    if (motion === "still" || !boardEl) return;
    const at = seatBox(view.seats?.[cue.seat ?? -1]?.id ?? "", b.rect);
    if (!at) return;
    add({ ...base, x: at.x, y: at.y, width: Math.min(at.w, 220), height: Math.min(at.h, 120) });
  }

  // ---- A double-faced card transforms, or day and night (#4448) ----

  // `src` as the face `face` of its card, at the size of `like`: an image
  // URL built by cardImage.ts ("/cards/{id}/image?size=…&face=N", face 0
  // with no face).
  function imageAs(src: string, like: string, face?: number): string {
    const [path, query = ""] = src.split("?");
    const q = new URLSearchParams(query);
    const size = new URLSearchParams(like.split("?")[1] ?? "").get("size");
    if (size) q.set("size", size);
    if (face !== undefined) {
      if (face > 0) q.set("face", String(face));
      else q.delete("face");
    }
    const out = q.toString();
    return out ? `${path}?${out}` : path;
  }

  // The permanent flips over its own tile, from the face it was to the
  // face its tile shows now, at the tile's size and turn. The face it was
  // is its other face, or, for a melded permanent, its first card's front
  // (its tile has just entered, so its size is the layout's, which its
  // entrance pop does not change). A permanent already gone, one with no
  // art, or a still cue draws nothing, and the sound plays.
  function placeTransform(
    cue: EffectCue,
    base: Omit<Fx, "key">,
    motion: EffectMotion,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): void {
    const id = cue.cardID ?? "";
    if (motion === "still" || !boardEl || !onBattlefieldNow().has(id)) return;
    const card = view.battlefield.cards.find((c) => c.instance_id === id);
    const el = findCardAnchor(boardEl, id, { accept: fieldTile });
    if (!card || !el || showsCardBack(card)) return;
    const to =
      el.querySelector<HTMLImageElement>("img")?.getAttribute("src") ?? cardImageURL(card, "small");
    if (!to) return;
    const first = card.melded_from?.[0]?.image;
    let from: string | null = null;
    if (cue.melded) from = first ? imageAs(first, to) : null;
    else if ((card.faces?.length ?? 0) >= 2) {
      from = imageAs(to, to, (card.active_face ?? 0) === 0 ? 1 : 0);
    }
    if (!from) return;
    const now = tileOf(el, b.rect, b.size);
    add({
      ...base,
      x: now.box.x,
      y: now.box.y,
      width: cue.melded ? el.offsetWidth || now.width : now.width,
      height: cue.melded ? el.offsetHeight || now.height : now.height,
      rot: now.rot,
      src: to,
      from,
    });
  }

  // The sky's wash covers the whole table. A still cue draws nothing, and
  // the sound plays.
  function placeSky(
    cue: EffectCue,
    base: Omit<Fx, "key">,
    motion: EffectMotion,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): void {
    if (motion === "still") return;
    add({ ...base, width: b.size.w, height: b.size.h, sky: cue.sky ?? "night" });
  }

  // Where an Aura's glints twinkle, as shares of the host's box from its centre.
  const ATT_GLINTS = [
    [-0.44, -0.34],
    [0.36, -0.46],
    [0.47, 0.12],
    [-0.47, 0.24],
    [0.12, 0.49],
    [-0.1, -0.5],
  ];

  function frame(log: GameView["log"]): void {
    const prime = reprimeNext;
    reprimeNext = false;
    const r = track(tracker, log, prime);
    tracker = r.tracker;
    // A reconnect or a replay move: forget what was playing.
    if (r.primed) clearAll();

    // What this frame shows, and any scene among it, which starts loading
    // now: the first time a card with a signature appears in the game.
    seenBefore = seenNow;
    seenNow = new Map();
    const shown = visibleCards(view);
    for (const c of shown) seenNow.set(c.instance_id, c);
    graveBefore = graveNow;
    graveNow = graveyardCards(view);
    signatures.notice(shown);
    // A signature's tint lasts until the turn ends (fx/tints.ts).
    syncTableTints(view.id, view.turn?.number ?? 0);

    if (r.fresh.length === 0) return;

    const s = get(settings);
    const motion = effectMotion({
      enabled: s.animations.enabled,
      effects: s.animations.effects,
      reduceMotion: s.accessibility.reduceMotion,
      osReduceMotion: osReducesMotion(),
    });
    const visibility = typeof document === "undefined" ? undefined : document.visibilityState;
    // The strike layer crumbles combat's deaths itself (ADR 0134).
    const strikes = strikesScheduled(
      combatMotion({
        enabled: s.animations.enabled,
        combat: s.animations.combat,
        reduceMotion: s.accessibility.reduceMotion,
      }),
      visibility,
    );
    const generic = planEffects(r.fresh, log, {
      isPlaneswalker: isPW,
      combatMotion: strikes,
      isCommander: isCommanderNow,
      // #4427: a permanent sacrificed for its own ability has just left,
      // so its type line is the one it had before it went.
      typeLineOf: (id) =>
        (lastViews.get(id) ?? view.battlefield.cards.find((c) => c.instance_id === id))?.type_line,
    });
    // Signatures take the batches they name, once their scene has loaded.
    const found = planSignatures(r.fresh, { sceneOf: sceneOfCard });
    // The creatures a tinting scene left standing keep its stain this
    // turn, whatever the motion: it is a mark, not an animation.
    for (const sig of found) {
      const tint = SCENES[sig.scene].tint;
      if (!tint) continue;
      const survivors = view.battlefield.cards
        .filter((c) => isCreature(c) && !sig.took.includes(c.instance_id))
        .map((c) => c.instance_id);
      tintSurvivors(survivors, tint, view.id, view.turn?.number ?? 0);
    }
    const sp = withSignatures(generic, found, (s) => signatures.ready(s) !== null, motion);
    const plan = sp.plan;
    // #4879: the players this frame says were decked out. Their sound is
    // EliminationFx's; the ghost is drawn here, only with motion.
    const decked = deckedSeats(r.fresh)
      .map((i) => view.seats[i])
      .filter((seat) => seat !== undefined);
    if (
      plan.cues.length === 0 &&
      plan.wipes.length === 0 &&
      sp.sigs.length === 0 &&
      decked.length === 0
    )
      return;

    const at = Date.now();
    // animations.speed, compressed on a busy table: many cues close
    // together each play shorter (tableEffects.ts busyScale). A decked-out
    // cue counts as a signature does.
    const speed =
      s.animations.speed * busy.frame(headlines(plan, sp.sigs.length + decked.length), at);
    const gameID = view.id;
    // Sounds always play (a hidden tab included); motion only when the
    // page is showing and the board can be measured.
    const b = effectsScheduled(visibility) ? boardSize() : null;

    // A sound at `delay` after the frame; one that is late is dropped, as a
    // late combat strike is (ADR 0134 §4).
    const sound = (
      name: Parameters<typeof play>[0],
      delay: number,
      volume: number,
      batch: number | null,
      soundKey?: string,
    ) => {
      const fire = () => play(name, { volume, batch: soundBatch(gameID, batch, soundKey) });
      if (delay <= 0) fire();
      else
        later(() => {
          if (!effectLate(at, delay, Date.now())) fire();
        }, delay);
    };

    for (const sig of sp.sigs) sound(SCENES[sig.scene].sound, 0, 1, sig.batch);

    // An exile wipe's cards are drawn into one vortex at the board's centre.
    const vortexBatches = new Set(plan.wipes.filter((w) => w.tone === "cold").map((w) => w.batch));
    for (const w of plan.wipes) {
      if (!sp.silencedWipes.has(w.batch)) sound("board_wipe", 0, 1, w.batch);
      if (!b) continue;
      const t = wipeTiming(motion, speed, w.tone);
      add({
        ...blank,
        kind: "wipe",
        motion,
        light: false,
        delay: 0,
        ms: t.ms,
        tone: w.tone,
        board: b.size,
        seed: `wipe-${w.batch}`,
        count: w.count,
        badge: L.effectBadgeWipe(w.count),
      });
    }

    let drawn = 0;
    for (const cue of plan.cues) {
      const t =
        (sp.inScene.has(cue.id) ? signatureCueTiming(cue, motion, speed) : null) ??
        cueTiming(cue, motion, speed);
      const snd = cueSound(cue);
      if (snd && !sp.silencedCues.has(cue.id)) {
        sound(snd.name, t.delayMs, snd.volume, cue.batch, cue.soundKey);
        // A sacrifice's knife has a sound of its own, as it lands.
        if (cue.kind === "sacrifice") {
          const knife =
            motion === "still" ? 0 : t.delayMs + sacrificeBeats(cue.weight === "light").stab * t.ms;
          sound(SACRIFICE_KNIFE_SOUND, knife, snd.volume, cue.batch);
        }
      }
      if (!b) continue;
      // A wipe says it once (its badge); its cards add nothing still.
      if (motion === "still" && cue.weight === "light") continue;
      if (++drawn > EFFECT_CUE_CAP) continue;
      const vortex =
        cue.kind === "exile" &&
        cue.weight === "light" &&
        cue.batch !== null &&
        vortexBatches.has(cue.batch) &&
        !sp.inScene.has(cue.id);
      drawCue(cue, motion, t, b, vortex);
    }

    // Drawn last, so a scene is over the cards it takes.
    if (b) for (const sig of sp.sigs) drawSignature(sig, motion, speed, b);
    if (b) for (const seat of decked) drawDecked(seat, motion, speed, b);
  }

  // ---- Decked out (#4879) ----

  function boxIn(el: Element, origin: { left: number; top: number }): Box {
    const r = el.getBoundingClientRect();
    return { x: r.left - origin.left, y: r.top - origin.top, w: r.width, h: r.height };
  }

  function byOwner(selector: string, attr: string, seatID: string): HTMLElement | null {
    if (!boardEl) return null;
    for (const el of boardEl.querySelectorAll<HTMLElement>(selector)) {
      if (el.getAttribute(attr) === seatID && hasSize(el)) return el;
    }
    return null;
  }

  function drawDecked(
    seat: GameView["seats"][number],
    motion: EffectMotion,
    speed: number,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): void {
    const t = deckedTiming(motion, speed);
    if (!t || !boardEl) return;
    // The mat is the seat's panel; a collapsed seat has none, and keeps
    // its stone alone.
    const pile = byOwner('[data-pile="library"]', "data-pile-owner", seat.id);
    const panel =
      pile?.closest<HTMLElement>(".seat-panel") ??
      findSeatAnchor(boardEl, seat.id, { accept: hasSize })?.closest<HTMLElement>(".seat-panel") ??
      null;
    if (!panel || !hasSize(panel)) return;
    const mat = boxIn(panel, b.rect);
    const matRect = panel.getBoundingClientRect();
    // No pile drawn: the ghost rises from the mat's lower middle.
    const lib = pile ? boxIn(pile, matRect) : { x: mat.w / 2 - 24, y: mat.h * 0.7, w: 48, h: 66 };
    const handEl = byOwner("[data-hand-owner]", "data-hand-owner", seat.id);
    const hand = handEl
      ? boxIn(handEl, b.rect)
      : { x: mat.x + mat.w * 0.25, y: mat.y + mat.h * 0.8, w: mat.w * 0.5, h: mat.h * 0.2 };

    // Faces where this player can see them (their own hand), backs
    // otherwise, from the frame before the cards left.
    const now = seat.hand;
    const before = lastHands.get(seat.id);
    const shown = (now?.cards?.length ?? 0) > 0 ? now.cards : (before?.cards ?? []);
    const count = Math.max(now?.count ?? 0, before?.count ?? 0, shown.length);
    const flings = deckedFlings(seat.id, count);
    const ch = Math.max(36, Math.min(120, hand.h * 0.85));
    const cw = Math.round(ch * (63 / 88));
    const cards = flings.map((f, i) => {
      const c = shown[i];
      const src = c && !showsCardBack(c) ? cardImageURL(c, "small") : "/card-back-small.jpg";
      return { src: src ?? "/card-back-small.jpg", w: cw, h: Math.round(ch), f };
    });

    const key = `decked-${++serial}`;
    decks = [
      ...decks,
      {
        key,
        seat: seat.id,
        ms: t.ms,
        mat,
        lib,
        hand,
        skull: Math.min(mat.w, mat.h / 1.2) * 0.96,
        diag: Math.hypot(b.size.w, b.size.h),
        wisps: deckedWisps(seat.id),
        cards,
      },
    ];
    later(() => {
      decks = decks.filter((d) => d.key !== key);
    }, t.ms + 80);
  }

  function deckedOf(node: HTMLElement, d: Decked) {
    const runs = deckedMotion(node, {
      ms: d.ms,
      lib: d.lib,
      mat: { w: d.mat.w, h: d.mat.h },
      skull: d.skull,
      diag: d.diag,
      wisps: d.wisps,
      flings: d.cards.map((c) => c.f),
    });
    return {
      destroy: () => {
        for (const r of runs) r.cancel();
        cancelTree(node);
      },
    };
  }

  // ---- Signatures ----

  function sceneCard(id: string, b: { rect: DOMRect; size: { w: number; h: number } }) {
    const where = cardTile(id, b);
    if (!where) return null;
    return {
      id,
      x: where.box.x,
      y: where.box.y,
      width: where.width,
      height: where.height,
      rot: where.rot,
      src: imageFor(id),
    } satisfies SceneCard;
  }

  // Where an entering card lands: its tile, measured once its entrance
  // has settled, as a planeswalker's ring is.
  function landingEl(sig: SignatureCue): HTMLElement | null {
    if (!boardEl) return null;
    return findCardAnchor(boardEl, sig.cardID, { accept: hasSize });
  }

  // The players' seats (their avatar circles), each marked a foe or not of
  // the card's controller: a scene may aim at one (Bolas breathes fire).
  function sceneSeats(cardID: string, b: { rect: DOMRect }): SceneSeat[] {
    if (!boardEl) return [];
    const controller = view.battlefield.cards.find((c) => c.instance_id === cardID)?.controller;
    const out: SceneSeat[] = [];
    for (const el of boardEl.querySelectorAll<HTMLElement>("[data-seat-id]")) {
      const id = el.dataset.seatId;
      if (!id || !hasSize(el) || out.some((s) => s.id === id)) continue;
      const r = el.getBoundingClientRect();
      out.push({
        id,
        x: r.left + r.width / 2 - b.rect.left,
        y: r.top + r.height / 2 - b.rect.top,
        r: Math.min(r.width, r.height) / 2,
        foe: !!controller && id !== controller,
      });
    }
    return out;
  }

  function landingOf(el: HTMLElement, sig: SignatureCue, b: { rect: DOMRect }): SceneCard {
    const t = tileOf(el, b.rect, { w: b.rect.width, h: b.rect.height });
    const src =
      el.querySelector<HTMLImageElement>("img")?.getAttribute("src") ?? imageFor(sig.cardID);
    return {
      id: sig.cardID,
      x: t.box.x,
      y: t.box.y,
      width: t.width,
      height: t.height,
      rot: t.rot,
      src,
    };
  }

  function drawSignature(
    sig: SignatureCue,
    motion: EffectMotion,
    speed: number,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): void {
    const spec = SCENES[sig.scene];
    const t = signatureTiming(spec, motion, speed);
    if (!t) return;
    const cards = sig.took.map((id) => sceneCard(id, b)).filter((c) => c !== null);
    if (spec.needsCards && cards.length === 0) return;
    const f: Omit<Fx, "key"> = {
      ...blank,
      kind: "signature",
      motion,
      light: false,
      delay: 0,
      ms: t.ms,
      board: b.size,
      seed: sig.id,
      cardID: sig.cardID,
      scene: sig.scene,
      sceneCards: cards,
      landing: null,
      k: t.k,
    };
    if (!sig.lands) {
      add(f);
      return;
    }
    const el = landingEl(sig);
    if (!el) {
      // Nowhere to land: the scene plays what it can without (a shadow).
      add(f);
      return;
    }
    void settled(el).then(() => {
      if (destroyed) return;
      const now = boardSize();
      if (!now) return;
      add({
        ...f,
        board: now.size,
        landing: landingOf(el, sig, now),
        seats: sceneSeats(sig.cardID, now),
      });
    });
  }

  // The first frame primes; a prime-key change primes the next one. The
  // key effect is declared first, so a key and a log that change together
  // prime that same frame.
  $effect(() => {
    void beatsPrimeKey;
    untrack(() => {
      reprimeNext = true;
    });
  });
  $effect(() => {
    const log = view.log;
    untrack(() => frame(log));
  });

  // Remember every battlefield card as of the latest frame: a permanent
  // that left is drawn from the frame before it went, and a planeswalker
  // is known for one by it.
  $effect(() => {
    const cards = view.battlefield.cards;
    untrack(() => {
      for (const c of cards) lastViews.set(c.instance_id, c);
    });
  });

  // Remember each seat's hand as of the latest frame, after frame() has
  // read the one before (#4879: a decked player's hand is flung from it).
  $effect(() => {
    const seats = view.seats;
    untrack(() => {
      for (const seat of seats ?? []) {
        if (seat.eliminated) continue;
        lastHands.set(seat.id, {
          count: seat.hand?.count ?? 0,
          cards: [...(seat.hand?.cards ?? [])],
        });
      }
    });
  });

  // Measure on every frame (after layout) and whenever the board or the
  // expanded board changes size, as the strike layer does.
  $effect(() => {
    void view;
    void $boardExpandLayout;
    if (!boardEl) return;
    // What is new in this frame is measured now as well (#4410): a
    // permanent or spell that goes in the next frame, before the
    // animation frame below has run, still has a box to be drawn from.
    untrack(() => measureBoxes(true));
    if (typeof requestAnimationFrame === "undefined") return;
    let raf = requestAnimationFrame(measure);
    if (typeof ResizeObserver === "undefined") return () => cancelAnimationFrame(raf);
    const obs = new ResizeObserver(() => {
      cancelAnimationFrame(raf);
      raf = requestAnimationFrame(measure);
    });
    obs.observe(boardEl);
    return () => {
      cancelAnimationFrame(raf);
      obs.disconnect();
    };
  });

  // Skip: a click anywhere or Escape, while a cue is playing. Capture and
  // passive, so the click still reaches what it was aimed at, and Escape
  // still closes whatever else listens for it.
  $effect(() => {
    if ((fx.length === 0 && decks.length === 0) || typeof window === "undefined") return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") clearAll();
    };
    const onPointer = () => clearAll();
    window.addEventListener("keydown", onKey, true);
    window.addEventListener("pointerdown", onPointer, { capture: true, passive: true });
    return () => {
      window.removeEventListener("keydown", onKey, true);
      window.removeEventListener("pointerdown", onPointer, true);
    };
  });

  onDestroy(() => {
    destroyed = true;
    for (const t of timers) clearTimeout(t);
    timers.clear();
  });

  // ---- Motion ----

  function cancelTree(node: HTMLElement): void {
    cancelAnimations(node);
    for (const el of node.querySelectorAll("*")) cancelAnimations(el);
  }

  // One cue's motion, started when it mounts (lib/effectCueMotion.ts).
  // Its removal is on a timer (add), so the tween is presentation only.
  // A still cue starts none.
  function motionOf(node: HTMLElement, f: Fx) {
    if (f.motion === "still") return {};
    if (f.kind === "dies") {
      // The existing crumble (ADR 0134), gated by animations.effects.
      const face = node.querySelector<HTMLElement>(".face");
      if (face) {
        void crumble(face, crumbleShards(f.cardID || f.seed, f.width, f.height, f.rot), {
          gate: "effects",
          ms: f.ms,
          delay: f.delay,
        });
      }
    } else if (f.kind === "token_create") {
      // The token's own tile pops in; only this run is cancelled with
      // the cue, never another animation on the tile.
      const tile = boardEl ? findCardAnchor(boardEl, f.cardID, { accept: hasSize }) : null;
      const pop = tile ? tokenPop(tile, f) : null;
      cueMotion(node, f);
      return {
        destroy: () => {
          pop?.cancel();
          cancelTree(node);
        },
      };
    } else if (f.kind === "commander_return") {
      // #4443: its card in the command zone waits, hidden, for the copy;
      // only this run is cancelled with the cue.
      const card = commandCardOf(f);
      const held = card ? commandArrive(card, f) : null;
      cueMotion(node, f);
      return {
        destroy: () => {
          held?.cancel();
          cancelTree(node);
        },
      };
    } else if (f.kind === "attach_equip" || f.kind === "attach_aura") {
      // #4444: the attachment's own tile slides onto its host; only this
      // run is cancelled with the cue (a token's pop on it is its own).
      const tile = boardEl ? findCardAnchor(boardEl, f.cardID, { accept: cmdLive }) : null;
      const slide = tile ? attachSlide(tile, f) : null;
      cueMotion(node, f);
      return {
        destroy: () => {
          slide?.cancel();
          cancelTree(node);
        },
      };
    } else if (f.kind === "transform") {
      // #4448: the permanent's own tile waits, hidden, under its flip;
      // only this run is cancelled with the cue.
      const tile = boardEl ? findCardAnchor(boardEl, f.cardID, { accept: fieldTile }) : null;
      const held = tile ? flipHold(tile, f) : null;
      cueMotion(node, f);
      return {
        destroy: () => {
          held?.cancel();
          cancelTree(node);
        },
      };
    } else if (f.kind === "reanimate") {
      // #4446: each returned card's own tile waits, hidden, for its copy
      // to land; only these runs are cancelled with the cue.
      const ids = f.tileIDs ?? [];
      const held = ids.flatMap((id, i) => {
        const tile = id && boardEl ? findCardAnchor(boardEl, id, { accept: fieldTile }) : null;
        return tile ? [riseArrive(tile, f, pileLandAt(ids.length, i, f.sweep === true))] : [];
      });
      cueMotion(node, f);
      return {
        destroy: () => {
          for (const h of held) h.cancel();
          cancelTree(node);
        },
      };
    } else {
      cueMotion(node, f);
    }
    return { destroy: () => cancelTree(node) };
  }

  // A signature's scene, started when it mounts: the kit's parts draw into
  // this node, and unmounting it (the cue's end, a skip) cancels them all.
  function sceneOf(node: HTMLElement, f: Fx) {
    const mod = f.scene ? signatures.ready(f.scene) : null;
    if (mod && f.motion !== "still") {
      const k = f.k ?? 1;
      try {
        mod.play({
          layer: node,
          board: f.board,
          seed: f.seed,
          short: f.motion === "short",
          t: (ms) => ms * k,
          cards: f.sceneCards ?? [],
          landing: f.landing ?? null,
          seats: f.seats ?? [],
        });
      } catch {
        // A scene that fails draws nothing; the cards' own cues still play.
      }
    }
    return { destroy: () => cancelTree(node) };
  }

  // The jagged cracks of a counterspell's shatter, and the glass.
  function burstOf(f: Fx) {
    return counterBurst(f.seed, f.width, f.height);
  }
  // The specks a fizzling spell comes apart into (#4450).
  function fizzleOf(f: Fx) {
    return fizzleMotes(f.seed, f.width, f.height);
  }
  // The crack across a planeswalker that dies (a fixed, readable shape).
  const PW_CRACK = "M50 0 L44 22 L56 40 L42 62 L54 80 L48 100";

  // A card that breaks (a destroy, a sacrifice's end, a destroy turned to
  // exile), the specks an exile's void pulls in, and a wipe's wind cuts.
  function shatterOf(f: Fx) {
    return shatterPlan(f.cardID || f.seed, f.width, f.height, {
      rot: f.rot,
      light: f.light || f.kind === "exile",
    });
  }
  // A permanent sacrificed for its own ability: its crack and its halves (#4427).
  function crackOf(f: Fx) {
    return crackPlan(f.cardID || f.seed, f.width, f.height, f.rot);
  }
  function motesOf(f: Fx) {
    return voidMotes(f.seed, f.width, f.height, f.light);
  }
  function cutsOf(f: Fx) {
    return windCuts(f.seed, Math.max(5, Math.min(8, 4 + f.count)));
  }
  // A shard that breaks has these; an exile's crumble moves them a little.
  const breaks = (k: FxKind) => k === "destroy" || k === "sacrifice";
  // The arms of an exile wipe's vortex.
  const SWIRL = swirlArms();

  // The open palm that stops a countered spell (a 0..100 box): its palm,
  // four fingers and a thumb (drawn as round-capped strokes), and a crease.
  const PALM_BODY = "M31 52 Q31 45 38 45 L64 45 Q71 45 71 52 L71 74 Q71 93 51 94 Q32 94 31 76 Z";
  const PALM_DIGITS = [
    "M38 50 L36.5 17",
    "M49 48 L49 10",
    "M60 50 L62 15",
    "M69.5 56 L73.5 28",
    "M34 72 Q25 63 17 52",
  ];
  const PALM_CREASE = "M37 68 Q50 61 65 66";
</script>

<div class="fx-layer" aria-hidden="true" data-table-effects>
  {#each fx as f (f.key)}
    {#if f.kind === "wipe"}
      <div
        class="wipe"
        data-effect="board_wipe"
        data-effect-motion={f.motion}
        data-tone={f.tone}
        use:motionOf={f}
      >
        {#if f.motion === "still"}
          <!-- Reduced motion (ADR 0209 §1): a brief static tint and the badge. -->
          <div class="tint still-tint"></div>
          <span class="badge" data-effect-badge style:left="50%" style:top="50%">{f.badge}</span>
        {:else if f.tone === "cold"}
          <!-- D11: the board swirls into a white void with a small darkness at its heart. -->
          {@const side = Math.hypot(f.board.w, f.board.h) * 1.1}
          {@const voidSide = Math.min(f.board.w, f.board.h) * 0.62}
          <div class="tint m"></div>
          <div class="haze m"></div>
          <svg
            class="swirl m"
            viewBox="-100 -100 200 200"
            style:width={`${side}px`}
            style:height={`${side}px`}
            style:left={`${f.board.w / 2 - side / 2}px`}
            style:top={`${f.board.h / 2 - side / 2}px`}
          >
            <defs>
              <radialGradient
                id={`swirl-${f.key}`}
                cx="0"
                cy="0"
                r="100"
                gradientUnits="userSpaceOnUse"
              >
                <stop offset="0" stop-color="#ffffff" stop-opacity="0.95" />
                <stop offset="0.45" stop-color="#d8ecff" stop-opacity="0.6" />
                <stop offset="1" stop-color="#7fb0ff" stop-opacity="0" />
              </radialGradient>
            </defs>
            {#each SWIRL as d, i (i)}
              <path {d} fill={`url(#swirl-${f.key})`} />
            {/each}
          </svg>
          <div
            class="whitevoid m"
            style:width={`${voidSide}px`}
            style:height={`${voidSide}px`}
            style:left={`${f.board.w / 2}px`}
            style:top={`${f.board.h / 2}px`}
          ></div>
          <div
            class="core m"
            style:width={`${voidSide * 0.16}px`}
            style:height={`${voidSide * 0.16}px`}
            style:left={`${f.board.w / 2}px`}
            style:top={`${f.board.h / 2}px`}
          ></div>
        {:else if f.tone === "dark"}
          <!-- D10: quick cuts slicing across the battlefield, like horizontal winds. -->
          <div class="tint m"></div>
          {#each cutsOf(f) as c, i (i)}
            <div
              class="cut m"
              data-at={c.at}
              data-len={c.len}
              data-dir={c.dir}
              style:top={`${c.y}%`}
              style:rotate={`${c.tilt}deg`}
              style:--cut-thick={`${c.thick}px`}
            >
              <div class="scar" style:transform-origin={c.dir > 0 ? "0 50%" : "100% 50%"}></div>
              <div class="gust" style:width={`${Math.max(220, f.board.w * 0.6)}px`}></div>
              <div class="blade" style:width={`${Math.max(160, f.board.w * 0.42)}px`}></div>
            </div>
          {/each}
        {:else}
          <div class="tint m"></div>
          <div class="sweep m"></div>
        {/if}
      </div>
    {:else if f.kind === "signature"}
      <!-- A signature's scene (fx/scenes/): never still, so never a badge. -->
      <div
        class="sig m"
        data-effect="signature"
        data-signature={f.scene}
        data-effect-motion={f.motion}
        use:sceneOf={f}
      ></div>
    {:else if f.motion === "still"}
      {#if f.badge}
        <span
          class="badge"
          data-effect={f.kind}
          data-effect-motion="still"
          data-effect-badge
          style:left={`${f.x}px`}
          style:top={`${f.y}px`}>{f.badge}</span
        >
      {/if}
    {:else if f.kind === "counter"}
      <!-- D7: a palm stops the spell, and the spell cracks. -->
      {@const burst = burstOf(f)}
      <div
        class="fx m"
        data-effect="counter"
        data-effect-motion={f.motion}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        <div class="spot" style:width={`${f.width}px`} style:height={`${f.height}px`}>
          <div class="flash"></div>
          <svg class="cracks" viewBox="0 0 100 100" preserveAspectRatio="none">
            {#each burst.cracks as d, i (i)}
              <path {d} pathLength="1" />
            {/each}
          </svg>
          {#each burst.shards as s, i (i)}
            <div
              class="glass"
              data-dx={s.dx}
              data-dy={s.dy}
              data-rot={s.rot}
              style:left={`${s.x}%`}
              style:top={`${s.y}%`}
              style:width={`${s.size}px`}
              style:height={`${s.size}px`}
            ></div>
          {/each}
          <div class="shock"></div>
          <svg class="palm" viewBox="0 0 100 100">
            <g opacity="0.9">
              <path class="palm-body" d={PALM_BODY} />
              {#each PALM_DIGITS as d, i (i)}
                <path class="palm-digit" {d} />
              {/each}
            </g>
            <path class="palm-crease" d={PALM_CREASE} />
          </svg>
        </div>
      </div>
    {:else if f.kind === "fizzle"}
      <!-- #4450: no palm and no glass: the spell thins into the air. -->
      <div
        class="fx m"
        data-effect="fizzle"
        data-effect-motion={f.motion}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        <div class="spot" style:width={`${f.width}px`} style:height={`${f.height}px`}>
          <div class="fmist"></div>
          <div class="fedge"></div>
          {#each fizzleOf(f) as s, i (i)}
            <div
              class="fmote"
              data-dx={s.dx}
              data-dy={s.dy}
              data-at={s.at}
              style:left={`${s.x}%`}
              style:top={`${s.y}%`}
              style:width={`${s.size}px`}
              style:height={`${s.size}px`}
            ></div>
          {/each}
        </div>
      </div>
    {:else if f.kind === "token_create"}
      <!-- #4441: a soft glow over the new token as its tile pops in. -->
      <div
        class="fx m"
        data-effect="token_create"
        data-effect-motion={f.motion}
        data-effect-card={f.cardID}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        <div
          class="tglow"
          style:width={`${f.width * 1.6}px`}
          style:height={`${f.height * 1.4}px`}
          style:--fx-rot={`${f.rot}deg`}
        ></div>
      </div>
    {:else if f.kind === "counters"}
      <!-- #4442: a ring over the power/toughness badge, and a tick that
           rises (+1/+1) or falls (-1/-1) above it. -->
      <div
        class="fx m"
        data-effect="counters"
        data-effect-motion={f.motion}
        data-effect-card={f.cardID}
        data-effect-tone={f.counterTone}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        <div
          class="cpulse"
          style:width={`${f.width + 10}px`}
          style:height={`${f.height + 8}px`}
        ></div>
        {#if f.counterTone !== "other"}
          <div class="ctick" style:top={`${-(f.height / 2 + 9)}px`}></div>
        {/if}
      </div>
    {:else if f.kind === "commander_cast"}
      <!-- #4443: golden light rising from the command zone to the stack. -->
      {@const beam = reachOf(f)}
      <div
        class="fx m"
        data-effect="commander_cast"
        data-effect-motion={f.motion}
        data-effect-card={f.cardID}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        <div
          class="cmd-bloom"
          style:width={`${f.width * 1.8}px`}
          style:height={`${f.height * 1.5}px`}
        ></div>
        <div
          class="cmd-beam"
          style:width={`${Math.max(16, f.width * 0.42)}px`}
          style:height={`${beam.len}px`}
          style:rotate={`${beam.turn}deg`}
        ></div>
        {#each CMD_MOTES as i (i)}
          <i class="cmd-mote"></i>
        {/each}
        <div
          class="cmd-crown"
          style:left={`${f.fly.x}px`}
          style:top={`${f.fly.y}px`}
          style:width={`${(f.dest?.w ?? f.width) + 14}px`}
          style:height={`${(f.dest?.h ?? f.height) + 14}px`}
        ></div>
      </div>
    {:else if f.kind === "commander_return"}
      <!-- #4443: the commander drifts home with a soft trail. -->
      {@const trail = reachOf(f)}
      <div
        class="fx m"
        data-effect="commander_return"
        data-effect-motion={f.motion}
        data-effect-card={f.cardID}
        data-effect-lands={f.landsOnCard ? "card" : "seat"}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        <div
          class="cmd-streak"
          style:width={`${Math.max(10, f.width * 0.22)}px`}
          style:height={`${trail.len}px`}
          style:rotate={`${trail.turn}deg`}
        ></div>
        {#each [0, 1] as g (g)}
          <div class="copy cmd-ghost">
            <div
              class="face"
              style:width={`${f.width}px`}
              style:height={`${f.height}px`}
              style:--fx-rot={`${f.rot}deg`}
            >
              {#if f.src}<img src={f.src} alt="" decoding="async" draggable="false" />{/if}
            </div>
          </div>
        {/each}
        <div class="copy cmd-copy">
          <div
            class="face"
            style:width={`${f.width}px`}
            style:height={`${f.height}px`}
            style:--fx-rot={`${f.rot}deg`}
          >
            {#if f.src}<img src={f.src} alt="" decoding="async" draggable="false" />{/if}
          </div>
        </div>
        <div
          class="cmd-home"
          style:left={`${f.fly.x}px`}
          style:top={`${f.fly.y}px`}
          style:width={`${(f.dest?.w ?? f.width) + 12}px`}
          style:height={`${(f.dest?.h ?? f.height) + 12}px`}
        ></div>
      </div>
    {:else if f.kind === "attach_equip"}
      <!-- #4444: the Equipment's tile slides under its host (attachSlide);
           a flash and sparks where the two meet, and a glint on the host. -->
      <div
        class="fx m"
        data-effect="attach_equip"
        data-effect-motion={f.motion}
        data-effect-card={f.cardID}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        <div class="att-rim" style:width={`${f.width}px`} style:height={`${f.height}px`}></div>
        <div class="att-seam" style:left={`${f.seam?.x ?? 0}px`} style:top={`${f.seam?.y ?? 0}px`}>
          <div class="att-flash"></div>
          {#each ATT_SPARKS as i (i)}
            <i class="att-spark"></i>
          {/each}
        </div>
      </div>
    {:else if f.kind === "attach_aura"}
      <!-- #4444: the Aura's tile slides under its host (attachSlide); a
           soft ring of light around the host, and glints at its edge. -->
      <div
        class="fx m"
        data-effect="attach_aura"
        data-effect-motion={f.motion}
        data-effect-card={f.cardID}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        <div
          class="att-halo"
          style:width={`${f.width + 10}px`}
          style:height={`${f.height + 10}px`}
        ></div>
        {#each ATT_GLINTS as [gx, gy], i (i)}
          <i class="att-glint" style:left={`${gx * f.width}px`} style:top={`${gy * f.height}px`}
          ></i>
        {/each}
      </div>
    {:else if f.kind === "mill" || f.kind === "discard"}
      <!-- #4445: the cards slide from the library (turning face up) or are
           tossed from the hand onto the graveyard pile, a stagger apart;
           five or more go as one sweep, with a streak behind them. -->
      {@const path = reachOf(f)}
      <div
        class="fx m"
        data-effect={f.kind}
        data-effect-motion={f.motion}
        data-effect-card={f.cardID}
        data-effect-count={f.count}
        data-effect-sweep={f.sweep ? "" : undefined}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        {#if f.sweep}
          <div
            class="pile-streak"
            data-kind={f.kind}
            style:width={`${Math.max(12, f.width * 0.5)}px`}
            style:height={`${path.len}px`}
            style:rotate={`${path.turn}deg`}
          ></div>
        {/if}
        {#each f.slides ?? [] as src, i (i)}
          <div class="copy pile-card">
            {#if f.kind === "mill" && src}
              <div class="flip">
                <div class="face" style:width={`${f.width}px`} style:height={`${f.height}px`}>
                  <img {src} alt="" decoding="async" draggable="false" />
                </div>
                <div
                  class="face back"
                  style:width={`${f.width}px`}
                  style:height={`${f.height}px`}
                ></div>
              </div>
            {:else}
              <div
                class="face"
                class:back-only={!src}
                style:width={`${f.width}px`}
                style:height={`${f.height}px`}
              >
                {#if src}<img {src} alt="" decoding="async" draggable="false" />{/if}
              </div>
            {/if}
          </div>
        {/each}
        <div
          class="pile-dust"
          data-kind={f.kind}
          style:left={`${f.fly.x}px`}
          style:top={`${f.fly.y}px`}
          style:width={`${f.width * 1.7}px`}
          style:height={`${f.height * 1.1}px`}
        ></div>
      </div>
    {:else if f.kind === "reanimate"}
      <!-- #4446: each card rises out of its graveyard pile in a dark-green
           glow and goes to where it went, a stagger apart; five or more
           go as one sweep. -->
      <div
        class="fx m"
        data-effect="reanimate"
        data-effect-motion={f.motion}
        data-effect-card={f.cardID}
        data-effect-count={f.count}
        data-effect-to={f.reanimateTo}
        data-effect-sweep={f.sweep ? "" : undefined}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        {#each f.paths ?? [] as p, i (i)}
          <div
            class="rise-glow"
            style:left={`${p.x}px`}
            style:top={`${p.y}px`}
            style:width={`${f.width * 1.9}px`}
            style:height={`${f.height * 1.25}px`}
          ></div>
        {/each}
        {#each f.paths ?? [] as p, i (i)}
          {@const src = f.slides?.[i] ?? null}
          <div class="copy pile-card" style:left={`${p.x}px`} style:top={`${p.y}px`}>
            <div
              class="rise-aura"
              style:width={`${f.width}px`}
              style:height={`${f.height}px`}
            ></div>
            {#if f.flips?.[i] && src}
              <div class="flip" data-flip="down">
                <div class="face" style:width={`${f.width}px`} style:height={`${f.height}px`}>
                  <img {src} alt="" decoding="async" draggable="false" />
                </div>
                <div
                  class="face back"
                  style:width={`${f.width}px`}
                  style:height={`${f.height}px`}
                ></div>
              </div>
            {:else}
              <div
                class="face"
                class:back-only={!src}
                style:width={`${f.width}px`}
                style:height={`${f.height}px`}
              >
                {#if src}<img {src} alt="" decoding="async" draggable="false" />{/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {:else if f.kind === "drain_table"}
      <!-- #4450: a thin dark line from the source to each player. -->
      <div
        class="fx m"
        data-effect="drain_table"
        data-effect-motion={f.motion}
        style:left="0px"
        style:top="0px"
        use:motionOf={f}
      >
        <svg
          class="drain-svg"
          width={f.board.w}
          height={f.board.h}
          viewBox={`0 0 ${f.board.w} ${f.board.h}`}
          aria-hidden="true"
        >
          {#each f.lines ?? [] as l, i (i)}
            <line class="drain-line" pathLength="1" x1={l.x1} y1={l.y1} x2={l.x2} y2={l.y2} />
          {/each}
        </svg>
        <div class="drain-source" style:left={`${f.x}px`} style:top={`${f.y}px`}></div>
      </div>
    {:else if f.kind === "monarch" || f.kind === "initiative"}
      <!-- #4450: a short glint on the new holder's panel. -->
      <div
        class="fx m"
        data-effect={f.kind}
        data-effect-motion={f.motion}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        <div class="glint-ring" style:width={`${f.width}px`} style:height={`${f.height}px`}></div>
        {#each GLINT_SPARKS as s, i (i)}
          <div class="glint-spark" data-dx={s.dx} data-dy={s.dy}></div>
        {/each}
      </div>
    {:else if f.kind === "library_search"}
      <!-- #4447: a few card backs flick up off the searched library's pile
           one after another and settle, over a cool light. Backs only. -->
      <div
        class="fx m"
        data-effect="library_search"
        data-effect-motion={f.motion}
        data-effect-library={f.library}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        <div
          class="riffle-glow"
          style:width={`${f.width * 2}px`}
          style:height={`${f.height * 1.5}px`}
        ></div>
        {#each RIFFLE.slice(0, f.count) as i (i)}
          <div class="copy riffle-card" style:transform-origin={`0px ${f.height / 2}px`}>
            <div
              class="face back-only"
              style:width={`${f.width}px`}
              style:height={`${f.height}px`}
            ></div>
          </div>
        {/each}
      </div>
    {:else if f.kind === "transform"}
      <!-- #4448: the card flips on its own vertical axis to its other face,
           over its tile, which waits hidden under it. -->
      <div
        class="fx m"
        data-effect="transform"
        data-effect-motion={f.motion}
        data-effect-card={f.cardID}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        <div class="turn" style:rotate={`${f.rot}deg`}>
          <div
            class="turn-rim"
            style:width={`${f.width + 10}px`}
            style:height={`${f.height + 10}px`}
          ></div>
          <div class="turn-card">
            <div class="face turn-from" style:width={`${f.width}px`} style:height={`${f.height}px`}>
              {#if f.from}<img src={f.from} alt="" decoding="async" draggable="false" />{/if}
            </div>
            <div class="face turn-to" style:width={`${f.width}px`} style:height={`${f.height}px`}>
              {#if f.src}<img src={f.src} alt="" decoding="async" draggable="false" />{/if}
            </div>
          </div>
        </div>
      </div>
    {:else if f.kind === "day_night"}
      <!-- #4448: the table takes the sky's tint as it becomes day or night. -->
      <div
        class="sky m"
        data-effect="day_night"
        data-effect-motion={f.motion}
        data-sky={f.sky}
        use:motionOf={f}
      >
        <div class="sky-tint"></div>
        <div class="sky-band" style:width={`${Math.max(160, f.board.w * 0.45)}px`}></div>
        <div class="sky-orb" style:left={`${f.board.w / 2}px`}></div>
      </div>
    {:else if f.kind === "pw_enter"}
      <div
        class="fx m"
        data-effect="pw_enter"
        data-effect-motion={f.motion}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        <div
          class="glow"
          style:width={`${f.width * 1.5}px`}
          style:height={`${f.height * 1.5}px`}
        ></div>
        <div
          class="ring"
          style:width={`${f.width + 10}px`}
          style:height={`${f.height + 10}px`}
        ></div>
      </div>
    {:else if f.kind === "crack"}
      <!-- #4427: snapped along a crack like a dry branch, and fallen away; no altar. -->
      {@const ck = crackOf(f)}
      <div
        class="fx m"
        data-effect="crack"
        data-effect-motion={f.motion}
        data-effect-card={f.cardID}
        data-crack={f.crack ?? "other"}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        {#each ck.halves as h, i (i)}
          <div
            class="copy crack-half"
            data-dx={h.dx}
            data-dy={h.dy}
            data-rot={h.rot}
            style:transform-origin={`${h.ox}px ${h.oy}px`}
          >
            <div
              class="face"
              style:width={`${f.width}px`}
              style:height={`${f.height}px`}
              style:--fx-rot={`${f.rot}deg`}
              style:clip-path={h.clip}
            >
              {#if f.src}<img src={f.src} alt="" decoding="async" draggable="false" />{/if}
            </div>
          </div>
        {/each}
        <svg
          class="crack-line"
          viewBox={`0 0 ${f.width} ${f.height}`}
          width={f.width}
          height={f.height}
          style:--fx-rot={`${f.rot}deg`}
        >
          <path class="under" d={ck.path} pathLength="1" />
          <path class="over" d={ck.path} pathLength="1" />
        </svg>
        <div class="crack-puff" style:width={`${Math.max(f.width, f.height) * 0.9}px`}></div>
      </div>
    {:else}
      {@const broken = breaks(f.kind) || (f.kind === "exile" && f.fromDestroy)}
      {@const sh = broken ? shatterOf(f) : null}
      {@const reach = Math.max(f.width, f.height)}
      <div
        class="fx m"
        data-effect={f.kind}
        data-effect-motion={f.motion}
        data-effect-light={f.light ? "" : undefined}
        data-effect-card={f.cardID}
        data-effect-from-destroy={f.fromDestroy ? "" : undefined}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        {#if f.kind === "exile" && !f.pull}
          <!-- D3: a white void the card spins into. -->
          <div class="void" style:width={`${reach * 1.6}px`}></div>
          <div class="vring" style:width={`${reach * 1.25}px`}></div>
          {#each motesOf(f) as o, i (i)}
            <div class="orbit" data-at={o.at} data-angle={o.angle} data-spin={o.spin} data-r={o.r}>
              <i class="speck" style:width={`${o.size}px`} style:height={`${o.size}px`}></i>
            </div>
          {/each}
        {:else if f.kind === "sacrifice" && !f.light}
          <!-- D5: a quick altar the card is laid on. -->
          <svg
            class="altar"
            viewBox="0 0 140 64"
            style:width={`${f.width * 1.7}px`}
            style:left={`${-f.width * 0.85}px`}
            style:top={`${f.height * 0.12}px`}
          >
            <defs>
              <linearGradient id={`altar-top-${f.key}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0" stop-color="#6c6472" />
                <stop offset="1" stop-color="#4a4350" />
              </linearGradient>
              <linearGradient id={`altar-front-${f.key}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0" stop-color="#2d2733" />
                <stop offset="1" stop-color="#17131b" />
              </linearGradient>
            </defs>
            <ellipse class="altar-glow" cx="70" cy="30" rx="70" ry="22" />
            <path d="M16 12 L124 12 L138 28 L2 28 Z" fill={`url(#altar-top-${f.key})`} />
            <rect x="2" y="28" width="136" height="24" fill={`url(#altar-front-${f.key})`} />
            <path class="altar-edge" d="M2 28 L138 28" />
            <rect x="12" y="52" width="116" height="8" rx="1" fill="#120f15" />
            {#each [14, 126] as cx (cx)}
              <rect x={cx - 3} y="2" width="6" height="12" rx="1" fill="#d9cdb4" />
              <ellipse class="flame" {cx} cy="-2" rx="2.6" ry="5" />
            {/each}
          </svg>
        {:else if f.kind === "pw_leave"}
          <!-- D9: a rift the walker leaves the plane through. -->
          <div
            class="rift"
            style:width={`${f.width * 0.9}px`}
            style:height={`${f.height * 1.5}px`}
          ></div>
        {/if}
        {#if f.kind === "bounce"}
          {#each [0, 1] as g (g)}
            <div class="copy ghost">
              <div
                class="face"
                style:width={`${f.width}px`}
                style:height={`${f.height}px`}
                style:--fx-rot={`${f.rot}deg`}
              >
                {#if f.src}<img src={f.src} alt="" decoding="async" draggable="false" />{/if}
              </div>
            </div>
          {/each}
        {/if}
        <div class="copy">
          {#if f.kind === "bounce"}
            <!-- D6: turned over and shown to the table on the way home. -->
            <div
              class="halo"
              style:width={`${f.width + 12}px`}
              style:height={`${f.height + 12}px`}
              style:--fx-rot={`${f.rot}deg`}
            ></div>
            <div class="flip">
              <div
                class="face"
                style:width={`${f.width}px`}
                style:height={`${f.height}px`}
                style:--fx-rot={`${f.rot}deg`}
              >
                {#if f.src}<img src={f.src} alt="" decoding="async" draggable="false" />{/if}
              </div>
              <div
                class="face back"
                style:width={`${f.width}px`}
                style:height={`${f.height}px`}
                style:--fx-rot={`${f.rot}deg`}
              ></div>
            </div>
          {:else}
            {#if !(f.kind === "exile" && f.fromDestroy)}
              <div
                class="face"
                style:width={`${f.width}px`}
                style:height={`${f.height}px`}
                style:--fx-rot={`${f.rot}deg`}
              >
                {#if f.src}
                  <img src={f.src} alt="" decoding="async" draggable="false" />
                {/if}
                {#if f.kind === "pw_leave"}
                  <svg class="crack" viewBox="0 0 100 100" preserveAspectRatio="none">
                    <path d={PW_CRACK} pathLength="1" />
                  </svg>
                {/if}
              </div>
            {/if}
            {#if sh}
              <!-- D1: the card breaks from the impact at its centre outward. -->
              <div
                class="shatter"
                style:width={`${f.width}px`}
                style:height={`${f.height}px`}
                style:--fx-rot={`${f.rot}deg`}
              >
                {#each sh.shards as s, i (i)}
                  <div
                    class="shard"
                    data-at={s.at}
                    data-dx={s.dx}
                    data-dy={s.dy}
                    data-rot={s.rot}
                    style:clip-path={s.clip}
                    style:transform-origin={`${s.ox}px ${s.oy}px`}
                  >
                    {#if f.src}<img src={f.src} alt="" decoding="async" draggable="false" />{/if}
                  </div>
                {/each}
                <svg
                  class="crk"
                  viewBox={`0 0 ${f.width} ${f.height}`}
                  width={f.width}
                  height={f.height}
                >
                  {#each sh.spokes as d, i (i)}
                    <path class="spoke under" {d} pathLength="1" />
                    <path class="spoke over" {d} pathLength="1" />
                  {/each}
                  {#each sh.rings as d, i (i)}
                    <path class="rim under" {d} pathLength="1" />
                    <path class="rim over" {d} pathLength="1" />
                  {/each}
                </svg>
                <div class="flare" style:left={`${sh.ix}px`} style:top={`${sh.iy}px`}></div>
                {#each sh.motes as m, i (i)}
                  <i
                    class="mote"
                    data-at={m.at}
                    data-dx={m.dx}
                    data-dy={m.dy}
                    style:left={`${sh.ix}px`}
                    style:top={`${sh.iy}px`}
                    style:width={`${m.size}px`}
                    style:height={`${m.size}px`}
                  ></i>
                {/each}
              </div>
            {/if}
          {/if}
        </div>
        {#if f.kind === "exile" && !f.pull}
          <div class="pinch"></div>
        {:else if f.kind === "sacrifice"}
          <!-- D5: the knife that stabs the card. -->
          {@const kh = f.height * 0.95}
          {@const kw = kh * 0.24}
          <div
            class="knife"
            style:width={`${kw}px`}
            style:height={`${kh}px`}
            style:left={`${-kw / 2}px`}
            style:top={`${-kh}px`}
          >
            <svg viewBox="0 0 24 100" preserveAspectRatio="none">
              <defs>
                <linearGradient id={`blade-${f.key}`} x1="0" y1="0" x2="1" y2="0">
                  <stop offset="0" stop-color="#f6f8fc" />
                  <stop offset="0.5" stop-color="#c9cfdb" />
                  <stop offset="1" stop-color="#7d8597" />
                </linearGradient>
              </defs>
              <path d="M8 40 L16 40 L15.2 72 L12 100 L8.8 72 Z" fill={`url(#blade-${f.key})`} />
              <path d="M12 43 L12 86" stroke="rgba(255,255,255,0.55)" stroke-width="0.8" />
              <rect x="1.5" y="35" width="21" height="5" rx="2.2" fill="#8a6a34" />
              <rect x="1.5" y="35" width="21" height="1.6" rx="0.8" fill="#d6b56a" />
              <rect x="9" y="13" width="6" height="22" rx="2" fill="#3a2418" />
              {#each [17, 22, 27, 32] as y (y)}
                <path d={`M9 ${y} L15 ${y - 2}`} stroke="#5e3b26" stroke-width="1" />
              {/each}
              <circle cx="12" cy="8.5" r="5" fill="#8a6a34" />
              <circle cx="12" cy="8.5" r="2.1" fill="#b3122e" />
            </svg>
            <div class="glint"></div>
          </div>
        {/if}
      </div>
    {/if}
  {/each}
  <!-- #4879: decked out, the library's ghost. The mat box clips the pulse,
       the wisps and the skull to the player's own seat; the flung cards
       are outside it, so they can leave the board. -->
  {#each decks as d (d.key)}
    <div class="decked m" data-decked-out={d.seat} use:deckedOf={d}>
      <div
        class="dk-mat"
        style:left="{d.mat.x}px"
        style:top="{d.mat.y}px"
        style:width="{d.mat.w}px"
        style:height="{d.mat.h}px"
      >
        <span
          class="dk-pulse"
          style:left="{d.lib.x}px"
          style:top="{d.lib.y}px"
          style:width="{d.lib.w}px"
          style:height="{d.lib.h}px"
        ></span>
        {#each d.wisps as w, i (i)}
          <span
            class="dk-wisp"
            style:left="{d.lib.x + d.lib.w * w.x}px"
            style:top="{d.lib.y + d.lib.h * 0.35}px"
            style:width="{d.lib.w * w.size}px"
            style:height="{d.lib.w * w.size * 2.2}px"
          ></span>
        {/each}
        <div
          class="dk-skull"
          style:left="{d.mat.w / 2 - d.skull / 2}px"
          style:top="{d.mat.h / 2 - d.skull * 0.6}px"
          style:width="{d.skull}px"
          style:height="{d.skull * 1.2}px"
        >
          <svg viewBox="0 0 100 120" width="100%" height="100%">
            <defs>
              <radialGradient id="{d.key}-bone" cx="50%" cy="45%" r="60%">
                <stop offset="0%" stop-color="#f2f8ff" stop-opacity="0.95" />
                <stop offset="45%" stop-color="#b9d4f7" stop-opacity="0.8" />
                <stop offset="80%" stop-color="#3f74c6" stop-opacity="0.7" />
                <stop offset="100%" stop-color="#173d8a" stop-opacity="0.85" />
              </radialGradient>
              <filter id="{d.key}-glow" x="-200%" y="-200%" width="500%" height="500%">
                <feGaussianBlur stdDeviation="1.6" />
              </filter>
            </defs>
            <!-- The cranium and cheeks, then the jaw. -->
            <path
              d="M50 4C24 4 8 22 8 48c0 16 6 26 14 32l2 14c0 5 4 8 9 8h34c5 0 9-3 9-8l2-14c8-6 14-16 14-32C92 22 76 4 50 4z"
              fill="url(#{d.key}-bone)"
              stroke="#1b3f86"
              stroke-opacity="0.6"
              stroke-width="1.5"
            />
            <path
              d="M30 100v12c0 4 4 6 8 6h24c4 0 8-2 8-6v-12z"
              fill="url(#{d.key}-bone)"
              stroke="#1b3f86"
              stroke-opacity="0.6"
              stroke-width="1.5"
            />
            <!-- The eye sockets, and their small glowing pupils. -->
            <ellipse cx="33" cy="52" rx="12" ry="13" fill="#061433" fill-opacity="0.9" />
            <ellipse cx="67" cy="52" rx="12" ry="13" fill="#061433" fill-opacity="0.9" />
            <circle
              cx="34"
              cy="54"
              r="5.5"
              fill="#6fd3ff"
              opacity="0.45"
              filter="url(#{d.key}-glow)"
            />
            <circle
              cx="66"
              cy="54"
              r="5.5"
              fill="#6fd3ff"
              opacity="0.45"
              filter="url(#{d.key}-glow)"
            />
            <circle cx="34" cy="54" r="2.4" fill="#d8f7ff" />
            <circle cx="66" cy="54" r="2.4" fill="#d8f7ff" />
            <!-- The nasal cavity. -->
            <path d="M50 66l-6 14c2 3 5 2 6 0 1 2 4 3 6 0z" fill="#061433" fill-opacity="0.9" />
            <!-- The teeth, upper and lower. -->
            {#each [0, 1, 2, 3, 4, 5] as t (t)}
              <rect
                x={33.5 + t * 5.6}
                y="94"
                width="4.6"
                height="9"
                rx="1.2"
                fill="#e8f2ff"
                fill-opacity="0.85"
                stroke="#2a5297"
                stroke-opacity="0.5"
                stroke-width="0.6"
              />
              <rect
                x={33.5 + t * 5.6}
                y="104"
                width="4.6"
                height="8"
                rx="1.2"
                fill="#e8f2ff"
                fill-opacity="0.8"
                stroke="#2a5297"
                stroke-opacity="0.5"
                stroke-width="0.6"
              />
            {/each}
          </svg>
        </div>
      </div>
      {#each d.cards as c, i (i)}
        <img
          class="dk-card"
          src={c.src}
          alt=""
          draggable="false"
          style:left="{d.hand.x + d.hand.w * (0.15 + 0.7 * c.f.x)}px"
          style:top="{d.hand.y + d.hand.h / 2}px"
          style:width="{c.w}px"
          style:height="{c.h}px"
        />
      {/each}
    </div>
  {/each}
</div>

<style>
  .fx-layer {
    position: absolute;
    inset: 0;
    pointer-events: none;
    /* The table's z-index ladder (DESIGN.md §4, ADR 0209 §4): above the
       action dock (55), below the coach (58). */
    z-index: 57;
    overflow: hidden;
  }
  /* A zero-size anchor at the card's centre: the motion moves .copy, so
     the face's own transform (the tap rotation) is never overwritten. */
  .fx {
    position: absolute;
    width: 0;
    height: 0;
  }
  .copy {
    position: absolute;
    left: 0;
    top: 0;
    width: 0;
    height: 0;
  }
  .face {
    position: absolute;
    left: 0;
    top: 0;
    border-radius: 8px;
    overflow: hidden;
    background: #0d1220;
    border: 1px solid #0a0e1a;
    box-shadow: 0 10px 24px rgba(0, 0, 0, 0.55);
    box-sizing: border-box;
    transform: translate(-50%, -50%) rotate(var(--fx-rot, 0deg));
  }
  .face img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  /* The crumble (animations.ts crumble): the shards take over from the
     face's own art, frame and shadow, and are free to fall out of its box. */
  .face:global(.crumbling) {
    overflow: visible;
    background: transparent;
    border-color: transparent;
    box-shadow: none;
  }
  .face:global(.crumbling) > img {
    visibility: hidden;
  }
  .face :global(.shard) {
    position: absolute;
    inset: 0;
    border-radius: 8px;
    overflow: hidden;
    background: #0d1220;
    filter: grayscale(0.55) brightness(0.85);
  }
  .face :global(.shard img) {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  /* D1, the shatter: the pieces of the card, each a full-size copy of
     its art clipped to its outline, turned with the card about its
     centre. The face shows until the first piece lets go. */
  .shatter {
    position: absolute;
    left: 0;
    top: 0;
    transform: translate(-50%, -50%) rotate(var(--fx-rot, 0deg));
  }
  .shard {
    position: absolute;
    inset: 0;
    border-radius: 8px;
    overflow: hidden;
    background: #0d1220;
  }
  .shard img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .crk {
    position: absolute;
    inset: 0;
    overflow: visible;
  }
  .crk path {
    fill: none;
    stroke-linejoin: round;
    stroke-linecap: round;
    stroke-dasharray: 1;
    stroke-dashoffset: 1;
  }
  /* A dark rift with a white-hot core, glowing ember at its edge. */
  .crk path.under {
    stroke: rgba(12, 4, 10, 0.88);
    stroke-width: 2.6;
  }
  .crk path.over {
    stroke: #fff1dc;
    stroke-width: 1.1;
    filter: drop-shadow(0 0 2.5px rgba(255, 96, 40, 0.95));
  }
  .crk path.rim.over {
    stroke-width: 0.9;
  }

  /* #4427, a permanent sacrificed for its own ability: two halves of its
     art either side of a crack, which fall away. A dry, pale split with a
     dark edge, not the shatter's ember: nothing died. */
  .crack-half .face {
    box-shadow: 0 6px 14px rgba(0, 0, 0, 0.45);
  }
  .crack-line {
    position: absolute;
    left: 0;
    top: 0;
    overflow: visible;
    transform: translate(-50%, -50%) rotate(var(--fx-rot, 0deg));
  }
  .crack-line path {
    fill: none;
    stroke-linejoin: round;
    stroke-linecap: round;
    stroke-dasharray: 1;
    stroke-dashoffset: 1;
  }
  .crack-line path.under {
    stroke: rgba(14, 10, 6, 0.85);
    stroke-width: 3;
  }
  .crack-line path.over {
    stroke: #f4ead2;
    stroke-width: 1.1;
    filter: drop-shadow(0 0 2px rgba(255, 236, 200, 0.8));
  }
  .crack-puff {
    position: absolute;
    left: 0;
    top: 0;
    aspect-ratio: 1;
    border-radius: 50%;
    transform: translate(-50%, -50%);
    background: radial-gradient(
      circle,
      rgba(206, 186, 150, 0.7) 0%,
      rgba(150, 128, 96, 0.35) 40%,
      rgba(120, 100, 70, 0) 70%
    );
    opacity: 0;
    pointer-events: none;
  }
  /* The dust off the crack is what was spent: a Treasure's gold, a
     Clue's paper, a Food's crumbs; a land's is earth. */
  .fx[data-crack="treasure"] .crack-puff {
    background: radial-gradient(
      circle,
      rgba(255, 226, 140, 0.8) 0%,
      rgba(214, 170, 60, 0.4) 40%,
      rgba(180, 130, 30, 0) 70%
    );
  }
  .fx[data-crack="clue"] .crack-puff {
    background: radial-gradient(
      circle,
      rgba(240, 240, 232, 0.75) 0%,
      rgba(190, 196, 206, 0.35) 40%,
      rgba(160, 168, 180, 0) 70%
    );
  }
  .fx[data-crack="food"] .crack-puff {
    background: radial-gradient(
      circle,
      rgba(232, 176, 110, 0.75) 0%,
      rgba(184, 120, 60, 0.35) 40%,
      rgba(150, 90, 40, 0) 70%
    );
  }
  .flare {
    position: absolute;
    width: 46px;
    height: 46px;
    margin: -23px 0 0 -23px;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      rgba(255, 255, 255, 1) 0%,
      rgba(255, 196, 140, 0.85) 22%,
      rgba(255, 80, 40, 0.35) 48%,
      rgba(255, 60, 30, 0) 70%
    );
    mix-blend-mode: screen;
    opacity: 0;
  }
  .fx[data-effect="sacrifice"] .flare {
    background: radial-gradient(
      circle,
      rgba(255, 240, 240, 1) 0%,
      rgba(255, 90, 110, 0.85) 24%,
      rgba(170, 10, 40, 0.4) 50%,
      rgba(120, 0, 30, 0) 72%
    );
  }
  .mote {
    position: absolute;
    margin: -1px 0 0 -1px;
    border-radius: 50%;
    background: #ffb27a;
    box-shadow: 0 0 4px 1px rgba(255, 110, 50, 0.8);
    opacity: 0;
  }
  .fx[data-effect="exile"] .mote {
    background: #2a2630;
    box-shadow: none;
  }
  /* Inside a board wipe the shatter is white and black, like the cuts
     (Ian, round 3): no warm flare, white sparks and black grit. */
  .fx[data-effect-light] .flare {
    background: radial-gradient(
      circle,
      rgba(255, 255, 255, 1) 0%,
      rgba(230, 234, 242, 0.75) 26%,
      rgba(20, 20, 24, 0.35) 52%,
      rgba(0, 0, 0, 0) 72%
    );
  }
  .fx[data-effect-light] .mote {
    background: #f4f5f8;
    box-shadow: 0 0 3px 1px rgba(0, 0, 0, 0.6);
  }
  .fx[data-effect-light] .mote:nth-of-type(odd) {
    background: #111114;
    box-shadow: 0 0 3px rgba(255, 255, 255, 0.5);
  }

  /* D3, exile: a white void the card spins into, specks of light drawn
     in on spirals, and a last pinpoint as it closes. */
  .void {
    position: absolute;
    left: 0;
    top: 0;
    aspect-ratio: 1;
    translate: -50% -50%;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      #ffffff 0%,
      rgba(255, 255, 255, 0.96) 18%,
      rgba(232, 244, 255, 0.7) 34%,
      rgba(190, 220, 255, 0.28) 52%,
      rgba(160, 200, 255, 0) 70%
    );
    opacity: 0;
  }
  .vring {
    position: absolute;
    left: 0;
    top: 0;
    aspect-ratio: 1;
    translate: -50% -50%;
    box-sizing: border-box;
    border-radius: 50%;
    border: 1.5px solid rgba(255, 255, 255, 0.95);
    box-shadow:
      0 0 14px 2px rgba(200, 230, 255, 0.8),
      inset 0 0 12px 1px rgba(200, 230, 255, 0.6);
    opacity: 0;
  }
  .orbit {
    position: absolute;
    left: 0;
    top: 0;
    width: 0;
    height: 0;
  }
  .speck {
    position: absolute;
    left: 0;
    top: 0;
    translate: -50% -50%;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 0 6px 2px rgba(210, 235, 255, 0.95);
    opacity: 0;
  }
  .pinch {
    position: absolute;
    left: -30px;
    top: -30px;
    width: 60px;
    height: 60px;
    border-radius: 50%;
    background:
      linear-gradient(
          90deg,
          rgba(255, 255, 255, 0) 0%,
          rgba(255, 255, 255, 0.95) 50%,
          rgba(255, 255, 255, 0) 100%
        )
        center / 100% 2px no-repeat,
      radial-gradient(circle, #fff 0%, rgba(220, 240, 255, 0.8) 18%, rgba(220, 240, 255, 0) 45%);
    opacity: 0;
  }

  /* D5, sacrifice: a stone altar, and the knife. */
  .altar {
    position: absolute;
    overflow: visible;
    opacity: 0;
  }
  .altar-glow {
    fill: rgba(150, 10, 40, 0.38);
    filter: blur(6px);
  }
  .altar-edge {
    stroke: rgba(210, 40, 60, 0.75);
    stroke-width: 1.2;
  }
  .flame {
    fill: #ffd27a;
    filter: drop-shadow(0 0 3px rgba(255, 170, 60, 0.95));
  }
  .knife {
    position: absolute;
    opacity: 0;
    filter: drop-shadow(0 6px 8px rgba(0, 0, 0, 0.6));
    z-index: 2;
  }
  .knife svg {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
  }
  .glint {
    position: absolute;
    left: 0;
    right: 0;
    top: 40%;
    bottom: 0;
    background: linear-gradient(
      180deg,
      rgba(255, 255, 255, 0) 30%,
      rgba(255, 255, 255, 0.95) 50%,
      rgba(255, 255, 255, 0) 70%
    );
    clip-path: polygon(33% 0, 67% 0, 63% 53%, 50% 100%, 37% 53%);
    mix-blend-mode: screen;
    opacity: 0;
  }

  /* D6, bounce: turned over in 3D (the back is the card back), a cool
     halo while it is held up to the table, and two after-images. */
  .flip {
    position: absolute;
    left: 0;
    top: 0;
    width: 0;
    height: 0;
    transform-style: preserve-3d;
  }
  .flip .face {
    backface-visibility: hidden;
  }
  .face.back {
    transform: translate(-50%, -50%) rotate(var(--fx-rot, 0deg)) rotateX(180deg);
    background:
      url("/card-back-small.jpg") center / cover no-repeat,
      radial-gradient(ellipse at 50% 50%, #5a3a22 0%, #2b1a10 60%, #160d08 100%);
  }
  .halo {
    position: absolute;
    left: 0;
    top: 0;
    border-radius: 12px;
    transform: translate(-50%, -50%) rotate(var(--fx-rot, 0deg));
    box-shadow:
      0 0 0 2px color-mix(in srgb, var(--block, #91c3ff) 85%, #fff),
      0 0 26px 6px color-mix(in srgb, var(--block, #91c3ff) 70%, transparent);
    opacity: 0;
  }
  .ghost .face {
    box-shadow: none;
    opacity: 0;
  }

  /* D9: the rift a planeswalker leaves the plane through. */
  .rift {
    position: absolute;
    left: 0;
    top: 0;
    translate: -50% -50%;
    border-radius: 50%;
    background: radial-gradient(
      ellipse,
      rgba(255, 255, 255, 0.95) 0%,
      color-mix(in srgb, var(--magenta, #b08aff) 80%, #fff) 26%,
      color-mix(in srgb, var(--magenta, #b08aff) 45%, transparent) 50%,
      transparent 72%
    );
    mix-blend-mode: screen;
    opacity: 0;
  }

  /* A planeswalker that dies: a crack runs down the card. */
  .crack {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    z-index: 2;
  }
  .crack path {
    fill: none;
    stroke: #fff;
    stroke-width: 3;
    stroke-linejoin: round;
    stroke-dasharray: 1;
    stroke-dashoffset: 1;
    filter: drop-shadow(0 0 4px var(--magenta, #b08aff));
  }

  /* A planeswalker that arrives: a ring and a soft glow in the accent. */
  .glow {
    position: absolute;
    left: 0;
    top: 0;
    translate: -50% -50%;
    border-radius: 16px;
    background: radial-gradient(
      ellipse,
      color-mix(in srgb, var(--accent) 65%, transparent) 0%,
      transparent 70%
    );
    opacity: 0;
  }
  .ring {
    position: absolute;
    left: 0;
    top: 0;
    translate: -50% -50%;
    box-sizing: border-box;
    border-radius: 12px;
    border: 2px solid color-mix(in srgb, var(--accent) 70%, #fff);
    box-shadow: 0 0 22px 4px color-mix(in srgb, var(--accent) 70%, transparent);
    opacity: 0;
  }

  /* A token that is created (#4441): a soft, warm shimmer over the card,
     screened so the art shows through it. */
  .tglow {
    position: absolute;
    left: 0;
    top: 0;
    translate: -50% -50%;
    rotate: var(--fx-rot, 0deg);
    border-radius: 18px;
    background: radial-gradient(
      ellipse,
      rgba(255, 250, 232, 0.75) 0%,
      color-mix(in srgb, var(--accent) 35%, rgba(255, 236, 190, 0.45)) 38%,
      transparent 70%
    );
    mix-blend-mode: screen;
    opacity: 0;
  }

  /* Counters put on a permanent (#4442): a ring over its power/toughness,
     green for +1/+1, purple for -1/-1, a soft gold for any other kind,
     and a small tick above it. */
  .fx[data-effect="counters"] {
    --fx-c: #8ef0a8;
    --fx-c-glow: rgba(74, 222, 128, 0.7);
  }
  .fx[data-effect="counters"][data-effect-tone="minus"] {
    --fx-c: #d2a4ff;
    --fx-c-glow: rgba(168, 85, 247, 0.75);
  }
  .fx[data-effect="counters"][data-effect-tone="other"] {
    --fx-c: #f3e2b0;
    --fx-c-glow: rgba(234, 200, 120, 0.6);
  }
  .cpulse {
    position: absolute;
    left: 0;
    top: 0;
    translate: -50% -50%;
    border-radius: 999px;
    border: 2px solid var(--fx-c);
    box-shadow:
      0 0 14px 3px var(--fx-c-glow),
      inset 0 0 8px var(--fx-c-glow);
    opacity: 0;
  }
  .ctick {
    position: absolute;
    left: 0;
    width: 12px;
    height: 8px;
    translate: -50% 0;
    background: var(--fx-c);
    clip-path: polygon(50% 0, 100% 100%, 0 100%);
    opacity: 0;
  }
  .fx[data-effect-tone="minus"] .ctick {
    clip-path: polygon(0 0, 100% 0, 50% 100%);
  }

  /* A commander (#4443). Cast: a golden bloom where it sat, a beam that
     rises to the stack with motes riding it, and a crown of light on the
     spell. Home: its copy drifts back with two after-images and a faint
     streak, and a soft halo meets it on its card. Warm gold over the
     table, so it reads on a dark board. */
  .cmd-bloom {
    position: absolute;
    left: 0;
    top: 0;
    translate: -50% -50%;
    border-radius: 50%;
    background: radial-gradient(
      ellipse,
      rgba(255, 246, 214, 0.9) 0%,
      rgba(255, 206, 92, 0.55) 36%,
      transparent 70%
    );
    mix-blend-mode: screen;
    opacity: 0;
  }
  .cmd-beam,
  .cmd-streak {
    position: absolute;
    left: 0;
    top: 0;
    /* Its foot on the start, pointing at the end (the inline rotate). */
    translate: -50% -100%;
    transform-origin: 50% 100%;
    border-radius: 999px;
    opacity: 0;
  }
  .cmd-beam {
    background: linear-gradient(
      to top,
      rgba(255, 184, 60, 0) 0%,
      rgba(255, 196, 80, 0.8) 30%,
      rgba(255, 232, 160, 0.95) 88%,
      rgba(255, 244, 210, 0) 100%
    );
    box-shadow: 0 0 20px 6px rgba(255, 180, 48, 0.5);
    filter: blur(1.5px);
  }
  .cmd-streak {
    background: linear-gradient(to top, rgba(255, 196, 90, 0) 0%, rgba(255, 210, 120, 0.7) 100%);
    box-shadow: 0 0 12px 3px rgba(255, 180, 48, 0.3);
    filter: blur(3px);
  }
  .cmd-mote {
    position: absolute;
    left: -3.5px;
    top: -3.5px;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #ffe9a8;
    box-shadow: 0 0 10px 3px rgba(255, 186, 60, 0.85);
    opacity: 0;
  }
  .cmd-crown,
  .cmd-home {
    position: absolute;
    translate: -50% -50%;
    border: 3px solid #ffd36a;
    box-shadow:
      0 0 24px 8px rgba(255, 184, 48, 0.7),
      inset 0 0 14px rgba(255, 184, 48, 0.6);
    opacity: 0;
  }
  .cmd-crown {
    border-radius: 12px;
  }
  .cmd-home {
    border-radius: 10px;
    border-width: 2px;
  }
  .cmd-copy .face {
    box-shadow:
      0 0 16px 4px rgba(255, 206, 92, 0.55),
      0 10px 24px rgba(0, 0, 0, 0.5);
  }
  .cmd-ghost {
    opacity: 0;
  }
  .cmd-ghost .face {
    box-shadow: 0 0 14px 3px rgba(255, 206, 92, 0.5);
    filter: sepia(0.4) brightness(1.15);
  }

  /* An attachment (#4444). Its own tile does the sliding; this is the
     light over its host. An Equipment clasps cold and bright: a white-
     silver flash where it meets the host, a fan of sparks and a glint
     along the host's edge. An Aura is soft: a pale violet ring of light
     around the host and a few glints twinkling at its edge. Both are
     screened, so the art shows through on a dark board. */
  .att-rim,
  .att-halo {
    position: absolute;
    left: 0;
    top: 0;
    translate: -50% -50%;
    opacity: 0;
  }
  .att-rim {
    border-radius: 8px;
    border: 1.5px solid rgba(236, 244, 255, 0.9);
    box-shadow:
      0 0 10px 2px rgba(200, 220, 255, 0.55),
      inset 0 0 8px rgba(200, 220, 255, 0.4);
  }
  .att-seam {
    position: absolute;
    width: 0;
    height: 0;
  }
  .att-flash {
    position: absolute;
    left: -18px;
    top: -18px;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      #ffffff 0%,
      rgba(226, 236, 255, 0.85) 30%,
      rgba(150, 180, 230, 0) 70%
    );
    mix-blend-mode: screen;
    opacity: 0;
  }
  .att-spark {
    position: absolute;
    left: -7px;
    top: -1.5px;
    width: 14px;
    height: 3px;
    border-radius: 999px;
    background: linear-gradient(to right, rgba(255, 255, 255, 0), #fffbe8);
    box-shadow: 0 0 6px 1px rgba(255, 236, 170, 0.8);
    opacity: 0;
  }
  .att-halo {
    border-radius: 12px;
    border: 2px solid rgba(232, 214, 255, 0.85);
    box-shadow:
      0 0 22px 7px rgba(186, 150, 255, 0.55),
      inset 0 0 16px rgba(206, 180, 255, 0.5);
    mix-blend-mode: screen;
  }
  .att-glint {
    position: absolute;
    width: 12px;
    height: 12px;
    /* Centred by margin: the motion animates translate. */
    margin: -6px 0 0 -6px;
    background: #fbf4ff;
    clip-path: polygon(50% 0, 62% 38%, 100% 50%, 62% 62%, 50% 100%, 38% 62%, 0 50%, 38% 38%);
    filter: drop-shadow(0 0 4px rgba(200, 160, 255, 0.95));
    opacity: 0;
  }

  /* Cards milled or discarded (#4445): plain card copies, a card back
     where the art is not known, a soft puff of dust where they land on
     the graveyard pile (parchment for a mill, ash for a discard), and a
     sweep's faint streak. */
  .face.back-only {
    background:
      url("/card-back-small.jpg") center / cover no-repeat,
      radial-gradient(ellipse at 50% 50%, #5a3a22 0%, #2b1a10 60%, #160d08 100%);
  }
  .pile-dust {
    position: absolute;
    translate: -50% -50%;
    border-radius: 50%;
    background: radial-gradient(
      ellipse,
      rgba(238, 226, 200, 0.55) 0%,
      rgba(214, 198, 166, 0.28) 40%,
      transparent 70%
    );
    filter: blur(2px);
    opacity: 0;
  }
  .pile-dust[data-kind="discard"] {
    background: radial-gradient(
      ellipse,
      rgba(214, 214, 222, 0.5) 0%,
      rgba(160, 162, 176, 0.24) 40%,
      transparent 70%
    );
  }
  .pile-streak {
    position: absolute;
    left: 0;
    top: 0;
    /* Its foot on the start, pointing at the graveyard (the inline rotate). */
    translate: -50% -100%;
    transform-origin: 50% 100%;
    border-radius: 999px;
    background: linear-gradient(to top, rgba(230, 216, 186, 0) 0%, rgba(236, 226, 204, 0.5) 100%);
    filter: blur(3px);
    opacity: 0;
  }
  .pile-streak[data-kind="discard"] {
    background: linear-gradient(to top, rgba(190, 192, 204, 0) 0%, rgba(206, 208, 220, 0.45) 100%);
  }

  /* A card returned from a graveyard (#4446): a dark-green glow, the
     Toxic Deluge stain's green (FxTint.svelte) lit from within, about the
     copy as it rises, and a bloom of it from the pile it leaves. */
  .rise-aura {
    position: absolute;
    left: 0;
    top: 0;
    translate: -50% -50%;
    border-radius: 9px;
    box-shadow:
      0 0 0 2px rgba(120, 196, 104, 0.85),
      0 0 18px 6px rgba(48, 132, 46, 0.8),
      0 0 40px 14px rgba(22, 64, 20, 0.7);
    opacity: 0;
  }
  .rise-glow {
    position: absolute;
    translate: -50% -50%;
    border-radius: 50%;
    background: radial-gradient(
      ellipse,
      rgba(132, 206, 112, 0.6) 0%,
      rgba(40, 112, 40, 0.42) 38%,
      rgba(20, 58, 22, 0.22) 58%,
      transparent 72%
    );
    filter: blur(3px);
    opacity: 0;
  }

  /* A permanent that transforms (#4448): two faces back to back, turned
     with a tapped card, so the flip is about the card's own vertical axis;
     a pale rim lights it as it turns edge-on. */
  .turn {
    position: absolute;
    left: 0;
    top: 0;
    width: 0;
    height: 0;
  }
  .turn-card {
    position: absolute;
    left: 0;
    top: 0;
    width: 0;
    height: 0;
    transform-style: preserve-3d;
  }
  .turn-card .face {
    backface-visibility: hidden;
    transform: translate(-50%, -50%);
  }
  .turn-card .face.turn-to {
    transform: translate(-50%, -50%) rotateY(180deg);
  }
  .turn-rim {
    position: absolute;
    left: 0;
    top: 0;
    translate: -50% -50%;
    border-radius: 12px;
    box-shadow:
      0 0 0 2px rgba(226, 232, 255, 0.85),
      0 0 22px 6px rgba(178, 192, 255, 0.55);
    opacity: 0;
  }

  /* The game becomes day or night (#4448): the sky's tint over the table,
     a band of its light crossing it, and the moon or the sun at the top. */
  .sky {
    position: absolute;
    inset: 0;
    overflow: hidden;
    --sky-tint: rgba(18, 28, 84, 0.34);
    --sky-light: rgba(150, 172, 255, 0.22);
  }
  .sky[data-sky="day"] {
    --sky-tint: rgba(255, 196, 96, 0.2);
    --sky-light: rgba(255, 236, 180, 0.3);
  }
  .sky-tint {
    position: absolute;
    inset: 0;
    background: linear-gradient(180deg, var(--sky-tint) 0%, transparent 85%);
    opacity: 0;
  }
  .sky-band {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    background: linear-gradient(90deg, transparent, var(--sky-light), transparent);
    filter: blur(8px);
    opacity: 0;
  }
  .sky-orb {
    position: absolute;
    top: 18px;
    width: 34px;
    height: 34px;
    margin-left: -17px;
    border-radius: 50%;
    background: #eef1ff;
    /* A crescent: the dark of the night sky over most of the disc. */
    box-shadow:
      inset -9px -3px 0 1px rgba(30, 38, 90, 0.95),
      0 0 18px 4px rgba(190, 204, 255, 0.55);
    opacity: 0;
  }
  .sky[data-sky="day"] .sky-orb {
    background: radial-gradient(circle, #fff6d6 0%, #ffd56a 60%, #ffb63a 100%);
    box-shadow: 0 0 26px 10px rgba(255, 206, 110, 0.6);
  }

  /* A table-wide drain (#4450): a thin dark line, source to player. */
  .drain-svg {
    position: absolute;
    left: 0;
    top: 0;
    overflow: visible;
  }
  .drain-line {
    stroke: rgba(92, 20, 48, 0.9);
    stroke-width: 2px;
    stroke-linecap: round;
    stroke-dasharray: 1;
    stroke-dashoffset: 1;
    filter: drop-shadow(0 0 3px rgba(150, 30, 70, 0.7));
    opacity: 0;
  }
  .drain-source {
    position: absolute;
    width: 18px;
    height: 18px;
    translate: -50% -50%;
    border-radius: 50%;
    background: radial-gradient(circle, rgba(150, 30, 70, 0.8), rgba(92, 20, 48, 0));
    opacity: 0;
  }

  /* The crown or the initiative changes hands (#4450). */
  .glint-ring {
    position: absolute;
    translate: -50% -50%;
    border-radius: 14px;
    border: 2px solid rgba(255, 224, 120, 0.9);
    box-shadow:
      0 0 14px 2px rgba(255, 206, 80, 0.65),
      inset 0 0 12px rgba(255, 224, 120, 0.4);
    opacity: 0;
  }
  .glint-spark {
    position: absolute;
    width: 6px;
    height: 6px;
    translate: -50% -50%;
    border-radius: 50%;
    background: rgba(255, 244, 196, 0.98);
    box-shadow: 0 0 6px 2px rgba(255, 214, 100, 0.8);
    opacity: 0;
  }
  .fx[data-effect="initiative"] .glint-ring {
    border-color: rgba(176, 140, 255, 0.9);
    box-shadow:
      0 0 14px 2px rgba(124, 84, 220, 0.65),
      inset 0 0 12px rgba(150, 110, 240, 0.4);
  }
  .fx[data-effect="initiative"] .glint-spark {
    background: rgba(226, 212, 255, 0.98);
    box-shadow: 0 0 6px 2px rgba(140, 100, 230, 0.8);
  }

  /* A library searched (#4447): card backs riffled off the pile, which
     start hidden (the motion shows each while it moves), and a cool,
     moonlit light under the pile while it is looked through. */
  .riffle-card {
    opacity: 0;
  }
  .riffle-card .face {
    box-shadow: 0 4px 10px rgba(0, 0, 0, 0.5);
  }
  .riffle-glow {
    position: absolute;
    translate: -50% -50%;
    border-radius: 50%;
    background: radial-gradient(
      ellipse,
      rgba(214, 228, 255, 0.75) 0%,
      rgba(136, 164, 228, 0.42) 40%,
      rgba(74, 96, 160, 0.16) 60%,
      transparent 72%
    );
    filter: blur(3px);
    opacity: 0;
  }

  /* A counterspell: the stack item cracks and its glass flies out. */
  .spot {
    position: absolute;
    left: 0;
    top: 0;
    translate: -50% -50%;
  }
  .flash {
    position: absolute;
    inset: -20%;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      rgba(255, 255, 255, 0.95) 0%,
      rgba(190, 220, 255, 0.5) 40%,
      rgba(190, 220, 255, 0) 70%
    );
    opacity: 0;
  }
  .cracks {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    overflow: visible;
  }
  .cracks path {
    fill: none;
    stroke: #f2f8ff;
    stroke-width: 1.6;
    stroke-linejoin: round;
    stroke-dasharray: 1;
    stroke-dashoffset: 1;
    vector-effect: non-scaling-stroke;
    filter: drop-shadow(0 0 3px rgba(150, 200, 255, 0.9));
  }
  .glass {
    position: absolute;
    translate: -50% -50%;
    clip-path: polygon(0 100%, 55% 0, 100% 80%);
    background: linear-gradient(135deg, rgba(235, 245, 255, 0.95), rgba(140, 190, 255, 0.6));
    opacity: 0;
  }

  /* A spell that fizzles (#4450): a cool grey haze, a faint edge and
     specks that drift off it. Quieter than the counterspell's light. */
  .fmist {
    position: absolute;
    inset: -8%;
    border-radius: 12px;
    background: radial-gradient(
      ellipse,
      rgba(214, 220, 236, 0.6) 0%,
      rgba(160, 168, 196, 0.35) 45%,
      rgba(160, 168, 196, 0) 75%
    );
    opacity: 0;
  }
  .fedge {
    position: absolute;
    inset: 0;
    box-sizing: border-box;
    border-radius: 8px;
    border: 1.5px solid rgba(226, 232, 246, 0.8);
    box-shadow: 0 0 10px 1px rgba(170, 180, 214, 0.55);
    opacity: 0;
  }
  .fmote {
    position: absolute;
    translate: -50% -50%;
    border-radius: 50%;
    background: rgba(236, 240, 250, 0.95);
    box-shadow: 0 0 4px rgba(190, 200, 230, 0.8);
    opacity: 0;
  }

  /* A board wipe: one sweep across the whole table, and a tint under it. */
  .wipe {
    position: absolute;
    inset: 0;
    --wipe-core: rgba(14, 8, 24, 0.85);
    --wipe-edge: rgba(255, 107, 107, 0.7);
    --wipe-tint: rgba(40, 8, 20, 0.22);
  }
  .wipe[data-tone="cold"] {
    --wipe-core: rgba(170, 225, 255, 0.55);
    --wipe-edge: rgba(235, 248, 255, 0.95);
    --wipe-tint: rgba(110, 190, 255, 0.16);
  }
  .wipe[data-tone="cool"] {
    --wipe-core: rgba(145, 195, 255, 0.5);
    --wipe-edge: rgba(215, 235, 255, 0.9);
    --wipe-tint: rgba(110, 160, 255, 0.14);
  }
  .sweep {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    opacity: 0;
    background: linear-gradient(
      90deg,
      transparent 0%,
      var(--wipe-core) 55%,
      var(--wipe-edge) 92%,
      transparent 100%
    );
  }
  .tint {
    position: absolute;
    inset: 0;
    background: var(--wipe-tint);
    opacity: 0;
  }
  /* The still form's tint is on, not animated: the badge is the message. */
  .tint.still-tint {
    opacity: 1;
  }

  /* D7: the palm that stops the spell, in the cracks' own cold light,
     and the ring it sends out. */
  .palm {
    position: absolute;
    left: 50%;
    top: 50%;
    width: 92%;
    aspect-ratio: 1;
    translate: -50% -50%;
    overflow: visible;
    filter: drop-shadow(0 0 6px rgba(150, 200, 255, 0.95))
      drop-shadow(0 0 16px rgba(120, 180, 255, 0.6));
    opacity: 0;
  }
  .palm-body {
    fill: #eef6ff;
  }
  .palm-digit {
    fill: none;
    stroke: #eef6ff;
    stroke-width: 10.5;
    stroke-linecap: round;
  }
  .palm-crease {
    fill: none;
    stroke: rgba(120, 170, 235, 0.75);
    stroke-width: 1.6;
    stroke-linecap: round;
  }
  .shock {
    position: absolute;
    left: 50%;
    top: 50%;
    width: 150%;
    aspect-ratio: 1;
    translate: -50% -50%;
    box-sizing: border-box;
    border-radius: 50%;
    border: 2px solid rgba(235, 245, 255, 0.95);
    box-shadow: 0 0 18px 3px rgba(150, 200, 255, 0.7);
    opacity: 0;
  }

  /* D10: the wind cuts. Each crosses the board on its own line: a
     bright blade with a dark edge, a faint gust behind it, and the slice
     it leaves on the battlefield for a moment. White and black only: the
     scars alternate, a white slash edged in black, then a black one
     edged in white. */
  .cut {
    position: absolute;
    left: 0;
    right: 0;
    height: 0;
  }
  .blade {
    position: absolute;
    left: 0;
    top: calc(var(--cut-thick, 2px) / -2);
    height: var(--cut-thick, 2px);
    border-radius: 50%;
    background: linear-gradient(
      90deg,
      rgba(255, 255, 255, 0) 0%,
      rgba(235, 238, 245, 0.5) 30%,
      #ffffff 72%,
      rgba(255, 255, 255, 0) 100%
    );
    box-shadow:
      0 0 6px 1px rgba(255, 255, 255, 0.7),
      0 0 18px 4px rgba(0, 0, 0, 0.7);
    opacity: 0;
  }
  .cut[data-dir="-1"] .blade {
    background: linear-gradient(
      270deg,
      rgba(255, 255, 255, 0) 0%,
      rgba(235, 238, 245, 0.5) 30%,
      #ffffff 72%,
      rgba(255, 255, 255, 0) 100%
    );
  }
  .gust {
    position: absolute;
    left: 0;
    top: -9px;
    height: 18px;
    border-radius: 50%;
    background: linear-gradient(
      90deg,
      rgba(220, 225, 240, 0) 0%,
      rgba(225, 230, 245, 0.32) 50%,
      rgba(220, 225, 240, 0) 100%
    );
    filter: blur(3px);
    opacity: 0;
  }
  .scar {
    position: absolute;
    left: 0;
    right: 0;
    top: -0.5px;
    height: 1px;
    background: linear-gradient(
      90deg,
      rgba(255, 255, 255, 0) 0%,
      rgba(255, 255, 255, 0.85) 20%,
      rgba(255, 255, 255, 0.85) 80%,
      rgba(255, 255, 255, 0) 100%
    );
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.75);
    opacity: 0;
  }
  .cut:nth-of-type(even) .scar {
    height: 2px;
    background: linear-gradient(
      90deg,
      rgba(0, 0, 0, 0) 0%,
      rgba(0, 0, 0, 0.9) 20%,
      rgba(0, 0, 0, 0.9) 80%,
      rgba(0, 0, 0, 0) 100%
    );
    box-shadow: 0 0 4px rgba(255, 255, 255, 0.55);
  }

  /* D11: the swirl the board turns into, the blur that bends it, and the
     white void at its centre with a small darkness at its heart. */
  .swirl {
    position: absolute;
    overflow: visible;
    filter: blur(5px);
    mix-blend-mode: screen;
    opacity: 0;
  }
  .haze {
    position: absolute;
    inset: 0;
    backdrop-filter: blur(2.5px) saturate(0.6) brightness(1.08);
    -webkit-backdrop-filter: blur(2.5px) saturate(0.6) brightness(1.08);
    mask-image: radial-gradient(ellipse at 50% 50%, #000 0%, #000 35%, transparent 75%);
    -webkit-mask-image: radial-gradient(ellipse at 50% 50%, #000 0%, #000 35%, transparent 75%);
    opacity: 0;
  }
  .whitevoid {
    position: absolute;
    translate: -50% -50%;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      #ffffff 0%,
      rgba(255, 255, 255, 0.95) 24%,
      rgba(225, 240, 255, 0.55) 46%,
      rgba(200, 225, 255, 0) 70%
    );
    z-index: 2;
    opacity: 0;
  }
  .core {
    position: absolute;
    translate: -50% -50%;
    border-radius: 50%;
    background: radial-gradient(circle, #000 0%, #04060c 55%, rgba(4, 6, 12, 0) 72%);
    box-shadow: 0 0 10px 2px rgba(255, 255, 255, 0.9);
    z-index: 3;
    opacity: 0;
  }

  /* A signature's scene: a board-sized layer its parts draw into
     (fx/scenes/kit.ts styles each part inline). */
  .sig {
    position: absolute;
    inset: 0;
    overflow: hidden;
  }

  /* The outcome badge, still: the information a cue's motion carries. */
  .badge {
    position: absolute;
    translate: -50% -50%;
    padding: 4px 10px;
    border-radius: 999px;
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    white-space: nowrap;
    color: #fff;
    background: color-mix(in srgb, var(--danger) 70%, #14060a);
    box-shadow:
      0 0 0 1px rgba(0, 0, 0, 0.45),
      0 6px 18px rgba(0, 0, 0, 0.55);
  }
  .badge[data-effect="exile"] {
    background: color-mix(in srgb, #7cc8ff 55%, #07121f);
  }
  .badge[data-effect="bounce"],
  .badge[data-effect="pw_enter"] {
    background: color-mix(in srgb, var(--block) 60%, #0a1220);
  }

  /* #4879: decked out. Every part is placed by the script and moved by
     deckedOutMotion.ts; at rest (before its beat) each is hidden or still. */
  .decked {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }
  .dk-mat {
    position: absolute;
    overflow: hidden;
    border-radius: 10px;
  }
  .dk-pulse {
    position: absolute;
    box-sizing: border-box;
    border: 2px solid rgba(255, 70, 70, 0.9);
    border-radius: 6px;
    background: rgba(255, 40, 40, 0.18);
    opacity: 0;
  }
  .dk-wisp {
    position: absolute;
    border-radius: 50%;
    background: radial-gradient(
      ellipse at 50% 60%,
      rgba(200, 232, 255, 0.9) 0%,
      rgba(96, 160, 255, 0.55) 45%,
      rgba(40, 90, 200, 0) 75%
    );
    filter: blur(3px);
    opacity: 0;
  }
  .dk-skull {
    position: absolute;
    opacity: 0;
    transform-origin: 50% 50%;
    filter: drop-shadow(0 0 18px rgba(80, 150, 255, 0.55));
    will-change: transform, opacity, filter;
  }
  .dk-skull svg {
    display: block;
    overflow: visible;
  }
  .dk-card {
    position: absolute;
    transform: translate(-50%, -50%);
    border-radius: 5px;
    box-shadow: 0 6px 14px rgba(0, 0, 0, 0.55);
    object-fit: cover;
  }

  /* The OS's reduced-motion signal (DESIGN.md §4, the third way): every
     moving piece goes and the badges stay. The script reads the same
     signal and draws only badges, so this is the belt to its braces. */
  @media (prefers-reduced-motion: reduce) {
    .m {
      display: none;
    }
  }
</style>
