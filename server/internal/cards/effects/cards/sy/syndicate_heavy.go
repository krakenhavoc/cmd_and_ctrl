package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Syndicate Heavy — Creature — Giant Rogue {2}{W/B}{W/B}, 4/4 (Ravnica: Clue Edition, #5929): "Extort. At the beginning of each end step, if you gained 4 or more life this turn, investigate."
//
// The intervening-if (CR 603.4) is checked on the trigger and again on resolution (Indulging Patrician).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01becf76-aa53-4fd5-94c1-665c64674219",
		Name:         "Syndicate Heavy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Extort("Syndicate Heavy"),
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B32EndStepAndYouGainedLifeThisTurnAtLeast(ev, source, g, 4)
			}, "Syndicate Heavy — investigate",
				effects.Do(effects.If{Cond: effects.CondLifeGained{AtLeast: 4}, Then: []effects.Applier{effects.CreateToken{Template: effects.ClueToken(), N: 1}}})),
		},
	})
}
