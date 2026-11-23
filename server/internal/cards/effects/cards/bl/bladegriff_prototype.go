package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "9fff90b6-98ef-4f78-b08d-7e46dfc03403",
		Name:            "Bladegriff Prototype",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisDealsCombatDamageToAPlayer("Bladegriff Prototype — destroy target nonland permanent of that player's choice that one of your opponents controls",
					effects.DestroyFirstLegalCardTarget),
				effects.Harmful(effects.OfThatPlayersChoice(effects.TargetPermanent("target nonland permanent of that player's choice that one of your opponents controls", effects.Nonland(), effects.Not(effects.YouControl()))))),
		},
	})
}
