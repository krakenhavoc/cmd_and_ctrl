package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashnod's Harvester — Artifact Creature — Construct {2}, 3/1:
//
//	"Whenever this creature attacks, exile target card from a graveyard.
//	 Unearth {1}{B}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ea87566-43da-4ebd-a904-ae8b8830c3a2",
		Name:         "Ashnod's Harvester",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Ashnod's Harvester — exile target card from a graveyard", effects.B27ExileChosenTarget),
				effects.TargetCardInGraveyard("target card from a graveyard")),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{1}{B}")},
	})
}
