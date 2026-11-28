package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ritual of Subdual — Enchantment {4}{G}{G} (Ice Age, S82 OLD batch 839, #5636):
//
//	"Cumulative upkeep {2}
//	 If a land is tapped for mana, it produces colorless mana instead of any other
//	 type."
//
// Infernal Darkness' replacement with colorless as the type: every mana a tapped
// land would produce is {C}, the amount unchanged.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0587218-206e-41ed-af3c-f06b7a668e90",
		Name:         "Ritual of Subdual",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.CumulativeUpkeep("Ritual of Subdual — cumulative upkeep {2}", "{2}")},
		Replacements: []game.ReplacementEffect{
			effects.OldB839LandsProduce("Ritual of Subdual: a land tapped for mana produces colorless mana", "C"),
		},
	})
}
