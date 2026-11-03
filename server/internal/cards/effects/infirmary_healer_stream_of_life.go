package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infirmary Healer // Stream of Life — Creature — Cat Cleric {1}{G}, 2/3 //
// Sorcery {X}{G} (preparation card, CR 722; Secrets of Strixhaven, #4456):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Stream of Life — "Target player gains X life."
//
// X is announced when the copy is cast and read back at resolution; at X = 0
// the spell gains nothing, so it declares XMatters.
//
// No simplification.
func init() {
	const id = "4cbbdebc-b807-4d7b-92dd-312637d2a5de"
	Register(Spec{
		OracleID:     id,
		Name:         "Infirmary Healer",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Stream of Life",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return GainLife{Player: item.Targets[0].ID, Amount: ctx.X()}.Apply(ctx)
		},
	})
}
