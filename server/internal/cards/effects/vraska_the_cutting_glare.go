package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vraska, the Cutting Glare — Legendary Creature — Gorgon Assassin
// {B}{B}{G}, 4/4:
//
//	"Deathtouch
//	 When Vraska enters, if you control six or more lands, destroy
//	 target permanent an opponent controls. They create a Treasure
//	 token."
//
// An intervening if (CR 603.4): the trigger does not go on the stack
// below six lands, and the count is checked again as it resolves. "They"
// is the destroyed permanent's controller, read before it is destroyed,
// and the Treasure is theirs even if the permanent was indestructible.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "580de511-863c-4b7a-9d2e-fcdf6e12d54b",
		Name:            "Vraska, the Cutting Glare",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{Targeting(
			On(game.EventETB, rfCreatureFEntersWithSixLands,
				"Vraska, the Cutting Glare — destroy target permanent an opponent controls; they create a Treasure",
				rfCreatureFDestroyThenTreasure),
			TargetPermanent("target permanent an opponent controls", OpponentControls()),
		)},
	})
}
