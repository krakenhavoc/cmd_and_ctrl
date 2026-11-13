package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aphetto Vulture — Creature — Zombie Bird {4}{B}{B}, 3/2 (Onslaught):
//
//	"Flying
//	 When this creature dies, you may put target Zombie card from your
//	 graveyard on top of your library."
//
// Golgari Thug's dies trigger as a "you may" (Mortuary Mire's shape),
// over Zombie cards. The trigger goes on the stack with the Vulture
// already in the graveyard, so it may target itself. With no Zombie card
// there it is removed for want of a target (CR 603.3d), and a card that
// leaves in response makes it do nothing (CR 608.2b). The put is the
// shared tuck, so a commander card gets the CR 903.9 offer on the way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aeba5c63-aa80-4f31-90fb-f404c9a3057d",
		Name:            "Aphetto Vulture",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					WhenThisDies("Aphetto Vulture — put target Zombie card from your graveyard on top of your library",
						PutChosenTargetOnTopOfLibrary),
					"Aphetto Vulture — put target Zombie card from your graveyard on top of your library?"),
				TargetCardInGraveyard("target Zombie card from your graveyard", HasSubtype("Zombie"), YouOwn()),
			),
		},
	})
}
