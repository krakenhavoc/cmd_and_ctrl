package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sandstorm Salvager — Creature — Human Artificer {2}{G}, 3/3 (The Big
// Score, #4468):
//
//	"When this creature enters, create a 3/3 colorless Golem artifact
//	 creature token.
//	 {2}, {T}: Put a +1/+1 counter on each creature token you control.
//	 They gain trample until end of turn."
//
// The creature tokens are read once at resolution (CR 611.2c); "they" is
// that same set, so a token that arrives later gains neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec19a9bc-6fc1-4150-976b-37a47b165016",
		Name:         "Sandstorm Salvager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sandstorm Salvager — create a 3/3 Golem", effects.Do(effects.CreateToken{Template: effects.B35GolemToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Put a +1/+1 counter on each creature token you control. They gain trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ids := effects.B35CreatureTokensYouControl(g, item.Controller)
				if err := (effects.GrantKeywordUntilEOT{Match: effects.B35IDSet(ids), Keywords: []string{"trample"}, Label: "Sandstorm Salvager — trample"}).Apply(ctx); err != nil {
					return err
				}
				return effects.B13PutCounterOnEach(ctx, ids)
			},
		}},
	})
}
