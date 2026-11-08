package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shefet Archfiend — Creature — Demon {5}{B}{B}, 5/5 (Aetherdrift, #4464):
//
//	"Flying
//	 When this creature enters, all other creatures get -2/-2 until end
//	 of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Demon of Dark Schemes' enters trigger: the -2/-2 locks its set as the
// trigger resolves (CR 611.2c) and spares the Archfiend itself. Cycling
// works from the hand, so it never competes with the battlefield text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "06355d3f-a82d-4f12-a399-016a12824795",
		Name:            "Shefet Archfiend",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose: game.Purpose{Sweep: game.Sweep{
			Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Shefet Archfiend — all other creatures get -2/-2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return BoostUntilEOT{
						Match:     And(Creature(), OtherThan(ctx.Source())),
						Power:     -2,
						Toughness: -2,
						Label:     "Shefet Archfiend — -2/-2",
					}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
