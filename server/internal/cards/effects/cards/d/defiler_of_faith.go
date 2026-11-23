package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defiler of Faith — Creature — Phyrexian Human {3}{W}{W}, 5/5 (Dominaria United,
// #5659):
//
//	"Vigilance
//	 As an additional cost to cast white permanent spells, you may pay 2
//	 life. Those spells cost {W} less to cast if you paid life this way.
//	 This effect reduces only the amount of white mana you pay.
//	 Whenever you cast a white permanent spell, create a 1/1 white Soldier
//	 creature token."
//
// The offer and its discount are the cycle's linked pair
// (pay_life_for_discount.go): each white permanent spell its controller
// casts may be announced with "pay 2 life", and one that is costs {W}
// less, never less generic mana. The token is
// made as the trigger resolves, before the spell that caused it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:             "88ac2d09-7993-4bce-8aae-b2683c5c25a5",
		Name:                 "Defiler of Faith",
		Completeness:         effects.CompletenessFull,
		PrintedKeywords:      []string{"vigilance"},
		GrantedOptionalCosts: []game.GrantedOptionalCost{effects.MayPayLifeToCast("Defiler of Faith", "W", 2)},
		CostModifiers:        []game.CostModifier{effects.CostsColorLessIfLifePaid("W")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCastAPermanentSpellOfColor("W", "Defiler of Faith — create a 1/1 white Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1})),
		},
	})
}
