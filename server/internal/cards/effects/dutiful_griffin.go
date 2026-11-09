package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "f91ff377-a997-4f24-926c-76edcbf91a6c",
		Name:            "Dutiful Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}, Sacrifice two enchantments: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{W}"), SacrificeN(2, "two enchantments", Enchantment())),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
