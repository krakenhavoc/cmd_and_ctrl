// legalActions.ts — ADR 0105 (#1789): what the viewer may do right now,
// per card, as O(1) lookups for the board's "ready" highlights.
//
// NOTHING HERE IS A RULE. The input is the server's own answer:
//
//   - `GameView.legal_actions`, the per-card digest of the viewer's
//     UNCAPPED legal-move list (ADR 0105 §1). It survives the 48-move
//     wire cap down to the ability row.
//   - `GameView.legal_moves`, the capped list itself, as a fallback
//     for a server older than the digest. The cap keeps at least one
//     move per (source, kind), so the card-level answers below are
//     still exact; only the alternatives within a card (a second
//     ability row, a second zone) can be missing.
//
// Absent digest AND absent list means "no information": the seat owes
// no decision, or the server predates S31. Every lookup then answers
// "nothing", so the board highlights nothing. It never DIMS anything
// on that basis either — the negative gates (timing.ts) keep their own
// permissive reading of an absent list (ADR 0105 §3).
//
// One lookup is built per snapshot, in a `$derived` in Game.svelte,
// and passed down; each card reads it in O(1). `highlightsLive` decides
// whether the board shows it at all (the setting, and smart autopass's
// verdict for the frame).

import type {
  CardView,
  GameView,
  LegalActionsView,
  LegalMoveView,
  LegalSourceView,
  ManaAbilityView,
} from "./protocol";
import type { AutopassVerdict } from "./autopassDecision";

/** A zone a card can be "ready" in — the piles, the hand, the table. */
export type ReadyZone = "hand" | "graveyard" | "exile" | "library" | "command" | "battlefield";

/** A digested move kind: never pass, choice or mulligan. */
export type ReadyKind = LegalSourceView["kinds"][number];

const NONE: readonly string[] = Object.freeze([]);
const NO_KINDS: readonly ReadyKind[] = Object.freeze([]);

// Moves with no card carry the nil UUID rather than omitting `source`
// (protocol.ts LegalMoveView.source), so join on a real ID only.
const NIL_UUID = "00000000-0000-0000-0000-000000000000";

export interface LegalActions {
  /** A digest or a move list arrived on this frame. False: no information. */
  readonly known: boolean;
  /**
   * The per-ROW answers are complete: the lookup was built from the
   * server's uncapped `legal_actions` digest. False for the capped
   * `legal_moves` fallback (the wire cap can cut a card's second
   * ability row, ADR 0105 §1) and for no information. A gate that
   * withholds a row because its ref is missing reads only an exact
   * lookup; a highlight may read either.
   */
  readonly exact: boolean;
  /**
   * A pass move exists. Undefined when the frame carried no list at
   * all, which callers must read as "don't know", never as "no".
   */
  readonly pass: boolean | undefined;
  /** The card has at least one legal move of any kind. */
  isReady(cardID: string): boolean;
  /** The distinct kinds of the card's moves, in enumeration order. */
  kinds(cardID: string): readonly ReadyKind[];
  /**
   * The card may be cast (or, for a land, played) from `zone`. A move
   * whose zone the wire did not name counts for any zone.
   */
  castableFrom(cardID: string, zone: ReadyZone): boolean;
  /** ADR 0093 refs of the card's live activated-ability rows. */
  readyAbilityRefs(cardID: string): readonly string[];
  /** Refs of the card's live mana-ability rows. */
  readyManaRefs(cardID: string): readonly string[];
  /** The card's live special-action kinds ("foretell", "plot", …). */
  readySpecialActions(cardID: string): readonly string[];
  /** The creature may be declared as an attacker. */
  canAttack(cardID: string): boolean;
  /** Whom the creature may attack: players, planeswalkers, battles. */
  attackTargets(cardID: string): readonly string[];
  /** The attackers this creature may block, alone or in a group. */
  blockableAttackers(cardID: string): readonly string[];
  /**
   * How many cards in `zone` have a legal move. Per-seat zones (hand,
   * graveyard, library, command) need `seatID`; the shared zones
   * (exile, battlefield) count every card without it and only the
   * seat's own (exile: owner, battlefield: controller) with it.
   */
  readyCount(zone: ReadyZone, seatID?: string): number;
}

