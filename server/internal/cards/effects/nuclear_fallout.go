package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nuclear Fallout — Sorcery {X}{B}{B}:
//
//	"Each creature gets twice -X/-X until end of turn. Each player gets
//	 X rad counters."
//
// #2042. "Twice -X/-X" is -2X/-2X, Toxic Deluge's layer-7c shrink rather
// than a destruction (the zero-toughness state-based action does the
// killing), snapshotted at resolution (CR 611.2c). Every player still in
// the game, the caster included, then gets X rad counters.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f09388e2-f190-4b48-ba85-3ffc33020392",
		Name:         "Nuclear Fallout",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true}},
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			if err := (BoostUntilEOT{
				Match:     Creature(),
				Power:     -2 * x,
				Toughness: -2 * x,
				Label:     "Nuclear Fallout — twice -X/-X",
			}).Apply(ctx); err != nil {
				return err
			}
			return eachPlayerGetsRadCounters(ctx.Game, ctx.Controller(), x)
		},
	})
}
