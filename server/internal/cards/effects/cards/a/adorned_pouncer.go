package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adorned Pouncer — Creature — Cat {1}{W}, 1/1 (Hour of Devastation, #5534):
//
//	"Double strike
//	 Eternalize {3}{W}{W} ({3}{W}{W}, Exile this card from your graveyard:
//	 Create a token that's a copy of it, except it's a 4/4 black Zombie Cat
//	 with no mana cost. Eternalize only as a sorcery.)"
//
// The keywords are printed data; eternalize is the shared Eternalize row
// (CR 702.129a, embalm.go), a graveyard ability that exiles the card as
// its cost and makes a 4/4 black Zombie token copy of it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "80b8ee3d-4174-4f47-a533-a3db1e703382",
		Name:            "Adorned Pouncer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Activated:       []effects.ActivatedAbility{effects.Eternalize("{3}{W}{W}")},
	})
}
