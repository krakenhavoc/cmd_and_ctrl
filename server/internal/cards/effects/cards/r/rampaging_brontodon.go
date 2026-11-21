package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rampaging Brontodon — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "bfea2229-a1ca-40c1-aaad-5f97add3f079", Name: "Rampaging Brontodon", Completeness: effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{effects.WheneverThisAttacks("Rampaging Brontodon — +1/+1 for each land you control", func(g *game.Game, item *game.StackItem) error {
			n := effects.B10LandsControlled(g, item.Controller)
			return (effects.BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n, Label: "Rampaging Brontodon — lands"}).Apply(effects.NewContext(g, item))
		})},
	})
}
