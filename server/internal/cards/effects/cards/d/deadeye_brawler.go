package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "a7c36b66-4f58-41af-b1c5-704ff0d9eac8",
		Name:            "Deadeye Brawler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", game.KeywordAscend},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.AllOf(effects.ThisDealtCombatDamageToAPlayer, effects.YouHaveTheCitysBlessingNow),
				"Deadeye Brawler — draw a card", deadeyeBrawlerDraw),
		},
	})
}

// deadeyeBrawlerDraw re-checks the blessing as it resolves (CR 603.4).
var deadeyeBrawlerDraw = effects.Do(effects.If{
	Cond: effects.CondCitysBlessing{},
	Then: []effects.Applier{effects.DrawCards{N: 1}},
})
