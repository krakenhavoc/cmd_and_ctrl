package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "1d705d77-4b1a-4a17-b3b7-e502f2f6f6e6",
		Name:            "Koma's Faithful",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Koma's Faithful — each player mills three cards", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, p := range apnapPlayers(g) {
					if err := (MillCards{Player: p, N: 3}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
