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
			card := c
			if card.IsLand() {
				// CR 305: main phase, empty stack, your turn, and one
				// per turn. The engine enforces the first three and
				// not the fourth; we enforce all four.
				if zone.from == "hand" && speed && landOwed {
					e.add(Move{
						Type:   TypeCastSpell,
						Player: e.seat,
						Kind:   KindLand,
						Label:  "Play " + card.Name,
						Source: card.InstanceID,
						Params: mustJSON(castParams{InstanceID: card.InstanceID.String(), FromZone: "hand"}),
					})
				}
				continue
			}
			e.castMovesForCard(card, zone.from, speed)
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
	if game.TargetModeFor(card.OracleID) != "" && game.TargetSpecFor(card.OracleID) == nil {
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
	if from == "command" {
		cost.Generic += p.CommanderCasts[card.InstanceID] * 2
	}
	x, ok := e.affordableX(cost, game.ManaSpendForCast(card))
	if !ok {
		return
	}

	// Modes → each choice of modes yields a target spec (at most one
	// chosen mode may carry a target clause; the engine rejects two).
	modeSets := [][]int{nil}
	if ms := game.ModeSpecFor(card.OracleID); ms != nil {
		modeSets = legalModeSets(ms)
		if len(modeSets) == 0 {
			return
		}
	}

	// Additional costs (CR 601.2f). Discards choose from the rest of
	// the hand; a sacrifice chooses from the seat's own permanents
	// matching the clause.
	addCost := game.AdditionalCostFor(card.OracleID)
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
			sacrificeSets = combinations(pool, 1, 1, e.opts.MaxExpansionPerSource)
			if len(sacrificeSets) == 0 {
				return
			}
		}
	}

	budget := e.opts.MaxExpansionPerSource
	for _, modes := range modeSets {
		spec := castTargetSpec(card.OracleID, modes)
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
	if cost.XSlots == 0 {
		return 0, e.canPay(cost, 0, spend)
	}
	best, ok := -1, false
	for x := 0; x <= e.opts.MaxX; x++ {
		if e.canPay(cost, x, spend) {
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
	if e.p.ManaPool.CanPayFor(cost, x, spend) {
		return true
	}
	_, ok := e.g.AutoTapForCostForEffect(e.seat, cost, x)
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
