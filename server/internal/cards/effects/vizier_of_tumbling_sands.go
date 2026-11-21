package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vizier of Tumbling Sands — Creature — Human Cleric {2}{U}, 1/3 (Commander
// Legends: Battle for Baldur's Gate):
//
//	"{T}: Untap another target permanent.
//	 Cycling {1}{U}
//	 When you cycle this card, untap target permanent."
//
// D'Vana Tendi's tap-to-untap-another ability, with cycling and a Choking
// Tethers cycled trigger that works from the graveyard (CR 702.29d) and
// untaps any target permanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "699fc7cf-1ad3-4c32-b3e4-49f9a2da4b26",
		Name:         "Vizier of Tumbling Sands",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Untap another target permanent.",
				Cost:    TapCost(),
				Targets: Another(TargetPermanent("another target permanent")),
				Effect:  b22UntapEachLegalTarget,
			},
			Cycling("{1}{U}"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				InGraveyard(On(game.EventCycle, Self, "Vizier of Tumbling Sands — cycled: untap target permanent",
					b22UntapEachLegalTarget)),
				TargetPermanent("target permanent"),
			),
		},
	})
}
