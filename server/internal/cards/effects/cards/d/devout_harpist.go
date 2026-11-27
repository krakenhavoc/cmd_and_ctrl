package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Devout Harpist — Creature — Human {W} 1/1:
//
//	"{T}: Destroy target Aura attached to a creature."
//
// The Aura must be attached to a creature when it is targeted and again when
// the ability resolves (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f57901c2-80c6-49d6-8de3-c96aaeb72bbf",
		Name:         "Devout Harpist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target Aura attached to a creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.BattlefieldSpec("permanent", "target Aura attached to a creature", effects.OldB717AuraOnACreature),
			Effect:  effects.DestroyFirstLegalCardTarget,
		}},
	})
}
