package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Koth of the Homestead — Legendary Creature — Human Citizen {2}{W},
// 2/3:
//
//	"Landfall — Whenever a land you control enters, you gain 1 life.
//	 Whenever a Plains you control enters, put a +1/+1 counter on target
//	 creature."
//
// Two separate abilities. A Plains triggers both, and the second targets
// any creature (Koth himself included) as it goes on the stack; with no
// creature at all it is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "567aeb60-a441-41b1-98f4-54510d299317",
		Name:         "Koth of the Homestead",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Koth of the Homestead — you gain 1 life", game.LandfallPayoff{LifeGain: 1}, effects.Do(effects.GainLife{Amount: 1})),
			effects.Targeting(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					return ok && c.IsLand() && c.HasSubtype(source.TextSubtype("Plains"))
				}, "Koth of the Homestead — a +1/+1 counter on target creature",
					effects.PutACounterOnTheTarget(game.CounterPlusOne)),
				effects.TargetCreature("target creature")),
		},
	})
}
