<script lang="ts">
  // TableEffects — the generic effect cues (ADR 0209 §1 and §3, #4396).
  //
  // When a permanent is destroyed, exiled, sacrificed or bounced, a spell
  // is countered, a planeswalker arrives or goes, or a board wipe clears
  // three or more permanents at once, this layer draws one cue over the
  // table and plays its sound, so a player sees what happened before they
  // read the log. The board already shows the frame's end state: this is
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
  // z 57 on DESIGN.md §4's ladder: above the action dock (55), below the
  // hints (58). aria-hidden: the log and the board carry the news.

  import { onDestroy, untrack } from "svelte";
  import { get } from "svelte/store";
  import type { CardView, GameView } from "../../protocol";
  import { settings } from "../../settings";
  import { cardAnchors, findCardAnchor, findSeatAnchor, hasSize } from "../../boardAnchor";
  import { boardExpandLayout } from "../../boardExpand";
  import { cardImageURL } from "../../cardImage";
  import { showsCardBack } from "../../cardBack";
  import { isCreature, isPlaneswalker } from "../../cardTypes";
  import { crumble } from "../../animations";
  import { cancelAnimations, settled } from "../../motion";
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
    cueSound,
    cueTiming,
    effectLate,
    effectMotion,
    effectsScheduled,
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
    type EffectCue,
    type EffectKind,
    type EffectMotion,
    type WipeTone,
  } from "../../tableEffects";
  import { cueMotion } from "../../effectCueMotion";
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
    // A signature: its scene, the cards its batch took, where it lands,
    // and the factor from the scene's ms to real ms.
    scene?: SceneID;
    sceneCards?: SceneCard[];
    landing?: SceneCard | null;
    seats?: SceneSeat[];
    k?: number;
  }

  let fx = $state<Fx[]>([]);
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
    const b = boardSize();
    if (!b || !boardEl) return;
    const live = onBattlefieldNow();
    for (const [id, el] of cardAnchors(boardEl)) {
      if (!live.has(id) || !hasSize(el)) continue;
      tiles.set(id, tileOf(el, b.rect, b.size));
      const src = el.querySelector<HTMLImageElement>("img")?.getAttribute("src");
      if (src) images.set(id, src);
    }
    for (const el of boardEl.querySelectorAll<HTMLElement>("[data-stack-item-id]")) {
      const id = el.dataset.stackItemId;
      if (!id || !hasSize(el)) continue;
      const r = el.getBoundingClientRect();
      stackBoxes.set(id, {
        x: r.left + r.width / 2 - b.rect.left,
        y: r.top + r.height / 2 - b.rect.top,
        w: r.width,
        h: r.height,
      });
    }
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
    pw_enter: L.effectBadgePwEnter,
    pw_leave: L.effectBadgePwLeave,
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
    if (cue.kind === "counter") {
      // The stack linger's badge says "Countered"; only the motion is ours.
      if (motion === "still") return;
      const box = cue.stackItemID ? stackBoxes.get(cue.stackItemID) : undefined;
      if (!box) return;
      add({ ...base, x: box.x, y: box.y, width: box.w, height: box.h });
      return;
    }
    const cardID = cue.cardID;
    if (!cardID) return;
    // A planeswalker that just entered is still in its entrance pop (the
    // tile is scaled down and fading in), so its box is read once that
    // has settled, not when the frame arrives. A leaver is drawn at once
    // from its cache, so the card never blinks out.
    if (cue.kind === "pw_enter") {
      const tile = boardEl ? findCardAnchor(boardEl, cardID, { accept: hasSize }) : null;
      if (!tile) return;
      void settled(tile).then(() => {
        if (destroyed) return;
        const now = boardSize();
        if (now) place(cue, base, motion, now);
      });
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
    const generic = planEffects(r.fresh, log, { isPlaneswalker: isPW, combatMotion: strikes });
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
    if (plan.cues.length === 0 && plan.wipes.length === 0 && sp.sigs.length === 0) return;

    const at = Date.now();
    // animations.speed, compressed on a busy table: many cues close
    // together each play shorter (tableEffects.ts busyScale).
    const speed = s.animations.speed * busy.frame(headlines(plan, sp.sigs.length), at);
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
    ) => {
      const fire = () => play(name, { volume, batch: soundBatch(gameID, batch) });
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
        sound(snd.name, t.delayMs, snd.volume, cue.batch);
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

  // Measure on every frame (after layout) and whenever the board or the
  // expanded board changes size, as the strike layer does.
  $effect(() => {
    void view;
    void $boardExpandLayout;
    if (!boardEl) return;
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
    if (fx.length === 0 || typeof window === "undefined") return;
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
        });
      }
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
</div>

<style>
  .fx-layer {
    position: absolute;
    inset: 0;
    pointer-events: none;
    /* The table's z-index ladder (DESIGN.md §4, ADR 0209 §4): above the
       action dock (55), below the hints (58). */
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
    transform: translate(-50%, -50%) rotate(var(--fx-rot, 0deg)) rotateY(180deg);
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
     it leaves on the battlefield for a moment. */
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
      rgba(255, 220, 220, 0.5) 30%,
      #ffffff 72%,
      rgba(255, 255, 255, 0) 100%
    );
    box-shadow:
      0 0 6px 1px rgba(255, 90, 90, 0.75),
      0 0 18px 4px rgba(40, 6, 16, 0.65);
    opacity: 0;
  }
  .cut[data-dir="-1"] .blade {
    background: linear-gradient(
      270deg,
      rgba(255, 255, 255, 0) 0%,
      rgba(255, 220, 220, 0.5) 30%,
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
      rgba(255, 120, 120, 0) 0%,
      rgba(255, 150, 150, 0.7) 20%,
      rgba(255, 150, 150, 0.7) 80%,
      rgba(255, 120, 120, 0) 100%
    );
    box-shadow: 0 0 4px rgba(255, 70, 70, 0.6);
    opacity: 0;
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

  /* The OS's reduced-motion signal (DESIGN.md §4, the third way): every
     moving piece goes and the badges stay. The script reads the same
     signal and draws only badges, so this is the belt to its braces. */
  @media (prefers-reduced-motion: reduce) {
    .m {
      display: none;
    }
  }
</style>
