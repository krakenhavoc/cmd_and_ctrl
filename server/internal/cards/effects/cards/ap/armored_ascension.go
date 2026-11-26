package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Armored Ascension — Enchantment — Aura {3}{W} (Magic 2011 / Planeswalker
// Decks, #5579):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 for each Plains you control and has
//	 flying."
//
// The bonus is counted live in layer 7c (CR 613.4c), by the Aura's controller
// ("you"), so a Plains played or lost changes it in the same beat; the
// flying is a layer-6 grant (CR 613.1f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "916db499-7f0a-45f7-8abe-94818a53e494",
		Name:         "Armored Ascension",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, effects.S82OldB476PlainsYouControl),
			effects.GrantToAttached("flying"),
		},
	})
}
