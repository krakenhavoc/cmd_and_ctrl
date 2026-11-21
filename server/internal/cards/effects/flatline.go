package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flatline — Instant {2}{U} (Doctor Who):
//
//	"Creatures your opponents control have base power and toughness 0/1
//	 until end of turn."
//
// A layer 7b set (CR 613.4b) until end of turn over the creatures your
// opponents control as the spell resolves: a continuous effect that changes
// characteristics locks its set when it begins (CR 611.2c), so a creature that
// enters later is unaffected. Counters and pumps still apply on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9c4a3ca-b966-4fd0-b189-ffbcbfb412b1",
		Name:         "Flatline",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ScopedEffectFor{
				Match:    And(Creature(), OpponentControls()),
				Mods:     game.SetBasePTMods(0, 1),
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Flatline — creatures your opponents control have base power and toughness 0/1 until end of turn",
			}.Apply(ctx)
		},
	})
}
