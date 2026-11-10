package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Metamorphic Blast — Instant {U} (Outlaws of Thunder Junction, #4468):
//
//	"Spree (Choose one or more additional costs.)
//	 + {1} — Until end of turn, target creature becomes a white Rabbit
//	   with base power and toughness 0/1.
//	 + {3} — Target player draws two cards."
//
// Three Steps Ahead's Spree (CR 702.172a), each bullet with its own
// target. The first bullet is Eccentric Apprentice's "becomes a <type>
// with base power and toughness" with a colour added: one effect that
// sets the creature's colour to white (layer 5), replaces its creature
// types with Rabbit (layer 4) and sets its base power and toughness to
// 0/1 (layer 7b), so counters and pumps still apply on top. It keeps its
// abilities and its other card types. The second bullet's player may be
// you or anyone else.
//
// No simplification.
func init() {
	const rabbitLabel = "Metamorphic Blast — a white Rabbit with base power and toughness 0/1 until end of turn"
	Register(Spec{
		OracleID:     "69c8e690-c0a2-4c4e-910d-123c9c5b922f",
		Name:         "Metamorphic Blast",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Until end of turn, target creature becomes a white Rabbit with base power and toughness 0/1.", "{1}",
				TargetCreature("target creature"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return untilEndOfTurn(ctx, t.ID, nil, rabbitLabel,
						game.SetColorsMod(ctx.TextColors("W")...),
						game.SetCreatureTypesMod(ctx.TextSubtypes("Rabbit")...),
						game.SetBasePowerMod(0), game.SetBaseToughnessMod(1))
				}),
			SpreeModeDoing("Target player draws two cards.", "{3}",
				TargetPlayer("target player"),
				TheModesTargetPlayerDrawsTwo),
		),
	})
}
