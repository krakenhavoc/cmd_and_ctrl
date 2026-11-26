package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Joraga Warcaller — Creature — Elf Warrior {G} 1/1 (Commander 2014):
//
//	"Multikicker {1}{G}
//	 This creature enters with a +1/+1 counter on it for each time it was
//	 kicked.
//	 Other Elf creatures you control get +1/+1 for each +1/+1 counter on
//	 this creature."
//
// The counters are placed as it enters (CR 614.1c, Everflowing Chalice's
// shape); the lord bonus reads the counters it has now, so a counter added
// or removed later moves every other Elf.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8959a147-75b1-4f4a-a699-cb354ff55e59",
		Name:         "Joraga Warcaller",
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			effects.Multikicker("{1}{G}", 20),
		},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			effects.CountersPerKick(game.CounterPlusOne, 1),
		},
		Static: []game.StaticAbility{
			effects.TribalScalingAnthem(effects.TribeFilter{Tribes: []string{"Elf"}, Others: true, YoursOnly: true},
				func(source *game.Card, _ *game.Game) int { return source.Counters[game.CounterPlusOne] }),
		},
	})
}
