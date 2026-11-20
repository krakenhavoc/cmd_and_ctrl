package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slagstone Refinery — Artifact {4} (The Brothers' War):
//
//	"Whenever this artifact or another nontoken artifact you control is
//	 put into a graveyard from the battlefield or is put into exile from
//	 the battlefield, create a tapped Powerstone token."
//
// One printed ability, one registered trigger, watching every
// leaves-the-battlefield event (the Refinery's own included, which is a
// leaves-the-battlefield trigger and so looks back, CR 603.10a). See
// nontokenArtifactYouControlWasPutIntoAGraveyardOrExileFromTheBattlefield.
// Each such artifact is its own event and its own trigger, so two
// artifacts dying together make two Powerstones.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a23c4378-9242-46a0-9f62-ca91f0baa428",
		Name:         "Slagstone Refinery",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.NontokenArtifactYouControlWasPutIntoAGraveyardOrExileFromTheBattlefield,
				"Slagstone Refinery — create a tapped Powerstone token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1})),
		},
	})
}
