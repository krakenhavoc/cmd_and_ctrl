package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kinetic Augur — Creature — Human Shaman {3}{R}, */4 (Core Set 2021, #5511):
//
//	"Trample
//	 Kinetic Augur's power is equal to the number of instant and sorcery
//	 cards in your graveyard.
//	 When this creature enters, discard up to two cards, then draw that
//	 many cards."
//
// Enigma Drake's power-defining ability (CR 604.3, layer 7a, counted in
// every zone) and Sokka, Bold Boomeranger's enters trigger: the discard is
// chosen as it resolves, and the draw is as many as were discarded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "98e0b1a8-902d-4285-b13c-e63f2e1692f1",
		Name:            "Kinetic Augur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:     "Kinetic Augur's power is equal to the number of instant and sorcery cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = fandanielInstantsAndSorceries(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kinetic Augur — discard up to two cards, then draw that many",
				func(g *game.Game, item *game.StackItem) error {
					return b39MayDiscardThenDraw(2, false,
						"Kinetic Augur — discard up to two cards, then draw that many",
						drawAsManyAsDiscarded)(NewContext(g, item))
				}),
		},
	})
}
