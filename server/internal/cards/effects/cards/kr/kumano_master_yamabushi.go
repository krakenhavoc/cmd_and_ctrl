package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kumano, Master Yamabushi — Legendary Creature — Human Shaman
// {3}{R}{R}, 4/4:
//
//	"{1}{R}: Kumano deals 1 damage to any target.
//	 If a creature dealt damage by Kumano this turn would die, exile it
//	 instead."
//
// The ping is Ballista Watcher's. The replacement reads the turn's record
// of which objects Kumano damaged (#1946, ExileIfDealtDamageByThisWouldDie),
// so combat damage counts as well as the ping, damage prevented to 0
// does not, and a creature that was damaged and then flickered is a new
// object. It is a static, so it works while Kumano is on the battlefield,
// including on a creature that dies at the same time as Kumano.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d776c7b-4ca1-48cd-88c7-dfbbe5a56a0b",
		Name:         "Kumano, Master Yamabushi",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: Kumano deals 1 damage to any target.",
			Cost:    effects.ManaCost("{1}{R}"),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
		Replacements: []game.ReplacementEffect{
			effects.ExileIfDealtDamageByThisWouldDie("Kumano, Master Yamabushi: exile it instead of dying"),
		},
	})
}
