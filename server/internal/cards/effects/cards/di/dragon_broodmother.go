package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon Broodmother — Creature — Dragon {2}{R}{R}{R}{G}, 4/4 (Alara Reborn,
// #5586):
//
//	"Flying
//	 At the beginning of each upkeep, create a 1/1 red and green Dragon
//	 creature token with flying and devour 2. (As the token enters, you may
//	 sacrifice any number of creatures. It enters with twice that many +1/+1
//	 counters on it.)"
//
// The trigger fires in every player's upkeep and the token is created under the
// Broodmother's controller. The token's devour 2 is a replacement on its own
// template (PrintedOldB503Dragon): its controller is asked as it enters which
// creatures to sacrifice, and the entering token is never a candidate
// (CR 614.13a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a9a6c2ba-30ba-4e45-9ea0-ea9833342069",
		Name:            "Dragon Broodmother",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Dragon Broodmother — create a 1/1 red and green Dragon creature token with flying and devour 2",
				effects.Do(effects.CreateToken{Template: effects.S82OldB503DragonToken(), N: 1})),
		},
	})
}
