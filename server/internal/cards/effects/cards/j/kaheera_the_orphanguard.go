package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kaheera, the Orphanguard — Legendary Creature — Cat Beast
// {1}{G/W}{G/W}, 3/2:
//
//	"Companion — Each creature card in your starting deck is a Cat,
//	 Elemental, Nightmare, Dinosaur, or Beast card. (...)
//	 Vigilance
//	 Each other creature you control that's a Cat, Elemental, Nightmare,
//	 Dinosaur, or Beast gets +1/+1 and has vigilance."
//
// The two statics read the layered view of each creature's types, so a
// changeling counts. Companion is the deck check's business
// (internal/deck companion registry, ADR 0172) and the {3} special action
// the engine's, not a card ability.
//
// No simplification.
func init() {
	f := effects.TribeFilter{
		Tribes:    []string{"Cat", "Elemental", "Nightmare", "Dinosaur", "Beast"},
		Others:    true,
		YoursOnly: true,
	}
	effects.Register(effects.Spec{
		OracleID:        "35428c77-9731-4d4d-8f75-e939a491cf86",
		Name:            "Kaheera, the Orphanguard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(f, 1, 1),
			effects.TribalKeywordGrant(f, "vigilance"),
		},
	})
}
