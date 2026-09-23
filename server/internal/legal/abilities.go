package legal

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activateParams is the activate_ability wire payload for a catalog
// ability (ability_index set). Free-form S13.1 announcements are not
// enumerated — nothing says what they do.
type activateParams struct {
	SourceCardID string       `json:"source_card_id"`
	AbilityIndex int          `json:"ability_index"`
	Targets      []targetWire `json:"targets,omitempty"`
	// Modes is the CR 602.2b mode choice of a modal activated
	// ability (#764), announced with the targets.
	Modes        []int    `json:"modes,omitempty"`
	SacrificeIDs []string `json:"sacrifice_ids,omitempty"`
	CrewIDs      []string `json:"crew_ids,omitempty"`
	// #625: the permanent a RemoveCounters cost removes from (omitted
	// for the self form) and, for "a counter" of any kind, the kind.
	CounterSourceIDs []string `json:"counter_source_ids,omitempty"`
	// #789: the per-permanent split, for a removal that is spread
	// across several permanents or whose count the activator
	// announces. Omitted for a fixed one-permanent cost.
	CounterCounts []int  `json:"counter_counts,omitempty"`
	CounterKind   string `json:"counter_kind,omitempty"`
	// #943: the per-permanent kind, for the one printed cost whose
	// parts may differ in kind (Tekuthal). Omitted whenever the
	// payment is of one kind, which counter_kind says.
	CounterKinds []string `json:"counter_kinds,omitempty"`
	// #660: the cards paid to a "Discard a creature card" cost.
	// Cycling's "Discard this card" sends none — the source is the
	// payment.
	DiscardIDs []string `json:"discard_ids,omitempty"`
	// #1213: the permanents paid to a "Return a permanent you
	// control to its owner's hand" cost. Omitted for every ability
	// that does not print the clause.
	ReturnIDs []string `json:"return_ids,omitempty"`
	XValue    int      `json:"x_value,omitempty"`
	// #917, CR 107.4f: how many of the mana component's Phyrexian
	// symbols this activation pays with 2 life each. Omitted for
	// every ability that prints none, which is nearly all of them.
	PhyrexianLife int  `json:"phyrexian_life,omitempty"`
	Strict        bool `json:"strict,omitempty"`
	AutoTap       bool `json:"auto_tap,omitempty"`
}

// activatedMoves enumerates catalog activated abilities on the
// seat's permanents. Requires priority (checked by the caller).
// Mirrors game.ActivateCatalogAbility's validation: split second,
// sorcery-speed flag, activation condition (#743), tap cost
// (untapped + not summoning sick for creatures), sacrifice costs
// payable, life cost payable, mana cost affordable, targets legal.
func (e *enumerator) activatedMoves() {
	g := e.g
	if g.SplitSecondActive {
		return
	}
	speed := sorcerySpeedOpen(g, e.seat)
	// #1210: the board-wide fast negative, taken ONCE for the whole
	// pass. Almost no game has a Cursed Totem in it, and without this
	// every ability row would walk the battlefield to be told so.
	restricted := g.AnyActivationRestrictionsForEffect()
	for i := range g.Battlefield.Cards {
		source := &g.Battlefield.Cards[i]
		if source.Controller != e.seat {
			continue
		}
		// CR 602.5a: an Arrested or Fettered permanent's activated
		// abilities are not moves. Same predicate
		// ActivateCatalogAbility gates on, so the engine can never
		// refuse an activation this list offered (#544).
		if !game.CanActivateAbilities(source) {
			continue
		}
		e.abilityMovesForSource(source, game.ZoneBattlefield, speed, restricted)
	}
	// #660, widened by #1221: every OTHER zone an ability can
	// function from. Cycling is an ordinary CR 602 activation whose
	// ability functions from hand (CR 702.29a), unearth is one whose
	// ability functions from a graveyard (CR 702.82a), and both go
	// through the same body below rather than through a second
	// enumerator — the zone predicate is the only thing that differs,
	// which is the whole point of ADR 0062 Decision 1.
	//
	// No CanActivateAbilities call: layer 6 removes a PERMANENT's
	// abilities, and ActivateCatalogAbility does not ask it off the
	// battlefield either. Offering a move the engine refuses, and
	// withholding one it would accept, are both #544.
	for _, zone := range e.abilityZones() {
		if zone.z == nil {
			continue
		}
		for i := range zone.z.Cards {
			src := &zone.z.Cards[i]
			// CR 108.4: off the battlefield and the stack the card's
			// OWNER is the "you" of its printed text, and that is the
			// check ActivateCatalogAbility makes. Free for the
			// per-seat piles, which hold only their owner's cards,
			// and the whole of the answer for exile, which is one
			// shared pile holding everybody's.
			if src.Owner != e.seat {
				continue
			}
			e.abilityMovesForSource(src, zone.kind, speed, restricted)
		}
	}
}

// abilityZone is one non-battlefield pile the activation walk visits,
// with the ZoneKind an ability's declaration is written in.
type abilityZone struct {
	z    *game.Zone
	kind game.ZoneKind
}

// abilityZones are the places a CR 602 activation can come out of
// besides the battlefield, in the order the moves are emitted. The
// activation sibling of castZones (cast.go, #1014), and deliberately
// shaped the same way: one walk, the zone a card is in deciding
// nothing but which declarations match it.
//
// Four piles rather than five. The LIBRARY is missing, and that is a
// rule rather than an omission: CR 401.2 makes a library hidden, no
// printed ability functions from one, and an enumerator that walked
// it would be reading cards the seat is not entitled to see in order
// to answer a question whose answer is always "nothing". If a card
// ever prints such an ability, this is the one line it needs — plus
// CR 401.5's trim to the top card, which the cast walk already shows
// how to write.
//
// Exile is the seat's own cards in the one shared pile, filtered by
// the caller: g.Exile holds every seat's exiled cards, and CR 108.4
// gives each of them to its owner.
func (e *enumerator) abilityZones() []abilityZone {
	g, p := e.g, e.p
	if p == nil {
		return nil
	}
	return []abilityZone{
		{p.Hand, game.ZoneHand},
		{p.Graveyard, game.ZoneGraveyard},
		{p.Command, game.ZoneCommand},
		{g.Exile, game.ZoneExile},
	}
}

