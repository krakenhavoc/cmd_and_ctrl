package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hallway Heckler // Vicious Verse — Creature — Elemental Sorcerer {2}{R},
// 2/3 // Sorcery {B/R} (preparation card, CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)
//	 {T}, Discard a card: Draw a card."
//
//	Vicious Verse — "Vicious Verse deals 1 damage to target opponent."
//
// No simplification.
func init() {
	const id = "d1af8018-35b5-4b6f-93c8-1ed6bd8c92b1"
	Register(Spec{
		OracleID:     id,
		Name:         "Hallway Heckler",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
		Activated: []ActivatedAbility{{
			Label: "{T}, Discard a card: Draw a card.",
			Cost:  Plus(TapCost(), DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Vicious Verse",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    viciousVerseResolve,
	})
}
