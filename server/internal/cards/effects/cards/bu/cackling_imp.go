package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cackling Imp — Creature — Imp {2}{B}{B}, 2/2 (Fifth Dawn, #5604):
//
//	"Flying
//	 {T}: Target player loses 1 life."
//
// The life loss is TargetPlayerLosesLife (Sweettooth Witch's body), a life
// change that names the Imp and no damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "adc8c55c-8fbe-4e07-b1be-8cb796db3d8a",
		Name:            "Cackling Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target player loses 1 life.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.Do(effects.TargetPlayerLosesLife{N: 1}),
		}},
	})
}
