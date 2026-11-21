package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Winged Hive Tyrant — Creature — Tyranid {3}{U}{R}, 4/4 (Warhammer 40,000
// Commander, S82 GAP batch 195, #5948):
//
//	"Flying, haste
//	 The Will of the Hive Mind — Other creatures you control with counters on
//	 them have flying and haste."
//
// One layer 6 grant over the other creatures its controller controls that have a
// counter of any kind on them as the layers are read (Fire Nation Salvagers'
// reading of "with a counter on it"), so a creature gains flying and haste the
// moment a counter lands and loses them when the last one goes.
//
// No simplification.
func init() {
	grant := effects.B16GrantKeywords(effects.GapB195OtherCreaturesWithCounters, "flying", "haste")
	grant.Label = "Other creatures you control with counters on them have flying and haste."
	effects.Register(effects.Spec{
		OracleID:        "6f0d8900-8460-4b31-8b67-eb4e470c3317",
		Name:            "Winged Hive Tyrant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Static:          []game.StaticAbility{grant},
	})
}
