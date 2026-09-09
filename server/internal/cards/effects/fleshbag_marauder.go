package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fleshbag Marauder — Creature — Zombie Warrior {2}{B}, 3/1:
//
//	"When this creature enters, each player sacrifices a creature of
//	 their choice."
//
// Note "each player" — YOU sacrifice too, and the Marauder itself is
// on the battlefield when its own trigger resolves, so it is a legal
// choice for its own effect. That is exactly how the card is played:
// eat the Marauder, keep your real creatures, and every opponent still
// loses one. Feeding it to a sacrifice outlet first is the other line,
// and it is why ExceptController is false here.
//
// A one-sided edict for three mana that leaves a 3/1 behind if nobody
// makes it eat itself.
func init() {
	Register(Spec{
		OracleID: "4b1bf05e-753e-4350-a913-894cf3cecc0c",
		Name:     "Fleshbag Marauder",
		OnETB: func(card *game.Card, ctx *Context) error {
			return EachPlayerSacrifices{
				Match: Creature(),
				Label: "a creature",
			}.Apply(ctx)
		},
	})
}
