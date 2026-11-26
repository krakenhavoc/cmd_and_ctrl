package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Siren of the Fanged Coast — {3}{U}{U} Creature — Siren 1/1 (Born of the Gods):
//
//	"Flying
//	 Tribute 3 (As this creature enters, an opponent of your choice may put three
//	 +1/+1 counters on it.)
//	 When this creature enters, if tribute wasn't paid, gain control of target
//	 creature."
//
// No stated duration (CR 611.2a), as Keiga's: control lasts indefinitely.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c5e7ba95-c9b9-45e8-9551-dba892d90683",
		Name:            "Siren of the Fanged Coast",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    effects.Tribute("Siren of the Fanged Coast", 3),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenEntersIfTributeWasntPaid("Siren of the Fanged Coast — gain control of target creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GainControl{Target: id, Controller: item.Controller, Duration: game.IndefiniteDuration(),
						Label: "Siren of the Fanged Coast — gain control (no stated duration)"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature")),
		},
	})
}
