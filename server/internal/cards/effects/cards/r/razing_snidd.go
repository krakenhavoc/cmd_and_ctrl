package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Razing Snidd — Creature — Beast {4}{B}{R}, 3/3 (Planeshift, S82 OLD b643, #5615):
//
//	"When this creature enters, return a black or red creature you control to its
//	 owner's hand.
//	 When this creature enters, each player sacrifices a land of their choice."
//
// Marsh Crocodile's mandatory return, then Akki Blizzard-Herder's edict over
// lands: each player chooses their own land and the sacrifices are one simultaneous
// event (CR 701.21). The Snidd is a legal pick for the return while it is on the
// battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bbcd2045-cf51-4346-a561-ebae8ffdc181",
		Name:         "Razing Snidd",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Razing Snidd — return a black or red creature you control to its owner's hand", effects.Do(effects.ReturnOneYouControl{
				Match:    effects.OldB643CreatureOfEitherColor("B", "R"),
				Question: "Razing Snidd — return a black or red creature you control to its owner's hand",
			})),
			effects.WhenThisEnters("Razing Snidd — each player sacrifices a land of their choice", effects.Do(effects.EachPlayerSacrifices{
				Match: effects.Land(),
				Label: "a land",
			})),
		},
	})
}
