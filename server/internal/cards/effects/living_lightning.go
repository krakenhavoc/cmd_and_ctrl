package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Living Lightning — Creature — Elemental Shaman {3}{R}, 3/2 (Ravnica:
// Clue Edition, Bloomburrow):
//
//	"When this creature dies, return target instant or sorcery card from
//	 your graveyard to your hand."
//
// Pinnacle Monk's return on a dies trigger. The target is chosen as the
// trigger goes on the stack and is re-checked as it resolves (CR 608.2b);
// with no instant or sorcery card in your graveyard it is removed for want
// of a target (CR 603.3d). The Lightning itself is a creature card and is
// never a legal choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f4738db-be2e-4d02-ad0c-1b74dd6bbfdc",
		Name:         "Living Lightning",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Living Lightning — return target instant or sorcery card from your graveyard to your hand",
				returnTargetInstantOrSorceryFromYourGraveyard),
				TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					YouOwn(), Or(Instant(), Sorcery())),
			),
		},
	})
}
