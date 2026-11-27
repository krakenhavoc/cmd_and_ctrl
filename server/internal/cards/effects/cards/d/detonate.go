package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Detonate — Sorcery {X}{R} (Mirrodin, S82 OLD, #5606):
//
//	"Destroy target artifact with mana value X. It can't be regenerated.
//	 Detonate deals X damage to that artifact's controller."
//
// The target clause is "mana value X" (WithManaValueEqualsX), judged against
// the X announced and again on resolution (CR 608.2b), so only an artifact
// whose mana value is exactly X can be picked. The destruction ignores
// regeneration shields (CR 701.19c). The artifact's controller is read
// before it is destroyed, and the damage is dealt whether or not it
// died (an indestructible artifact still costs its controller X).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "daa90a75-c600-41bd-9311-ec21cf51480b",
		Name:         "Detonate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact with mana value X", effects.Artifact()).WithManaValueEqualsX(),
		XMatters:     true,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			controller := uuid.Nil
			if c, found := ctx.Game.LookupCardForEffect(id); found {
				controller = c.Controller
			}
			// The damage is the spell's next sentence, so it follows the
			// destruction as a keyed tail and is dealt whether or not the
			// artifact died (an indestructible one still costs its controller X).
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{id},
				effects.WithCarry(ctx.TailN(detonateDamage, ctx.X()), controller),
				game.DestroyOptions{CantBeRegenerated: true})
		},
	})
}

// detonateDamage is "Detonate deals X damage to that artifact's controller".
// An on-disk key: never renamed, never reused.
var detonateDamage = effects.CardTail("then/s82-old-b603-detonate-damage", func(ctx *effects.Context, o game.TailOutcome) error {
	controller := effects.TailCarry0(o)
	if controller == uuid.Nil || o.Tail.N <= 0 {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: o.Tail.N}.Apply(ctx)
})
