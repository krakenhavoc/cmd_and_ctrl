package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Momentum Rumbler — Creature — Dinosaur {3}{R}, 3/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"Whenever this creature attacks, if it doesn't have first strike, put a
//	 first strike counter on it.
//	 Whenever this creature attacks, if it has first strike, it gains double
//	 strike until end of turn."
//
// Two printed triggers, each with an intervening "if" (CR 603.4) read through
// the layered view at the attack and again as the trigger resolves, so a
// Rumbler that has first strike from anywhere gets double strike and one that
// does not gets the counter. They are mutually exclusive when the attack is
// declared; the first attack puts the counter and the next attack gains double
// strike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "502230fb-3a7a-439c-adfa-f7820deacf87",
		Name:         "Momentum Rumbler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.S82OldB113ThisHasFirstStrike(false),
				effects.On(game.EventAttack, effects.S82OldB113ThisAttackedWith(false), "Momentum Rumbler — put a first strike counter on it",
					effects.Do(effects.CounterOnThis{Kind: game.CounterFirstStrike, N: 1}))),
			effects.InterveningIf(effects.S82OldB113ThisHasFirstStrike(true),
				effects.On(game.EventAttack, effects.S82OldB113ThisAttackedWith(true), "Momentum Rumbler — it gains double strike until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"double strike"}, Label: "Momentum Rumbler — double strike"}.Apply(effects.NewContext(g, item))
					})),
		},
	})
}
