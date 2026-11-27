package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merrow Bonegnawer — {B} Creature — Merfolk Rogue, 1/1:
//
//	"{T}: Target player exiles a card from their graveyard.
//	 Whenever you cast a black spell, you may untap this creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dad36dc0-39aa-46f9-819b-cad7a578c037",
		Name:         "Merrow Bonegnawer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target player exiles a card from their graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.S82OldB524TargetExilesOwnGraveyardCard("Merrow Bonegnawer"),
		}},
		Triggered: []game.TriggeredAbility{effects.EveB1UntapOnCast("Merrow Bonegnawer", "B")},
	})
}
