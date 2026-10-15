package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bronze Guardian — Artifact Creature — Golem {4}{W}, */5 (#3656):
//
//	"Double strike
//	 Ward {2}
//	 Other artifacts you control have ward {2}.
//	 Bronze Guardian's power is equal to the number of artifacts you
//	 control."
//
// Mendicant Core's layer 7a power over the controller's artifacts
// (itself included). The printed ward is the shared ward trigger and the
// grant to the other artifacts is WardGranted, whose predicate is the
// same "other artifacts you control" shape the printed text uses.
//
// No simplification.
func init() {
	const label = "Bronze Guardian — ward {2}"
	Register(Spec{
		OracleID:        "922a030c-857b-4c61-93df-b70ebb32abd3",
		Name:            "Bronze Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static:          []game.StaticAbility{PowerEqualToArtifactsYouControl()},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), label),
			WardGranted(WardMana("{2}"), "Bronze Guardian — other artifacts you control have ward {2}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID != source.InstanceID && target.Controller == source.Controller && target.IsArtifact()
				}),
		},
	})
}
