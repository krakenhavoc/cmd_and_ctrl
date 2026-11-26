package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Requiem Angel — Creature — Angel {5}{W} 5/5 (Commander 2014):
//
//	"Flying
//	 Whenever another non-Spirit creature you control dies, create a 1/1
//	 white Spirit creature token with flying."
//
// The dying creature is judged as it last existed (CR 603.10a): its
// creature types, its controller, and "another" is object identity.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "59f10ddb-8287-4d2c-8c60-1080b108fa78",
		Name:            "Requiem Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && !effects.LeftAsSubtype(ev, dead, source.TextSubtype("Spirit"))
			}, "Requiem Angel — create a 1/1 white Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1})),
				game.Purpose{DeathPayoff: true, Tokens: 1}),
		},
	})
}
