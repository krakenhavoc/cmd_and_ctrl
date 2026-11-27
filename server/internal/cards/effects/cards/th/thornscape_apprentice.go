package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thornscape Apprentice (Invasion, S82 OLD batch 647, #5616):
//
//	Creature — Human Wizard {G}, 1/1:
//	"{R}, {T}: Target creature gains first strike until end of turn.
//	{W}, {T}: Tap target creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebb68e03-b2a5-4c5c-8e3b-968cd241d1f0",
		Name:         "Thornscape Apprentice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.S82OldB647TargetGainsRow(effects.Plus(effects.ManaCost("{R}"), effects.TapCost()), "{R}, {T}", "first strike", "first strike"),
			effects.S82OldB647TapTargetRow("{W}", "{W}, {T}"),
		},
	})
}
