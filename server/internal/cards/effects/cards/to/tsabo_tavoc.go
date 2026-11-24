package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tsabo Tavoc — Legendary Creature — Phyrexian Horror {5}{B}{R}, 7/4 (Secret
// Lair Drop):
//
//	"First strike, protection from legendary creatures
//	 {B}{B}, {T}: Destroy target legendary creature. It can't be regenerated."
//
// First strike is printed data. Protection from legendary creatures is the
// closed-grammar quality game.ProtectionQualityLegendaryCreatures (a creature
// source with the legendary supertype), so damage, enchanting, blocking and
// targeting by one are stopped. The destruction ignores regeneration (CR 701.19c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f98f1f71-99ec-4ffb-9d8e-acbab173f972",
		Name:            "Tsabo Tavoc",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", game.ProtectionFromLegendaryCreatures},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{B}, {T}: Destroy target legendary creature. It can't be regenerated.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{B}{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target legendary creature", effects.Legendary()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyTheTargetPermanentNoRegen(item, effects.NewContext(g, item))
			},
		}},
	})
}
