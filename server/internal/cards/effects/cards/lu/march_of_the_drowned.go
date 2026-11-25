package cardslu

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// March of the Drowned — Sorcery {B} (Ixalan, S82 older sets):
//
//	"Choose one —
//	 • Return target creature card from your graveyard to your hand.
//	 • Return two target Pirate cards from your graveyard to your hand."
//
// The second bullet needs exactly two targets (CR 601.2c): with fewer than two
// Pirate cards in the graveyard it cannot be chosen. Each target is rechecked as
// the spell resolves and a card that has left is skipped (CR 608.2b).
//
// No simplification.
func init() {
	back := effects.EoeABCharmReturn
	effects.Register(effects.Spec{
		OracleID:     "e9e55768-ed62-4998-9922-dbcdeab12aee",
		Name:         "March of the Drowned",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()), back),
			effects.ModeDoing("Return two target Pirate cards from your graveyard to your hand.",
				effects.TargetCardInGraveyard("two target Pirate cards from your graveyard", effects.YouOwn(), effects.Subtype("Pirate")).WithCount(2, 2), back),
		),
	})
}
