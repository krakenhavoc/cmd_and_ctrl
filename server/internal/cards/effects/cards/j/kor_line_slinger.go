package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kor Line-Slinger — {1}{W} Creature — Kor Scout 0/1 (Rise of the Eldrazi, S82 OLD #5581):
//
//	{T}: Tap target creature with power 3 or less.
//
// Ephara's Warden's ability.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a852283-42d4-4197-91b9-5c952c7325ca",
		Name:         "Kor Line-Slinger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Tap target creature with power 3 or less.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature with power 3 or less", effects.PowerLE(3)),
			Effect:  effects.TapTheFirstLegalTarget,
		}},
	})
}
