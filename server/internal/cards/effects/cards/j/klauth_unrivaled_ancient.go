package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Klauth, Unrivaled Ancient — Legendary Creature — Dragon {5}{R}{G}, 4/4
// (Adventures in the Forgotten Realms Commander):
//
//	"Flying, haste
//	 Whenever Klauth attacks, add X mana in any combination of colors, where
//	 X is the total power of attacking creatures. Spend this mana only to
//	 cast spells. Until end of turn, you don't lose this mana as steps and
//	 phases end."
//
// X is counted as the ability resolves (CR 608.2h), over every attacking
// creature at its current power, Klauth included; a total below 1 adds
// nothing. The mana is one slot per point, each any colour chosen as it is
// added, restricted to casting spells (CR 106.6) and kept through the
// turn's steps and phases (KeepManaUntilEndOfTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8cefa5d6-5151-4e75-8b72-f210a4c751ea",
		Name:            "Klauth, Unrivaled Ancient",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Klauth, Unrivaled Ancient — add X mana in any combination of colors, X the total power of attacking creatures",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := effects.AfrB18TotalPowerOfAttackingCreatures(g)
					if x <= 0 {
						return nil
					}
					return effects.AddMana{
						Amount:       effects.ManaN(x, game.AllColors...),
						Restrictions: []string{effects.ManaRestrictCast},
						Riders:       []game.ManaSpendRider{effects.KeepManaUntilEndOfTurn()},
					}.Apply(ctx)
				}),
		},
	})
}
