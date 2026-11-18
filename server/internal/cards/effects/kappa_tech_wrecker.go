package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kappa Tech-Wrecker — Creature — Turtle Ninja {1}{G}, 1/3 (Neon Dynasty):
//
//	"Ninjutsu {1}{G}
//	 This creature enters with a deathtouch counter on it.
//	 Whenever this creature deals combat damage to a player, you may remove
//	 a deathtouch counter from it. When you do, exile target artifact or
//	 enchantment that player controls."
//
// Biting-Palm Ninja's shape (s82_neo_b06_helpers.go) with a deathtouch
// counter and a targeted reflexive trigger: the target is chosen as it goes
// on the stack, from what the damaged player controls, and re-checked as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2bc92c16-cd1e-4c32-9102-4593d77a9a9f",
		Name:         "Kappa Tech-Wrecker",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{1}{G}")},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterDeathtouch, 1, "Kappa Tech-Wrecker: enters with a deathtouch counter on it"),
		},
		Triggered: []game.TriggeredAbility{
			Optional(WheneverThisDealsCombatDamageToAPlayer("Kappa Tech-Wrecker — you may remove a deathtouch counter from it; when you do, exile target artifact or enchantment that player controls",
				s82NeoB06RemoveKeywordCounterThen(game.CounterDeathtouch,
					"Kappa Tech-Wrecker — exile target artifact or enchantment that player controls",
					kappaTechWreckerExileBody)),
				"Kappa Tech-Wrecker — remove a deathtouch counter? (When you do, exile target artifact or enchantment that player controls.)"),
		},
	})
}

var kappaTechWreckerExileBody = game.ReflexiveBody("s82-neo-b06/kappa-tech-wrecker-exile",
	simpleBody(exileFirstLegalCardTarget),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetPermanent("target artifact or enchantment that player controls", Or(Artifact(), Enchantment()), ControlledBy(p.Player))
	})
