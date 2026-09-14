package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7eccd10d-f224-4467-af25-f2b768993f9f",
		Name:         "Thraben Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Thraben Charm deals damage equal to twice the number of creatures you control to target creature.",
				TargetCreature("target creature")),
			Mode("Destroy target enchantment.", TargetPermanent("target enchantment", Enchantment())),
			Mode("Exile any number of target players' graveyards.",
				TargetPlayer("any number of target players").WithCount(0, 0)),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				n := b14CreaturesControlled(ctx.Game, ctx.Controller())
				return DealDamage{Source: ctx.Source(), Target: id, Amount: 2 * n}.Apply(ctx)
			case ctx.HasMode(1):
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return DestroyTarget{Target: id}.Apply(ctx)
			case ctx.HasMode(2):
				return b30ExileTargetGraveyards(ctx)
			}
			return nil
		},
	})
}
