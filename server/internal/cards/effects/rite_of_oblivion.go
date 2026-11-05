package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "99957aa7-b654-41b5-9b96-7bce47545d25",
		Name:             "Rite of Oblivion",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost:   SacrificeCost("a nonland permanent", Nonland()),
		Targets:          TargetPermanent("target nonland permanent", Nonland()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{W}{B}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileTheFirstTargetIfLegal(ctx)
		},
	})
}
