package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dutiful Griffin — Creature — Griffin {3}{W}{W}, 3/3 (Wilds of Eldraine,
// #4471):
//
//	"Flying
//	 {2}{W}, Sacrifice two enchantments: Return this card from your
//	 graveyard to your hand."
//
// Blessed Ghoul's graveyard return with Breya's two-permanent sacrifice
// in the cost: the ability is activated from the graveyard, and the two
// enchantments (the Griffin is not on the battlefield, so it is never one
// of them) are chosen and sacrificed as the cost is paid (CR 601.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f91ff377-a997-4f24-926c-76edcbf91a6c",
		Name:            "Dutiful Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, Sacrifice two enchantments: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.SacrificeN(2, "two enchantments", effects.Enchantment())),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
