package legal

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castParams is the cast_spell wire payload this package emits. Every
// cast is sent strict + auto_tap: the bot pays for what it casts, and
// affordability is decided here so the engine never has to reject.
type castParams struct {
	InstanceID string `json:"instance_id"`
	FromZone   string `json:"from_zone,omitempty"`
	// AlternativeCost is the key the cast claims (CR 118.9) — the
	// offer a granted permission synthesises for a non-hand cast
	// (ADR 0066), since a zone a permission prices cannot be cast
	// from without paying that price.
	AlternativeCost string       `json:"alternative_cost,omitempty"`
	Targets         []targetWire `json:"targets,omitempty"`
	Modes           []int        `json:"modes,omitempty"`
	XValue          int          `json:"x_value,omitempty"`
	DiscardIDs      []string     `json:"discard_ids,omitempty"`
	SacrificeIDs    []string     `json:"sacrifice_ids,omitempty"`
	Strict          bool         `json:"strict,omitempty"`
	AutoTap         bool         `json:"auto_tap,omitempty"`
	// Face is the printed face being cast or played (ADR 0034).
	// Omitted — the front — for every single-faced card.
	Face int `json:"face,omitempty"`
}

// castMoves enumerates land drops and spell casts from the seat's
// hand and command zone, plus every card a granted permission opens
// in a graveyard, in exile, or on top of a library (ADR 0066).
//
// The granted half goes through game.Game.CastPermissionForLocked —
// the same function CastSpell validates with — so a bot can never be
// offered a cast the engine will refuse, and never be offered one at
// the wrong price. That is the enumerator half of #673 for GRANTED
// permissions; the printed ones (a card whose own text declares
// flashback, escape or warp) are still not enumerated.
func (e *enumerator) castMoves() {
	g, p := e.g, e.p
	if g.SplitSecondActive {
		return
	}
	speed := sorcerySpeedOpen(g, e.seat)
	// #500: the allowance is the player's, not a literal one — a
	// controlled Exploration or a one-turn grant raises it. Same
	// helper the engine's own refusal reads, so the enumerator can
	// never offer a land play CastSpell will reject.
	landOwed := g.LandDropsRemainingLocked(e.seat) > 0

	e.grantedCastMoves(speed, landOwed)

	for _, zone := range []struct {
		z    *game.Zone
		from string
	}{{p.Hand, "hand"}, {p.Command, "command"}} {
		if zone.z == nil {
			continue
		}
		for _, c := range zone.z.Cards {
			// ADR 0034: a modal DFC is two playable objects sharing
			// one instance, so enumerate each face as its own move
			// and let the bot pick between them. CastableFaces
			// returns [0] for everything else, so this loop runs once
			// for every single-faced card and the enumeration is
			// unchanged for them.
			//
			// The face is materialised onto a COPY, exactly as
			// CastSpell does, so all the type, cost and catalog reads
			// below see the chosen half without any of them learning
			// about faces.
			for _, face := range c.CastableFaces() {
				card := c
				card.SetFace(face)
				if card.IsLand() {
					// CR 305: main phase, empty stack, your turn, and
					// the per-turn land-play allowance. The engine
					// enforces all four since #500; the check stays
					// here so a bot is never OFFERED a move that
					// would be refused.
					if zone.from == "hand" && speed && landOwed {
						e.add(Move{
							Type:   TypeCastSpell,
							Player: e.seat,
							Kind:   KindLand,
							Label:  "Play " + card.Name,
							Source: card.InstanceID,
							Params: mustJSON(castParams{
								InstanceID: card.InstanceID.String(),
								FromZone:   "hand",
								Face:       face,
							}),
						})
					}
					continue
				}
				e.castMovesForCard(card, zone.from, speed, nil)
			}
		}
	}
}

