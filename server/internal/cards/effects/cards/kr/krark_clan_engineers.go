package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krark-Clan Engineers — Creature — Goblin Artificer {3}{R}, 2/2 (Fifth Dawn, #5604):
//
//	"{R}, Sacrifice two artifacts: Destroy target artifact."
//
// Goblin Replica's artifact kill, paid with {R} and Metalwork Colossus's
// "sacrifice two artifacts" (one simultaneous exit).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0964288c-0397-428a-9a84-784aaea4e040",
		Name:         "Krark-Clan Engineers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice two artifacts: Destroy target artifact.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeN(2, "two artifacts", effects.Artifact())),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
