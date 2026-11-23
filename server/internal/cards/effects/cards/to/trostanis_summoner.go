package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trostani's Summoner — Creature — Elf Shaman {5}{G}{W}, 1/1:
//
//	"When this creature enters, create a 2/2 white Knight creature token
//	 with vigilance, a 3/3 green Centaur creature token, and a 4/4 green
//	 Rhino creature token with trample."
//
// Three tokens in one resolution; each template is the printed token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "874e0b54-bc24-4887-abe1-1ecfd3a3abae",
		Name:         "Trostani's Summoner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Trostani's Summoner — a 2/2 Knight with vigilance, a 3/3 Centaur and a 4/4 Rhino with trample", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range []string{"2/2 white Knight with vigilance", "3/3 green Centaur", "4/4 green Rhino with trample"} {
					if err := (effects.CreateToken{Template: effects.TokenCard(t), N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
