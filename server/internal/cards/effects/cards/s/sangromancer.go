package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sangromancer — Creature — Vampire Shaman {2}{B}{B}, 3/3 (EDHREC
// rank 2008):
//
//	"Flying
//	 Whenever a creature an opponent controls dies, you may gain 3
//	 life.
//	 Whenever an opponent discards a card, you may gain 3 life."
//
// Two optional lifegain triggers, one per printed ability: the
// dead creature's controller is the one it had as it died
// (leftUnderControlOf, #1682), and the discarding player is the
// discard event's Actor.
// A wheel against three opponents is one prompt per card, as
// printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "920445ab-0ac2-4de7-bc1c-f5e58eb4424c",
		Name:            "Sangromancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B18OpponentsCreatureDied(ev, source, g)
			}, "Sangromancer — gain 3 life (a creature died)", b18GainThree), "Sangromancer — an opponent's creature died. Gain 3 life?"),
			effects.Optional(effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B18OpponentDiscarded(ev, source)
			}, "Sangromancer — gain 3 life (a card was discarded)", b18GainThree), "Sangromancer — an opponent discarded. Gain 3 life?"),
		},
	})
}

// b18GainThree is Sangromancer's shared effect body.
func b18GainThree(g *game.Game, item *game.StackItem) error {
	return effects.GainLife{Player: item.Controller, Amount: 3}.Apply(effects.NewContext(g, item))
}
