package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tendershoot Dryad — Creature — Dryad {4}{G}, 2/2 (EDHREC rank
// 1301):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 At the beginning of each upkeep, create a 1/1 green Saproling
//	 creature token.
//	 Saprolings you control get +2/+2 as long as you have the city's
//	 blessing."
//
// A Saproling every upkeep — every player's, as printed — and a lord
// once the board is wide. The token is an ordinary upkeep trigger
// with no actor test; the anthem is a Layer 7c static over Saprolings
// the controller controls (effective subtypes, so a changeling
// counts).
//
// SANDBOX GAP, weaker than printed: the city's blessing is a player
// designation that, once earned, lasts the rest of the game, and the
// engine has no per-player designation to keep it in. The anthem
// therefore reads the condition live — "as long as you control ten
// or more permanents" — so a board that shrinks below ten loses the
// bonus where the printed card would keep it. Never stronger: the
// condition that earns the blessing is the same one.
func init() {
	Register(Spec{
		OracleID:     "a336c10a-b5bd-47ff-ba2d-31e27af1e15a",
		Name:         "Tendershoot Dryad",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The city's blessing isn't kept once earned — Saprolings get +2/+2 only while you control ten or more permanents."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(_ game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return true
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Tendershoot Dryad — create a Saproling",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{Controller: item.Controller, Template: b11GreenSaprolingToken(), N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller &&
					target.IsCreature() && target.HasSubtype("Saproling") &&
					b11PermanentsControlled(g, source.Controller) >= 10
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 2
			},
		}},
	})
}
