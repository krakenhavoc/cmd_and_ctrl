package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ransom Note — Artifact — Clue {1} (Murders at Karlov Manor Commander):
//
//	"When this artifact enters, surveil 1.
//	 {2}, Sacrifice this artifact: Choose one —
//	 • Cloak the top card of your library.
//	 • Goad target creature.
//	 • Draw a card."
//
// A modal activated ability (Cankerbloom's shape): the mode and the
// goad's target are announced with the activation and the sacrifice is
// paid up front. Cloak is the shared CloakTopOfLibrary, goad is the
// shared goadAndScheduleClear over the mode's target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49c0aac5-d650-4ada-a194-3bf77f104bc1",
		Name:         "Ransom Note",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ransom Note — surveil 1", Do(Surveil{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this artifact: Choose one — cloak the top card of your library; goad target creature; or draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Modes: ChooseOne(
				ModeDoing("Cloak the top card of your library.", nil,
					func(item *game.StackItem, ctx *Context, _ int) error {
						return CloakTopOfLibrary{Player: item.Controller}.Apply(ctx)
					}),
				ModeDoing("Goad target creature.",
					TargetCreature("target creature"),
					func(item *game.StackItem, ctx *Context, occ int) error {
						t, ok := ModeTarget(ctx, occ)
						if !ok || !onBattlefield(ctx.Game, t.ID) {
							return nil
						}
						return goadAndScheduleClear(ctx, []uuid.UUID{t.ID})
					}),
				ModeDoing("Draw a card.", nil,
					func(item *game.StackItem, ctx *Context, _ int) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
					}),
			),
		}},
	})
}
