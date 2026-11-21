package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elite Interceptor // Rejoinder — Creature — Human Wizard {W}, 1/2 //
// Sorcery {1}{W} (Secrets of Strixhaven, #4456; preparation card, CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Rejoinder — "You may tap or untap target creature.
//	 Draw a card."
//
// "Enters prepared" is SelfEntersPrepared. Rejoinder targets a creature
// as it is cast; as it resolves the caster picks tap it, untap it or leave
// it (the "you may", Hidden Strings' three-way question), and the draw
// comes after that question is answered. The draw is the spell's own
// step: it happens whenever the spell resolves at all, so a target that
// became illegal costs only the tap or untap (CR 608.2b).
//
// No simplification.
func init() {
	const oracle = "d97de156-5a6d-4dc6-953e-d4bad63cbad2"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Elite Interceptor",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersPrepared()},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Rejoinder",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.SocB23MayTapOrUntap(ctx, "Rejoinder", t.ID)
				}
				return nil
			}),
			effects.DrawCards{N: 1},
		),
	})
}