// grantedCastMoves enumerates the casts and land plays a granted
// permission opens out of a graveyard, exile or the top of a library.
//
// One shape per zone and no expansion beyond what castMovesForCard
// already does: the permission decides whether the card may be cast
// at all, and the price it names becomes the alternative cost the
// move claims.
//
// NOT enumerated, deliberately: a permission whose price has a CARD
// component — escape's "exile three other cards from your graveyard"
// (Underworld Breach, The Grim Captain's Locker). Those need a
// combination search over the graveyard and an alt_cost_ids payload,
// which is the same machinery the printed escape costs want and
// belongs with the rest of #673 rather than bolted on here. A bot
// simply does not take those lines yet; it is never offered one it
// cannot pay for.
func (e *enumerator) grantedCastMoves(speed, landOwed bool) {
	g, p := e.g, e.p
	// The fast negative, and the same one the view takes: in almost
	// every game nothing grants anything, and the three walks below
	// are pure cost. An enumeration runs on every bot decision.
	if !g.AnyCastPermissionsForEffect() {
		return
	}
	zones := []struct {
		z    *game.Zone
		from string
	}{
		{p.Graveyard, "graveyard"},
		{g.Exile, "exile"},
		{p.Library, "library"},
	}
	for _, zone := range zones {
		if zone.z == nil {
			continue
		}
		cards := zone.z.Cards
		if zone.from == "library" {
			// CR 401.5: only the top card is ever open, and the top is
			// the LAST element. Checking one card rather than walking
			// the library also keeps this loop from touching hidden
			// information it has no business reading.
			if len(cards) == 0 {
				continue
			}
			cards = cards[len(cards)-1:]
		}
		for _, c := range cards {
			perm := g.CastPermissionForLocked(e.seat, c, zone.z.Kind)
			if !perm.Active(e.seat, g.Turn.Number) {
				continue
			}
			card := c
			if face, ok := perm.GrantsFace(e.seat, g.Turn.Number); ok {
				card.SetFace(face)
			}
			offer := perm.AlternativeCostFor(card)
			if offer != nil && offer.PaysCards() {
				continue
			}
			if card.IsLand() {
				// CR 305.1: playing a land is not casting, so a
				// cast-only permission strands it, and CR 305.2's
				// allowance still has to be there.
				if perm.CastOnly || !speed || !landOwed {
					continue
				}
				e.add(Move{
					Type:   TypeCastSpell,
					Player: e.seat,
					Kind:   KindLand,
					Label:  "Play " + card.Name + " from " + zone.from,
					Source: card.InstanceID,
					Params: mustJSON(castParams{
						InstanceID: card.InstanceID.String(),
						FromZone:   zone.from,
						Face:       card.ActiveFace,
					}),
				})
				continue
			}
			if offer != nil && offer.Life > 0 && p.Life <= offer.Life {
				// CR 119.4 / CR 118.4: life is a cost, so a player who
				// cannot pay it cannot claim the offer. Strictly below,
				// not at — paying down to exactly zero is legal but
				// loses the game to the next state-based check, and a
				// bot offered that line would take it.
				continue
			}
			e.castMovesForCard(card, zone.from, speed, perm)
		}
	}
}

