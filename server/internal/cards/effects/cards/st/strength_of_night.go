package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strength of Night — Instant {2}{G} (Apocalypse, #5614):
//
//	"Kicker {B} (You may pay an additional {B} as you cast this spell.)
//	 Creatures you control get +1/+1 until end of turn. If this spell was
//	 kicked, Zombie creatures you control get an additional +2/+2 until end
//	 of turn."
//
// The creatures you control as it resolves are pumped +1/+1, and when kicked
// the Zombies among them get a second +2/+2 (CR 611.2c): the set is fixed when
// the effect begins, so the type is read then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "0045cf16-86dd-4417-ae36-88ca63b30c26",
		Name:          "Strength of Night",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerPump},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{B}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
				Label: "Strength of Night — creatures you control get +1/+1",
			}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasKicked() {
				return nil
			}
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype(ctx.TextSubtype("Zombie"))),
				Power: 2, Toughness: 2,
				Label: "Strength of Night — kicked: Zombie creatures you control get an additional +2/+2",
			}.Apply(ctx)
		},
	})
}
