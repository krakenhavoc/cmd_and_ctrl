package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tallyman of Nurgle — Creature — Astartes Warrior {2}{B}, 2/3 (Warhammer
// 40,000 Commander, #5948):
//
//	"Lifelink
//	 The Seven-fold Chant — At the beginning of your end step, if a creature
//	 died this turn, you draw a card and you lose 1 life. If seven or more
//	 creatures died this turn, instead you draw seven cards and you lose 7 life."
//
// Bulette's end-step trigger over the table-wide tally of creatures that died
// this turn (tokens included). The intervening "if" (CR 603.4) is the
// trigger-time test and is rechecked as the ability resolves; the two branches
// are exclusive, so the seven-card one is "instead" of the one-card one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7cc4a635-8730-4018-a8fd-b29fb5479928",
		Name:            "Tallyman of Nurgle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, effects.S82AfrB07AnyCreatureDiedWhen),
				"Tallyman of Nurgle — draw a card and lose 1 life; if seven or more creatures died this turn, instead draw seven cards and lose 7 life",
				effects.Do(
					effects.If{Cond: effects.CondGapB189SevenOrMoreCreaturesDied{}, Then: []effects.Applier{effects.DrawCards{N: 7}, effects.GainLife{Amount: -7}}},
					effects.If{Cond: effects.CondGapB189FewerThanSevenCreaturesDied{}, Then: []effects.Applier{effects.DrawCards{N: 1}, effects.GainLife{Amount: -1}}},
				)),
		},
	})
}
