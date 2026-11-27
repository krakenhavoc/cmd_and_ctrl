package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Déjà Vu — Sorcery {2}{U} (Portal Second Age, S82 older sets batch 735, #5625):
//
//	"Return target sorcery card from your graveyard to your hand."
//
// The target is a sorcery card in your graveyard as it is chosen and as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7408b9c5-7266-4627-be4e-b691cf5c622c",
		Name:         "Déjà Vu",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target sorcery card from your graveyard", effects.Sorcery(), effects.YouOwn()),
		OnResolve:    effects.ReturnTargetFromGraveyardToHand,
	})
}
