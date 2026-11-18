package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reito Sentinel — Artifact Creature — Construct {3}, 3/3 (Kamigawa: Neon
// Dynasty, S82):
//
//	"Defender
//	 When this creature enters, target player mills three cards.
//	 {3}: Put target card from a graveyard on the bottom of its owner's
//	 library."
//
// The enter trigger is Exhibition Tidecaller's player-target mill
// (TargetPlayerMills); the activation is Dutiful Knowledge Seeker's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e574bfdb-505b-4d61-a376-6d26f03cd9b1",
		Name:            "Reito Sentinel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Reito Sentinel — target player mills three cards", TargetPlayerMills(3)),
				TargetPlayer("target player")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}: Put target card from a graveyard on the bottom of its owner's library.",
			Cost:    ManaCost("{3}"),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect:  socB37PutTargetGraveyardCardOnBottom,
		}},
	})
}
