package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lingering Souls — Sorcery for {2}{W}:
//
//	"Create two 1/1 white Spirit creature tokens with flying.
//	 Flashback {1}{B}"
//
// The card that explains why flashback is a cast PATH rather than a
// second mana cost: the {2}{W} and the {1}{B} are different colours,
// so a deck plays this off white mana and buys the second half with
// black. Nothing in the engine has to know that — the offer carries
// its own cost string and the zone it is claimable from, and the
// mana gate charges whichever one was claimed.
//
// The Spirit token is the same template Doomed Traveler uses, and
// its flying is real (printedCharacteristic folds Card.Keywords into
// the layer engine, S21 sub-PR 1).
func init() {
	Register(Spec{
		OracleID:         "0b8c3337-04dd-4798-8203-6d8b8cfb936b",
		Name:             "Lingering Souls",
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return CreateToken{
				Controller: item.Controller,
				Template:   SpiritToken(),
				N:          2,
			}.Apply(ctx)
		},
	})
}
