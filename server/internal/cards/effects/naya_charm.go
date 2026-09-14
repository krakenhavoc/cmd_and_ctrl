package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Naya Charm — Instant {R}{G}{W} (EDHREC rank 3119):
//
//	"Choose one —
//	 • Naya Charm deals 3 damage to target creature.
//	 • Return target card from a graveyard to its owner's hand.
//	 • Tap all creatures target player controls."
//
// Three modes, each with its own target clause: a bolt at a
// creature, a Regrowth from ANY graveyard (the card goes to its
// owner's hand, whoever that is), and a one-sided Sleep over the
// chosen player's creatures (b29TapAllCreaturesControlledBy).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "20cc0f3e-de69-4e5f-88b3-ae7e6c6b5996",
		Name:         "Naya Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Naya Charm deals 3 damage to target creature.", TargetCreature("target creature")),
			Mode("Return target card from a graveyard to its owner's hand.", TargetCardInGraveyard("target card from a graveyard")),
			Mode("Tap all creatures target player controls.", TargetPlayer("target player")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				switch {
				case ctx.HasMode(0) && t.Kind == game.TargetCard:
					return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}.Apply(ctx)
				case ctx.HasMode(1) && t.Kind == game.TargetCard:
					return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
				case ctx.HasMode(2) && t.Kind == game.TargetPlayer:
					return b29TapAllCreaturesControlledBy(ctx, t.ID)
				}
			}
			return nil
		},
	})
}
