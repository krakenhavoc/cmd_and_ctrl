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
	InstanceID   string       `json:"instance_id"`
	FromZone     string       `json:"from_zone,omitempty"`
	Targets      []targetWire `json:"targets,omitempty"`
	Modes        []int        `json:"modes,omitempty"`
	XValue       int          `json:"x_value,omitempty"`
	DiscardIDs   []string     `json:"discard_ids,omitempty"`
	SacrificeIDs []string     `json:"sacrifice_ids,omitempty"`
	Strict       bool         `json:"strict,omitempty"`
	AutoTap      bool         `json:"auto_tap,omitempty"`
	// Face is the printed face being cast or played (ADR 0034).
	// Omitted — the front — for every single-faced card.
	Face int `json:"face,omitempty"`
}

// castMoves enumerates land drops and spell casts from the seat's
// hand and command zone. Requires priority (checked by the caller).
//
// Sources: hand and command only. Exile (impulse permissions) is a
// follow-up — the permission check is per-card and the catalog's
// impulse cards are not in the curated bot decks yet.
func (e *enumerator) castMoves() {
	g, p := e.g, e.p
	if g.SplitSecondActive {
		return
	}
	speed := sorcerySpeedOpen(g, e.seat)
	landOwed := g.LandsPlayedThisTurnFor(e.seat) < 1

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
					// one per turn. The engine enforces the first
					// three and not the fourth; we enforce all four.
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
				e.castMovesForCard(card, zone.from, speed)
			}
		}
	}
}

// castMovesForCard expands one non-land card into concrete casts:
// every legal (modes × targets × additional-cost payment) combination
// the seat can afford, capped at MaxExpansionPerSource.
func (e *enumerator) castMovesForCard(card game.Card, from string, speed bool) {
	g, p := e.g, e.p

	// Timing (CR 307.1 / 702.8): instants and flash any time the seat
	// holds priority; everything else needs the sorcery-speed window.
	requiresSorcerySpeed := !card.IsInstant() && !game.HasKeyword(&card, "flash")
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
	cost, err := game.ParseCost(card.ManaCost)
	if err != nil {
		return
	}
	// CR 118.6: a spell with no mana cost (Ancestral Vision, Living
	// End) can't be cast by paying it, and ParseCost reads that empty
	// string as a free {0}. Every move this function builds pays the
	// printed cost, so none of them is legal; the engine refuses the
	// cast with ErrNoManaCost.
	if game.HasNoManaCost(card) {
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
	if from == "command" {
		fromZone = game.ZoneCommand
	}
	cost, err = e.g.ApplyCostModifiersForEffect(cost, game.CostQuery{
		Card:       card,
		Controller: e.seat,
		FromZone:   fromZone,
	})
	if err != nil {
		return
	}
	x, ok := e.affordableX(cost, game.ManaSpendForCast(card))
	if !ok {
		return
	}

	// Modes → each choice of modes yields a target spec (at most one
	// chosen mode may carry a target clause; the engine rejects two).
	modeSets := [][]int{nil}
	if ms := game.ModeSpecFor(game.CatalogKey(card)); ms != nil {
		modeSets = legalModeSets(ms)
		if len(modeSets) == 0 {
			return
		}
	}

	// Additional costs (CR 601.2f). Discards choose from the rest of
	// the hand; a sacrifice chooses from the seat's own permanents
	// matching the clause.
	addCost := game.AdditionalCostFor(game.CatalogKey(card))
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
	for _, modes := range modeSets {
		spec := castTargetSpec(game.CatalogKey(card), modes)
		targetSets := [][]game.TargetRef{nil}
		if spec != nil {
			targetSets = e.legalTargetSets(spec, budget)
			if len(targetSets) == 0 {
				continue
			}
		}
		for _, targets := range targetSets {
			for _, discards := range discardSets {
				for _, sacs := range sacrificeSets {
					if budget <= 0 {
						return
					}
					budget--
					label := "Cast " + card.Name
					if from == "command" {
						label += " from the command zone"
					}
					if x > 0 {
						label += fmt.Sprintf(" for X=%d", x)
					}
					label += targetLabel(g, targets)
					e.add(Move{
						Type:   TypeCastSpell,
						Player: e.seat,
						Kind:   KindCast,
						Label:  label,
						Source: card.InstanceID,
						Params: mustJSON(castParams{
							InstanceID:   card.InstanceID.String(),
							FromZone:     from,
							Targets:      wireTargets(targets),
							Modes:        modes,
							XValue:       x,
							DiscardIDs:   idStrings(discards),
							SacrificeIDs: idStrings(sacs),
							Strict:       true,
							AutoTap:      true,
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

// affordableX reports whether the seat can pay cost right now — from
// the floating pool, or by the auto-tapper's plan — and, for an {X}
// cost, the largest X it can pay up to MaxX. Mirrors the engine's
// auto_tap + strict path: the pool is consulted first, then a plan
// is sought for the WHOLE cost (the engine does not net floating
// mana against the plan).
func (e *enumerator) affordableX(cost game.ParsedCost, spend game.ManaSpendContext) (int, bool) {
	return e.affordableXFrom(cost, spend, 0)
}

// affordableXFrom is affordableX with a floor on the announcement —
// an activated ability whose printed text says "X can't be 0" (Helm
// of Obedience) has a floor of 1, and a seat that cannot pay for
// X=1 has no legal activation at all rather than a free one at X=0.
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

// castTargetSpec mirrors game.castTargetSpec: the card-level spec
// wins; otherwise the single chosen mode that carries a clause. Two
// targeted modes never reach here — legalModeSets excludes them.
func castTargetSpec(oracleID string, modes []int) *game.TargetSpec {
	if spec := game.TargetSpecFor(oracleID); spec != nil {
		return spec
	}
	ms := game.ModeSpecFor(oracleID)
	if ms == nil {
		return nil
	}
	for _, m := range modes {
		if m >= 0 && m < len(ms.Options) && ms.Options[m].Targets != nil {
			return ms.Options[m].Targets
		}
	}
	return nil
}

// legalModeSets lists every distinct mode selection of size
// Min..Max, excluding any with more than one targeted option (the
// engine's castTargetSpec rejects those).
func legalModeSets(ms *game.ModeSpec) [][]int {
	n := len(ms.Options)
	hi := ms.Max
	if hi <= 0 || hi > n {
		hi = n
	}
	var out [][]int
	var rec func(start int, cur []int)
	rec = func(start int, cur []int) {
		if len(cur) >= ms.Min && len(cur) <= hi {
			targeted := 0
			for _, m := range cur {
				if ms.Options[m].Targets != nil {
					targeted++
				}
			}
			if targeted <= 1 {
				out = append(out, append([]int(nil), cur...))
			}
		}
		if len(cur) == hi {
			return
		}
		for i := start; i < n; i++ {
			rec(i+1, append(cur, i))
		}
	}
	rec(0, nil)
	return out
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
