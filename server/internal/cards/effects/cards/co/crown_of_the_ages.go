package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crown of the Ages — Artifact {2} (Ice Age, S82 OLD #5636):
//
//	"{4}, {T}: Attach target Aura attached to a creature to another creature."
//
// Autumn-Tail, Kitsune Sage's move, for {4} and a tap: the target is an Aura on a
// creature; the controller picks the new host among the creatures it can enchant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b772ef2-b760-400b-8e5b-f4b98e6c935d",
		Name:         "Crown of the Ages",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Attach target Aura attached to a creature to another creature.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Targets: effects.BattlefieldSpec("permanent", "target Aura attached to a creature", effects.OldB717AuraOnACreature),
			Effect:  effects.CrownOfTheAgesMove,
		}},
	})
}
