package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mark of the Oni — Enchantment — Aura {2}{B} (Betrayers of Kamigawa, S82 OLD b594):
//
//	"Enchant creature
//	 You control enchanted creature.
//	 At the beginning of the end step, if you control no Demons, sacrifice this Aura."
//
// Control Magic and Glimmervoid's end-step trigger. It watches every end step, not only
// its controller's; the "if" is intervening (CR 603.4): it triggers only when the Aura's
// controller controls no Demon as the end step begins, and does nothing if one has arrived
// by the time it resolves. "Demon" is a word of the Aura's text (CR 612.2). A creature
// that is itself a Demon under the Aura's controller counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ce00fdf-c4da-492a-81c9-8c93dd0bf620",
		Name:         "Mark of the Oni",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control no Demons",
				Test: func(ctx *effects.Context) bool {
					return !effects.ControlsA(ctx.TextSubtype("Demon"))(ctx.Game, ctx.Controller())
				},
			}, effects.On(game.EventBeginEndStep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B15EndStepBegan(ev) && !effects.ControlsA(source.TextSubtype("Demon"))(g, source.Controller)
				},
				"Mark of the Oni — if you control no Demons, sacrifice this Aura",
				effects.SacrificeThisIfStillOnBattlefield)),
		},
	})
}
