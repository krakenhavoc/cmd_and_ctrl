package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tyrranax Rex — Creature — Phyrexian Dinosaur {4}{G}{G}{G}, 8/8 (March of
// the Machine: Phyrexia: All Will Be One):
//
//	"This spell can't be countered.
//	 Trample, ward {4}, haste
//	 Toxic 4"
//
// The uncounterable line is a spell property, read while the card is on the
// stack (CR 101.2). Trample, haste and toxic are the engine's keywords; ward
// {4} is the ward trigger over a mana payment (CR 702.21).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6e42da0c-151e-468d-91cb-5a5b117a9298",
		Name:            "Tyrranax Rex",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"trample", "ward", "haste", "toxic 4"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{4}"), "Tyrranax Rex — ward {4}"),
		},
	})
}
