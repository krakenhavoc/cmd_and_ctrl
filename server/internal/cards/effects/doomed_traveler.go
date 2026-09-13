package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doomed Traveler — 1/1 Creature — Human Soldier for {W}:
//
//	"When Doomed Traveler dies, create a 1/1 white Spirit creature
//	token with flying."
//
// S19 sub-PR 4: a mandatory LTB trigger that creates one token via
// the CreateToken primitive when it resolves. cardDied gates the
// trigger to graveyard-only.
//
// Flying on the Spirit token is real: SpiritToken carries it on
// Card.Keywords, which printedCharacteristic folds into the layer
// engine (S21 sub-PR 1). The "cosmetic" note here outlived the fix;
// corrected in the #338 sweep.
func init() {
	Register(Spec{
		OracleID:     "a30907c0-fbde-4fd3-a8c7-f304305fcea7",
		Name:         "Doomed Traveler",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Spirit token is created colorless instead of white, so anything that cares about a creature's color doesn't see it."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Doomed Traveler — create a 1/1 Spirit",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   SpiritToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
