package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyre of Heroes — Artifact {2} (Kaldheim):
//
//	"{2}, {T}, Sacrifice a creature: Search your library for a creature card
//	 that shares a creature type with the sacrificed creature and has mana
//	 value equal to 1 plus that creature's mana value. Put that card onto the
//	 battlefield, then shuffle. Activate only as a sorcery."
//
// Birthing Pod's shape with a creature-type rider: the creature is paid at
// announce, and the body reads it back (S82OldB061PyreOfHeroesSearch). The
// fetched creature enters untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "691ba449-40e1-4976-8dab-9b1741ea37b8",
		Name:         "Pyre of Heroes",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}, {T}, Sacrifice a creature: Search your library for a creature card that shares a creature type with the sacrificed creature and has mana value equal to 1 plus that creature's mana value. Put that card onto the battlefield, then shuffle. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeACreature()),
			SorcerySpeed: true,
			Effect:       effects.S82OldB061PyreOfHeroesSearch,
		}},
	})
}
