package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "e574bfdb-505b-4d61-a376-6d26f03cd9b1",
		Name:            "Reito Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Reito Sentinel — target player mills three cards", effects.TargetPlayerMills(3)),
				effects.TargetPlayer("target player")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Put target card from a graveyard on the bottom of its owner's library.",
			Cost:    effects.ManaCost("{3}"),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  effects.SocB37PutTargetGraveyardCardOnBottom,
		}},
	})
}
