package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Supply Drop — Artifact {3} (The Brothers' War, S82):
//
//	"Flash
//	 When this artifact enters, target creature you control gets +2/+2 until
//	 end of turn.
//	 {4}, {T}, Sacrifice this artifact: Draw a card."
//
// The enters trigger targets as it goes on the stack (CR 603.3d); with no
// creature of yours it is removed. The sacrifice-and-tap ability is Bear
// Trap's cost shape.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9594b8e0-a176-447a-a52e-e1a77175b748",
		Name:            "Supply Drop",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Supply Drop — target creature you control gets +2/+2 until end of turn",
				Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
					return BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Supply Drop — +2/+2"}.Apply(ctx)
				}))),
				TargetCreature("target creature you control", YouControl())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{4}"), TapCost(), SacrificeThis()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
