package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omen of the Dead — Enchantment {B} (Theros Beyond Death, #5514):
//
//	"Flash
//	 When this enchantment enters, return target creature card from
//	 your graveyard to your hand.
//	 {2}{B}, Sacrifice this enchantment: Scry 2."
//
// Omen of the Sea's frame: flash, an enters trigger, and a sacrifice-
// for-scry-2 activated ability, each Omen paying {2} and one mana of its
// own color. The enters trigger is Vampire Soulcaller's return of a
// creature card from its controller's graveyard to hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0aa9ee99-69a8-452b-a0d5-001384bdff3c",
		Name:            "Omen of the Dead",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Omen of the Dead — return target creature card from your graveyard to your hand",
					b27ReturnChosenGraveyardCardToHand),
				TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice this enchantment: Scry 2.",
			Purpose: game.Purpose{Scry: 2, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{B}"), SacrificeThis()),
			Effect:  Do(Scry{N: 2}),
		}},
	})
}
