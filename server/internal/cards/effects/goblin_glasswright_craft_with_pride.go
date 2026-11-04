package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Glasswright // Craft with Pride — a preparation card (CR 722, ADR
// 0090), Secrets of Strixhaven (#4456):
//
//	Goblin Glasswright {1}{R}, Creature — Goblin Sorcerer 2/2:
//	"This creature enters prepared. (While it's prepared, you may cast a copy
//	 of its spell. Doing so unprepares it.)"
//	Craft with Pride {R}, Sorcery:
//	"Create a Treasure token."
//
// "Enters prepared" is SelfEntersPrepared; the engine makes, casts and
// unprepares the copy. Craft with Pride is one Treasure for its controller.
//
// No simplification.
func init() {
	const id = "23c3de07-8aa2-48e9-848c-50b5a40bd133"
	Register(Spec{
		OracleID:     id,
		Name:         "Goblin Glasswright",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Craft with Pride",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(CreateToken{Template: TreasureToken(), N: 1}),
	})
}
