package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Patron of the Moon — Legendary Creature — Spirit {5}{U}{U}, 5/4
// (Betrayers of Kamigawa, #5602):
//
//	"Moonfolk offering (You may cast this spell any time you could cast an
//	 instant by sacrificing a Moonfolk and paying the difference in mana
//	 costs between this and the sacrificed Moonfolk. Mana cost includes
//	 color.)
//	 Flying
//	 {1}: Put up to two land cards from your hand onto the battlefield
//	 tapped."
//
// Moonfolk offering is Offering (offering.go, CR 702.48). The ability is
// Lt. Arex's put with a limit of two: as it resolves its controller picks
// none, one or two land cards from their hand, and they enter tapped as
// one batch. They are put, not played, so no land drop is used.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e514546a-ca03-48ba-90e1-7a0e145137f2",
		Name:            "Patron of the Moon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{effects.Offering("Moonfolk", effects.Subtype("Moonfolk"))},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Put up to two land cards from your hand onto the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{1}"),
			Effect: effects.Do(effects.PutFromHandOntoBattlefield{
				Match:    effects.Land(),
				Tapped:   true,
				Optional: true,
				Max:      2,
				Label:    "Patron of the Moon — put up to two land cards from your hand onto the battlefield tapped",
			}),
		}},
	})
}
