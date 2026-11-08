package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "bbea8ade-a641-4132-b306-f39fe2f14311",
		Name:            "Vampire Soulcaller",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Vampire Soulcaller — return target creature card from your graveyard to your hand",
					b27ReturnChosenGraveyardCardToHand),
				TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn())),
		},
	})
}
