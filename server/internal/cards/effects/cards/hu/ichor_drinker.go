package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ichor Drinker — Creature — Phyrexian Vampire {B}, 1/1 (March of the
// Machine):
//
//	"Lifelink
//	 {B}, Exile this card from your graveyard: Incubate 2. Activate
//	 only as a sorcery."
//
// The ability functions only from the graveyard, and its cost exiles
// the card itself (the shape embalm uses), so the card is gone before
// the Incubator is made. The incubate (CR 701.53a, ADR 0185) is the
// activator's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "470a8ee9-1870-44d8-8ade-64ec8084b9f7",
		Name:            "Ichor Drinker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{B}, Exile this card from your graveyard: Incubate 2. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{B}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.Do(effects.Incubate{N: 2}),
		}},
	})
}