// abilityMovesForSource enumerates one card's activated abilities,
// from the zone that card is actually in. Split out of activatedMoves
// so the battlefield loop and the hand loop share every line of the
// cost solve: the mana and Phyrexian affordability, the sacrifice,
// crew, counter and discard payments, the mode and target expansion
// and the per-source budget.
//
// The ability list comes from game.ActivatedAbilitiesForCard, which is
// the one accessor every consumer reads through — so a designation-
// gated ability (ADR 0071) is absent here exactly as it is absent from
// the activation path and the wire.
func (e *enumerator) abilityMovesForSource(source *game.Card, zone game.ZoneKind, speed, restricted bool) {
	g, p := e.g, e.p
	abilities := game.ActivatedAbilitiesForCard(*source)
	// #662: an activated ability's source is the permanent — or, since
	// #660, the HAND CARD — that has it, which is what CR 702.16b tests
	// protection against. Never the seat: a red player's colourless
	// artifact ability may still point at a pro-red creature.
	abilitySrc := game.SourceObject(e.seat, source)
	for idx, ab := range abilities {
		// CR 113.6: the ability has to function from the zone the
		// card is in. Same predicate the engine gates on, so a
		// hand card's battlefield abilities are never offered and
		// a permanent's cycling never is either.
		if !game.AbilityFunctionsFromZone(ab, zone) {
			continue
		}
		// #1210, CR 602.5a: the board-wide "can't be activated"
		// gate — Cursed Totem, Linvala, Collector Ouphe, Pithing
		// Needle. The SAME function ActivateCatalogAbility calls,
		// in the same place relative to the zone and timing checks,
		// so a policy is never offered an activation the engine
		// refuses (#544). Behind the board-wide fast negative the
		// caller took once: nothing restricts anything in almost
		// every game, and the walk is otherwise per ability.
		if restricted && g.ActivationGateLocked(e.seat, *source, zone,
			game.ActivationAbility{Label: ab.Label}) != nil {
			continue
		}
		if (ab.SorcerySpeed || ab.Cost.Loyalty != nil) && !speed {
			continue
		}
		// CR 602.1b (#743): the "Activate only if …" gate, with
		// the same arguments ActivateCatalogAbility passes, so a
		// policy is never offered Tectonic Edge while no
		// opponent has four lands (#544).
		if ab.Condition != nil && !ab.Condition(g, e.seat, source.InstanceID) {
			continue
		}
		// #1181: "Activate each exhaust ability only once". Same
		// reader ActivateCatalogAbility and the view use, so a policy
		// is never offered an exhaust ability this object has already
		// spent (#544).
		if g.AbilityExhausted(e.seat, source.InstanceID, ab) {
			continue
		}
		// CR 606.3 / 606.5: a loyalty ability of a PERMANENT you
		// control, one activation per turn, and enough counters to
		// pay a −N. Mirrors ActivateCatalogAbility so a policy never
		// proposes a move the engine will bounce — including the
		// IsPlaneswalker test both sides dropped in #1157, because
		// CR 606 asks about the permanent and its cost symbol, not
		// about the card type.
		if ab.Cost.Loyalty != nil {
			if g.LoyaltyActivatedThisTurn[source.InstanceID] {
				continue
			}
			if n := *ab.Cost.Loyalty; n < 0 && source.Counters[game.CounterLoyalty] < -n {
				continue
			}
		}
		if ab.Cost.Tap {
			if source.Tapped {
				continue
			}
			if source.IsCreature() && game.HasSummoningSickness(source) {
				continue
			}
		}
		// CR 119.4 and CR 119.8, through the one predicate
		// ActivateCatalogAbility validates with (#544, #1200): a
		// policy is never offered a life cost the engine refuses,
		// including one a locked life total makes unpayable.
		if !g.CanPayLifeLocked(p, ab.Cost.Life) {
			continue
		}
		// CR 602.2b: X is announced with the activation, so the
		// enumerator has to pick one. It picks the LARGEST
		// affordable value at or above the cost's printed floor,
		// exactly as castMovesForCard does for an {X} spell, and
		// emits ONE move for it.
		//
		// One move, not one per value in 0..MaxX, and that is the
		// #544 lesson applied rather than re-learned: the
		// expansion budget below is shared with the target and
		// sacrifice sets, so an X that added an arity to the
		// cross product would spend the budget on near-duplicate
		// activations of the first target and never reach the
		// second. X consumes no budget at all here.
		//
		// The floor is the other half, and since #810 it has two
		// sources, both settled by enumeratedXFloor (x.go).
		// Helm of Obedience's printed "X can't be 0" means an
		// activator who cannot afford X=1 has no legal
		// activation, and offering one at X=0 would be exactly
		// the bug #544 describes — an enumeration the engine
		// refuses. Soothsaying's "{X}: Look at the top X cards"
		// prints no floor, so X=0 IS legal (CR 602.2b) and the
		// engine accepts it — but it costs nothing, does
		// nothing, and comes straight back, which is CR 732.2a's
		// repeatable no-op. Neither is a move worth offering, so
		// both are answered by the same floor.
		xValue := 0
		phyrexianLife := 0
		if ab.Cost.Mana != "" {
			// #1184: the PRICED cost, not the printed one — the same
			// function ActivateCatalogAbility pays through, so a Boom
			// Scholar's "{2} less to activate" is visible to the
			// policy as an activation it can now afford rather than
			// one it is never offered. #544's rule with the sign the
			// other way round: an enumerator that priced at the
			// printed cost would silently hide legal moves.
			cost, err := g.AbilityManaCostForEffect(e.seat, *source, zone, ab)
			if err != nil {
				continue
			}
			// #352: an activated ability's mana is paid under an
			// activation context keyed on the SOURCE permanent,
			// mirroring payAbilityManaCostLocked. So is the
			// exclusion: a {T} ability cannot tap its own source
			// for mana, and an enumerator that thought it could
			// would offer activations the engine refuses.
			var excluded map[uuid.UUID]bool
			if ab.Cost.Tap {
				excluded = map[uuid.UUID]bool{source.InstanceID: true}
			}
			floor := enumeratedXFloor(game.CatalogAbilityKey(*source), ab.Cost.FloorX())
			// #917, CR 107.4f: the announcement has TWO numbers
			// when the cost prints a Phyrexian symbol — the X and
			// how many symbols are paid with 2 life each — and
			// they are solved together, because striking a symbol
			// changes what X the pool can afford.
			x, life, ok := e.affordablePayment(cost, game.ManaSpendForAbility(*source), floor, ab.Cost.Life, excluded)
			if !ok {
				continue
			}
			xValue, phyrexianLife = x, life
		} else if ab.Cost.DemandsX() {
			// Unreachable — DemandsX reads the same string — but
			// a cost that demanded X with no mana component
			// would be an unannouncable ability, so refuse it
			// rather than emit an activation at X=0.
			continue
		}
		// Crew (CR 702.122a). The engine rejects a crew
		// activation that names no creatures, so an enumerator
		// that skipped this offered a move that could only ever
		// bounce — the same class of defect as #544, one cost
		// component over. The pick is the cheapest set that
		// clears the number; any set the engine accepts is a
		// legal answer, and offering all of them would be a
		// combinatorial expansion for a choice the policy has no
		// information to make.
		var crewIDs []uuid.UUID
		if ab.Cost.Crew > 0 {
			crewIDs = e.crewPayment(ab.Cost.Crew)
			if crewIDs == nil {
				continue
			}
		}
		sacrificeSets := [][]uuid.UUID{nil}
		if ab.Cost.SacrificeOther != nil {
			pool := e.sacrificePool(source.InstanceID, ab.Cost.SacrificeSelf, ab.Cost.SacrificeOther)
			// #1213: a VARIABLE count is an announcement, so the
			// enumerator offers a bounded ladder of counts rather
			// than one payment — see variableSacrificePayments.
			if game.SacrificeCostVariable(ab.Cost.SacrificeOther) {
				sacrificeSets = e.variableSacrificePayments(pool, ab.Cost.SacrificeOther, source.InstanceID)
			} else {
				sacrificeSets = e.sacrificePayments(pool, ab.Cost.SacrificeOther, source.InstanceID)
			}
			if len(sacrificeSets) == 0 {
				continue
			}
		}
		// #1213: a "Return a permanent you control to its owner's
		// hand" cost. One move per candidate, cheapest-to-keep first
		// and then in payment order, so a capped budget spends itself
		// on the permanent the seat misses least. Nothing payable
		// means no move at all — #544.
		returnSets := [][]uuid.UUID{nil}
		if rc := ab.Cost.ReturnToHand; !rc.Empty() {
			returnSets = e.returnPayments(g.ReturnToHandOptionsForEffect(e.seat, source.InstanceID, rc), rc, source.InstanceID)
			if len(returnSets) == 0 {
				continue
			}
		}
		// #625: a "remove N counters" cost. One move per (permanent,
		// kind) that could pay, most counters first — so a capped
		// budget spends itself on the payments that hurt least
		// (a 6-loyalty walker before a 1-loyalty one) — and the
		// ability is not offered at all when nothing can pay. The
		// candidates come from the same walk the engine's option
		// view and validation agree with, so an offered payment is
		// one ActivateCatalogAbility accepts (#544).
		counterChoices := []counterChoice{{}}
		if rc := ab.Cost.RemoveCounters; rc != nil {
			counterChoices = counterPaymentChoices(g.CounterCostOptionsForEffect(e.seat, source.InstanceID, rc), rc)
			if len(counterChoices) == 0 {
				continue
			}
		}
		// #789: CR 118.3's other direction — a cost that PUTS a
		// counter on cannot be paid by a permanent that can't
		// have one. Same predicate ActivateCatalogAbility
		// refuses on, so the list never offers what the engine
		// bounces (#544).
		if ac := ab.Cost.AddCounter; ac != nil && !g.CanPlaceCounterForEffect(e.seat, source.InstanceID, ac) {
			continue
		}
		// #660: a "Discard N cards" cost. Solved the way crew is —
		// ONE payment, the cheapest set, rather than one move per
		// subset: a discard's subsets are the powerset of a
		// seven-card hand, and the policy has no information here
		// that a different pick would use. "Cheapest" is hand
		// order over the cards the clause admits, with the source
		// excluded (an ability activated from hand cannot pay
		// itself). Nothing payable means no move at all — #544.
		var discardIDs []uuid.UUID
		if dc := ab.Cost.DiscardCards; dc != nil {
			opts := g.DiscardCostOptionsForEffect(e.seat, source.InstanceID, dc)
			if len(opts) < dc.N {
				continue
			}
			discardIDs = opts[:dc.N]
		}
		budget := e.opts.MaxExpansionPerSource
		// #764: a modal activated ability announces its modes with
		// its targets (CR 602.2b), so the enumerator expands the
		// same product a modal cast does.
		modeSets := [][]int{nil}
		if ab.Modes != nil {
			modeSets = e.legalModeSets(abilitySrc, ab.Modes)
			if len(modeSets) == 0 {
				continue
			}
		}
		type announcement struct {
			modes   []int
			targets []game.TargetRef
		}
		var announcements []announcement
		for _, modes := range modeSets {
			steps := game.AnnouncedClauses(ab.Targets, ab.Modes, modes)
			sets := e.legalStepSets(abilitySrc, steps, budget)
			for _, ts := range sets {
				announcements = append(announcements, announcement{modes: modes, targets: ts})
			}
		}
		if len(announcements) == 0 {
			continue
		}
		// #74: the life and loyalty components ride the Move
		// rather than the params, because the params are the
		// action payload and the dispatcher reads neither — it
		// reads them off the ability. Without this a policy
		// cannot tell "Pay 7 life: Draw seven cards" from a
		// free ability and activates itself to death.
		loyalty := 0
		if ab.Cost.Loyalty != nil {
			loyalty = *ab.Cost.Loyalty
		}
		for _, ann := range announcements {
			targets := ann.targets
			for _, sacs := range sacrificeSets {
				// #1213: "Sacrifice X Treasures" announces its count
				// as X (CR 602.2b), so the move's x_value IS the
				// payment it carries. Register refuses a cost that
				// also puts {X} in its mana component, so there is
				// never a second claimant on this number.
				xValue := xValue
				if game.SacrificeCountFromX(ab.Cost.SacrificeOther) {
					xValue = len(sacs)
				}
				for _, rets := range returnSets {
					for _, cc := range counterChoices {
						if budget <= 0 {
							break
						}
						budget--
						label := source.Name + ": " + ab.Label
						if xValue > 0 {
							label += fmt.Sprintf(" for X=%d", xValue)
						}
						if phyrexianLife > 0 {
							label += fmt.Sprintf(" paying %d life for Phyrexian mana",
								phyrexianLife*game.PhyrexianLifePerSymbol)
						}
						label += sacrificeLabel(g, sacs)
						label += returnLabel(g, rets)
						label += cc.label(g)
						label += targetLabel(g, targets)
						// #74: the life on the Move is what the
						// controller pays at announce, so the
						// Phyrexian half counts — a policy that saw
						// only the printed component would read a
						// four-life activation as free.
						cost := moveCost(ab.Cost.Life+phyrexianLife*game.PhyrexianLifePerSymbol, loyalty)
						for _, price := range cc.prices() {
							cost = withCounterPrice(cost, price)
						}
						e.add(Move{
							Type:   TypeActivateAbility,
							Player: e.seat,
							Kind:   KindActivate,
							Label:  label,
							Source: source.InstanceID,
							Cost:   cost,
							Params: mustJSON(activateParams{
								SourceCardID:     source.InstanceID.String(),
								AbilityIndex:     idx,
								Targets:          wireTargets(targets),
								Modes:            ann.modes,
								SacrificeIDs:     idStrings(sacs),
								CrewIDs:          idStrings(crewIDs),
								CounterSourceIDs: cc.wireIDs(),
								CounterCounts:    cc.wireCounts(),
								CounterKind:      cc.wireKind(),
								CounterKinds:     cc.wireKinds(),
								DiscardIDs:       idStrings(discardIDs),
								ReturnIDs:        idStrings(rets),
								XValue:           xValue,
								PhyrexianLife:    phyrexianLife,
								Strict:           true,
								AutoTap:          true,
							}),
						})
					}
				}
			}
		}
	}
}

