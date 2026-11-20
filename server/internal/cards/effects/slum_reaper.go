package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slum Reaper — Creature — Horror {3}{B}, 4/2 (Return to Ravnica, #5566):
//
//	"When this creature enters, each player sacrifices a creature of
//	 their choice."
//
// Fleshbag Marauder's edict: "each player" includes you, and the Reaper is
// on the battlefield as its trigger resolves, so it is a legal choice for
// its own effect.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0cf742c1-feb5-4108-9976-37fa58528212",
		Name:         "Slum Reaper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WhenThisEnters("Slum Reaper — each player sacrifices a creature", Do(EachPlayerSacrifices{
				Match: Creature(),
				Label: "a creature",
			}))),
		},
	})
}
