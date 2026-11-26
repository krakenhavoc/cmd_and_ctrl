package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hellion Crucible — Land (Magic 2013, S82 OLD, #5568):
//
//	"{T}: Add {C}.
//	 {1}{R}, {T}: Put a pressure counter on this land.
//	 {1}{R}, {T}, Remove two pressure counters from this land and sacrifice it:
//	 Create a 4/4 red Hellion creature token with haste."
//
// The first ability is a mana ability; the other two are ordinary activated
// abilities. The third pays its counters and sacrifice as costs (CR 118.3), so it
// cannot be activated with fewer than two pressure counters, and the token is
// created with haste, as printed on the ability (the Hellion token card itself
// has none).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c238ef51-4b46-43d5-a70b-40270a96a1fd",
		Name:         "Hellion Crucible",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}.",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{R}, {T}: Put a pressure counter on this land.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost()),
				Effect:  effects.PutACounterOnThis("pressure"),
			},
			{
				Label:   "{1}{R}, {T}, Remove two pressure counters from this land and sacrifice it: Create a 4/4 red Hellion creature token with haste.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost(), effects.RemoveCountersFromThis("pressure", 2), effects.SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("4/4 red Hellion")).WithKeywords("haste"), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