// maxEnumeratedVariableCounts caps how many different COUNTS the
// enumerator offers for one variable-count cost — "Sacrifice one or
// more artifacts" (Radiant Lotus), "Sacrifice X Treasures" (Grim
// Hireling). Not a rule; a policy, documented in docs/bot.md beside
// maxEnumeratedCostPayments and maxEnumeratedRepeats.
//
// THREE, and the reason is maxEnumeratedCostPayments' corollary
// (ADR 0033 §1) rather than squeamishness about combinatorics: a
// variable in a COST must not become an arity of the target/mode cross
// product. An open count over a ten-artifact board is ten counts, each
// of them the same ability with the same target at a different size,
// and a budget spent ten ways here would never reach a second target.
//
// The counts offered are the SMALLEST ones: the floor, and the next
// two above it. Small is the conservative direction for a cost — it
// spends the least board — and the permanents inside each payment are
// already the cheapest the seat's own policy can name
// (cheapestFuelFirst), so payment number one is the best answer the
// policy has and the next two are that answer paying more.
const maxEnumeratedVariableCounts = 3

// variableSacrificePayments turns a VARIABLE sacrifice clause's
// candidate pool into the payments the enumerator offers (#1213).
//
// One payment per COUNT, from the clause's floor upward, capped at
// maxEnumeratedVariableCounts and at what the board can actually pay.
// Each payment is a prefix of one order: the seat's own fuel price
// first (cheapestFuelFirst — what is this permanent worth to KEEP),
// then game.SacrificePaymentOrderForEffect's policy-neutral tie-break
// (tokens, then lower mana value, then the source last). So the counts
// nest, and a bot asked to sacrifice three eats the same two it would
// have eaten to sacrifice two.
//
// The floor is at least ONE even for "Sacrifice X", whose printed
// floor is zero: an announcement of X=0 sacrifices nothing and does
// nothing, which is #810's "a move whose whole effect is X is not
// worth offering at X=0" one component over.
//
// Nil when the board cannot reach the floor, so the ability is not
// offered at all (#544).
func (e *enumerator) variableSacrificePayments(pool []uuid.UUID, spec *game.TargetSpec, sourceID uuid.UUID) [][]uuid.UUID {
	lo, _ := game.SacrificeCostBounds(spec, 0)
	if lo < 1 {
		lo = 1
	}
	if len(pool) < lo {
		return nil
	}
	ordered := e.g.SacrificePaymentOrderForEffect(e.cheapestFuelFirst(pool), sourceID)
	var out [][]uuid.UUID
	for n := lo; n <= len(ordered) && len(out) < maxEnumeratedVariableCounts; n++ {
		// The engine validates the count with the same predicate, so
		// a payment offered here is one the announce path accepts.
		// For a CountFromX clause the announced X IS the count, which
		// is why both arguments are n.
		if !game.SacrificeCountLegal(spec, n, n) {
			break
		}
		out = append(out, ordered[:n])
	}
	return out
}

