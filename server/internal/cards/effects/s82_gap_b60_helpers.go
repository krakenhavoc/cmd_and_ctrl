package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b60_helpers.go — shared bodies for the Fallout (PIP) artifacts of
// GAP batch 60 (#5948).

// s82GapB60BobbleheadCount is "X, where X is the number of Bobbleheads you
// control as you activate this ability" (Agility and Endurance Bobblehead):
// the bound on "up to X target creatures", read once at announce (CR 601.2b,
// 602.2b; TargetSpec.MaxFrom). The activating Bobblehead is on the
// battlefield, so it counts itself.
var s82GapB60BobbleheadCount = game.RegisterDivideAmount("s82-gap-b60/bobbleheads-you-control",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		return game.DivideSpec{Total: pipBobbleheadsYouControl(g, a.Controller)}
	})

// s82GapB60UpToXCreatures is the clause "up to X target creatures you
// control" with X the Bobbleheads count above.
func s82GapB60UpToXCreatures() *game.TargetSpec {
	t := TargetCreature("up to X target creatures you control, where X is the number of Bobbleheads you control as you activate this ability", YouControl()).WithCount(0, 0)
	t.MaxFrom = s82GapB60BobbleheadCount
	return t
}

// s82GapB60FortifyAbility is the CR 702.67a fortify ability: "{cost}: Attach
// this Fortification to target land you control. Activate only as a
// sorcery." Equip's twin for a land host; the attach is the same
// AttachSourceToTarget.
func s82GapB60FortifyAbility(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label:        "Fortify " + cost,
		Cost:         ManaCost(cost),
		Targets:      TargetPermanent("target land you control", Land(), YouControl()),
		SorcerySpeed: true,
		Effect:       AttachSourceToTarget,
	}
}

// s82GapB60ManaColorsOfLastTap lists the colours of the mana the land
// `landID` added in its most recent mana ability, read off the event log: the
// EventManaAdded events between the land's latest EventManaAbilityActivated
// and the EventManaAbilityResolved with sequence `resolvedSeq`.
func s82GapB60ManaColorsOfLastTap(g *game.Game, landID uuid.UUID, resolvedSeq uint64) []string {
	var colors []string
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Seq >= resolvedSeq {
			continue
		}
		if ev.Source != landID {
			continue
		}
		if ev.Kind == game.EventManaAbilityActivated {
			break
		}
		if ev.Kind == game.EventManaAdded {
			colors = append(colors, ev.Colors...)
		}
	}
	return colors
}