// ---- building ---------------------------------------------------------

type Sources = ReadonlyMap<string, LegalSourceView>;

function fromDigest(d: LegalActionsView): Map<string, LegalSourceView> {
  return new Map(Object.entries(d.sources ?? {}));
}

// Mutable twin of LegalSourceView for the fallback fold.
interface Acc {
  kinds: ReadyKind[];
  moves: number;
  abilities: string[];
  mana_abilities: string[];
  special_actions: string[];
  zones: string[];
  faces: number[];
  attack_targets: string[];
  blocks: string[];
}

function push<T>(list: T[], v: T | undefined | null): void {
  if (v === undefined || v === null || (v as unknown) === "") return;
  if (!list.includes(v)) list.push(v);
}

function str(v: unknown): string | undefined {
  return typeof v === "string" && v !== "" ? v : undefined;
}

// fromMoves is the client twin of the server's digestLegalMoves
// (server/internal/protocol/legal_actions.go), over the CAPPED list:
// it reads the same params, so on any frame under the cap it builds
// the same entries the server would have sent.
function fromMoves(moves: readonly LegalMoveView[]): Map<string, LegalSourceView> {
  const out = new Map<string, Acc>();
  const entry = (id: string): Acc => {
    let e = out.get(id);
    if (!e) {
      e = {
        kinds: [],
        moves: 0,
        abilities: [],
        mana_abilities: [],
        special_actions: [],
        zones: [],
        faces: [],
        attack_targets: [],
        blocks: [],
      };
      out.set(id, e);
    }
    return e;
  };
  for (const m of moves) {
    if (m.kind === "pass" || m.kind === "choice" || m.kind === "mulligan") continue;
    const source = m.source;
    if (!source || source === NIL_UUID) continue;
    const e = entry(source);
    e.moves++;
    push(e.kinds, m.kind);
    const p = m.params ?? {};
    switch (m.kind) {
      case "cast":
      case "land":
        push(e.zones, str(p.from_zone));
        if (m.kind === "cast") push(e.faces, typeof p.face === "number" ? p.face : 0);
        break;
      case "activate":
        push(e.abilities, str(p.ref));
        break;
      case "mana":
        push(e.mana_abilities, str(p.ref));
        break;
      case "special_action":
        push(e.special_actions, str(p.kind));
        break;
      case "attack":
        push(e.attack_targets, str(p.target));
        break;
      case "block": {
        // declare_blockers names a group (the two creatures a menace
        // attacker takes); every creature in it is a candidate.
        const group = Array.isArray(p.blocks) ? (p.blocks as Record<string, unknown>[]) : null;
        if (m.type === "declare_blockers" && group) {
          const counted = new Set([source]);
          for (const b of group) {
            const blocker = str(b?.blocker);
            if (!blocker) continue;
            const be = blocker === source ? e : entry(blocker);
            if (!counted.has(blocker)) {
              counted.add(blocker);
              be.moves++;
              push(be.kinds, "block");
            }
            push(be.blocks, str(b?.attacker));
          }
        } else {
          push(e.blocks, str(p.attacker));
        }
        break;
      }
    }
  }
  return out as unknown as Map<string, LegalSourceView>;
}

function cardsIn(view: GameView, zone: ReadyZone, seatID: string | undefined): CardView[] {
  if (zone === "exile" || zone === "battlefield") {
    const cards = (zone === "exile" ? view.exile : view.battlefield)?.cards ?? [];
    if (seatID === undefined) return cards;
    return cards.filter((c) => (zone === "exile" ? c.owner : c.controller) === seatID);
  }
  if (seatID === undefined) return [];
  const seat = view.seats?.find((s) => s.id === seatID);
  return seat?.[zone]?.cards ?? [];
}

