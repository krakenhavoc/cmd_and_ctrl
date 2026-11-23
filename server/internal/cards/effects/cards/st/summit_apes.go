package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summit Apes — Creature — Ape {3}{G}, 5/2 (Worldwake, #5582):
//
//	"As long as you control a Mountain, this creature has menace. (It can't
//	 be blocked except by two or more creatures.)"
//
// Kargan Dragonrider's shape: a layer-6 grant to itself while its
// controller controls a Mountain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01799149-4bca-417b-9e89-14493fd46c79",
		Name:         "Summit Apes",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{summitApesMenace()},
	})
}

func summitApesMenace() game.StaticAbility {
	ab := effects.KeywordGrant(effects.SelfWhileYouControlA(effects.Subtype("Mountain")), "menace")
	ab.Label = "As long as you control a Mountain, this creature has menace."
	return ab
}
