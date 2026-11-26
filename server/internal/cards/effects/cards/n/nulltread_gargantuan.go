package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nulltread Gargantuan — Creature — Beast {1}{G}{U}, 5/6 (Alara Reborn, #5586):
//
//	"When this creature enters, put a creature you control on top of its owner's
//	 library."
//
// Emancipation Angel's resolution-time pick, not a target: you choose one of your
// own creatures as the trigger resolves, the Gargantuan itself included, and it
// goes on top of its owner's library (a token ceases to exist). Mandatory with at
// least one creature, which the Gargantuan itself always is while it is there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acc74d59-f587-46a2-a90b-865f721d494d",
		Name:         "Nulltread Gargantuan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Nulltread Gargantuan — put a creature you control on top of its owner's library",
				effects.S82OldB504PutACreatureYouControlOnTop("Nulltread Gargantuan — put a creature you control on top of its owner's library")),
		},
	})
}