// returnPayments turns a return-to-hand clause's candidate pool into
// the payments the enumerator offers (#1213) — one move per candidate
// for the one-permanent clause every printed card has, and the first
// Count of the order for a hypothetical larger one.
//
// The order is cheapestFuelFirst then SacrificePaymentOrderForEffect,
// the same two-stage order the sacrifice payments use, because the
// question is the same one: which permanent does this seat miss least.
// A capped budget therefore returns the token before the bomb.
//
// Nil when the pool cannot reach the clause's count, so the ability is
// not offered at all (#544) — which is CR 118.3 read through the
// enumerator.
func (e *enumerator) returnPayments(pool []uuid.UUID, rc *game.ReturnToHandCost, sourceID uuid.UUID) [][]uuid.UUID {
	if rc.Empty() || len(pool) < rc.Count {
		return nil
	}
	ordered := e.g.SacrificePaymentOrderForEffect(e.cheapestFuelFirst(pool), sourceID)
	if rc.Count > 1 {
		return [][]uuid.UUID{ordered[:rc.Count]}
	}
	return combinations(ordered, 1, 1, e.opts.MaxExpansionPerSource)
}

// sacrificeLabel names what a payment eats, so two moves that differ
// only in which permanent paid read differently in a log or a trace.
// Empty for a payment of nothing.
func sacrificeLabel(g *game.Game, ids []uuid.UUID) string {
	if len(ids) == 0 {
		return ""
	}
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = cardName(g, id)
	}
	return " (sacrificing " + strings.Join(names, ", ") + ")"
}