function lookup(
  view: GameView | null,
  sources: Sources,
  pass: boolean | undefined,
  exact = false,
): LegalActions {
  const known = pass !== undefined;
  const counts = new Map<string, number>();
  const get = (id: string): LegalSourceView | undefined => sources.get(id);
  return {
    known,
    exact: known && exact,
    pass,
    isReady: (id) => sources.has(id),
    kinds: (id) => get(id)?.kinds ?? NO_KINDS,
    castableFrom(id, zone) {
      const e = get(id);
      if (!e || !e.kinds.some((k) => k === "cast" || k === "land")) return false;
      const zones = e.zones ?? [];
      return zones.length === 0 || zones.includes(zone);
    },
    readyAbilityRefs: (id) => get(id)?.abilities ?? NONE,
    readyManaRefs: (id) => get(id)?.mana_abilities ?? NONE,
    readySpecialActions: (id) => get(id)?.special_actions ?? NONE,
    canAttack: (id) => get(id)?.kinds.includes("attack") ?? false,
    attackTargets: (id) => get(id)?.attack_targets ?? NONE,
    blockableAttackers: (id) => get(id)?.blocks ?? NONE,
    readyCount(zone, seatID) {
      if (!view || sources.size === 0) return 0;
      const key = `${zone}|${seatID ?? ""}`;
      let n = counts.get(key);
      if (n === undefined) {
        n = cardsIn(view, zone, seatID).filter((c) => sources.has(c.instance_id)).length;
        counts.set(key, n);
      }
      return n;
    },
  };
}

/** The lookup that knows nothing: every answer is "no". */
export const NO_LEGAL_ACTIONS: LegalActions = lookup(null, new Map(), undefined);

/**
 * legalActionsOf builds the frame's lookup: from `legal_actions` when
 * the server sent the digest, else from `legal_moves`, else
 * NO_LEGAL_ACTIONS.
 */
export function legalActionsOf(view: GameView | null | undefined): LegalActions {
  if (!view) return NO_LEGAL_ACTIONS;
  if (view.legal_actions) {
    return lookup(view, fromDigest(view.legal_actions), view.legal_actions.pass === true, true);
  }
  if (view.legal_moves) {
    return lookup(
      view,
      fromMoves(view.legal_moves),
      view.legal_moves.some((m) => m.kind === "pass"),
    );
  }
  return NO_LEGAL_ACTIONS;
}

// ---- when the board shows it -----------------------------------------

/**
 * highlightsLive is ADR 0105 §3's timing rule, minus the part the data
 * already carries (a frame with no digest highlights nothing anyway):
 * the player has the setting on, and smart autopass is not about to
 * pass this frame — a window the client is skipping must not flash.
 * A bluff or a hold leaves the window with the player, so it shows.
 */
export function highlightsLive(enabled: boolean, verdict: AutopassVerdict | null): boolean {
  return enabled && verdict !== "pass";
}

/**
 * visibleHighlights is what the board draws from: the frame's lookup
 * while highlights are live, otherwise the one that knows nothing.
 * Only the positive treatment reads it. The gates (dimming, disabled
 * buttons) read the frame through timing.ts and never through this,
 * so turning highlights off withholds no click.
 */
export function visibleHighlights(actions: LegalActions, live: boolean): LegalActions {
  return live ? actions : NO_LEGAL_ACTIONS;
}

// ---- presentation: the mana-noise rule (ADR 0105 §4) -----------------

function isLand(card: CardView): boolean {
  return (card.type_line ?? "").toLowerCase().includes("land");
}

