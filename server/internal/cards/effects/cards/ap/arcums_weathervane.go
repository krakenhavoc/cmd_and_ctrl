package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcum's Weathervane — Artifact {2} (Ice Age, S82 OLD, #5636):
//
//	"{2}, {T}: Target snow land is no longer snow.
//	 {2}, {T}: Target nonsnow basic land becomes snow."
//
// Each is a layer-4 supertype change with no duration, so it lasts as long as the
// land stays on the battlefield (CR 611.2a, #6199).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4e57fba-5fa3-4fd5-8f0a-7b61593ade72",
		Name:         "Arcum's Weathervane",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Target snow land is no longer snow.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target snow land", effects.And(effects.Land(), effects.Snow())),
			Effect:  effects.OldB833TargetSnowChange("Arcum's Weathervane", false),
		}, {
			Label:   "{2}, {T}: Target nonsnow basic land becomes snow.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target nonsnow basic land", effects.OldB833BasicLand(effects.Not(effects.Snow()))),
			Effect:  effects.OldB833TargetSnowChange("Arcum's Weathervane", true),
		}},
	})
}
