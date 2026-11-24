package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectacle Mage — Creature — Bird Shaman {1}{U}{R}, 2/2 (Strixhaven: School
// of Mages, #5502):
//
//	"Flying
//	 Instant and sorcery spells you cast with mana value 5 or greater cost
//	 {1} less to cast."
//
// Thryx's discount narrowed to instants and sorceries: it reads the mana value
// the spell has on the stack, X included (CR 202.3e), and covers only spells
// its controller casts, never a copy (CR 707.10). The reduction changes only
// what is paid (CR 118.7), not the mana value.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "71920c05-7061-4ba5-8903-cffd3d2f85f4",
		Name:            "Spectacle Mage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Instant and sorcery spells you cast with mana value 5 or greater cost {1} less to cast.",
				effects.YourSpell(), effects.SpellManaValueAtLeast(5),
				func(q game.CostQuery) bool { return q.Card.IsInstant() || q.Card.IsSorcery() }),
		},
	})
}
