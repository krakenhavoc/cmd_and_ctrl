package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merciless Executioner — Creature — Orc Warrior {2}{B}, 3/1 (EDHREC
// rank 2205):
//
//	"When this creature enters, each player sacrifices a creature of
//	 their choice."
//
// Fleshbag Marauder with a different type line, and the same body:
// EachPlayerSacrifices with the controller included — "each player"
// means you too, and the Executioner is on the battlefield when its
// own trigger resolves, so feeding it to itself is the printed line.
// Every player picks their own creature in their own prompt; a
// player with none is skipped (CR 701.17b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3c45d50-9038-41df-bb2f-9bc40071845b",
		Name:         "Merciless Executioner",
		Completeness: CompletenessFull,
		OnETB: func(_ *game.Card, ctx *Context) error {
			return EachPlayerSacrifices{
				Match: Creature(),
				Label: "a creature",
			}.Apply(ctx)
		},
	})
}
