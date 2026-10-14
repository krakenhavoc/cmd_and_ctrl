package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Polymorphist's Jest — Instant {1}{U}{U}:
//
//	"Until end of turn, each creature target player controls loses all
//	 abilities and becomes a blue Frog with base power and toughness
//	 1/1."
//
// Sudden Spoiling's shape with two more mods in the same record: one
// continuous effect, one timestamp, pinned to the creatures that player
// controls as it resolves (CR 611.2c), so a creature cast afterwards is
// untouched and one that blinks is a new object (CR 400.7):
//
//   - layer 4, "becomes a Frog": the creature types are replaced and
//     nothing else (game.SetCreatureTypesMod, #2144, CR 205.1a). An
//     artifact creature is still an artifact, a legendary one still
//     legendary;
//   - layer 5, "blue": the colours are set, so it is blue and no other
//     colour;
//   - layer 6, "loses all abilities" (CR 613.1f);
//   - layer 7b, "base power and toughness 1/1": a set, so an anthem
//     and counters still apply on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5bc66b2d-a93c-449d-a090-b61e1613aa15",
		Name:         "Polymorphist's Jest",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return untilEndOfTurn(ctx, uuid.Nil, And(Creature(), ControlledBy(item.Targets[0].ID)),
				"Polymorphist's Jest — a blue 1/1 Frog with no abilities",
				game.SetCreatureTypesMod("Frog"), game.SetColorsMod("U"), game.LoseAllAbilitiesMod(),
				game.SetBasePowerMod(1), game.SetBaseToughnessMod(1))
		},
	})
}
