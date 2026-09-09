package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thraben Inspector — 1/2 Creature — Human Soldier for {W}:
//
//	"When Thraben Inspector enters the battlefield, investigate."
//
// Investigating is "create a Clue token", and the Clue is where the
// value is: a card in artifact form that costs {2} to cash in. The
// whole loop — make the token, crack it later for the draw — only
// became possible in S21, because a Clue is nothing but its
// activated ability.
func init() {
	Register(Spec{
		OracleID: "caa02547-66e3-4e27-a2d3-5e94f3e7a069",
		Name:     "Thraben Inspector",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Thraben Inspector — investigate",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   ClueToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
