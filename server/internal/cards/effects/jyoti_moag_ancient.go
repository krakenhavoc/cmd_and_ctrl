package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jyoti, Moag Ancient — Legendary Creature — Elemental {2}{G}{U}, 2/4
// (Modern Horizons 3 Commander, #5468):
//
//	"When Jyoti enters, create a 1/1 green Forest Dryad land creature
//	 token for each time you've cast your commander from the command zone
//	 this game. (They're affected by summoning sickness.)
//	 At the beginning of each combat, land creatures you control get
//	 +X/+X until end of turn, where X is Jyoti's power."
//
// The count is Commander's Insignia's: every cast of a commander you own
// from the command zone this game, both partners' summed, read as the
// enters trigger resolves. The tokens are Awaken the Woods' Forest
// Dryads, which enter summoning sick like any creature.
//
// The combat trigger fires at the beginning of every combat, on any
// player's turn. X is Jyoti's power as the trigger resolves, or as it
// last existed if it has left (CR 608.2h), and never less than zero. The
// creatures it pumps are the land creatures you control as it resolves
// (CR 611.2c), so one that becomes a land creature later that turn is not
// pumped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67f0cef6-8ef4-4c39-b306-2b580036894e",
		Name:         "Jyoti, Moag Ancient",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Jyoti, Moag Ancient — create a Forest Dryad for each time you've cast your commander",
				func(g *game.Game, item *game.StackItem) error {
					n := socB3CommanderCastsFromCommandZone(g, item.Controller)
					if n <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 green Forest Dryad"), N: n}.Apply(NewContext(g, item))
				}),
			On(game.EventStepBegan, StepBegan(game.StepBeginCombat, false),
				"Jyoti, Moag Ancient — land creatures you control get +X/+X",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x := sourcePowerNow(ctx)
					return BoostUntilEOT{
						Match:     And(Land(), Creature(), YouControl()),
						Power:     x,
						Toughness: x,
						Label:     "Jyoti, Moag Ancient — +X/+X",
					}.Apply(ctx)
				}),
		},
	})
}
