package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moonmist — Instant {1}{G}:
//
//	"Transform all Humans. Prevent all combat damage that would be dealt
//	 this turn by creatures other than Werewolves and Wolves. (Only
//	 double-faced cards can be transformed.)"
//
// Every Human permanent transforms as the spell resolves; one that
// can't transform is left alone (the reminder text; the ruling: "any
// double-faced Human", not just Werewolves). Then #2026's negation over
// two subtypes, read as each creature would deal combat damage (CR
// 609.7b; the rulings: "Whether or not a creature is a Werewolf or a
// Wolf is checked only as combat damage is dealt", and creatures that
// weren't on the battlefield as Moonmist resolved are caught).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "2c306e87-e3c8-4066-9be5-570e2f2d6bad",
		Name:         "Moonmist",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			// The set of Humans is read once, before any of them turns
			// over (CR 608.2h), so a Human whose back face is not a Human
			// is not turned back by its own transformation.
			var humans []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.HasSubtype("Human") {
					humans = append(humans, c.InstanceID)
				}
			}
			for _, id := range humans {
				if err := ctx.Game.TransformPermanentForEffect(id); err != nil && !errors.Is(err, game.ErrCardNotFound) {
					return err
				}
			}
			return combatShieldAgainstCreatures(exceptSubtypes("Werewolf", "Wolf")).Apply(ctx)
		},
	})
}
