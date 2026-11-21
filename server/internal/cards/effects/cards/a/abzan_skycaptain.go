package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abzan Skycaptain — Creature — Bird Soldier {3}{W}, 2/2 (Fate Reforged,
// #5552):
//
//	"Flying
//	 When this creature dies, bolster 2. (Choose a creature with the
//	 least toughness among creatures you control and put two +1/+1
//	 counters on it.)"
//
// Flying is printed. The dies trigger bolsters as it resolves (CR 701.39a);
// the Skycaptain is in the graveyard by then, so it is never the chosen
// creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "866fc4b9-f661-444b-b7fd-929814dbc69d",
		Name:            "Abzan Skycaptain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Abzan Skycaptain — bolster 2", effects.Do(effects.Bolster{N: 2})),
		},
	})
}
