package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strongarm Thug — Creature — Human Mercenary {2}{B}, 1/1 (Mercadian Masques,
// S82 OLD batch 699, #5619):
//
//	"When this creature enters, you may return target Mercenary card from your
//	 graveyard to your hand."
//
// Moriok Scavenger's optional enters trigger with a Mercenary clause: with no
// Mercenary card in the graveyard there is no trigger (CR 603.3d), and one that
// leaves in response makes it fizzle (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4d46c1b-7742-4ac8-9794-b669ee4d2cc1",
		Name:         "Strongarm Thug",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Strongarm Thug — you may return target Mercenary card from your graveyard to your hand", effects.ReturnTargetedCardToHand),
				effects.TargetCardInGraveyard("target Mercenary card from your graveyard", effects.OfSubtype("Mercenary"), effects.YouOwn())),
				"Strongarm Thug — return the Mercenary card to your hand?"),
		},
	})
}
