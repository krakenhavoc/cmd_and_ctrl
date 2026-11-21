package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treason of Isengard — Sorcery {2}{U} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Put up to one target instant or sorcery card from your graveyard on top
//	 of your library.
//	 Amass Orcs 2."
//
// "Up to one" target, so the spell resolves, and amasses, with no target
// chosen. The card (if still in the graveyard as the spell resolves, CR 608.2b)
// goes on top of its owner's library, then the keyword action (CR 701.47).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b04ef0bd-5e96-45d4-afa8-1f3bc8544e1a",
		Name:         "Treason of Isengard",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 2, Answers: game.AnswerValue},
		Targets: effects.TargetCardInGraveyard("up to one target instant or sorcery card from your graveyard",
			effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())).WithCount(0, 1),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.PutIntoLibrary{Card: t.ID, Depth: 1}.Apply(ctx)
					}
				}
				return nil
			}),
			effects.Amass{Subtype: "Orc", N: 2},
		),
	})
}
