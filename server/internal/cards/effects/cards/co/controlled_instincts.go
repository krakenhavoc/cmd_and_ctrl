package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Controlled Instincts — Enchantment — Aura {U} (Conflux):
//
//	"Enchant red or green creature
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Bind the Monster's untap restriction. "Red or green" is read through the
// creature's colours (a colour word in the clause, CR 612.2), so a creature
// that stops being either falls off as a state-based action (CR 704.5m). An
// untap effect other than the untap step still untaps it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "7d56d2e2-e83b-4505-af8e-1fd207e161af",
		Name:                  "Controlled Instincts",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(effects.Or(effects.OfColor("R"), effects.OfColor("G"))),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
	})
}
