package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abzan Advantage — {1}{W} Instant (Fate Reforged):
//
//	"Target player sacrifices an enchantment of their choice. Bolster 1."
//
// The targeted player chooses what to sacrifice (CR 701.17); the bolster is the
// caster's and follows the sacrifice from its continuation, whether or not the
// player had an enchantment to give.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7b2f844-2ba2-4465-bb80-358a4d915f36",
		Name:         "Abzan Advantage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return ctx.Game.PlayerSacrificesThenKeyedForEffect(ctx.Source(), victim,
				effects.SacrificeSpec("an enchantment", effects.Enchantment()),
				"Abzan Advantage — sacrifice an enchantment", 1,
				s82OldB339AdvantageBolster, []uuid.UUID{ctx.Controller()})
		},
	})
}

// s82OldB339AdvantageBolster is "Bolster 1" for the caster once the sacrifice is
// done (carry: the caster). An on-disk key: never renamed, never reused.
var s82OldB339AdvantageBolster = game.RegisterPromptRunThen("prompt-run/s82-old-b339-abzan-advantage-bolster", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) < 1 {
		return nil
	}
	return effects.Bolster{N: 1}.Apply(effects.KeyedContext(g, r.Carry[0], r.Source))
})
