package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forge Boss — Creature — Human Warrior {2}{B}{R}, 3/4 (Streets of New
// Capenna, S82):
//
//	"Whenever you sacrifice one or more other creatures, this creature
//	 deals 2 damage to each opponent. This ability triggers only once each
//	 turn."
//
// "One or more" is one trigger per batch of simultaneous sacrifices
// (OncePerBatch), and "only once each turn" shuts every later batch of the
// turn out (TriggersOnlyOnceEachTurn, #5245). The Boss itself does not
// count; the damage comes from the Boss as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c234d2ab-21d5-41a9-9ab8-47a8bd53a686",
		Name:         "Forge Boss",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.On(game.EventSacrifice,
				effects.SncB13YouSacrificedAnotherCreature,
				"Forge Boss — 2 damage to each opponent",
				effects.Do(effects.DamageEachOpponent{N: 2})))),
		},
	})
}
