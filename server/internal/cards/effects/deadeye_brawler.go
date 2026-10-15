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

// deadeyeBrawlerDraw re-checks the blessing as it resolves (CR 603.4).
var deadeyeBrawlerDraw = Do(If{
	Cond: CondCitysBlessing{},
	Then: []Applier{DrawCards{N: 1}},
})
