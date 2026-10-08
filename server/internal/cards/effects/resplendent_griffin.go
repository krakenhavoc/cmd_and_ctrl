package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Resplendent Griffin — Creature — Griffin {1}{W}{U}, 2/2:
//
//	"Flying
//	 Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 Whenever this creature attacks, if you have the city's blessing,
//	 put a +1/+1 counter on it."
//
// The "if" is intervening (CR 603.4): checked as the attack is declared,
// so without the blessing no trigger goes on the stack, and again as the
// trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f8bf5df2-b783-4796-b2c6-eb4dd4167d62",
		Name:            "Resplendent Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordAscend},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, AllOf(ThisAttacked, YouHaveTheCitysBlessingNow),
				"Resplendent Griffin — put a +1/+1 counter on it", blessedSelfCounter),
		},
	})
}

// blessedSelfCounter is "put a +1/+1 counter on this creature" behind
// an intervening "if you have the city's blessing" (Resplendent Griffin,
// Mausoleum Harpy): the blessing is asked again as the trigger resolves
// (CR 603.4).
func blessedSelfCounter(g *game.Game, item *game.StackItem) error {
	if !YouHaveTheCitysBlessing(g, item.Controller) || !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	return g.AddCounterForEffect(item.SourceCardID, game.CounterPlusOne, 1)
}
