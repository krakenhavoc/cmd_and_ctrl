package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sentinel of the Nameless City — Creature — Merfolk Warrior Scout
// {2}{G}, 3/4 (The Lost Caverns of Ixalan, #4470):
//
//	"Vigilance
//	 Whenever this creature enters or attacks, create a Map token."
//
// Vigilance is the printed keyword (CR 702.20a). The trigger is one
// ability with two events (WhenThisEntersOrAttacks, Sun Titan's shape),
// and the Map is the catalog's token (CR 111.10s) with its own
// sorcery-speed explore ability, as Cartographer's Companion makes it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a5bbfe32-2fb9-48a0-bc8a-e9899aac1fda",
		Name:            "Sentinel of the Nameless City",
		Purpose:         game.Purpose{Tokens: 1},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Sentinel of the Nameless City — create a Map token",
				effects.Do(effects.CreateToken{Template: effects.MapToken(), N: 1})),
		},
	})
}
