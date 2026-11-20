package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barrier Breach — Instant {2}{G} (Ikoria: Lair of Behemoths, #5512):
//
//	"Exile up to three target enchantments.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Each target still legal as it resolves is exiled (CR 608.2b), the
// shared exileEveryLegalCardTarget body. Cycling from hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "581425e6-3128-4f83-9efc-17028fc65078",
		Name:         "Barrier Breach",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("up to three target enchantments", Enchantment()).WithCount(0, 3),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileEveryLegalCardTarget(ctx.Game, item)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
