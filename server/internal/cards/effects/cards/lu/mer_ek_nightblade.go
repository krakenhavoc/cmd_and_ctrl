package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mer-Ek Nightblade — Creature — Orc Assassin {3}{B}, 2/3 (Khans of
// Tarkir, #5554):
//
//	"Outlast {B} ({B}, {T}: Put a +1/+1 counter on this creature.
//	 Outlast only as a sorcery.)
//	 Each creature you control with a +1/+1 counter on it has
//	 deathtouch."
//
// Outlast is the shared sorcery-speed tap-and-pay counter row
// (rprB04Outlast, CR 702.107a). The static is Abzan Battle Priest's
// layer-6 grant over every creature you control with a +1/+1 counter,
// this one included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56a8554c-d427-44ca-a01d-9e12c89acd1e",
		Name:         "Mer-Ek Nightblade",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RprB04Outlast("{B}")},
		Static: []game.StaticAbility{
			effects.CreaturesYouControlWithCountersHave(game.CounterPlusOne, false, "deathtouch"),
		},
	})
}
