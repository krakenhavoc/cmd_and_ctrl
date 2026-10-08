package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadeye Brawler — Creature — Human Pirate {2}{U}{B}, 2/4:
//
//	"Deathtouch
//	 Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 Whenever this creature deals combat damage to a player, if you have
//	 the city's blessing, draw a card."
//
// The "if" is intervening (CR 603.4): checked as the damage is dealt and
// again as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a7c36b66-4f58-41af-b1c5-704ff0d9eac8",
		Name:            "Deadeye Brawler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", game.KeywordAscend},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, AllOf(ThisDealtCombatDamageToAPlayer, YouHaveTheCitysBlessingNow),
				"Deadeye Brawler — draw a card", deadeyeBrawlerDraw),
		},
	})
}

func deadeyeBrawlerDraw(g *game.Game, item *game.StackItem) error {
	if !YouHaveTheCitysBlessing(g, item.Controller) {
		return nil // CR 603.4
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
}
