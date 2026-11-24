package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blind Hunter — Creature — Bat {2}{W}{B}, 2/2 (Guildpact):
//
//	"Flying
//	 Haunt (When this creature dies, exile it haunting target creature.)
//	 When this creature enters or the creature it haunts dies, target
//	 player loses 2 life and you gain 2 life."
//
// Haunt is CR 702.55 (#6083). The second ability is one ability with two
// conditions; from exile its controller is the card's owner (CR 108.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a731f5bc-892c-4857-9590-4fa31c1caf00",
		Name:            "Blind Hunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Haunt("Blind Hunter"),
			effects.Targeting(
				effects.WhenThisEntersOrTheCreatureItHauntsDies("Blind Hunter — target player loses 2 life and you gain 2 life",
					effects.Do(effects.TargetPlayerLosesLife{N: 2}, effects.GainLife{Amount: 2})),
				effects.TargetPlayer("target player"),
			),
		},
	})
}