// castMovesForCard expands one non-land card into concrete casts:
// every legal (modes × targets × additional-cost payment) combination
// the seat can afford, capped at MaxExpansionPerSource.
func (e *enumerator) castMovesForCard(card game.Card, from string, speed bool, perm *game.CastPermission) {
	g, p := e.g, e.p

	// Timing (CR 307.1 / 702.8): instants and flash any time the seat
	// holds priority; everything else needs the sorcery-speed window.
	// ADR 0066: a granted permission may say otherwise, and the engine
	// reads the same field in the same order, so the two cannot drift.
	requiresSorcerySpeed := !card.IsInstant() && !game.HasKeyword(&card, "flash")
	if perm != nil {
		switch perm.Timing {
		case game.TimingFlash:
			requiresSorcerySpeed = false
		case game.TimingSorcery:
			requiresSorcerySpeed = true
		}
	}
	if requiresSorcerySpeed && !speed {
		return
	}

	// A card the catalog marks as targeted the S13.1 way (free-form
	// target_mode, no structured spec) cannot be enumerated: the
	// engine demands a target but nothing says which are legal.
	if game.TargetModeFor(game.CatalogKey(card)) != "" && game.TargetSpecFor(game.CatalogKey(card)) == nil {
		return
	}

	// Cost. A cost the parser can't read is not enumerable: since
	// #289 the engine rejects such a cast with ErrUnparseableCost
	// (split and adventure cards import a joined "{1}{R} // {1}{U}"),
	// so offering the move would hand the client an action that is
	// guaranteed to fail. Enumerating it free — what this used to
	// do, mirroring the engine's old silent downgrade — is worse:
	// it advertises a free spell that isn't one.
	// ADR 0066: the permission's price, when it has one, IS the cost
	// this cast pays (CR 118.9). Read through the same CastCostFor the
	// engine prices with.
	offer := perm.AlternativeCostFor(card)
	cost, err := game.ParseCost(game.CastCostFor(card, offer, perm).Paid)
	if err != nil {
		return
	}
	// CR 118.6: a spell with no mana cost (Ancestral Vision, Living
	// End) can't be cast by paying it, and ParseCost reads that empty
	// string as a free {0}. Every move this function builds pays the
	// printed cost, so none of them is legal; the engine refuses the
	// cast with ErrNoManaCost.
	if game.HasNoManaCost(card) && offer == nil && (perm == nil || perm.Cost == "") {
		return
	}
	if from == "command" {
		cost.Generic += p.CommanderCasts[card.InstanceID] * 2
	}
	// S28: the board's cost modifiers (CR 601.2f). Same reasoning as
	// the parse gate above — a move enumerated at the printed price
	// while a Sphere of Resistance sits on the table is a move the
	// engine will reject for insufficient mana, and a bot that keeps
	// picking rejected moves stalls. A modifier the engine refuses to
	// price (ErrCostModifier) makes the cast unenumerable for the
	// same reason an unparseable cost does.
	//
	// Priced at X=0 even though affordableX is about to search for a
	// bigger X. The only modifier kind X can change the answer for is
	// a CostFloor (Trinisphere), and X=0 is the branch where the
	// floor applies — so the search starts from the most expensive
	// reading and can only narrow the X it offers. Conservative in
	// the direction that never advertises an unaffordable move.
	fromZone := game.ZoneHand
	switch from {
	case "command":
		fromZone = game.ZoneCommand
	case "graveyard":
		fromZone = game.ZoneGraveyard
	case "exile":
		fromZone = game.ZoneExile
	case "library":
		fromZone = game.ZoneLibrary
	}
	// ADR 0048 addendum §14: when something on the board or the card
	// itself prices by target (Fireball's surcharge, Price of Fame's
	// discount), one price up front is not the price — and a
	// nil-targets price cannot even be used as a gate, because a
	// target-reading REDUCTION makes the real cast cheaper than it.
	// So the up-front price and its early return run only when
	// nothing reads targets, which is every board without such a
	// card; otherwise each (modes, targets) set below is priced on its
	// own and carries its own X.
	spend := game.ManaSpendForCast(card)
	perTarget := e.g.CastPriceReadsTargetsForEffect(card)
	// Additional costs (CR 601.2f). Read before the X search because
	// one of them can PRICE X: Toxic Deluge's "pay X life" is the
	// whole of its X, and the mana cost it prints has no {X} slot at
	// all (#957). The payments themselves are expanded below.
	addCost := game.AdditionalCostFor(game.CatalogKey(card))
	// #810: the one X rule. A spell whose whole effect is X (Fireball,
	// Stroke of Genius) is not offered at X=0, where it would resolve
	// for nothing; a spell with a fixed rider still is. #957 is the
	// same rule reaching the non-mana half of the price: the floor is
	// the same one, and xCeilingFromCost is what a "pay X life" cost
	// can pay for (CR 119.4, less one so the seat survives its own
	// sweep). Both live in x.go.
	xFloor := enumeratedXFloor(game.CatalogKey(card), 0)
	xLifeCeiling := xCeilingFromCost(addCost, p.Life)
	x := 0
	if !perTarget {
		priced, err := e.g.ApplyCostModifiersForEffect(cost, game.CostQuery{
			Card:       card,
			Controller: e.seat,
			FromZone:   fromZone,
		})
		if err != nil {
			return
		}
		var ok bool
		x, ok = e.announcedX(priced, spend, xFloor, xLifeCeiling)
		if !ok {
			return
		}
	}

	// Modes → each choice of modes yields its own clause list, and
	// each clause its own picks (#764). Options with no legal target
	// are dropped before any combination is built, so the budget is
	// never spent on selections the engine would refuse (ADR 0065
	// §6).
	modeSpec := game.ModeSpecFor(game.CatalogKey(card))
	modeSets := [][]int{nil}
	if modeSpec != nil {
		modeSets = e.legalModeSets(modeSpec)
		if len(modeSets) == 0 {
			return
		}
	}

	// The additional cost's PAYMENTS (CR 601.2f), read above. Discards
	// choose from the rest of the hand; a sacrifice chooses from the
	// seat's own permanents matching the clause. A pay-X-life needs no
	// payment set of its own — the announced X is the payment, and it
	// was priced with the rest of X above.
	discardSets := [][]uuid.UUID{nil}
	sacrificeSets := [][]uuid.UUID{nil}
	if addCost != nil && !addCost.Empty() {
		if addCost.DiscardCards > 0 {
			var pool []uuid.UUID
			for _, h := range p.Hand.Cards {
				if h.InstanceID != card.InstanceID {
					pool = append(pool, h.InstanceID)
				}
			}
			discardSets = combinations(pool, addCost.DiscardCards, addCost.DiscardCards, e.opts.MaxExpansionPerSource)
			if len(discardSets) == 0 {
				return
			}
		}
		if addCost.Sacrifice != nil {
			// Cost, not target — see SpecCandidatesForEffect.
			lt := g.SpecCandidatesForEffect(e.seat, addCost.Sacrifice)
			var pool []uuid.UUID
			for _, id := range lt.Cards {
				if c := findBattlefield(g, id); c != nil && c.Controller == e.seat {
					pool = append(pool, id)
				}
			}
			// #747: N from the clause, one payment per cast for N ≥ 2,
			// nothing offered when the caster controls fewer than N.
			sacrificeSets = e.sacrificePayments(pool, addCost.Sacrifice, uuid.Nil)
			if len(sacrificeSets) == 0 {
				return
			}
		}
	}

	budget := e.opts.MaxExpansionPerSource
	cardSpec := game.TargetSpecFor(game.CatalogKey(card))
	for _, modes := range modeSets {
		// The budget is spent MODES-outermost: every mode selection
		// gets at least one target set before any gets a second, so a
		// bot is never offered only the first bullet of a charm
		// (ADR 0065 §6).
		steps := game.AnnouncedClauses(cardSpec, modeSpec, modes)
		// #619, CR 601.2c. Crackle with Power's target count IS X
		// ("deals five times X damage to each of up to X targets"),
		// and the announce path refuses any cast where the two
		// disagree. So for such a step the enumerator does not pick an
		// X and a target list independently — it picks the targets,
		// and the X it announces is how many it picked.
		xSteps := stepsCountedByX(steps)
		if len(xSteps) > 0 {
			if cost.XSlots == 0 {
				// The step's X is announced by a cost this package
				// cannot price — Waterbender's Restoration's
				// waterbend {X}, paid by tapping artifacts and
				// creatures. The only announcement the enumerator
				// could make for it is X=0, which buys no targets and
				// a spell that does nothing, and any larger one would
				// be a move the engine refuses for an unpaid cost. So
				// the cast is not enumerable, the same answer an
				// unparseable cost gets.
				continue
			}
			// The arities to generate are bounded by the largest X
			// the seat could announce. Under a per-target price that
			// is not known until each set is priced, so open the step
			// to MaxX and let the affordability check below drop what
			// cannot be paid; the budget caps the expansion either
			// way.
			bound := x
			if perTarget {
				bound = e.opts.MaxX
			}
			if bound < 1 {
				continue
			}
			steps = openXCountedSteps(steps, bound)
		}
		targetSets := e.legalStepSets(steps, budget)
		if len(targetSets) == 0 {
			continue
		}
		for _, targets := range targetSets {
			setX := x
			if perTarget {
				// §14: priced with this set's targets. An unaffordable
				// set is skipped before any budget is spent on it, so
				// a Fireball the seat can pay for at one target is not
				// crowded out by the three-target sets it cannot.
				priced, err := e.g.ApplyCostModifiersForEffect(cost, game.CostQuery{
					Card:       card,
					Controller: e.seat,
					FromZone:   fromZone,
					Targets:    targets,
				})
				if err != nil {
					continue
				}
				var ok bool
				setX, ok = e.announcedX(priced, spend, xFloor, xLifeCeiling)
				if !ok {
					continue
				}
			}
			if len(xSteps) > 0 {
				// setX is the largest announcement this seat can pay
				// for; a set that filled more X-counted slots than
				// that is a cast it cannot make. The cost is monotonic
				// in X, so the comparison is the whole affordability
				// check.
				k, ok := announcedXCount(steps, xSteps, targets)
				if !ok || k < 1 || k > setX {
					continue
				}
				setX = k
			}
			for _, discards := range discardSets {
				for _, sacs := range sacrificeSets {
					if budget <= 0 {
						return
					}
					budget--
					label := "Cast " + card.Name
					switch from {
					case "command":
						label += " from the command zone"
					case "graveyard", "exile", "library":
						label += " from " + from
					}
					if setX > 0 {
						label += fmt.Sprintf(" for X=%d", setX)
					}
					label += targetLabel(g, targets)
					e.add(Move{
						Type:   TypeCastSpell,
						Player: e.seat,
						Kind:   KindCast,
						Label:  label,
						Source: card.InstanceID,
						Params: mustJSON(castParams{
							InstanceID:      card.InstanceID.String(),
							FromZone:        from,
							AlternativeCost: offerKey(offer),
							Targets:         wireTargets(targets),
							Modes:           modes,
							XValue:          setX,
							DiscardIDs:      idStrings(discards),
							SacrificeIDs:    idStrings(sacs),
							Strict:          true,
							AutoTap:         true,
							// ADR 0034: `card` has already had
							// SetFace applied by the caller, so
							// ActiveFace IS the face this move casts.
							Face: card.ActiveFace,
						}),
					})
				}
			}
		}
	}
}