/**
 * notableManaRefs narrows a card's ready mana refs to the ones worth a
 * pip. Every untapped land has a legal mana move, and a board where
 * they all glowed would drown the one signal #1621 needs (Vivi's free
 * ability), so a land's {T} mana ability is not marked. A mana ability
 * IS marked when it is not the obvious kind: a nonland source, a
 * source in hand (a Spirit Guide), or a row without a {T} cost (Vivi,
 * a Treasure's sacrifice, a Lotus Petal).
 *
 * Presentation only: legality is `readyManaRefs`, from the digest.
 */
export function notableManaRefs(
  card: CardView,
  readyRefs: readonly string[],
  zone: ReadyZone,
): string[] {
  if (readyRefs.length === 0) return [];
  if (zone === "hand" || !isLand(card)) return [...readyRefs];
  const rows: ManaAbilityView[] = [
    ...(card.mana_abilities ?? []),
    ...(card.zone_mana_abilities ?? []),
  ];
  return readyRefs.filter((ref) => {
    const row = rows.find((r) => r.ref === ref);
    // A land row the frame does not describe: assume the ordinary
    // {T}: Add one — the quiet answer.
    return row !== undefined && row.tap_cost !== true;
  });
}

// ---- presentation: pips and menu rows (ADR 0105 §2, sub-PR 3) --------

/**
 * What a card's pips say. `abilities` is how many of its
 * activated-ability rows are live: the bolt pip, with a count from two
 * up, and the ready ring. `mana` is whether a mana ability worth
 * marking is live: the drop pip, after §4's noise rule, with no ring
 * of its own on a permanent. `special` is how many CR 116.2 special
 * actions are live (foretell, suspend and plot from a hand; turn face
 * up on a face-down permanent): the star pip, with the ready ring
 * (sub-PR 4).
 */
export interface ReadyPips {
  readonly abilities: number;
  readonly mana: boolean;
  readonly special: number;
}

export const NO_PIPS: ReadyPips = Object.freeze({ abilities: 0, mana: false, special: 0 });

/**
 * readyPips reads one card's pips off the lookup. Legality comes from
 * the digest alone (`readyAbilityRefs`, `readyManaRefs`,
 * `readySpecialActions`). The only decision made here is §4's
 * presentation rule for the drop pip (`notableManaRefs`). The lookup
 * that knows nothing gives NO_PIPS: highlights off, autopass passing,
 * a spectator, an opponent's card.
 *
 * Every special-action kind counts, including one this client has no
 * name for (a Room's unlock, or whatever lands next). The star is the
 * generic glyph, so a new kind lights with no client change (ADR 0105
 * §2, "Room doors").
 */
export function readyPips(legal: LegalActions, card: CardView, zone: ReadyZone): ReadyPips {
  const abilities = legal.readyAbilityRefs(card.instance_id).length;
  const mana = notableManaRefs(card, legal.readyManaRefs(card.instance_id), zone).length > 0;
  const special = legal.readySpecialActions(card.instance_id).length;
  if (abilities === 0 && !mana && special === 0) return NO_PIPS;
  return { abilities, mana, special };
}

/** The card wears at least one pip. */
export function hasPips(p: ReadyPips): boolean {
  return p.abilities > 0 || p.mana || p.special > 0;
}

/**
 * handHasAction is ADR 0105 §2's "not castable, but not dead either":
 * a hand card the server would let the viewer DO something with other
 * than cast it. That is a special action (foretell, plot, suspend), an
 * ability that functions from the hand (cycling), or a mana ability
 * that does (a Spirit Guide). Such a card is not dimmed even when it
 * cannot be cast.
 *
 * The hand reads it off the frame's FULL lookup (`legalGate`), never
 * off `visibleHighlights`. Dimming is the negative half, and the
 * highlight setting and autopass only ever turn off the positive one
 * (§3, §6). The lookup that knows nothing answers false, so an absent
 * digest dims exactly what it dimmed before.
 */
export function handHasAction(gate: LegalActions, cardID: string): boolean {
  return (
    gate.readySpecialActions(cardID).length > 0 ||
    gate.readyAbilityRefs(cardID).length > 0 ||
    gate.readyManaRefs(cardID).length > 0
  );
}

