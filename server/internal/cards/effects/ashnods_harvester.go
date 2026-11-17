package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ashnod's Harvester — Artifact Creature — Construct {2}, 3/1:
//
//	"Whenever this creature attacks, exile target card from a graveyard.
//	 Unearth {1}{B}"
func init() {
	Register(Spec{
		OracleID:     "9ea87566-43da-4ebd-a904-ae8b8830c3a2",
		Name:         "Ashnod's Harvester",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Ashnod's Harvester — exile target card from a graveyard", b27ExileChosenTarget),
				TargetCardInGraveyard("target card from a graveyard")),
		},
		Activated: []ActivatedAbility{Unearth("{1}{B}")},
	})
}
