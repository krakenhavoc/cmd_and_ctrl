package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surveillance Room — Land (Marvel Super Heroes, #4455):
//
//	"When this land enters, surveil 1.
//	 {T}: Add {C}.
//	 {1}, {T}: Add one mana of any color."
//
// Conduit Pylons' shape: the colorless mana ability, the filter ability
// and a surveil trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3648376-dc8b-409b-b2d1-c29e326a059c",
		Name:         "Surveillance Room",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.PainlessColorless(),
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Surveillance Room — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