// affordableXFrom reports whether the seat can pay cost right now —
// from the floating pool, or by the auto-tapper's plan — and, for an
// {X} cost, the largest X it can pay, at or above `floor`, up to MaxX.
// Mirrors the engine's auto_tap + strict path: the pool is consulted
// first, then a plan is sought for the WHOLE cost (the engine does not
// net floating mana against the plan).
//
// The floor is the announcement's lower bound, and it has two sources,
// both settled by enumeratedXFloor (x.go) before the call: a printed
// "X can't be 0" (Helm of Obedience), and #810's rule that a move
// whose whole effect is X is not worth offering at X=0. A seat that
// cannot pay for the floor has no move at all rather than a free one.
//
// The scan still starts at the floor and still breaks on the first
// unaffordable value, because the cost is monotonic in X: every
// extra point of X buys the same XSlots generic symbols. Nothing
// here enumerates a RANGE — exactly one X comes back, so X never
// enters an expansion cross product (see activatedMoves for why
// that matters).
func (e *enumerator) affordableXFrom(cost game.ParsedCost, spend game.ManaSpendContext, floor int) (int, bool) {
	return e.affordableXExcluding(cost, spend, floor, nil)
}

// affordableXExcluding is affordableXFrom with sources the payment
// may not use — the one caller is an activated ability whose cost
// includes {T}, which cannot tap its own source for mana.
func (e *enumerator) affordableXExcluding(
	cost game.ParsedCost,
	spend game.ManaSpendContext,
	floor int,
	excluded map[uuid.UUID]bool,
) (int, bool) {
	if cost.XSlots == 0 {
		// No {X}: the floor is meaningless and X is always zero.
		return 0, e.canPayExcluding(cost, 0, spend, excluded)
	}
	if floor < 0 {
		floor = 0
	}
	best, ok := -1, false
	for x := floor; x <= e.opts.MaxX; x++ {
		if e.canPayExcluding(cost, x, spend, excluded) {
			best, ok = x, true
			continue
		}
		break
	}
	return best, ok
}

