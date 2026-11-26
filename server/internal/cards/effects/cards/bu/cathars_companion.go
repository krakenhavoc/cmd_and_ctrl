package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cathar's Companion — Creature — Dog {2}{W}, 3/1 (Jumpstart, #5510):
//
//	"Whenever you cast a noncreature spell, this creature gains indestructible
//	 until end of turn. (Damage and effects that say "destroy" don't destroy
//	 it.)"
//
// Seeker of the Way's cast trigger with indestructible.
//
// No simplification.
func init() {
	const label = "Cathar's Companion — this creature gains indestructible until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "668f646b-94f0-4f1a-badd-2d69b559d8a2",
		Name:         "Cathar's Companion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou, label,
				effects.ThisGainsKeywordUntilEndOfTurn("indestructible", label)),
		},
	})
}
