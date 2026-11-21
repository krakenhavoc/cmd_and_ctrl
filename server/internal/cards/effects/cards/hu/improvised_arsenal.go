package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Improvised Arsenal — {1}{R} Artifact — Equipment (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Equipped creature gets +1/+0 for each artifact you control.
//	 {4}{R}: Create a token that's a copy of this Equipment.
//	 Equip {R}"
//
// The bonus is Cranial Plating's (artifactsControlledBy, read live in
// layer 7c), and it counts the Arsenal itself and every token copy of it.
// The copy is Bloodforged Battle-Axe's token copy of the source, made
// unattached; once the ability is on the stack it no longer needs the
// Arsenal (CR 113.7a), so an Arsenal destroyed in response is copied from
// its last-known copiable values (CR 707.2, CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f02aa99-481c-4089-8f7c-cfe7eaa7dda0",
		Name:         "Improvised Arsenal",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 0, effects.ArtifactsControlledBy),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{4}{R}: Create a token that's a copy of this Equipment.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{4}{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			effects.EquipAbility("{R}"),
		},
	})
}
