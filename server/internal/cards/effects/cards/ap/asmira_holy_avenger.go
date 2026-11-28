package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Asmira, Holy Avenger — Legendary Creature — Human Cleric {2}{G}{W}, 2/3
// (Mirage, S82 OLD b805):
//
//	"Flying
//	 At the beginning of each end step, put a +1/+1 counter on Asmira for
//	 each creature put into your graveyard from the battlefield this turn."
//
// Every end step, whoever's it is, reads the per-owner tally of creatures put
// into the Asmira controller's graveyard from the battlefield this turn
// (PlayerTurnTally.CreaturesToGraveyard: the graveyard is the owner's, CR
// 404.1, whoever controlled the creature; tokens count). It is not an
// intervening "if", so the trigger always goes on the stack and puts nothing
// on when the count is zero. Asmira herself dying earlier does not matter: a
// source that has left gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a0397f06-d557-4c9f-af89-238ab38ada3e",
		Name:            "Asmira, Holy Avenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AnyPlayer,
				"Asmira, Holy Avenger — put a +1/+1 counter on her for each creature put into your graveyard from the battlefield this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := g.CreaturesToGraveyardThisTurn(item.Controller)
					if n <= 0 || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
