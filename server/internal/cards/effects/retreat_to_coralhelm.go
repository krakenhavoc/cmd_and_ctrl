package effects

import (
	"github.com/google/uuid"

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
	t := Landfall("Retreat to Coralhelm — choose one: you may tap or untap target creature, or scry 1",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("You may tap or untap target creature.", TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				tgt, ok := ModeTarget(ctx, occ)
				if !ok || tgt.Kind != game.TargetCard {
					return nil
				}
				return PickOption{
					Question: "Retreat to Coralhelm — tap or untap that creature",
					Options: []game.ChoiceOption{
						{Label: "Tap it"},
						{Label: "Untap it"},
						{Label: "Do neither"},
					},
					ThenKey: tapOrUntapAnswer,
					Carry:   []uuid.UUID{tgt.ID},
				}.Apply(ctx)
			}),
		ModeWithPurpose(ModeDoing("Scry 1.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error { return Scry{N: 1}.Apply(ctx) }),
			game.Purpose{Scry: 1}),
	)
	Register(Spec{
		OracleID:     "c450f6b7-bb2a-4b8c-aec7-d27555137b63",
		Name:         "Retreat to Coralhelm",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
