package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Griffin Aerie — Enchantment {1}{W} (Jumpstart):
//
//	"At the beginning of your end step, if you gained 3 or more life this
//	 turn, create a 2/2 white Griffin creature token with flying."
//
// Indulging Patrician's intervening if (CR 603.4): checked as your end step
// begins and again as the trigger resolves. The tally counts life gained,
// not net life, so a gain of 3 followed by a loss of 3 is still "gained 3".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7cab62dc-a256-4da8-a3ea-b3903ec37077",
		Name:         "Griffin Aerie",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.B32EndStepAndYouGainedLifeThisTurnAtLeast(ev, source, g, 3)
			}, "Griffin Aerie — create a 2/2 white Griffin creature token with flying",
				effects.Do(effects.If{
					Cond: effects.CondLifeGained{AtLeast: 3},
					Then: []effects.Applier{effects.CreateToken{Template: effects.TokenCard("2/2 white Griffin with flying"), N: 1}},
				})),
		},
	})
}
