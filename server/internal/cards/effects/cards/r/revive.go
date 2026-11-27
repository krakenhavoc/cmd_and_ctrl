package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Revive — Sorcery {1}{G} (Magic 2013, S82 OLD, #5568):
//
//	"Return target green card from your graveyard to your hand."
//
// Ritual of Restoration's return, over a green card. Colour is the card's own as
// it lies in the graveyard (a colourless or multicoloured card with green in its
// colours counts). A target gone in response does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51991d28-137e-4690-a674-d852b904bd1a",
		Name:         "Revive",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target green card in your graveyard", effects.OfColor("G"), effects.YouOwn()),
		OnResolve:    effects.ReturnTargetFromGraveyardToHand,
	})
}
