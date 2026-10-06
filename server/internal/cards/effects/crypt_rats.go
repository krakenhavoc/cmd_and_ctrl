package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crypt Rats — Creature — Rat {2}{B}, 1/1:
//
//	"{X}: This creature deals X damage to each creature and each
//	 player. Spend only black mana on X."
//
// Pestilence with the mana up front: one activation, any size, and the
// Rats are a creature, so they take their share. The body is
// Pestilence's (b23DamageEachCreatureAndEachPlayer) with X for 1, one
// damage instance, every player hit, the controller included.
//
// "Spend only black mana on X" is a cost-side spend restriction
// (#1600, SpendOnlyOnX): the X mana is folded into X black symbols at
// payment, so the pool, the auto-tapper, the auto-tap preview and the
// bot's enumerator all refuse a red or colourless X — and all agree on
// the largest X the board can pay in black. Under Chromatic Orrery any
// mana may pay it (CR 609.4b: the restriction is about how the cost is
// paid, and the Orrery changes exactly that).
//
// XMatters: an X of 0 deals no damage, so the enumerator does not offer
// it as a move.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "104095ed-55e3-408e-bf70-4fe06bb16d2f",
		Name:         "Crypt Rats",
		Completeness: CompletenessFull,
		XMatters:     true,
		Activated: []ActivatedAbility{{
			Label:   "{X}: This creature deals X damage to each creature and each player. Spend only black mana on X.",
			Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true}},
			Cost:    Plus(ManaCost("{X}"), SpendOnlyOnX("B")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return b23DamageEachCreatureAndEachPlayer(ctx, ctx.X())
			},
		}},
	})
}
