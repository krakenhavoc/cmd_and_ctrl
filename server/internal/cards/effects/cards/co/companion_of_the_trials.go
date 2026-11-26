package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Companion of the Trials — {2}{W} Creature — Bird Soldier 2/2 (Amonkhet, #5535):
//
//	"Flying
//	 {1}{W}: Untap target creature. Activate only if you control a Gideon
//	 planeswalker."
//
// The gate is read as the ability is activated (CR 602.5b): a planeswalker
// with the subtype Gideon, whatever its name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3aa654cb-c0a0-48e7-901a-c690120b1510",
		Name:            "Companion of the Trials",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}{W}: Untap target creature. Activate only if you control a Gideon planeswalker.",
			Cost:      effects.ManaCost("{1}{W}"),
			Targets:   effects.TargetCreature("target creature"),
			Condition: effects.ControlsAtLeast(1, func(c game.Card) bool { return c.IsPlaneswalker() && c.HasSubtype("Gideon") }),
			Effect:    effects.UntapEachLegalTarget,
		}},
	})
}
