package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Adelbert Steiner — Legendary Creature — Human Knight {1}{W}, 2/1
// (Final Fantasy, #4462):
//
//	"Lifelink
//	 Adelbert Steiner gets +1/+1 for each Equipment you control."
//
// A layer 7c self-modify whose count is re-read on every layer pass, so
// the bonus follows the board: every Equipment its controller controls
// counts, attached to it or not. "You" is Steiner's controller (CR
// 109.5). Equipment is an artifact type, which no text change reaches.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "657a4500-4b47-4c78-945b-a7791af7d10e",
		Name:            "Adelbert Steiner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Static: []game.StaticAbility{{
			Label:     "Adelbert Steiner gets +1/+1 for each Equipment you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := permanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtype("Equipment") })
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
