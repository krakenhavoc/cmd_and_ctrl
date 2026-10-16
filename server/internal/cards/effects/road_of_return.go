package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Road of Return — Sorcery {G}{G}:
//
//	"Choose one —
//	 • Return target permanent card from your graveyard to your hand.
//	 • Put your commander into your hand from the command zone.
//	 Entwine {2} (Choose both if you pay the entwine cost.)"
//
// The first bullet is Restore with Empathy's return; the second is
// Command Beacon's move, read for the commanders you own that are in
// the command zone right now — with two (partners, a Background) you
// choose which (CR 608.2c), with none nothing happens. A commander
// going to a hand is a CR 903.9 destination like any other.
//
// Entwined (CR 702.42, #3866) the graveyard return comes first (CR
// 702.42b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "0007c283-5b7a-4c00-9ca1-b455c8dff8c3",
		Name:          "Road of Return",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{2}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Return target permanent card from your graveyard to your hand.",
				TargetCardInGraveyard("target permanent card from your graveyard", YouOwn(), Permanent()),
				ReturnTheModesGraveyardTargetToHand),
			ModeDoing("Put your commander into your hand from the command zone.", nil, roadOfReturnCommander),
		)),
	})
}

// roadOfReturnCommander is Command Beacon's move, narrowed to the
// commanders the caster owns.
func roadOfReturnCommander(item *game.StackItem, ctx *Context, _ int) error {
	return putCommanderIntoHand(ctx.Game, item, "Road of Return", true)
}
