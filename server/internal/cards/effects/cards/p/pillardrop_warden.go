package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pillardrop Warden — Creature — Spirit Dwarf {3}{R}, 1/5 (Strixhaven: School
// of Mages, #5502):
//
//	"Reach
//	 {2}, {T}, Sacrifice this creature: Return target instant or sorcery card
//	 from your graveyard to your hand. Activate only as a sorcery."
//
// Sacrificing the Warden is a cost paid as the ability is activated (CR
// 602.2b); the card returns to its owner's hand on resolution if it is still an
// instant or sorcery card in a graveyard of yours (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6ee2a7ba-a676-4d05-b214-7db52e41afe7",
		Name:            "Pillardrop Warden",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}, {T}, Sacrifice this creature: Return target instant or sorcery card from your graveyard to your hand. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
			Effect:       effects.ReturnTargetCardToHand,
		}},
	})
}
