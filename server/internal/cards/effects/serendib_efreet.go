package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serendib Efreet — Creature — Efreet {2}{U}, 3/4 (Jumpstart 2020):
//
//	"Flying
//	 At the beginning of your upkeep, this creature deals 1 damage to you."
//
// Flying is the printed keyword. The upkeep trigger is unconditional:
// it deals the damage from the Efreet even when the Efreet has left
// the battlefield in response, using the source as it last existed
// (CR 608.2h), so a lifelink or infect it had at the time applies.
//
// No simplification. Deck request #4894 (Jumpstart Azorius).
func init() {
	Register(Spec{
		OracleID:        "dd520d83-297a-487f-b8f3-4997bbc056e0",
		Name:            "Serendib Efreet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Serendib Efreet — this creature deals 1 damage to you",
				func(g *game.Game, item *game.StackItem) error {
					return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
