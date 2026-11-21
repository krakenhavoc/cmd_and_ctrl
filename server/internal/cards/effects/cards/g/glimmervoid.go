package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glimmervoid — Land (Mirrodin, #5606):
//
//	"At the beginning of the end step, if you control no artifacts,
//	 sacrifice this land.
//	 {T}: Add one mana of any color."
//
// The trigger watches every end step, not only its controller's. Its
// "if" is an intervening if (CR 603.4): it triggers only when its
// controller controls no artifacts as the end step begins, and does
// nothing if an artifact has arrived by the time it resolves. A land
// that has already left is not sacrificed. The mana ability is the
// any-colour rock's (samiAnyColorMana).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b92e9854-4527-4133-8615-e282a213e7e3",
		Name:         "Glimmervoid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control no artifacts",
				Test: func(ctx *effects.Context) bool {
					return !effects.ControlsAnyCard(ctx.Game, ctx.Controller(), effects.Artifact())
				},
			}, effects.On(game.EventBeginEndStep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B15EndStepBegan(ev) && !effects.ControlsAnyCard(g, source.Controller, effects.Artifact())
				},
				"Glimmervoid — if you control no artifacts, sacrifice it",
				effects.SacrificeThisIfStillOnBattlefield)),
		},
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Tap: true}, "{T}: Add one mana of any color"),
		},
	})
}
