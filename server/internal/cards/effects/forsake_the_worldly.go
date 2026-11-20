package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forsake the Worldly — Instant {2}{W}
// (Amonkhet, #5535):
//
//	"Exile target artifact or enchantment.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Revoke Existence, plus cycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f60fb81a-969a-476c-a227-5231bbed4ad4",
		Name:         "Forsake the Worldly",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileFirstTarget(ctx.Game, item)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
