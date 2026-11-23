package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Proven Combatant — Creature — Human Warrior {U}, 1/1
// (Hour of Devastation, #5534):
//
//	"Eternalize {4}{U}{U} ({4}{U}{U}, Exile this card from your graveyard:
//	 Create a token that's a copy of it, except it's a 4/4 black Zombie Human
//	 Warrior with no mana cost. Eternalize only as a sorcery.)"
//
// The keywords are printed data; eternalize is the shared Eternalize row
// (CR 702.129a, embalm.go), a graveyard ability that exiles the card as
// its cost and makes a 4/4 black Zombie token copy of it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24ce3dbe-2b61-43a6-ab16-32b37e385127",
		Name:         "Proven Combatant",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Eternalize("{4}{U}{U}")},
	})
}
