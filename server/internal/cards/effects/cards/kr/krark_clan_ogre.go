package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krark-Clan Ogre — Creature — Ogre {3}{R}{R}, 3/3 (Fifth Dawn, #5604):
//
//	"{R}, Sacrifice an artifact: Target creature can't block this turn."
//
// Arena Athlete's "can't block this turn" on any target creature, paid with
// {R} and an artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5de63529-7dea-4018-bf46-afb4bc8b69fd",
		Name:         "Krark-Clan Ogre",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice an artifact: Target creature can't block this turn.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Effect:  effects.TargetCreatureCantBlockThisTurn("Krark-Clan Ogre — target creature can't block this turn"),
		}},
	})
}
