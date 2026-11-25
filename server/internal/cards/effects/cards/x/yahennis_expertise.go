package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yahenni's Expertise — Sorcery {2}{B}{B} (Aether Revolt, S82 older sets, #5536):
//
//	"All creatures get -3/-3 until end of turn.
//	 You may cast a spell with mana value 3 or less from your hand without
//	 paying its mana cost."
//
// The shrink is applied first to the creatures on the battlefield as the
// spell resolves (CR 611.2c); the free cast then waits for the prompt the
// offer asks (CR 608.2c), after state-based actions have been checked only
// once the whole spell has finished, as for Rishkar's Expertise.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "692babfa-9b74-415e-a560-abae4ccb006b",
		Name:         "Yahenni's Expertise",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 3}},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostUntilEOT{Match: effects.Creature(), Power: -3, Toughness: -3, Label: "Yahenni's Expertise — -3/-3"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				effects.OfferFreeHandCast(ctx, ctx.Controller(), yahennisExpertiseFreeCast,
					"a spell with mana value 3 or less", effects.SpellWithManaValueAtMost(3))
				return nil
			}),
		),
	})
}

var yahennisExpertiseFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/yahennis-expertise-cast", "Yahenni's Expertise")
