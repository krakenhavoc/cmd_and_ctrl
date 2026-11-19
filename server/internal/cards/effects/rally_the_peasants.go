package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rally the Peasants — Instant {2}{W} (Innistrad Remastered):
//
//	"Creatures you control get +2/+0 until end of turn.
//	 Flashback {2}{R}"
//
// The creatures are those you control as the spell resolves (CR 611.2c).
func init() {
	Register(Spec{
		OracleID:         "3f0a0a15-613a-4864-9309-a340dc4cf94d",
		Name:             "Rally the Peasants",
		Purpose:          game.Purpose{Answers: game.AnswerPump},
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{
				Match: And(Creature(), YouControl()),
				Power: 2,
				Label: "Rally the Peasants — +2/+0",
			}.Apply(ctx)
		},
	})
}
