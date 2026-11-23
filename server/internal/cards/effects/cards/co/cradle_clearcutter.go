package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cradle Clearcutter — Artifact Creature — Golem {6}, 3/6 (The Brothers'
// War, #5656): "Prototype {2}{G} — 1/3", "{T}: Add an amount of {G} equal
// to this creature's power."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out.
// The mana ability reads the Clearcutter's power as it is activated: one
// {G} from a prototyped Clearcutter, three from one cast normally, more
// or fewer with counters and continuous effects.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "225f53ef-da42-47a0-b773-d2aad3338c72",
		Name:         "Cradle Clearcutter",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			Label:          "Add an amount of {G} equal to this creature's power",
			ProducedAmount: effects.TdmGreenEqualToPower,
		}},
	})
}
