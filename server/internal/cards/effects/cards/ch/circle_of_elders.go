package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Circle of Elders — Creature — Human Shaman {2}{G}{G}, 2/4 (Dragons of Tarkir,
// S82 older sets):
//
//	"Vigilance
//	 Formidable — {T}: Add {C}{C}{C}. Activate only if creatures you control have
//	 total power 8 or greater."
//
// A mana ability (CR 605.1a) with the formidable total as its "activate only if"
// (CR 602.5b), read as it is activated through the layers, the Elder's own power
// included; the engine refuses the tap while it is summoning sick (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f720f80a-74bc-4681-ad4f-d33152ff988b",
		Name:            "Circle of Elders",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:      effects.ManaAbilityCost{Tap: true},
			Produced:  "{C}{C}{C}",
			Condition: effects.OldB324Formidable,
			Label:     effects.OldB324FormidableLabel("{T}: Add {C}{C}{C}."),
		}},
	})
}
