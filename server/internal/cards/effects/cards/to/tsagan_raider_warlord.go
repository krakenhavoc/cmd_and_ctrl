package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tsagan, Raider Warlord — Legendary Creature — Dinosaur Berserker
// {R}{W}{B}, 1/4 (Mystery Booster Commander Edition):
//
//	"Double strike
//	 Start your engines! (If you have no speed, it starts at 1. It
//	 increases once on each of your turns when an opponent loses life.
//	 Max speed is 4.)
//	 Whenever Tsagan attacks, creatures you control get +1/+0 until end of
//	 turn for each creature you control with first strike or double strike.
//	 Max speed — Tsagan has deathtouch and other creatures you control
//	 have first strike."
//
// The count is taken as the trigger resolves (CR 608.2h) and the set that
// gets the bonus is locked then (CR 611.2c). It reads each creature's
// effective keywords, so a creature given first strike by Tsagan's own
// max-speed static counts. Both max-speed clauses are layer 6 grants
// switched on by the controller's speed (CR 702.178a).
//
// No simplification.
func init() {
	const label = "Tsagan, Raider Warlord — creatures you control get +1/+0 until end of turn for each creature you control with first strike or double strike"
	statics := effects.MaxSpeedSelfKeywords("deathtouch")
	statics = append(statics, effects.MaxSpeedStatic(effects.KeywordGrant(effects.OtherCreaturesYouControl, "first strike")))
	effects.Register(effects.Spec{
		OracleID:        "3194c964-1693-40b6-b4f2-64fe056e7cd9",
		Name:            "Tsagan, Raider Warlord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike", effects.StartYourEngines},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				me := ctx.Controller()
				n := effects.CountControlled(g, me, func(c game.Card) bool {
					return c.IsCreature() && (game.HasKeyword(&c, "first strike") || game.HasKeyword(&c, "double strike"))
				})
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: n, Label: "Tsagan, Raider Warlord — +1/+0 for each first or double strike creature"}.Apply(ctx)
			}),
		},
		Static: statics,
	})
}
