package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reaper King — Legendary Artifact Creature — Scarecrow {2/W}{2/U}{2/B}{2/R}{2/G},
// 6/6 (Secret Lair Drop):
//
//	"Other Scarecrow creatures you control get +1/+1.
//	 Whenever another Scarecrow you control enters, destroy target
//	 permanent."
//
// The cost is five {2/C} hybrid symbols, each paid with its colour or two
// mana (mana value 10). The anthem is a layer-7c tribal bonus over the other
// Scarecrow creatures its controller has; the trigger is on the entry of
// another Scarecrow under its controller's control (a token too, and one that
// enters under its control from anywhere) and targets any permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70034860-5198-421f-871d-7c1676337b6e",
		Name:         "Reaper King",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Scarecrow"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.HasSubtypeWord("Scarecrow")
			}, "Reaper King — destroy target permanent", effects.DestroyFirstLegalTarget),
				effects.TargetPermanent("target permanent")),
		},
	})
}
