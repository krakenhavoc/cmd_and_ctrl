package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sarevok, Deathbringer — {3}{B} Legendary Creature — Human Knight 3/4
//
//	"At the beginning of each player's end step, if no permanents left the
//	 battlefield this turn, that player loses X life, where X is Sarevok's
//	 power.
//	 Choose a Background"
//
// An intervening "if" (CR 603.4), read as the end step begins and again as
// the trigger resolves. X is Sarevok's power as it resolves, or as he last
// existed once he has left (CR 608.2h). "Choose a Background" is the
// deck-construction rule (CR 702.124k) read from the oracle text.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e215ede-e4ac-482a-a2e1-cacb20affa70",
		Name:         "Sarevok, Deathbringer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.GapB213CondNoPermanentLeft,
				effects.On(game.EventBeginEndStep,
					func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
						return effects.GapB213NoPermanentLeft(g)
					},
					"Sarevok, Deathbringer — that player loses X life, where X is Sarevok's power",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						info, ok := ctx.SourcePermanent()
						if !ok || item.Trigger == nil {
							return nil
						}
						x := info.Power
						if x <= 0 {
							return nil
						}
						return g.ChangePlayerLifeForEffect(ctx.Source(), item.Trigger.Event.Actor, -x)
					})),
		},
	})
}
