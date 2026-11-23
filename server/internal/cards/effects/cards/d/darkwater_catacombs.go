package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darkwater Catacombs — Land (EDHREC rank 323):
//
//	"{1}, {T}: Add {U}{B}."
//
// An Odyssey "tap-and-a-tax" dual: no colourless option at all, one
// ability, {1} to filter into a fixed two colours. Mechanically the
// Signet cycle's mana component (`ManaAbilityCost.Mana`, closed by
// #356) on a land instead of an artifact, paid like a Signet's (see
// signets.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4869a530-757f-4364-8d8e-4dc8001f433c",
		Name:         "Darkwater Catacombs",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{U}{B}",
				Label:    "{1}, {T}: Add {U}{B}",
			},
		},
	})
}
