package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boggart Birth Rite — Kindred Sorcery — Goblin {B} (Lorwyn, S82 OLD, #5593):
//
//	"Return target Goblin card from your graveyard to your hand."
//
// Ritual of Restoration's regrowth over any Goblin card (a Kindred Goblin
// included, so the clause reads the card's subtypes, not creature types), the
// target re-checked as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "030e9e3e-ce78-4c3b-bd83-969c07c70459",
		Name:         "Boggart Birth Rite",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target Goblin card in your graveyard", effects.HasSubtype("Goblin"), effects.YouOwn()),
		OnResolve:    effects.ReturnTargetFromGraveyardToHand,
	})
}
