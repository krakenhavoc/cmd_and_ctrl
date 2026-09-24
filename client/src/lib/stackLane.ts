// stackLane — the one model of "what is on the stack, and what does it
// mean", shared by every way the board draws the stack (#1467).
//
// #1467 asked for the stack to move from a small docked card to the
// centre of the table, in three competing designs (a fan of cards with
// target arrows, a spotlight on the next item, a compact numbered
// ribbon) that a player picks between in Settings. Three designs over
// one set of rules is exactly the case client/README.md's "pull the
// decision into lib and test it" is for: if each design derived its
// own names, targets and ordering from the wire, the three would drift
// apart the first time one was fixed. So every derivation lives here,
// the docked StackOverlay reads it too, and a design is presentation
// only.
//
// What the model answers, per item:
//
//   - the name to show, verbatim (e2e specs look for labels such as
//     "Mulldrifter — draw two cards", so an ability's label is never
//     reworded);
//   - which card to draw and which card the hover preview may zoom
//     (for an ability, its source permanent — `previewableCard` is the
//     rule for whether the viewer may see it at all, #697);
//   - who controls it, and their seat colour;
//   - a short effect summary, when one can be stated honestly;
//   - each target resolved to a name and to WHERE it lives — a
//     permanent, a player, another stack item, or a card in some other
//     zone — plus a viewer-relative phrase ("your Lightning Bolt");
//   - the chips the compact overlay has always shown.
//
// And for the whole stack: resolution order (top first), the pending
// triggers after it, the plain-words line for the top item, and who
// holds priority.
//
// HONESTY. The effect summary never invents rules text. It has three
// sources and nothing else:
//
//   1. an ability's own label, which the server writes as
//      "<card> — <what happens>" (AGENTS.md §7); the part after the
//      dash is the summary;
//   2. a modal spell's chosen mode labels (#764), which are oracle
//      bullets;
//   3. the card's own oracle text, when the host can supply it and it
//      is one of two shapes read strictly — a lone "Counter target
//      <spell>." sentence, or a lone "<Name> deals N damage to <…>."
//      sentence. "Counter target spell unless its controller pays
//      {3}." does NOT count, because "counters your Lightning Bolt"
//      would be a claim the card does not make.
//
// Anything else summarises to null, and the top-item line falls back
// to "<name> resolves next". A wrong summary on the stack is worse than
// none: the player is deciding whether to respond to it.

import type { CardView, PlayerView, StackItemView, ZoneView } from "./protocol";
import { seatColor } from "./colors";
import { doubledTriggerLabel } from "./triggerDoubling";
import { previewableCard } from "./stackPreview";

/** The four ways the board can draw the stack (settings.display.stackStyle). */
export type StackStyle = "compact" | "fan" | "spotlight" | "ribbon";

/** Every style, in the order the settings panel lists them. */
export const STACK_STYLES: readonly StackStyle[] = ["compact", "fan", "spotlight", "ribbon"];

/** The floating styles — every style but the docked "compact" panel. */
export type StackLaneStyle = Exclude<StackStyle, "compact">;

export function isStackStyle(v: unknown): v is StackStyle {
  return typeof v === "string" && (STACK_STYLES as readonly string[]).includes(v);
}

/** Where an item is: on the stack, or a trigger still waiting to go on it. */
export type StackLaneItemKind = "spell" | "triggered" | "activated" | "pending";

/**
 * Where a target lives. The issue names the first three; "card" is a
 * card in another zone — a graveyard card Eternal Witness returns, an
 * exiled card — which is none of them and must not be passed off as a
 * permanent.
 */
export type StackLaneTargetKind = "permanent" | "player" | "stack" | "card";

export interface StackLaneTarget {
  kind: StackLaneTargetKind;
  /** Card instance id, player id, or stack item id. */
  id: string;
  /** The bare name: "Birds of Paradise", "Bot 2", "Lightning Bolt". */
  name: string;
  /**
   * Whose it is: a permanent's or stack item's CONTROLLER, a card's
   * OWNER, or for a player target the player themself. Null when the
   * target could not be resolved against the snapshot.
   */
  ownerSeat: string | null;
  /** That seat's colour (seatColor), or the spectator grey. */
  ownerColor: string;
  /** True when the target is the viewer, or belongs to them. */
  isViewers: boolean;
  /** For a stack target: its 1-based position, top = 1. */
  stackPosition: number | null;
  /**
   * Viewer-relative words for the target: "your Lightning Bolt",
   * "Bot 2's Birds of Paradise", "you", "Bot 2".
   */
  phrase: string;
}

