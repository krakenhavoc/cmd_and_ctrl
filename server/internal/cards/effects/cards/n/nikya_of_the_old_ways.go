package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nikya of the Old Ways — Legendary Creature — Centaur Druid {3}{R}{G}, 5/5
// (Ravnica Allegiance):
//
//	"You can't cast noncreature spells.
//	 Whenever you tap a land for mana, add one mana of any type that land
//	 produced."
//
// The restriction is Nullhide Ferox's (it binds the controller only); the
// second ability is Mirari's Wake's triggered mana ability (CR 605.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebf3fd80-9574-489b-9428-f4251834bdc4",
		Name:         "Nikya of the Old Ways",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			effects.YouCantCastUnless("You can't cast noncreature spells.", effects.Not(effects.Creature()), nil),
		},
		ManaTriggers: []game.ManaTrigger{
			effects.WheneverYouTapALandForMana(
				"Nikya of the Old Ways — add one mana of any type that land produced",
				effects.AddsOneManaOfAnyTypeProduced()),
		},
	})
}
