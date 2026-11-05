package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// There's Coffee in that Nebula — Sorcery {1}{W} (Star Trek, #4973):
//
//	"Look at the top five cards of your library. You may reveal a Plains
//	 card or nonland permanent card with mana value 3 or less from among
//	 them and put it into your hand. Put the rest on the bottom in a
//	 random order.
//	 Beam me up {2}{W} (You may cast this card from your graveyard for
//	 {2}{W} if you also return a creature you control to its owner's
//	 hand. Then exile this spell.)"
//
// Horn of the Mark's look-and-take over five cards; "with mana value 3
// or less" qualifies the nonland permanent card only. Beam me up is
// trekB4BeamMeUp.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "887614f9-f12f-47bc-ae56-1438cd7cad10",
		Name:             "There's Coffee in that Nebula",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{trekB4BeamMeUp("{2}{W}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return LookAtTopThenMayTakeToHand(5,
				Or(HasSubtype("Plains"), And(Nonland(), Permanent(), ManaValueLE(3))), 1,
				"There's Coffee in that Nebula — you may reveal a Plains card or nonland permanent card with mana value 3 or less and put it into your hand")(ctx.Game, item)
		},
	})
}
