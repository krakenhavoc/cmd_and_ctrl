package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wildwood Geist — {4}{G} Creature — Spirit 3/3: "During your turn, this
// creature gets +2/+2."
//
// A layer 7c static that applies only while its controller is the active
// player.
//
// No simplification.
func init() {
	duringYourTurn := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
	}
	effects.Register(effects.Spec{
		OracleID:     "9fb4aa2d-6414-4176-9070-e5230e100c6a",
		Name:         "Wildwood Geist",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "During your turn, this creature gets +2/+2.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: duringYourTurn,
			Apply:     effects.PumpBy(2, 2),
		}},
	})
}
