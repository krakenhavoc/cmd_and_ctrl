package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Team Pennant — Artifact — Equipment {1} (Strixhaven: School of Mages,
// #5502):
//
//	"Equipped creature gets +1/+1 and has vigilance and trample.
//	 Equip creature token {1}
//	 Equip {3}"
//
// Batterbone's attached pump and keywords, and Brotherhood Regalia's two equip
// costs: the cheaper one is restricted to a creature token you control and the
// plain one takes any creature you control. Both are sorcery-speed equip
// abilities (CR 702.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fa51915-7c0b-4018-802a-90fd431f0036",
		Name:         "Team Pennant",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("vigilance", "trample"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipOnlyAbility("Equip creature token {1}", "{1}",
				effects.TargetCreature("target creature token you control", effects.YouControl(), effects.IsTokenPredicate())),
			effects.EquipAbility("{3}"),
		},
	})
}