// returnLabel is sacrificeLabel one verb over (#1213).
func returnLabel(g *game.Game, ids []uuid.UUID) string {
	if len(ids) == 0 {
		return ""
	}
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = cardName(g, id)
	}
	return " (returning " + strings.Join(names, ", ") + ")"
}

// affordablePayment solves an activated ability's mana component for
// the pair CR 602.2b makes the activator announce: the X, and how
// many of the cost's Phyrexian symbols are paid with 2 life each
// (CR 107.4f, #917). Returns false when nothing the activator can
// announce pays for it, which is what stops the ability being offered
// at all (#544).
//
// MANA FIRST, always. A life payment is a real cost, so the
// enumerator never spends a life total to save mana the board could
// have produced — it reaches for life only when the mana half alone
// cannot pay, and then takes the FIRST count that works, which is the
// cheapest in life. Without this Birthing Pod would simply not be
// offered to a seat holding {1} and no green source, which is the
// gap #917 names.
//
// Bounded twice, and both bounds are the ones the engine validates
// against, so an offered payment is one ActivateCatalogAbility
// accepts: by the symbols the cost actually prints, and by CR 119.4 —
// a player may pay life only down to 0. `reservedLife` is the
// ability's own printed Life component, held back so the two
// together can never claim more than the seat has.
//
// The strike itself is game.PhyrexianLifePlan — the same function the
// engine reduces the cost with, reading the same pool — so the
// enumerator and the payment cannot disagree about which symbol the
// life buys.
func (e *enumerator) affordablePayment(
	cost game.ParsedCost,
	spend game.ManaSpendContext,
	floor int,
	reservedLife int,
	excluded map[uuid.UUID]bool,
) (int, int, bool) {
	if x, ok := e.affordableXExcluding(cost, spend, floor, excluded); ok {
		return x, 0, true
	}
	budget := e.p.Life - reservedLife
	for n := 1; n <= cost.PhyrexianSymbols(); n++ {
		reduced, life := game.PhyrexianLifePlan(cost, e.p.ManaPool, spend, n)
		if life > budget {
			break
		}
		if x, ok := e.affordableXExcluding(reduced, spend, floor, excluded); ok {
			return x, n, true
		}
	}
	return 0, 0, false
}

// counterPart is one permanent's share of a counter payment: which
// kind comes off it, and how many. The kind is per part because
// #943's any-kind among cost is paid in whatever kinds the
// permanents hold; every other shape repeats one kind.
type counterPart struct {
	cardID uuid.UUID
	kind   string
	n      int
}

// counterChoice is one way to pay a RemoveCounters cost: the zero
// value stands for "the ability has no counter component".
type counterChoice struct {
	parts []counterPart
	total int
	// count orders the choices: the counters on the permanent the
	// payment comes off, so a capped budget spends itself on the
	// payments that hurt least (a 6-loyalty walker before a
	// 1-loyalty one).
	count int
	// fromOther is true for the "from a permanent you control" and
	// "from among …" forms, which name the permanents on the wire;
	// the self form does not.
	fromOther bool
	// anyKind is true when the COST prints no kind, so the payment
	// has to name the kinds it removed — counter_kind when the parts
	// share one, counter_kinds when they do not (#943).
	anyKind bool
	// explicitCounts is true when the payment has to spell out the
	// per-permanent split — an among payment, and a variable one,
	// where there is no printed count for the engine to assume.
	explicitCounts bool
}

// oneKind is the kind every part removes, or "" when the parts mix
// kinds — which only an any-kind among payment can do.
func (cc counterChoice) oneKind() string {
	if len(cc.parts) == 0 {
		return ""
	}
	kind := cc.parts[0].kind
	for _, p := range cc.parts[1:] {
		if p.kind != kind {
			return ""
		}
	}
	return kind
}

// counterPaymentChoices turns the engine's (permanent, kinds) options
// into the payments the enumerator offers — one move per payment, and
// the same "one answer, not every answer" discipline crewPayment and
// sacrificePayments follow (#544: an expansion the budget cannot
// afford starves the target loop).
//
//	fixed, one permanent   one choice per (permanent, kind), most
//	                       counters first — unchanged from #625.
//	among                  ONE payment: take from the permanents
//	                       holding the most first until the clause's
//	                       N is covered. The sets differ only in which
//	                       permanents are drained, and the policy has
//	                       nothing to choose between them with.
//	any-kind among         the same walk, with every kind in the pool
//	                       rather than one (#943, Tekuthal): the parts
//	                       carry the kinds they drain, and the order
//	                       is still most counters first, within a
//	                       permanent as well as across them.
//	variable               ONE payment per permanent: every counter it
//	                       holds. A variable removal is printed on
//	                       cards that turn counters into mana, so the
//	                       largest payment is the only one worth
//	                       offering — a bot that took fewer would be
//	                       choosing to waste them.
//
// Empty when the cost cannot be paid at all, which is what stops the
// ability being offered (#544).
func counterPaymentChoices(opts []game.CounterCostOption, rc *game.CounterRemovalCost) []counterChoice {
	if rc == nil {
		return []counterChoice{{}}
	}
	if rc.Among {
		// One pool, drained biggest first. For a printed kind the
		// pool is that kind's counters; for an any-kind among cost
		// every kind on every matched permanent is in it, and a part
		// records which kind it took (#943). Bounded by the board
		// either way: at most one part per (permanent, kind), and
		// still exactly ONE payment.
		var parts []counterPart
		need, best := rc.N, 0
		for _, o := range opts {
			if need <= 0 {
				break
			}
			for _, k := range o.Kinds {
				if need <= 0 {
					break
				}
				if rc.Counter != "" && k.Kind != rc.Counter {
					continue
				}
				take := k.Count
				if take > need {
					take = need
				}
				parts = append(parts, counterPart{cardID: o.CardID, kind: k.Kind, n: take})
				need -= take
				if k.Count > best {
					best = k.Count
				}
			}
		}
		if need > 0 {
			return nil
		}
		return []counterChoice{{
			parts: parts, total: rc.N, count: best,
			fromOther: true, anyKind: rc.Counter == "", explicitCounts: true,
		}}
	}
	fromOther := rc.From != nil
	var out []counterChoice
	for _, o := range opts {
		for _, k := range o.Kinds {
			n := rc.N
			if rc.Variable {
				n = k.Count
			}
			if n < 1 {
				continue
			}
			out = append(out, counterChoice{
				parts:          []counterPart{{cardID: o.CardID, kind: k.Kind, n: n}},
				total:          n,
				count:          k.Count,
				fromOther:      fromOther,
				anyKind:        rc.Counter == "",
				explicitCounts: rc.Variable,
			})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].count > out[j].count })
	return out
}

