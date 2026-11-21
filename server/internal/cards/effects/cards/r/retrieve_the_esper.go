package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Retrieve the Esper — Sorcery {3}{U} (Final Fantasy, #4462):
//
//	"Create a 3/3 blue Robot Warrior artifact creature token. Then if
//	 this spell was cast from a graveyard, put two +1/+1 counters on
//	 that token.
//	 Flashback {5}{U}"
//
// Antiquities on the Loose's shape: the cast zone is the spell's
// announce record. The token is the one this spell made, so the
// counters wait for it (TokenSpec.WithCounters enters it with them,
// before any enters trigger, which is what "put on that token" yields
// for a token no replacement touches).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "9d6fdbd4-ba51-4df1-b383-9b709b36a8d7",
		Name:             "Retrieve the Esper",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{5}{U}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			spec := effects.Token(effects.TokenCard("3/3 blue Robot Warrior artifact"))
			if from := item.CastFromZone; from != "" && from != game.ZoneHand {
				spec = spec.WithCounters(game.CounterPlusOne, 2)
			}
			return effects.CreateTokenAdvanced{Spec: spec, N: 1}.Apply(ctx)
		},
	})
}
