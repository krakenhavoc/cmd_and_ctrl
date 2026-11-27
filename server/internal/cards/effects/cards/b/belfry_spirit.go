package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Belfry Spirit — Creature — Spirit {3}{W}{W}, 1/1 (Guildpact):
//
//	"Flying
//	 Haunt (When this creature dies, exile it haunting target creature.)
//	 When this creature enters or the creature it haunts dies, create two
//	 1/1 black Bat creature tokens with flying."
//
// Haunt is CR 702.55 (#6083): one ability with two conditions. From exile
// the tokens are the card owner's (CR 108.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c171065f-7947-4161-a481-1ea2d3ce8422",
		Name:            "Belfry Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Haunt("Belfry Spirit"),
			effects.WhenThisEntersOrTheCreatureItHauntsDies("Belfry Spirit — create two 1/1 black Bat creature tokens with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black Bat with flying"), N: 2})),
		},
	})
}
