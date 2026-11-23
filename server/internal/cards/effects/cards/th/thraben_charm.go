package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thraben Charm — Instant {1}{W} (EDHREC rank 3171):
//
//	"Choose one —
//	 • Thraben Charm deals damage equal to twice the number of
//	   creatures you control to target creature.
//	 • Destroy target enchantment.
//	 • Exile any number of target players' graveyards."
//
// The go-wide deck's charm. Three modes, each with its own target
// clause: the damage is twice the caster's creature count as the
// spell resolves (the Charm is the source, so it is not combat
// damage and no creature of yours is dealing it); the enchantment
// is destroyed if it is still there; the graveyards are Stonespeaker
// Crystal's zero-to-unbounded player clause, exiled whole.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7eccd10d-f224-4467-af25-f2b768993f9f",
		Name:         "Thraben Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Thraben Charm deals damage equal to twice the number of creatures you control to target creature.",
				effects.TargetCreature("target creature")),
			effects.Mode("Destroy target enchantment.", effects.TargetPermanent("target enchantment", effects.Enchantment())),
			effects.Mode("Exile any number of target players' graveyards.",
				effects.TargetPlayer("any number of target players").WithCount(0, 0)),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				n := effects.B14CreaturesControlled(ctx.Game, ctx.Controller())
				return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 2 * n}.Apply(ctx)
			case ctx.HasMode(1):
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.DestroyTarget{Target: id}.Apply(ctx)
			case ctx.HasMode(2):
				return effects.B30ExileTargetGraveyards(ctx)
			}
			return nil
		},
	})
}
