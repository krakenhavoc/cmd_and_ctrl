package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragoon's Wyvern — Creature — Drake {2}{U}, 2/1 (Final Fantasy, #4462):
//
//	"Flying
//	 When this creature enters, create a 1/1 colorless Hero creature
//	 token."
//
// Magitek Armor's enters trigger on a flier: the Hero is the token job
// select makes, made on its own. Flying rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f328012f-4696-47bb-ab43-247bdcb13f03",
		Name:            "Dragoon's Wyvern",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dragoon's Wyvern — create a 1/1 colorless Hero creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard(effects.JobSelectHeroToken), N: 1})),
		},
	})
}
