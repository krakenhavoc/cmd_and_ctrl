package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elder Druid — Creature — Elf Cleric Druid {3}{G}, 2/2 (Ice Age, S82 OLD,
// #5636):
//
//	"{3}{G}, {T}: You may tap or untap target artifact, creature, or land."
//
// The tap-or-untap pick is made as it resolves (Dawn Charm-style option pick);
// a target that has left takes the clause with it (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a440a4f7-843f-4186-8ca9-df99f6aa321e",
		Name:         "Elder Druid",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}, {T}: You may tap or untap target artifact, creature, or land.",
			Cost:    effects.Plus(effects.ManaCost("{3}{G}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
			Effect:  effects.B38TapOrUntapTarget("Elder Druid"),
		}},
	})
}
