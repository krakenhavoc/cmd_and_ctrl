package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Retreat to Coralhelm — Enchantment {2}{U} (Duskmourn Commander, #5368):
//
//	"Landfall — Whenever a land you control enters, choose one —
//	 • You may tap or untap target creature.
//	 • Scry 1."
//
// A modal landfall trigger (CR 700.2): the bullet is chosen as the trigger goes
// on the stack, and the first is offered only with a creature to target. "You
// may tap or untap" is the answer to a three-way pick at resolution (tap it,
// untap it, or neither), the keyed pick Merrow Reejerey and Gandalf the Grey use;
// a target that left in response takes the clause with it (CR 608.2b). The
// second bullet is a plain scry.
//
// No simplification.
func init() {
	t := effects.Landfall("Retreat to Coralhelm — choose one: you may tap or untap target creature, or scry 1",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("You may tap or untap target creature.", effects.TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				tgt, ok := effects.ModeTarget(ctx, occ)
				if !ok || tgt.Kind != game.TargetCard {
					return nil
				}
				return effects.PickOption{
					Question: "Retreat to Coralhelm — tap or untap that creature",
					Options: []game.ChoiceOption{
						{Label: "Tap it"},
						{Label: "Untap it"},
						{Label: "Do neither"},
					},
					ThenKey: effects.TapOrUntapAnswer,
					Carry:   []uuid.UUID{tgt.ID},
				}.Apply(ctx)
			}),
		effects.ModeWithPurpose(effects.ModeDoing("Scry 1.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error { return effects.Scry{N: 1}.Apply(ctx) }),
			game.Purpose{Scry: 1}),
	)
	effects.Register(effects.Spec{
		OracleID:     "c450f6b7-bb2a-4b8c-aec7-d27555137b63",
		Name:         "Retreat to Coralhelm",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
