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
  import { EASE, cancelAnimations, run, settled } from "../../motion";
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
    EFFECT_CUE_CAP,
    counterBurst,
    cueSound,
    cueTiming,
    effectLate,
    effectMotion,
    effectsScheduled,
    planEffects,
    slashGeometry,
    soundBatch,
    wipeTiming,
    type EffectCue,
    type EffectKind,
    type EffectMotion,
    type WipeTone,
  } from "../../tableEffects";
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
    place(cue, base, motion, b);
  }

  // place measures the card behind a cue and mounts it, when it can be found.
  function place(
    cue: EffectCue,
    base: Omit<Fx, "key">,
    motion: EffectMotion,
    b: { rect: DOMRect; size: { w: number; h: number } },
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
    const speed = s.animations.speed;
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

    for (const w of plan.wipes) {
      if (!sp.silencedWipes.has(w.batch)) sound("board_wipe", 0, 1, w.batch);
      if (!b) continue;
      const t = wipeTiming(motion, speed);
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
      if (snd && !sp.silencedCues.has(cue.id)) sound(snd.name, t.delayMs, snd.volume, cue.batch);
      if (!b) continue;
      // A wipe says it once (its badge); its cards add nothing still.
      if (motion === "still" && cue.weight === "light") continue;
      if (++drawn > EFFECT_CUE_CAP) continue;
      drawCue(cue, motion, t, b);
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

  function timing(f: Fx, easing: string = EASE.power2Out): Parameters<typeof run>[2] {
    // fill both: a wipe's per-card copy holds its first frame until its
    // turn, and a finished cue stays gone until its owner removes it.
    return { duration: f.ms, delay: f.delay, easing, fill: "both" };
  }

  // One cue's motion, started when it mounts. Its removal is on a timer
  // (add), so the tween is presentation only. A still cue starts none.
  function motionOf(node: HTMLElement, f: Fx) {
    if (f.motion === "still") return {};
    const copy = node.querySelector<HTMLElement>(".copy");
    const face = node.querySelector<HTMLElement>(".face");
    const q = (sel: string) => node.querySelector<HTMLElement>(sel);
    switch (f.kind) {
      case "destroy": {
        if (copy) {
          run(
            copy,
            [
              { offset: 0, translate: "0px 0px", scale: "1", opacity: 1 },
              { offset: 0.3, translate: "0px 0px", scale: "1", opacity: 1 },
              { offset: 0.45, translate: "0px 2px", scale: "1", opacity: 1 },
              { offset: 1, translate: "0px 12px", scale: "0.96", opacity: 0 },
            ],
            timing(f),
          );
        }
        if (face) {
          run(
            face,
            [
              { offset: 0, filter: "brightness(1)" },
              { offset: 0.28, filter: "brightness(1.7)" },
              { offset: 0.42, filter: "brightness(0.3) saturate(0.3)" },
              { offset: 1, filter: "brightness(0.2) saturate(0.2)" },
            ],
            timing(f, EASE.linear),
          );
        }
        const bar = q(".slash .bar");
        if (bar) {
          run(
            bar,
            [
              { offset: 0, scale: "0 1", opacity: 1 },
              { offset: 0.28, scale: "1 1", opacity: 1 },
              { offset: 0.6, scale: "1 1", opacity: 1 },
              { offset: 1, scale: "1 1", opacity: 0 },
            ],
            timing(f, EASE.power2In),
          );
        }
        break;
      }
      case "exile": {
        if (copy) {
          run(
            copy,
            [
              { offset: 0, scale: "1", rotate: "0deg", translate: "0px 0px", opacity: 1 },
              { offset: 0.25, scale: "1.04", rotate: "-2deg", translate: "0px 0px", opacity: 1 },
              { offset: 1, scale: "0.04", rotate: "-50deg", translate: "0px -6px", opacity: 0 },
            ],
            timing(f, EASE.power2In),
          );
        }
        if (face) {
          run(
            face,
            [
              { offset: 0, filter: "grayscale(0) brightness(1)" },
              { offset: 0.3, filter: "grayscale(0.8) brightness(1.5) hue-rotate(160deg)" },
              { offset: 1, filter: "grayscale(1) brightness(2.4) hue-rotate(180deg)" },
            ],
            timing(f, EASE.linear),
          );
        }
        const portal = q(".portal");
        if (portal) {
          run(
            portal,
            [
              { offset: 0, scale: "0.1", opacity: 0 },
              { offset: 0.3, scale: "0.9", opacity: 1 },
              { offset: 0.7, scale: "1.15", opacity: 0.7 },
              { offset: 1, scale: "0.3", opacity: 0 },
            ],
            timing(f, EASE.sineInOut),
          );
        }
        break;
      }
      case "sacrifice": {
        if (copy) {
          run(
            copy,
            [
              { offset: 0, translate: "0px 0px", scale: "1 1", opacity: 1 },
              { offset: 0.15, translate: "0px -4px", scale: "1.03 1.03", opacity: 1 },
              { offset: 1, translate: "0px 30px", scale: "0.9 0.5", opacity: 0 },
            ],
            timing(f, EASE.power2In),
          );
        }
        if (face) {
          run(
            face,
            [
              { offset: 0, filter: "brightness(1)" },
              { offset: 1, filter: "brightness(0.25) saturate(1.6) hue-rotate(-25deg)" },
            ],
            timing(f, EASE.linear),
          );
        }
        const well = q(".well");
        if (well) {
          run(
            well,
            [
              { offset: 0, scale: "0.2", opacity: 0 },
              { offset: 0.4, scale: "1", opacity: 0.75 },
              { offset: 1, scale: "1.1", opacity: 0 },
            ],
            timing(f, EASE.sineInOut),
          );
        }
        break;
      }
      case "bounce": {
        if (copy) {
          run(
            copy,
            [
              { offset: 0, translate: "0px 0px", scale: "1", opacity: 1, easing: EASE.power2Out },
              {
                offset: 0.25,
                translate: "0px -12px",
                scale: "1.07",
                opacity: 1,
                easing: EASE.power2InOut,
              },
              { offset: 1, translate: `${f.fly.x}px ${f.fly.y}px`, scale: "0.3", opacity: 0 },
            ],
            timing(f, EASE.linear),
          );
        }
        if (face) {
          run(
            face,
            [
              { offset: 0, filter: "brightness(1)" },
              { offset: 0.25, filter: "brightness(1.3)" },
              { offset: 1, filter: "brightness(1.1)" },
            ],
            timing(f, EASE.linear),
          );
        }
        break;
      }
      case "dies": {
        // The existing crumble (ADR 0134), gated by animations.effects.
        if (face) {
          void crumble(face, crumbleShards(f.cardID || f.seed, f.width, f.height, f.rot), {
            gate: "effects",
            ms: f.ms,
          });
        }
        break;
      }
      case "pw_leave": {
        if (copy) {
          run(
            copy,
            [
              { offset: 0, translate: "0px 0px", scale: "1", opacity: 1 },
              { offset: 0.2, translate: "-3px 0px", scale: "1", opacity: 1 },
              { offset: 0.3, translate: "3px 0px", scale: "1", opacity: 1 },
              { offset: 0.4, translate: "-2px 0px", scale: "1", opacity: 1 },
              { offset: 0.5, translate: "0px 0px", scale: "1", opacity: 1 },
              { offset: 1, translate: "0px 6px", scale: "1.06", opacity: 0 },
            ],
            timing(f, EASE.power2In),
          );
        }
        if (face) {
          run(
            face,
            [
              { offset: 0, filter: "brightness(1)" },
              { offset: 0.3, filter: "brightness(2)" },
              { offset: 0.5, filter: "brightness(1.2)" },
              { offset: 1, filter: "brightness(0.5)" },
            ],
            timing(f, EASE.linear),
          );
        }
        const crack = q(".crack");
        if (crack) {
          run(
            crack,
            [{ opacity: 1 }, { opacity: 1, offset: 0.6 }, { opacity: 0 }],
            timing(f, EASE.linear),
          );
          for (const p of crack.querySelectorAll("path")) {
            run(
              p,
              [
                { strokeDashoffset: 1 },
                { strokeDashoffset: 0, offset: 0.3 },
                { strokeDashoffset: 0 },
              ],
              timing(f, EASE.power2Out),
            );
          }
        }
        break;
      }
      case "pw_enter": {
        const ring = q(".ring");
        if (ring) {
          run(
            ring,
            [
              { offset: 0, scale: "0.85", opacity: 0 },
              { offset: 0.3, scale: "1", opacity: 1 },
              { offset: 1, scale: "1.18", opacity: 0 },
            ],
            timing(f, EASE.power2Out),
          );
        }
        const glow = q(".glow");
        if (glow) {
          run(
            glow,
            [
              { offset: 0, scale: "0.6", opacity: 0 },
              { offset: 0.35, scale: "1", opacity: 0.9 },
              { offset: 1, scale: "1.4", opacity: 0 },
            ],
            timing(f, EASE.power2Out),
          );
        }
        break;
      }
      case "counter": {
        const flash = q(".flash");
        if (flash) {
          run(
            flash,
            [
              { offset: 0, scale: "0.4", opacity: 0.9 },
              { offset: 0.25, scale: "1.1", opacity: 0.6 },
              { offset: 1, scale: "1.4", opacity: 0 },
            ],
            timing(f, EASE.power2Out),
          );
        }
        const cracks = q(".cracks");
        if (cracks) {
          run(
            cracks,
            [{ opacity: 1 }, { opacity: 1, offset: 0.55 }, { opacity: 0 }],
            timing(f, EASE.linear),
          );
          for (const p of cracks.querySelectorAll("path")) {
            run(
              p,
              [
                { strokeDashoffset: 1 },
                { strokeDashoffset: 0, offset: 0.35 },
                { strokeDashoffset: 0 },
              ],
              timing(f, EASE.power2Out),
            );
          }
        }
        for (const el of node.querySelectorAll<HTMLElement>(".glass")) {
          const dx = Number(el.dataset.dx);
          const dy = Number(el.dataset.dy);
          const rot = Number(el.dataset.rot);
          run(
            el,
            [
              { offset: 0, translate: "0px 0px", rotate: "0deg", opacity: 0 },
              { offset: 0.2, translate: "0px 0px", rotate: "0deg", opacity: 1 },
              { offset: 1, translate: `${dx}px ${dy}px`, rotate: `${rot}deg`, opacity: 0 },
            ],
            timing(f, EASE.power2Out),
          );
        }
        break;
      }
      case "wipe": {
        const band = q(".sweep");
        if (band) {
          const bandW = Math.max(120, f.board.w * 0.34);
          band.style.width = `${bandW}px`;
          run(
            band,
            [
              { offset: 0, translate: `${-bandW}px 0px`, opacity: 0 },
              { offset: 0.1, translate: `${-bandW * 0.7}px 0px`, opacity: 1 },
              { offset: 0.9, translate: `${f.board.w - bandW * 0.3}px 0px`, opacity: 1 },
              { offset: 1, translate: `${f.board.w}px 0px`, opacity: 0 },
            ],
            timing(f, EASE.power1Out),
          );
        }
        const tint = q(".tint");
        if (tint) {
          run(
            tint,
            [{ opacity: 0 }, { opacity: 1, offset: 0.3 }, { opacity: 0 }],
            timing(f, EASE.linear),
          );
        }
        break;
      }
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
      {@const slash = slashGeometry(f.width, f.height)}
      <div
        class="fx m"
        data-effect={f.kind}
        data-effect-motion={f.motion}
        data-effect-light={f.light ? "" : undefined}
        data-effect-card={f.cardID}
        style:left={`${f.x}px`}
        style:top={`${f.y}px`}
        use:motionOf={f}
      >
        {#if f.kind === "exile"}
          <div class="portal" style:width={`${Math.max(f.width, f.height) * 1.6}px`}></div>
        {:else if f.kind === "sacrifice"}
          <div
            class="well"
            style:width={`${f.width * 1.3}px`}
            style:height={`${f.height * 0.5}px`}
          ></div>
        {/if}
        <div class="copy">
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
        </div>
        {#if f.kind === "destroy" && !f.light}
          <div
            class="slash"
            style:--slash-angle={`${slash.angle}deg`}
            style:--slash-len={`${slash.length}px`}
          >
            <div class="bar"></div>
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

  /* Destroy: a dark slash along the card's diagonal. The outer box is
     turned about the card's centre; the bar scales out from its start. */
  .slash {
    position: absolute;
    left: 0;
    top: 0;
    width: 0;
    height: 0;
    rotate: var(--slash-angle, 0deg);
  }
  .slash .bar {
    position: absolute;
    left: calc(var(--slash-len, 0px) / -2);
    top: -2px;
    width: var(--slash-len, 0px);
    height: 4px;
    border-radius: 2px;
    transform-origin: 0 50%;
    background: linear-gradient(
      90deg,
      rgba(8, 6, 14, 0) 0%,
      rgba(8, 6, 14, 0.95) 12%,
      #fff 50%,
      rgba(8, 6, 14, 0.95) 88%,
      rgba(8, 6, 14, 0) 100%
    );
    box-shadow:
      0 0 10px 2px rgba(120, 10, 30, 0.8),
      0 0 2px 0 rgba(255, 255, 255, 0.9);
  }

  /* Exile: a cold portal the card is drawn into. */
  .portal {
    position: absolute;
    left: 0;
    top: 0;
    aspect-ratio: 1;
    translate: -50% -50%;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      rgba(230, 248, 255, 0.95) 0%,
      rgba(120, 195, 255, 0.55) 34%,
      rgba(60, 110, 255, 0) 70%
    );
    mix-blend-mode: screen;
  }

  /* Sacrifice: a dark well under the card as it is pulled down. */
  .well {
    position: absolute;
    left: 0;
    top: 0;
    translate: -50% 0%;
    margin-top: 10%;
    border-radius: 50%;
    background: radial-gradient(
      ellipse,
      rgba(0, 0, 0, 0.9) 0%,
      rgba(40, 6, 30, 0.6) 50%,
      rgba(0, 0, 0, 0) 75%
    );
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
