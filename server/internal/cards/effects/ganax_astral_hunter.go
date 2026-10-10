package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ganax, Astral Hunter — Legendary Creature — Dragon {4}{R}, 3/4
// (EDHREC rank 1548):
//
//	"Flying
//	 Whenever Ganax or another Dragon you control enters, create a
//	 Treasure token. (It's an artifact with "{T}, Sacrifice this
//	 token: Add one mana of any color.")
//	 Choose a Background (You can have a Background as a second
//	 commander.)"
//
// A Treasure per Dragon, Ganax included. One printed ability with two
// conditions is one TriggeredAbility: the source's own entry, or
// another Dragon entering under the same control. Effective
// subtypes, so a changeling counts.
//
// Choose a Background is a deck-construction rule (CR 702.124k) that
// internal/deck reads off the oracle text: Ganax may share the command
// zone with a Background (#2874).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3112971-3af4-46e8-b2d6-a8759b39d0d1",
		Name:            "Ganax, Astral Hunter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b14SelfOrDragonYouControlEntered(ev, source, g)
			}, "Ganax, Astral Hunter — create a Treasure", Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
