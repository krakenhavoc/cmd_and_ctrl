package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spectral Sailor — Creature — Spirit Pirate {U}, 1/1 (EDHREC rank
// 1948):
//
//	"Flash
//	 Flying
//	 {3}{U}: Draw a card."
//
// The one-drop flash flier that turns spare mana into cards. Two
// printed keywords and a mana-only activated ability — no tap, so it
// can be activated any number of times and while summoning sick, as
// printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a8fdbcdf-479d-4582-9ad5-9fbd4c740c29",
		Name:            "Spectral Sailor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Activated: []ActivatedAbility{{
			Label: "{3}{U}: Draw a card.",
			Cost:  ManaCost("{3}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
