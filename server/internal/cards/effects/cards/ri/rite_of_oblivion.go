package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rite of Oblivion — Sorcery {W}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a nonland
//	 permanent.
//	 Exile target nonland permanent.
//	 Flashback {2}{W}{B}"
//
// The sacrifice is paid on every cast, flashback included (the 2021
// ruling, CR 601.2f).
//
// No simplification. Star Trek Commander (#4974).
func init() {
	effects.Register(effects.Spec{
		OracleID:         "99957aa7-b654-41b5-9b96-7bce47545d25",
		Name:             "Rite of Oblivion",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost:   effects.SacrificeCost("a nonland permanent", effects.Nonland()),
		Targets:          effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{2}{W}{B}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTheFirstTargetIfLegal(ctx)
		},
	})
}
