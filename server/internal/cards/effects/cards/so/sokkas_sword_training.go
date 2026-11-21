package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sokka's Sword Training — Instant — Lesson {1}{W} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Target creature gets +2/+2 until end of turn.
//	 Create a Clue token. (It's an artifact with "{2}, Sacrifice this
//	 token: Draw a card.")"
//
// Cunning Maneuver's shape at +2/+2. Creating the Clue is not
// investigating, as the ruling has it. With its only target gone the
// spell does not resolve (CR 608.2b), so there is no Clue either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19963c90-e7b2-4935-adcd-d41f0faf401a",
		Name:         "Sokka's Sword Training",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Sokka's Sword Training — +2/+2 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.CreateToken{Template: effects.ClueToken(), N: 1}.Apply(ctx)
		},
	})
}
