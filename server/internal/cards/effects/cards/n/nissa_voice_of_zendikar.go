package cardsn

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Nissa, Voice of Zendikar — Legendary Planeswalker — Nissa {1}{G}{G},
// loyalty 3 (Oath of the Gatewatch, S82 OLD #5545):
//
//	"+1: Create a 0/1 green Plant creature token.
//	 −2: Put a +1/+1 counter on each creature you control.
//	 −7: You gain X life and draw X cards, where X is the number of lands you
//	 control."
//
// −2 is Basri's Solidarity's set (the creatures you control as it resolves).
// −7 counts the lands once, as it resolves, and gains and draws that many.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b0f36bf7-bdde-4576-9eca-ec02d41326ab",
		Name:            "Nissa, Voice of Zendikar",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Create a 0/1 green Plant creature token.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 green Plant"), N: 1}),
			},
			{
				Label:  "−2: Put a +1/+1 counter on each creature you control.",
				Cost:   effects.LoyaltyCost(-2),
				Effect: effects.Do(effects.Instruction(effects.PutAPlusOneCounterOnEachCreatureYouControlStep)),
			},
			{
				Label:  "−7: You gain X life and draw X cards, where X is the number of lands you control.",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.Do(effects.Instruction(nissaVoiceOfZendikarUltimate)),
			},
		},
	})
}

func nissaVoiceOfZendikarUltimate(ctx *effects.Context) error {
	x := len(effects.LandsControlledByPlayer(ctx.Game, ctx.Controller()))
	if err := (effects.GainLife{Amount: x}).Apply(ctx); err != nil {
		return err
	}
	return effects.DrawCards{N: x}.Apply(ctx)
}
