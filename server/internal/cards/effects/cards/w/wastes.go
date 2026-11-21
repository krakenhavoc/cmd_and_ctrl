package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wastes — Basic Land (Oath of the Gatewatch, #5545):
//
//	"{T}: Add {C}."
//
// A basic land with no basic land type, so no intrinsic mana ability (CR
// 305.6): its {C} ability is the printed one, and the Spec declares it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "05d24b0c-904a-46b6-b42a-96a4d91a0dd4",
		Name:          "Wastes",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
	})
}