// `spend` is the #352 spend context — what the mana would be paid
// for — so restricted mana in the pool counts toward a cast it may
// legally fund and toward no other.
func (e *enumerator) canPay(cost game.ParsedCost, x int, spend game.ManaSpendContext) bool {
	return e.canPayExcluding(cost, x, spend, nil)
}

// canPayExcluding is canPay with sources the auto-tapper may not
// reach for. It has to mirror exactly what the engine excludes, or
// the enumerator's answer and the engine's answer disagree — which
// is the #544 failure mode, one cost component over.
func (e *enumerator) canPayExcluding(
	cost game.ParsedCost,
	x int,
	spend game.ManaSpendContext,
	excluded map[uuid.UUID]bool,
) bool {
	if e.p.ManaPool.CanPayFor(cost, x, spend) {
		return true
	}
	_, ok := e.g.AutoTapForCostForEffectExcluding(e.seat, cost, x, excluded)
	return ok
}

// legalModeSets lists every distinct mode selection of size
// Min..Max, excluding any with more than one targeted option (the
// engine's castTargetSpec rejects those).
func (e *enumerator) legalModeSets(ms *game.ModeSpec) [][]int {
	// ADR 0065 §6, "prefer the modes that have legal targets": an
	// option whose clause cannot be filled is dropped before any
	// combination is built, so the budget never goes on a selection
	// the engine would refuse at announce.
	options := e.g.ChoosableModeOptionsForEffect(e.seat, ms)
	if !game.EnoughChoosableModes(len(options), ms) {
		return nil
	}
	hi := ms.Max
	if hi <= 0 || (!ms.Repeatable && hi > len(options)) {
		hi = len(options)
	}
	budget := e.opts.MaxExpansionPerSource
	var out [][]int
	add := func(sel []int) bool {
		out = append(out, append([]int(nil), sel...))
		return len(out) < budget
	}
	if ms.Min == 0 {
		if !add(nil) {
			return out
		}
	}
	lo := ms.Min
	if lo < 1 {
		lo = 1
	}
	// CR 700.2d: the all-one-option selections first, so a
	// repeatable spec whose only legal option is one mode is not
	// crowded out by mixed multisets.
	if ms.Repeatable {
		for _, opt := range options {
			for n := lo; n <= hi; n++ {
				sel := make([]int, n)
				for i := range sel {
					sel[i] = opt
				}
				if !add(sel) {
					return out
				}
			}
		}
	}
	var rec func(start int, cur []int) bool
	rec = func(start int, cur []int) bool {
		if len(cur) >= lo && len(cur) <= hi {
			if !(ms.Repeatable && len(cur) == 1) && !add(cur) {
				return false
			}
		}
		if len(cur) == hi {
			return true
		}
		for i := start; i < len(options); i++ {
			if !rec(i+1, append(cur, options[i])) {
				return false
			}
		}
		return true
	}
	rec(0, nil)
	return out
}

