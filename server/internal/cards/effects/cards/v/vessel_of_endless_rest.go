package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vessel of Endless Rest — {3} Artifact:
//
//	"When this artifact enters, put target card from a graveyard on the
//	 bottom of its owner's library.
//	 {T}: Add one mana of any color."
//
// The target is any card in any graveyard; it goes to the bottom of its
// owner's library (CR 400.3), no shuffle. Transplant Theorist's tuck.
func init() {
	t := effects.Targeting(
		effects.WhenThisEnters("Vessel of Endless Rest — put target card from a graveyard on the bottom of its owner's library",
			effects.S82One13TuckTargetToLibraryBottom),
		effects.TargetCardInGraveyard("target card from a graveyard"),
	)
	effects.Register(effects.Spec{
		OracleID:      "8031285b-2add-42da-aa97-03c3f2fff226",
		Name:          "Vessel of Endless Rest",
		Completeness:  effects.CompletenessFull,
		Triggered:     []game.TriggeredAbility{t},
		ManaAbilities: []effects.ManaAbility{effects.Cmdr2B01TapForAnyColor()},
	})
}
