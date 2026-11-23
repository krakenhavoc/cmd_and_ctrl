package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steadfast Sentinel — Creature — Human Cleric {3}{W}, 2/3
// (Hour of Devastation, #5534):
//
//	"Vigilance
//	 Eternalize {4}{W}{W} ({4}{W}{W}, Exile this card from your graveyard:
//	 Create a token that's a copy of it, except it's a 4/4 black Zombie Human
//	 Cleric with no mana cost. Eternalize only as a sorcery.)"
//
// The keywords are printed data; eternalize is the shared Eternalize row
// (CR 702.129a, embalm.go), a graveyard ability that exiles the card as
// its cost and makes a 4/4 black Zombie token copy of it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "38bfc0bc-66a6-4d9e-9186-5ccaef773014",
		Name:            "Steadfast Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []effects.ActivatedAbility{effects.Eternalize("{4}{W}{W}")},
	})
}
