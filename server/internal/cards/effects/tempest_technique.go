package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tempest Technique — Enchantment — Aura {3}{W} (Tarkir: Dragonstorm
// Commander):
//
//	"Storm (When you cast this spell, copy it for each spell cast before it
//	 this turn. You may choose new targets for the copies. Copies become
//	 tokens.)
//	 Enchant creature you control
//	 Enchanted creature gets +1/+1 for each enchantment you control."
//
// Amphibian Downpour's storm on an Aura (each copy is an Aura spell that
// targets its own creature, and becomes a token as it resolves) with
// Ethereal Armor's count, re-read on every layer recompute: every
// enchantment its controller has, the Aura and its token copies included.
// The clause is "enchant creature you control", judged at announce and again
// as state-based actions look (CR 704.5m).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e86bd848-d712-47c5-a82b-8fd5ce1de76a",
		Name:         "Tempest Technique",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered:    []game.TriggeredAbility{Storm()},
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, enchantmentsControlledBy),
		},
	})
}
