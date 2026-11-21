package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twin-Silk Spider — Creature — Spider {2}{G}, 1/2 (Modern Horizons, #5519):
//
//	"Reach
//	 When this creature enters, create a 1/2 green Spider creature token
//	 with reach."
//
// Brood Weaver's token on an enters trigger instead of a dies one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a2193302-402a-4cf0-8014-5facecad647b",
		Name:            "Twin-Silk Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Twin-Silk Spider — create a 1/2 green Spider creature token with reach",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/2 green Spider with reach"), N: 1})),
		},
	})
}
