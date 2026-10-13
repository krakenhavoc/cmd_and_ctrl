package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Return to Dust — Instant {2}{W}{W} (EDHREC rank 1576):
//
//	"Exile target artifact or enchantment. If you cast this spell
//	 during your main phase, you may exile up to one other target
//	 artifact or enchantment."
//
// Two clauses (Clauses): the required "target artifact or enchantment"
// and an "up to one other" clause whose predicate is only true during
// the caster's own main phase, so outside it the second clause has no
// legal target at all and a second pick is refused at announce rather
// than silently ignored. The step cannot change between cast and
// resolution (a step ends only with an empty stack), so the same test
// at resolution (CR 608.2b) agrees.
func init() {
	mainPhaseOnly := func(g *game.Game, caster uuid.UUID, _ game.Card) bool {
		return b14InYourMainPhase(g, caster)
	}
	Register(Spec{
		OracleID:     "3029df1d-d02a-4fed-8ab4-000a2096f823",
		Name:         "Return to Dust",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Distinct(TargetPermanent("up to one other target artifact or enchantment (your main phase only)",
				And(Or(Artifact(), Enchantment()), mainPhaseOnly)).WithCount(0, 1)),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range item.Targets {
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
