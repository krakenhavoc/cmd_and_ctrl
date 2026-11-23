package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Identity Echo — Enchantment {2}{R} (Reality Fracture):
//
//	"{3}{R}: Exile target creature or planeswalker you control. Reveal
//	 cards from the top of your library until you reveal a creature or
//	 planeswalker card. Put that card onto the battlefield and the rest
//	 on the bottom of your library in a random order. Activate only as a
//	 sorcery."
//
// The same exile-and-reveal sentence as Jace, Multiverse Architect's −3
// (rfExileTargetThenRevealCreatureOrPlaneswalker). A token exiled this
// way ceases to exist (CR 111.7); the reveal still happens.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4dbf8c43-d7c5-4037-875e-6accd87aaf6f",
		Name:         "Identity Echo",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{R}: Exile target creature or planeswalker you control. Reveal cards from the top of your library until you reveal a creature or planeswalker card. Put that card onto the battlefield and the rest on the bottom of your library in a random order. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{R}"),
			Targets:      effects.TargetPermanent("target creature or planeswalker you control", effects.Or(effects.Creature(), effects.Planeswalker()), effects.YouControl()),
			SorcerySpeed: true,
			Effect:       effects.RfExileTargetThenRevealCreatureOrPlaneswalker,
		}},
	})
}