// legalStepSets is the cartesian product of each clause's legal
// picks, in step order, capped at `budget` (#764, ADR 0065 §6). An
// announcement with no steps yields the single empty set, which is
// how an untargeted cast stays one move.
func (e *enumerator) legalStepSets(steps []game.AnnouncedClause, budget int) [][]game.TargetRef {
	out := [][]game.TargetRef{nil}
	for i := range steps {
		clause := steps[i].Clause
		picks := e.legalTargetSets(&clause, budget)
		if len(picks) == 0 {
			return nil
		}
		next := make([][]game.TargetRef, 0, budget)
		for _, prefix := range out {
			for _, pick := range picks {
				if len(next) >= budget {
					break
				}
				combined := append([]game.TargetRef(nil), prefix...)
				skip := false
				for _, p := range pick {
					p.Mode, p.Slot = steps[i].Mode, steps[i].Slot
					// CR 601.2c: a Distinct clause may not repeat an
					// object an earlier clause took.
					if clause.Distinct {
						for _, seen := range prefix {
							if seen.ID == p.ID {
								skip = true
							}
						}
					}
					combined = append(combined, p)
				}
				if skip {
					continue
				}
				next = append(next, combined)
			}
		}
		if len(next) == 0 {
			return nil
		}
		out = next
	}
	return out
}

// stepsCountedByX lists the indexes of the announcement's steps whose
// target count is the announced X rather than a printed constant
// (Crackle with Power, Doppelgang, Heliod's Intervention's first
// mode). Empty — the overwhelming majority — means nothing here ties
// X to anything.
func stepsCountedByX(steps []game.AnnouncedClause) []int {
	var out []int
	for i := range steps {
		if steps[i].Clause.CountFromX {
			out = append(out, i)
		}
	}
	return out
}

