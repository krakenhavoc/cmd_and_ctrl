package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primordial Sage — 4/5 Creature — Spirit for {4}{G}{G} (EDHREC rank
// 4426):
//
//	"Whenever you cast a creature spell, you may draw a card."
//
// Beast Whisperer before Beast Whisperer, at twice the price and with
// a body that survives a sweeper's worth of damage. Green's answer to
// running out of cards, and the reason it still shows up is that the
// trigger is on CAST — it fires even if the creature is countered or
// never resolves. Roadmap batch 42 (#449), "no new machinery".
//
// "You may" is a real prompt, not an auto-draw: OptionalPrompt puts
// the question to the controller when the trigger is harvested, and
// declining leaves the library alone. That matters in a deck that
// cares about its own library count, and it matters when the draw
// would kill you.
//
// Triggers on the cast, so the Sage sees a creature spell that is
// later countered, and does NOT see a creature that entered without
// being cast (a token, a reanimation, a Sneak Attack).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "499f1e1e-d96e-481f-a5d9-eb38da927cd7",
		Name:         "Primordial Sage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(
				WheneverYouCast(Creature(), "Primordial Sage — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					}),
				"Primordial Sage — draw a card?"),
		},
	})
}
