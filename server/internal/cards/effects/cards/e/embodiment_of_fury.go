package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embodiment of Fury — Creature — Elemental {3}{R}, 4/3 (Oath of the Gatewatch,
// #5545):
//
//	"Trample
//	 Land creatures you control have trample.
//	 Landfall — Whenever a land you control enters, you may have target land you
//	 control become a 3/3 Elemental creature with haste until end of turn. It's
//	 still a land."
//
// The grant is a layer 6 keyword over the land creatures you control, read
// live. The landfall target is chosen as the trigger goes on the stack and the
// "may" asked as it resolves (Roil Elemental's shape); the animation is
// Mobilized District's: the Creature type and Elemental subtype, base 3/3 and
// haste until end of turn, the land keeping its land type and abilities. The
// entering land may be its own target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0e8adf34-0834-4027-9cda-18df69239e2b",
		Name:            "Embodiment of Fury",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{effects.KeywordGrant(effects.B31LandCreaturesYouControl, "trample")},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.Landfall("Embodiment of Fury — you may have target land you control become a 3/3 Elemental creature with haste until end of turn", embodimentOfFuryAnimate),
				effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl())),
				"Embodiment of Fury — have target land you control become a 3/3 Elemental creature with haste until end of turn?"),
		},
	})
}

func embodimentOfFuryAnimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.ScopedEffectFor{
		Target: id,
		Mods: []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Elemental")...),
			game.SetBasePowerMod(3),
			game.SetBaseToughnessMod(3),
			game.AddKeywordsMod("haste"),
		},
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Embodiment of Fury — a 3/3 Elemental creature with haste until end of turn",
	}.Apply(ctx)
}
