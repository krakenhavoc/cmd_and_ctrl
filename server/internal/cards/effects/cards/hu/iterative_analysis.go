package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Iterative Analysis — Conspiracy (Conspiracy, #6102):
//
//	"Hidden agenda (Start the game with this conspiracy face down in the
//	 command zone and secretly choose a card name. You may turn this
//	 conspiracy face up any time and reveal that name.)
//	 Whenever you cast an instant or sorcery spell with the chosen name,
//	 you may draw a card."
//
// Hidden agenda is the engine's (ADR 0227). Face up, a cast trigger
// harvested from the command zone, with the optional draw; face down it
// has no abilities (CR 315.5b), so a spell cast before the reveal draws
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79172853-240c-4924-8b20-edf469946df4",
		Name:         "Iterative Analysis",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaOneName,
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(
				effects.On(game.EventCast, effects.YouCastASpellWithTheChosenName(iterativeAnalysisInstantOrSorcery),
					"Iterative Analysis — draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Iterative Analysis — draw a card?", 1),
		},
	})
}

func iterativeAnalysisInstantOrSorcery(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.IsInstant() || c.IsSorcery()
}
