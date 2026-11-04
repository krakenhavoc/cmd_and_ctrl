package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4f02aa99-481c-4089-8f7c-cfe7eaa7dda0",
		Name:         "Improvised Arsenal",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 0, artifactsControlledBy),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{4}{R}: Create a token that's a copy of this Equipment.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    ManaCost("{4}{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(NewContext(g, item))
				},
			},
			EquipAbility("{R}"),
		},
	})
}
