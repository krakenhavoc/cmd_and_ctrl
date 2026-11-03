package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Squee, Goblin Nabob — Legendary Creature — Goblin {2}{R}, 1/1
// (Secrets of Strixhaven Commander, #4834):
//
//	"At the beginning of your upkeep, you may return this card from your
//	 graveyard to your hand."
//
// A trigger that works only from the graveyard (CR 113.6, InGraveyard),
// where "your" is the card's owner (CR 108.4). The "you may" is the
// CR 603.5 prompt, and the card may have left the graveyard before the
// ability resolves, when nothing happens (returnThisCardFromYourGraveyardToHand).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ef760f4-bbec-4735-8d60-fc24d973452a",
		Name:         "Squee, Goblin Nabob",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InGraveyard(Optional(AtYourUpkeep("Squee, Goblin Nabob — return this card from your graveyard to your hand",
				returnThisCardFromYourGraveyardToHand), "Squee, Goblin Nabob — return this card from your graveyard to your hand?")),
		},
	})
}
