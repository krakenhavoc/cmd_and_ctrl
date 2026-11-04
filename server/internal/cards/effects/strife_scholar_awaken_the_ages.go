package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strife Scholar // Awaken the Ages — Creature — Orc Sorcerer {2}{R}, 3/2
// // Sorcery {5}{R} (preparation card, CR 722; Secrets of Strixhaven, #4456):
//
//	"Ward—Pay 2 life.
//	 This creature enters prepared."
//
//	Awaken the Ages — "Create two 2/2 red and white Spirit creature tokens."
//
// Hexing Squelcher's ward (WardLife) and Carnivorous Cultivator's entry
// replacement (SelfEntersPrepared). Antiquities on the Loose makes the same
// two Spirits from the shared token row.
//
// No simplification.
func init() {
	const id = "e85c6bd2-db15-4052-88c0-3ae25414f0ca"
	Register(Spec{
		OracleID:     id,
		Name:         "Strife Scholar",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(2), "Strife Scholar — ward, pay 2 life"),
		},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Awaken the Ages",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("2/2 red and white Spirit"), N: 2}.Apply(ctx)
		},
	})
}
