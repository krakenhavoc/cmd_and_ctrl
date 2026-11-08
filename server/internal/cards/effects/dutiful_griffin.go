package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dutiful Griffin — Creature — Griffin {3}{W}{W}, 4/4 (Wilds of Eldraine,
// #4471):
//
//	"Flying
//	 {2}{W}, Sacrifice two enchantments: Return this card from your
//	 graveyard to your hand."
//
// Bestial Bloodline's graveyard return (returnThisCardFromYourGraveyardToYourHand,
// working only while the card is in the graveyard) with Breya, Etherium
// Shaper's two-permanent sacrifice cost (SacrificeN) over enchantments.
// It targets nothing, so it declares what it answers (value).
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
