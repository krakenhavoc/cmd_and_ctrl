package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyshroud Elf — {1}{G} Creature — Elf Druid, 1/1 (Tempest, S82 OLD batch 770,
// #5628):
//
//	"{T}: Add {G}.
//	 {1}: Add {R} or {W}."
//
// Two mana abilities (CR 605.1a): the tap ability adds {G}; the second costs {1}
// paid from the pool or another source, and adds {R} or {W}, the colour chosen as
// it resolves (Boros Cluestone's choice, Bog Initiate's {1} cost). The second
// needs no tap, so it can be used while the Elf is tapped or summoning sick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84a666a0-e453-4b34-815c-48dbfc9b4f4b",
		Name:         "Skyshroud Elf",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{G}", Label: "{T}: Add {G}"},
			{Cost: effects.ManaAbilityCost{Mana: "{1}"}, Produced: "{R|W}", Label: "{1}: Add {R} or {W}"},
		},
	})
}
