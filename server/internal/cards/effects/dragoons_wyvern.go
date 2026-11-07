package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "f328012f-4696-47bb-ab43-247bdcb13f03",
		Name:            "Dragoon's Wyvern",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dragoon's Wyvern — create a 1/1 colorless Hero creature token",
				Do(CreateToken{Template: TokenCard(jobSelectHeroToken), N: 1})),
		},
	})
}
