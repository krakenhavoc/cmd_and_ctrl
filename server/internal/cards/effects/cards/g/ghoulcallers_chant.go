package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghoulcaller's Chant — {B} Sorcery (Innistrad, S82 OLD b447, #5572):
//
//	"Choose one —
//	 • Return target creature card from your graveyard to your hand.
//	 • Return two target Zombie cards from your graveyard to your hand."
//
// Unbury's shape: the second bullet is two target clauses, distinct because
// "two target" is one instance of the word (CR 115.3), each judged again as the
// spell resolves; a card that has left the graveyard by then is skipped and the
// other is still returned (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b685fff3-f1b4-42d5-9b7b-e0885bcc92d2",
		Name:         "Ghoulcaller's Chant",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
				effects.EoeABCharmReturn),
			effects.ModeDoing("Return two target Zombie cards from your graveyard to your hand.",
				effects.Clauses(
					effects.TargetCardInGraveyard("target Zombie card from your graveyard", effects.HasSubtype("Zombie"), effects.YouOwn()),
					effects.Distinct(effects.TargetCardInGraveyard("another target Zombie card from your graveyard", effects.HasSubtype("Zombie"), effects.YouOwn())),
				),
				effects.EoeABCharmReturn),
		),
	})
}
