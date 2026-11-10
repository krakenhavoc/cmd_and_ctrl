package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hammerhead Tyrant — Creature — Dragon {4}{U}{U}, 6/6 (Tarkir: Dragonstorm
// Commander):
//
//	"Flying
//	 Whenever you cast a spell, return up to one target nonland permanent an
//	 opponent controls with mana value less than or equal to that spell's mana
//	 value to its owner's hand."
//
// "That spell's mana value" is the cast event's Amount, the mana value the
// spell had as it was cast with its X chosen (the ruling; CR 202.3e), carried
// on the trigger and judged in the target walk (ManaValueAtMostTheEventAmount),
// both as the ability goes on the stack and again on resolution (CR 608.2b).
// A permanent with {X} in its cost has X = 0. The clause is "up to one", so
// the trigger goes on the stack even with nothing to return, and a land is
// never a target.
//
// Seam: game.Event.Amount on EventCast (the spell's mana value).
func init() {
	Register(Spec{
		OracleID:        "7830f600-c7c7-48f8-8020-3493c5745eed",
		Name:            "Hammerhead Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverYouCast(nil,
				"Hammerhead Tyrant — return up to one target nonland permanent an opponent controls with mana value less than or equal to that spell's mana value to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return BounceToHand{Target: t.ID}.Apply(ctx)
						}
					}
					return nil
				}),
				ManaValueAtMostTheEventAmount(
					TargetPermanent("up to one target nonland permanent an opponent controls with mana value less than or equal to that spell's mana value",
						Nonland(), OpponentControls()).WithCount(0, 1))),
		},
	})
}
