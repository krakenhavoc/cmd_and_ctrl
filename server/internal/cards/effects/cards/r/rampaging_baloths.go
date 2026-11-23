package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rampaging Baloths — Creature — Beast {4}{G}{G}, 6/6 (EDHREC rank
// 379):
//
//	"Trample
//	 Landfall — Whenever a land you control enters, create a 4/4
//	 green Beast creature token."
//
// The landfall finisher: a 4/4 per land drop, two per fetchland.
// Tireless Provisioner's trigger with a bigger payoff — the same
// ETB-filtered-to-lands condition, firing for played, fetched and
// bounced-and-replayed lands alike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2d3e6549-6cc6-434f-a189-ba3b55e64c34",
		Name:            "Rampaging Baloths",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Rampaging Baloths — create a 4/4 Beast (landfall)", game.LandfallPayoff{Tokens: 1}, effects.Do(effects.CreateToken{
				Template: effects.TokenCard("4/4 green Beast"),
				N:        1,
			})),
		},
	})
}
