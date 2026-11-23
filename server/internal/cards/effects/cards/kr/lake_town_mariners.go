package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lake-town Mariners // Gone Fishing — an Adventure card (CR 715,
// The Hobbit, #4454), oracle dba1d231…:
//
//	Lake-town Mariners — Creature — Human Citizen {4}{U}{U}, 6/5
//	  "Vigilance
//	   Ward {2}"
//	Gone Fishing — Instant — Adventure {3}{U}
//	  "Exile two target creatures and/or lands you control, then return
//	   them to the battlefield under their owner's control. (Then exile
//	   this card. You may cast the creature later from exile.)"
//
// Both targets are required (WithCount(2, 2)). The blink is Brago's
// (bragoBlinkTargets): the cards that reached exile come back together as
// one entry under their owners' control, each as a new object (CR 400.7),
// a token not at all (CR 111.8). A target that left in response is
// skipped (CR 608.2b).
//
// No simplification.

// lakeTownMarinersOracleID is shared by both faces of the card.
const LakeTownMarinersOracleID = "dba1d231-6b03-44ad-ab8c-be0c8fe1b6db"

func init() {
	effects.Register(effects.Spec{
		OracleID:        LakeTownMarinersOracleID,
		Name:            "Lake-town Mariners",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Lake-town Mariners — ward {2}"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     LakeTownMarinersOracleID + "#1",
		Name:         "Gone Fishing",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("two target creatures and/or lands you control",
			effects.Or(effects.Creature(), effects.Land()), effects.YouControl()).WithCount(2, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BragoBlinkTargets(ctx.Game, item)
		},
	})
}
