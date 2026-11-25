package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deathless Ancient — Creature — Vampire Knight {4}{B}{B}, 4/4 (Ixalan, S82
// older sets):
//
//	"Flying
//	 Tap three untapped Vampires you control: Return this card from your
//	 graveyard to your hand."
//
// An ability that works from the graveyard (Earthquake Dragon's zone and
// return), paid by tapping three untapped Vampires (not the {T} symbol, CR
// 302.6: a Vampire that arrived this turn may pay). The card itself is in the
// graveyard, so it never pays for itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "990dde55-8ee3-41ad-874b-60864ec28ff4",
		Name:            "Deathless Ancient",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap three untapped Vampires you control: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  3,
				Filter: effects.TargetPermanent("three untapped Vampires you control", effects.Subtype("Vampire")),
				Label:  "three untapped Vampires you control",
			}},
			Zones:  []game.ZoneKind{game.ZoneGraveyard},
			Effect: effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