func (cc counterChoice) wireIDs() []string {
	if !cc.fromOther {
		return nil
	}
	out := make([]string, 0, len(cc.parts))
	for _, p := range cc.parts {
		out = append(out, p.cardID.String())
	}
	return out
}

// wireCounts is the per-permanent split, sent only when the engine
// cannot assume the printed count: an among payment and a variable
// one. A fixed single-permanent payment sends nothing, which is the
// shape every #625 client already speaks.
func (cc counterChoice) wireCounts() []int {
	if !cc.explicitCounts {
		return nil
	}
	out := make([]int, 0, len(cc.parts))
	for _, p := range cc.parts {
		out = append(out, p.n)
	}
	return out
}

// wireKind is the kind sent as counter_kind: set only when the COST
// prints no kind and the payment is of ONE kind, which is every
// any-kind payment but a mixed Tekuthal split.
func (cc counterChoice) wireKind() string {
	if !cc.anyKind {
		return ""
	}
	return cc.oneKind()
}

// wireKinds is #943's per-part kind array, sent only when the parts
// do not share a kind — an any-kind AMONG payment that mixes them.
// Everything else says its kind once in counter_kind, so a move this
// enumerator builds is a payload every client already speaks.
func (cc counterChoice) wireKinds() []string {
	if !cc.anyKind || cc.oneKind() != "" {
		return nil
	}
	out := make([]string, 0, len(cc.parts))
	for _, p := range cc.parts {
		out = append(out, p.kind)
	}
	return out
}

// prices is the counter component of the move's cost, one entry per
// permanent the payment drains, in the kind it drains, so a policy
// sees what the payment costs IT rather than only what the card
// charges.
func (cc counterChoice) prices() []CounterPrice {
	out := make([]CounterPrice, 0, len(cc.parts))
	for _, p := range cc.parts {
		out = append(out, CounterPrice{CardID: p.cardID, Counter: p.kind, N: p.n})
	}
	return out
}

// label names the payment in the move label when it is a choice the
// reader could not infer from the ability text: which permanents, how
// many came off each, and for "a counter", which kind.
func (cc counterChoice) label(g *game.Game) string {
	if cc.total == 0 {
		return ""
	}
	if len(cc.parts) == 1 && !cc.fromOther && !cc.anyKind && !cc.explicitCounts {
		// A printed self cost with a printed count: the ability text
		// already says it.
		return ""
	}
	kind := cc.oneKind()
	froms := make([]string, 0, len(cc.parts))
	for _, p := range cc.parts {
		name := cardName(g, p.cardID)
		switch {
		case kind == "":
			// A payment that mixes kinds names the kind per part —
			// "2 +1/+1 from Bear, 1 loyalty from Jace" — because the
			// reader cannot infer it from the ability text or from
			// the total (#943).
			name = fmt.Sprintf("%d %s from %s", p.n, p.kind, name)
		case len(cc.parts) > 1 || cc.total != p.n:
			name = fmt.Sprintf("%d from %s", p.n, name)
		default:
			name = "from " + name
		}
		froms = append(froms, name)
	}
	if kind == "" {
		return fmt.Sprintf(" (removing %d counters: %s)", cc.total, strings.Join(froms, ", "))
	}
	what := kind + " counter"
	if cc.total > 1 {
		what = fmt.Sprintf("%d %s counters", cc.total, kind)
	}
	return " (removing " + what + " " + strings.Join(froms, ", ") + ")"
}

// sacrificePool lists the seat's permanents that satisfy a
// SacrificeOther clause, excluding the source itself when the cost
// also sacrifices the source (the engine rejects paying one
// permanent twice).
func (e *enumerator) sacrificePool(sourceID uuid.UUID, selfToo bool, spec *game.TargetSpec) []uuid.UUID {
	// Sacrificing is a cost, not targeting, so this is the
	// candidate walk rather than the legal-TARGET walk — the
	// hexproof / shroud gate must not shrink the pool.
	lt := e.g.SpecCandidatesForEffect(e.seat, spec)
	var pool []uuid.UUID
	for _, id := range lt.Cards {
		if selfToo && id == sourceID {
			continue
		}
		if c := findBattlefield(e.g, id); c != nil && c.Controller == e.seat {
			pool = append(pool, id)
		}
	}
	return pool
}

// sacrificePayments turns a sacrifice clause's candidate pool into the
// payments the enumerator offers, one per move (#747, ADR 0020
// addendum §15). `sourceID` is the ability's source, or uuid.Nil for a
// spell's additional cost.
//
//   - N = 1: every candidate is its own payment, in board order, up to
//     MaxExpansionPerSource — unchanged from before #747.
//   - N ≥ 2: ONE payment, the first N of
//     game.SacrificePaymentOrderForEffect (tokens first, then lower
//     mana value, then the source last, then board order). The sets
//     differ only in which permanents are lost, and the target and
//     sacrifice loops share one budget: ten Treasures choose five is
//     252 sets, which would spend the whole budget on the first target
//     and never reach the second (#544). crewPayment answers the same
//     kind of choice with one answer.
//
// Nil when the pool has fewer than N candidates, so the ability or
// spell is not offered at all (#544).
func (e *enumerator) sacrificePayments(pool []uuid.UUID, spec *game.TargetSpec, sourceID uuid.UUID) [][]uuid.UUID {
	n := game.SacrificeCostCount(spec)
	if n <= 1 {
		return combinations(pool, 1, 1, e.opts.MaxExpansionPerSource)
	}
	if len(pool) < n {
		return nil
	}
	ordered := e.g.SacrificePaymentOrderForEffect(pool, sourceID)
	return [][]uuid.UUID{ordered[:n]}
}

