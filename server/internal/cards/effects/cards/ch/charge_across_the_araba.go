package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charge Across the Araba — Instant — Arcane {4}{W} (Saviors of Kamigawa,
// #5601):
//
//	"Sweep — Return any number of Plains you control to their owner's hand.
//	 Creatures you control get +1/+1 until end of turn for each Plains
//	 returned this way."
//
// The Plains are picked as the spell resolves (OldB591AskSweep) and returned
// together; the creatures you control after that get +N/+N, where N is how
// many reached a hand (CR 611.2c: the set is fixed now).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd76e1a4-d6ff-419a-a881-1b77eb6ae581",
		Name:         "Charge Across the Araba",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OldB591AskSweep(ctx, "Plains", "Charge Across the Araba — return any number of Plains you control to their owner's hand", chargeAcrossTheArabaPick)
		},
	})
}

// chargeAcrossTheArabaPick returns the Plains, then pumps. An on-disk key.
var chargeAcrossTheArabaPick = effects.AskedPick("choose-cards/charge-across-the-araba-plains", func(ctx *effects.Context, r game.CardsChosen) error {
	n, err := effects.OldB591ReturnSwept(ctx, r.Picked)
	if err != nil || n <= 0 {
		return err
	}
	return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: n, Toughness: n, Label: "Charge Across the Araba — +1/+1 for each Plains returned"}.Apply(ctx)
})
