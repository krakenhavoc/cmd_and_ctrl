package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beacon Hawk — Creature — Bird {1}{W}, 1/1 (Dissension, #5598):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, you may untap
//	 target creature.
//	 {W}: This creature gets +0/+1 until end of turn."
//
// The untap is Mistblade Shinobi's optional, targeted combat-damage
// trigger with untapTheTarget; the pump is thisGetsRow.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "edf7c98a-c2ed-4ba2-ac08-e83db0d968cb",
		Name:            "Beacon Hawk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer("Beacon Hawk — untap target creature", effects.UntapTheTarget),
				effects.TargetCreature("target creature")), "Beacon Hawk — untap the target creature?"),
		},
		Activated: []effects.ActivatedAbility{effects.ThisGetsRow("{W}", 0, 1, "Beacon Hawk")},
	})
}
