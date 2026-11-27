package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tangleroot — Artifact {3} (Mirrodin, #5606):
//
//	"Whenever a player casts a creature spell, that player adds {G}."
//
// A triggered ability, not a mana ability (it has no cost and does not resolve
// as part of a mana ability, CR 605.1b): it goes on the stack above the spell,
// and the caster, named by the cast event, adds {G} as it resolves, in time to
// spend it before the step ends.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b94efd70-705d-4cdf-9990-3a065a541f28",
		Name:         "Tangleroot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.OldB607CastACreatureSpell, "Tangleroot — that player adds {G}", effects.OldB607CasterAddsG),
		},
	})
}
