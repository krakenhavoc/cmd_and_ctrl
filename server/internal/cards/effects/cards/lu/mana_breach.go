package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mana Breach — Enchantment {2}{U} (Exodus, S82 OLD batch 743, #5626):
//
//	"Whenever a player casts a spell, that player returns a land they control
//	 to its owner's hand."
//
// Aether Barrier's any-player cast trigger: the caster, whoever they are,
// chooses one of the lands they control as the trigger resolves, and it goes
// to its owner's hand (a stolen land goes back to its owner). A caster with no
// land returns nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "221d0305-94db-4791-9dc7-89ee96c7fcd6",
		Name:         "Mana Breach",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB743APlayerCastASpell, "Mana Breach — that player returns a land they control to its owner's hand", effects.S82OldB743ManaBreach),
		},
	})
}
