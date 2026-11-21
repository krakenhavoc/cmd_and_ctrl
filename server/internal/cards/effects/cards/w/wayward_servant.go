package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wayward Servant — Creature — Zombie {W}{B}, 2/2:
//
//	"Whenever another Zombie you control enters, each opponent loses
//	 1 life and you gain 1 life."
//
// "Another" excludes the Servant's own entry, so it never drains for
// itself, and "you control" is read as the entering permanent's
// controller. Every opponent loses 1 (a life LOSS, not damage) and the
// controller gains 1 once, however many opponents there are.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2916f041-f080-450f-ae79-dc09076d971a",
		Name:         "Wayward Servant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Zombie"))
			}, "Wayward Servant — each opponent loses 1 life and you gain 1 life", effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
