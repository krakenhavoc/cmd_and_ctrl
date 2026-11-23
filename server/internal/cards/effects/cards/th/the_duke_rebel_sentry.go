package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Duke, Rebel Sentry — Legendary Creature — Human Rebel Ally {W},
// 0/1 (Avatar: The Last Airbender Eternal, #5075):
//
//	"The Duke enters with a +1/+1 counter on him.
//	 {T}, Remove a counter from The Duke: Put a +1/+1 counter on another
//	 target creature you control. It gains hexproof until end of turn."
//
// The enters-with is a replacement (Scholar of New Horizons'), so he is
// a 1/2 from the moment he enters. The cost removes one counter of any
// kind from him; the +1/+1 counter and the hexproof land on the other
// creature only if it is still a legal target as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c55047d-badc-49bb-b791-e9759577e4eb",
		Name:         "The Duke, Rebel Sentry",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 1, "The Duke enters with a +1/+1 counter on him"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove a counter from The Duke: Put a +1/+1 counter on another target creature you control. It gains hexproof until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis("", 1)),
			Targets: effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target := effects.FirstLegalBattlefieldTarget(ctx)
				if target == uuid.Nil {
					return nil
				}
				if err := (effects.AddCounter{Target: target, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"hexproof"},
					Label: "The Duke, Rebel Sentry — hexproof until end of turn"}.Apply(ctx)
			},
		}},
	})
}
