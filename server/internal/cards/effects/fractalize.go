package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fractalize — Instant {X}{U} (Secrets of Strixhaven, #4456):
//
//	"Until end of turn, target creature becomes a green and blue Fractal with
//	 base power and toughness each equal to X plus 1. (It loses all other
//	 colors and creature types.)"
//
// One continuous effect until end of turn (CR 611.2a), one timestamp, on the
// creature chosen when the spell was cast, with three layers in the same
// record (Polymorphist's Jest's shape, Secret Identity's single target):
//
//   - layer 4, "becomes a Fractal": the creature types are replaced and
//     nothing else (game.SetCreatureTypesMod, CR 205.1a), so an artifact
//     creature is still an artifact and a legendary one still legendary;
//   - layer 5, "green and blue": the colours are set, so it is exactly those
//     two and no other colour (CR 613.1e);
//   - layer 7b, "base power and toughness each equal to X plus 1": a set, so
//     counters and pumps still apply on top. X is the value announced as the
//     spell was cast (CR 107.3f) and is fixed for the effect's life.
//
// Its abilities are kept: the card does not say "loses all abilities".
//
// X is read, but X=0 is still a real move (the creature becomes a 1/1 Fractal
// until end of turn), so the card leaves XMatters unset and takes an entry in
// xMattersAllowlist.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e7bd032-8946-4d72-9910-b224823bc940",
		Name:         "Fractalize",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			size := ctx.X() + 1
			return untilEndOfTurn(ctx, t.ID, nil,
				"Fractalize — becomes a green and blue Fractal with base power and toughness equal to X plus 1 until end of turn",
				game.SetCreatureTypesMod(ctx.TextSubtypes("Fractal")...),
				game.SetColorsMod(ctx.TextColors("G", "U")...),
				game.SetBasePowerMod(size), game.SetBaseToughnessMod(size))
		},
	})
}
