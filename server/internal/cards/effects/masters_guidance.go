package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "66de5b97-0de0-422c-a351-bfe55ad4503a",
		Name:         "Master's Guidance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(OncePerBatch(On(game.EventAttack, avB34YouAttackedWithTwoLegendaryCreatures,
				"Master's Guidance — put a +1/+1 counter on each of up to two target attacking creatures",
				putPlusOneCounterOnEachLegalTarget)),
				TargetCreature("up to two target attacking creatures", AttackingCreature()).WithCount(0, 2)),
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ByYou(ev, source, lki, g) && b24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
			}, "Master's Guidance — draw a card", Do(If{
				Cond: CondPowerAtLeast{N: 4},
				Then: []Applier{DrawCards{N: 1}},
			})),
		},
	})
}
