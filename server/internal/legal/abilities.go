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
	CounterKind      string   `json:"counter_kind,omitempty"`
	XValue           int      `json:"x_value,omitempty"`
	Strict           bool     `json:"strict,omitempty"`
	AutoTap          bool     `json:"auto_tap,omitempty"`
}

// activatedMoves enumerates catalog activated abilities on the
// seat's permanents. Requires priority (checked by the caller).
// Mirrors game.ActivateCatalogAbility's validation: split second,
// sorcery-speed flag, activation condition (#743), tap cost
// (untapped + not summoning sick for creatures), sacrifice costs
// payable, life cost payable, mana cost affordable, targets legal.
func (e *enumerator) activatedMoves() {
	g, p := e.g, e.p
	if g.SplitSecondActive {
		return
	}
	speed := sorcerySpeedOpen(g, e.seat)
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
		abilities := game.ActivatedAbilitiesForCard(*source)
		for idx, ab := range abilities {
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
			// CR 606: a loyalty ability needs a planeswalker, one
			// activation per turn, and enough counters to pay a −N.
			// Mirrors ActivateCatalogAbility so a policy never
			// proposes a move the engine will bounce.
			if ab.Cost.Loyalty != nil {
				if !source.IsPlaneswalker() || g.LoyaltyActivatedThisTurn[source.InstanceID] {
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
			if ab.Cost.Life > 0 && p.Life < ab.Cost.Life {
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
			if ab.Cost.Mana != "" {
				cost, err := game.ParseCost(ab.Cost.Mana)
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
				x, ok := e.affordableXExcluding(cost, game.ManaSpendForAbility(*source), floor, excluded)
				if !ok {
					continue
				}
				xValue = x
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
				sacrificeSets = e.sacrificePayments(pool, ab.Cost.SacrificeOther, source.InstanceID)
				if len(sacrificeSets) == 0 {
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
			budget := e.opts.MaxExpansionPerSource
			// #764: a modal activated ability announces its modes with
			// its targets (CR 602.2b), so the enumerator expands the
			// same product a modal cast does.
			modeSets := [][]int{nil}
			if ab.Modes != nil {
				modeSets = e.legalModeSets(ab.Modes)
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
				sets := e.legalStepSets(steps, budget)
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
					for _, cc := range counterChoices {
						if budget <= 0 {
							break
						}
						budget--
						label := source.Name + ": " + ab.Label
						if xValue > 0 {
							label += fmt.Sprintf(" for X=%d", xValue)
						}
						label += cc.label(g)
						label += targetLabel(g, targets)
						cost := moveCost(ab.Cost.Life, loyalty)
						if cc.n > 0 {
							cost = withCounterPrice(cost, CounterPrice{CardID: cc.cardID, Counter: cc.kind, N: cc.n})
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
								CounterKind:      cc.wireKind,
								XValue:           xValue,
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

// counterChoice is one way to pay a RemoveCounters cost: the zero
// value stands for "the ability has no counter component".
type counterChoice struct {
	cardID uuid.UUID
	kind   string
	n      int
	count  int
	// fromOther is true for the "from a permanent you control" form,
	// which names the permanent on the wire; the self form does not.
	fromOther bool
	// wireKind is the kind sent as counter_kind: set only for "a
	// counter" of any kind, where the engine needs to be told.
	wireKind string
}

// counterPaymentChoices flattens the engine's (permanent, kinds)
// options into one choice per (permanent, kind), most counters first
// across the whole set.
func counterPaymentChoices(opts []game.CounterCostOption, rc *game.CounterRemovalCost) []counterChoice {
	var out []counterChoice
	for _, o := range opts {
		for _, k := range o.Kinds {
			cc := counterChoice{cardID: o.CardID, kind: k.Kind, n: rc.N, count: k.Count, fromOther: rc.From != nil}
			if rc.Counter == "" {
				cc.wireKind = k.Kind
			}
			out = append(out, cc)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].count > out[j].count })
	return out
}

func (cc counterChoice) wireIDs() []string {
	if !cc.fromOther {
		return nil
	}
	return []string{cc.cardID.String()}
}

// label names the payment in the move label when it is a choice the
// reader could not infer from the ability text: which permanent, and
// for "a counter", which kind.
func (cc counterChoice) label(g *game.Game) string {
	if cc.n == 0 || (!cc.fromOther && cc.wireKind == "") {
		return ""
	}
	what := cc.kind + " counter"
	if cc.n > 1 {
		what = fmt.Sprintf("%d %s counters", cc.n, cc.kind)
	}
	return " (removing " + what + " from " + cardName(g, cc.cardID) + ")"
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
}

// manaMoves enumerates mana abilities on the seat's permanents.
// Casts already auto-tap, so a policy rarely needs these for plain
// "{T}: Add" sources; they matter for sacrifice sources (Treasure,
// Lotus Petal, Ashnod's Altar) the auto-tapper never touches.
// Requires priority (checked by the caller). Mana abilities ignore
// split second (CR 702.61b).
func (e *enumerator) manaMoves() {
	g := e.g
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
		abilities := game.ManaAbilitiesForCard(*source)
		for idx, ab := range abilities {
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
			if ab.LifeCost > 0 && e.p.Life < ab.LifeCost {
				continue
			}
			// A mana component in the cost has to be already floating
			// — the activation path deliberately does not auto-tap
			// into a mana ability, so a Signet with an empty pool is
			// not a legal move.
			if ab.ManaCost != "" {
				cost, err := game.ParseCost(ab.ManaCost)
				if err != nil || !e.p.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(*source)) {
					continue
				}
			}
			sacrificeSets := [][]uuid.UUID{nil}
			if ab.SacrificeOther != nil {
				pool := e.sacrificePool(source.InstanceID, ab.SacrificeCost, ab.SacrificeOther)
				sacrificeSets = e.sacrificePayments(pool, ab.SacrificeOther, source.InstanceID)
				if len(sacrificeSets) == 0 {
					continue
				}
			}
			for _, sacs := range sacrificeSets {
				label := source.Name + ": " + ab.Label
				if ab.Label == "" {
					label = source.Name + ": add " + ab.Produced
				}
				if len(sacs) > 0 {
					names := make([]string, len(sacs))
					for i, id := range sacs {
						names[i] = cardName(g, id)
					}
					label += " (sacrificing " + strings.Join(names, ", ") + ")"
				}
				e.add(Move{
					Type:   TypeActivateManaAbility,
					Player: e.seat,
					Kind:   KindMana,
					Label:  label,
					Source: source.InstanceID,
					// Mana Confluence's "Pay 1 life" is the same
					// invisible cost an activated ability's is (#74).
					Cost: moveCost(ab.LifeCost, 0),
					Params: mustJSON(manaParams{
						CardID:       source.InstanceID.String(),
						AbilityIndex: idx,
						SacrificeIDs: idStrings(sacs),
					}),
				})
			}
		}
	}
}
