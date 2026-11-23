package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shepherding Spirits — Creature — Spirit {4}{W}{W}, 4/5 (Duskmourn:
// House of Horror, #4466):
//
//	"Flying
//	 Plainscycling {2} ({2}, Discard this card: Search your library for
//	 a Plains card, reveal it, put it into your hand, then shuffle.)"
//
// Plainscycling is the shared Typecycling constructor (CR 702.29e), as on
// Daggermaw Megalodon: any card with the Plains subtype, basic or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e8ff3e99-eca9-44a5-baac-99ee34df3c36",
		Name:            "Shepherding Spirits",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Plainscycling", "{2}", "a Plains card", effects.IsLandWithSubtype("Plains")),
		},
	})
}
