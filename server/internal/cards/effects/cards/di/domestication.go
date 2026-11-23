package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Domestication — Enchantment — Aura {2}{U}{U} (S82 GAP, #5948):
//
//	"Enchant creature
//	 You control enchanted creature.
//	 At the beginning of your end step, if enchanted creature's power is 4 or
//	 greater, sacrifice this Aura."
//
// Control Magic's control effect; the end-step trigger's intervening "if"
// is checked as the step begins and again as it resolves (CR 603.4).
func init() {
	const label = "Domestication — sacrifice this Aura"
	effects.Register(effects.Spec{
		OracleID:     "86da33ac-f8ac-48c8-a50b-22fd761dff8f",
		Name:         "Domestication",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "enchanted creature's power is 4 or greater",
				Test: auraHostPowerAtLeastFour,
			}, effects.AtYourEndStep(label, effects.SacrificeThisIfStillOnBattlefield)),
		},
	})
}

// auraHostPowerAtLeastFour reads the enchanted creature's power through the
// layers as the ability's source Aura sits on it.
func auraHostPowerAtLeastFour(ctx *effects.Context) bool {
	host := effects.AttachedHostFor(ctx.Game, ctx.Source())
	return host != nil && effects.B31CurrentPowerOf(ctx.Game, host.InstanceID) >= 4
}