// crewPayment picks a set of untapped creatures the seat controls
// whose total effective power reaches `crew` (CR 702.122a), or nil
// when no such set exists.
//
// Greedy from the biggest power down, so the set is as small as the
// board allows and the fewest blockers are spent. Summoning sickness
// is deliberately NOT filtered — tapping to crew is not paying a
// {T} cost, so a creature cast this turn may crew (CR 702.122b), and
// game.validateCrewCostLocked agrees.
//
// One answer, not every answer: the printed number is a floor, so a
// five-power creature crews a Vehicle that says 3 and so does a pair
// of two-power ones. Enumerating the subsets would be an exponential
// expansion for a choice the policy has nothing to decide it with.
func (e *enumerator) crewPayment(crew int) []uuid.UUID {
	type candidate struct {
		id    uuid.UUID
		power int
	}
	var pool []candidate
	for i := range e.g.Battlefield.Cards {
		c := &e.g.Battlefield.Cards[i]
		if c.Controller != e.seat || !c.IsCreature() || c.Tapped {
			continue
		}
		pool = append(pool, candidate{id: c.InstanceID, power: c.CurrentPower()})
	}
	sort.SliceStable(pool, func(i, j int) bool { return pool[i].power > pool[j].power })
	total := 0
	var out []uuid.UUID
	for _, c := range pool {
		if total >= crew {
			break
		}
		// A 0-power creature can never move the total, and naming it
		// would only tap a blocker for nothing.
		if c.power <= 0 {
			continue
		}
		out = append(out, c.id)
		total += c.power
	}
	if total < crew {
		return nil
	}
	return out
}

type manaParams struct {
	CardID       string   `json:"card_id"`
	AbilityIndex int      `json:"ability_index"`
	SacrificeIDs []string `json:"sacrifice_ids,omitempty"`
	// #789: a mana ability's counter cost is paid with exactly the
	// fields an activated ability's is — one component, one payment
	// shape, whichever ability kind carries it.
	CounterSourceIDs []string `json:"counter_source_ids,omitempty"`
	CounterCounts    []int    `json:"counter_counts,omitempty"`
	CounterKind      string   `json:"counter_kind,omitempty"`
	CounterKinds     []string `json:"counter_kinds,omitempty"`
	// #1213: a mana ability's discard cost is paid with exactly the
	// field an activated ability's is, for the same reason the
	// counter fields above share theirs — Skirge Familiar's
	// "Discard a card: Add {B}" is #660's component with a second
	// owner, not a second component.
	DiscardIDs []string `json:"discard_ids,omitempty"`
}

// manaMoves enumerates mana abilities on the seat's permanents and —
// since #1228 — on the seat's own cards in every other zone a mana
// ability can function from (CR 113.6).
// Casts already auto-tap, so a policy rarely needs these for plain
// "{T}: Add" sources; they matter for sacrifice sources (Treasure,
// Lotus Petal, Ashnod's Altar) the auto-tapper never touches.
// Requires priority (checked by the caller). Mana abilities ignore
// split second (CR 702.61b).
func (e *enumerator) manaMoves() {
	g := e.g
	// #1210: the board-wide "can't be activated" fast negative, taken
	// once for the whole pass, exactly as activatedMoves takes it.
	restricted := g.AnyActivationRestrictionsForEffect()
	for i := range g.Battlefield.Cards {
		source := &g.Battlefield.Cards[i]
		if source.Controller != e.seat {
			continue
		}
		// CR 602.5a, the mana half: Arrest stops these too, Faith's
		// Fetters deliberately does not. Same predicate
		// ActivateManaAbility gates on (#544).
		if !game.CanActivateManaAbilities(source) {
			continue
		}
		e.manaMovesForSource(source, game.ZoneBattlefield, restricted)
	}
	// #1228: the walk activatedMoves has made since #660, for the
	// CR 605 ability kind. A Spirit Guide's "Exile this card from
	// your hand: Add {R}" is an ordinary mana ability whose ability
	// functions from a hand, so it goes through the same body below
	// rather than through a second enumerator — the zone predicate is
	// the only thing that differs.
	//
	// abilityZones() is reused whole rather than trimmed to the one
	// pile game.supportedManaAbilityZones admits: the per-ability
	// predicate inside the body is what decides, and walking the same
	// four piles the activated half already walks keeps ONE list of
	// "where can an ability come from" in this file. The piles are
	// scanned either way, for the abilities loop above them.
	//
	// No CanActivateManaAbilities call, for the reason the activated
	// walk makes none: layer 6 removes a PERMANENT's abilities, and
	// ActivateManaAbility does not ask it off the battlefield either.
	for _, zone := range e.abilityZones() {
		if zone.z == nil {
			continue
		}
		for i := range zone.z.Cards {
			src := &zone.z.Cards[i]
			// CR 108.4: off the battlefield the card's OWNER is the
			// "you" of its printed text, and that is the check
			// ActivateManaAbility makes. Free for the per-seat piles
			// and the whole of the answer for exile, which is one
			// shared pile holding everybody's cards.
			if src.Owner != e.seat {
				continue
			}
			e.manaMovesForSource(src, zone.kind, restricted)
		}
	}
}

