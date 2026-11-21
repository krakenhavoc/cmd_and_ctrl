package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bedhead Beastie — Creature — Beast {4}{R}{R}, 5/6 (Duskmourn, #4466):
//
//	"Menace
//	 Mountaincycling {2} ({2}, Discard this card: Search your library
//	 for a Mountain card, reveal it, put it into your hand, then
//	 shuffle.)"
//
// Mountaincycling is typecycling (CR 702.29e) for the Mountain subtype.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7143d82a-931a-4918-bb14-37bc5b9a4d99",
		Name:            "Bedhead Beastie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Mountaincycling", "{2}", "a Mountain card", effects.IsLandWithSubtype("Mountain")),
		},
	})
}
