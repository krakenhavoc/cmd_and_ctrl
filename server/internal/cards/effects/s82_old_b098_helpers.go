package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b098_helpers.go — Jumpstart enters and dies triggers (S82 older
// sets, batch 098).

// S82OldB098TargetCreatureGetsUntilEOT is "target creature gets +P/+T until
// end of turn" as a trigger's whole Effect: the first still-legal target
// card gets the boost, and a target that has become illegal is skipped
// (CR 608.2b).
func S82OldB098TargetCreatureGetsUntilEOT(label string, power, toughness int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, id := range LegalTargetCardIDs(ctx) {
			return BoostUntilEOT{Target: id, Power: power, Toughness: toughness, Label: label}.Apply(ctx)
		}
		return nil
	}
}
