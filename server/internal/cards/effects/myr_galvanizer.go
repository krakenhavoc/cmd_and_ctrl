package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Myr Galvanizer — Artifact Creature — Myr {3}, 2/2 (Scars of Mirrodin,
// #5577):
//
//	"Other Myr creatures you control get +1/+1.
//	 {1}, {T}: Untap each other Myr you control."
//
// Goblin Chieftain's "other … you control" lord half (TribalAnthem), and
// Deepway Navigator's untap of each other permanent of the type, read as
// the ability resolves. "Myr" is a word of the card's text (CR 612.2).
// The untap reads every Myr permanent, not only creatures, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e070028e-468a-41e9-9ddd-a8cd01551266",
		Name:         "Myr Galvanizer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Myr"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Untap each other Myr you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect: somB1UntapEachOther(func(ctx *Context, c game.Card) bool {
				return c.HasSubtype(ctx.TextSubtype("Myr"))
			}),
		}},
	})
}
