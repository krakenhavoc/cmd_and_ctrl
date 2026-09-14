package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sensei's Divining Top — Artifact {1}:
//
//	"{1}: Look at the top three cards of your library, then put them
//	 back in any order."
//	"{T}: Draw a card, then put this artifact on top of its owner's
//	 library."
//
// The card that makes "put them back in any order" worth
// implementing. The two abilities are a loop: look at three, arrange
// them, tap to draw the one you put on top — and the Top itself goes
// back on the library, so next turn's draw is the Top again and the
// whole thing repeats. Neither half is interesting alone; together
// they are why this is banned in Legacy.
//
// The tap ability is deliberately ordered the way it prints. "Draw a
// card, THEN put this on top" means the draw comes off the library as
// it stands — the card the player just arranged — and only afterwards
// does the Top land above it. Tucking first would draw the Top back
// into hand, which is a different and much worse card.
//
// Sandbox simplification: NONE for the abilities themselves. Note
// that tucking the Top is a zone change to a hidden zone, so its
// knowers are cleared — the controller does not get to keep seeing it
// sitting on top of their library, which matches paper: you know
// it is there, you just cannot read the cards under it.
//
// Shipped unreviewed and audited in #74's theme-deck pass, which runs
// the whole loop — look at three, arrange, tap, draw the arranged
// card, tuck — and then has Dark Confidant flip the tucked Top
// straight back for 1 life (blue_draw_deck_smoke_test.go).
func init() {
	Register(Spec{
		OracleID:     "13575cf9-65c1-4861-b21e-eb2155e07766",
		Name:         "Sensei's Divining Top",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "{1}: Look at the top three cards of your library, then put them back in any order.",
				Cost:  game.AbilityCost{Mana: "{1}"},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return LookAtTop{Player: item.Controller, N: 3}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "{T}: Draw a card, then put this artifact on top of its owner's library.",
				Cost:  game.AbilityCost{Tap: true},
				Effect: func(g *game.Game, item *game.StackItem) error {
					// Order is load-bearing: draw off the library as
					// the player left it, THEN tuck. Tucking first
					// would draw the Top itself.
					if err := g.DrawNForEffect(item.Controller, 1); err != nil {
						return err
					}
					return g.TuckToLibraryForEffect(item.SourceCardID, false)
				},
			},
		},
	})
}
