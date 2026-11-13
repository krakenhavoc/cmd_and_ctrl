package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gainLifeEqualToSacrificedToughness is the whole Effect of "…,
// Sacrifice a creature: You gain life equal to the sacrificed
// creature's toughness." (Miren, the Moaning Well; Diamond Valley,
// #5641). The toughness is the creature's as it last existed on the
// battlefield (counters and an anthem's bonus count), and a creature
// whose toughness was zero or less gains nothing.
func gainLifeEqualToSacrificedToughness(g *game.Game, item *game.StackItem) error {
	fed, ok := b17PermanentSacrificedToPay(g, item)
	if !ok {
		return nil
	}
	toughness := departedCreatureToughness(g, fed)
	if toughness <= 0 {
		return nil
	}
	return GainLife{Player: item.Controller, Amount: toughness}.Apply(NewContext(g, item))
}
