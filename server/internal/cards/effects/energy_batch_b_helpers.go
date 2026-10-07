package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// energy_batch_b_helpers.go — bodies shared by ADR 0129 PR 1's energy
// cards (#1995). Append-only.

// ebTargetCreatureGainsUntilEOT is "Target creature gains <keyword> until
// end of turn" as an activated ability's body (Spontaneous Artist's
// haste). A target that left legality in response does nothing
// (CR 608.2b).
func ebTargetCreatureGainsUntilEOT(keyword, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return GrantKeywordUntilEOT{Target: id, Keywords: []string{keyword}, Label: label}.Apply(ctx)
	}
}

// ebYouGetEnergyPerCreatureYouControl is "you get {E} for each creature
// you control", counted as the ability resolves (Aetherwind Basker).
func ebYouGetEnergyPerCreatureYouControl(g *game.Game, item *game.StackItem) error {
	return GetEnergy{N: b04CreaturesControlled(g, item.Controller)}.Apply(NewContext(g, item))
}

// ebDrawACard is "Draw a card." as an activated ability's body.
func ebDrawACard(g *game.Game, item *game.StackItem) error {
	return DrawCards{N: 1}.Apply(NewContext(g, item))
}

// ebYouGetEnergy is "You get N {E}." as an activated or triggered
// ability's body.
func ebYouGetEnergy(n int) Effect {
	return Do(GetEnergy{N: n})
}
