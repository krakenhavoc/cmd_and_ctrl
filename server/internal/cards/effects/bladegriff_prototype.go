package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bladegriff Prototype — Artifact Creature {5}, 3/2:
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, destroy target
//	 nonland permanent of that player's choice that one of your opponents
//	 controls."
//
// "That player" is the player dealt the damage. The clause is
// OfThatPlayersChoice (the damaged player is asked, game.triggerChooserOf);
// legality is judged for Bladegriff's controller, so "one of your opponents
// controls" is written from that side as Not(YouControl()).
func init() {
	Register(Spec{
		OracleID:        "9fff90b6-98ef-4f78-b08d-7e46dfc03403",
		Name:            "Bladegriff Prototype",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisDealsCombatDamageToAPlayer("Bladegriff Prototype — destroy target nonland permanent of that player's choice that one of your opponents controls",
					destroyFirstLegalCardTarget),
				Harmful(OfThatPlayersChoice(TargetPermanent("target nonland permanent of that player's choice that one of your opponents controls", Nonland(), Not(YouControl()))))),
		},
	})
}
