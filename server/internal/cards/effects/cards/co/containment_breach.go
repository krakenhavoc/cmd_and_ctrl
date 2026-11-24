package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Containment Breach — Sorcery — Lesson {2}{G} (Strixhaven: School of Mages,
// #5502):
//
//	"Destroy target artifact or enchantment. If its mana value is 2 or less,
//	 create a 1/1 black and green Pest creature token with "When this token
//	 dies, you gain 1 life.""
//
// Root Out's target and destroy. The mana value is read from the target as
// the spell resolves, before the destruction (CR 608.2h), so a permanent that
// survives (indestructible) or is replaced still counts, and the Pest is the
// caster's. Being a Lesson is its type line, which learn reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "043dd2e5-21e4-49b9-90b2-831e587b2238",
		Name:         "Containment Breach",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			mv := effects.FirstTargetManaValue(ctx.Game, item)
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			if mv > 2 {
				return nil
			}
			return effects.CreateToken{Template: effects.PestToken(), N: 1}.Apply(ctx)
		},
	})
}
