package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ceta Disciple — Creature — Merfolk Wizard {U}, 1/1 (Apocalypse, S82 OLD,
// #5614):
//
//	"{R}, {T}: Target creature gets +2/+0 until end of turn.
//	 {G}, {T}: Add one mana of any color."
//
// A targeted tap ability and a mana ability with a mana cost (Rootrider Faun's
// filter shape). Both have {T} in the cost, so summoning sickness applies
// (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f085c2d-bd1b-4889-882f-367e0e970f44",
		Name:         "Ceta Disciple",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.OldB630DiscipleBoostRow("Ceta Disciple", "{R}", 2)},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{G}"},
			Produced: "{W|U|B|R|G}",
			Label:    "{G}, {T}: Add one mana of any color",
		}},
	})
}
