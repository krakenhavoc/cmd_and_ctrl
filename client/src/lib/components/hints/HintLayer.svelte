<script lang="ts">
  // HintLayer — the first-use hints, everywhere (ADR 0125 §3.5, §3.6).
  //
  // Mounted once, in App.svelte. Every POLL_MS it asks the engine
  // (lib/hints/queue.ts, through lib/hints/runtime.ts) which hint, if
  // any, belongs on screen, measures that hint's anchor, places the card
  // (lib/hints/place.ts) and publishes it as `activeTip`. It draws the
  // card itself, except for a hint whose place is the Settings dialog,
  // which the dialog's HintSlot draws inside its focus trap.
  //
  // What it reads, and nothing else: the route and the session (the
  // place and the context, with whether the person has a finished game:
  // lib/hints/endedGame.ts), the settings (`help.seen`, `help.tipsOff`,
  // the key bindings for copy, reduced motion), whether Settings is
  // open, and at the table the moment Game.svelte publishes.
  //
  // It also owns two accessibility pieces: the polite live region that
  // reads a tip out once when it appears ("Tip: <title>. <body>"), and
  // the anchor's aria-describedby, pointed at the tip's body while it
  // shows, so someone who tabs to the anchor hears it.
  //
  // With no hints at all (ADR 0125 PR 4 ships none) it never polls.

  import { get } from "svelte/store";
  import { route } from "../../router";
  import { session } from "../../session";
  import { settings, settingsOpen } from "../../settings";
  import { adminModeOn } from "../../admin";
  import { signedInUserID } from "../../myGames";
  import { HINTS } from "../../hints";
  import { anchorOf, placeOfRoute, type Hint, type HintContext } from "../../hints/hint";
  import { tableState } from "../../hints/tableMoment";
  import { hasEndedGameFor } from "../../hints/endedGame";
  import { CARD_WIDTH, isPhone, placeCard, type Rect } from "../../hints/place";
  import {
    activeTip,
    hintEngine,
    setActiveTip,
    TIP_BODY_ID,
    type ActiveTip,
  } from "../../hints/runtime";
  import { anchorRect, labelSelector, resolveAnchor, sameRect } from "../../tutorialAnchor";
  import { POLL_MS, copyText, type CopyContext } from "../../tutorial";
  import { effectiveBindings, formatChord, isMacLike } from "../../shortcuts";
  import { L } from "../../labels";
  import TipCard from "./TipCard.svelte";

  interface Props {
    /** Every hint; the collected ones unless a test hands it others. */
    hints?: readonly Hint[];
    pollMs?: number;
  }

  const { hints = HINTS, pollMs = POLL_MS }: Props = $props();

  // A new visit is a new route, or the Settings dialog opening.
  let routeKey = "";
  let routeVisit = 0;
  let settingsWasOpen = false;
  let settingsVisit = 0;

  // The anchor carrying aria-describedby, and what it had before.
  let described: { el: Element; had: string | null } | null = null;
  function describe(el: Element | null): void {
    if (described?.el === el) return;
    if (described) {
      if (described.had === null) described.el.removeAttribute("aria-describedby");
      else described.el.setAttribute("aria-describedby", described.had);
      described = null;
    }
    if (!el) return;
    const had = el.getAttribute("aria-describedby");
    el.setAttribute("aria-describedby", had ? `${had} ${TIP_BODY_ID}` : TIP_BODY_ID);
    described = { el, had };
  }

  let announcement = $state("");
  let announced: string | null = null;

  /** What the card must not cover: the dock, the stack pile, a dialog, the coach. */
  function avoidRects(anchor: Element): Rect[] {
    const sel = [
      `[aria-label="${L.actions}"]`,
      labelSelector(L.stackPile.any),
      '[role="dialog"]',
      '[role="alertdialog"]',
      ".coach-slot",
    ].join(", ");
    const out: Rect[] = [];
    for (const el of document.querySelectorAll(sel)) {
      // The anchor's own container (a dialog it sits in, the dock it is)
      // is not something to keep clear of.
      if (el.contains(anchor) || anchor.contains(el)) continue;
      const r = el.getBoundingClientRect();
      if (r.width > 0 && r.height > 0) {
        out.push({ left: r.left, top: r.top, width: r.width, height: r.height });
      }
    }
    return out;
  }

  function copyContext(): CopyContext {
    const keys = effectiveBindings(get(settings).shortcuts.bindings);
    const mac = isMacLike();
    return {
      helpKey: keys.toggleHelp ? formatChord(keys.toggleHelp, mac) : "",
      settingsKey: keys.openSettings ? formatChord(keys.openSettings, mac) : "",
      nextKey: keys.passPriority ? formatChord(keys.passPriority, mac) : "",
    };
  }

  function prefersReducedMotion(): boolean {
    return (
      typeof matchMedia === "function" && matchMedia("(prefers-reduced-motion: reduce)").matches
    );
  }

  function sameTip(a: ActiveTip | null, b: ActiveTip | null): boolean {
    if (a === null || b === null) return a === b;
    const pa = a.placement;
    const pb = b.placement;
    return (
      a.hint === b.hint &&
      a.title === b.title &&
      a.body === b.body &&
      a.stripBottom === b.stripBottom &&
      a.reduceMotion === b.reduceMotion &&
      sameRect(a.ring, b.ring) &&
      pa.kind === pb.kind &&
      (pa.kind !== "card" ||
        (pb.kind === "card" &&
          pa.side === pb.side &&
          Math.abs(pa.left - pb.left) < 0.5 &&
          Math.abs(pa.top - pb.top) < 0.5))
    );
  }

  function show(tip: ActiveTip | null, anchor: Element | null): void {
    if (!sameTip(get(activeTip), tip)) setActiveTip(tip);
    describe(tip ? anchor : null);
    const id = tip?.hint.id ?? null;
    if (id !== announced) {
      announced = id;
      announcement = tip ? `Tip: ${tip.title}. ${tip.body}` : "";
    }
  }

  function tick(): void {
    const r = get(route);
    const key = JSON.stringify(r);
    if (key !== routeKey) {
      routeKey = key;
      routeVisit++;
    }
    const sOpen = get(settingsOpen);
    if (sOpen && !settingsWasOpen) settingsVisit++;
    settingsWasOpen = sOpen;

    const place = sOpen ? "settings" : placeOfRoute(r.name);
    if (place === null) {
      show(null, null);
      return;
    }
    const s = get(session);
    const table = place === "table" ? get(tableState) : null;
    const ctx: HintContext = {
      place,
      route: r.name,
      signedIn: signedInUserID(s) !== null,
      adminMode: adminModeOn(s),
      adminToken: s?.principal.role === "admin",
      hasEndedGame: hasEndedGameFor(s),
      moment: table?.moment ?? null,
      view: table?.view ?? null,
      viewerID: table?.viewerID ?? null,
    };
    const conf = get(settings);
    const resolves = (h: Hint): boolean => {
      const a = anchorOf(h, ctx);
      return !!a && anchorRect([a]) !== null;
    };
    const hint = hintEngine().tick({
      hints,
      ctx,
      visit: sOpen ? `settings#${settingsVisit}` : `route#${routeVisit}`,
      seen: conf.help.seen,
      tipsOff: conf.help.tipsOff,
      now: Date.now(),
      resolves,
    });
    const a = hint ? anchorOf(hint, ctx) : null;
    const anchor = a ? resolveAnchor(a) : null;
    const ring = a ? anchorRect([a]) : null;
    if (!hint || !anchor || !ring) {
      show(null, null);
      return;
    }
    const card = document.querySelector<HTMLElement>(`[aria-label="${L.tip}"]`);
    const size = {
      width: card?.offsetWidth || CARD_WIDTH,
      height: card?.offsetHeight || 132,
    };
    const viewport = { width: window.innerWidth, height: window.innerHeight };
    const placement = isPhone(viewport.width)
      ? ({ kind: "strip" } as const)
      : placeCard({ anchor: ring, card: size, viewport, avoid: avoidRects(anchor) });
    if (placement.kind === "wait") {
      show(null, null);
      return;
    }
    // At the table the phone strip sits on the dock bar.
    const dock = place === "table" ? document.querySelector(`[aria-label="${L.actions}"]`) : null;
    const dockTop = dock?.getBoundingClientRect().top ?? viewport.height;
    const copy = copyContext();
    show(
      {
        hint,
        title: copyText(hint.title, copy),
        body: copyText(hint.body, copy),
        ring,
        placement,
        stripBottom: Math.max(6, viewport.height - dockTop + 6),
        reduceMotion: conf.accessibility.reduceMotion || prefersReducedMotion(),
      },
      anchor,
    );
  }

  $effect(() => {
    if (hints.length === 0) return;
    tick();
    const t = setInterval(tick, pollMs);
    return () => {
      clearInterval(t);
      show(null, null);
    };
  });
</script>

<!-- The polite live region: a tip is read out once when it appears. -->
<div class="hint-live" aria-live="polite">{announcement}</div>

{#if $activeTip && $activeTip.hint.place !== "settings"}
  <TipCard tip={$activeTip} />
{/if}

<style>
  .hint-live {
    position: absolute;
    width: 1px;
    height: 1px;
    margin: -1px;
    padding: 0;
    overflow: hidden;
    clip: rect(0 0 0 0);
    white-space: nowrap;
    border: 0;
  }
</style>
