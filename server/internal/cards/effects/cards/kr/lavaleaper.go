package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lavaleaper — Creature — Elemental {3}{R}, 4/4 (Lorwyn Eclipsed,
// #4458):
//
//	"All creatures have haste.
//	 Whenever a player taps a basic land for mana, that player adds one
//	 mana of any type that land produced."
//
// The haste is Mass Hysteria's layer 6 grant over every creature on the
// battlefield. The second ability is Heartbeat of Spring's triggered
// mana ability (it does not use the stack, the ruling) narrowed to a
// land with the basic supertype, as it was when it produced.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a107fb84-6874-4964-a729-4425dad98330",
		Name:         "Lavaleaper",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(effects.B29AllCreatures, "haste"),
		},
		ManaTriggers: []game.ManaTrigger{lavaleaperBasicLandTrigger()},
	})
}

// lavaleaperBasicLandTrigger is "whenever a player taps a basic land for
// mana, that player adds one mana of any type that land produced".
func lavaleaperBasicLandTrigger() game.ManaTrigger {
	t := effects.WheneverAPlayerTapsALandForMana(
		"Lavaleaper — add one mana of any type that basic land produced",
		effects.AddsOneManaOfAnyTypeProduced())
	anyLand := t.AppliesTo
	t.AppliesTo = func(prod game.ManaProduced, source *game.Card, g *game.Game) bool {
		return anyLand(prod, source, g) && prod.Source.HasSupertype("basic")
	}
	return t
}
