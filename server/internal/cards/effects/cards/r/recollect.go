package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Recollect — Sorcery {2}{G} (Ravnica: City of Guilds, #5600):
//
//	"Return target card from your graveyard to your hand."
//
// Regrowth's body: any card in its caster's graveyard returns to hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b38fafb-3999-46dc-928b-1677625c1943",
		Name:         "Recollect",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn()),
		OnResolve:    effects.ReturnTargetFromGraveyardToHand,
	})
}
