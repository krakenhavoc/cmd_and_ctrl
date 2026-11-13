package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Break Asunder — Sorcery {2}{G}{G} (Scourge):
//
//	"Destroy target artifact or enchantment.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Naturalize at sorcery speed, and cycling as the keyword's own hand
// ability (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d2c53737-c265-46e3-a779-52c6b4f82d7d",
		Name:         "Break Asunder",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyFirstLegalCardTarget(ctx.Game, item)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
