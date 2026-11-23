package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Autumn Willow, Harmony — {3}{G}{G} Legendary Creature — Avatar 3/3
// (Mystery Booster Commander, S82 GAP batch 3, #5948):
//
//	"Hexproof
//	 When Autumn Willow enters, create a 1/1 green Forest Dryad land
//	 creature token. (It's affected by summoning sickness.)
//	 Whenever you tap a land creature for mana, add an additional {G}."
//
// The Dryad is the printed Forest Dryad token (Token Land Creature —
// Forest Dryad), and a Forest taps for {G} by its land type. The extra
// {G} is a triggered mana ability (CR 605.1b, Spec.ManaTriggers) for any
// permanent of yours that is a land and a creature when it is tapped for
// mana, read through the layers, so an animated land counts; it is
// "whenever you tap", so it does not need the mana to be {G}.
//
// No simplification.
func init() {
	trigger := effects.WheneverYouTapALandForMana(
		"Autumn Willow, Harmony — add an additional {G}", effects.AddsFixedMana("{G}"))
	land := trigger.AppliesTo
	trigger.AppliesTo = func(prod game.ManaProduced, source *game.Card, g *game.Game) bool {
		return prod.Source.IsCreature() && land(prod, source, g)
	}
	effects.Register(effects.Spec{
		OracleID:        "01f9cfd9-47da-4dbe-a21e-53eccad3864a",
		Name:            "Autumn Willow, Harmony",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Autumn Willow, Harmony — create a 1/1 green Forest Dryad land creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Forest Dryad"), N: 1})),
		},
		ManaTriggers: []game.ManaTrigger{trigger},
	})
}
