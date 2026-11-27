package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coma Veil — Enchantment — Aura {4}{U} (Shards of Alara, S82 OLD batch 517,
// #5588):
//
//	"Enchant artifact or creature
//	 Enchanted permanent doesn't untap during its controller's untap step."
//
// Bitter Chill's untap lock on an artifact or a creature host. The enchant
// clause is the spec the CR 704.5m legality check re-runs, so a host that
// stops being either takes the Aura with it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "c062bd12-34cf-4177-92d4-8ce91206989d",
		Name:                  "Coma Veil",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.TargetPermanent("enchant artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
	})
}