// openXCountedSteps returns a copy of `steps` with every X-counted
// clause opened to 1..bound, so legalStepSets expands one target set
// per arity and the caller can read the arity back off the set.
//
// The engine's own resolveStepCountsFromX pins the same clauses to a
// single announced X; this is that operation with the X still unknown,
// which is the whole difference between validating an announcement and
// building one.
//
// The floor is 1, not 0. X=0 on such a step is a spell cast with no
// targets, which for every card in the catalog that has one means a
// spell that does nothing — #810's rule, and the reason this needs no
// zero case of its own. A CountFromX card with a fixed rider would
// want one; none exists, and x.go is where that would be decided.
func openXCountedSteps(steps []game.AnnouncedClause, bound int) []game.AnnouncedClause {
	out := append([]game.AnnouncedClause(nil), steps...)
	for _, i := range stepsCountedByX(out) {
		out[i].Clause.Min, out[i].Clause.Max = 1, bound
		out[i].Clause.CountFromX = false
	}
	return out
}

// announcedXCount reads the X a target set is announcing: how many
// refs answer the X-counted steps. False when two such steps disagree,
// which is a set no announcement could cover — one X, one count
// (CR 601.2c).
//
// Matched on (Mode, Slot), the coordinates legalStepSets stamps onto
// every ref, which is exactly what the engine's stepTargetCount reads.
func announcedXCount(steps []game.AnnouncedClause, xSteps []int, targets []game.TargetRef) (int, bool) {
	k := -1
	for _, i := range xSteps {
		n := 0
		for _, t := range targets {
			if t.Kind == game.TargetSelf || t.Kind == game.TargetNone {
				continue
			}
			if t.Mode == steps[i].Mode && t.Slot == steps[i].Slot {
				n++
			}
		}
		if k >= 0 && n != k {
			return 0, false
		}
		k = n
	}
	return k, k >= 0
}

// legalTargetSets expands a target clause into concrete target
// lists: the empty list when Min is 0, then every k-subset of the
// legal candidates for k in max(Min,1)..Max, players before cards,
// stopping at budget entries.
func (e *enumerator) legalTargetSets(spec *game.TargetSpec, budget int) [][]game.TargetRef {
	lt := e.g.LegalTargetsForEffect(e.seat, spec)
	cands := make([]game.TargetRef, 0, len(lt.Players)+len(lt.Cards))
	for _, id := range lt.Players {
		cands = append(cands, game.TargetRef{Kind: game.TargetPlayer, ID: id})
	}
	for _, id := range lt.Cards {
		cands = append(cands, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	lo, hi := spec.Min, spec.Max
	if hi <= 0 || hi > len(cands) {
		hi = len(cands)
	}
	var out [][]game.TargetRef
	if lo == 0 {
		out = append(out, nil)
	}
	if lo < 1 {
		lo = 1
	}
	if lo > len(cands) {
		return out
	}
	for k := lo; k <= hi && len(out) < budget; k++ {
		var rec func(start int, cur []game.TargetRef)
		rec = func(start int, cur []game.TargetRef) {
			if len(out) >= budget {
				return
			}
			if len(cur) == k {
				out = append(out, append([]game.TargetRef(nil), cur...))
				return
			}
			for i := start; i < len(cands); i++ {
				rec(i+1, append(cur, cands[i]))
			}
		}
		rec(0, nil)
	}
	return out
}

// combinations returns every k-subset of pool for k in min..max, in
// pool order, up to limit entries. Returns nil when the pool is too
// small for min.
func combinations(pool []uuid.UUID, lo, hi, limit int) [][]uuid.UUID {
	if lo > len(pool) {
		return nil
	}
	if hi > len(pool) {
		hi = len(pool)
	}
	var out [][]uuid.UUID
	for k := lo; k <= hi && len(out) < limit; k++ {
		var rec func(start int, cur []uuid.UUID)
		rec = func(start int, cur []uuid.UUID) {
			if len(out) >= limit {
				return
			}
			if len(cur) == k {
				out = append(out, append([]uuid.UUID(nil), cur...))
				return
			}
			for i := start; i < len(pool); i++ {
				rec(i+1, append(cur, pool[i]))
			}
		}
		rec(0, nil)
	}
	return out
}

// offerKey is the wire key of a resolved alternative cost, or "" when
// the cast pays the printed price. Nil-safe so the emit site needs no
// guard.
func offerKey(offer *game.AlternativeCost) string {
	if offer == nil {
		return ""
	}
	return offer.Key
}
