package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brago's Favor — Conspiracy (Conspiracy, #6102):
//
//	"Hidden agenda (Start the game with this conspiracy face down in the
//	 command zone and secretly choose a card name. You may turn this
//	 conspiracy face up any time and reveal that name.)
//	 Spells with the chosen name you cast cost {1} less to cast."
//
// Hidden agenda is the engine's (ADR 0227). Face up, a generic {1}
// reduction in the CR 601.2f pass, which reads cost modifiers from the
// command zone; face down it has none (CR 315.5b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "395b312b-8fe6-4c28-b91d-620b5fcb9e6d",
		Name:         "Brago's Favor",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaOneName,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Brago's Favor — spells with the chosen name you cast cost {1} less", effects.SpellsWithTheChosenNameYouCast()),
		},
	})
}
