package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elephant Grass — Enchantment {G} (Visions, S82 older sets batch 789, #5631):
//
//	"Cumulative upkeep {1}
//	 Black creatures can't attack you.
//	 Nonblack creatures can't attack you unless their controller pays {2} for
//	 each creature they control that's attacking you."
//
// Two clauses over the same attackers, split by colour: a black creature is
// barred outright (an attack restriction, CR 508.1c), and every other
// attacker costs {2} (the attack tax of Propaganda, CR 508.1h). The tax
// applies only to attackers that are not black as they are declared, read
// through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53cc4b93-538d-4193-8bbb-5b73d3ce9d79",
		Name:         "Elephant Grass",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.CreaturesMatchingCantAttackYou("Black creatures can't attack you.",
				game.PermanentQuery{Types: []string{"creature"}, Colors: []string{"B"}}),
		},
		AttackTaxes: []game.AttackTax{{
			Label: "Nonblack creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.",
			AppliesTo: func(q game.AttackTaxQuery) bool {
				src := q.Source
				return !q.Attacker.HasColor(src.TextColor("B"))
			},
			ManaCost: func(game.AttackTaxQuery) string { return "{2}" },
		}},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Elephant Grass — cumulative upkeep {1}", "{1}"),
		},
	})
}