export type StackLaneChipTone = "plain" | "flag" | "manual";

export interface StackLaneChip {
  label: string;
  tone: StackLaneChipTone;
  title?: string;
}

/**
 * An effect summary, and which of the honest sources it came from.
 * `counter` and `damage` come from the oracle text; `text` from an
 * ability label or a mode label.
 */
export type StackLaneEffect =
  | { kind: "counter"; text: string }
  | { kind: "damage"; text: string; amount: string; recipients: string }
  | { kind: "text"; text: string };

export interface StackLaneItem {
  id: string;
  kind: StackLaneItemKind;
  /** The display name, verbatim — the spell's card name or the ability's label. */
  name: string;
  /**
   * The name of the card behind an ability (its source), or the
   * spell's own name. Null when the source cannot be named.
   */
  sourceName: string | null;
  /** The card whose art represents the item (the spell, or an ability's source). */
  artCard: CardView | null;
  /** The card the hover preview may zoom, or null when the viewer may not see it (#697). */
  previewCard: CardView | null;
  casterSeat: string;
  casterName: string;
  casterSeatNum: number;
  casterColor: string;
  casterIsViewer: boolean;
  /** A short effect summary, or null when none can be stated honestly. */
  effect: StackLaneEffect | null;
  targets: StackLaneTarget[];
  /** The compact overlay's target line, "→ a / b", exactly as it always read. "" with no targets. */
  targetText: string;
  /** True for the stack item that resolves next. Never true for a pending trigger. */
  isTop: boolean;
  /** 1-based position on the stack, top = 1. Null for a pending trigger. */
  position: number | null;
  /** The CR 603.2d doubled-trigger chip text, or null. */
  doubledLabel: string | null;
  /** Ids of the other stack items that target this one. */
  targetedBy: string[];
  /** The overlay's chips, in the order it has always drawn them. */
  chips: StackLaneChip[];
  /** The wire item, for handlers that send it back (counter, target). */
  raw: StackItemView;
}

export interface StackLanePriority {
  /** True when the viewer holds priority. */
  viewerHolds: boolean;
  /** Seat id of the holder, or null. */
  holderSeat: string | null;
  /** Name of the holder, or null. */
  holderName: string | null;
  /** #1307: Board's public-timing "considering a response" read, passed through. */
  considering: boolean;
}

export interface StackLaneModel {
  /** Every item: the stack top-first, then the pending triggers. */
  items: StackLaneItem[];
  /** Just the stack, top-first. */
  stackItems: StackLaneItem[];
  /** Just the pending triggers, in the order the server sent them. */
  pendingTriggers: StackLaneItem[];
  /** The item that resolves next, or null. */
  top: StackLaneItem | null;
  /** The plain-words line for the top item, from the viewer's side. "" when nothing is live. */
  summary: string;
  priority: StackLanePriority;
  splitSecond: boolean;
  /** True iff there are stack items or pending triggers. */
  live: boolean;
}

export interface StackLaneInput {
  stack: ZoneView;
  stackItems: StackItemView[] | undefined;
  pendingTriggers: StackItemView[] | undefined;
  seats: PlayerView[];
  battlefield?: ZoneView;
  exile?: ZoneView;
  /** The viewer's seat id; null for a spectator (no "your"). */
  viewerID: string | null;
  /** The priority holder's seat id, or null. */
  priorityHolder: string | null;
  splitSecondActive: boolean;
  considering?: boolean;
  /**
   * The card's oracle text, when the host has it (the /cards cache).
   * Optional: without it spells summarise only from their mode labels.
   */
  oracleTextFor?: (card: CardView) => string | null | undefined;
}

