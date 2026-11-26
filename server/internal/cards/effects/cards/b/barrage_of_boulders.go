package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barrage of Boulders — Khans of Tarkir (S82 OLD b347, #5554):
//
//	Sorcery {2}{R}: Barrage of Boulders deals 1 damage to each creature you don't
//	control. Ferocious — If you control a creature with power 4 or greater,
//	creatures can't block this turn.
//
// Ferocious is read through YouControlPowerFourOrGreater (current power, layered). "Creatures can't block" covers every creature, entering later ones
// included (CR 611.2c: it changes no characteristic).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "71a756f3-cd1e-4047-90ac-eda096b90647",
		Name:         "Barrage of Boulders",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.OpponentControls()), 1); err != nil {
				return err
			}
			if !effects.FerociousYouControlPowerFourOrGreater(ctx) {
				return nil
			}
			return effects.RestrictUntilEOT{Scope: game.ScopeCreatures, Restrictions: game.CantBlock, Label: "Barrage of Boulders — creatures can't block"}.Apply(ctx)
		},
	})
}
