package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const (
	elrondScryLabel   = "Elrond, Master of Healing — put a +1/+1 counter on each of up to X target creatures, where X is the number of cards looked at while scrying"
	elrondTargetLabel = "Elrond, Master of Healing — draw a card"
)

// Elrond, Master of Healing — Legendary Creature — Elf Noble {2}{G}{U}, 4/4
// (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Whenever you scry, put a +1/+1 counter on each of up to X target
//	 creatures, where X is the number of cards looked at while scrying this
//	 way.
//	 Whenever a creature you control with a +1/+1 counter on it becomes the
//	 target of a spell or ability an opponent controls, you may draw a
//	 card."
//
// The first ability sizes its target clause from the event it triggered
// on (TargetsFrom hands it the scry's LookedAt, Celeborn's count): a scry
// of 2 offers up to two targets. The second is Cactarantula's targeted-by-
// an-opponent draw, for any creature of its controller's that has a +1/+1
// counter when it is targeted (read as the event fired).
//
// No simplification.
func init() {
	scry := game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventScry},
		AppliesTo: effects.ByYou,
		Key:       elrondScryLabel,
		TargetsFrom: func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
			x := tc.Event.LookedAt
			if x <= 0 {
				return nil
			}
			return effects.TargetCreature("up to X target creatures").WithCount(0, x)
		},
		Effect: effects.PutPlusOneCounterOnEachLegalTarget,
	}
	effects.Register(effects.Spec{
		OracleID:     "28aefa3a-16d3-4207-bb8c-588b722a4041",
		Name:         "Elrond, Master of Healing",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			scry,
			effects.OptionalDraw(effects.On(game.EventBecomesTarget, effects.GapB125CounteredCreatureTargetedByOpponent,
				elrondTargetLabel, effects.Do(effects.DrawCards{N: 1})),
				"Elrond, Master of Healing — draw a card?", 1),
		},
	})
}
