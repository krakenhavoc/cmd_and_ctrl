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
// The target clause is "up to two" (WithCount(0, 2)). The total is X,
// or 2X when X is 6 or more, and it is split across the targets still
// legal when the spell resolves (CR 608.2b), so a target that left in
// response takes nothing and the other takes the whole amount.
//
// DECLARED SIMPLIFICATION (weaker than printed), Fury's exactly: the
// DIVISION is made for the player, as evenly as possible in the order
// the targets were picked, the odd point to the first pick. A 5/1
// split of 6 is not offered; 3/3 is. StackItem.Distribution rides the
// stack item but nothing writes or reads it yet (#1563).
func init() {
	Register(Spec{
		OracleID:     "78301998-fd9b-4cd5-afad-dbcb43cac2a7",
		Name:         "Shatterskull Smashing",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The damage is divided as evenly as possible among the targets you pick, in the order you pick them, rather than however you choose.",
		},
		XMatters: true,
		Targets: TargetPermanent("up to two target creatures and/or planeswalkers",
			Or(Creature(), Planeswalker())).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			total := ctx.X()
			if total >= 6 {
				total *= 2
			}
			return b22DamageDividedEvenly(ctx, total)
		},
	})
}
