package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dross Harvester — Creature — Horror {1}{B}{B}, 4/4 (Mirrodin, #5606):
//
//	"Protection from white
//	 At the beginning of your end step, you lose 4 life.
//	 Whenever a creature dies, you gain 2 life."
//
// Protection is the printed keyword (CR 702.16a). The end-step loss is
// a life loss, not damage (CR 119.3), so nothing prevents it. The life
// gain is Black Market's death watch: every creature that goes to a
// graveyard from the battlefield, anyone's, the Harvester itself
// included, since a dies trigger looks back at the creature as it left
// (CR 603.10a). A bounced or exiled creature does not count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5520ada0-f173-4b5e-a0b8-a5938a1c0a4b",
		Name:            "Dross Harvester",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from white"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Dross Harvester — you lose 4 life", effects.Do(effects.GainLife{Amount: -4})),
			effects.TriggerWithPurpose(effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.DiedCreature(ev, g)
				return ok
			}, "Dross Harvester — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})), game.Purpose{DeathPayoff: true}),
		},
	})
}
