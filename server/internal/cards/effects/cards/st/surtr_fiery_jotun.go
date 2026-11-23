package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surtr, Fiery Jötun — Legendary Creature — Giant God Warrior {3}{R}{R}, 5/5
// (Assassin's Creed, #5948):
//
//	"Trample
//	 Whenever you cast a historic spell, Surtr deals 3 damage to any target.
//	 (Artifacts, legendaries, and Sagas are historic.)"
//
// Jhoira's historic cast watch (youCastAHistoricSpell: an artifact, a
// legendary or a Saga, read off the spell on the stack) with Omnath's
// "deals 3 damage to any target" body. The target is chosen as the trigger
// goes on the stack; Surtr, gone by resolution, still deals the damage as the
// object it was (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "424084cf-418f-47b4-915f-b254c8adb3bd",
		Name:            "Surtr, Fiery Jötun",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCast, effects.YouCastAHistoricSpell,
				"Surtr, Fiery Jötun — 3 damage to any target", effects.DamageToFirstTarget(3)),
				effects.TargetAny()),
		},
	})
}
