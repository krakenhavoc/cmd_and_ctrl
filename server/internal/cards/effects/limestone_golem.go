package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Limestone Golem — Artifact Creature — Golem {6}, 3/4 (Odyssey, #5613):
//
//	"{2}, Sacrifice this creature: Target player draws a card."
//
// The sacrifice is part of the cost (CR 602.2b). The target player, if
// still legal as the ability resolves (CR 608.2b), draws: any player,
// its controller included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0dddf5a9-fdf3-49e5-8103-a8cfd62f55c6",
		Name:         "Limestone Golem",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this creature: Target player draws a card.",
			Purpose: ForTargets(game.TargetPurpose{Slot: 0, Draws: 1}),
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Targets: TargetPlayer("target player"),
			Effect:  theTargetPlayerDraws(1),
		}},
	})
}
