package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shimmercreep — Creature — Elemental {4}{B}, 3/5 (Lorwyn Eclipsed,
// #4458):
//
//	"Menace
//	 Vivid — When this creature enters, each opponent loses X life and you
//	 gain X life, where X is the number of colors among permanents you
//	 control."
//
// Vivid is an ability word with no rules meaning (CR 207.2c). X is counted
// once, as the trigger resolves (the ruling), through Explosive Prodigy's
// colorsAmongPermanentsYouControl. Each opponent loses X, and you gain X
// once, not once per opponent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "52244c23-7680-4bed-8ef8-8d10e289d3f7",
		Name:            "Shimmercreep",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Shimmercreep — each opponent loses X life and you gain X life (vivid)",
				func(g *game.Game, item *game.StackItem) error {
					x := len(colorsAmongPermanentsYouControl(g, item.Controller))
					if x == 0 {
						return nil
					}
					ctx := NewContext(g, item)
					if err := (EachOpponentLosesLife{N: x}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Amount: x}.Apply(ctx)
				}),
		},
	})
}
