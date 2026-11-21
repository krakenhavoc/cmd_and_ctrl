package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cliffhaven Vampire — Creature — Vampire Warrior Ally {2}{W}{B}, 2/4:
//
//	"Flying
//	 Whenever you gain life, each opponent loses 1 life."
//
// Watches EventChangeLife for a positive delta on the controller, the
// event every lifegain path emits (lifelink included), so it fires once
// per gain event — two lifelink damage events are two triggers, a
// single 5-point gain is one. The drain is life LOSS, not damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1915a311-209b-4628-8122-af3055a78fed",
		Name:            "Cliffhaven Vampire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Target == source.Controller && ev.Amount > 0
			}, "Cliffhaven Vampire — each opponent loses 1 life", effects.Do(effects.EachOpponentLosesLife{N: 1})),
		},
	})
}
