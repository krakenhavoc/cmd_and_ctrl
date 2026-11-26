package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Double Stroke — Conspiracy (Conspiracy, S82 OLD batch 361, #5557):
//
//	"Hidden agenda (Start the game with this conspiracy face down in the
//	 command zone and secretly choose a card name. You may turn this
//	 conspiracy face up any time and reveal that name.)
//	 Whenever you cast an instant or sorcery spell with the chosen name, you
//	 may copy it. You may choose new targets for the copy."
//
// Hidden agenda is the engine's (ADR 0227). Face down it has no abilities
// (CR 315.5b). The copy is created, not cast, so it triggers nothing; its
// new targets are its controller's to choose (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e25ab152-e166-44df-8f92-668bab0e9619",
		Name:         "Double Stroke",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaOneName,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.On(game.EventCast, effects.YouCastASpellWithTheChosenName(doubleStrokeInstantOrSorcery),
					"Double Stroke — copy that spell; you may choose new targets for the copy",
					effects.CopyTheTriggeringSpellWithNewTargets),
				"Double Stroke — copy that spell?"),
		},
	})
}

func doubleStrokeInstantOrSorcery(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.IsInstant() || c.IsSorcery()
}
