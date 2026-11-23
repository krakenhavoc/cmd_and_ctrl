package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "104095ed-55e3-408e-bf70-4fe06bb16d2f",
		Name:         "Crypt Rats",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}: This creature deals X damage to each creature and each player. Spend only black mana on X.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true}},
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.SpendOnlyOnX("B")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.B23DamageEachCreatureAndEachPlayer(ctx, ctx.X())
			},
		}},
	})
}
