package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cruel Revival — Instant {4}{B} (Onslaught):
//
//	"Destroy target non-Zombie creature. It can't be regenerated.
//	 Return up to one target Zombie card from your graveyard to your
//	 hand."
//
// Two target clauses: slot 0 is the creature, slot 1 the optional
// Zombie card, which the caster may leave empty. Each is checked again
// as the spell resolves (CR 608.2b), so a creature that left in response
// still lets the Zombie come back, and the other way round. The
// destruction ignores regeneration shields (CR 701.19c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "836d074a-8888-42a0-8dce-11ffce00d2f5",
		Name:         "Cruel Revival",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target non-Zombie creature", Not(OfCreatureType("Zombie"))),
			TargetCardInGraveyard("up to one target Zombie card from your graveyard", YouOwn(), HasSubtype("Zombie")).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if victim, ok := ctx.ClauseTarget(0); ok && victim.Kind == game.TargetCard {
				if err := (DestroyTarget{Target: victim.ID, CantBeRegenerated: true}).Apply(ctx); err != nil {
					return err
				}
			}
			// The return is not about the destroyed creature: its target
			// is read fresh here, after the destruction.
			if zombie, legal := ctx.ClauseTarget(1); legal && zombie.Kind == game.TargetCard {
				return ReturnFromGraveyard{Target: zombie.ID, Dest: game.ZoneHand}.Apply(ctx)
			}
			return nil
		},
	})
}
