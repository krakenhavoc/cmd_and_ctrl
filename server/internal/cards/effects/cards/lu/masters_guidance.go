package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Master's Guidance — Enchantment {2}{G} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Whenever you attack with two or more legendary creatures, put a +1/+1
//	 counter on each of up to two target attacking creatures.
//	 At the beginning of your end step, if you control a creature with
//	 power 4 or greater, draw a card."
//
// The attack trigger is one per declaration (OncePerBatch), and only when
// two or more of the creatures you declared are legendary, read through
// the layered view; a creature put onto the battlefield attacking was not
// declared and does not count (the rulings). Its targets may be any
// attacking creatures, legendary or not (the rulings).
//
// The draw is an intervening "if" (CR 603.4), Bugenhagen's
// CondPowerAtLeast at 4: checked as the end step begins and again as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66de5b97-0de0-422c-a351-bfe55ad4503a",
		Name:         "Master's Guidance",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OncePerBatch(effects.On(game.EventAttack, effects.AvB34YouAttackedWithTwoLegendaryCreatures,
				"Master's Guidance — put a +1/+1 counter on each of up to two target attacking creatures",
				effects.PutPlusOneCounterOnEachLegalTarget)),
				effects.TargetCreature("up to two target attacking creatures", effects.AttackingCreature()).WithCount(0, 2)),
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ByYou(ev, source, lki, g) && effects.B24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
			}, "Master's Guidance — draw a card", effects.Do(effects.If{
				Cond: effects.CondPowerAtLeast{N: 4},
				Then: []effects.Applier{effects.DrawCards{N: 1}},
			})),
		},
	})
}