// manaMovesForSource enumerates one card's mana abilities, from the
// zone that card is actually in. Split out of manaMoves (#1228) so
// the battlefield loop and the other-zone loop share every line of
// the cost solve — the sacrifice, counter and discard payments, the
// affordability of a mana component, the CR 903.4f narrowing, the
// exhaust gate and the per-source budget — exactly as
// abilityMovesForSource is shared by the CR 602 pair.
//
// The ability list comes from game.ManaAbilitiesForCard, which is the
// one accessor every consumer reads through, so the intrinsic
// basic-land half (CR 305.6) is present here as it is on the
// activation path — and filtered out again by the zone predicate for
// a land sitting in a hand, which is exactly what should happen.
func (e *enumerator) manaMovesForSource(source *game.Card, zone game.ZoneKind, restricted bool) {
	g := e.g
	abilities := game.ManaAbilitiesForCard(*source)
	for idx, ab := range abilities {
		// CR 113.6 (#1228): the ability has to function from the zone
		// the card is in. Same predicate the engine gates on, so a
		// Forest sitting in a hand is never offered as a mana source
		// and a Spirit Guide's hand ability is never offered from the
		// battlefield either.
		if !game.ManaAbilityFunctionsFromZone(ab, zone) {
			continue
		}
		// #1183: "Activate each exhaust ability only once", and
		// this object has. First, exactly as ActivateManaAbility
		// checks it, and through the same one reader — #544's
		// rule is that a bot is never offered a move the engine
		// refuses.
		if g.ManaAbilityExhausted(e.seat, source.InstanceID, ab) {
			continue
		}
		// #1210, CR 602.5a: the board-wide "can't be activated"
		// gate. Cursed Totem does NOT exempt mana abilities, so a
		// Birds of Paradise under one is not a move — the same
		// function ActivateManaAbility calls, with Mana: true, so
		// the restriction decides and not this call site (#544).
		if restricted && g.ActivationGateLocked(e.seat, *source, zone,
			game.ActivationAbility{Label: ab.Label, Mana: true}) != nil {
			continue
		}
		// #352: the activation gate first, exactly as
		// ActivateManaAbility checks it — Temple of the False
		// God is not a move with four lands out.
		if ab.Condition != nil && !ab.Condition(g, e.seat, source.InstanceID) {
			continue
		}
		// CR 903.4f (#844): "any color in your commander's color
		// identity" adds nothing for a seat with no commander, or
		// a colourless one. Tapping Command Tower for no mana is
		// legal and pointless; it is not a move worth offering,
		// and a bot that took it would just lose a land.
		if game.ManaAbilityAddsNoMana(g, e.seat, source.InstanceID, ab) {
			continue
		}
		if ab.TapCost {
			if source.Tapped {
				continue
			}
			if source.IsCreature() && game.HasSummoningSickness(source) {
				continue
			}
		}
		// CR 119.4 and CR 119.8, the same predicate one path
		// over (#544, #1200).
		if !g.CanPayLifeLocked(e.p, ab.LifeCost) {
			continue
		}
		// A mana component in the cost has to be already floating
		// — the activation path deliberately does not auto-tap
		// into a mana ability, so a Signet with an empty pool is
		// not a legal move.
		//
		// #1191: the PRICED cost, not the printed one — the same
		// function ActivateManaAbility pays through, so Boom
		// Scholar's "{2} less to activate" reaching Loot, the
		// Pathfinder's mana half is visible to the policy as an
		// activation it can now afford, mirroring the CR 602 arm
		// above (#544, sign reversed: pricing at the printed cost
		// would silently hide a legal move).
		if ab.ManaCost != "" {
			cost, err := g.ManaAbilityManaCostForEffect(e.seat, *source, ab)
			if err != nil || !e.p.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(*source)) {
				continue
			}
		}
		sacrificeSets := [][]uuid.UUID{nil}
		if ab.SacrificeOther != nil {
			pool := e.sacrificePool(source.InstanceID, ab.SacrificeCost, ab.SacrificeOther)
			// #1213: the same two-way split the CR 602 path makes.
			// A mana ability announces no X, so only the OPEN form
			// can reach here (effects.Register refuses CountFromX
			// on a mana ability).
			if game.SacrificeCostVariable(ab.SacrificeOther) {
				sacrificeSets = e.variableSacrificePayments(pool, ab.SacrificeOther, source.InstanceID)
			} else {
				sacrificeSets = e.sacrificePayments(pool, ab.SacrificeOther, source.InstanceID)
			}
			if len(sacrificeSets) == 0 {
				continue
			}
		}
		// #1213: a "Discard N cards" cost on a mana ability
		// (Skirge Familiar). Solved exactly as the activated
		// path solves its own — ONE payment, the cheapest set in
		// hand order, rather than one move per subset of the
		// hand. Nothing payable means no move at all (#544), and
		// the walk is the engine's own so an offered payment is
		// one ActivateManaAbility accepts.
		var manaDiscardIDs []uuid.UUID
		if dc := ab.DiscardCards; dc != nil && dc.N > 0 {
			opts := g.DiscardCostOptionsForEffect(e.seat, source.InstanceID, dc)
			if len(opts) < dc.N {
				continue
			}
			manaDiscardIDs = opts[:dc.N]
		}
		// #789: the counter components, enumerated by the SAME
		// functions the activated path uses, because it is the
		// same component. A Vivid land with no charge counters is
		// not a five-colour move, and Ramos with four +1/+1
		// counters is not a move at all.
		counterChoices := []counterChoice{{}}
		if rc := ab.RemoveCounters; rc != nil {
			counterChoices = counterPaymentChoices(g.CounterCostOptionsForEffect(e.seat, source.InstanceID, rc), rc)
			if len(counterChoices) == 0 {
				continue
			}
		}
		if ac := ab.AddCounter; ac != nil && !g.CanPlaceCounterForEffect(e.seat, source.InstanceID, ac) {
			continue
		}
		for _, sacs := range sacrificeSets {
			for _, cc := range counterChoices {
				label := source.Name + ": " + ab.Label
				if ab.Label == "" {
					label = source.Name + ": add " + ab.Produced
				}
				label += sacrificeLabel(g, sacs)
				label += cc.label(g)
				// Mana Confluence's "Pay 1 life" is the same
				// invisible cost an activated ability's is (#74),
				// and so is a charge counter: the params name the
				// permanent but never the price.
				cost := moveCost(ab.LifeCost, 0)
				for _, price := range cc.prices() {
					cost = withCounterPrice(cost, price)
				}
				e.add(Move{
					Type:   TypeActivateManaAbility,
					Player: e.seat,
					Kind:   KindMana,
					Label:  label,
					Source: source.InstanceID,
					Cost:   cost,
					Params: mustJSON(manaParams{
						CardID:           source.InstanceID.String(),
						AbilityIndex:     idx,
						SacrificeIDs:     idStrings(sacs),
						CounterSourceIDs: cc.wireIDs(),
						CounterCounts:    cc.wireCounts(),
						CounterKind:      cc.wireKind(),
						CounterKinds:     cc.wireKinds(),
						DiscardIDs:       idStrings(manaDiscardIDs),
					}),
				})
			}
		}
	}
}
