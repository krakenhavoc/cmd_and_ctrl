package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire Soulcaller — Creature — Vampire Warlock {4}{B}, 3/2
// (Foundations, #4465):
//
//	"Flying
//	 This creature can't block.
//	 When this creature enters, return target creature card from your
//	 graveyard to your hand."
//
// Vampire Interloper's body with Raise Dead on entry. The return isn't
// optional, so the trigger has no "you may": the creature card chosen
// as the trigger goes on the stack returns to your hand if it is still
// in your graveyard as the trigger resolves (CR 608.2b). With no
// creature card in your graveyard, the trigger is removed for want of a
// target (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bbea8ade-a641-4132-b306-f39fe2f14311",
		Name:            "Vampire Soulcaller",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Vampire Soulcaller — return target creature card from your graveyard to your hand",
					effects.B27ReturnChosenGraveyardCardToHand),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn())),
		},
	})
}
