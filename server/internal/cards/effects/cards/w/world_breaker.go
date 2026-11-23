package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// World Breaker — Creature — Eldrazi {6}{G}, 5/7 (Foundations Commander):
//
//	"Devoid (This card has no color.)
//	 When you cast this spell, exile target artifact, enchantment, or
//	 land.
//	 Reach
//	 {2}{C}, Sacrifice a land: Return this card from your graveyard to
//	 your hand."
//
// Smothering Abomination's devoid keyword, Abundant Maw's cast trigger
// (it resolves above the spell, so it exiles even if the Breaker is
// countered) and Blessed Ghoul's graveyard activation. The {C} is
// colourless mana specifically, and the land is sacrificed as a cost.
//
// No simplification.
func init() {
	cast := effects.WhenYouCastThisSpell("World Breaker — exile target artifact, enchantment, or land", effects.ExileFirstTarget)
	cast.Targets = effects.TargetPermanent("target artifact, enchantment, or land", effects.Or(effects.Artifact(), effects.Enchantment(), effects.Land()))
	effects.Register(effects.Spec{
		OracleID:        "deeb7e22-b207-42de-b9d4-f790403d52a9",
		Name:            "World Breaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"devoid", "reach"},
		Triggered:       []game.TriggeredAbility{cast},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{C}, Sacrifice a land: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{C}"), effects.B08SacrificeALand()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