// The special-action kinds this client has words for. A kind missing
// here still lights its pip and still gets its menu row (the row's
// label is the server's); it only borrows the generic name.
const SPECIAL_ACTION_NAMES: Readonly<Record<string, string>> = Object.freeze({
  foretell: "foretell",
  suspend: "suspend",
  plot: "plot",
  turn_face_up: "turn face up",
  unlock: "unlock a door",
});

/** The generic name of a special action, and of the star pip. */
export const SPECIAL_ACTION_GENERIC = "special action";

/**
 * specialActionName is the short name of a special-action kind, for
 * the star pip's tooltip: "foretell", "turn face up", and, for a kind
 * this client does not know, the generic "special action".
 */
export function specialActionName(kind: string): string {
  return Object.hasOwn(SPECIAL_ACTION_NAMES, kind)
    ? SPECIAL_ACTION_NAMES[kind]
    : SPECIAL_ACTION_GENERIC;
}

/**
 * specialPipTitle is the star pip's tooltip: the live kinds' names,
 * deduplicated, in the digest's order.
 */
export function specialPipTitle(kinds: readonly string[]): string {
  const names: string[] = [];
  for (const k of kinds) {
    const n = specialActionName(k);
    if (!names.includes(n)) names.push(n);
  }
  return names.length > 0 ? names.join(", ") : SPECIAL_ACTION_GENERIC;
}

/**
 * pipCount is the number printed on a pip: nothing for one, the count
 * from two up. At the `small` card size the stylesheet hides it and
 * the pip shows alone (ADR 0105 §7).
 */
export function pipCount(n: number): string {
  return n >= 2 ? String(n) : "";
}

/**
 * readyFirst is the menus' row order: the rows the server would accept
 * right now first, then the rest, each half in its original order. It
 * is stable, so a menu with no ready row is exactly the menu it was.
 */
export function readyFirst<T>(rows: readonly T[], isReady: (row: T) => boolean): T[] {
  const ready: T[] = [];
  const rest: T[] = [];
  for (const r of rows) (isReady(r) ? ready : rest).push(r);
  return [...ready, ...rest];
}

/**
 * digestRefusesRow is the server's verdict on one activated-ability
 * row: true when the frame carries the exact digest, the row names its
 * ref, and that ref is not among the card's live activate refs. It
 * replaces the ability popover's client-side sorcery-speed gate (ADR
 * 0105 §1, sub-PR 3).
 *
 * Anything less is "no information" and answers false, so nothing is
 * newly withheld: no digest (the seat owes no decision, or an older
 * server), the capped `legal_moves` fallback, or a row with no ref.
 * Pass the frame's FULL lookup here, never `visibleHighlights`: this is
 * a gate, and the highlight setting must not open it.
 */
export function digestRefusesRow(
  gate: LegalActions,
  cardID: string,
  ref: string | undefined,
): boolean {
  if (!gate.exact || !ref) return false;
  return !gate.readyAbilityRefs(cardID).includes(ref);
}

/**
 * digestRefusesSpecial is digestRefusesRow for a special-action row:
 * true when the frame carries the exact digest and the card's live
 * special-action kinds leave `kind` out. The row's own `available` is
 * the server's timing answer. This adds what `available` does not say:
 * the price, since the enumerator runs a real auto-tap solve (ADR 0105,
 * Context). The digest names kinds, not a Room's doors, so a Room's
 * two unlock rows share one answer.
 *
 * As with digestRefusesRow, anything less than the exact digest is no
 * information and refuses nothing, and the caller passes the frame's
 * FULL lookup, never `visibleHighlights`.
 */
export function digestRefusesSpecial(gate: LegalActions, cardID: string, kind: string): boolean {
  if (!gate.exact || !kind) return false;
  return !gate.readySpecialActions(cardID).includes(kind);
}
