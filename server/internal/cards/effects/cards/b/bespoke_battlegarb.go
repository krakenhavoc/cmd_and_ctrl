package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bespoke Battlegarb — Artifact — Equipment {1}{R} (Wilds of Eldraine,
// #5164):
//
//	"Equipped creature gets +2/+0.
//	 Celebration — At the beginning of combat on your turn, if two or
//	 more nonland permanents entered the battlefield under your control
//	 this turn, attach this Equipment to up to one target creature you
//	 control.
//	 Equip {2}"
//
// Celebration is an intervening if (celebration.go, CR 603.4): checked
// as combat begins and again as the ability resolves. The Battlegarb
// entering this turn counts toward it. The attach is the equip
// ability's own resolution (AttachSourceToTarget): no target, a target
// that is no longer legal, or an Equipment that has left or come back
// as a new object attaches nothing (CR 701.3b, CR 400.7).
//
// No simplification.
func init() {
	attach := effects.AtBeginningOfYourCombatIfCelebration("Bespoke Battlegarb — attach this Equipment to up to one target creature you control",
		effects.AttachSourceToTarget)
	attach.Targets = effects.TargetCreature("up to one target creature you control", effects.YouControl()).WithCount(0, 1)
	effects.Register(effects.Spec{
		OracleID:     "b2ff15b8-1be4-4f94-b4cc-aa959a50288e",
		Name:         "Bespoke Battlegarb",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered:    []game.TriggeredAbility{attach},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
