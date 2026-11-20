package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Koma's Faithful — Creature — Elf Cleric {2}{B}, 3/1:
//
//	"Lifelink
//	 When this creature dies, each player mills three cards."
//
// A dies trigger (Abyssal Gatekeeper's) that mills every player in APNAP
// order (Disturbing Conversion's loop).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1d705d77-4b1a-4a17-b3b7-e502f2f6f6e6",
		Name:            "Koma's Faithful",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Koma's Faithful — each player mills three cards", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, p := range effects.ApnapPlayers(g) {
					if err := (effects.MillCards{Player: p, N: 3}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
