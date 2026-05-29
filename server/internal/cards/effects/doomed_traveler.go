package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doomed Traveler — 1/1 Creature — Human Soldier for {W}:
//
//	"When Doomed Traveler dies, create a 1/1 white Spirit creature
//	token with flying."
//
// S19 sub-PR 4: a mandatory LTB trigger that creates one token via
// the CreateToken primitive. Flying on the Spirit token is cosmetic
// (token keywords aren't mechanically enforced yet). cardDied gates
// the trigger to graveyard-only.
func init() {
	Register(Spec{
		OracleID: "a30907c0-fbde-4fd3-a8c7-f304305fcea7",
		Name:     "Doomed Traveler",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				ctx := NewContext(g, nil)
				_ = CreateToken{
					Controller: source.Controller,
					Template:   SpiritToken(),
					N:          1,
				}.Apply(ctx)
				return nil
			},
		}},
	})
}
