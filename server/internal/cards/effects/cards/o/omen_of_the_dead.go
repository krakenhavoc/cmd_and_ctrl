package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "0aa9ee99-69a8-452b-a0d5-001384bdff3c",
		Name:            "Omen of the Dead",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Omen of the Dead — return target creature card from your graveyard to your hand",
					effects.B27ReturnChosenGraveyardCardToHand),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice this enchantment: Scry 2.",
			Purpose: game.Purpose{Scry: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.Scry{N: 2}),
		}},
	})
}
