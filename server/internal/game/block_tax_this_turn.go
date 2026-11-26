package game

import (
	"fmt"

	"github.com/google/uuid"
)

// block_tax_this_turn.go — "This turn, creatures can't block unless their
// controller pays {X} for each blocking creature they control" (War Cadence,
// S82 OLD b385; CR 509.1d).
//
// A ScopedEffect record (ScopeGame, until end of turn) with one
// ModBlockTaxThisTurn mod whose Amount is X, the generic mana each blocking
// creature costs. Every blocking creature of every player owes it, read where
// the permanents' own block taxes are priced (priceBlockPlansLocked), so the
// declaration is priced, refused and paid exactly as Archangel of Tithes'.
// Being a game rule from a resolved ability it outlives its source (CR 611.2).

// ModBlockTaxThisTurn is "creatures can't block this turn unless their
// controller pays {Amount} for each of those creatures". Reads Amount.
const ModBlockTaxThisTurn ModKind = "blockTaxThisTurn"

// blockTaxThisTurnModProblem is the registration and restore check. "" when
// the mod is sound or is another kind.
func blockTaxThisTurnModProblem(m Mod) string {
	if m.Kind != ModBlockTaxThisTurn {
		return ""
	}
	if m.Amount < 1 {
		return fmt.Sprintf("a blockTaxThisTurn mod charges {%d}", m.Amount)
	}
	return ""
}

// BlockTaxThisTurnForEffect records that, until end of turn, each creature
// that blocks costs its controller {x}. Returns false, writing nothing, for an
// x below 1 (a tax of nothing). `label` is the clause as printed.
//
// Caller must hold g.mu (write).
func (g *Game) BlockTaxThisTurnForEffect(source uuid.UUID, x int, label string) bool {
	if x < 1 {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(source, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModBlockTaxThisTurn, Amount: x}}, g.UntilEndOfTurnDuration(), label)
}

// priceBlockTaxesThisTurnLocked adds, for one blocking creature, the line each
// live ModBlockTaxThisTurn record charges.
//
// Caller holds g.mu.
func (g *Game) priceBlockTaxesThisTurnLocked(out *BlockTaxPrice, b interface{ WriteString(string) (int, error) }, blk *Card) {
	for _, e := range g.ScopedEffects {
		if len(e.Mods) != 1 || e.Mods[0].Kind != ModBlockTaxThisTurn || e.Mods[0].Amount < 1 {
			continue
		}
		mana := fmt.Sprintf("{%d}", e.Mods[0].Amount)
		_, _ = b.WriteString(mana)
		out.Lines = append(out.Lines, BlockTaxLine{
			Source:  e.Source.ID,
			Label:   e.Label,
			Blocker: blk.InstanceID,
			Mana:    mana,
		})
	}
}
