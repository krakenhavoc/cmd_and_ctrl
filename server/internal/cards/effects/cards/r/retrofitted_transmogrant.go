package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Retrofitted Transmogrant — Artifact Creature — Zombie {B}, 1/1
// (Modern Horizons 3):
//
//	"{3}{B}: Return this card from your graveyard to the battlefield tapped
//	 with two +1/+1 counters on it."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "058dae67-807e-4542-ae51-f2b1216ffaed",
		Name:         "Retrofitted Transmogrant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}: Return this card from your graveyard to the battlefield tapped with two +1/+1 counters on it.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  gapB49ReturnTappedWithTwoCounters,
		}},
	})
}

func gapB49ReturnTappedWithTwoCounters(g *game.Game, item *game.StackItem) error {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(item.SourceCardID, item.Controller, true,
		map[string]int{game.CounterPlusOne: 2}, game.KeyedTail{})
}
