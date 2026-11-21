package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "20cc0f3e-de69-4e5f-88b3-ae7e6c6b5996",
		Name:         "Naya Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Naya Charm deals 3 damage to target creature.", effects.TargetCreature("target creature")), effects.ForTargets(effects.DamageToTarget(0, 3))),
			effects.Mode("Return target card from a graveyard to its owner's hand.", effects.TargetCardInGraveyard("target card from a graveyard")),
			effects.Mode("Tap all creatures target player controls.", effects.TargetPlayer("target player")),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				switch {
				case ctx.HasMode(0) && t.Kind == game.TargetCard:
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}.Apply(ctx)
				case ctx.HasMode(1) && t.Kind == game.TargetCard:
					return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
				case ctx.HasMode(2) && t.Kind == game.TargetPlayer:
					return effects.B29TapAllCreaturesControlledBy(ctx, t.ID)
				}
			}
			return nil
		},
	})
}
