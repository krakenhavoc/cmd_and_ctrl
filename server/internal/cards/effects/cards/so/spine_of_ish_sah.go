package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spine of Ish Sah — Artifact {7} (EDHREC rank 2568):
//
//	"When this artifact enters, destroy target permanent.
//	 When this artifact is put into a graveyard from the battlefield,
//	 return it to its owner's hand."
//
// Seven mana for a Vindicate that comes back. The entry trigger is a
// targeted ETB over any permanent (destroyChosenTargetTrigger, the
// Acidic Slime shape — the controller picks as it goes on the
// stack, an indestructible target survives). The second is a dies
// trigger: cardDied is the graveyard-only exit, and on resolution
// the Spine goes from the graveyard to its owner's hand if it is
// still there — a Spine exiled from the graveyard in response stays
// exiled (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02f062f5-8012-4440-ac12-49fc49822106",
		Name:         "Spine of Ish Sah",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.B06SelfETB,
				Targets:   effects.TargetPermanent("target permanent"),
				Key:       "Spine of Ish Sah — destroy target permanent",
				Effect:    effects.DestroyChosenPermanent,
			},
			effects.WhenThisDies("Spine of Ish Sah — return it to its owner's hand", effects.ReturnThisCardFromYourGraveyardToHand),
		},
	})
}
