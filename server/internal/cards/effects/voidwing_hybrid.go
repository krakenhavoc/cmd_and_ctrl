package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voidwing Hybrid — Creature — Phyrexian Bat {U}{B}, 2/1 (Phyrexia: All
// Will Be One, #5653):
//
//	"Flying
//	 Toxic 1 (Players dealt combat damage by this creature also get a
//	 poison counter.)
//	 When you proliferate, return this card from your graveyard to your
//	 hand."
//
// A graveyard ability (InGraveyard: it works there and nowhere else,
// "you" being the card's owner) off WheneverYouProliferate. It is not
// optional, and it returns the object that triggered: a Hybrid that left
// the graveyard in response stays where it went (CR 400.7). Toxic is the
// engine's damage tail (ADR 0056).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0dab592c-a230-408d-a383-044e561dbe5d",
		Name:            "Voidwing Hybrid",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "toxic 1"},
		Triggered: []game.TriggeredAbility{
			InGraveyard(WheneverYouProliferate("Voidwing Hybrid — return this card from your graveyard to your hand",
				returnThisCardObjectFromYourGraveyardToHand)),
		},
	})
}

// returnThisCardObjectFromYourGraveyardToHand is "return this card from
// your graveyard to your hand" for a trigger that fired from the
// graveyard: only the object that triggered comes back, not the same card
// after it left and was put there again (CR 400.7).
func returnThisCardObjectFromYourGraveyardToHand(g *game.Game, item *game.StackItem) error {
	if !cardObjectStillInGraveyard(g, item) {
		return nil
	}
	return returnThisCardFromYourGraveyardToHand(g, item)
}
