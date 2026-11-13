package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Night // Day — split card (Apocalypse, #5614):
//
//	Night — Instant {B}: "Target creature gets -1/-1 until end of turn."
//	Day — Instant {2}{W}: "Creatures target player controls get +1/+1
//	 until end of turn."
//
// Each half is an ADR 0034 face (ADR 0103): Night under the bare oracle
// ID, Day under "#1". Only one half can be cast (CR 709.3; the card has
// no fuse). Night is the shared boostTargetUntilEOT with a negative
// boost. Day's set is the creatures the targeted player controls as it
// resolves (CR 611.2c, How to Start a Riot's shape), so a creature that
// player gets afterwards is not pumped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb32ab3e-ef76-4afb-959f-f8cc2d18416e",
		Name:         "Night",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -1, -1)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return boostTargetUntilEOT("Night — -1/-1", -1, -1)(ctx.Game, item)
		},
	})
	Register(Spec{
		OracleID:     "cb32ab3e-ef76-4afb-959f-f8cc2d18416e#1",
		Name:         "Day",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			p, ok := ctx.ClauseTarget(0)
			if !ok || p.Kind != game.TargetPlayer {
				return nil
			}
			return BoostUntilEOT{
				Match:     And(Creature(), controlledBy(p.ID)),
				Power:     1,
				Toughness: 1,
				Label:     "Day — +1/+1",
			}.Apply(ctx)
		},
	})
}
