package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cunning Maneuver — Instant {1}{R} (Avatar: The Last Airbender, #4459):
//
//	"Target creature gets +3/+1 until end of turn.
//	 Create a Clue token."
//
// Creating the Clue is not investigating (the ruling). With its only
// target gone the spell doesn't resolve, so no Clue (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96335212-ec57-4beb-bab1-11c8eaa62119",
		Name:         "Cunning Maneuver",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 1)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{
					Target: t.ID,
					Power:  3, Toughness: 1,
					Label: "Cunning Maneuver — +3/+1 until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.CreateToken{Template: effects.ClueToken(), N: 1}.Apply(ctx)
		},
	})
}
