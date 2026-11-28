package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thran Quarry — Land {—} (Urza's Saga, S82 OLD batch 731, #5624):
//
//	"At the beginning of the end step, if you control no creatures, sacrifice
//	 this land.
//	 {T}: Add one mana of any color."
//
// Glimmervoid's shape with creatures: the trigger is the beginning of EVERY
// end step, not just yours, and its "if" is intervening (CR 603.4), checked as
// the step begins and again as the trigger resolves, so a creature that
// arrives in response saves the land. The mana ability is a plain any-color tap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57b4da3f-361a-4cbe-b77f-190ec33eefd8",
		Name:         "Thran Quarry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control no creatures",
				Test: func(ctx *effects.Context) bool {
					return !effects.ControlsAnyCard(ctx.Game, ctx.Controller(), effects.Creature())
				},
			}, effects.On(game.EventBeginEndStep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B15EndStepBegan(ev) && !effects.ControlsAnyCard(g, source.Controller, effects.Creature())
				},
				"Thran Quarry — if you control no creatures, sacrifice it",
				effects.SacrificeThisIfStillOnBattlefield)),
		},
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Tap: true}, "{T}: Add one mana of any color"),
		},
	})
}
