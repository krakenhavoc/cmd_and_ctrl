package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sylvan Anthem — Enchantment {G}{G} (EDHREC rank 2808):
//
//	"Green creatures you control get +1/+1.
//	 Whenever a green creature you control enters, scry 1."
//
// The mono-green Glorious Anthem with a cantrip-ish rider. The anthem
// is a layer 7c static over the controller's green creatures — colour
// read from the card's computed colours, with the mana-cost fallback
// for fixtures — and the scry is an ordinary ETB watcher (Ivy Lane
// Denizen's condition); the Anthem is not a creature, so no "another"
// guard is needed. Scry queues the prompt and nothing moves until the
// controller answers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5ab5eefd-47bf-4cdf-bc29-b517e4f6adc0",
		Name:         "Sylvan Anthem",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller && target.HasColor("G")
		}, 1, 1)},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b23AnotherGreenCreatureYouControlEntered(ev, source, g)
			}, "Sylvan Anthem — scry 1", Do(Scry{N: 1})),
		},
	})
}
