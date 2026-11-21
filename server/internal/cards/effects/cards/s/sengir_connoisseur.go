package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sengir Connoisseur — Creature — Vampire {3}{B}{B}, 3/3 (Dominaria United,
// S82 DMU batch 4):
//
//	"Flying
//	 Whenever one or more other creatures die, put a +1/+1 counter on this
//	 creature. This ability triggers only once each turn."
//
// Morbid Opportunist's trigger shape (a death batch collapsed to one trigger,
// and a second batch the same turn gated out) with a counter on this creature
// in place of the draw. Flying is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8a4d505e-b884-4b8b-93d5-495992f3858e",
		Name:            "Sengir Connoisseur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				if !effects.AnotherCreatureDied(ev, source, ch, g) {
					return false
				}
				return !effects.B11TriggeredThisTurn(g, source.InstanceID, sengirConnoisseurLabel)
			}, sengirConnoisseurLabel, effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}

const sengirConnoisseurLabel = "Sengir Connoisseur — put a +1/+1 counter on this creature"
