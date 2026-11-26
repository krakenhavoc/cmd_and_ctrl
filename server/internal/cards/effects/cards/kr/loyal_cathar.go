package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loyal Cathar // Unhallowed Cathar — Creature — Human Soldier {W}{W} 2/2 //
// Creature — Zombie Soldier 2/1:
//
//	Loyal Cathar: "Vigilance
//	 When this creature dies, return it to the battlefield transformed under
//	 your control at the beginning of the next end step."
//	Unhallowed Cathar: "This creature can't block."
//
// The dies trigger schedules a delayed trigger (CR 603.7) over the graveyard
// object the Cathar became; a card that has left the graveyard by then stays
// where it went (CR 400.7). It returns under the trigger's controller.
func init() {
	const oracle = "bf28a2ed-cad7-48a7-9dec-06e43af73e9c"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Loyal Cathar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Loyal Cathar — return it to the battlefield transformed at the beginning of the next end step",
				effects.OldB443ReturnThisCardTransformedAtNextEndStep("Loyal Cathar — return it transformed")),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Unhallowed Cathar",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
