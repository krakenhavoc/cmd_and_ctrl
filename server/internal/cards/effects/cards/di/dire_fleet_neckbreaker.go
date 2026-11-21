package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dire Fleet Neckbreaker — Creature — Orc Pirate {2}{B}{R}, 3/2 (Rivals
// of Ixalan, #5531):
//
//	"Attacking Pirates you control get +2/+0."
//
// Instigator Gang's attacking-creatures pump narrowed to Pirates, the way
// Blight Mound narrows it to Pests. "Pirates" is a word of the card's text
// (CR 612.2), and the Neckbreaker pumps itself when it attacks.
//
// No simplification.
func init() {
	pirates := effects.AttackingCreaturesYouControlGetPower(2)
	pirates.Label = "Attacking Pirates you control get +2/+0."
	attacking := pirates.AppliesTo
	pirates.AppliesTo = func(target *game.Card, g *game.Game, source *game.Card) bool {
		return attacking(target, g, source) && target.HasSubtype(source.TextSubtype("Pirate"))
	}
	effects.Register(effects.Spec{
		OracleID:     "bcf34a1c-162a-400c-b1db-9d1bb62bf7e9",
		Name:         "Dire Fleet Neckbreaker",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{pirates},
	})
}
