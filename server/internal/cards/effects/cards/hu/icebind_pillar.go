package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icebind Pillar — Snow Artifact {2}{U} (Dominaria Remastered):
//
//	"{S}, {T}: Tap target artifact or creature."
//
// {S} is snow mana (Frost Raptor); the tap is Avacynian Priest's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18050051-63f6-4091-bf67-7a0701a0791d",
		Name:         "Icebind Pillar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{S}, {T}: Tap target artifact or creature.",
			Cost:    effects.Plus(effects.ManaCost("{S}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
