package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ajani Goldmane — Legendary Planeswalker — Ajani {2}{W}{W}, loyalty 4
// (Secret Lair Drop, a reversible card whose two faces are the same card):
//
//	"+1: You gain 2 life.
//	 −1: Put a +1/+1 counter on each creature you control. Those creatures gain
//	 vigilance until end of turn.
//	 −6: Create a white Avatar creature token. It has "This token's power and
//	 toughness are each equal to your life total.""
//
// −1's "those creatures" are the ones that got a counter: the creatures you
// control as it resolves, all of which get one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dfb5f660-fd8b-4b7b-934c-6a71cd182f15",
		Name:            "Ajani Goldmane",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: You gain 2 life.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.Do(effects.GainLife{Amount: 2}),
			},
			{
				Label: "−1: Put a +1/+1 counter on each creature you control. Those creatures gain vigilance until end of turn.",
				Cost:  effects.LoyaltyCost(-1),
				Effect: effects.Do(effects.Instruction(effects.PutAPlusOneCounterOnEachCreatureYouControlStep),
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"vigilance"}, Label: "Ajani Goldmane — vigilance"}.Apply(ctx)
					})),
			},
			{
				Label:   "−6: Create a white Avatar creature token. It has \"This token's power and toughness are each equal to your life total.\"",
				Cost:    effects.LoyaltyCost(-6),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.OldB007AvatarToken(), N: 1}),
			},
		},
	})
}
