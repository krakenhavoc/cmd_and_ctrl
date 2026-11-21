package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sarcophagus, Ship of the Dead — Legendary Artifact — Spacecraft {2}{R}
// (Star Trek, #4973), 4/4:
//
//	"Other creatures you control get +1/+0.
//	 Whenever a creature dies, if Sarcophagus isn't a creature, put a
//	 charge counter on it.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 10+.)
//	 10+ | Flying"
//
// The anthem is printed above the threshold bar, so it is on from the
// moment the Spacecraft lands. The dies trigger counts any creature,
// anyone's, and its "if" is an intervening one (CR 603.4): checked as a
// creature dies and again as the trigger resolves, so once Sarcophagus
// has become a creature (10+ charge counters, or any other effect) it
// stops gaining counters, and a trigger already waiting does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22b86863-c992-405b-949a-bdf8859884ef",
		Name:         "Sarcophagus, Ship of the Dead",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(10, 4, 4),
			effects.ThresholdKeywords(10, "flying"),
			effects.TribalAnthem(effects.TribeFilter{Others: true, YoursOnly: true}, 1, 0),
		),
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "Sarcophagus isn't a creature",
				Test: func(ctx *effects.Context) bool {
					c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
					return ok && effects.OnBattlefield(ctx.Game, ctx.Source()) && !c.IsCreature()
				},
			}, effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died && !source.IsCreature()
			}, "Sarcophagus, Ship of the Dead — put a charge counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterCharge, N: 1}))),
		},
	})
}
