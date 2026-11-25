package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Regal Bloodlord — Creature — Vampire Soldier {3}{W}{B}, 2/4 (Core Set 2019):
//
//	"Flying
//	 At the beginning of each end step, if you gained life this turn, create
//	 a 1/1 black Bat creature token with flying."
//
// Crested Sunmare's trigger: an intervening "if" (CR 603.4) checked when each
// end step begins and again on resolution, counting life gained by its
// controller this turn whatever the net change.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ce9d3382-23ca-4d35-ad30-5e2b75517012",
		Name:            "Regal Bloodlord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B33EndStepAndYouGainedLifeThisTurn(ev, source, g)
			}, "Regal Bloodlord — create a 1/1 black Bat creature token with flying",
				effects.Do(effects.If{
					Cond: effects.CondLifeGained{},
					Then: []effects.Applier{effects.CreateToken{Template: effects.TokenCard("1/1 black Bat with flying"), N: 1}},
				})),
		},
	})
}
