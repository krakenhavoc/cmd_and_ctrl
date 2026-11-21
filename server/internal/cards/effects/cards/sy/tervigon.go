package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tervigon — Creature — Tyranid {X}{1}{G}, 0/0:
//
//	"Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or
//	 more, draw a card when it enters.)
//	 Trample
//	 Spawn Termagants — Whenever this creature deals combat damage to a
//	 player, create that many 1/1 green Tyranid creature tokens."
//
// "That many" is the damage the event dealt.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "a82842b8-cc2e-44fa-b869-7bccc16ca74f",
		Name:                       "Tervigon",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		PrintedKeywords:            []string{"trample"},
		EntersWithCountersFromCast: effects.RavenousEntry(),
		Triggered: []game.TriggeredAbility{
			effects.RavenousDraw("Tervigon"),
			effects.WheneverThisDealsCombatDamageToAPlayer("Tervigon — create that many 1/1 green Tyranid creature tokens",
				func(g *game.Game, item *game.StackItem) error {
					n := item.Trigger.Event.Amount
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 green Tyranid"), N: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
