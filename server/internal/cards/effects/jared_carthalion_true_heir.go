package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jared Carthalion, True Heir — Legendary Creature — Human Warrior {R}{G}{W}, 3/3:
//
//	"When Jared Carthalion enters, target opponent becomes the monarch. You can't become the monarch this turn.
//	 If damage would be dealt to Jared Carthalion while you're the monarch, prevent that damage and put that many +1/+1 counters on it."
//
// #2039 (ADR 0096 amendment, 2026-10-08): "You can't become the monarch
// this turn" is a stored ModCantBecomeMonarch record swept at cleanup,
// read by becomeMonarchLocked, so it holds against the combat-damage
// steal and every other route. The opponent is crowned first, then the
// bar is registered, in printed order. The prevention static is ADR 0108
// §8 (#1906) with a "while you're the monarch" condition.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "b48da54c-002f-4afb-9fdf-2a3e1cf77383",
		Name:         "Jared Carthalion, True Heir",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Jared Carthalion — target opponent becomes the monarch; you can't become the monarch this turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								if err := (BecomeTheMonarch{Player: t.ID}).Apply(ctx); err != nil {
									return err
								}
							}
						}
						return CantBecomeTheMonarchThisTurn{}.Apply(ctx)
					}),
				TargetPlayer("target opponent", Opponent())),
		},
		Replacements: []game.ReplacementEffect{
			PreventDamageDealtTo(PreventionStatic{
				To:    ToThisCreature,
				While: whileItsControllerIsTheMonarch,
				Then:  thatManyCountersOnThisBody,
				Label: "Jared Carthalion — prevent damage to it and put that many +1/+1 counters on it",
			}),
		},
	})
}

// whileItsControllerIsTheMonarch is "while you're the monarch".
func whileItsControllerIsTheMonarch(g *game.Game, src *game.Card) bool {
	return src.Controller != uuid.Nil && YoureTheMonarch(g, src.Controller)
}
