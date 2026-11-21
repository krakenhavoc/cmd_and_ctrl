package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "49c0aac5-d650-4ada-a194-3bf77f104bc1",
		Name:         "Ransom Note",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ransom Note — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice this artifact: Choose one — cloak the top card of your library; goad target creature; or draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
			Modes: effects.ChooseOne(
				effects.ModeDoing("Cloak the top card of your library.", nil,
					func(item *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.CloakTopOfLibrary{Player: item.Controller}.Apply(ctx)
					}),
				effects.ModeDoing("Goad target creature.",
					effects.TargetCreature("target creature"),
					func(item *game.StackItem, ctx *effects.Context, occ int) error {
						t, ok := effects.ModeTarget(ctx, occ)
						if !ok || !effects.OnBattlefield(ctx.Game, t.ID) {
							return nil
						}
						return effects.GoadAndScheduleClear(ctx, []uuid.UUID{t.ID})
					}),
				effects.ModeDoing("Draw a card.", nil,
					func(item *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
					}),
			),
		}},
	})
}
