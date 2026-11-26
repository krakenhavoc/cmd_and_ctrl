package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gelectrode — Creature — Weird {1}{U}{R}, 0/1 (Duel Decks: Izzet vs. Golgari):
//
//	"{T}: This creature deals 1 damage to any target.
//	 Whenever you cast an instant or sorcery spell, you may untap this
//	 creature."
//
// Rod of Ruin's ping with {T} as the whole cost (so the creature must have been
// under your control since your turn began, CR 302.6), and a "you may" cast
// trigger that untaps it only while it is still the permanent that triggered
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8039df5c-b26b-4277-8226-4c08a5a69fbb",
		Name:         "Gelectrode",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DamageToFirstTarget(1),
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Gelectrode — untap this creature", effects.UntapThisPermanentEffect),
				"Gelectrode — untap it?"),
		},
	})
}