/** The actions a lane style wires to its buttons. Every style gets the same set. */
export interface StackLaneControls {
  /** Counter an item (counter_spell / counter_ability). */
  counter(item: StackLaneItem): void;
  /** Pass priority; undefined when the viewer may not pass right now. */
  pass: (() => void) | undefined;
  /** True when an in-progress targeting prompt may point at this item. */
  targetable(item: StackLaneItem): boolean;
  /** Complete that prompt on this item. */
  target(item: StackLaneItem): void;
  /** #322 hover preview: call on pointerenter / focusin and pointerleave / focusout. */
  hoverEnter(item: StackLaneItem): void;
  hoverLeave(item: StackLaneItem): void;
}

/** Props every floating lane style component takes. */
export interface StackLaneStyleProps {
  model: StackLaneModel;
  controls: StackLaneControls;
}

/** Whether the lane has anything to show — cheap enough for Board to ask every frame. */
export function stackLaneLive(
  stackItems: StackItemView[] | undefined,
  pendingTriggers: StackItemView[] | undefined,
): boolean {
  return (stackItems?.length ?? 0) > 0 || (pendingTriggers?.length ?? 0) > 0;
}

// ---------------------------------------------------------------- //

const EM_DASH = " — ";

/**
 * The effect half of a label written "<card> — <what happens>", or
 * null when the label has no dash (a cost-colon label, a bare name).
 */
export function labelEffect(label: string | undefined): string | null {
  if (!label) return null;
  const at = label.indexOf(EM_DASH);
  if (at < 0) return null;
  const tail = label.slice(at + EM_DASH.length).trim();
  return tail === "" ? null : tail;
}

/** The source half of such a label, or null. */
function labelSource(label: string | undefined): string | null {
  if (!label) return null;
  const at = label.indexOf(EM_DASH);
  if (at <= 0) return null;
  return label.slice(0, at).trim() || null;
}

