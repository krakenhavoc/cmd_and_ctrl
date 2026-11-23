package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alquist Proft, Master Sleuth — Legendary Creature — Human Detective
// {1}{W}{U}, 3/3 (Murders at Karlov Manor, #4469):
//
//	"Vigilance
//	 When Alquist Proft enters, investigate. (Create a Clue token. It's
//	 an artifact with '{2}, Sacrifice this token: Draw a card.')
//	 {X}{W}{U}{U}, {T}, Sacrifice a Clue: You draw X cards and gain X
//	 life."
//
// The entry is Novice Inspector's Clue. The activated ability's cost is
// the X mana, the tap and one Clue you control (Tamiyo's Journal's
// sacrifice clause with a count of one), so sacrificing it is
// sacrificing a Clue for every card that watches that. The X announced
// (CR 602.2b) is the number of cards drawn and then the life gained, the
// gain waiting for the draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dd79a994-7dad-49d4-8d8f-62dbedde4a38",
		Name:            "Alquist Proft, Master Sleuth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		XMatters:        true,
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Alquist Proft, Master Sleuth — investigate", effects.Do(effects.Investigate{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{W}{U}{U}, {T}, Sacrifice a Clue: You draw X cards and gain X life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{X}{W}{U}{U}"), effects.TapCost(), effects.SacrificeN(1, "a Clue", effects.OfSubtype("Clue"))),
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{N: ctx.X()}.Apply(ctx) }),
				effects.Instruction(func(ctx *effects.Context) error { return effects.GainLife{Amount: ctx.X()}.Apply(ctx) }),
			),
		}},
	})
}
