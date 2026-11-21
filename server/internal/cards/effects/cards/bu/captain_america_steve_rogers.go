package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain America, Steve Rogers — Legendary Creature — Human Soldier Hero
// {3}{W}{W}, 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Lifelink
//	 Whenever Captain America attacks, put a +1/+1 counter on another target
//	 creature you control. That creature gains indestructible until end of
//	 turn."
//
// Lonis, Genetics Expert's "another target creature you control" clause on an
// attack trigger. The target is chosen as the trigger goes on the stack; with
// no other creature there is no trigger. If it left the battlefield in
// response, neither half happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d2084a6b-3248-4ed4-8fb5-d9c439d1ee72",
		Name:            "Captain America, Steve Rogers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventAttack, effects.ThisAttacked,
				"Captain America, Steve Rogers — put a +1/+1 counter on another target creature you control; it gains indestructible until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: "Captain America, Steve Rogers"}.Apply(ctx)
				}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
