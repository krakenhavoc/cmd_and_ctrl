package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelic Cub — Creature — Cat Angel {1}{W}, 1/1 (Foundations Jumpstart,
// #5929):
//
//	"Whenever this creature becomes the target of a spell or ability for
//	 the first time each turn, put a +1/+1 counter on it.
//	 As long as this creature has three or more +1/+1 counters on it, it
//	 has flying."
//
// The trigger is valiant's scan of the turn's event log with no limit on
// who controls the spell or ability (angelicCubTargetedFirstTime); the
// counter goes on only while the Cub is still the targeted permanent. The
// flying is Skyknight Squire's layer 6 static, timestamped at the third
// counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8c110db-4d98-4087-a3b0-c7fe88b5e38a",
		Name:         "Angelic Cub",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.AngelicCubTargetedFirstTime,
				"Angelic Cub — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Static: []game.StaticAbility{{
			Label:            "As long as this creature has three or more +1/+1 counters on it, it has flying.",
			Layer:            game.Layer6Ability,
			AppliesTo:        effects.SkyknightSquireHasThreeCounters,
			StartsAtCounters: effects.SkyknightSquireThirdCounter,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "flying")
			},
		}},
	})
}
