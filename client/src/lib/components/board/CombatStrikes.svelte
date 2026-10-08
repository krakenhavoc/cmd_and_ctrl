<script lang="ts">
  // CombatStrikes — combat you can watch (ADR 0134).
  //
  // When combat damage lands, each attacker that dealt it lunges at
  // what it hit, touches, and snaps back, and what it hit shakes at the
  // moment of contact. The board already shows the frame's end state
  // (life, badges, dead creatures gone): this is presentation over it,
  // and nothing waits for it (ADR 0134 §4).
  //
  // A tile cannot cross the table itself (its row scrolls, its panel
  // clips, and a creature that died has no tile left), so what moves is
  // an ART-ONLY COPY drawn here, in a layer over the whole board, while
  // the live tile hides for the flight (Card.svelte reads the shared
  // clock's striking set). Copies start from the live tile, or from the
  // last box cached for a creature that died; a box cached on a board
  // of another size is never used, and that strike is skipped.
  //
  // PR 2's polish: a lethal hit (a creature that dies, a player the hit
  // eliminates) shakes harder and flashes red; a creature that died
  // crumbles after its hit (a dead attacker once it is home); and
  // trample's excess draws a streak from the blockers to what it hit.
  //
  // Every rule (who moves, at what, who braces, who died, which hit is
  // lethal, the lunge's length, the streak's segment, the shards, the
  // timings, the gates) is in lib/combatStrikes.ts, pure and
  // unit-tested. The beats come from the board's shared clock
  // (lib/combatCues.svelte.ts), the same one CombatArrows' text cue
  // reads. This file measures, draws and flies.
  //
  // z 37: above the combat arrows (35, 36), below the pile (38), the
  // attention strip (40), the linger and the dice (41) and the dock
  // (55). aria-hidden, and it takes no pointer events.

  import { onDestroy, untrack } from "svelte";
  import { get } from "svelte/store";
  import { gsap } from "gsap";
  import type { CardView, GameView } from "../../protocol";
  import { settings } from "../../settings";
  import { cardAnchors, findCardAnchor, findSeatAnchor, hasSize } from "../../boardAnchor";
  import { boardExpandLayout } from "../../boardExpand";
  import { cardImageURL } from "../../cardImage";
  import { showsCardBack } from "../../cardBack";
  import { crumble, impactShake, lunge, streak } from "../../animations";
  import { play } from "../../sounds";
  import { keepArrowCache, type ScheduledCue } from "../../combatBeats";
  import type { CombatCues } from "../../combatCues.svelte";
  import {
    aimBox,
    combatMotion,
    crumbleShards,
    crumbleTiming,
    lungeVector,
    streakSegment,
    strikeLate,
    strikeTimeline,
    strikesFor,
    usableTile,
    type CachedTile,
    type StreakSegment,
    type StrikeBox,
    type StrikeTarget,
  } from "../../combatStrikes";

  interface Props {
    view: GameView;
    boardEl: HTMLElement | null;
    cues: CombatCues;
  }

  const { view, boardEl, cues: givenCues }: Props = $props();
  // Fixed for the life of the layer: the clock is never swapped.
  const cues = untrack(() => givenCues);

  interface Copy {
    key: string;
    cardID: string;
    // "lunge": a mover's flight; "die": a dead target taking the lethal
    // shake, then crumbling.
    mode: "lunge" | "die";
    src: string | null;
    // Centre, in board pixels, and the tile's size at rest and rotation.
    x: number;
    y: number;
    width: number;
    height: number;
    rot: number;
    to: { x: number; y: number };
    // It died in this beat: it crumbles (a mover once it is home).
    dies: boolean;
    flash: boolean;
    // The flash at contact is the lethal one: red (ADR 0134 PR 2).
    lethal: boolean;
    // Whether this copy hid a live tile (and so must unhide it).
    hides: boolean;
  }

  // Trample's streak from the blockers to what took the excess.
  interface Streak extends StreakSegment {
    key: string;
  }

  let copies = $state<Copy[]>([]);
  let streaks = $state<Streak[]>([]);
  let serial = 0;
  let destroyed = false;

  // Each battlefield tile's last measured box, and the image it showed,
  // by instance ID. Not reactive: nothing renders from them directly.
  // Kept through combat and until every strike has played, so a
  // creature that died can still strike (ADR 0134 §2).
  const tiles = new Map<string, CachedTile>();
  const images = new Map<string, string>();
  const lastViews = new Map<string, CardView>();
  const timers = new Set<ReturnType<typeof setTimeout>>();

  function later(fn: () => void, ms: number): void {
    const t = setTimeout(() => {
      timers.delete(t);
      if (!destroyed) fn();
    }, ms);
    timers.add(t);
  }

  function onBattlefieldNow(): Set<string> {
    return new Set(view.battlefield.cards.map((c) => c.instance_id));
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

  // Measure every live battlefield tile, and drop what nothing needs.
  function measure(): void {
    const b = boardSize();
    if (!b || !boardEl) return;
    const live = onBattlefieldNow();
    for (const [id, el] of cardAnchors(boardEl)) {
      if (!live.has(id) || !hasSize(el)) continue;
      tiles.set(id, tileOf(el, b.rect, b.size));
      const img = el.querySelector<HTMLImageElement>("img");
      const src = img?.getAttribute("src");
      if (src) images.set(id, src);
    }
    prune();
  }

  function prune(): void {
    // Kept while combat lasts, a cue is due, a copy flies or a contact is
    // still coming.
    if (keepArrowCache(view.turn?.step, cues.pending) || copies.length > 0 || timers.size > 0) {
      return;
    }
    const live = onBattlefieldNow();
    for (const id of [...tiles.keys()]) if (!live.has(id)) tiles.delete(id);
    for (const id of [...images.keys()]) if (!live.has(id)) images.delete(id);
    for (const id of [...lastViews.keys()]) if (!live.has(id)) lastViews.delete(id);
  }

  // A card's box now: its live tile if it is on the battlefield, else
  // the cached one, if that was measured on a board of this size.
  function cardTile(
    id: string,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): { tile: CachedTile; live: boolean } | null {
    if (boardEl && view.battlefield.cards.some((c) => c.instance_id === id)) {
      const el = findCardAnchor(boardEl, id, { accept: hasSize });
      if (el) return { tile: tileOf(el, b.rect, b.size), live: true };
    }
    const cached = usableTile(tiles.get(id), b.size);
    return cached ? { tile: cached, live: false } : null;
  }

  function seatElement(seat: number): HTMLElement | null {
    const id = view.seats[seat]?.id;
    if (!boardEl || !id) return null;
    return findSeatAnchor(boardEl, id, { accept: hasSize });
  }

  function targetBox(
    t: StrikeTarget,
    b: { rect: DOMRect; size: { w: number; h: number } },
  ): StrikeBox | null {
    if (t.kind === "card") return cardTile(t.cardID, b)?.tile.box ?? null;
    const el = seatElement(t.seat);
    if (!el) return null;
    const r = el.getBoundingClientRect();
    return {
      x: r.left + r.width / 2 - b.rect.left,
      y: r.top + r.height / 2 - b.rect.top,
      w: r.width,
      h: r.height,
    };
  }

  function imageFor(id: string): string | null {
    const cached = images.get(id);
    if (cached) return cached;
    const card = lastViews.get(id) ?? view.battlefield.cards.find((c) => c.instance_id === id);
    if (!card) return null;
    return showsCardBack(card) ? "/card-back-small.jpg" : cardImageURL(card, "small");
  }

  function addCopy(c: Omit<Copy, "key">, lifeMs: number): void {
    const key = `strike-${++serial}`;
    copies = [...copies, { ...c, key }];
    if (c.hides) cues.strikeStart(c.cardID);
    later(() => removeCopy(key), lifeMs);
  }

  function removeCopy(key: string): void {
    const c = copies.find((x) => x.key === key);
    if (!c) return;
    copies = copies.filter((x) => x.key !== key);
    if (c.hides) cues.strikeEnd(c.cardID);
    if (copies.length === 0) prune();
  }

  function playCue(cue: ScheduledCue): void {
    if (!cue.strikes) return;
    // Decided again at cue time, so turning reduced motion on mid-combat
    // starts no further strike (ADR 0134 §5).
    const s = get(settings);
    if (
      !combatMotion({
        enabled: s.animations.enabled,
        combat: s.animations.combat,
        reduceMotion: s.accessibility.reduceMotion,
      })
    ) {
      return;
    }
    if (strikeLate(cue.frameAt, cue.atMs, Date.now())) return;
    const tl = strikeTimeline(s.animations.speed);

    // The hit is heard as it lands: once per beat, at first contact
    // (ADR 0134 §7). With strikes on, Game.svelte leaves combat_resolve
    // to this, so it plays even when the board cannot be measured and
    // no copy flies. A late beat plays neither: the board has moved on.
    later(() => play("combat_resolve"), tl.contactMs);

    const b = boardEl ? boardSize() : null;
    if (!b) return;
    measure();
    const plan = strikesFor(cue, cue.log);

    // Trample's streaks, drawn at contact: from the blockers' centroid
    // to each player, planeswalker or battle that took the excess.
    const trample: StreakSegment[] = [];

    for (const m of plan.movers) {
      const from = cardTile(m.cardID, b);
      if (!from) continue;
      const aims = m.aim.map((t) => targetBox(t, b)).filter((x): x is StrikeBox => x !== null);
      const aim = aimBox(aims);
      if (!aim) continue;
      for (const t of m.through) {
        const to = targetBox(t, b);
        const seg = to ? streakSegment(aim, to) : null;
        if (seg) trample.push(seg);
      }
      const v = lungeVector(from.tile.box, aim);
      addCopy(
        {
          cardID: m.cardID,
          mode: "lunge",
          src: imageFor(m.cardID),
          x: from.tile.box.x,
          y: from.tile.box.y,
          width: from.tile.width,
          height: from.tile.height,
          rot: from.tile.rot,
          to: { x: v.x, y: v.y },
          dies: m.dies,
          flash: m.flash,
          lethal: m.lethal,
          hides: from.live,
        },
        // A creature that died is home at totalMs and crumbles there.
        m.dies ? crumbleTiming(tl, "mover").endMs : tl.totalMs,
      );
    }

    // Every target shakes at the same contact (CR 510.2), and trample's
    // streaks draw then, during the contact hold.
    later(() => {
      const now = boardSize();
      if (!now) return;
      for (const seg of trample) addStreak(seg, tl.streakDrawMs + tl.streakFadeMs);
      for (const imp of plan.impacts) {
        if (imp.target.kind === "seat") {
          const el = seatElement(imp.target.seat);
          if (el) void impactShake(el, { lethal: imp.lethal });
          continue;
        }
        const id = imp.target.cardID;
        const where = cardTile(id, now);
        if (!where) continue;
        if (where.live) {
          const el = boardEl ? findCardAnchor(boardEl, id, { accept: hasSize }) : null;
          if (el) void impactShake(el, { lethal: imp.lethal });
          continue;
        }
        // A creature that died: the lethal shake, then it crumbles, in
        // its cached place.
        if (!imp.dies) continue;
        const death = crumbleTiming(tl, "target");
        addCopy(
          {
            cardID: id,
            mode: "die",
            src: imageFor(id),
            x: where.tile.box.x,
            y: where.tile.box.y,
            width: where.tile.width,
            height: where.tile.height,
            rot: where.tile.rot,
            to: { x: 0, y: 0 },
            dies: true,
            flash: false,
            lethal: true,
            hides: false,
          },
          // Mounted at contact, so its life runs from there.
          death.endMs - tl.contactMs,
        );
      }
    }, tl.contactMs);
  }

  function addStreak(seg: StreakSegment, lifeMs: number): void {
    const key = `streak-${++serial}`;
    streaks = [...streaks, { ...seg, key }];
    later(() => {
      streaks = streaks.filter((s) => s.key !== key);
    }, lifeMs);
  }

  function reset(): void {
    for (const t of timers) clearTimeout(t);
    timers.clear();
    copies = [];
    streaks = [];
    cues.clearStriking();
    tiles.clear();
    images.clear();
    lastViews.clear();
  }

  // Each frame the clock folds in is also a measurement, taken at once
  // during combat: Board's frame effect runs after the DOM has the
  // frame, so the tiles measured here are this frame's, and a creature
  // that dies in a later frame was measured in this one. Outside combat
  // the animation-frame measurement below is enough.
  const unsubscribe = cues.subscribe({
    onCue: playCue,
    onReset: reset,
    onFrame: () => {
      if (keepArrowCache(view.turn?.step, 0)) measure();
    },
  });

  onDestroy(() => {
    destroyed = true;
    unsubscribe();
    for (const t of timers) clearTimeout(t);
    timers.clear();
    for (const c of copies) if (c.hides) cues.strikeEnd(c.cardID);
  });

  // Remember every battlefield card as of the latest frame: a creature
  // that died is drawn from the frame before it left.
  $effect(() => {
    const cards = view.battlefield.cards;
    untrack(() => {
      for (const c of cards) lastViews.set(c.instance_id, c);
    });
  });

  // Measure on every frame (after layout) and whenever the board or the
  // expanded board changes size, as CombatArrows does.
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

  // One copy's motion, started when it mounts. Its removal is on a
  // timer (addCopy), so the tween is presentation only.
  function fly(node: HTMLElement, c: Copy) {
    const face = node.querySelector<HTMLElement>(".face");
    // A creature that died breaks up after its hit (ADR 0134 question 7,
    // PR 2): the shards are planned in combatStrikes.ts, seeded by the
    // card, and made inside the face, so they go when the copy does.
    let gone = false;
    const crumbleNow = () => {
      if (gone || destroyed || !face) return;
      void crumble(face, crumbleShards(c.cardID, c.width, c.height, c.rot));
    };
    if (c.mode === "lunge") {
      void lunge(node, c.to, { flash: c.flash ? face : null, lethal: c.lethal }).then(() => {
        if (c.dies) crumbleNow();
      });
    } else if (face) {
      void impactShake(face, { lethal: true }).then(crumbleNow);
    }
    return {
      destroy() {
        gone = true;
        gsap.killTweensOf(node);
        if (face) {
          gsap.killTweensOf(face);
          for (const s of face.querySelectorAll(".shard")) gsap.killTweensOf(s);
        }
      },
    };
  }

  // One streak's draw, started when it mounts. Its removal is on a
  // timer (addStreak).
  function draw(node: HTMLElement) {
    void streak(node);
    return {
      destroy() {
        gsap.killTweensOf(node);
      },
    };
  }
</script>

<div class="strike-layer" aria-hidden="true" data-combat-strikes>
  {#each streaks as s (s.key)}
    <div
      class="streak"
      data-strike-streak
      style:left={`${s.x}px`}
      style:top={`${s.y}px`}
      style:width={`${s.length}px`}
      style:--streak-angle={`${s.angle}deg`}
    >
      <div class="bar" use:draw></div>
    </div>
  {/each}
  {#each copies as c (c.key)}
    <div
      class="copy"
      data-strike-copy={c.cardID}
      data-strike-mode={c.mode}
      style:left={`${c.x}px`}
      style:top={`${c.y}px`}
      use:fly={c}
    >
      <div
        class="face"
        style:width={`${c.width}px`}
        style:height={`${c.height}px`}
        style:--strike-rot={`${c.rot}deg`}
      >
        {#if c.src}
          <img src={c.src} alt="" decoding="async" draggable="false" />
        {/if}
      </div>
    </div>
  {/each}
</div>

<style>
  .strike-layer {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 37;
    overflow: hidden;
  }
  /* A zero-size anchor at the tile's centre: GSAP moves it, so the
     face's own transform (the shake and the tap rotation) is never
     overwritten by the flight. */
  .copy {
    position: absolute;
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
    transform: translate(-50%, -50%) translateX(var(--impact-x, 0px))
      rotate(var(--strike-rot, 0deg));
    filter: brightness(calc(1 + var(--impact-glow, 0)));
  }
  .face img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  /* The lethal flash (ADR 0134 PR 2): a red wash over the copy, at
     --impact-glow. Transparent unless --impact-wash is set, so an
     ordinary flash is the brightness filter alone, as before. */
  .face::after {
    content: "";
    position: absolute;
    inset: 0;
    z-index: 2;
    pointer-events: none;
    background: var(--impact-wash, transparent);
    opacity: calc(var(--impact-glow, 0) * 0.6);
  }
  /* The crumble: the shards (made by animations.ts crumble, so they
     carry no scoped class) take over from the face's own art, frame and
     shadow, and are free to fall out of its box. */
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
  /* Trample's streak: a bar along the segment, turned from its start.
     The outer box carries the angle and the inner bar the draw, so the
     tween never overwrites the rotation. */
  .streak {
    position: absolute;
    height: 0;
    transform: rotate(var(--streak-angle, 0deg));
    transform-origin: 0 0;
  }
  .streak .bar {
    position: absolute;
    left: 0;
    top: -2px;
    width: 100%;
    height: 4px;
    border-radius: 2px;
    transform-origin: 0 50%;
    background: linear-gradient(
      90deg,
      rgba(255, 214, 140, 0) 0%,
      rgba(255, 226, 170, 0.9) 55%,
      rgba(255, 246, 220, 1) 100%
    );
    box-shadow: 0 0 8px rgba(255, 200, 120, 0.75);
  }
</style>
