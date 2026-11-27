package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akuta, Born of Ash — Legendary Creature — Spirit {2}{B}{B}, 3/2 (Saviors of
// Kamigawa, #5601):
//
//	"Haste
//	 At the beginning of your upkeep, if you have more cards in hand than each
//	 opponent, you may sacrifice a Swamp. If you do, return Akuta, Born of Ash
//	 from your graveyard to the battlefield."
//
// The upkeep ability works from the graveyard (CR 113.6) and its "if" is an
// intervening if (CR 603.4), asked as the upkeep begins and again as the
// trigger resolves. "You may sacrifice a Swamp" is a pick of zero or one Swamp
// (read through the layers, so a land that is a Swamp counts); Akuta returns
// only if one was sacrificed and Akuta is still in the graveyard.
//
// No simplification.
func init() {
	const label = "Akuta, Born of Ash — you may sacrifice a Swamp; if you do, return Akuta from your graveyard to the battlefield"
	effects.Register(effects.Spec{
		OracleID:        "29169659-309f-4e1c-86ba-1d36420ed41e",
		Name:            "Akuta, Born of Ash",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.InterveningIfYou("you have more cards in hand than each opponent", effects.MoreCardsInHandThanEachOpponent,
				effects.AtYourUpkeep(label, effects.OldB587AkutaSacrificeASwamp))),
		},
	})
}
