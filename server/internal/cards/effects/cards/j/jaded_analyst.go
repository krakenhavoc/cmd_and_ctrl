package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jaded Analyst — Creature — Human Detective {1}{U}, 3/2 (Murders at
// Karlov Manor, #4469):
//
//	"Defender
//	 Whenever you draw your second card each turn, this creature loses
//	 defender and gains vigilance until end of turn."
//
// The trigger is Knights of Dol Amroth's condition
// (b41DrewYourSecondCardThisTurn). Losing defender is a layer-6
// ability-removing effect with its own timestamp (CR 613.1f,
// thisLosesKeywordUntilEndOfTurn), so the Analyst can attack for the
// rest of the turn; vigilance is a layer-6 grant beside it. Both end at
// cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e43400fe-3df5-4942-abc3-b9bba34fc7cd",
		Name:            "Jaded Analyst",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Jaded Analyst — loses defender and gains vigilance until end of turn", jadedAnalystUnleashed),
		},
	})
}

// jadedAnalystUnleashed removes defender from the Analyst and grants it
// vigilance, both until end of turn.
func jadedAnalystUnleashed(g *game.Game, item *game.StackItem) error {
	if err := effects.ThisLosesKeywordUntilEndOfTurn("defender", "Jaded Analyst — loses defender until end of turn")(g, item); err != nil {
		return err
	}
	return effects.ThisGainsKeywordUntilEndOfTurn("vigilance", "Jaded Analyst — gains vigilance until end of turn")(g, item)
}
