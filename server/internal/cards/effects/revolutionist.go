package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Revolutionist — Creature — Human Wizard {5}{R}, 3/3 (Modern Horizons 2,
// #5501):
//
//	"When this creature enters, return target instant or sorcery card
//	 from your graveyard to your hand.
//	 Madness {3}{R} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// Pinnacle Monk's enter trigger (removed for want of a target with no
// instant or sorcery card in your graveyard, CR 603.3d) and the engine's
// madness (CR 702.35a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d00607b1-77ce-4b89-ad04-763d0d390c5c",
		Name:         "Revolutionist",
		Completeness: CompletenessFull,
		Madness:      "{3}{R}",
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Revolutionist — return target instant or sorcery card from your graveyard to your hand",
				returnTargetInstantOrSorceryFromYourGraveyard),
				TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					YouOwn(), Or(Instant(), Sorcery())),
			),
		},
	})
}
