package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleric of Life's Bond — Creature — Vampire Cleric {W}{B}, 2/2 (Zendikar
// Rising, S82 OLD #5508):
//
//	"Whenever another Cleric you control enters, you gain 1 life.
//	 Whenever you gain life for the first time each turn, put a +1/+1
//	 counter on this creature."
//
// Another Cleric is any other creature you control with the Cleric type as
// it enters (a changeling counts, CR 702.73a). The first life gain of a turn
// is read off the turn's log, so a gain earlier in the turn, before the
// Cleric arrived, makes a later one not the first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7721c5e-b3c5-45c8-b757-935603dfe44b",
		Name:         "Cleric of Life's Bond",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Cleric"))
			}, "Cleric of Life's Bond — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
			effects.On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.SocB28FirstLifeGainThisTurn(ev, source, g)
			}, "Cleric of Life's Bond — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
