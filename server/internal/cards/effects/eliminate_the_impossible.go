package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eliminate the Impossible — Instant {1}{U}:
//
//	"Investigate. Creatures your opponents control get -2/-0 until end
//	 of turn. If any of them are suspected, they're no longer suspected.
//	 (To investigate, create a Clue token. It's an artifact with "{2},
//	 Sacrifice this token: Draw a card.")"
//
// The -2/-0 reaches the creatures your opponents control as the spell
// resolves and no others (CR 611.2c), and "any of them" is the same
// set: UnsuspectAll with the same match, so a suspected creature you
// control is left alone and one an opponent gains control of later this
// turn is not retroactively included.
//
// No simplification.
func init() {
	opponentsCreatures := func() CardPredicate { return And(Creature(), OpponentControls()) }
	Register(Spec{
		OracleID:     "c78d4cec-6764-4cce-96d2-a2d85a57b218",
		Name:         "Eliminate the Impossible",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (CreateToken{Template: ClueToken(), N: 1}).Apply(ctx); err != nil {
				return err
			}
			if err := (BoostUntilEOT{
				Match: opponentsCreatures(), Power: -2,
				Label: "Eliminate the Impossible — -2/-0 until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return UnsuspectAll{Match: opponentsCreatures()}.Apply(ctx)
		},
	})
}
