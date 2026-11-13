package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Paradox Engine — Legendary Artifact {5} (Aether Revolt; Kaladesh
// Inventions, #5538):
//
//	"Whenever you cast a spell, untap all nonland permanents you
//	 control."
//
// A cast trigger: it goes on the stack above the spell that triggered it
// and resolves first, while that spell is still waiting. It untaps every
// tapped nonland permanent its controller controls as it resolves (the
// Engine itself included); lands stay tapped, and other players'
// permanents are left alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d38937a6-ab1c-4de8-a278-d81a6ce0d921",
		Name:         "Paradox Engine",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(nil, "Paradox Engine — untap all nonland permanents you control", func(g *game.Game, item *game.StackItem) error {
				return b16UntapAllYouControlMatching(NewContext(g, item), item.Controller, func(c game.Card) bool { return !c.IsLand() })
			}),
		},
	})
}
