package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Petravark — Creature — Nightmare Beast {3}{R}, 2/2 (Torment, #5612):
//
//	"When this creature enters, exile target land.
//	 When this creature leaves the battlefield, return the exiled card to
//	 the battlefield under its owner's control."
//
// Faceless Butcher's linked pair (CR 607.2a) aimed at a land: the leave
// trigger returns the land the entry trigger exiled, if it is still in
// exile, under its owner's control. If Petravark leaves before its entry
// trigger resolves, the leave trigger finds nothing and the land is then
// exiled for good, as the Oracle rulings say. Petravark is not a land, so
// it never targets itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aedf022c-6065-42cb-85bb-225ae1650a5c",
		Name:         "Petravark",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(petravarkExileLabel, exileFirstLegalCardTarget),
				TargetPermanent("target land", Land())),
			On(game.EventLTB, Self, "Petravark — return the exiled card to the battlefield under its owner's control",
				b41ReturnCardsExiledWithToTheBattlefield(petravarkExileLabel)),
		},
	})
}

// petravarkExileLabel is the entry trigger's key, which is what the leave
// trigger reads the linked exile by.
const petravarkExileLabel = "Petravark — exile target land"