function escapeRegExp(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

/**
 * oracleEffect reads the two shapes the summary may state from a
 * card's oracle text, and nothing else. Each must be a whole
 * paragraph of one sentence, so a rider in the same paragraph ("…
 * unless its controller pays {3}", "… and you gain 3 life") rejects
 * the read rather than being silently dropped.
 */
export function oracleEffect(
  text: string | null | undefined,
  cardName: string,
  xValue?: number,
): StackLaneEffect | null {
  if (!text) return null;
  const paragraphs = text
    .split("\n")
    .map((p) => p.trim())
    .filter((p) => p !== "");
  const name = escapeRegExp(cardName);
  const damage = new RegExp(`^(?:${name}|This spell) deals (\\d+|X) damage to ([^.]+)\\.$`);
  for (const p of paragraphs) {
    const counter = /^Counter target ([^.]+)\.$/.exec(p);
    if (counter) {
      if (/\bunless\b|\binstead\b/i.test(counter[1])) return null;
      return { kind: "counter", text: `counter target ${counter[1]}` };
    }
    const dmg = damage.exec(p);
    if (dmg) {
      const recipients = dmg[2].trim();
      // "any target and you gain 3 life" is two effects; stating one of
      // them would read as the whole spell. "each creature and each
      // player" is one recipient list and is fine.
      if (/ and /.test(recipients) && !/^each [a-z ]+ and each [a-z ]+$/.test(recipients)) {
        return null;
      }
      const amount = dmg[1] === "X" && typeof xValue === "number" ? String(xValue) : dmg[1];
      return {
        kind: "damage",
        text: `${amount} damage to ${recipients}`,
        amount,
        recipients,
      };
    }
  }
  return null;
}

/** "a", "a and b", "a, b and c". */
export function joinPhrases(parts: string[]): string {
  if (parts.length <= 1) return parts[0] ?? "";
  return `${parts.slice(0, -1).join(", ")} and ${parts[parts.length - 1]}`;
}

function possessive(name: string): string {
  return name.endsWith("s") ? `${name}'` : `${name}'s`;
}

// ---------------------------------------------------------------- //

export function buildStackLane(input: StackLaneInput): StackLaneModel {
  const {
    stack,
    seats,
    battlefield,
    exile,
    viewerID,
    priorityHolder,
    splitSecondActive,
    considering = false,
    oracleTextFor,
  } = input;
  const wireStack = input.stackItems ?? [];
  const wirePending = input.pendingTriggers ?? [];

  // The spell cards on the stack, by instance id — a spell item's id
  // IS its card's instance id.
  const stackCards = new Map<string, CardView>();
  for (const c of stack.cards) stackCards.set(c.instance_id, c);

  // Every card an item might need to name or draw: stack (spells),
  // battlefield (ability sources, most targets), exile and graveyards
  // (dies-trigger sources, exiled / graveyard targets). The zone is
  // kept so a target can say where it lives.
  const anyCard = new Map<string, { card: CardView; zone: "stack" | "battlefield" | "other" }>();
  for (const c of stack.cards) anyCard.set(c.instance_id, { card: c, zone: "stack" });
  for (const c of battlefield?.cards ?? []) {
    anyCard.set(c.instance_id, { card: c, zone: "battlefield" });
  }
  for (const c of exile?.cards ?? []) anyCard.set(c.instance_id, { card: c, zone: "other" });
  for (const s of seats) {
    for (const c of s.graveyard?.cards ?? []) {
      anyCard.set(c.instance_id, { card: c, zone: "other" });
    }
  }

  const seatByID = new Map<string, PlayerView>();
  for (const s of seats) seatByID.set(s.id, s);

  // The wire sends stack_items bottom..top; the user-facing order is
  // "top resolves first", so reverse.
  const topFirst = [...wireStack].reverse();
  const positionByID = new Map<string, number>();
  topFirst.forEach((it, i) => positionByID.set(it.id, i + 1));

  const seatNum = (id: string | null | undefined): number =>
    id ? (seatByID.get(id)?.seat ?? -1) : -1;

  function artCardFor(item: StackItemView): CardView | undefined {
    if (item.kind === "spell") return stackCards.get(item.id);
    return anyCard.get(item.source_card_id)?.card;
  }

  function titleFor(item: StackItemView): string {
    if (item.kind === "spell") {
      const c = stackCards.get(item.id);
      if (c?.name) return c.name;
    }
    if (item.label) return item.label;
    const src = anyCard.get(item.source_card_id)?.card;
    if (src?.name) return src.name;
    return item.kind === "triggered" ? "trigger" : "ability";
  }

  // The overlay's old target line, kept byte-for-byte.
  function targetText(item: StackItemView): string {
    if (!item.targets || item.targets.length === 0) return "";
    const parts = item.targets
      .map((t) => {
        if (t.kind === "self") return "self";
        if (t.kind === "none") return "—";
        if (t.kind === "player") return seatByID.get(t.id ?? "")?.name ?? "player";
        return anyCard.get(t.id ?? "")?.card.name ?? "card";
      })
      .join(" / ");
    return `→ ${parts}`;
  }

  function whose(seatID: string | null, thing: string): string {
    if (seatID !== null && viewerID !== null && seatID === viewerID) return `your ${thing}`;
    const owner = seatID ? seatByID.get(seatID)?.name : undefined;
    return owner ? `${possessive(owner)} ${thing}` : thing;
  }

  function resolveTargets(item: StackItemView): StackLaneTarget[] {
    const out: StackLaneTarget[] = [];
    for (const t of item.targets ?? []) {
      if (!t.id || t.kind === "self" || t.kind === "none") continue;
      if (t.kind === "player") {
        const p = seatByID.get(t.id);
        const name = p?.name ?? "player";
        const mine = viewerID !== null && t.id === viewerID;
        out.push({
          kind: "player",
          id: t.id,
          name,
          ownerSeat: p ? t.id : null,
          ownerColor: seatColor(p?.seat ?? -1),
          isViewers: mine,
          stackPosition: null,
          phrase: mine ? "you" : name,
        });
        continue;
      }
      // A card slot: another stack item, a permanent, or a card
      // somewhere else.
      const onStack = wireStack.find((s) => s.id === t.id);
      if (onStack) {
        const owner = onStack.controller;
        const name = titleFor(onStack);
        out.push({
          kind: "stack",
          id: t.id,
          name,
          ownerSeat: owner,
          ownerColor: seatColor(seatNum(owner)),
          isViewers: viewerID !== null && owner === viewerID,
          stackPosition: positionByID.get(t.id) ?? null,
          phrase: whose(owner, name),
        });
        continue;
      }
      const found = anyCard.get(t.id);
      const card = found?.card;
      const name = card?.name || "card";
      const owner = card
        ? found.zone === "battlefield"
          ? card.controller || card.owner
          : card.owner || card.controller
        : null;
      out.push({
        // Unresolvable counts as "card": it makes no claim that the
        // target is on the battlefield.
        kind: found?.zone === "battlefield" ? "permanent" : "card",
        id: t.id,
        name,
        ownerSeat: owner ?? null,
        ownerColor: seatColor(seatNum(owner)),
        isViewers: viewerID !== null && owner === viewerID,
        stackPosition: null,
        phrase: whose(owner ?? null, name),
      });
    }
    return out;
  }

  function effectFor(item: StackItemView, card: CardView | undefined): StackLaneEffect | null {
    if (item.kind === "spell") {
      if (item.mode_labels && item.mode_labels.length > 0) {
        return { kind: "text", text: item.mode_labels.join("; ") };
      }
      if (card && oracleTextFor) {
        return oracleEffect(oracleTextFor(card), card.name, item.x_value);
      }
      return null;
    }
    const tail = labelEffect(item.label);
    return tail ? { kind: "text", text: tail } : null;
  }

  function chipsFor(item: StackItemView, kind: StackLaneItemKind): StackLaneChip[] {
    const chips: StackLaneChip[] = [];
    if (kind === "pending") return chips;
    const card = item.kind === "spell" ? stackCards.get(item.id) : undefined;
    if (item.kind !== "spell") chips.push({ label: item.kind, tone: "plain" });
    const doubled = doubledTriggerLabel(item.doubled_by, item.doubled_by_name);
    if (doubled) chips.push({ label: doubled, tone: "flag" });
    if (card?.auto) {
      chips.push({
        label: "auto",
        tone: "flag",
        title: "this card auto-resolves — effect fires when priority passes to empty stack",
      });
    }
    // "manual": the moment the expectation forms. The spell is on the
    // stack, everyone is looking at it, and in a second it will
    // resolve and appear to do nothing. Saying so here is what reports
    // #321 / #324 / #325 / #332 / #333 each needed and none of them
    // got. Transient by construction — the chip leaves with the stack
    // item, so it never becomes board furniture.
    if (card?.unimplemented) {
      chips.push({
        label: "manual",
        tone: "manual",
        title:
          "this card's rules aren't implemented yet — it resolves with no effect, so resolve it by hand",
      });
    }
    if (item.x_value) chips.push({ label: `X = ${item.x_value}`, tone: "plain" });
    // S22: an overloaded Rift wipes and a hard-cast one bounces one thing.
    if (item.alt_cost) chips.push({ label: item.alt_cost, tone: "flag" });
    // #1267 (CR 702.174): a promised gift changes what the spell does,
    // and who gets it is public.
    if (item.gift_to) {
      const to = seatByID.get(item.gift_to)?.name;
      chips.push({
        label: `Gift → ${to ?? "opponent"}`,
        tone: "flag",
        title: `the gift was promised to ${to ?? "an opponent"} — they get it before the spell's other effects`,
      });
    }
    // #764: the chosen bullets, in announce order (CR 608.2c) and with
    // repeats (CR 700.2d). Raw indexes ("modes: 0, 2") are the fallback
    // only: nobody at the table can read them, because the caster's
    // hand card is gone once the spell is on the stack.
    if (item.mode_labels && item.mode_labels.length > 0) {
      for (const m of item.mode_labels) chips.push({ label: m, tone: "plain" });
    } else if (item.modes && item.modes.length > 0) {
      chips.push({ label: `modes: ${item.modes.join(", ")}`, tone: "plain" });
    }
    if (item.split_second) chips.push({ label: "split-second", tone: "flag" });
    if (item.hold_priority) chips.push({ label: "held priority", tone: "flag" });
    return chips;
  }

  function build(item: StackItemView, kind: StackLaneItemKind, index: number): StackLaneItem {
    const art = artCardFor(item);
    const caster = seatByID.get(item.controller);
    const sourceName =
      item.kind === "spell"
        ? (stackCards.get(item.id)?.name ?? null)
        : (anyCard.get(item.source_card_id)?.card.name ?? labelSource(item.label));
    const isStack = kind !== "pending";
    return {
      id: item.id,
      kind,
      name: kind === "pending" ? item.label || "trigger" : titleFor(item),
      sourceName: sourceName || null,
      artCard: art ?? null,
      previewCard: previewableCard(art),
      casterSeat: item.controller,
      casterName: caster?.name ?? "?",
      casterSeatNum: caster?.seat ?? 0,
      casterColor: seatColor(caster?.seat ?? 0),
      casterIsViewer: viewerID !== null && item.controller === viewerID,
      effect: effectFor(item, item.kind === "spell" ? stackCards.get(item.id) : undefined),
      targets: resolveTargets(item),
      targetText: targetText(item),
      isTop: isStack && index === 0,
      position: isStack ? index + 1 : null,
      doubledLabel: doubledTriggerLabel(item.doubled_by, item.doubled_by_name),
      targetedBy: [],
      chips: chipsFor(item, kind),
      raw: item,
    };
  }

  const stackOut = topFirst.map((it, i) => build(it, it.kind, i));
  const pendingOut = wirePending.map((it, i) => build(it, "pending", i));

  // targetedBy: the reverse edge, so a design can say "targeted by
  // Counterspell" on the item under threat.
  const byID = new Map(stackOut.map((it) => [it.id, it] as const));
  for (const it of [...stackOut, ...pendingOut]) {
    for (const t of it.targets) {
      if (t.kind !== "stack") continue;
      const hit = byID.get(t.id);
      if (hit && hit.id !== it.id) hit.targetedBy.push(it.id);
    }
  }

  const holder = priorityHolder ? (seatByID.get(priorityHolder) ?? null) : null;
  const priority: StackLanePriority = {
    viewerHolds: viewerID !== null && priorityHolder === viewerID,
    holderSeat: holder?.id ?? null,
    holderName: holder?.name ?? null,
    considering,
  };

  const top = stackOut[0] ?? null;
  return {
    items: [...stackOut, ...pendingOut],
    stackItems: stackOut,
    pendingTriggers: pendingOut,
    top,
    summary: summarize(top, pendingOut),
    priority,
    splitSecond: splitSecondActive,
    live: stackOut.length > 0 || pendingOut.length > 0,
  };
}

/**
 * summarize is the plain-words line for the top item. The subject is
 * the spell's name, or "<source>'s trigger / ability" for an ability
 * (its label is already "<source> — <effect>", so repeating it whole
 * would say the effect twice).
 */
export function summarize(top: StackLaneItem | null, pending: StackLaneItem[] = []): string {
  if (!top) {
    if (pending.length === 1) return `${pending[0].name} is waiting to go on the stack`;
    if (pending.length > 1) return `${pending.length} triggers are waiting to go on the stack`;
    return "";
  }
  const subject =
    top.kind === "spell" || !top.sourceName
      ? top.name
      : `${possessive(top.sourceName)} ${top.kind === "triggered" ? "trigger" : "ability"}`;
  const head = `${subject} resolves next`;
  const targets = joinPhrases(top.targets.map((t) => t.phrase));
  const eff = top.effect;

  if (eff?.kind === "counter") {
    const countered = top.targets.filter((t) => t.kind === "stack");
    if (countered.length > 0) {
      return `${head} and counters ${joinPhrases(countered.map((t) => t.phrase))}`;
    }
    return targets ? `${head}, targeting ${targets}` : head;
  }
  if (eff?.kind === "damage") {
    if (targets) return `${head} — ${eff.amount} damage to ${targets}`;
    // No announced target: the recipients are a quantifier the card
    // prints ("each creature"), quoted as printed. A "target" clause
    // with no target resolved says only the amount.
    if (!/\btarget\b/.test(eff.recipients)) return `${head} — ${eff.text}`;
    return `${head} — ${eff.amount} damage`;
  }
  if (eff?.kind === "text") {
    return targets ? `${head} — ${eff.text}, targeting ${targets}` : `${head} — ${eff.text}`;
  }
  return targets ? `${head}, targeting ${targets}` : head;
}
