package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jacques le Vert — Legendary Creature — Human Warrior {1}{R}{G}{W}, 3/2
// (Legends, #5639):
//
//	"Green creatures you control get +0/+2."
//
// Sylvan Anthem's layer 7c static at +0/+2: the controller's green
// creatures, colour read from the computed colours (TextColor, so a
// text change that rewrites the word is honoured). Jacques is green, so
// he is one of them and is a 3/4 on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c030ca14-33cb-40b3-a1f5-b6d0cb0efd49",
		Name:         "Jacques le Vert",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller && target.HasColor(source.TextColor("G"))
		}, 0, 2)},
	})
}
