package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bog Witch — Creature — Human Spellshaper {2}{B}, 1/1:
//
//	"{B}, {T}, Discard a card: Add {B}{B}{B}."
//
// Skirge Familiar's discard component on a mana ability, with a {B}
// mana component and a tap beside it. The discard is a discard
// (EventDiscardCard fires, madness applies) and does not pause the
// mana ability (CR 605.3b). Not auto-tappable: which card to pitch is
// a decision, and the {B} is a mana component. The {T} makes it
// subject to summoning sickness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85573cba-07ae-4421-a167-a8569f85c0f7",
		Name:         "Bog Witch",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{B}", DiscardCards: effects.DiscardACard().DiscardCards},
			Produced: "{B}{B}{B}",
			Label:    "{B}, {T}, Discard a card: Add {B}{B}{B}",
		}},
	})
}
