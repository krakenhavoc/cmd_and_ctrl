package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merfolk Seer — Creature — Merfolk Wizard {2}{U}, 2/2 (Mirage, S82 OLD b800,
// #5633):
//
//	"When this creature dies, you may pay {1}{U}. If you do, draw a card."
//
// Carrion Thrash's optional payment behind a dies trigger: the payment is
// asked as the trigger resolves (CR 603.5, 118.12), and the draw is made only
// if the mana was really paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2351a9fa-17b5-41fc-93d0-34731d54f915",
		Name:         "Merfolk Seer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Merfolk Seer — you may pay {1}{U}; if you do, draw a card", effects.OldB800MerfolkSeerDies),
		},
	})
}
