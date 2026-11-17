package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wipe Clean — Instant {1}{W} (Scourge):
//
//	"Exile target enchantment.
//	 Cycling {3} ({3}, Discard this card: Draw a card.)"
//
// The shared exile of the first legal card target, over enchantments, and
// cycling as the keyword's own hand ability (CR 702.29a). An enchantment
// that left in response leaves nothing to exile (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5cbc4f3-4e73-405f-9daf-e5b7c73fa6ee",
		Name:         "Wipe Clean",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target enchantment", Enchantment()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
		Activated: []ActivatedAbility{Cycling("{3}")},
	})
}
