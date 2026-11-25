package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mists of Littjara — Enchantment — Aura {1}{U} (Kaldheim, #5505):
//
//	"Flash
//	 Enchant creature or Vehicle
//	 Enchanted creature gets -3/-0."
//
// "Enchant creature or Vehicle" is the permanent spec the CR 704.5m
// check re-runs, so an uncrewed Vehicle (not a creature) keeps it.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c0162a6c-12d9-4b84-811a-ee6e39401863",
		Name:            "Mists of Littjara",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.TargetPermanent("enchant creature or Vehicle", effects.Or(effects.Creature(), effects.HasSubtype("Vehicle"))),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static:          []game.StaticAbility{effects.PumpAttached(-3, 0)},
	})
}
