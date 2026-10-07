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
//     ability row, a second zone) can be missing. Since ADR 0122 §6.1
//     the frame says when they are: `legal_moves_truncated` is set
//     exactly when the cap dropped one. This lookup does not read it
//     and `exact` stays false for the fallback; the flag is there for
//     the reader that one day needs the alternatives.
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
  /**
   * #1918: the server's hint when EVERY cast the card has would do
   * nothing right now (an overloaded Counterflux with no spell to
   * counter). Undefined when any cast would do something, or none is
   * idle. The casts are still legal; this only explains them.
   */
  castIdleHint(cardID: string): string | undefined;
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
  /**
   * ADR 0130 §5: the creature may be exerted as it attacks right now
   * (CR 701.43d) — the server offers its attacks with `exert: true`
   * beside the plain ones. Read off the digest, never oracle text.
   */
  canExertOnAttack(cardID: string): boolean;
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
  cast_idle_hint?: string;
  exert_on_attack?: boolean;
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
  // #1918: per card, its cast moves, how many are idle, the first hint.
  const idle = new Map<Acc, { casts: number; idle: number; hint?: string }>();
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
        if (m.kind === "cast") {
          push(e.faces, typeof p.face === "number" ? p.face : 0);
          let tl = idle.get(e);
          if (!tl) idle.set(e, (tl = { casts: 0, idle: 0 }));
          tl.casts++;
          if (m.idle_hint) {
            tl.idle++;
            tl.hint ??= m.idle_hint;
          }
        }
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
        // ADR 0130 §5: the twin move that exerts it as it attacks.
        if (p.exert === true) e.exert_on_attack = true;
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
  // The server's rule: the hint only when EVERY cast is idle. The wire
  // cap keeps one move per (source, kind, targets_stack), so a capped
  // list can drop a live cast; the digest, which is uncapped, is the
  // answer whenever the server sends it.
  for (const [e, tl] of idle) {
    if (tl.idle === tl.casts && tl.hint) e.cast_idle_hint = tl.hint;
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
    castIdleHint: (id) => get(id)?.cast_idle_hint || undefined,
    readyAbilityRefs: (id) => get(id)?.abilities ?? NONE,
    readyManaRefs: (id) => get(id)?.mana_abilities ?? NONE,
    readySpecialActions: (id) => get(id)?.special_actions ?? NONE,
    canAttack: (id) => get(id)?.kinds.includes("attack") ?? false,
    attackTargets: (id) => get(id)?.attack_targets ?? NONE,
    canExertOnAttack: (id) => get(id)?.exert_on_attack === true,
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

/**
 * acrossActions is the lookup an OPPONENT's panel reads (ADR 0106 §1
 * decisions 5 and 6, #1793). The digest is the viewer's own moves, so
 * the only entry it can hold for a permanent another player controls is
 * an "Any player may activate this ability" row (CR 602.2) the viewer
 * may activate right now. This answers exactly that: the live refs of
 * `cards`' any-player rows, for the cards in it that `viewerID` does
 * not control, and nothing for every other card or question. So the
 * bolt pip and ring on an opponent's permanent mean what they mean on
 * the viewer's own, and nothing else of the viewer's (a combat ring, a
 * cast) leaks onto that panel through this lookup.
 *
 * Pass `visibleHighlights`'s lookup for the drawing and the frame's
 * full lookup for the popover's gate; `known`, `exact` and `pass` are
 * the input's, so the gate keeps its "no information" reading. A
 * spectator (null) gets the lookup that knows nothing.
 */
export function acrossActions(
  actions: LegalActions,
  cards: readonly CardView[],
  viewerID: string | null,
): LegalActions {
  if (!viewerID || !actions.known) return NO_LEGAL_ACTIONS;
  const refs = new Map<string, readonly string[]>();
  for (const c of cards) {
    if ((c.controller || c.owner) === viewerID) continue;
    const own = new Set(
      (c.activated_abilities ?? [])
        .filter((a) => (a.any_player || a.opponents_only || a.owner_only) && a.ref)
        .map((a) => a.ref),
    );
    if (own.size === 0) continue;
    const live = actions.readyAbilityRefs(c.instance_id).filter((r) => own.has(r));
    refs.set(c.instance_id, live);
  }
  const abilities = (id: string): readonly string[] => refs.get(id) ?? NONE;
  const ACTIVATE: readonly ReadyKind[] = Object.freeze(["activate"]);
  return {
    known: actions.known,
    exact: actions.exact,
    pass: actions.pass,
    isReady: (id) => abilities(id).length > 0,
    kinds: (id) => (abilities(id).length > 0 ? ACTIVATE : NO_KINDS),
    castableFrom: () => false,
    castIdleHint: () => undefined,
    readyAbilityRefs: abilities,
    readyManaRefs: () => NONE,
    readySpecialActions: () => NONE,
    canAttack: () => false,
    attackTargets: () => NONE,
    canExertOnAttack: () => false,
    blockableAttackers: () => NONE,
    readyCount: () => 0,
  };
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
 * a spectator. An opponent's card reads `acrossActions`, which lights
 * only its any-player rows (ADR 0106 §1).
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

// ---- presentation: combat (ADR 0105 §2, sub-PR 5) ---------------------

/** What a battlefield click means right now (Game.svelte's combatMode). */
export type CombatMode = "idle" | "attack" | "block";

/**
 * The combat half of the ready treatment, as instance-ID sets.
 *
 * `candidates` are the viewer's creatures the server would let them
 * declare right now: an attacker in declare attackers (`canAttack`:
 * the digest's `attack_targets` is non-empty), a blocker in declare
 * blockers (`blocks` is non-empty). They wear the ready ring.
 *
 * `targets` are what the SELECTED creature may be declared against:
 * the players, planeswalkers and battles in its `attack_targets`, or
 * the attackers in its `blocks`. They wear the ready ring too, drawn
 * so it reads over an attacker's red ring, since an attacker a blocker
 * may block is always attacking (Card.svelte's `combatTarget`).
 */
export interface CombatRings {
  readonly candidates: ReadonlySet<string>;
  readonly targets: ReadonlySet<string>;
}

const NO_IDS: ReadonlySet<string> = Object.freeze(new Set<string>());
export const NO_COMBAT_RINGS: CombatRings = Object.freeze({
  candidates: NO_IDS,
  targets: NO_IDS,
});

/**
 * combatRings reads the combat rings off the lookup. Legality is the
 * digest's alone (`canAttack`, `attackTargets`, `blockableAttackers`).
 * The one decision made here is presentation: a creature already
 * declared (attacking, or blocking with room for another block, #1706)
 * is not a candidate, because its red or blue ring already says it is
 * in combat and the two would fight.
 *
 * Pass the HIGHLIGHT lookup (`legal`): these are rings, and the
 * setting and autopass suppression turn them off. The lookup that
 * knows nothing gives NO_COMBAT_RINGS.
 */
export function combatRings(
  legal: LegalActions,
  mode: CombatMode,
  selectedID: string | null | undefined,
  cards: readonly CardView[],
): CombatRings {
  if (!legal.known || mode === "idle") return NO_COMBAT_RINGS;
  const candidates = new Set<string>();
  for (const c of cards) {
    if (mode === "attack") {
      if (!c.attacking_target && legal.canAttack(c.instance_id)) candidates.add(c.instance_id);
    } else if (!c.blocking_target && legal.blockableAttackers(c.instance_id).length > 0) {
      candidates.add(c.instance_id);
    }
  }
  const targets = new Set<string>(
    !selectedID
      ? []
      : mode === "attack"
        ? legal.attackTargets(selectedID)
        : legal.blockableAttackers(selectedID),
  );
  if (candidates.size === 0 && targets.size === 0) return NO_COMBAT_RINGS;
  return { candidates, targets };
}

/**
 * attackTargetOpen is the gate on a defending PLAYER's click while an
 * attacker is selected: whether the server would accept `attackerID`
 * declared against `targetID`. Read it off the frame's FULL lookup
 * (`legalGate`), never `visibleHighlights`, so the highlight setting
 * opens or shuts nothing.
 *
 * Two cases are "no information" and keep today's rule, which offers
 * every opponent still in the game:
 *   - no list on this frame (an older server, or a seat that owes
 *     nothing);
 *   - an attacker that is already declared. Re-pointing it at another
 *     defender is legal until the declaration locks in, but the
 *     enumerator never lists it (a declared creature drops out of the
 *     list), so there the digest is silent rather than "no".
 */
export function attackTargetOpen(
  gate: LegalActions,
  attacker: CardView | null | undefined,
  attackerID: string,
  targetID: string,
): boolean {
  if (!gate.known) return true;
  if (attacker?.attacking_target) return true;
  return gate.attackTargets(attackerID).includes(targetID);
}

/**
 * attackTargetListed is the stricter question for a PERMANENT defender
 * (a planeswalker, a battle): the digest names it as a target of the
 * selected attacker. Clicking one on the board to declare the attack
 * arrives with its ring (ADR 0105 §7: a ring must point at something
 * the player can do), so no list means no click, which is what the
 * board did before. Read it off the FULL lookup, as attackTargetOpen.
 */
export function attackTargetListed(
  gate: LegalActions,
  attackerID: string | null | undefined,
  targetID: string,
): boolean {
  if (!attackerID || !gate.known) return false;
  return gate.attackTargets(attackerID).includes(targetID);
}

// ---- accessibility (ADR 0105 §7, sub-PR 6) ----------------------------

// The screen-reader phrase for each special-action kind. A kind this
// client has no words for gets the generic phrase, as its pip gets the
// generic star.
const SPECIAL_ACTION_PHRASES: Readonly<Record<string, string>> = Object.freeze({
  foretell: "can be foretold",
  suspend: "can be suspended",
  plot: "can be plotted",
  turn_face_up: "can be turned face up",
  unlock: "has a door you can unlock",
});

/** The phrase for a ready card none of the specific phrases explain. */
export const READY_PHRASE_GENERIC = "has an action available";

/** The phrase a special-action kind adds to a ready card's name. */
export function specialActionPhrase(kind: string): string {
  return Object.hasOwn(SPECIAL_ACTION_PHRASES, kind)
    ? SPECIAL_ACTION_PHRASES[kind]
    : "has a special action";
}

/**
 * readyPhrases is what a ready card's accessible name gains (ADR 0105
 * §7): "castable", "playable land", "has an ability you can activate",
 * "can attack", "can block", a special action's phrase ("can be
 * foretold"), and, for a mana ability worth a drop pip, "has a mana
 * ability you can use". Each phrase is the spoken twin of something
 * drawn (the ring, a pip, a combat ring), so it reads the same lookup
 * the drawing does, the HIGHLIGHT lookup, and applies the §4 noise
 * rule the drop pip applies: a land's ordinary {T} mana ability adds
 * nothing.
 *
 * `combatTarget` is the ring sub-PR 5 draws on what the SELECTED
 * creature may be declared against. That card is an opponent's and has
 * no moves of its own, so it says what may be done TO it: an attacker
 * "can be blocked", a planeswalker or battle "can be attacked".
 *
 * The lookup that knows nothing gives no phrases.
 */
export function readyPhrases(
  legal: LegalActions,
  card: CardView,
  zone: ReadyZone,
  combatTarget = false,
): string[] {
  if (combatTarget) return [card.attacking_target ? "can be blocked" : "can be attacked"];
  const id = card.instance_id;
  if (!legal.isReady(id)) return [];
  const out: string[] = [];
  const add = (p: string) => {
    if (!out.includes(p)) out.push(p);
  };
  if (legal.castableFrom(id, zone)) {
    const kinds = legal.kinds(id);
    // #1918: the muted ring says why it is muted, aloud as well.
    const idle = idleReadyHint(legal, id, zone);
    if (idle) {
      const why = idle.replace(/\.$/, "");
      add(`castable, but ${why.charAt(0).toLowerCase()}${why.slice(1)}`);
    } else if (kinds.includes("cast")) add("castable");
    if (kinds.includes("land")) add("playable land");
  }
  if (legal.readyAbilityRefs(id).length > 0) add("has an ability you can activate");
  if (notableManaRefs(card, legal.readyManaRefs(id), zone).length > 0) {
    add("has a mana ability you can use");
  }
  for (const k of legal.readySpecialActions(id)) add(specialActionPhrase(k));
  // A creature already declared wears its red or blue ring, not the
  // ready one (combatRings), so it is not announced as a candidate.
  if (!card.attacking_target && legal.canAttack(id)) add("can attack");
  if (!card.blocking_target && legal.blockableAttackers(id).length > 0) add("can block");
  return out;
}

/**
 * idleReadyHint is #1918's muted ring: the server's hint when the card
 * is ready in `zone` ONLY because it may be cast there, and every cast
 * it has would do nothing right now (an overloaded Counterflux with no
 * spell to counter). The card still draws a ring, because the cast is
 * legal, but a muted one with this as its tooltip, so it doesn't look
 * like a play worth making.
 *
 * Undefined, and an ordinary ring, the moment the card has anything
 * else: a cast that would do something (the server leaves
 * `cast_idle_hint` off then), a land play, a live ability, a mana
 * ability or a special action. Nothing here judges the board; the hint
 * is the server's, and so is the "every cast" rule.
 */
export function idleReadyHint(
  legal: LegalActions,
  cardID: string,
  zone: ReadyZone,
): string | undefined {
  if (!legal.castableFrom(cardID, zone)) return undefined;
  const hint = legal.castIdleHint(cardID);
  if (!hint) return undefined;
  if (legal.kinds(cardID).some((k) => k !== "cast")) return undefined;
  if (
    legal.readyAbilityRefs(cardID).length > 0 ||
    legal.readyManaRefs(cardID).length > 0 ||
    legal.readySpecialActions(cardID).length > 0
  ) {
    return undefined;
  }
  return hint;
}

/**
 * readyCardLabel is a card's accessible name: its name, plus the
 * phrases when it wears the ready ring. A ring the phrases cannot
 * explain (a caller that drew it without handing the lookup down)
 * still says something, so the ring is never silent.
 */
export function readyCardLabel(name: string, ready: boolean, phrases: readonly string[]): string {
  if (!ready) return name;
  const said = phrases.length > 0 ? phrases.join(", ") : READY_PHRASE_GENERIC;
  return name ? `${name}, ${said}` : said;
}

/**
 * withAvailable is a ready button's or menu row's accessible name: the
 * label plus "available" (ADR 0105 §7). Not ready: the label alone.
 */
export function withAvailable(label: string, ready: boolean): string {
  return ready ? `${label}, available` : label;
}

/** The bolt pip's accessible name: what it opens onto. */
export function boltPipLabel(n: number): string {
  return n >= 2
    ? `${n} abilities you can activate — open actions`
    : "Activate an ability — open actions";
}

/** The drop pip's accessible name. */
export const DROP_PIP_LABEL = "Mana ability available — open actions";

/** The star pip's accessible name: the live kinds, by name. */
export function starPipLabel(kinds: readonly string[]): string {
  const title = specialPipTitle(kinds);
  return `${title.charAt(0).toUpperCase()}${title.slice(1)} available — open actions`;
}

/**
 * The combat pip a candidate wears beside its ring: a sword for a
 * creature that may attack, a shield for one that may block (ADR 0105
 * §7: the kind is carried by pip shape). Decorative: a declaration has
 * no popover behind it, so the pip is no button, and the card's
 * accessible name already says "can attack" or "can block".
 */
export type CombatPip = "attack" | "block";

/**
 * combatPipFor reads a card's combat pip off the rings and the lookup.
 * `ids` is every instance the card stands for (a token group's
 * members), so a group shows a pip when any member is a candidate.
 * Not a candidate: null.
 */
export function combatPipFor(
  legal: LegalActions,
  combat: CombatRings,
  ids: readonly string[],
): CombatPip | null {
  for (const id of ids) {
    if (!combat.candidates.has(id)) continue;
    if (legal.canAttack(id)) return "attack";
    if (legal.blockableAttackers(id).length > 0) return "block";
  }
  return null;
}

/**
 * actionableCount is how many of `seatID`'s cards carry a highlight
 * right now (a ring or a pip): the cards readyPhrases has something to
 * say about, in every zone the board draws them in. It is the N
 * of the phase display's "N actions available". It counts cards, not
 * moves, so the number is what the player sees: one ring per card, not
 * one per target permutation. A basic land's mana is no more an action
 * worth announcing than it is worth a pip (§4).
 *
 * Pass the HIGHLIGHT lookup: with highlights off, or autopass about to
 * pass, it is the lookup that knows nothing and the count is 0.
 */
export function actionableCount(
  legal: LegalActions,
  view: GameView | null | undefined,
  seatID: string,
): number {
  if (!view || !legal.known) return 0;
  const zones: ReadyZone[] = ["hand", "battlefield", "graveyard", "exile", "library", "command"];
  let n = 0;
  for (const zone of zones) {
    for (const c of cardsIn(view, zone, seatID)) {
      if (readyPhrases(legal, c, zone).length > 0) n++;
    }
  }
  return n;
}

/** The live region's sentence for `n` actionable cards. */
export function readyAnnouncement(n: number): string {
  return n === 1 ? "1 action available" : `${n} actions available`;
}

/**
 * The phase display's live-region state. `live` is whether the seat is
 * inside a decision that has already been announced; `text` is what
 * the region holds.
 */
export interface ReadyAnnouncer {
  readonly live: boolean;
  readonly text: string;
}

export const QUIET_ANNOUNCER: ReadyAnnouncer = Object.freeze({ live: false, text: "" });

/**
 * announceArrival is ADR 0105 §7's "once, when priority arrives". Fed
 * each frame's actionableCount, it fills the region on the frame the
 * count goes from nothing to something, and leaves it alone on every
 * later frame of the same decision, so a screen reader hears the line
 * once rather than on every broadcast. When the decision ends (the
 * count drops to 0: priority moved on, highlights went off, autopass
 * took the window) the region empties, so the same sentence is news
 * again next time.
 */
export function announceArrival(prev: ReadyAnnouncer, count: number): ReadyAnnouncer {
  if (count <= 0) return prev.live || prev.text !== "" ? QUIET_ANNOUNCER : prev;
  if (prev.live) return prev;
  return { live: true, text: readyAnnouncement(count) };
}
