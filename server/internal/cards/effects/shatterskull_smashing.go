package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shatterskull Smashing — Sorcery {X}{R}{R}, the front face of a
// modal double-faced card (Edea steal-and-sac deck, #1565):
//
//	"Shatterskull Smashing deals X damage divided as you choose among
//	 up to two target creatures and/or planeswalkers. If X is 6 or
//	 more, Shatterskull Smashing deals twice X damage divided as you
//	 choose among them instead."
//
// The back face, Shatterskull, the Hammer Pass, is a row of the MDFC
// land cycle in mdfc_lands.go (pay 3 life or it enters tapped), under
// "<oracle_id>#1". This file is the front face, under the bare ID.
//
// The target clause is "up to two" (WithCount(0, 2)) and it divides
// X — or twice X once X is 6 or more — as the caster chooses (#1563,
// CR 601.2d): DivideXDoublingFrom(6) is the doubling, read against the
// X announced at CR 601.2b, so the amount the division has to add up
// to is fixed at announce. Each target is assigned at least 1, so X=1
// can take only one target and X=0 none. A target that left in
// response takes nothing and its share is lost (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "78301998-fd9b-4cd5-afad-dbcb43cac2a7",
		Name:         "Shatterskull Smashing",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets: TargetPermanent("up to two target creatures and/or planeswalkers",
			Or(Creature(), Planeswalker())).WithCount(0, 2).Dividing(DivideXDoublingFrom(6)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DealDividedDamage(ctx)
		},
	})
}
