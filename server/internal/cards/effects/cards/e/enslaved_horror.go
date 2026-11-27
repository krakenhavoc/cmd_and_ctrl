package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enslaved Horror — Creature — Horror {3}{B}, 4/4 (Mercadian Masques, S82 OLD
// batch 699, #5619):
//
//	"When this creature enters, each other player may return a creature card
//	 from their graveyard to the battlefield."
//
// Exhume's walk with a real "may" and the controller left out
// (S82OldB699EnslavedHorrorReturn): each other player with a creature card in
// their graveyard is asked in APNAP order, taking one or none, and the chosen
// cards enter together under their owners' control (CR 101.4, 110.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7920a2c-d14a-4ed6-8341-96751322f4e7",
		Name:         "Enslaved Horror",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Enslaved Horror — each other player may return a creature card from their graveyard to the battlefield",
				effects.S82OldB699EnslavedHorrorReturn),
		},
	})
}
