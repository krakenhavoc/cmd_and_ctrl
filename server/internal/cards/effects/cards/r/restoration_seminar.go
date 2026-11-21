package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restoration Seminar — Sorcery — Lesson {5}{W}{W} (Secrets of Strixhaven,
// #4456):
//
//	"Return target nonland permanent card from your graveyard to the
//	 battlefield.
//	 Paradigm (Then exile this spell. After you first resolve a spell
//	 with this name, you may cast a copy of it from exile without paying
//	 its mana cost at the beginning of each of your first main phases.)"
//
// Dread Return's body with a wider target: any nonland permanent card in
// your graveyard. Paradigm is the printed keyword and the engine does the
// rest (game/paradigm.go, CR 702.192). Each free copy picks its own
// target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fd0a27a9-7d7e-4a16-9dd0-0715e6c97b36",
		Name:            "Restoration Seminar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordParadigm},
		Targets:         effects.TargetCardInGraveyard("target nonland permanent card from your graveyard", effects.YouOwn(), effects.NonlandPermanentCard()),
		Purpose:         effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnFirstLegalGraveyardTargetToBattlefield(ctx.Game, item)
		},
	})
}
