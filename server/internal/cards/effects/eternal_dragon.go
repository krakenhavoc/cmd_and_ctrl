package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eternal Dragon — Creature — Dragon Spirit {5}{W}{W}, 5/5:
//
//	"Flying
//	 {3}{W}{W}: Return this card from your graveyard to your hand. Activate
//	 only during your upkeep.
//	 Plainscycling {2}"
//
// The return works only from the graveyard (Abzan Devotee's shape) and only
// in your own upkeep step. Plainscycling is the cycling family's Typecycling
// (CR 702.29e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "04d8615c-3883-4251-9790-1d8a4a40e142",
		Name:            "Eternal Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			{
				Label:     "{3}{W}{W}: Return this card from your graveyard to your hand. Activate only during your upkeep.",
				Purpose:   game.Purpose{Answers: game.AnswerValue},
				Cost:      ManaCost("{3}{W}{W}"),
				Zones:     []game.ZoneKind{game.ZoneGraveyard},
				Condition: DuringYourUpkeep(),
				Effect:    returnThisCardFromYourGraveyardToYourHand,
			},
			Typecycling("Plainscycling", "{2}", "a Plains card", IsLandWithSubtype("Plains")),
		},
	})
}
