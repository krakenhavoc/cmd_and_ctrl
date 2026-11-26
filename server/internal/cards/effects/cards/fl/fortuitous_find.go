package cardsfl

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Fortuitous Find — Sorcery {2}{B} (Kaladesh, #5539):
//
//	"Choose one or both —
//	 • Return target artifact card from your graveyard to your hand.
//	 • Return target creature card from your graveyard to your hand."
//
// Rise to Glory's "one or both" shape with return-to-hand bullets. Each bullet
// has its own target, chosen as the spell is cast, and the modes run in the
// order printed; a target that left the graveyard is skipped (CR 608.2b). An
// artifact creature card may be both targets, and is returned once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8cd3815d-92a6-4336-9da7-b18b1afd5460",
		Name:         "Fortuitous Find",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Return target artifact card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn()),
				effects.NeoB08ReturnTheModesTargetToHand),
			effects.ModeDoing("Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
				effects.NeoB08ReturnTheModesTargetToHand),
		),
	})
}
